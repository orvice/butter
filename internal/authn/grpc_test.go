package authn

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/auth"
	wsctx "go.orx.me/apps/butter/internal/workspace"
)

func TestMetadataSource_LowercasesCanonicalName(t *testing.T) {
	md := metadata.New(map[string]string{
		"authorization":  "Bearer t",
		"x-workspace-id": "ws-1",
	})
	src := metadataSource{md: md}

	if got := src.Get("Authorization"); got != "Bearer t" {
		t.Errorf("Authorization: got %q", got)
	}
	if got := src.Get("X-Workspace-ID"); got != "ws-1" {
		t.Errorf("X-Workspace-ID: got %q", got)
	}
	if got := src.Get("Missing"); got != "" {
		t.Errorf("missing header should be empty, got %q", got)
	}
}

func TestMetadataSource_NilSafe(t *testing.T) {
	src := metadataSource{}
	if got := src.Get("Authorization"); got != "" {
		t.Fatalf("nil metadata should yield empty string, got %q", got)
	}
}

// callUnary is a tiny helper so each test case stays terse.
func callUnary(t *testing.T, interceptor grpc.UnaryServerInterceptor, fullMethod string, md metadata.MD) (context.Context, error) {
	t.Helper()
	ctx := metadata.NewIncomingContext(context.Background(), md)
	var handlerCtx context.Context
	_, err := interceptor(ctx, struct{}{}, &grpc.UnaryServerInfo{FullMethod: fullMethod}, func(c context.Context, _ any) (any, error) {
		handlerCtx = c
		return "ok", nil
	})
	return handlerCtx, err
}

func TestUnaryServerInterceptor_PublicMethodBypass(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	interceptor := UnaryServerInterceptor(r, nil)

	// No metadata at all — public method should still reach the handler.
	ctx, err := callUnary(t, interceptor, "/agents.v1.AuthService/Login", nil)
	if err != nil {
		t.Fatalf("public method should not error, got %v", err)
	}
	if ctx == nil {
		t.Fatalf("handler should have been called")
	}
}

func TestUnaryServerInterceptor_RejectsMissingAuth(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	interceptor := UnaryServerInterceptor(r, nil)

	_, err := callUnary(t, interceptor, "/agents.v1.AgentService/ListAgents", metadata.New(nil))
	if err == nil {
		t.Fatalf("expected Unauthenticated, got nil err")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %T", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestUnaryServerInterceptor_PropagatesAuthCtx(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	interceptor := UnaryServerInterceptor(r, nil)

	md := metadata.New(map[string]string{
		"authorization":  "Bearer secret",
		"x-workspace-id": "ws-7",
	})
	ctx, err := callUnary(t, interceptor, "/agents.v1.AgentService/ListAgents", md)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !auth.IsAdmin(ctx) {
		t.Fatalf("root token should produce admin ctx")
	}
	if id, ok := wsctx.FromContext(ctx); !ok || id != "ws-7" {
		t.Fatalf("workspace not propagated: got %q ok=%v", id, ok)
	}
}

// fakeServerStream implements grpc.ServerStream with a swappable context;
// only Context() is exercised by the stream interceptor under test.
type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *fakeServerStream) Context() context.Context { return s.ctx }

func TestStreamServerInterceptor_PropagatesAuthCtx(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	interceptor := StreamServerInterceptor(r, nil)

	md := metadata.New(map[string]string{
		"authorization":  "Bearer secret",
		"x-workspace-id": "ws-8",
	})
	incoming := metadata.NewIncomingContext(context.Background(), md)
	ss := &fakeServerStream{ctx: incoming}

	var observed context.Context
	err := interceptor(nil, ss, &grpc.StreamServerInfo{FullMethod: "/agents.v1.AgentService/Stream"},
		func(_ any, wrapped grpc.ServerStream) error {
			observed = wrapped.Context()
			return nil
		})
	if err != nil {
		t.Fatalf("stream interceptor returned error: %v", err)
	}
	if observed == nil {
		t.Fatalf("handler did not see the wrapped stream")
	}
	if id, ok := wsctx.FromContext(observed); !ok || id != "ws-8" {
		t.Fatalf("workspace not propagated to stream handler: got %q ok=%v", id, ok)
	}
}

func TestStreamServerInterceptor_RejectsMissingAuth(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	interceptor := StreamServerInterceptor(r, nil)

	ss := &fakeServerStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(nil))}
	err := interceptor(nil, ss, &grpc.StreamServerInfo{FullMethod: "/agents.v1.AgentService/Stream"},
		func(any, grpc.ServerStream) error { return errors.New("handler should not run") })
	if err == nil {
		t.Fatalf("expected error")
	}
	if st, _ := status.FromError(err); st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

// TestPublicMethods_NoStaleEntries fails loudly if someone removes a
// public method from the gRPC list without also removing the matching
// HTTP path in isPublicPath, and vice versa. The two lists are
// hand-maintained twins; this guards the invariant.
func TestPublicMethods_StaysInSyncWithHTTPList(t *testing.T) {
	wantGRPC := []string{
		"/agents.v1.AuthService/Login",
		"/agents.v1.AuthService/ListOAuthProviders",
		"/agents.v1.AuthService/BeginOAuthFlow",
		"/agents.v1.AuthService/CompleteOAuthFlow",
		"/agents.v1.DaemonConnectorService/Connect",
	}
	got := PublicMethods()
	if len(got) != len(wantGRPC) {
		t.Fatalf("PublicMethods size %d, want %d", len(got), len(wantGRPC))
	}
	for _, m := range wantGRPC {
		if _, ok := got[m]; !ok {
			t.Errorf("missing public method: %s", m)
		}
	}
}
