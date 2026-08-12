package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/notify"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/forum"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	defaultMaxStoredOutputBytes = 4096
	automationAuthorKind        = "system"
	// automationSessionPrefix marks every session coordinate the engine mints
	// for an invoke_agent turn; the resume listeners rely on it to skip
	// ordinary chat traffic (mirrors cron's cronSessionPrefix).
	automationSessionPrefix = "automation:"
)

// Session coordinates for an invoke_agent turn are deterministic from the
// automation name, workspace, and run ID, so invokeAgent (which runs the turn)
// and waitForInput / the resume path all derive the same values.
func automationSessionAppName(name string) string       { return automationSessionPrefix + name }
func automationSessionUserID(workspaceID string) string { return automationSessionPrefix + workspaceID }
func automationSessionID(runID string) string           { return automationSessionPrefix + runID }

var ErrAutomationDisabled = errors.New("automation disabled")

type runnerService interface {
	ResolveAgentRef(workspaceID, agentID string) (string, bool)
	RunTurnSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (*runner.TurnResult, error)
}

type notifySender interface {
	Send(ctx context.Context, workspaceID string, target *agentsv1.NotifyTarget, msg notify.Message) error
}

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// sessionDeleter is the slice of the ADK session.Service the engine needs to
// delete a run's per-run session once its execution is terminal, so paused
// sessions do not accumulate (mirrors cron's cleanup, #134). Optional.
type sessionDeleter interface {
	Delete(ctx context.Context, req *session.DeleteRequest) error
}

type EngineOptions struct {
	Runner          runnerService
	NotifyGroupRepo configrepo.NotifyGroupRepository
	Notifier        notifySender
	ForumRepo       forum.Repository
	HTTPClient      httpDoer
	// SessionService, if set, deletes a run's session when the run reaches a
	// terminal state via the resume path.
	SessionService sessionDeleter
	// Context is the base context for resume-path work driven by listeners
	// (which carry no context of their own). Defaults to context.Background().
	Context context.Context
}

// Engine executes automation definitions and records run/step state.
type Engine struct {
	defRepo  DefinitionRepo
	runRepo  RunRepo
	stepRepo StepRunRepo

	runner          runnerService
	notifyGroupRepo configrepo.NotifyGroupRepository
	notifier        notifySender
	forumRepo       forum.Repository
	httpClient      httpDoer
	sessionSvc      sessionDeleter
	baseCtx         context.Context

	mu      sync.Mutex
	running map[string]*runningAutomation
}

type runningAutomation struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// SetTelegramDelivery wires the unified Telegram sender so notify-group steps
// can reach Telegram Destinations. Without it, Telegram targets fail
// explicitly rather than being skipped.
func (e *Engine) SetTelegramDelivery(delivery notify.TelegramDelivery) {
	if sender, ok := e.notifier.(*notify.Sender); ok {
		sender.SetTelegramDelivery(delivery)
	}
}

func NewEngine(defRepo DefinitionRepo, runRepo RunRepo, stepRepo StepRunRepo, opts EngineOptions) *Engine {
	notifier := opts.Notifier
	if notifier == nil {
		notifier = notify.NewSender(nil)
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: notify.DefaultHTTPTimeout}
	}
	baseCtx := opts.Context
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	return &Engine{
		defRepo:         defRepo,
		runRepo:         runRepo,
		stepRepo:        stepRepo,
		runner:          opts.Runner,
		notifyGroupRepo: opts.NotifyGroupRepo,
		notifier:        notifier,
		forumRepo:       opts.ForumRepo,
		httpClient:      httpClient,
		sessionSvc:      opts.SessionService,
		baseCtx:         baseCtx,
		running:         make(map[string]*runningAutomation),
	}
}

// RunNow loads a workspace-scoped automation by name and executes it with a
// manual trigger.
func (e *Engine) RunNow(ctx context.Context, workspaceID, name, triggerPayloadJSON string) (*agentsv1.AutomationRun, error) {
	if e == nil || e.defRepo == nil {
		return nil, errors.New("automation engine definition repo is not configured")
	}
	a, err := e.defRepo.Get(ctx, workspaceID, name)
	if err != nil {
		return nil, err
	}
	if !a.GetEnabled() {
		return nil, ErrAutomationDisabled
	}
	return e.Execute(ctx, a, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, triggerPayloadJSON)
}

