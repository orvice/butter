package http

import (
	"crypto/subtle"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/runtime/telegram"
)

// TelegramWebhookPath is the public callback route. It carries only the
// immutable Channel ID: the workspace is read off the Channel, so a caller
// cannot assert tenancy, and the URL stays valid across renames.
const TelegramWebhookPath = "/api/telegram/webhook/:channel_id"

// telegramSecretHeader is the header Telegram echoes from setWebhook.
const telegramSecretHeader = "X-Telegram-Bot-Api-Secret-Token"

// TelegramReceiver is the receive-path seam the handler drives.
type TelegramReceiver interface {
	// Authenticate verifies the per-Channel secret in constant time and
	// returns the Channel. It is called before the body is parsed.
	Authenticate(ctx *gin.Context, channelID, secret string) (telegram.AuthenticatedChannel, error)
	// Deliver routes one raw update.
	Deliver(ctx *gin.Context, channel telegram.AuthenticatedChannel, raw []byte) (telegram.Decision, error)
}

// TelegramWebhookHandler serves Telegram callbacks on every Pod.
//
// Its contract with Telegram is narrow and deliberate: 200 means "durably
// accepted", and anything else means "retry". Returning 200 for an update we
// failed to enqueue would lose it permanently, so infrastructure failures
// answer 503 even though the request itself was well-formed.
type TelegramWebhookHandler struct {
	receiver TelegramReceiver
}

func NewTelegramWebhookHandler(receiver TelegramReceiver) *TelegramWebhookHandler {
	return &TelegramWebhookHandler{receiver: receiver}
}

func (h *TelegramWebhookHandler) Register(r *gin.Engine) {
	r.POST(TelegramWebhookPath, h.Handle)
}

func (h *TelegramWebhookHandler) Handle(c *gin.Context) {
	logger := log.FromContext(c.Request.Context())
	channelID := c.Param("channel_id")

	if h.receiver == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "telegram receive is not configured"})
		return
	}

	// Authenticate before parsing: an unauthenticated caller must not be
	// able to make this Pod spend anything on their payload.
	channel, err := h.receiver.Authenticate(c, channelID, c.GetHeader(telegramSecretHeader))
	if err != nil {
		switch {
		case errors.Is(err, telegram.ErrUnauthorized):
			logger.Warn("rejected telegram webhook with an invalid secret", "channel_id", channelID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		case errors.Is(err, telegram.ErrChannelNotReceiving):
			// The Channel exists but is not accepting updates. Telegram
			// should stop rather than retry; the reconciler will remove the
			// registration.
			logger.Info("ignoring telegram webhook for a non-receiving channel", "channel_id", channelID)
			c.JSON(http.StatusOK, gin.H{"ok": true, "ignored": true})
		case errors.Is(err, telegram.ErrChannelNotFound):
			logger.Warn("telegram webhook for an unknown channel", "channel_id", channelID)
			c.JSON(http.StatusNotFound, gin.H{"error": "unknown channel"})
		default:
			// We could not tell whether this is authentic, so ask Telegram
			// to retry instead of dropping a possibly-valid update.
			logger.Error("telegram webhook authentication failed", "channel_id", channelID, "err", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "temporarily unavailable"})
		}
		return
	}

	// No application-level body size limit: infrastructure caps remain
	// possible, but Butter must not silently truncate an update Telegram
	// considers valid.
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Warn("could not read telegram webhook body", "channel_id", channelID, "err", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not read request"})
		return
	}

	decision, err := h.receiver.Deliver(c, channel, raw)
	if err != nil {
		if decision == telegram.DecisionIgnored && isPermanent(err) {
			// Malformed JSON will never succeed; acknowledge so Telegram
			// stops retrying it.
			logger.Warn("rejecting malformed telegram update", "channel_id", channelID, "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "malformed update"})
			return
		}
		logger.Error("telegram update was not accepted", "channel_id", channelID, "err", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "temporarily unavailable"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "decision": string(decision)})
}

// isPermanent distinguishes "this payload can never work" from "we could not
// process it right now". Only the former is acknowledged.
func isPermanent(err error) bool {
	return errors.Is(err, telegram.ErrMalformedUpdate)
}

// ConstantTimeSecretEqual compares a presented secret with the expected one
// without leaking length or content through timing.
func ConstantTimeSecretEqual(presented, expected string) bool {
	if expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(presented), []byte(expected)) == 1
}
