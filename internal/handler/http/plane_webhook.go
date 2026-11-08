package http

import (
	"butterfly.orx.me/core/log"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	planeruntime "go.orx.me/apps/butter/internal/runtime/plane"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"io"
	"net/http"
)

const PlaneWebhookPath = "/api/plane/webhook/:app_id"

type PlaneReceiver interface {
	FindApp(context.Context, string) (*agentsv1.PlaneApp, error)
	Deliver(context.Context, *agentsv1.PlaneApp, http.Header, []byte) (planeruntime.Decision, error)
}
type PlaneWebhookHandler struct{ receiver PlaneReceiver }

func NewPlaneWebhookHandler(r PlaneReceiver) *PlaneWebhookHandler {
	return &PlaneWebhookHandler{receiver: r}
}
func (h *PlaneWebhookHandler) Register(r *gin.Engine) { r.POST(PlaneWebhookPath, h.Handle) }
func (h *PlaneWebhookHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	if h.receiver == nil {
		c.JSON(503, gin.H{"error": "plane receive is not configured"})
		return
	}
	app, err := h.receiver.FindApp(ctx, c.Param("app_id"))
	if err != nil {
		if errors.Is(err, planeruntime.ErrAppNotFound) {
			c.JSON(404, gin.H{"error": "unknown app"})
		} else {
			c.JSON(503, gin.H{"error": "temporarily unavailable"})
		}
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, planeruntime.MaxWebhookBytes))
	if err != nil {
		if _, ok := err.(*http.MaxBytesError); ok {
			c.JSON(413, gin.H{"error": "payload too large"})
		} else {
			c.JSON(503, gin.H{"error": "could not read request"})
		}
		return
	}
	decision, err := h.receiver.Deliver(ctx, app, c.Request.Header, body)
	if err == nil {
		c.JSON(200, gin.H{"ok": true, "decision": string(decision)})
		return
	}
	logger := log.FromContext(ctx)
	logger.Warn("rejected plane webhook", "app_id", app.GetId(), "err", err)
	switch {
	case errors.Is(err, planeruntime.ErrUnauthorized), errors.Is(err, planeruntime.ErrIdentity):
		c.JSON(401, gin.H{"error": "unauthorized"})
	case errors.Is(err, planeruntime.ErrMalformed):
		c.JSON(400, gin.H{"error": "malformed webhook"})
	default:
		c.JSON(503, gin.H{"error": "temporarily unavailable"})
	}
}
