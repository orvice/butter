package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	apitokenmemory "go.orx.me/apps/butter/internal/repo/apitoken/memory"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/service"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func setupAuthRouter(cfg *config.AppConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(APITokenAuthMiddleware(cfg, nil))
	NewHealthHandler(service.NewHealthService(repo.NewHealthRepository(), cfg)).Register(r)
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestAPITokenAuthMiddleware_NoAuthConfiguredRejects(t *testing.T) {
	// Default: no api_token, no repos, allow_unauthenticated=false → reject.
	r := setupAuthRouter(&config.AppConfig{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when no auth is configured, got %d", w.Code)
	}
}

func TestAPITokenAuthMiddleware_AllowUnauthenticated(t *testing.T) {
	// Explicit opt-in: dev fallback grants admin to every request.
	r := setupAuthRouter(&config.AppConfig{
		Auth: config.AuthConfig{AllowUnauthenticated: true},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when allow_unauthenticated=true, got %d", w.Code)
	}
}

func TestAPITokenAuthMiddleware_MissingHeader(t *testing.T) {
	r := setupAuthRouter(&config.AppConfig{APIToken: "secret-token"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAPITokenAuthMiddleware_InvalidToken(t *testing.T) {
	r := setupAuthRouter(&config.AppConfig{APIToken: "secret-token"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAPITokenAuthMiddleware_ValidToken(t *testing.T) {
	r := setupAuthRouter(&config.AppConfig{APIToken: "secret-token"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPITokenAuthMiddleware_PingBypass(t *testing.T) {
	r := setupAuthRouter(&config.AppConfig{APIToken: "secret-token"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestIsPublicPathAllowsDaemonConnectorRPCs(t *testing.T) {
	paths := []string{
		"/api/agents.v1.DaemonConnectorService/Connect",
		"/api/agents.v1.DaemonConnectorService/Register",
		"/api/agents.v1.DaemonConnectorService/Poll",
		"/api/agents.v1.DaemonConnectorService/ReportTaskUpdate",
		"/api/agents.v1.DaemonConnectorService/Unregister",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			if !isPublicPath(path) {
				t.Fatalf("expected %q to bypass HTTP API auth", path)
			}
		})
	}
}

// A workspace API token authenticates as itself: the request carries the
// token's workspace and the API-token marker the session policy relies on,
// and no user.
func TestAuthMiddleware_WorkspaceAPITokenMarksTheContext(t *testing.T) {
	store := apitokenmemory.New()
	if err := store.Create(t.Context(), &agentsv1.APIToken{
		Id: "tok-1", WorkspaceId: "ws-a",
		Kind: agentsv1.APITokenKind_API_TOKEN_KIND_USER, Scopes: []string{"api:*"},
	}, hashSecret("bt_secret")); err != nil {
		t.Fatalf("seed token: %v", err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(APITokenAuthMiddleware(&config.AppConfig{}, func() apitoken.Repository { return store }))
	var tokenID, wsID string
	var hasUser bool
	r.GET("/protected", func(c *gin.Context) {
		tokenID, _ = auth.APITokenFromContext(c.Request.Context())
		wsID, _ = wsctx.FromContext(c.Request.Context())
		_, hasUser = auth.UserFromContext(c.Request.Context())
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer bt_secret")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if tokenID != "tok-1" || wsID != "ws-a" || hasUser {
		t.Fatalf("context: token=%q workspace=%q user=%v, want tok-1, ws-a, no user", tokenID, wsID, hasUser)
	}
}
