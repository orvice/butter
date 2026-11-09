package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/gin-gonic/gin"
	planeruntime "go.orx.me/apps/butter/internal/runtime/plane"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const PlaneWebhookPath = "/api/plane/webhook/:app_id"

type PlaneReceiver interface {
	FindApp(context.Context, string) (*agentsv1.PlaneApp, error)
	Deliver(context.Context, *agentsv1.PlaneApp, http.Header, []byte) (planeruntime.Decision, error)
}

// ErrIdentity/ErrMalformed from Deliver are authenticated permanent failures.
// A 200 rejected receipt is safe only once their diagnostic is persisted.
// Missing persistence must remain retryable, never an unconditional success.
type planeWebhookRejectionRecorder interface {
	RecordWebhookRejection(context.Context, *agentsv1.PlaneApp, []byte, error) (string, error)
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
	digest := sha256.Sum256([]byte(app.GetId()))
	logger := log.FromContext(ctx).With("app_hash", hex.EncodeToString(digest[:6]))
	decision, err := h.receiver.Deliver(ctx, app, c.Request.Header, body)
	if err == nil {
		logger.Info("plane webhook received", "decision", string(decision))
		c.JSON(200, gin.H{"ok": true, "decision": string(decision)})
		return
	}
	logger.Warn("rejected plane webhook", "err", err)
	switch {
	case errors.Is(err, planeruntime.ErrUnauthorized):
		c.JSON(401, gin.H{"error": "unauthorized"})
	case errors.Is(err, planeruntime.ErrIdentity), errors.Is(err, planeruntime.ErrMalformed):
		recorder, ok := h.receiver.(planeWebhookRejectionRecorder)
		if !ok {
			c.JSON(503, gin.H{"error": "temporarily unavailable"})
			return
		}
		recordCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		receipt, recordErr := recorder.RecordWebhookRejection(recordCtx, app, body, err)
		if recordErr != nil || receipt == "" {
			logger.Error("plane webhook rejection receipt unavailable")
			c.JSON(503, gin.H{"error": "temporarily unavailable"})
			return
		}
		receiptDigest := sha256.Sum256([]byte(receipt))
		logger.Info("plane webhook acknowledged", "decision", string(planeruntime.DecisionRejected), "receipt_hash", hex.EncodeToString(receiptDigest[:6]))
		c.JSON(200, gin.H{"ok": true, "decision": string(planeruntime.DecisionRejected), "receipt_id": receipt})
	default:
		c.JSON(503, gin.H{"error": "temporarily unavailable"})
	}
}
