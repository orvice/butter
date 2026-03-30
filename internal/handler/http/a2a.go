package http

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// A2AHandler serves the A2A protocol endpoints for enabled agents.
type A2AHandler struct {
	cfg *config.AppConfig // read agents at request time (populated by Butterfly after init)

	mu        sync.RWMutex
	runnerSvc *runner.Service
}

// NewA2AHandler creates an A2A handler. The runner service must be set
// via SetRunnerService before requests can be processed.
func NewA2AHandler(cfg *config.AppConfig) *A2AHandler {
	return &A2AHandler{cfg: cfg}
}

// SetRunnerService sets the runner service after bootstrap completes.
func (h *A2AHandler) SetRunnerService(svc *runner.Service) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.runnerSvc = svc
}

func (h *A2AHandler) getRunner() *runner.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.runnerSvc
}

// Register registers A2A routes on the Gin engine.
func (h *A2AHandler) Register(r *gin.Engine) {
	r.GET("/a2a/:agent_name/.well-known/agent.json", h.AgentCard)
	r.POST("/a2a/:agent_name", h.TaskSend)
}

// findA2AAgent returns the agent proto config if it exists and has enable_a2a set.
func (h *A2AHandler) findA2AAgent(name string) *agentsv1.Agent {
	for i := range h.cfg.Agents {
		if h.cfg.Agents[i].GetName() == name && h.cfg.Agents[i].GetEnableA2A() {
			return &h.cfg.Agents[i]
		}
	}
	return nil
}

// AgentCard serves the A2A agent card for a specific agent.
func (h *A2AHandler) AgentCard(c *gin.Context) {
	agentName := c.Param("agent_name")

	ag := h.findA2AAgent(agentName)
	if ag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	card := buildAgentCard(ag)
	c.JSON(http.StatusOK, card)
}

// TaskSend handles A2A tasks/send JSON-RPC requests.
func (h *A2AHandler) TaskSend(c *gin.Context) {
	agentName := c.Param("agent_name")

	ag := h.findA2AAgent(agentName)
	if ag == nil {
		c.JSON(http.StatusOK, jsonRPCError(nil, -32001, "agent not found"))
		return
	}

	svc := h.getRunner()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service not ready"})
		return
	}

	var req JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, jsonRPCError(req.ID, -32700, "parse error"))
		return
	}

	if req.Method != "tasks/send" {
		c.JSON(http.StatusOK, jsonRPCError(req.ID, -32601, "method not found"))
		return
	}

	// Extract input text from A2A message params.
	input := extractInputText(&req.Params)
	if input == "" {
		c.JSON(http.StatusOK, jsonRPCError(req.ID, -32602, "empty input"))
		return
	}

	// Use task ID as session ID, or generate one.
	sessionID := req.Params.ID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        uuid.New().String(),
		SessionId:   sessionID,
		UserId:      "a2a",
		ChannelName: "a2a:" + agentName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_A2A,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		Metadata: map[string]string{
			"task_id": sessionID,
		},
	}

	result, err := svc.Run(c.Request.Context(), agentName, input, ctxInfo, nil, nil)
	if err != nil {
		c.JSON(http.StatusOK, jsonRPCError(req.ID, -32000, err.Error()))
		return
	}

	c.JSON(http.StatusOK, jsonRPCResult(req.ID, sessionID, result))
}

// --- A2A protocol types ---

// JSONRPCRequest represents an A2A JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  TaskSendParams `json:"params"`
}

// TaskSendParams contains the params for a tasks/send request.
type TaskSendParams struct {
	ID      string  `json:"id,omitempty"`
	Message Message `json:"message"`
}

// Message represents an A2A message with role and parts.
type Message struct {
	Role  string        `json:"role"`
	Parts []MessagePart `json:"parts"`
}

// MessagePart represents a part of an A2A message.
type MessagePart struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// AgentCard represents an A2A agent card.
type AgentCard struct {
	Name         string     `json:"name"`
	Description  string     `json:"description,omitempty"`
	URL          string     `json:"url"`
	Version      string     `json:"version"`
	Capabilities Capabilities `json:"capabilities"`
}

// Capabilities describes what the agent supports.
type Capabilities struct {
	Streaming bool `json:"streaming"`
}

func buildAgentCard(ag *agentsv1.Agent) AgentCard {
	return AgentCard{
		Name:        ag.GetName(),
		Description: ag.GetDescription(),
		URL:         "/a2a/" + ag.GetName(),
		Version:     "0.1.0",
		Capabilities: Capabilities{
			Streaming: false,
		},
	}
}

func extractInputText(params *TaskSendParams) string {
	for _, part := range params.Message.Parts {
		if part.Text != "" {
			return part.Text
		}
	}
	return ""
}

// --- JSON-RPC response helpers ---

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type jsonRPCErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type taskResult struct {
	ID     string  `json:"id"`
	Status status  `json:"status"`
	Output Message `json:"output,omitempty"`
}

type status struct {
	State string `json:"state"`
}

func jsonRPCError(id any, code int, message string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   jsonRPCErrorObj{Code: code, Message: message},
	}
}

func jsonRPCResult(id any, taskID, text string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: taskResult{
			ID:     taskID,
			Status: status{State: "completed"},
			Output: Message{
				Role: "agent",
				Parts: []MessagePart{
					{Type: "text", Text: text},
				},
			},
		},
	}
}
