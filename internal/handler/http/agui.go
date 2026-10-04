package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	aguitypes "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/aguitool"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/interrupt"
	"go.orx.me/apps/butter/internal/runtime/runstate"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	"go.orx.me/apps/butter/internal/runtime/streamorch"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// aguiSessionPrefix namespaces AG-UI sessions. threadId is chosen by the
// client, so it must not become a bare session ID in the namespace shared with
// every other adapter (chat-, openai-, tg:…).
const aguiSessionPrefix = "agui-"

// aguiAppName is the ContextInfo channel name for AG-UI runs, and part of the
// ADK session key alongside the user ID and session ID.
const aguiAppName = "agui"

// aguiRequestInputPayloadKey is the key ADK's workflow engine reads a human
// input response from. It exports no constant for it; internal/runtime/interrupt
// documents the same wire format.
const aguiRequestInputPayloadKey = "payload"

// AGUISessionLeaseKeyPrefix namespaces the Redis leases that serialize AG-UI
// runs per (caller, thread) across Pods.
const AGUISessionLeaseKeyPrefix = "butter:agui:lease:session:"

// AGUISessionLeaseTTL bounds how long a crashed Pod blocks one AG-UI thread.
// The lease is renewed during the run, so the TTL only matters for crash
// recovery; it has to exceed a renewal interval comfortably, not a whole turn.
// The run state recorded next to the lease lasts as long.
const AGUISessionLeaseTTL = 5 * time.Minute

// AGUIRunnerService is the subset of runner.Service the AG-UI handler needs:
// streamorch.Runner, so the orchestrator can be driven directly, and the
// registry lookup that maps an agent_id to the name it runs under.
type AGUIRunnerService interface {
	streamorch.Runner
	ResolveAgentRef(workspaceID, agentID string) (string, bool)
}

// AGUIHandler serves the AG-UI protocol endpoint: to signed-in dashboard users
// for every agent the runner can run, and to API and root tokens for the
// agents that opted in via enable_agui.
//
// The endpoint is stateful in AG-UI terms: the server-side session is
// authoritative, so the client's message history is not replayed into the
// runner. See docs/api.md and docs/research/ag-ui-integration.md.
type AGUIHandler struct {
	agentRepo configrepo.AgentRepository

	mu            sync.RWMutex
	runnerSvc     AGUIRunnerService
	sessionGuard  sessionguard.Guard
	sessionSvc    session.Service
	sessionTitler AGUISessionTitler
	invocations   invocation.Repository
	runStates     runstate.Store
	maxRun        time.Duration
	// heartbeat paces the comments a Detached Run's observers send while it
	// is quiet; tests shorten it.
	heartbeat time.Duration
	// deleteWait bounds how long deleting a thread waits for its lease;
	// tests shorten it.
	deleteWait time.Duration
	// readAttempts and readBackoff pace how a thread read starts over
	// (agui_read.go); tests change them.
	readAttempts int
	readBackoff  time.Duration

	// runs are this process's Detached Runs in flight.
	runs *aguiRuns

	// titles tracks background thread titles in flight, so tests can wait
	// for them.
	titles sync.WaitGroup
}

// NewAGUIHandler creates an AG-UI handler with the given agent repository.
func NewAGUIHandler(repo configrepo.AgentRepository) *AGUIHandler {
	return &AGUIHandler{
		agentRepo:    repo,
		maxRun:       AGUIDefaultMaxRunDuration,
		heartbeat:    aguiHeartbeatInterval,
		deleteWait:   aguiDeleteLeaseWait,
		readAttempts: aguiReadAttempts,
		readBackoff:  aguiReadBackoff,
		runs:         newAGUIRuns(),
	}
}

// SetRunnerService sets the runner service after bootstrap completes.
func (h *AGUIHandler) SetRunnerService(svc AGUIRunnerService) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.runnerSvc = svc
}

func (h *AGUIHandler) getRunner() AGUIRunnerService {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.runnerSvc
}

// SetSessionGuard wires cross-Pod session serialization. Without a guard only
// runner.acquireSessionTurn's in-process serialization applies, so two Pods
// could interleave one thread's history — set it whenever Redis is available.
func (h *AGUIHandler) SetSessionGuard(guard sessionguard.Guard) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessionGuard = guard
}

func (h *AGUIHandler) getSessionGuard() sessionguard.Guard {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessionGuard
}

