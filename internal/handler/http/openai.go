package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// OpenAIRunnerService defines the subset of runner.Service the OpenAI handler needs.
type OpenAIRunnerService interface {
	Run(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
	RunSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
}

// OpenAIHandler serves OpenAI-compatible API endpoints for enabled agents.
type OpenAIHandler struct {
	agentRepo configrepo.AgentRepository

	mu        sync.RWMutex
	runnerSvc OpenAIRunnerService
}

// NewOpenAIHandler creates an OpenAI handler with the given agent repository.
func NewOpenAIHandler(repo configrepo.AgentRepository) *OpenAIHandler {
	return &OpenAIHandler{agentRepo: repo}
}

// SetRunnerService sets the runner service after bootstrap completes.
func (h *OpenAIHandler) SetRunnerService(svc OpenAIRunnerService) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.runnerSvc = svc
}

func (h *OpenAIHandler) getRunner() OpenAIRunnerService {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.runnerSvc
}

// Register registers OpenAI-compatible routes on the Gin engine.
func (h *OpenAIHandler) Register(r *gin.Engine) {
	r.GET("/api/v1/models", h.ListModels)
	r.POST("/api/v1/chat/completions", h.ChatCompletions)
}

// chatCompletionRequest is the OpenAI-compatible request body.
type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatCompletionResponse is the OpenAI ChatCompletion response format.
type chatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []chatCompletionChoice `json:"choices"`
	Usage   chatCompletionUsage    `json:"usage"`
}

type chatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type chatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// chatRunContext holds the validated and prepared data for a chat completions request.
type chatRunContext struct {
	workspaceID string
	req         chatCompletionRequest
	agent       *agentsv1.Agent
	svc         OpenAIRunnerService
	parts       []*genai.Part
	ctxInfo     *agentsv1.ContextInfo
}

// ChatCompletions handles POST /api/v1/chat/completions.
func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	rc, ok := h.validateAndPrepare(c)
	if !ok {
		return
	}

	if rc.req.Stream {
		h.handleStream(c, rc)
		return
	}

	h.handleNonStream(c, rc)
}

// validateAndPrepare parses the request, resolves the agent, checks the runner,
// and builds the run context. Returns false if an error response was already sent.
func (h *OpenAIHandler) validateAndPrepare(c *gin.Context) (*chatRunContext, bool) {
	workspaceID, ok := wsctx.FromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, openaiError("unauthorized", "invalid_request_error"))
		return nil, false
	}

	var req chatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, openaiError("invalid request body", "invalid_request_error"))
		return nil, false
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, openaiError("messages must not be empty", "invalid_request_error"))
		return nil, false
	}

	agent, err := h.agentRepo.GetAgent(c.Request.Context(), workspaceID, req.Model)
	if err != nil || agent == nil || !agent.GetEnableOpenaiApi() {
		c.JSON(http.StatusNotFound, openaiError("model not found: "+req.Model, "invalid_request_error"))
		return nil, false
	}

	svc := h.getRunner()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, openaiError("runner not available", "server_error"))
		return nil, false
	}

	invocationID := uuid.NewString()
	parts := []*genai.Part{{Text: formatMessages(req.Messages)}}
	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        invocationID,
		ChannelName: "openai-api",
		SessionId:   "openai-" + uuid.NewString(),
		UserId:      "openai-user",
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		WorkspaceId: workspaceID,
	}

	return &chatRunContext{
		workspaceID: workspaceID,
		req:         req,
		agent:       agent,
		svc:         svc,
		parts:       parts,
		ctxInfo:     ctxInfo,
	}, true
}

