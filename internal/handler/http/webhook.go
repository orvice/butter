package http

import (
	"context"
	"io"
	"net/http"

	"butterfly.orx.me/core/log"
	"github.com/gin-gonic/gin"
)

// WebhookSyncTrigger verifies webhook signatures and triggers
// sync + publish for a workspace.
type WebhookSyncTrigger interface {
	VerifyWebhookSignature(ctx context.Context, ws string, body []byte, signatureHeader, tokenHeader string) bool
	TriggerSyncAndPublish(ctx context.Context, ws string) error
}

// WebhookHandler receives repository push events from GitHub or GitLab
// webhooks and triggers synchronization + publication.
type WebhookHandler struct {
	trigger WebhookSyncTrigger
}

func NewWebhookHandler(trigger WebhookSyncTrigger) *WebhookHandler {
	return &WebhookHandler{trigger: trigger}
}

// Handle processes incoming webhook payloads.
// Route: POST /api/webhooks/repository/:workspace_id
func (h *WebhookHandler) Handle(c *gin.Context) {
	ws := c.Param("workspace_id")
	if ws == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id required"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	ctx := c.Request.Context()
	githubSig := c.GetHeader("X-Hub-Signature-256")
	gitlabToken := c.GetHeader("X-Gitlab-Token")

	if !h.trigger.VerifyWebhookSignature(ctx, ws, body, githubSig, gitlabToken) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
		return
	}

	logger := log.FromContext(ctx)
	logger.Info("webhook received, triggering sync", "workspace_id", ws)

	go func() {
		bgCtx := context.Background()
		if err := h.trigger.TriggerSyncAndPublish(bgCtx, ws); err != nil {
			logger.Error("webhook-triggered sync failed", "workspace_id", ws, "err", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "accepted"})
}
