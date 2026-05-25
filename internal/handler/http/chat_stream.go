package http

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/runtime/runner"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ChatStreamHandler serves streaming chat responses over Server-Sent Events.
type ChatStreamHandler struct {
	mu        sync.RWMutex
	runnerSvc *runner.Service
}

func NewChatStreamHandler() *ChatStreamHandler {
	return &ChatStreamHandler{}
}

// SetRunnerService sets the runner service after bootstrap completes.
func (h *ChatStreamHandler) SetRunnerService(svc *runner.Service) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.runnerSvc = svc
}

func (h *ChatStreamHandler) getRunner() *runner.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.runnerSvc
}

// Register registers chat streaming routes on the Gin engine.
func (h *ChatStreamHandler) Register(r *gin.Engine) {
	r.POST("/api/chat/stream", h.Stream)
}

type ChatStreamRequest struct {
	AgentName     string `json:"agent_name"`
	Message       string `json:"message"`
	AppName       string `json:"app_name"`
	UserID        string `json:"user_id"`
	SessionID     string `json:"session_id"`
	ModelOverride string `json:"model_override"`
}

type chatStreamPayload struct {
	InvocationID string              `json:"invocation_id,omitempty"`
	SessionID    string              `json:"session_id,omitempty"`
	AgentName    string              `json:"agent_name,omitempty"`
	Response     string              `json:"response,omitempty"`
	TextDelta    string              `json:"text_delta,omitempty"`
	Error        string              `json:"error,omitempty"`
	Event        *chatStreamRunEvent `json:"event,omitempty"`
}

type chatStreamRunEvent struct {
	EventID       string `json:"event_id,omitempty"`
	InvocationID  string `json:"invocation_id,omitempty"`
	Author        string `json:"author,omitempty"`
	Branch        string `json:"branch,omitempty"`
	Partial       bool   `json:"partial,omitempty"`
	FinalResponse bool   `json:"final_response,omitempty"`
	ContentJSON   string `json:"content_json,omitempty"`
	Timestamp     string `json:"timestamp,omitempty"`
}

type chatStreamMessage struct {
	Event string
	Data  chatStreamPayload
}

// Stream accepts a chat message and streams runner events back to the client as SSE.
func (h *ChatStreamHandler) Stream(c *gin.Context) {
	svc := h.getRunner()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "runner service not available"})
		return
	}

	var req ChatStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if req.AgentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_name is required"})
		return
	}
	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	appName := req.AppName
	if appName == "" {
		appName = "api"
	}
	userID := req.UserID
	if userID == "" {
		userID = "api"
	}
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "chat-" + uuid.NewString()
	}
	invocationID := uuid.NewString()
	if id, err := uuid.NewV7(); err == nil {
		invocationID = id.String()
	}

	reqCtx := c.Request.Context()
	workspaceID, hasWorkspace := wsctx.FromContext(reqCtx)
	if !hasWorkspace && !auth.IsAdmin(reqCtx) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace required (set X-Workspace-ID header)"})
		return
	}
	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        invocationID,
		SessionId:   sessionID,
		UserId:      userID,
		ChannelName: appName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		WorkspaceId: workspaceID,
	}

	logger := log.FromContext(reqCtx)
	logger.Info("streaming chat started",
		"workspace_id", workspaceID,
		"agent", req.AgentName,
		"session_id", sessionID,
		"invocation_id", invocationID,
		"message_len", len(req.Message),
	)

	w := c.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	messages := make(chan chatStreamMessage, 32)

	send := func(event string, payload chatStreamPayload) bool {
		select {
		case messages <- chatStreamMessage{Event: event, Data: payload}:
			return true
		case <-reqCtx.Done():
			return false
		}
	}

	go func() {
		defer close(messages)
		parts := []*genai.Part{genai.NewPartFromText(req.Message)}
		response, err := svc.RunSSE(reqCtx, req.AgentName, parts, req.ModelOverride, ctxInfo, func(evt *session.Event) {
			textParts := eventTextParts(evt)
			for _, text := range textParts {
				_ = send("text_delta", chatStreamPayload{
					InvocationID: invocationID,
					SessionID:    sessionID,
					AgentName:    req.AgentName,
					TextDelta:    text,
				})
			}
			if len(textParts) > 0 && eventHasOnlyTextParts(evt) {
				return
			}
			_ = send("agent_event", chatStreamPayload{
				InvocationID: invocationID,
				SessionID:    sessionID,
				AgentName:    req.AgentName,
				Event:        eventToChatStreamRunEvent(evt),
			})
		}, nil)
		if err != nil {
			_ = send("error", chatStreamPayload{
				InvocationID: invocationID,
				SessionID:    sessionID,
				AgentName:    req.AgentName,
				Error:        err.Error(),
			})
			return
		}
		_ = send("final", chatStreamPayload{
			InvocationID: invocationID,
			SessionID:    sessionID,
			AgentName:    req.AgentName,
			Response:     response,
		})
	}()

	c.SSEvent("invocation_started", chatStreamPayload{
		InvocationID: invocationID,
		SessionID:    sessionID,
		AgentName:    req.AgentName,
	})
	w.Flush()

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				return
			}
			c.SSEvent(msg.Event, msg.Data)
			w.Flush()
		case <-reqCtx.Done():
			logger.Info("streaming chat client disconnected", "invocation_id", invocationID)
			return
		}
	}
}

func eventToChatStreamRunEvent(evt *session.Event) *chatStreamRunEvent {
	if evt == nil {
		return nil
	}
	out := &chatStreamRunEvent{
		EventID:       evt.ID,
		InvocationID:  evt.InvocationID,
		Author:        evt.Author,
		Branch:        evt.Branch,
		Partial:       evt.Partial,
		FinalResponse: evt.IsFinalResponse(),
		Timestamp:     evt.Timestamp.UTC().Format(time.RFC3339Nano),
	}
	if evt.Content != nil {
		if data, err := json.Marshal(evt.Content); err == nil {
			out.ContentJSON = string(data)
		}
	}
	return out
}

func eventHasOnlyTextParts(evt *session.Event) bool {
	if evt == nil || evt.Content == nil || len(evt.Content.Parts) == 0 {
		return false
	}
	for _, part := range evt.Content.Parts {
		if part == nil {
			continue
		}
		if part.Text == "" && part.Thought == false {
			return false
		}
		if part.FunctionCall != nil || part.FunctionResponse != nil || part.CodeExecutionResult != nil || part.ExecutableCode != nil || part.InlineData != nil || part.FileData != nil {
			return false
		}
	}
	return true
}

func eventTextParts(evt *session.Event) []string {
	if evt == nil || !evt.Partial || evt.Content == nil {
		return nil
	}
	out := make([]string, 0, len(evt.Content.Parts))
	for _, part := range evt.Content.Parts {
		if part == nil || part.Text == "" || part.Thought {
			continue
		}
		out = append(out, part.Text)
	}
	return out
}
