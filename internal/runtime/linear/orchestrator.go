package linear

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"time"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/linearapi"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	"go.orx.me/apps/butter/internal/runtime/linearconn"
	"go.orx.me/apps/butter/internal/runtime/memoryhook"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// AppName is the ADK app every Linear Agent Session lives under. It is
	// fixed rather than the App's display name, which can be renamed.
	AppName = "linear"
	// ChannelType tags Linear turns in invocation records and traces.
	ChannelType = "linear"

	// defaultMaxRun applies when a Linear App leaves max_run_seconds unset.
	defaultMaxRun = 1800 * time.Second
	// postTimeout bounds one activity post.
	postTimeout = 15 * time.Second
)

// ErrSessionBusy means another owner holds the delivery's processing
// record. The event is left unacknowledged and redelivered.
var ErrSessionBusy = errors.New("linear delivery is being handled elsewhere")

// AgentRunner is the slice of the runner service the orchestrator needs.
type AgentRunner interface {
	// ResolveAgentRef maps a workspace-scoped agent_id to a runnable name.
	ResolveAgentRef(workspaceID, agentID string) (string, bool)
	// RunTurnSSE runs one turn.
	RunTurnSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string,
		ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback,
		onCompaction runner.CompactionCallback) (*runner.TurnResult, error)
	// GetSession loads a session; an error means there is none yet.
	GetSession(ctx context.Context, channelName, sessionID, userID string) (session.Session, error)
}

// TokenProvider hands out installation access tokens.
type TokenProvider interface {
	AccessToken(ctx context.Context, workspaceID, installationID string) (string, error)
	MarkRejected(ctx context.Context, workspaceID, installationID string, cause error)
}

// Orchestrator turns one accepted Linear event into an Agent turn answered
// in the Linear Agent Session.
//
// It re-reads the App and Installation rather than trusting the queued
// snapshot for authorization: an App disabled, or a user removed from the
// allowlist, after acceptance must not get an Agent run.
type Orchestrator struct {
	repo       linearrepo.Repository
	runner     AgentRunner
	tokens     TokenProvider
	linear     *linearapi.Client
	coord      SessionCoordinator
	baseURL    func(ctx context.Context) string
	processing linearprocessing.Repository
	// preAgentBackoff is the base delay between pre-Agent retries.
	preAgentBackoff time.Duration
}

func NewOrchestrator(repo linearrepo.Repository, agents AgentRunner, tokens TokenProvider, client *linearapi.Client) *Orchestrator {
	return &Orchestrator{
		repo: repo, runner: agents, tokens: tokens, linear: client,
		coord:           NewMemoryCoordinator(),
		preAgentBackoff: defaultPreAgentBackoff,
	}
}

// SetSessionCoordinator wires the cross-Pod session lease and follow-up
// list. The default serializes sessions within this process only.
func (o *Orchestrator) SetSessionCoordinator(coord SessionCoordinator) {
	if coord != nil {
		o.coord = coord
	}
}

// SetProcessingRepo wires the retry-boundary state machine (ADR-0009).
func (o *Orchestrator) SetProcessingRepo(repo linearprocessing.Repository) { o.processing = repo }

// SetExternalBaseURL wires where the dashboard lives, for the link from a
// Linear Agent Session back to the Butter session.
func (o *Orchestrator) SetExternalBaseURL(provider func(ctx context.Context) string) { o.baseURL = provider }

// SessionID derives the Butter session of one Linear Agent Session and
// Agent. Re-pointing the App to another Agent starts a fresh history.
func SessionID(appID, agentSessionID, agentID string) string {
	return "linear:" + appID + ":" + agentSessionID + ":" + agentID
}

// SessionUserID is the session's user: the Linear organization, scoped to
// the App. A Linear Agent Session belongs to everyone on the issue, not to
// one person in it.
func SessionUserID(appID, organizationID string) string {
	return "linear:" + appID + ":" + organizationID
}

// prompt is one message a turn answers, with the processing record that
// tracks it. Only the event being handled has a claim lease; follow-ups a
// holder drains are serialized by the session lease instead.
type prompt struct {
	text   string
	userID string
	record *agentsv1.LinearProcessingRecord
	lease  string
}

