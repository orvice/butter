package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/service"
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
