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
	"go.orx.me/apps/butter/internal/runtime/linearconn"
	"go.orx.me/apps/butter/internal/runtime/memoryhook"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
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

// ErrSessionBusy means another turn holds the Agent Session's lease. The
// event is left unacknowledged and redelivered.
var ErrSessionBusy = errors.New("linear agent session is busy")

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
	repo    linearrepo.Repository
	runner  AgentRunner
	tokens  TokenProvider
	linear  *linearapi.Client
	guard   sessionguard.Guard
	baseURL func(ctx context.Context) string
}

func NewOrchestrator(repo linearrepo.Repository, agents AgentRunner, tokens TokenProvider, client *linearapi.Client) *Orchestrator {
	return &Orchestrator{repo: repo, runner: agents, tokens: tokens, linear: client}
}

// SetSessionGuard wires the per-session lease that serializes turns.
func (o *Orchestrator) SetSessionGuard(guard sessionguard.Guard) { o.guard = guard }

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

// turn is one event being handled, with the context its activities need.
type turn struct {
	o     *Orchestrator
	event *Event
	app   *agentsv1.LinearApp
	token string
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
	token, err := o.tokens.AccessToken(ctx, event.WorkspaceID, event.InstallationID)
	if errors.Is(err, linearconn.ErrNeedsReinstall) || errors.Is(err, linearconn.ErrNoToken) {
		// Nothing can be posted without a token; the dashboard shows why.
		logger.Warn("cannot answer linear event", "err", err)
		return nil
	}
	if err != nil {
		return err
	}
	t := &turn{o: o, event: event, app: app, token: token}
	if event.AppRevision != app.GetRevision() {
		logger.Info("linear app changed after acceptance",
			"accepted_revision", event.AppRevision, "current_revision", app.GetRevision())
	}

	if !app.GetInboundEnabled() {
		_ = t.fail(ctx, "This Linear App is disabled in Butter, so nothing will run. A workspace admin can enable it again.")
		return nil
	}
	if !admitted(app.GetAllowedUserIds(), event.PromptingUserID) {
		logger.Info("linear user not admitted", "prompting_user_id", event.PromptingUserID)
		_ = t.fail(ctx, "You are not allowed to use this agent. Ask a Butter workspace admin to add your Linear user ID to the Linear App's allowlist.")
		return nil
	}
	if event.Stop {
		_ = t.post(ctx, linearapi.Activity{Type: linearapi.ActivityResponse, Body: "Nothing is running, so there is nothing to stop."})
		return nil
	}

	agentName, ok := o.runner.ResolveAgentRef(event.WorkspaceID, app.GetAgentId())
	if !ok {
		_ = t.fail(ctx, fmt.Sprintf("The Agent this Linear App routes to (%s) is not available in Butter right now.", app.GetAgentId()))
		return nil
	}
	sessionID := SessionID(app.GetId(), event.AgentSessionID, app.GetAgentId())
	userID := SessionUserID(app.GetId(), event.OrganizationID)
	if event.Action == ActionCreated {
		// Linear marks a session unresponsive without an activity within
		// seconds of its creation: acknowledge before any slow work.
		_ = t.post(ctx, linearapi.Activity{Type: linearapi.ActivityThought, Body: acknowledgement(event.Issue, agentName)})
		o.linkSession(ctx, t, sessionID, userID)
	}

	runCtx := ctx
	if o.guard != nil {
		leaseCtx, release, acquired, err := o.guard.Acquire(ctx, sessionID)
		if err != nil {
			return err
		}
		if !acquired {
			logger.Debug("linear session busy; deferring")
			return ErrSessionBusy
		}
		defer release()
		runCtx = leaseCtx
	}
	return o.run(runCtx, t, agentName, sessionID, userID)
}

// run invokes the Agent and posts its outcome.
func (o *Orchestrator) run(ctx context.Context, t *turn, agentName, sessionID, userID string) error {
	logger := t.logger(ctx)
	hasHistory := false
	if sess, err := o.runner.GetSession(ctx, AppName, sessionID, userID); err == nil && sess != nil && sess.Events().Len() > 0 {
		hasHistory = true
	}
	input := turnInput(t.event, hasHistory)

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
		SessionId:   sessionID,
		UserId:      userID,
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
	if t.event.PromptingUserID != "" {
		ctxInfo.Metadata[memoryhook.PrincipalMetadataKey] = "linear:" + t.event.PromptingUserID
	}
	logger.Info("invoking linear agent", "agent", agentName, "session_id", sessionID, "has_history", hasHistory)

	started := time.Now()
	result, err := o.runner.RunTurnSSE(runCtx, agentName, []*genai.Part{{Text: input}}, "", ctxInfo, nil, nil)
	if err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
			_ = t.fail(ctx, fmt.Sprintf("The agent timed out after %s and was stopped.", formatElapsed(time.Since(started))))
		} else {
			_ = t.fail(ctx, "The agent could not finish: "+sanitizeError(err))
		}
		logger.Warn("linear agent turn failed", "err", err)
		// The Agent may have run tools: never rerun it automatically.
		return nil
	}
	output := strings.TrimSpace(result.Output)
	if output == "" {
		output = "The agent finished without a text reply."
	}
	_ = t.post(ctx, linearapi.Activity{Type: linearapi.ActivityResponse, Body: truncateResponse(output, o.externalURL(ctx, sessionID, userID))})
	return nil
}

// linkSession points the Linear Agent Session at the Butter session in the
// dashboard. Best effort.
func (o *Orchestrator) linkSession(ctx context.Context, t *turn, sessionID, userID string) {
	link := o.externalURL(ctx, sessionID, userID)
	if link == "" {
		return
	}
	postCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), postTimeout)
	defer cancel()
	if err := o.linear.SetSessionExternalURL(postCtx, t.token, t.event.AgentSessionID, "Butter", link); err != nil {
		t.logger(ctx).Warn("could not link linear session to butter", "err", err)
	}
}

func (o *Orchestrator) externalURL(ctx context.Context, sessionID, userID string) string {
	if o.baseURL == nil {
		return ""
	}
	base := strings.TrimRight(o.baseURL(ctx), "/")
	if base == "" {
		return ""
	}
	return base + "/sessions/detail?" + url.Values{"app": {AppName}, "user": {userID}, "sid": {sessionID}}.Encode()
}

// admitted applies the allowlist: empty admits everyone, otherwise only
// listed users, and an unknown user is never listed.
func admitted(allowlist []string, userID string) bool {
	if len(allowlist) == 0 {
		return true
	}
	return userID != "" && slices.Contains(allowlist, strings.ToLower(userID))
}
