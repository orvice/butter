package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	wsctx "go.orx.me/apps/butter/internal/workspace"
)

// OpenAIHandler serves OpenAI-compatible API endpoints for enabled agents.
type OpenAIHandler struct {
	agentRepo configrepo.AgentRepository
}

// NewOpenAIHandler creates an OpenAI handler with the given agent repository.
func NewOpenAIHandler(repo configrepo.AgentRepository) *OpenAIHandler {
	return &OpenAIHandler{agentRepo: repo}
}

// Register registers OpenAI-compatible routes on the Gin engine.
func (h *OpenAIHandler) Register(r *gin.Engine) {
	r.GET("/api/v1/models", h.ListModels)
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