// Execute runs the provided automation definition. It is used by manual runs
// today and by scheduled/event triggers as they are wired in.
func (e *Engine) Execute(ctx context.Context, a *agentsv1.Automation, triggerType agentsv1.AutomationTriggerType, triggerPayloadJSON string) (*agentsv1.AutomationRun, error) {
	if e == nil || e.runRepo == nil || e.stepRepo == nil {
		return nil, errors.New("automation engine repositories are not configured")
	}
	if a == nil {
		return nil, errors.New("automation is required")
	}
	if triggerType == agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_UNSPECIFIED {
		triggerType = a.GetTrigger().GetType()
	}
	if triggerType == agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_UNSPECIFIED {
		triggerType = agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL
	}
	switch effectiveAutomationConcurrency(a.GetPolicy()) {
	case agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_ALLOW:
		return e.executeRun(ctx, a, triggerType, triggerPayloadJSON)
	case agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_SKIP,
		agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED:
		return e.executeWithConcurrency(ctx, a, triggerType, triggerPayloadJSON, false, false)
	case agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_QUEUE:
		return e.executeWithConcurrency(ctx, a, triggerType, triggerPayloadJSON, true, false)
	case agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_REPLACE:
		return e.executeWithConcurrency(ctx, a, triggerType, triggerPayloadJSON, true, true)
	default:
		return nil, fmt.Errorf("unsupported automation concurrency policy %s", a.GetPolicy().GetConcurrency())
	}
}

func (e *Engine) executeWithConcurrency(ctx context.Context, a *agentsv1.Automation, triggerType agentsv1.AutomationTriggerType, payload string, wait, replace bool) (*agentsv1.AutomationRun, error) {
	key := automationID(a.GetWorkspaceId(), a.GetName())
	for {
		e.mu.Lock()
		active := e.running[key]
		if active == nil {
			runCtx, cancel := context.WithCancel(ctx)
			current := &runningAutomation{cancel: cancel, done: make(chan struct{})}
			e.running[key] = current
			e.mu.Unlock()

			defer func() {
				cancel()
				e.mu.Lock()
				if e.running[key] == current {
					delete(e.running, key)
				}
				e.mu.Unlock()
				close(current.done)
			}()
			return e.executeRun(runCtx, a, triggerType, payload)
		}

		if !wait {
			e.mu.Unlock()
			return e.recordSkipped(ctx, a, triggerType, payload, "previous execution still running")
		}
		if replace {
			active.cancel()
		}
		done := active.done
		e.mu.Unlock()
		select {
		case <-done:
		case <-ctx.Done():
			return e.recordSkipped(context.WithoutCancel(ctx), a, triggerType, payload, "execution context ended before automation could start")
		}
	}
}

func (e *Engine) executeRun(ctx context.Context, a *agentsv1.Automation, triggerType agentsv1.AutomationTriggerType, triggerPayloadJSON string) (*agentsv1.AutomationRun, error) {
	start := time.Now().UTC()
	triggerPreview, _ := truncateUTF8(triggerPayloadJSON, effectiveMaxOutputBytes(a.GetPolicy()))
	run := &agentsv1.AutomationRun{
		Id:                 uuid.NewString(),
		AutomationName:     a.GetName(),
		TriggerType:        triggerType,
		Status:             agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING,
		TriggerPayloadJson: triggerPreview,
		StartedAt:          timestamppb.New(start),
		WorkspaceId:        a.GetWorkspaceId(),
	}
	if run.TriggerPayloadJson == "" {
		run.TriggerPayloadJson = "{}"
	}
	if err := e.runRepo.Save(ctx, run); err != nil {
		return nil, err
	}

	runCtx := ctx
	cancel := func() {}
	if timeout := a.GetPolicy().GetTimeout(); timeout != nil && timeout.AsDuration() > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout.AsDuration())
	}
	defer cancel()

	payload, err := parseJSONPayload(triggerPayloadJSON)
	if err != nil {
		return e.finishRun(runCtx, run, agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED, fmt.Sprintf("invalid trigger payload: %v", err)), nil
	}
	state := newExecutionState(a, triggerType, payload)
	if skipped, reason, err := conditionsSkipped(a.GetConditions(), state.roots); err != nil {
		return e.finishRun(runCtx, run, agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED, err.Error()), nil
	} else if skipped {
		return e.finishRun(runCtx, run, agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SKIPPED, reason), nil
	}

	for i, step := range a.GetSteps() {
		select {
		case <-runCtx.Done():
			return e.finishRun(runCtx, run, runStatusFromErr(runCtx.Err()), runCtx.Err().Error()), nil
		default:
		}
		stepRun, output, pending, err := e.executeStep(runCtx, a, run, step, int32(i+1), state)
		if err != nil {
			status := agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED
			if stepRun.GetStatus() == agentsv1.AutomationStepRunStatus_AUTOMATION_STEP_RUN_STATUS_CANCELLED {
				status = agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_CANCELLED
			}
			return e.finishRun(runCtx, run, status, err.Error()), nil
		}
		if len(pending) > 0 {
			// The invoked Workflow Agent paused on a Human Input node (ADR-0003).
			// The run is not finished: it waits for a human reply on its session
			// coordinates. Later steps do not run until a reply resumes the
			// workflow and HandleTurn finalizes the run (Option A).
			agentName, agentID := e.resolveWaitAgent(a, step.GetInvokeAgent())
			return e.waitForInput(runCtx, a, run, agentName, agentID, pending), nil
		}
		state.recordStepOutput(step.GetName(), output)
	}

	return e.finishRun(runCtx, run, agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED, ""), nil
}

