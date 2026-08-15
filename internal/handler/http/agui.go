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

	"go.orx.me/apps/butter/internal/aguitool"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/interrupt"
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
const AGUISessionLeaseTTL = 5 * time.Minute

// AGUIRunnerService is the subset of runner.Service the AG-UI handler needs.
// It matches streamorch.Runner so the orchestrator can be driven directly.
type AGUIRunnerService = streamorch.Runner

// AGUIHandler serves the AG-UI protocol endpoint for agents that opted in via
// enable_agui.
//
// The endpoint is stateful in AG-UI terms: the server-side session is
// authoritative, so the client's message history is not replayed into the
// runner. See docs/api.md and docs/research/ag-ui-integration.md.
type AGUIHandler struct {
	agentRepo configrepo.AgentRepository

	mu           sync.RWMutex
	runnerSvc    AGUIRunnerService
	sessionGuard sessionguard.Guard
	sessionSvc   session.Service
}

// NewAGUIHandler creates an AG-UI handler with the given agent repository.
func NewAGUIHandler(repo configrepo.AgentRepository) *AGUIHandler {
	return &AGUIHandler{agentRepo: repo}
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
}

// aguiErrorResponse is the body for failures that happen before the SSE stream
// opens. Once streaming has started, errors are RUN_ERROR events instead.
type aguiErrorResponse struct {
	Error string `json:"error"`
}

// aguiRunContext is the validated, prepared state for one AG-UI run.
type aguiRunContext struct {
	input       aguitypes.RunAgentInput
	agent       *agentsv1.Agent
	svc         AGUIRunnerService
	parts       []*genai.Part
	ctxInfo     *agentsv1.ContextInfo
	clientTools []aguitool.Declaration
}

// RunAgent handles POST /api/agui/:agent_id.
func (h *AGUIHandler) RunAgent(c *gin.Context) {
	rc, ok := h.validateAndPrepare(c)
	if !ok {
		return
	}

	logger := log.FromContext(c.Request.Context())
	logger.Info("agui run started",
		"workspace_id", rc.ctxInfo.GetWorkspaceId(),
		"agent", rc.agent.GetName(),
		"agent_id", rc.agent.GetAgentId(),
		"thread_id", rc.input.ThreadID,
		"run_id", rc.input.RunID,
		"session_id", rc.ctxInfo.GetSessionId(),
		"resume_entries", len(rc.input.Resume),
	)

	// One turn per (caller, thread) at a time across the whole fleet. The
	// lease is taken before the stream opens so a busy thread is an HTTP
	// error the client can retry, not a stream that dies mid-run. Client
	// disconnect cancels the request context, which ends the run and releases
	// the lease through the same defer.
	runCtx := c.Request.Context()
	if guard := h.getSessionGuard(); guard != nil {
		leaseCtx, release, acquired, err := guard.Acquire(runCtx, aguiSessionKey(rc.ctxInfo))
		if err != nil {
			logger.Error("agui session lease unavailable",
				"thread_id", rc.input.ThreadID, "err", err)
			c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "session lock unavailable, retry later"})
			return
		}
		if !acquired {
			c.JSON(http.StatusConflict, aguiErrorResponse{Error: "a run is already in progress for this thread, retry after it finishes"})
			return
		}
		defer release()
		runCtx = leaseCtx
	}

	// Client-declared frontend tools ride the run context: the aguitool
	// toolset resolves them per invocation, so the agent sees them for this
	// run only.
	if len(rc.clientTools) > 0 {
		runCtx = aguitool.WithClientTools(runCtx, rc.clientTools)
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	sink := newAGUISink(rc.input.ThreadID, rc.input.RunID, uuid.NewString(),
		newAGUISSEEmitter(c.Request.Context(), c.Writer, c.Writer.Flush))

	runErr := streamorch.Run(runCtx, rc.svc,
		streamorch.AgentRef{Name: rc.agent.GetName(), ID: rc.agent.GetAgentId()},
		rc.parts, "", rc.ctxInfo, sink)
	if runErr == nil {
		return
	}
	// A cancelled lease context with a live request means the lease was lost
	// (fenced out by expiry or takeover); name it instead of reporting a bare
	// context cancellation.
	if errors.Is(runErr, context.Canceled) && c.Request.Context().Err() == nil {
		runErr = errors.New("session lease lost, the run was cancelled")
	}

	logger.Error("agui run failed",
		"workspace_id", rc.ctxInfo.GetWorkspaceId(),
		"agent_id", rc.agent.GetAgentId(),
		"thread_id", rc.input.ThreadID,
		"err", runErr,
	)
	// The status and headers are already committed, so the failure is reported
	// in-band as RUN_ERROR rather than as an HTTP error.
	if err := sink.Error(runErr); err != nil {
		logger.Error("agui failed to emit RUN_ERROR", "err", err)
	}
}

