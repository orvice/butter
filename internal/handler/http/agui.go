package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"butterfly.orx.me/core/log"
	aguitypes "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
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

	mu        sync.RWMutex
	runnerSvc AGUIRunnerService
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
	input   aguitypes.RunAgentInput
	agent   *agentsv1.Agent
	svc     AGUIRunnerService
	parts   []*genai.Part
	ctxInfo *agentsv1.ContextInfo
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

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	sink := newAGUISink(rc.input.ThreadID, rc.input.RunID, uuid.NewString(),
		newAGUISSEEmitter(c.Request.Context(), c.Writer, c.Writer.Flush))

	runErr := streamorch.Run(c.Request.Context(), rc.svc,
		streamorch.AgentRef{Name: rc.agent.GetName(), ID: rc.agent.GetAgentId()},
		rc.parts, "", rc.ctxInfo, sink)
	if runErr == nil {
		return
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

	parts, err := aguiInputParts(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, aguiErrorResponse{Error: err.Error()})
		return nil, false
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

	return &aguiRunContext{input: input, agent: agent, svc: svc, parts: parts, ctxInfo: ctxInfo}, true
}

// validateAGUIInput enforces the Phase 1 contract. Everything the endpoint does
// not implement yet is rejected rather than silently ignored, so a client never
// believes a capability took effect when it did not.
func validateAGUIInput(input *aguitypes.RunAgentInput) error {
	if strings.TrimSpace(input.ThreadID) == "" {
		return errors.New("threadId is required")
	}
	if len(input.Tools) > 0 {
		return errors.New("client-supplied tools are not supported yet")
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

// aguiInputParts builds the run's input parts.
//
// A resume request answers pending Interrupts by ID. Because
// interrupt.Resume passes parts through untouched once they carry a
// FunctionResponse, the client's explicit addressing takes precedence over
// butter's implicit oldest-first resume (ADR-0002).
//
// Otherwise only the trailing user message is sent: RunAgentInput.Messages is
// the client's full history, but the server-side session is authoritative, so
// replaying it would duplicate the conversation.
func aguiInputParts(input *aguitypes.RunAgentInput) ([]*genai.Part, error) {
	if len(input.Resume) > 0 {
		parts := make([]*genai.Part, 0, len(input.Resume))
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
		return parts, nil
	}

	text := latestAGUIUserText(input.Messages)
	if text == "" {
		return nil, errors.New("messages must end with a non-empty user message")
	}
	return []*genai.Part{{Text: text}}, nil
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