// waitForInput records a run paused on a Human Input node: WAITING_INPUT with
// the session coordinates a reply must address. finished_at stays unset — the
// run is not terminal — until a reply resumes it (HandleTurn) or the session is
// deleted (HandleSessionDeleted). Automations have no top-level delivery
// config, so the node's question is surfaced only through the paused step's
// recorded output (issue #176).
func (e *Engine) waitForInput(ctx context.Context, a *agentsv1.Automation, run *agentsv1.AutomationRun, agentName, agentID string, pending []runner.PendingInput) *agentsv1.AutomationRun {
	run.Status = agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT
	run.FinishedAt = nil
	run.DurationMs = 0
	run.Error = ""
	run.AgentName = agentName
	run.AgentId = agentID
	run.SessionAppName = automationSessionAppName(a.GetName())
	run.SessionUserId = automationSessionUserID(a.GetWorkspaceId())
	run.SessionId = automationSessionID(run.GetId())
	_ = e.runRepo.Save(context.WithoutCancel(ctx), run)
	log.FromContext(ctx).Info("automation run waiting for human input",
		"automation", a.GetName(),
		"workspace_id", a.GetWorkspaceId(),
		"run_id", run.GetId(),
		"agent", agentName,
		"pending_interrupts", len(pending),
	)
	return run
}

// HandleTurn observes the outcome of every agent turn (registered as a runner
// turn listener). It closes the resume loop: a human answers a paused
// automation workflow by messaging its per-run session — through ReplySession
// or any other entry point — and when a turn on that session ends with no
// pending Interrupt, the session's WAITING_INPUT run reaches SUCCEEDED with the
// resumed output. Mirrors cron's HandleTurn (ADR-0003).
//
// The initial pausing turn is invisible here: it is still Interrupted (early
// return), and it fires before waitForInput persists the WAITING_INPUT record.
func (e *Engine) HandleTurn(ctxInfo *agentsv1.ContextInfo, turn *runner.TurnResult, runErr error) {
	if ctxInfo == nil || turn == nil || runErr != nil || turn.Interrupted() {
		return
	}
	e.finalizeWaitingBySession(ctxInfo.GetChannelName(), ctxInfo.GetUserId(), ctxInfo.GetSessionId(),
		"completed by session reply", true,
		func(run *agentsv1.AutomationRun) {
			run.Status = agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED
			e.recordResumedOutput(run, turn.Output)
		})
}

// HandleSessionDeleted observes session deletions (registered as a session
// delete listener on the SessionService RPC). Deleting a paused session is the
// documented way to abandon its workflow (ADR-0002), so the session's
// WAITING_INPUT run transitions to CANCELLED — otherwise it waits forever on a
// session that no longer exists (mirrors cron's HandleSessionDeleted, #132).
func (e *Engine) HandleSessionDeleted(appName, userID, sessionID string) {
	// cleanupSession=false: the session was already deleted; nothing to clean.
	e.finalizeWaitingBySession(appName, userID, sessionID,
		"cancelled by session deletion", false,
		func(run *agentsv1.AutomationRun) {
			run.Status = agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_CANCELLED
			run.Error = "session deleted before a human answered"
		})
}