func (h *OpenAIHandler) handleNonStream(c *gin.Context, rc *chatRunContext) {
	output, err := rc.svc.Run(c.Request.Context(), rc.agent.GetName(), rc.parts, "", rc.ctxInfo, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, openaiError(err.Error(), "server_error"))
		return
	}

	resp := chatCompletionResponse{
		ID:     "chatcmpl-" + uuid.NewString(),
		Object: "chat.completion",
		Model:  rc.req.Model,
		Choices: []chatCompletionChoice{{
			Index:        0,
			Message:      chatMessage{Role: "assistant", Content: output},
			FinishReason: "stop",
		}},
		Usage: chatCompletionUsage{}, // TODO: future token tracking
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OpenAIHandler) handleStream(c *gin.Context, rc *chatRunContext) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)

	completionID := "chatcmpl-" + uuid.NewString()
	isFirst := true
	wroteContent := false

	writeContentChunk := func(content string) {
		if content == "" {
			return
		}
		chunk := streamChunk{
			ID:     completionID,
			Object: "chat.completion.chunk",
			Model:  rc.req.Model,
			Choices: []streamChunkChoice{{
				Index: 0,
				Delta: streamDelta{Content: content},
			}},
		}
		if isFirst {
			chunk.Choices[0].Delta.Role = "assistant"
			isFirst = false
		}
		data, _ := json.Marshal(chunk)
		c.Writer.WriteString("data: " + string(data) + "\n\n")
		c.Writer.Flush()
		wroteContent = true
	}

	onEvent := func(evt *session.Event) {
		if evt == nil || !evt.Partial || evt.Content == nil {
			return
		}
		for _, part := range evt.Content.Parts {
			if part == nil || part.Text == "" {
				continue
			}
			writeContentChunk(part.Text)
		}
	}

	output, err := rc.svc.RunSSE(c.Request.Context(), rc.agent.GetName(), rc.parts, "", rc.ctxInfo, onEvent, nil)
	if err != nil {
		// If streaming hasn't started (no chunks sent), we could return JSON error,
		// but headers are already committed. Write an SSE error event instead.
		errChunk := streamChunk{
			ID:     completionID,
			Object: "chat.completion.chunk",
			Model:  rc.req.Model,
			Choices: []streamChunkChoice{{
				Index:        0,
				Delta:        streamDelta{},
				FinishReason: "error",
			}},
		}
		data, _ := json.Marshal(errChunk)
		c.Writer.WriteString("data: " + string(data) + "\n\n")
		c.Writer.Flush()
	} else if !wroteContent {
		writeContentChunk(output)
	}

	c.Writer.WriteString("data: [DONE]\n\n")
	c.Writer.Flush()
}

// streamChunk is the SSE chunk format for streaming responses.
type streamChunk struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created,omitempty"`
	Model   string              `json:"model"`
	Choices []streamChunkChoice `json:"choices"`
}

type streamChunkChoice struct {
	Index        int         `json:"index"`
	Delta        streamDelta `json:"delta"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

type streamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// formatMessages converts the messages array into structured text.
func formatMessages(messages []chatMessage) string {
	var b strings.Builder
	for i, msg := range messages {
		if i > 0 {
			b.WriteByte('\n')
		}
		switch msg.Role {
		case "system":
			b.WriteString("[System] ")
		case "user":
			b.WriteString("[User] ")
		case "assistant":
			b.WriteString("[Assistant] ")
		default:
			b.WriteString("[" + msg.Role + "] ")
		}
		b.WriteString(msg.Content)
	}
	return b.String()
}

// ListModels returns an OpenAI-format model list of agents with enable_openai_api=true.
func (h *OpenAIHandler) ListModels(c *gin.Context) {
	workspaceID, ok := wsctx.FromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, openaiError("unauthorized", "invalid_request_error"))
		return
	}

	agents, err := h.agentRepo.ListAgents(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, openaiError(err.Error(), "server_error"))
		return
	}

	models := make([]openaiModel, 0, len(agents))
	for _, ag := range agents {
		if ag.GetEnableOpenaiApi() {
			models = append(models, openaiModel{
				ID:      ag.GetName(),
				Object:  "model",
				Created: 0,
				OwnedBy: "butter",
			})
		}
	}

	c.JSON(http.StatusOK, openaiModelList{
		Object: "list",
		Data:   models,
	})
}

// --- OpenAI response types ---

type openaiModelList struct {
	Object string        `json:"object"`
	Data   []openaiModel `json:"data"`
}

type openaiModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type openaiErrorResponse struct {
	Error openaiErrorBody `json:"error"`
}

type openaiErrorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func openaiError(message, errType string) openaiErrorResponse {
	return openaiErrorResponse{
		Error: openaiErrorBody{
			Message: message,
			Type:    errType,
		},
	}
}