// SetSessionService wires the ADK session store. It is what lets the handler
// validate client tool results against the pending calls actually recorded on
// the session (ADR-0002: session events are the single source of truth).
func (h *AGUIHandler) SetSessionService(svc session.Service) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessionSvc = svc
}

func (h *AGUIHandler) getSessionService() session.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessionSvc
}

// SetInvocationRepo wires the store of Invocation records. A Detached Run
// owns its record there (ADR-0016 decision 3); the store stamps each record
// it creates with this process as its owner. Without it, detaching is
// refused.
func (h *AGUIHandler) SetInvocationRepo(repo invocation.Repository) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.invocations = repo
}

func (h *AGUIHandler) getInvocations() invocation.Repository {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.invocations
}

// SetRunStateStore wires where every run records its run state next to its
// thread lease (ADR-0016 decision 6), on which a Stop is accepted (decision
// 4): Redis with several Pods, the in-process store otherwise. Thread reads
// read it to tell a run in flight. Without one, runs record none and cannot
// detach, and reads never see a run.
func (h *AGUIHandler) SetRunStateStore(store runstate.Store) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.runStates = store
}

func (h *AGUIHandler) getRunStateStore() runstate.Store {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.runStates
}

// SetMaxRunDuration bounds a Detached Run: past it the run is cancelled and
// ends FAILED. A non-positive duration keeps the current bound.
func (h *AGUIHandler) SetMaxRunDuration(d time.Duration) {
	if d <= 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.maxRun = d
}

func (h *AGUIHandler) maxRunDuration() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.maxRun
}

func (h *AGUIHandler) heartbeatInterval() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.heartbeat
}

func (h *AGUIHandler) deleteLeaseWait() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.deleteWait
}

func (h *AGUIHandler) readRetry() (attempts int, backoff time.Duration) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return max(h.readAttempts, 1), h.readBackoff
}

// Shutdown ends the Detached Runs in flight for a graceful process exit. Each
// is cancelled, ends FAILED with a shutdown reason and releases its thread's
// lease; Shutdown waits for them until ctx ends. Detached runs requested
// afterwards are refused with 503. Runs without the opt-in end with their
// requests.
func (h *AGUIHandler) Shutdown(ctx context.Context) error {
	return h.runs.shutdown(ctx)
}

// aguiSessionKey identifies one caller's AG-UI thread for serialization. The
// user ID is part of the key for the same reason it is part of the ADK session
// key: two users sharing a threadId are two conversations, not one.
func aguiSessionKey(ctxInfo *agentsv1.ContextInfo) string {
	return ctxInfo.GetUserId() + ":" + ctxInfo.GetSessionId()
}

// Register registers the AG-UI route on the Gin engine. The path segment is the
// agent's immutable agent_id, the sole agent reference on protocol interfaces.
func (h *AGUIHandler) Register(r *gin.Engine) {
	r.POST("/api/agui/:agent_id", h.RunAgent)
	r.GET("/api/agui/:agent_id/threads/:thread_id/ui", h.UISnapshot)
	r.GET("/api/agui/:agent_id/threads/:thread_id/messages", h.ThreadMessages)
	r.POST("/api/agui/:agent_id/threads/:thread_id/stop", h.StopRun)
}

// aguiErrorResponse is the body for failures that happen before the SSE stream
// opens. Once streaming has started, errors are RUN_ERROR events instead.
type aguiErrorResponse struct {
	Error string `json:"error"`
}