// finalizeWaitingBySession closes every WAITING_INPUT run recorded for the
// session coordinates: finalize sets the terminal state, then the record is
// persisted. action names the transition in logs; cleanupSession deletes the
// now-terminal run's session — pass false when the session is already gone.
func (e *Engine) finalizeWaitingBySession(appName, userID, sessionID, action string, cleanupSession bool, finalize func(run *agentsv1.AutomationRun)) {
	// Automation sessions are the only ones that can carry waiting runs; skip
	// the repo lookup for ordinary chat traffic.
	if !strings.HasPrefix(sessionID, automationSessionPrefix) {
		return
	}
	ctx := e.baseCtx
	logger := log.FromContext(ctx)
	waiting, err := e.runRepo.ListWaitingBySession(ctx, appName, userID, sessionID)
	if err != nil {
		logger.Error("failed to look up waiting automation runs", "session_id", sessionID, "err", err)
		return
	}
	now := time.Now().UTC()
	for _, run := range waiting {
		finalize(run)
		run.FinishedAt = timestamppb.New(now)
		run.DurationMs = now.Sub(run.GetStartedAt().AsTime()).Milliseconds()
		if saveErr := e.runRepo.Save(ctx, run); saveErr != nil {
			logger.Error("failed to finalize waiting automation run",
				"automation", run.GetAutomationName(),
				"workspace_id", run.GetWorkspaceId(),
				"run_id", run.GetId(),
				"err", saveErr,
			)
			continue
		}
		logger.Info("waiting automation run "+action,
			"automation", run.GetAutomationName(),
			"workspace_id", run.GetWorkspaceId(),
			"run_id", run.GetId(),
			"duration_ms", run.GetDurationMs(),
		)
		if cleanupSession {
			e.cleanupSession(run.GetSessionAppName(), run.GetSessionUserId(), run.GetSessionId())
		}
	}
}

// recordResumedOutput writes the workflow's post-answer output onto the run's
// paused step — the highest-order step run, which is the invoke_agent step that
// waited. Automations store output on step runs (the run has no output field),
// so this is where the "real output" of a resumed run becomes observable.
func (e *Engine) recordResumedOutput(run *agentsv1.AutomationRun, output string) {
	if e.stepRepo == nil || strings.TrimSpace(output) == "" {
		return
	}
	steps, err := e.stepRepo.ListByRun(e.baseCtx, run.GetWorkspaceId(), run.GetId())
	if err != nil || len(steps) == 0 {
		return
	}
	paused := steps[len(steps)-1]
	paused.OutputJson = marshalJSON(map[string]any{"response": output})
	paused.Truncated = false
	if err := e.stepRepo.Save(e.baseCtx, paused); err != nil {
		log.FromContext(e.baseCtx).Warn("failed to record resumed automation step output",
			"run_id", run.GetId(), "step", paused.GetStepName(), "err", err)
	}
}

// cleanupSession deletes a terminal run's per-run session. Failures only log —
// a leftover session is debris, not an error the run should surface.
func (e *Engine) cleanupSession(appName, userID, sessionID string) {
	if e.sessionSvc == nil {
		return
	}
	if err := e.sessionSvc.Delete(e.baseCtx, &session.DeleteRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	}); err != nil {
		log.FromContext(e.baseCtx).Warn("failed to delete finished automation session",
			"app_name", appName, "session_id", sessionID, "err", err)
	}
}

