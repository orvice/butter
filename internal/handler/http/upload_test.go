package http

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/service"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// uploadTestRouter wires UploadHandler with a service that reports Enabled()
// but has no real S3 client, so the auth gate is exercised in isolation —
// the S3 PutObject call (if reached) errors out below the auth check.
func uploadTestRouter(t *testing.T, user *agentsv1.User, workspaceID string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := c.Request.Context()
		if user != nil {
			ctx = auth.WithAuthenticated(ctx, user, &auth.Session{})
		}
		if workspaceID != "" {
			ctx = workspace.WithID(ctx, workspaceID)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	svc := service.NewUploadServiceLazy(func() config.StaticConfig {
		return config.StaticConfig{S3Bucket: "test-bucket"}
	})
	NewUploadHandler(svc).Register(r)
	return r
}

func multipartImage(t *testing.T) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile("file", "icon.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("\x89PNG\r\n\x1a\n")); err != nil {
		t.Fatalf("write image bytes: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return body, w.FormDataContentType()
}

func TestUploadAvatarFor_AgentAllowsNonAdmin(t *testing.T) {
	r := uploadTestRouter(t, &agentsv1.User{Id: "u1", Role: "user"}, "ws1")
	body, ct := multipartImage(t)

	req := httptest.NewRequest(http.MethodPost, "/api/uploads/avatar/agent/my-agent", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusForbidden {
		t.Fatalf("non-admin should be allowed to upload agent icon; got 403 body=%s", w.Body.String())
	}
}

func TestUploadAvatarFor_AgentRequiresWorkspace(t *testing.T) {
	r := uploadTestRouter(t, &agentsv1.User{Id: "u1", Role: "user"}, "")
	body, ct := multipartImage(t)

	req := httptest.NewRequest(http.MethodPost, "/api/uploads/avatar/agent/my-agent", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("agent avatar upload without workspace should 403; got %d body=%s", w.Code, w.Body.String())
	}
}

func TestUploadAvatarFor_UserRequiresSelf(t *testing.T) {
	r := uploadTestRouter(t, &agentsv1.User{Id: "u1", Role: "user"}, "")
	body, ct := multipartImage(t)

	req := httptest.NewRequest(http.MethodPost, "/api/uploads/avatar/user/u2", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("non-admin uploading other user's avatar should 403; got %d body=%s", w.Code, w.Body.String())
	}
}

func TestUploadAvatarFor_OtherKindRequiresAdmin(t *testing.T) {
	r := uploadTestRouter(t, &agentsv1.User{Id: "u1", Role: "user"}, "")
	body, ct := multipartImage(t)

	req := httptest.NewRequest(http.MethodPost, "/api/uploads/avatar/workspace/ws1", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("non-admin uploading workspace avatar should 403; got %d body=%s", w.Code, w.Body.String())
	}
}