// turn is one event being handled, with the context its activities need.
type turn struct {
	o     *Orchestrator
	event *Event
	app   *agentsv1.LinearApp
	token string
	// own is the handled event's prompt; its record and claim come from
	// the delivery.
	own *prompt
}

func (t *turn) logger(ctx context.Context) *slog.Logger {
	return log.FromContext(ctx).With(
		"workspace_id", t.event.WorkspaceID, "app_id", t.event.AppID,
		"agent_session_id", t.event.AgentSessionID, "delivery_id", t.event.DeliveryID)
}

// post delivers one activity, redacted and bounded. A lost activity must
// never fail a run, so failures are logged, and a rejected token marks the
// installation for reinstall.
func (t *turn) post(ctx context.Context, a linearapi.Activity) error {
	a.Body = boundBody(a.Body)
	a.Parameter = boundParameter(a.Parameter)
	postCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), postTimeout)
	defer cancel()
	err := t.o.linear.CreateActivity(postCtx, t.token, t.event.AgentSessionID, a)
	if err != nil {
		if errors.Is(err, linearapi.ErrUnauthorized) {
			t.o.tokens.MarkRejected(ctx, t.event.WorkspaceID, t.event.InstallationID, err)
		}
		t.logger(ctx).Warn("could not post linear activity", "type", a.Type, "err", err)
	}
	return err
}

func (t *turn) fail(ctx context.Context, body string) error {
	return t.post(ctx, linearapi.Activity{Type: linearapi.ActivityError, Body: body})
}

// routing is where a turn runs: the Agent and the derived session.
type routing struct {
	agentName string
	sessionID string
	userID    string
}

const (
	cutShortMessage = "The previous run was cut short before it finished, so it was not repeated. Send a message to continue in the same session."
	queuedMessage   = "Queued — this will run after the current task finishes."
)