func (e *Engine) recordSkipped(ctx context.Context, a *agentsv1.Automation, triggerType agentsv1.AutomationTriggerType, triggerPayloadJSON, reason string) (*agentsv1.AutomationRun, error) {
	now := time.Now().UTC()
	triggerPreview, _ := truncateUTF8(triggerPayloadJSON, effectiveMaxOutputBytes(a.GetPolicy()))
	run := &agentsv1.AutomationRun{
		Id:                 uuid.NewString(),
		AutomationName:     a.GetName(),
		TriggerType:        triggerType,
		Status:             agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SKIPPED,
		TriggerPayloadJson: triggerPreview,
		Error:              reason,
		StartedAt:          timestamppb.New(now),
		FinishedAt:         timestamppb.New(now),
		WorkspaceId:        a.GetWorkspaceId(),
	}
	if run.TriggerPayloadJson == "" {
		run.TriggerPayloadJson = "{}"
	}
	if err := e.runRepo.Save(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func (e *Engine) finishRun(ctx context.Context, run *agentsv1.AutomationRun, status agentsv1.AutomationRunStatus, errText string) *agentsv1.AutomationRun {
	finished := time.Now().UTC()
	run.Status = status
	run.FinishedAt = timestamppb.New(finished)
	run.DurationMs = finished.Sub(run.GetStartedAt().AsTime()).Milliseconds()
	run.Error = errText
	_ = e.runRepo.Save(context.WithoutCancel(ctx), run)
	return run
}

func (e *Engine) executeStep(ctx context.Context, a *agentsv1.Automation, run *agentsv1.AutomationRun, step *agentsv1.AutomationStep, order int32, state *executionState) (*agentsv1.AutomationStepRun, string, []runner.PendingInput, error) {
	policy := effectiveStepPolicy(a.GetPolicy(), step.GetPolicy())
	inputJSON := stepInputJSON(step)
	inputPreview, _ := truncateUTF8(inputJSON, effectiveMaxOutputBytes(policy))
	now := time.Now().UTC()
	stepRun := &agentsv1.AutomationStepRun{
		Id:             uuid.NewString(),
		RunId:          run.GetId(),
		AutomationName: a.GetName(),
		StepName:       step.GetName(),
		StepType:       step.GetType(),
		Status:         agentsv1.AutomationStepRunStatus_AUTOMATION_STEP_RUN_STATUS_RUNNING,
		InputJson:      inputPreview,
		StartedAt:      timestamppb.New(now),
		Order:          order,
		WorkspaceId:    a.GetWorkspaceId(),
	}
	if err := e.stepRepo.Save(ctx, stepRun); err != nil {
		return nil, "", nil, err
	}

	output, invocationID, pending, attempts, err := e.executeStepWithRetry(ctx, a, run, step, policy, state)
	finished := time.Now().UTC()
	stepRun.AttemptCount = attempts
	stepRun.FinishedAt = timestamppb.New(finished)
	stepRun.DurationMs = finished.Sub(stepRun.GetStartedAt().AsTime()).Milliseconds()
	stepRun.InvocationId = invocationID
	stepRun.OutputJson, stepRun.Truncated = truncateOutputJSON(output, effectiveMaxOutputBytes(policy))
	if err != nil {
		stepRun.Status = stepStatusFromErr(err)
		stepRun.Error = err.Error()
	} else {
		// A step that paused on a Human Input node still ran its turn without
		// error and produced the node's question as output — it is SUCCEEDED;
		// the run (not the step) carries the WAITING_INPUT signal via pending.
		stepRun.Status = agentsv1.AutomationStepRunStatus_AUTOMATION_STEP_RUN_STATUS_SUCCEEDED
	}
	if saveErr := e.stepRepo.Save(context.WithoutCancel(ctx), stepRun); saveErr != nil && err == nil {
		return stepRun, output, pending, saveErr
	}
	return stepRun, output, pending, err
}

func (e *Engine) executeStepWithRetry(ctx context.Context, a *agentsv1.Automation, run *agentsv1.AutomationRun, step *agentsv1.AutomationStep, policy *agentsv1.AutomationPolicy, state *executionState) (string, string, []runner.PendingInput, int32, error) {
	totalAttempts := int32(1)
	if retry := policy.GetRetry(); retry != nil && retry.GetMaxAttempts() > 0 {
		totalAttempts += retry.GetMaxAttempts()
	}

	var (
		output       string
		invocationID string
		pending      []runner.PendingInput
		err          error
	)
	for attempt := int32(1); attempt <= totalAttempts; attempt++ {
		attemptCtx := ctx
		cancel := func() {}
		if timeout := policy.GetTimeout(); timeout != nil && timeout.AsDuration() > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, timeout.AsDuration())
		}
		output, invocationID, pending, err = e.executeStepAction(attemptCtx, a, run, step, state)
		cancel()
		if err == nil {
			return output, invocationID, pending, attempt, nil
		}
		if attempt == totalAttempts || ctx.Err() != nil {
			if ctx.Err() != nil {
				return output, invocationID, nil, attempt, ctx.Err()
			}
			return output, invocationID, nil, attempt, err
		}
		backoff := time.Duration(0)
		if retry := policy.GetRetry(); retry != nil && retry.GetBackoff() != nil {
			backoff = retry.GetBackoff().AsDuration()
		}
		if backoff <= 0 {
			continue
		}
		timer := time.NewTimer(backoff)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return output, invocationID, nil, attempt, ctx.Err()
		}
	}
	return output, invocationID, nil, totalAttempts, err
}

