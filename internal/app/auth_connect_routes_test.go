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

// TestConnectMigratedServices_Routing is a smoke test that each migrated
// service is reachable via Connect at its /api-prefixed URL and that error
// responses use Connect's wire format ({code, message}). It does not exercise
// service-specific behavior — service tests in internal/application/*_test.go
// already do that against the underlying Twirp-shaped signatures.
func TestConnectMigratedServices_Routing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.AppConfig{Auth: config.AuthConfig{AllowUnauthenticated: true}}
	router, _ := SetupRoutes(cfg, daemon.NewRegistry())
	engine := gin.New()
	router(engine)

	cases := []struct {
		name string
		url  string
	}{
		// AllowUnauthenticated grants admin without a user, and the
		// workspace repo isn't wired, so ListWorkspaces returns an
		// empty list. Reaching that branch proves the route is mounted.
		{"workspace_list", "/api/agents.v1.WorkspaceService/ListWorkspaces"},
		// APITokenService has no repo wired in this test, so the
		// service returns FailedPrecondition. We only care that the
		// response is connect-shaped, not 404.
		{"apitoken_list", "/api/agents.v1.APITokenService/ListAPITokens"},
		{"agent_list", "/api/agents.v1.AgentService/ListAgents"},
		{"mcpserver_list", "/api/agents.v1.MCPServerService/ListMCPServers"},
		{"modelprovider_list", "/api/agents.v1.ModelProviderService/ListModelProviders"},
		{"notifygroup_list", "/api/agents.v1.NotifyGroupService/ListNotifyGroups"},
		{"remoteagent_list", "/api/agents.v1.RemoteAgentService/ListRemoteAgents"},
		{"channel_list", "/api/agents.v1.ChannelService/ListChannels"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.url, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			if w.Code == http.StatusNotFound {
				t.Fatalf("route not mounted: 404 from %s", tc.url)
			}
			var body map[string]any
			if w.Body.Len() == 0 {
				t.Fatalf("empty body from %s (status %d)", tc.url, w.Code)
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body from %s: %v; raw=%q", tc.url, err, w.Body.String())
			}
			// Error responses must carry connect's "message" key; success
			// responses won't have "message" but also won't have twirp's
			// "msg" stand-in.
			if _, hasMsg := body["msg"]; hasMsg {
				t.Fatalf("response still uses twirp-style \"msg\" key: %v", body)
			}
		})
	}
}