// Handle implements EventHandler.
func (o *Orchestrator) Handle(ctx context.Context, event *Event) error {
	logger := log.FromContext(ctx).With("workspace_id", event.WorkspaceID, "app_id", event.AppID,
		"agent_session_id", event.AgentSessionID, "delivery_id", event.DeliveryID)

	app, err := o.repo.GetApp(ctx, event.WorkspaceID, event.AppID)
	if errors.Is(err, linearrepo.ErrNotFound) {
		logger.Info("skipping linear event for a deleted app")
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := o.repo.GetInstallation(ctx, event.WorkspaceID, event.InstallationID); errors.Is(err, linearrepo.ErrNotFound) {
		logger.Info("skipping linear event for a removed installation")
		return nil
	} else if err != nil {
		return err
	}
	var token string
	err = o.retryPreAgent(ctx, func() error {
		var tokenErr error
		token, tokenErr = o.tokens.AccessToken(ctx, event.WorkspaceID, event.InstallationID)
		if errors.Is(tokenErr, linearconn.ErrNeedsReinstall) || errors.Is(tokenErr, linearconn.ErrNoToken) {
			return permanent(tokenErr)
		}
		return tokenErr
	})
	if isPermanent(err) {
		// Nothing can be posted without a token; the dashboard shows why.
		logger.Warn("cannot answer linear event", "err", err)
		return nil
	}
	if err != nil {
		return err
	}
	t := &turn{o: o, event: event, app: app, token: token,
		own: &prompt{text: event.PromptText, userID: event.PromptingUserID}}
	if event.AppRevision != app.GetRevision() {
		logger.Info("linear app changed after acceptance",
			"accepted_revision", event.AppRevision, "current_revision", app.GetRevision())
	}

	action, err := o.claim(ctx, t)
	if errors.Is(err, linearprocessing.ErrInProgress) {
		logger.Debug("linear delivery is being handled elsewhere; deferring")
		return ErrSessionBusy
	}
	if err != nil {
		return err
	}
	switch action {
	case linearprocessing.ClaimAcknowledge:
		logger.Info("acknowledging linear delivery without repeating completed or uncertain work")
		// A holder that crashed may have left follow-ups behind: pick them
		// up if the session is free.
		return o.recoverSession(ctx, t)
	case linearprocessing.ClaimReportInterrupted:
		_ = t.fail(ctx, cutShortMessage)
		o.release(ctx, t.own)
		return o.recoverSession(ctx, t)
	case linearprocessing.ClaimResumeDelivery:
		defer o.release(ctx, t.own)
		_ = o.deliver(ctx, t, []*prompt{t.own})
		return nil
	}
	defer o.release(ctx, t.own)
	ctx, stopHeartbeat := o.heartbeat(ctx, t.own)
	defer stopHeartbeat()

	if !app.GetInboundEnabled() {
		return o.answer(ctx, t, linearapi.ActivityError,
			"This Linear App is disabled in Butter, so nothing will run. A workspace admin can enable it again.")
	}
	if !admitted(app.GetAllowedUserIds(), event.PromptingUserID) {
		logger.Info("linear user not admitted", "prompting_user_id", event.PromptingUserID)
		return o.answer(ctx, t, linearapi.ActivityError,
			"You are not allowed to use this agent. Ask a Butter workspace admin to add your Linear user ID to the Linear App's allowlist.")
	}
	if event.Stop {
		return o.answer(ctx, t, linearapi.ActivityResponse, "Nothing is running, so there is nothing to stop.")
	}

	route, ok := o.route(t)
	if !ok {
		return o.answer(ctx, t, linearapi.ActivityError,
			fmt.Sprintf("The Agent this Linear App routes to (%s) is not available in Butter right now.", app.GetAgentId()))
	}
	if event.Action == ActionCreated && t.own.record.GetAttempts() <= 1 {
		// Linear marks a session unresponsive without an activity within
		// seconds of its creation: acknowledge before any slow work.
		_ = t.post(ctx, linearapi.Activity{Type: linearapi.ActivityThought, Body: acknowledgement(event.Issue, route.agentName)})
		o.linkSession(ctx, t, route)
	}

	hold, backlog, err := o.coord.EnqueueOrAcquire(ctx, route.sessionID, FollowUp{
		RecordID:        t.own.record.GetId(),
		DeliveryID:      event.DeliveryID,
		Text:            event.PromptText,
		PromptingUserID: event.PromptingUserID,
	})
	if err != nil {
		return err
	}
	if hold == nil {
		// Another turn holds the session: this message waits for it and
		// the delivery is done.
		if err := o.recordStatus(ctx, t.own, agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_QUEUED, ""); err != nil {
			logger.Warn("could not record queued linear follow-up", "err", err)
		}
		_ = t.post(ctx, linearapi.Activity{Type: linearapi.ActivityThought, Body: queuedMessage})
		return nil
	}
	prompts := append(o.loadFollowUps(ctx, t, backlog), t.own)
	return o.holdSession(ctx, t, hold, route, prompts)
}

// route resolves the App's Agent and the session it runs in.
func (o *Orchestrator) route(t *turn) (routing, bool) {
	agentName, ok := o.runner.ResolveAgentRef(t.event.WorkspaceID, t.app.GetAgentId())
	if !ok {
		return routing{}, false
	}
	return routing{
		agentName: agentName,
		sessionID: SessionID(t.app.GetId(), t.event.AgentSessionID, t.app.GetAgentId()),
		userID:    SessionUserID(t.app.GetId(), t.event.OrganizationID),
	}, true
}

// holdSession runs turns while holding the session: first prompts, then
// whatever follow-ups queued meanwhile, until release-or-drain releases.
func (o *Orchestrator) holdSession(ctx context.Context, t *turn, hold SessionHold, route routing, prompts []*prompt) error {
	logger := t.logger(ctx)
	runCtx, stop := contextWithPeerCancellation(ctx, hold.Context())
	defer stop()
	o.sweepInterrupted(runCtx, t, prompts)
	for {
		if len(prompts) > 0 {
			if err := o.run(runCtx, t, route, prompts); err != nil {
				hold.Abandon()
				return err
			}
		}
		if runCtx.Err() != nil {
			// Shutting down or fenced out: leave queued follow-ups for the
			// next holder, and this delivery for redelivery.
			hold.Abandon()
			return runCtx.Err()
		}
		next, err := hold.ReleaseOrDrain(ctx)
		if err != nil {
			logger.Warn("lost the linear session while draining follow-ups", "err", err)
			return nil
		}
		if len(next) == 0 {
			return nil
		}
		prompts = o.loadFollowUps(ctx, t, next)
	}
}

// recoverSession takes a free session to settle what a crashed holder left:
// records it was running, and follow-ups it never drained.
func (o *Orchestrator) recoverSession(ctx context.Context, t *turn) error {
	route, ok := o.route(t)
	if !ok {
		return nil
	}
	hold, backlog, err := o.coord.TryAcquire(ctx, route.sessionID)
	if err != nil || hold == nil {
		return nil
	}
	return o.holdSession(ctx, t, hold, route, o.loadFollowUps(ctx, t, backlog))
}

// loadFollowUps turns queued follow-ups into prompts with their records.
// Follow-ups drained together advance under one invocation ID.
func (o *Orchestrator) loadFollowUps(ctx context.Context, t *turn, followUps []FollowUp) []*prompt {
	prompts := make([]*prompt, 0, len(followUps))
	invocation := ""
	for _, f := range followUps {
		p := &prompt{text: f.Text, userID: f.PromptingUserID}
		if o.processing != nil && f.RecordID != "" {
			record, err := o.processing.Get(ctx, t.event.WorkspaceID, f.RecordID)
			if err != nil {
				t.logger(ctx).Warn("could not load queued linear follow-up", "record_id", f.RecordID, "err", err)
			} else {
				if invocation == "" {
					invocation = record.GetInvocationId()
				}
				record.InvocationId = invocation
				p.record = record
			}
		}
		prompts = append(prompts, p)
	}
	return prompts
}

// sweepInterrupted settles records of this session a dead holder left
// running. Only one holder runs a session at a time, so any PROCESSING
// record not in hand is from a turn that will never finish.
func (o *Orchestrator) sweepInterrupted(ctx context.Context, t *turn, inHand []*prompt) {
	if o.processing == nil {
		return
	}
	stale, err := o.processing.List(ctx, linearprocessing.Filter{
		WorkspaceID:    t.event.WorkspaceID,
		AppID:          t.event.AppID,
		AgentSessionID: t.event.AgentSessionID,
		Status:         agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING,
	})
	if err != nil {
		t.logger(ctx).Warn("could not look for interrupted linear turns", "err", err)
		return
	}
	swept := 0
	for _, record := range stale {
		if slices.ContainsFunc(inHand, func(p *prompt) bool { return p.record.GetId() == record.GetId() }) {
			continue
		}
		linearprocessing.MarkInterruptedUncertain(record)
		if _, err := o.processing.Update(ctx, record); err != nil {
			t.logger(ctx).Warn("could not settle an interrupted linear turn", "record_id", record.GetId(), "err", err)
			continue
		}
		swept++
	}
	if swept > 0 {
		_ = t.fail(ctx, cutShortMessage)
	}
}

// run invokes the Agent for prompts, persists its reply, then posts it.
func (o *Orchestrator) run(ctx context.Context, t *turn, route routing, prompts []*prompt) error {
	logger := t.logger(ctx)
	hasHistory := false
	if sess, err := o.runner.GetSession(ctx, AppName, route.sessionID, route.userID); err == nil && sess != nil && sess.Events().Len() > 0 {
		hasHistory = true
	}
	texts := make([]string, 0, len(prompts))
	principal := ""
	for _, p := range prompts {
		texts = append(texts, p.text)
		if p.userID != "" {
			principal = p.userID
		}
	}
	input := turnInput(t.event, texts, hasHistory)

	maxRun := defaultMaxRun
	if t.app.MaxRunSeconds != nil {
		maxRun = time.Duration(t.app.GetMaxRunSeconds()) * time.Second
	}
	runCtx := ctx
	if maxRun > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, maxRun)
		defer cancel()
	}

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        prompts[0].record.GetInvocationId(),
		SessionId:   route.sessionID,
		UserId:      route.userID,
		ChannelName: AppName,
		ChannelType: ChannelType,
		ChatId:      t.event.AgentSessionID,
		WorkspaceId: t.event.WorkspaceID,
		Metadata: map[string]string{
			"linear_app_id":          t.app.GetId(),
			"linear_app":             t.app.GetDisplayName(),
			"linear_organization_id": t.event.OrganizationID,
			"linear_agent_session":   t.event.AgentSessionID,
			"linear_issue":           t.event.Issue.Identifier,
		},
	}
	if principal != "" {
		ctxInfo.Metadata[memoryhook.PrincipalMetadataKey] = "linear:" + principal
	}

	// From here the Agent may run tools with side effects: a crash is no
	// longer safely retryable.
	for _, p := range prompts {
		if err := o.recordStatus(ctx, p, agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING, ""); err != nil {
			return err
		}
	}
	logger.Info("invoking linear agent", "agent", route.agentName, "session_id", route.sessionID,
		"has_history", hasHistory, "prompts", len(prompts))

	started := time.Now()
	result, err := o.runner.RunTurnSSE(runCtx, route.agentName, []*genai.Part{{Text: input}}, "", ctxInfo, nil, nil)
	if err != nil {
		if ctx.Err() != nil {
			// Shutting down or fenced out mid-turn: the records stay
			// PROCESSING for the reclaim to settle honestly.
			return ctx.Err()
		}
		body := "The agent could not finish: " + sanitizeError(err)
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			body = fmt.Sprintf("The agent timed out after %s and was stopped.", formatElapsed(time.Since(started)))
		}
		logger.Warn("linear agent turn failed", "err", err)
		// The Agent may have run tools: dead-letter, never rerun.
		o.recordUncertain(ctx, t, prompts, err)
		_ = t.fail(ctx, body)
		return nil
	}
	output := strings.TrimSpace(result.Output)
	if output == "" {
		output = "The agent finished without a text reply."
	}
	// Persist the reply before posting it: a failed post is then a resend,
	// never a rerun.
	if err := o.persistReply(ctx, prompts, linearapi.Activity{
		Type: linearapi.ActivityResponse,
		Body: truncateResponse(output, o.externalURL(ctx, route)),
	}); err != nil {
		o.recordUncertain(ctx, t, prompts, err)
		return nil
	}
	_ = o.deliver(ctx, t, prompts)
	return nil
}