func (e *Engine) executeStepAction(ctx context.Context, a *agentsv1.Automation, run *agentsv1.AutomationRun, step *agentsv1.AutomationStep, state *executionState) (string, string, []runner.PendingInput, error) {
	switch step.GetType() {
	case agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT:
		return e.invokeAgent(ctx, a, run, step.GetInvokeAgent())
	case agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK:
		out, err := e.callWebhook(ctx, step.GetCallWebhook())
		return out, "", nil, err
	case agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP:
		out, err := e.sendNotifyGroup(ctx, a.GetWorkspaceId(), step.GetSendNotifyGroup())
		return out, "", nil, err
	case agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CREATE_FORUM_POST:
		out, err := e.createForumPost(ctx, a, step.GetCreateForumPost())
		return out, "", nil, err
	default:
		return "", "", nil, fmt.Errorf("unsupported automation step type %s", step.GetType())
	}
}

// resolveWaitAgent derives the (runtime name, agent_id) pair a paused run
// records so a human reply addresses the right agent. Falls back to the
// step's agent_id when resolution misses — the run already paused, so the
// record must still carry the reference.
func (e *Engine) resolveWaitAgent(a *agentsv1.Automation, step *agentsv1.AutomationInvokeAgentStep) (string, string) {
	name, ok := e.runner.ResolveAgentRef(a.GetWorkspaceId(), step.GetAgentId())
	if !ok {
		return "", step.GetAgentId()
	}
	return name, step.GetAgentId()
}

func (e *Engine) invokeAgent(ctx context.Context, a *agentsv1.Automation, run *agentsv1.AutomationRun, step *agentsv1.AutomationInvokeAgentStep) (string, string, []runner.PendingInput, error) {
	if e.runner == nil {
		return "", "", nil, errors.New("runner service is not configured")
	}
	if step.GetAgentId() == "" {
		return "", "", nil, errors.New("invoke_agent.agent_id is required")
	}
	agentName, ok := e.runner.ResolveAgentRef(a.GetWorkspaceId(), step.GetAgentId())
	if !ok {
		return "", "", nil, fmt.Errorf("agent %q not found in workspace %q", step.GetAgentId(), a.GetWorkspaceId())
	}
	input := step.GetInput()
	if input == "" {
		input = "execute"
	}
	invocationID := uuid.NewString()
	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        invocationID,
		ChannelName: automationSessionAppName(a.GetName()),
		SessionId:   automationSessionID(run.GetId()),
		UserId:      automationSessionUserID(a.GetWorkspaceId()),
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChannelType: "automation",
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		WorkspaceId: a.GetWorkspaceId(),
	}
	turn, err := e.runner.RunTurnSSE(ctx, agentName, []*genai.Part{genai.NewPartFromText(input)}, step.GetModelOverride(), ctxInfo, nil, nil)
	if turn == nil {
		turn = &runner.TurnResult{}
	}
	out := marshalJSON(map[string]any{"invocation_id": invocationID, "response": turn.Output})
	if err != nil {
		return out, invocationID, nil, err
	}
	return out, invocationID, turn.Pending, nil
}

