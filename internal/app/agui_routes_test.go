package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
)

// TestAGUI_RouteRegistration verifies the AG-UI endpoint is mounted by
// SetupRoutes and sits behind the shared auth middleware. The handler's own
// behaviour is covered in internal/handler/http; what can only be checked here
// is the wiring: that the route exists at all, and that it is *not* on the
// public allowlist.
func TestAGUI_RouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := `{"threadId":"t-1","runId":"r-1","messages":[{"id":"m1","role":"user","content":"hi"}]}`

	// With auth wired and no credentials presented, the request must be
	// rejected by the middleware before ever reaching the handler.
	t.Run("requires_auth", func(t *testing.T) {
		cfg := &config.AppConfig{APIToken: "secret-token"}
		router, _ := SetupRoutes(cfg, daemon.NewRegistry())
		engine := gin.New()
		router(engine)

		req := httptest.NewRequest(http.MethodPost, "/api/agui/writer", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d want 401, body=%q", w.Code, w.Body.String())
		}
	})

	// Past the middleware the route must resolve to the AG-UI handler rather
	// than 404-ing as an unknown path. No agent exists in this bare setup, so
	// the handler's own "agent not found" is the expected answer — and a 404
	// carrying that message can only have come from the handler, since gin's
	// routing 404 has an empty body.
	t.Run("route_is_mounted_and_workspace_scoped", func(t *testing.T) {
		cfg := &config.AppConfig{
			APIToken: "secret-token",
			Auth:     config.AuthConfig{AllowUnauthenticated: true},
		}
		router, _ := SetupRoutes(cfg, daemon.NewRegistry())
		engine := gin.New()
		router(engine)

		newRequest := func(workspace string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPost, "/api/agui/writer", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer secret-token")
			if workspace != "" {
				req.Header.Set("X-Workspace-ID", workspace)
			}
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			return w
		}

		// Without a workspace header the handler refuses before doing any work.
		if w := newRequest(""); w.Code != http.StatusUnauthorized ||
			!strings.Contains(w.Body.String(), "workspace required") {
			t.Fatalf("no-workspace: got %d %q, want 401 requiring a workspace", w.Code, w.Body.String())
		}

		// With one, the request reaches the agent lookup.
		w := newRequest("default")
		if w.Code != http.StatusNotFound {
			t.Fatalf("status: got %d want 404 from the handler, body=%q", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "agent not found") {
			t.Fatalf("body %q does not look like the AG-UI handler's response; "+
				"a gin routing 404 would have an empty body", w.Body.String())
		}
	})
}
