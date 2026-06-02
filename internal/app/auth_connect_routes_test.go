package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
)

// TestAuthService_ConnectRouting verifies AuthService is reachable via the
// Connect protocol at the /api-prefixed URL the frontend still uses, and that
// the wire-level error shape is Connect's {code, message} rather than Twirp's
// {code, msg}.
func TestAuthService_ConnectRouting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.AppConfig{
		Auth: config.AuthConfig{AllowUnauthenticated: true},
	}
	router, _ := SetupRoutes(cfg, daemon.NewRegistry())

	engine := gin.New()
	router(engine)

	// ListOAuthProviders is on the public allowlist and returns an empty
	// list when no providers are configured. A successful round-trip proves
	// the connect handler is mounted at /api/agents.v1.AuthService/.
	t.Run("public_endpoint_round_trip", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/agents.v1.AuthService/ListOAuthProviders",
			strings.NewReader(`{}`),
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d want 200, body=%q", w.Code, w.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode body: %v; raw=%q", err, w.Body.String())
		}
	})

	// Me requires an authenticated user. Under allow_unauthenticated the
	// middleware grants admin but does not populate the user context, so the
	// service returns Unauthenticated. Connect serializes that as HTTP 401
	// with {"code":"unauthenticated","message":...}. The "message" field
	// (not "msg") is the protocol marker we care about.
	t.Run("error_shape_is_connect", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/agents.v1.AuthService/Me",
			strings.NewReader(`{}`),
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d want 401, body=%q", w.Code, w.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode body: %v; raw=%q", err, w.Body.String())
		}
		if _, ok := body["message"]; !ok {
			t.Fatalf("expected connect-style {code,message} body, got %v", body)
		}
		if _, ok := body["msg"]; ok {
			t.Fatalf("body still contains twirp-style \"msg\" field: %v", body)
		}
		if got := body["code"]; got != "unauthenticated" {
			t.Fatalf("code: got %v want unauthenticated", got)
		}
	})

	// CORS preflight must not be rejected by AuthMiddleware. Without this
	// the browser-side Connect-Web client cannot reach the endpoint at all
	// when the dashboard is hosted on a different origin.
	t.Run("options_preflight_passes_auth", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodOptions,
			"/api/agents.v1.AuthService/Login",
			nil,
		)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized {
			t.Fatalf("OPTIONS preflight was rejected by auth middleware (401)")
		}
	})
}