// aguiCodedError is the body of a pre-stream rejection a client tells apart
// by code without parsing the message: a rejected form submission, whose
// client keeps the draft and shows the message and per-field errors, and the
// Detached Run refusals.
type aguiCodedError struct {
	Error       string            `json:"error"`
	Code        string            `json:"code,omitempty"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
}

// aguiRefusal is a rejection before the stream opens: the status and the
// JSON body to answer with.
type aguiRefusal struct {
	status int
	body   any
}

func refuseAGUI(status int, err error) *aguiRefusal {
	return &aguiRefusal{status: status, body: aguiErrorResponse{Error: err.Error()}}
}

func (r *aguiRefusal) write(c *gin.Context) { c.JSON(r.status, r.body) }

// aguiRunContext is the validated, prepared state for one AG-UI run.
type aguiRunContext struct {
	input       aguitypes.RunAgentInput
	agent       *agentsv1.Agent
	agentName   string // the name the runner registers agent under
	svc         AGUIRunnerService
	parts       []*genai.Part
	ctxInfo     *agentsv1.ContextInfo
	clientTools []aguitool.Declaration

	// Shared state mapping for this run: stateArmed gates the whole feature
	// (off without a session store), stateInitial is the pre-run
	// client-visible authoritative state, and stateSnapshot requests the
	// corrective/initial STATE_SNAPSHOT.
	stateArmed    bool
	stateInitial  map[string]any
	stateSnapshot bool

	// a2uiNegotiated is set when the client declared A2UI
	// (forwardedProps.butterA2UI). It becomes live for the run only if the
	// session's binding matches.
	a2uiNegotiated bool
	// formEntries are the resume entries that submit forms; they are
	// validated under the lease, before the run starts.
	formEntries []aguiFormEntry
	// toolResults are the trailing tool-role messages; they are validated
	// against the session's pending calls under the lease, and their parts
	// follow the resume parts.
	toolResults []aguitypes.Message
	// detach is set when the client asked for a Detached Run
	// (forwardedProps.butterRun).
	detach bool
}

// RunAgent handles POST /api/agui/:agent_id.
func (h *AGUIHandler) RunAgent(c *gin.Context) {
	rc, ok := h.validateAndPrepare(c)
	if !ok {
		return
	}

	log.FromContext(c.Request.Context()).Info("agui run started",
		"workspace_id", rc.ctxInfo.GetWorkspaceId(),
		"agent", rc.agentName,
		"agent_id", rc.agent.GetAgentId(),
		"thread_id", rc.input.ThreadID,
		"run_id", rc.input.RunID,
		"session_id", rc.ctxInfo.GetSessionId(),
		"invocation_id", rc.ctxInfo.GetUuid(),
		"resume_entries", len(rc.input.Resume),
		"detached", rc.detach,
	)

	// One turn per (caller, thread) at a time across the whole fleet. The
	// lease is taken before the stream opens so a busy thread is an HTTP
	// error the client can retry, not a stream that dies mid-run, and every
	// check that reads the session runs under it.
	run, ok := h.beginRun(c, rc)
	if !ok {
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	if run.detached == nil {
		// The run lives in its request, as it always has: the sink writes to
		// the response, and a client disconnect cancels the request context,
		// which ends the run and releases the lease.
		run.openSink(newAGUISSEEmitter(c.Request.Context(), c.Writer, c.Writer.Flush))
		run.execute()
		return
	}
	// A Detached Run belongs to its own goroutine, which holds the lease
	// until its terminal state is recorded. The response is only its first
	// observer: a disconnect detaches the observer, never the run.
	fanout := run.detached.fanout
	run.openSink(fanout.emit)
	observer := fanout.attach()
	go run.execute()
	h.observe(c, observer)
}

// acquireThread takes the thread's cross-Pod session lease on ctx, answering
// 503 (lease infrastructure down) or 409 (busy) itself when it cannot.
// Without a guard it is a no-op. The returned context is cancelled if the
// lease is lost, and ends with ctx.
func (h *AGUIHandler) acquireThread(ctx context.Context, c *gin.Context, ctxInfo *agentsv1.ContextInfo) (context.Context, func(), bool) {
	guard := h.getSessionGuard()
	if guard == nil {
		return ctx, func() {}, true
	}
	leaseCtx, release, acquired, err := guard.Acquire(ctx, aguiSessionKey(ctxInfo))
	if err != nil {
		log.FromContext(ctx).Error("agui session lease unavailable",
			"session_id", ctxInfo.GetSessionId(), "err", err)
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "session lock unavailable, retry later"})
		return nil, nil, false
	}
	if !acquired {
		c.JSON(http.StatusConflict, aguiErrorResponse{Error: "a run is already in progress for this thread, retry after it finishes"})
		return nil, nil, false
	}
	return leaseCtx, release, true
}

// aguiTarget is the agent a request addresses, resolved for its caller: the
// workspace it lives in, its config, and the runner that runs it under name.
type aguiTarget struct {
	workspaceID string
	agent       *agentsv1.Agent
	runner      AGUIRunnerService
	name        string
}

// resolveAgent finds the agent the route addresses in the request's workspace,
// answering 401, 404 or 503 itself when the caller cannot reach it.
//
// A signed-in dashboard user reaches every agent in the workspace. Any other
// caller, an API token or the root token, reaches only an agent with
// enable_agui: like enable_a2a and enable_openai_api for their protocols, the
// flag opts an agent in to programmatic access. Either way an agent the runner
// cannot run (provisioning, being deleted, deleted, or not loaded) is refused
// here, as resolveAgentRunnerRef refuses it on AgentService, instead of
// opening a stream that can only fail.
func (h *AGUIHandler) resolveAgent(c *gin.Context) (aguiTarget, bool) {
	ctx := c.Request.Context()
	workspaceID, hasWorkspace := wsctx.FromContext(ctx)
	if !hasWorkspace {
		c.JSON(http.StatusUnauthorized, aguiErrorResponse{Error: "workspace required (set X-Workspace-ID header)"})
		return aguiTarget{}, false
	}
	agentID := c.Param("agent_id")
	notFound := aguiErrorResponse{Error: "agent not found: " + agentID}
	agent, err := h.agentRepo.GetAgent(ctx, workspaceID, agentID)
	if err != nil || agent == nil || !aguiReachable(ctx, agent) {
		c.JSON(http.StatusNotFound, notFound)
		return aguiTarget{}, false
	}
	svc := h.getRunner()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "runner not available"})
		return aguiTarget{}, false
	}
	name, ok := svc.ResolveAgentRef(workspaceID, agentID)
	if !ok {
		c.JSON(http.StatusNotFound, notFound)
		return aguiTarget{}, false
	}
	return aguiTarget{workspaceID: workspaceID, agent: agent, runner: svc, name: name}, true
}

// aguiReachable reports whether the request's caller may address agent: a
// signed-in dashboard user may address any agent, other callers only one that
// enabled AG-UI.
func aguiReachable(ctx context.Context, agent *agentsv1.Agent) bool {
	if _, signedIn := auth.UserFromContext(ctx); signedIn {
		return true
	}
	return agent.GetEnableAgui()
}

// validateAndPrepare resolves the workspace, agent and runner, decodes and
// validates the RunAgentInput, and builds the run's parts and ContextInfo.
// It returns false when an error response was already written.
func (h *AGUIHandler) validateAndPrepare(c *gin.Context) (*aguiRunContext, bool) {
	ctx := c.Request.Context()

	target, ok := h.resolveAgent(c)
	if !ok {
		return nil, false
	}

	var input aguitypes.RunAgentInput
	if status, err := bindAGUIInput(c, &input); err != nil {
		c.JSON(status, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}
	if err := validateAGUIInput(&input); err != nil {
		c.JSON(http.StatusBadRequest, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}
	if input.RunID == "" {
		input.RunID = uuid.NewString()
	}
	negotiated, err := a2ui.Negotiate(input.ForwardedProps)
	if err != nil {
		c.JSON(http.StatusBadRequest, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}
	detach, err := negotiateDetach(input.ForwardedProps)
	if err != nil {
		c.JSON(http.StatusBadRequest, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}

	ctxInfo, err := streamorch.NewContextInfo(streamorch.ContextInfoInput{
		AppName:       aguiAppName,
		UserID:        aguiUserID(ctx),
		SessionID:     aguiSessionPrefix + input.ThreadID,
		SessionPrefix: aguiSessionPrefix,
		WorkspaceID:   target.workspaceID,
		HasWorkspace:  true,
		IsAdmin:       auth.IsAdmin(ctx),
		Source:        agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:      agentsv1.ChatType_CHAT_TYPE_PRIVATE,
	})
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}

	parts, formEntries, toolResults, err := aguiInputParts(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}

	rc := &aguiRunContext{
		input:          input,
		agent:          target.agent,
		agentName:      target.name,
		svc:            target.runner,
		parts:          parts,
		ctxInfo:        ctxInfo,
		clientTools:    clientToolDeclarations(input.Tools),
		a2uiNegotiated: negotiated,
		formEntries:    formEntries,
		toolResults:    toolResults,
		detach:         detach,
	}
	if refusal := h.checkStores(rc); refusal != nil {
		refusal.write(c)
		return nil, false
	}
	return rc, true
}

// checkStores refuses, before the lease, what the handler cannot serve
// without the stores it depends on, instead of leaving a client believing a
// capability took effect.
func (h *AGUIHandler) checkStores(rc *aguiRunContext) *aguiRefusal {
	unavailable := func(msg string) *aguiRefusal {
		return refuseAGUI(http.StatusServiceUnavailable, errors.New(msg))
	}
	hasSessions := h.getSessionService() != nil
	if !hasSessions {
		if clientState, _ := rc.input.State.(map[string]any); len(clientState) > 0 {
			return unavailable("shared state is not accepted: session service unavailable")
		}
		if len(rc.toolResults) > 0 {
			return unavailable("tool results are not accepted: session service unavailable")
		}
	}
	if rc.detach {
		if !hasSessions {
			return unavailable("detached runs are not available: session service unavailable")
		}
		if h.getInvocations() == nil {
			return unavailable("detached runs are not available: invocation store unavailable")
		}
		// A Stop reaches a Detached Run through its run state.
		if h.getRunStateStore() == nil {
			return unavailable("detached runs are not available: run state store unavailable")
		}
	}
	return nil
}

// prepareSharedState resolves the run's shared-state baseline from the
// session as the run starts from it, read under the lease.
//
// Ownership model: the server-side session owns state; the client holds a
// mirror. The client's RunAgentInput.State is used for validation only —
// when it is absent or diverges from the authoritative state, the run opens
// with a corrective STATE_SNAPSHOT rather than adopting client changes, so a
// client edit is answered visibly instead of being silently kept or dropped.
// Without a session store the feature is off (checkStores refuses a client
// state then).
func (rc *aguiRunContext) prepareSharedState(sess session.Session) {
	clientState, _ := rc.input.State.(map[string]any)
	authoritative := aguiVisibleState(sessionStateMap(sess))
	rc.stateArmed = true
	rc.stateInitial = authoritative
	if len(clientState) == 0 {
		// No mirror yet: baseline the client when there is anything to see.
		rc.stateSnapshot = len(authoritative) > 0
		return
	}
	rc.stateSnapshot = !aguiStatesEqual(authoritative, aguiVisibleState(clientState))
}

// validateAGUIInput enforces the endpoint's contract. Everything the endpoint
// does not implement is rejected rather than silently ignored, so a client
// never believes a capability took effect when it did not.
func validateAGUIInput(input *aguitypes.RunAgentInput) error {
	if strings.TrimSpace(input.ThreadID) == "" {
		return errors.New("threadId is required")
	}
	seenTools := make(map[string]bool, len(input.Tools))
	for _, t := range input.Tools {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			return errors.New("tools require a name")
		}
		if seenTools[name] {
			return errors.New("duplicate tool name: " + name)
		}
		seenTools[name] = true
	}
	switch input.State.(type) {
	case nil, map[string]any:
	default:
		return errors.New("state must be a JSON object")
	}
	seenResume := make(map[string]bool, len(input.Resume))
	for _, entry := range input.Resume {
		if entry.Status == aguitypes.ResumeStatusCancelled {
			// Butter has no way to abandon a pending Interrupt: the workflow
			// would stay paused forever with no path back. Failing loudly beats
			// treating a cancellation as an answer.
			return errors.New("cancelling an interrupt is not supported")
		}
		if entry.InterruptID == "" {
			return errors.New("resume entries require an interruptId")
		}
		if seenResume[entry.InterruptID] {
			return errors.New("duplicate resume entry for interruptId: " + entry.InterruptID)
		}
		seenResume[entry.InterruptID] = true
	}
	return nil
}

// clientToolDeclarations converts the request's AG-UI tool declarations into
// the run-scoped form the aguitool toolset resolves.
func clientToolDeclarations(tools []aguitypes.Tool) []aguitool.Declaration {
	if len(tools) == 0 {
		return nil
	}
	decls := make([]aguitool.Declaration, 0, len(tools))
	for _, t := range tools {
		decls = append(decls, aguitool.Declaration{
			Name:        strings.TrimSpace(t.Name),
			Description: t.Description,
			Parameters:  t.Parameters,
		})
	}
	return decls
}

// aguiInputParts builds the run's input parts. Any error is a 400.
//
// A resume request answers pending Interrupts by ID. Because
// interrupt.Resume passes parts through untouched once they carry a
// FunctionResponse, the client's explicit addressing takes precedence over
// butter's implicit oldest-first resume (ADR-0002). Trailing tool-role
// messages answer pending frontend tool calls the same way and may be
// combined with resume entries in one request. They are returned as
// toolResults: they become parts only once validated against the session's
// pending calls under the lease (toolResultParts), after the resume parts.
//
// Otherwise only the trailing user message is sent, its text and images
// (aguiUserParts): RunAgentInput.Messages is the client's full history, but
// the server-side session is authoritative, so replaying it would duplicate
// the conversation.
//
// A resume entry whose payload is an A2UI form submission (butterForm) is
// returned as a form entry: its answer is only known once the submission is
// validated against the session under the lease (resolveFormSubmissions).
func aguiInputParts(input *aguitypes.RunAgentInput) ([]*genai.Part, []aguiFormEntry, []aguitypes.Message, error) {
	var parts []*genai.Part
	var forms []aguiFormEntry
	for _, entry := range input.Resume {
		payload := entry.Payload
		sub, isForm, err := a2ui.SubmissionOf(entry.Payload)
		if err != nil {
			return nil, nil, nil, err
		}
		if isForm {
			forms = append(forms, aguiFormEntry{partIndex: len(parts), interruptID: entry.InterruptID, submission: sub})
			payload = nil
		}
		parts = append(parts, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				ID:   entry.InterruptID,
				Name: workflow.WorkflowInputFunctionCallName,
				Response: map[string]any{
					aguiRequestInputPayloadKey: payload,
				},
			},
		})
	}

	results, err := trailingToolResults(input.Messages)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(parts) > 0 || len(results) > 0 {
		return parts, forms, results, nil
	}

	userParts, err := aguiUserParts(input.Messages)
	if err != nil {
		return nil, nil, nil, err
	}
	return userParts, nil, nil, nil
}

// trailingToolResults returns the tool-role messages that terminate the
// client's history — the results of frontend tool calls the previous run
// ended on. Earlier tool messages are replayed history of already-answered
// calls; the authoritative record of those lives in the server-side session,
// so they are ignored.
func trailingToolResults(messages []aguitypes.Message) ([]aguitypes.Message, error) {
	var reversed []aguitypes.Message
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != aguitypes.RoleTool {
			break
		}
		reversed = append(reversed, messages[i])
	}
	if len(reversed) == 0 {
		return nil, nil
	}
	results := make([]aguitypes.Message, 0, len(reversed))
	seen := make(map[string]bool, len(reversed))
	for i := len(reversed) - 1; i >= 0; i-- {
		msg := reversed[i]
		if msg.ToolCallID == "" {
			return nil, errors.New("tool result messages require a toolCallId")
		}
		if seen[msg.ToolCallID] {
			return nil, errors.New("duplicate tool result for toolCallId: " + msg.ToolCallID)
		}
		seen[msg.ToolCallID] = true
		if _, ok := msg.ContentString(); !ok && msg.Error == "" {
			return nil, errors.New("tool result content must be a string (toolCallId " + msg.ToolCallID + ")")
		}
		results = append(results, msg)
	}
	return results, nil
}

// toolResultParts validates client tool results against the pending calls
// recorded on the thread's session, read under the lease, and converts them
// into the FunctionResponse parts ADK resumes on. Pairing is by FunctionCall
// ID; the name comes from the session's own record of the call, never from
// the client. A nil session (a thread with none yet) waits on nothing. Any
// error is a 400.
func toolResultParts(sess session.Session, results []aguitypes.Message) ([]*genai.Part, error) {
	if sess == nil {
		return nil, errors.New("no pending tool calls for this thread")
	}
	pending := interrupt.PendingToolCalls(sess)
	names := make(map[string]string, len(pending))
	for _, call := range pending {
		names[call.ID] = call.Name
	}

	parts := make([]*genai.Part, 0, len(results))
	for _, msg := range results {
		name, ok := names[msg.ToolCallID]
		if !ok {
			return nil, errors.New("unknown or already answered toolCallId: " + msg.ToolCallID)
		}
		parts = append(parts, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				ID:       msg.ToolCallID,
				Name:     name,
				Response: toolResultResponse(msg),
			},
		})
	}
	return parts, nil
}

// toolResultResponse shapes one client tool result as the FunctionResponse
// payload the model reads. A JSON-object result passes through; anything else
// is wrapped, and a client-reported error is surfaced as one so the model
// knows the tool failed rather than returned prose.
func toolResultResponse(msg aguitypes.Message) map[string]any {
	if msg.Error != "" {
		return map[string]any{"error": msg.Error}
	}
	content, _ := msg.ContentString()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(content), &decoded); err == nil && decoded != nil {
		return decoded
	}
	return map[string]any{"result": content}
}

// aguiUserID derives the ADK session's user ID from the authenticated caller.
// It is part of the session key, so two users sharing a threadId still get
// separate histories.
func aguiUserID(ctx context.Context) string {
	if user, ok := auth.UserFromContext(ctx); ok {
		if id := user.GetId(); id != "" {
			return id
		}
		if name := user.GetUsername(); name != "" {
			return name
		}
	}
	return "agui-user"
}