// linkSession points the Linear Agent Session at the Butter session in the
// dashboard. Best effort.
func (o *Orchestrator) linkSession(ctx context.Context, t *turn, route routing) {
	link := o.externalURL(ctx, route)
	if link == "" {
		return
	}
	postCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), postTimeout)
	defer cancel()
	if err := o.linear.SetSessionExternalURL(postCtx, t.token, t.event.AgentSessionID, "Butter", link); err != nil {
		t.logger(ctx).Warn("could not link linear session to butter", "err", err)
	}
}

func (o *Orchestrator) externalURL(ctx context.Context, route routing) string {
	if o.baseURL == nil {
		return ""
	}
	base := strings.TrimRight(o.baseURL(ctx), "/")
	if base == "" {
		return ""
	}
	return base + "/sessions/detail?" + url.Values{"app": {AppName}, "user": {route.userID}, "sid": {route.sessionID}}.Encode()
}

// admitted applies the allowlist: empty admits everyone, otherwise only
// listed users, and an unknown user is never listed.
func admitted(allowlist []string, userID string) bool {
	if len(allowlist) == 0 {
		return true
	}
	return userID != "" && slices.Contains(allowlist, strings.ToLower(userID))
}

// contextWithPeerCancellation cancels ctx's child when peer ends too.
func contextWithPeerCancellation(ctx, peer context.Context) (context.Context, func()) {
	merged, cancel := context.WithCancel(ctx)
	stopPeer := context.AfterFunc(peer, cancel)
	return merged, func() {
		stopPeer()
		cancel()
	}
}
