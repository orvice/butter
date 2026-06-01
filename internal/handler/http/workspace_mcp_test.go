package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	authrepo "go.orx.me/apps/butter/internal/repo/auth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestWorkspaceMCPRequiresValidatedWorkspaceForNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := authrepo.WithAuthenticated(c.Request.Context(), &agentsv1.User{Id: "user-1", Role: "member"}, nil)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	RegisterWorkspaceMCP(r, nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusNoContent)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(nethttp.MethodPost, "/api/workspaces/ws-a/mcp", nil)
	r.ServeHTTP(w, req)
	if w.Code != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestWorkspaceMCPAllowsAdminPathWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(authrepo.WithAdmin(context.Background()))
		c.Next()
	})
	RegisterWorkspaceMCP(r, nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusNoContent)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(nethttp.MethodPost, "/api/workspaces/ws-a/mcp", nil)
	r.ServeHTTP(w, req)
	if w.Code != nethttp.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
}
