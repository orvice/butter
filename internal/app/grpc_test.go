package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// TestGRPCWebDispatcher_PathDispatch verifies that the dispatcher
// intercepts paths shaped like a gRPC procedure under the agents.v1
// package and lets every other path fall through to the rest of the Gin
// chain. The wrapped server has no services registered so intercepted
// requests will end with a non-2xx status — the test only asserts
// whether the sentinel handler past the dispatcher ran.
func TestGRPCWebDispatcher_PathDispatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	wrapped := NewGRPCWebHandler(grpc.NewServer())

	cases := []struct {
		name             string
		path             string
		method           string
		expectIntercept  bool
		expectSentinelOK bool
	}{
		{
			name:             "agents.v1 RPC path is intercepted",
			path:             "/agents.v1.AgentService/ListAgents",
			method:           http.MethodPost,
			expectIntercept:  true,
			expectSentinelOK: false,
		},
		{
			name:             "twirp path under /api falls through",
			path:             "/api/agents.v1.AgentService/ListAgents",
			method:           http.MethodPost,
			expectIntercept:  false,
			expectSentinelOK: true,
		},
		{
			name:             "REST path falls through",
			path:             "/ping",
			method:           http.MethodGet,
			expectIntercept:  false,
			expectSentinelOK: true,
		},
		{
			name:             "service-only path (no method segment) falls through",
			path:             "/agents.v1.AgentService",
			method:           http.MethodPost,
			expectIntercept:  false,
			expectSentinelOK: true,
		},
		{
			name:             "extra trailing segment falls through",
			path:             "/agents.v1.AgentService/ListAgents/extra",
			method:           http.MethodPost,
			expectIntercept:  false,
			expectSentinelOK: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(grpcWebDispatcher(wrapped))

			var sentinelRan bool
			handler := func(c *gin.Context) {
				sentinelRan = true
				c.Status(http.StatusOK)
			}
			r.Any("/*path", handler)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(w, req)

			if tc.expectSentinelOK && !sentinelRan {
				t.Fatalf("expected request to reach the downstream handler, status=%d", w.Code)
			}
			if tc.expectIntercept && sentinelRan {
				t.Fatalf("dispatcher should have aborted, but downstream handler ran (status=%d)", w.Code)
			}
		})
	}
}

// TestGRPCWebDispatcher_PreflightIntercepted exercises the CORS
// preflight branch through the wrapped server's own helper. Browsers
// hit grpc-web endpoints with OPTIONS before the real call.
func TestGRPCWebDispatcher_PreflightIntercepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	wrapped := NewGRPCWebHandler(grpc.NewServer())

	r := gin.New()
	var sentinelRan bool
	r.Use(grpcWebDispatcher(wrapped))
	r.Any("/*path", func(c *gin.Context) {
		sentinelRan = true
	})

	req := httptest.NewRequest(http.MethodOptions, "/agents.v1.AgentService/ListAgents", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "x-grpc-web,content-type,authorization")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if sentinelRan {
		t.Fatalf("preflight should be handled by the grpc-web wrapper, not the downstream handler")
	}
}
