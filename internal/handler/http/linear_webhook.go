package http

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"butterfly.orx.me/core/log"
	linearruntime "go.orx.me/apps/butter/internal/runtime/linear"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// LinearWebhookPath is the public per-App webhook route. It carries only the
// immutable App ID: the workspace is read off the App, so a sender cannot
// assert tenancy.
const LinearWebhookPath = "/api/linear/webhook/:app_id"

// linearWebhookPrefix is what the auth middleware lets through unauthenticated.
const linearWebhookPrefix = "/api/linear/webhook/"

// linearMaxWebhookBytes caps a delivery. The signature covers the whole
// body, so it has to be read before anything is verified.
const linearMaxWebhookBytes = 1 << 20

// LinearReceiver is the receive-path seam the handler drives.
type LinearReceiver interface {
	FindApp(ctx context.Context, appID string) (*agentsv1.LinearApp, error)
	Deliver(ctx context.Context, app *agentsv1.LinearApp, header http.Header, body []byte) (linearruntime.Decision, error)
}

// LinearWebhookHandler serves Linear deliveries on every Pod.
//
// Its contract with Linear is narrow: 200 means "durably accepted" (or
// deliberately ignored), and anything else tells Linear whether a retry can
// help. Returning 200 for a delivery we failed to enqueue would lose it, so
// infrastructure failures answer 503.
type LinearWebhookHandler struct {
	receiver LinearReceiver
}

func NewLinearWebhookHandler(receiver LinearReceiver) *LinearWebhookHandler {
	return &LinearWebhookHandler{receiver: receiver}
}

func (h *LinearWebhookHandler) Register(r *gin.Engine) {
	r.POST(LinearWebhookPath, h.Handle)
}

func (h *LinearWebhookHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	logger := log.FromContext(ctx)
	appID := c.Param("app_id")
	if h.receiver == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "linear receive is not configured"})
		return
	}

	app, err := h.receiver.FindApp(ctx, appID)
	if err != nil {
		if errors.Is(err, linearruntime.ErrAppNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "unknown app"})
			return
		}
		logger.Error("linear webhook app lookup failed", "app_id", appID, "err", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "temporarily unavailable"})
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, linearMaxWebhookBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "payload too large"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not read request"})
		return
	}

	decision, err := h.receiver.Deliver(ctx, app, c.Request.Header, body)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"ok": true, "decision": string(decision)})
	case errors.Is(err, linearruntime.ErrUnauthorized), errors.Is(err, linearruntime.ErrStale):
		logger.Warn("rejected linear webhook", "app_id", appID, "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case errors.Is(err, linearruntime.ErrMalformed):
		logger.Warn("rejected malformed linear webhook", "app_id", appID, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed webhook"})
	default:
		logger.Error("linear webhook was not accepted", "app_id", appID, "err", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "temporarily unavailable"})
	}
}
