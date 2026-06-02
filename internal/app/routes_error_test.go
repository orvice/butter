package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"github.com/twitchtv/twirp"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
)

func TestWriteError_TwirpCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"invalid_argument", twirp.NewError(twirp.InvalidArgument, "bad"), http.StatusBadRequest, "bad"},
		{"permission_denied", twirp.NewError(twirp.PermissionDenied, "nope"), http.StatusForbidden, "nope"},
		{"not_found", twirp.NewError(twirp.NotFound, "missing"), http.StatusNotFound, "missing"},
		{"already_exists", twirp.NewError(twirp.AlreadyExists, "dup"), http.StatusConflict, "dup"},
		{"failed_precondition", twirp.NewError(twirp.FailedPrecondition, "pre"), http.StatusPreconditionFailed, "pre"},
		{"internal", twirp.NewError(twirp.Internal, "boom"), http.StatusInternalServerError, "boom"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := runWriteError(t, tc.err)
			if status != tc.wantStatus {
				t.Fatalf("status: got %d want %d", status, tc.wantStatus)
			}
			if body["error"] != tc.wantMsg {
				t.Fatalf("body: got %v want %q", body, tc.wantMsg)
			}
		})
	}
}

func TestWriteError_ConnectCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name       string
		code       connect.Code
		wantStatus int
	}{
		{"invalid_argument", connect.CodeInvalidArgument, http.StatusBadRequest},
		{"unauthenticated", connect.CodeUnauthenticated, http.StatusUnauthorized},
		{"permission_denied", connect.CodePermissionDenied, http.StatusForbidden},
		{"not_found", connect.CodeNotFound, http.StatusNotFound},
		{"already_exists", connect.CodeAlreadyExists, http.StatusConflict},
		{"failed_precondition", connect.CodeFailedPrecondition, http.StatusPreconditionFailed},
		{"unimplemented", connect.CodeUnimplemented, http.StatusNotImplemented},
		{"unavailable", connect.CodeUnavailable, http.StatusServiceUnavailable},
		{"deadline", connect.CodeDeadlineExceeded, http.StatusGatewayTimeout},
		{"resource_exhausted", connect.CodeResourceExhausted, http.StatusTooManyRequests},
		{"internal", connect.CodeInternal, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := connect.NewError(tc.code, errors.New("msg"))
			status, body := runWriteError(t, err)
			if status != tc.wantStatus {
				t.Fatalf("status: got %d want %d", status, tc.wantStatus)
			}
			if body["error"] != "msg" {
				t.Fatalf("body: got %v want msg", body)
			}
		})
	}
}

func TestWriteError_ConfigRepoSentinels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("not_found", func(t *testing.T) {
		status, _ := runWriteError(t, configrepo.ErrNotFound)
		if status != http.StatusNotFound {
			t.Fatalf("status: got %d want 404", status)
		}
	})
	t.Run("already_exists", func(t *testing.T) {
		status, _ := runWriteError(t, configrepo.ErrAlreadyExists)
		if status != http.StatusConflict {
			t.Fatalf("status: got %d want 409", status)
		}
	})
}

func runWriteError(t *testing.T, err error) (int, map[string]any) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	writeError(c, err)
	var body map[string]any
	if w.Body.Len() > 0 {
		if jerr := json.Unmarshal(w.Body.Bytes(), &body); jerr != nil {
			t.Fatalf("unmarshal body: %v", jerr)
		}
	}
	return w.Code, body
}
