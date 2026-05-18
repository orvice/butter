package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestChatStreamRequiresWorkspaceForNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewChatStreamHandler()
	h.SetRunnerService(&runner.Service{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := auth.WithAuthenticated(c.Request.Context(), &agentsv1.User{Id: "u1", Role: "member"}, &auth.Session{})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	h.Register(r)

	body := bytes.NewBufferString(`{"agent_name":"assistant","message":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/chat/stream", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}