// validateAndPrepare resolves the workspace, agent and runner, decodes and
// validates the RunAgentInput, and builds the run's parts and ContextInfo.
// It returns false when an error response was already written.
func (h *AGUIHandler) validateAndPrepare(c *gin.Context) (*aguiRunContext, bool) {
	ctx := c.Request.Context()

	workspaceID, hasWorkspace := wsctx.FromContext(ctx)
	if !hasWorkspace {
		c.JSON(http.StatusUnauthorized, aguiErrorResponse{Error: "workspace required (set X-Workspace-ID header)"})
		return nil, false
	}

	agentID := c.Param("agent_id")
	agent, err := h.agentRepo.GetAgent(ctx, workspaceID, agentID)
	if err != nil || agent == nil || !agent.GetEnableAgui() {
		c.JSON(http.StatusNotFound, aguiErrorResponse{Error: "agent not found: " + agentID})
		return nil, false
	}

	svc := h.getRunner()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "runner not available"})
		return nil, false
	}

	var input aguitypes.RunAgentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, aguiErrorResponse{Error: "invalid request body"})
		return nil, false
	}
	if err := validateAGUIInput(&input); err != nil {
		c.JSON(http.StatusBadRequest, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}
	if input.RunID == "" {
		input.RunID = uuid.NewString()
	}

	ctxInfo, err := streamorch.NewContextInfo(streamorch.ContextInfoInput{
		AppName:       aguiAppName,
		UserID:        aguiUserID(ctx),
		SessionID:     aguiSessionPrefix + input.ThreadID,
		SessionPrefix: aguiSessionPrefix,
		WorkspaceID:   workspaceID,
		HasWorkspace:  hasWorkspace,
		IsAdmin:       auth.IsAdmin(ctx),
		Source:        agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:      agentsv1.ChatType_CHAT_TYPE_PRIVATE,
	})
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}

	parts, status, err := h.aguiInputParts(ctx, &input, ctxInfo)
	if err != nil {
		c.JSON(status, aguiErrorResponse{Error: err.Error()})
		return nil, false
	}

	return &aguiRunContext{
		input:       input,
		agent:       agent,
		svc:         svc,
		parts:       parts,
		ctxInfo:     ctxInfo,
		clientTools: clientToolDeclarations(input.Tools),
	}, true
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
	if !aguiStateIsEmpty(input.State) {
		return errors.New("shared state is not supported yet")
	}
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

// aguiStateIsEmpty reports whether the request carries no shared state. An
// absent field, JSON null, and an empty object all mean "none".
func aguiStateIsEmpty(state any) bool {
	switch v := state.(type) {
	case nil:
		return true
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

// aguiInputParts builds the run's input parts and, on failure, the HTTP
// status the rejection maps to.
//
// A resume request answers pending Interrupts by ID. Because
// interrupt.Resume passes parts through untouched once they carry a
// FunctionResponse, the client's explicit addressing takes precedence over
// butter's implicit oldest-first resume (ADR-0002). Trailing tool-role
// messages answer pending frontend tool calls the same way and may be
// combined with resume entries in one request.
//
// Otherwise only the trailing user message is sent: RunAgentInput.Messages is
// the client's full history, but the server-side session is authoritative, so
// replaying it would duplicate the conversation.
func (h *AGUIHandler) aguiInputParts(ctx context.Context, input *aguitypes.RunAgentInput, ctxInfo *agentsv1.ContextInfo) ([]*genai.Part, int, error) {
	var parts []*genai.Part
	for _, entry := range input.Resume {
		parts = append(parts, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				ID:   entry.InterruptID,
				Name: workflow.WorkflowInputFunctionCallName,
				Response: map[string]any{
					aguiRequestInputPayloadKey: entry.Payload,
				},
			},
		})
	}

	results, err := trailingToolResults(input.Messages)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if len(results) > 0 {
		toolParts, status, err := h.toolResultParts(ctx, results, ctxInfo)
		if err != nil {
			return nil, status, err
		}
		parts = append(parts, toolParts...)
	}
	if len(parts) > 0 {
		return parts, 0, nil
	}

	text := latestAGUIUserText(input.Messages)
	if text == "" {
		return nil, http.StatusBadRequest, errors.New("messages must end with a non-empty user message")
	}
	return []*genai.Part{{Text: text}}, 0, nil
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
// recorded on the session and converts them into the FunctionResponse parts
// ADK resumes on. Pairing is by FunctionCall ID; the name comes from the
// session's own record of the call, never from the client.
func (h *AGUIHandler) toolResultParts(ctx context.Context, results []aguitypes.Message, ctxInfo *agentsv1.ContextInfo) ([]*genai.Part, int, error) {
	svc := h.getSessionService()
	if svc == nil {
		return nil, http.StatusServiceUnavailable, errors.New("tool results are not accepted: session service unavailable")
	}
	resp, err := svc.Get(ctx, &session.GetRequest{
		AppName:   aguiAppName,
		UserID:    ctxInfo.GetUserId(),
		SessionID: ctxInfo.GetSessionId(),
	})
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("no pending tool calls for this thread")
	}
	pending := interrupt.PendingToolCalls(resp.Session)
	names := make(map[string]string, len(pending))
	for _, call := range pending {
		names[call.ID] = call.Name
	}

	parts := make([]*genai.Part, 0, len(results))
	for _, msg := range results {
		name, ok := names[msg.ToolCallID]
		if !ok {
			return nil, http.StatusBadRequest, errors.New("unknown or already answered toolCallId: " + msg.ToolCallID)
		}
		parts = append(parts, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				ID:       msg.ToolCallID,
				Name:     name,
				Response: toolResultResponse(msg),
			},
		})
	}
	return parts, 0, nil
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

// latestAGUIUserText returns the text of the last user message. Content arrives
// either as a plain string or as multimodal fragments; Phase 1 is text-only, so
// non-text fragments are skipped.
func latestAGUIUserText(messages []aguitypes.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != aguitypes.RoleUser {
			continue
		}
		if s, ok := msg.ContentString(); ok {
			return strings.TrimSpace(s)
		}
		if fragments, ok := msg.ContentInputContents(); ok {
			var b strings.Builder
			for _, fragment := range fragments {
				if fragment.Text == "" {
					continue
				}
				if b.Len() > 0 {
					b.WriteByte('\n')
				}
				b.WriteString(fragment.Text)
			}
			return strings.TrimSpace(b.String())
		}
		return ""
	}
	return ""
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