func (e *Engine) callWebhook(ctx context.Context, step *agentsv1.AutomationCallWebhookStep) (string, error) {
	if e.httpClient == nil {
		return "", errors.New("http client is not configured")
	}
	if step.GetUrl() == "" {
		return "", errors.New("call_webhook.url is required")
	}
	method := strings.ToUpper(strings.TrimSpace(step.GetMethod()))
	if method == "" {
		method = http.MethodPost
	}
	body := []byte("{}")
	if strings.TrimSpace(step.GetPayloadJson()) != "" {
		if !json.Valid([]byte(step.GetPayloadJson())) {
			return "", errors.New("call_webhook.payload_json must be valid JSON")
		}
		body = []byte(step.GetPayloadJson())
	}
	req, err := http.NewRequestWithContext(ctx, method, step.GetUrl(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range step.GetHeaders() {
		req.Header.Set(k, v)
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return marshalJSON(map[string]any{"error": err.Error()}), fmt.Errorf("call webhook: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	out := marshalJSON(map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(respBody),
	})
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return out, nil
}

func (e *Engine) sendNotifyGroup(ctx context.Context, workspaceID string, step *agentsv1.AutomationSendNotifyGroupStep) (string, error) {
	if e.notifyGroupRepo == nil || e.notifier == nil {
		return "", errors.New("notify group repository or sender is not configured")
	}
	if step.GetNotifyGroupName() == "" {
		return "", errors.New("send_notify_group.notify_group_name is required")
	}
	group, err := e.notifyGroupRepo.GetNotifyGroup(ctx, workspaceID, step.GetNotifyGroupName())
	if err != nil {
		return "", fmt.Errorf("load notify group %q: %w", step.GetNotifyGroupName(), err)
	}
	if !group.GetEnabled() {
		return marshalJSON(map[string]any{"notify_group": group.GetName(), "delivered": 0}), nil
	}
	delivered := 0
	var errs []string
	msg := notify.Message{Title: step.GetTitle(), Text: step.GetMessage()}
	for _, target := range group.GetTargets() {
		if target == nil || !target.GetEnabled() {
			continue
		}
		if err := e.notifier.Send(ctx, group.GetWorkspaceId(), target, msg); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", target.GetName(), err))
			continue
		}
		delivered++
	}
	out := marshalJSON(map[string]any{
		"notify_group": group.GetName(),
		"delivered":    delivered,
		"errors":       errs,
	})
	if len(errs) > 0 && delivered == 0 {
		return out, errors.New(strings.Join(errs, "; "))
	}
	return out, nil
}

func (e *Engine) createForumPost(ctx context.Context, a *agentsv1.Automation, step *agentsv1.AutomationCreateForumPostStep) (string, error) {
	if e.forumRepo == nil {
		return "", errors.New("forum repository is not configured")
	}
	if step.GetThreadId() == "" {
		return "", errors.New("create_forum_post.thread_id is required")
	}
	if _, err := e.forumRepo.GetThread(ctx, a.GetWorkspaceId(), step.GetThreadId()); err != nil {
		return "", fmt.Errorf("load forum thread %q: %w", step.GetThreadId(), err)
	}
	now := timestamppb.New(time.Now().UTC())
	post := &agentsv1.ForumPost{
		Id:              uuid.NewString(),
		ThreadId:        step.GetThreadId(),
		Body:            step.GetBody(),
		AuthorUserId:    "automation:" + a.GetName(),
		AuthorKind:      automationAuthorKind,
		AuthorAgentName: a.GetName(),
		CreatedAt:       now,
		UpdatedAt:       now,
		WorkspaceId:     a.GetWorkspaceId(),
	}
	if err := e.forumRepo.CreatePost(ctx, post); err != nil {
		return "", fmt.Errorf("create forum post: %w", err)
	}
	return marshalJSON(map[string]any{"post_id": post.GetId(), "thread_id": post.GetThreadId()}), nil
}

type executionState struct {
	roots map[string]any
	steps map[string]any
}

func newExecutionState(a *agentsv1.Automation, triggerType agentsv1.AutomationTriggerType, payload any) *executionState {
	steps := make(map[string]any)
	roots := map[string]any{
		"payload": payload,
		"context": map[string]any{
			"automation_name": a.GetName(),
			"workspace_id":    a.GetWorkspaceId(),
			"trigger_type":    triggerType.String(),
			"steps":           steps,
		},
		"steps": steps,
	}
	return &executionState{roots: roots, steps: steps}
}

func (s *executionState) recordStepOutput(name, output string) {
	if s == nil || name == "" {
		return
	}
	var parsed any
	if strings.TrimSpace(output) != "" && json.Unmarshal([]byte(output), &parsed) == nil {
		s.steps[name] = parsed
		return
	}
	s.steps[name] = output
}

func conditionsSkipped(conditions []*agentsv1.AutomationCondition, roots map[string]any) (bool, string, error) {
	for _, cond := range conditions {
		ok, err := conditionMatches(cond, roots)
		if err != nil {
			return false, "", err
		}
		if !ok {
			return true, fmt.Sprintf("condition %q did not match", cond.GetSelector()), nil
		}
	}
	return false, "", nil
}

func conditionMatches(cond *agentsv1.AutomationCondition, roots map[string]any) (bool, error) {
	value, exists := selectValue(roots, cond.GetSelector())
	actual := valueToString(value)
	switch cond.GetOperator() {
	case agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_EQUALS:
		return exists && actual == cond.GetValue(), nil
	case agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS:
		return !exists || actual != cond.GetValue(), nil
	case agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_CONTAINS:
		return exists && strings.Contains(actual, cond.GetValue()), nil
	case agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH:
		if !exists {
			return false, nil
		}
		return regexp.MatchString(cond.GetValue(), actual)
	case agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_EXISTS:
		return exists, nil
	case agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS:
		return !exists, nil
	default:
		return false, fmt.Errorf("unsupported automation condition operator %s", cond.GetOperator())
	}
}

func selectValue(root any, selector string) (any, bool) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nil, false
	}
	current := root
	for _, part := range strings.Split(selector, ".") {
		if part == "" {
			return nil, false
		}
		switch typed := current.(type) {
		case map[string]any:
			next, ok := typed[part]
			if !ok {
				return nil, false
			}
			current = next
		default:
			return nil, false
		}
	}
	return current, true
}

