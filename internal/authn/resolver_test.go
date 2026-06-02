package authn

import (
	"context"
	"strings"
	"testing"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/auth"
	wsctx "go.orx.me/apps/butter/internal/workspace"
)

// staticHeaders mimics HeaderSource for tests; values are matched
// case-insensitively so the same fixture works for both Gin (canonical
// MIME form) and gRPC metadata (lowercased) call sites.
type staticHeaders map[string]string

func (s staticHeaders) Get(name string) string {
	if v, ok := s[name]; ok {
		return v
	}
	return s[strings.ToLower(name)]
}

func TestResolver_NoAuthWired_RejectByDefault(t *testing.T) {
	r := New(&config.AppConfig{}, nil, nil, nil)
	got := r.Resolve(context.Background(), staticHeaders{})
	if got.Outcome != OutcomeRejected {
		t.Fatalf("expected reject when nothing is wired, got %v", got.Outcome)
	}
}

func TestResolver_AllowUnauthenticated_GrantsAdmin(t *testing.T) {
	cfg := &config.AppConfig{Auth: config.AuthConfig{AllowUnauthenticated: true}}
	r := New(cfg, nil, nil, nil)
	got := r.Resolve(context.Background(), staticHeaders{})
	if got.Outcome != OutcomeAuthenticated {
		t.Fatalf("expected authenticated, got %v", got.Outcome)
	}
	if !auth.IsAdmin(got.Ctx) {
		t.Fatalf("expected admin context in dev fallback")
	}
}

func TestResolver_AllowUnauthenticated_ReadsConfigLazily(t *testing.T) {
	cfg := &config.AppConfig{}
	r := New(cfg, nil, nil, nil)

	cfg.Auth.AllowUnauthenticated = true

	got := r.Resolve(context.Background(), staticHeaders{})
	if got.Outcome != OutcomeAuthenticated {
		t.Fatalf("expected authenticated after allow_unauthenticated is loaded, got %v", got.Outcome)
	}
	if !auth.IsAdmin(got.Ctx) {
		t.Fatalf("expected admin context in dev fallback")
	}
}

func TestResolver_RootToken_Match(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	got := r.Resolve(context.Background(), staticHeaders{"Authorization": "Bearer secret"})
	if got.Outcome != OutcomeAuthenticated {
		t.Fatalf("expected authenticated, got %v", got.Outcome)
	}
	if !auth.IsAdmin(got.Ctx) {
		t.Fatalf("root token should grant admin")
	}
}

func TestResolver_RootToken_ReadsConfigLazily(t *testing.T) {
	cfg := &config.AppConfig{}
	r := New(cfg, nil, nil, nil)

	cfg.APIToken = "loaded-secret"

	got := r.Resolve(context.Background(), staticHeaders{"Authorization": "Bearer loaded-secret"})
	if got.Outcome != OutcomeAuthenticated {
		t.Fatalf("expected authenticated after token is loaded, got %v", got.Outcome)
	}
	if !auth.IsAdmin(got.Ctx) {
		t.Fatalf("root token should grant admin")
	}
}

func TestResolver_RootToken_Mismatch(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	got := r.Resolve(context.Background(), staticHeaders{"Authorization": "Bearer nope"})
	if got.Outcome != OutcomeRejected {
		t.Fatalf("expected reject for bad token, got %v", got.Outcome)
	}
}

func TestResolver_BearerParsing(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)

	cases := map[string]Outcome{
		"":                 OutcomeRejected, // missing
		"secret":           OutcomeRejected, // no scheme
		"Token secret":     OutcomeRejected, // wrong scheme
		"Bearer":           OutcomeRejected, // empty token
		"Bearer ":          OutcomeRejected, // empty after trim
		"Bearer secret":    OutcomeAuthenticated,
		"Bearer  secret  ": OutcomeAuthenticated, // surrounding ws is tolerated
	}
	for hdr, want := range cases {
		got := r.Resolve(context.Background(), staticHeaders{"Authorization": hdr})
		if got.Outcome != want {
			t.Errorf("Authorization=%q: got %v, want %v", hdr, got.Outcome, want)
		}
	}
}

// TestResolver_HeaderSourceCaseInsensitive guards the abstraction that
// lets the gRPC interceptor (lowercased metadata keys) reuse the same
// Resolver as the Gin middleware (canonical MIME headers).
func TestResolver_HeaderSourceCaseInsensitive(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	lower := staticHeaders{"authorization": "Bearer secret"}
	if got := r.Resolve(context.Background(), lower); got.Outcome != OutcomeAuthenticated {
		t.Fatalf("lowercase Authorization should authenticate, got %v", got.Outcome)
	}
}

func TestResolver_WorkspaceHeader_NoRepo_AcceptsVerbatim(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	got := r.Resolve(context.Background(), staticHeaders{
		"Authorization":  "Bearer secret",
		"X-Workspace-ID": "ws-123",
	})
	if got.Outcome != OutcomeAuthenticated {
		t.Fatalf("expected auth, got %v", got.Outcome)
	}
	id, ok := wsctx.FromContext(got.Ctx)
	if !ok || id != "ws-123" {
		t.Fatalf("expected workspace ws-123 in ctx, got %q ok=%v", id, ok)
	}
}

func TestResolver_NoWorkspaceHeader_LeavesCtxUnset(t *testing.T) {
	r := New(&config.AppConfig{APIToken: "secret"}, nil, nil, nil)
	got := r.Resolve(context.Background(), staticHeaders{"Authorization": "Bearer secret"})
	if _, ok := wsctx.FromContext(got.Ctx); ok {
		t.Fatalf("workspace should be unset when no header is sent")
	}
}

func TestHashSecret_StableAndDifferent(t *testing.T) {
	a := HashSecret("hello")
	b := HashSecret("hello")
	if a != b {
		t.Fatalf("HashSecret should be deterministic")
	}
	if HashSecret("hello") == HashSecret("world") {
		t.Fatalf("different inputs should hash differently")
	}
}
