package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeTelegramWebhookRouter struct {
	handlers map[string]http.Handler
}

func (r fakeTelegramWebhookRouter) TelegramWebhookHandler(channelName string) (http.Handler, bool) {
	h, ok := r.handlers[channelName]
	return h, ok
}

func TestTelegramWebhookRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	router := fakeTelegramWebhookRouter{
		handlers: map[string]http.Handler{
			"main": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("method = %s, want POST", r.Method)
				}
				w.WriteHeader(http.StatusAccepted)
			}),
		},
	}
	registerTelegramWebhookRoute(engine, func() telegramWebhookRouter { return router })

	req := httptest.NewRequest(http.MethodPost, "/webhooks/telegram/main", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	req = httptest.NewRequest(http.MethodPost, "/webhooks/telegram/missing", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing channel status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestTelegramWebhookRouteUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerTelegramWebhookRoute(engine, func() telegramWebhookRouter { return nil })

	req := httptest.NewRequest(http.MethodPost, "/webhooks/telegram/main", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