func valueToString(v any) string {
	switch typed := v.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func parseJSONPayload(raw string) (any, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	var out any
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func stepInputJSON(step *agentsv1.AutomationStep) string {
	if step == nil {
		return "{}"
	}
	b, err := protojson.Marshal(step)
	if err != nil {
		return marshalJSON(map[string]any{"name": step.GetName(), "type": step.GetType().String()})
	}
	return string(b)
}

func marshalJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func effectiveAutomationConcurrency(policy *agentsv1.AutomationPolicy) agentsv1.AutomationConcurrencyPolicy {
	if policy == nil || policy.GetConcurrency() == agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED {
		return agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_ALLOW
	}
	return policy.GetConcurrency()
}

func effectiveStepPolicy(base, override *agentsv1.AutomationPolicy) *agentsv1.AutomationPolicy {
	out := &agentsv1.AutomationPolicy{}
	if base != nil {
		out.Timeout = base.GetTimeout()
		out.Retry = base.GetRetry()
		out.Concurrency = base.GetConcurrency()
		out.MaxOutputBytes = base.GetMaxOutputBytes()
	}
	if override != nil {
		if override.GetTimeout() != nil {
			out.Timeout = override.GetTimeout()
		}
		if override.GetRetry() != nil {
			out.Retry = override.GetRetry()
		}
		if override.GetMaxOutputBytes() > 0 {
			out.MaxOutputBytes = override.GetMaxOutputBytes()
		}
	}
	if out.Timeout != nil && out.Timeout.AsDuration() < 0 {
		out.Timeout = durationpb.New(0)
	}
	return out
}

func effectiveMaxOutputBytes(policy *agentsv1.AutomationPolicy) int32 {
	if policy != nil && policy.GetMaxOutputBytes() > 0 {
		return policy.GetMaxOutputBytes()
	}
	return defaultMaxStoredOutputBytes
}

func truncateOutputJSON(s string, maxBytes int32) (string, bool) {
	if strings.TrimSpace(s) == "" {
		return "{}", false
	}
	return truncateUTF8(s, maxBytes)
}

func truncateUTF8(s string, maxBytes int32) (string, bool) {
	if maxBytes <= 0 || len(s) <= int(maxBytes) {
		return s, false
	}
	max := int(maxBytes)
	for max > 0 && !utf8.ValidString(s[:max]) {
		max--
	}
	return s[:max], true
}

func runStatusFromErr(err error) agentsv1.AutomationRunStatus {
	if errors.Is(err, context.Canceled) {
		return agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_CANCELLED
	}
	return agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED
}

func stepStatusFromErr(err error) agentsv1.AutomationStepRunStatus {
	if errors.Is(err, context.Canceled) {
		return agentsv1.AutomationStepRunStatus_AUTOMATION_STEP_RUN_STATUS_CANCELLED
	}
	return agentsv1.AutomationStepRunStatus_AUTOMATION_STEP_RUN_STATUS_FAILED
}
