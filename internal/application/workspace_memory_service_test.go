package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	memoryconfigmem "go.orx.me/apps/butter/internal/repo/memoryconfig/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeMem0 is a stand-in mem0 OSS server: it accepts `validKey` (or any
// request when validKey is empty) and records the last X-API-Key it saw.
type fakeMem0 struct {
	srv      *httptest.Server
	validKey string

	mu      sync.Mutex
	lastKey string
	calls   int
}

func newFakeMem0(t *testing.T, validKey string) *fakeMem0 {
	f := &fakeMem0{validKey: validKey}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.lastKey = r.Header.Get("X-API-Key")
		f.calls++
		validKey := f.validKey
		f.mu.Unlock()
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		if validKey != "" && r.Header.Get("X-API-Key") != validKey {
			http.Error(w, `{"detail":"Invalid API key"}`, http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeMem0) setValidKey(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.validKey = key
}

func (f *fakeMem0) seen() (string, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastKey, f.calls
}

type workspaceMemoryFixture struct {
	svc     *WorkspaceMemoryConfigServiceServer
	repo    *memoryconfigmem.Store
	keyring *secretbox.Keyring
}

func newWorkspaceMemoryFixture(t *testing.T) *workspaceMemoryFixture {
	t.Helper()
	wsRepo := workspacememory.New()
	if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: "ws-a", Name: "ws-a", Slug: "ws-a"}); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	for _, m := range []struct{ user, role string }{
		{"owner-user", "owner"},
		{"admin-user", "admin"},
		{"member-user", "member"},
	} {
		if _, err := wsRepo.AddMember(t.Context(), &agentsv1.WorkspaceMember{WorkspaceId: "ws-a", UserId: m.user, Role: m.role}); err != nil {
			t.Fatalf("seed member %s: %v", m.user, err)
		}
	}
	repo := memoryconfigmem.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	svc := NewWorkspaceMemoryConfigServiceServer(repo)
	svc.SetKeyring(keyring)
	svc.SetWorkspaceRepo(wsRepo)
	return &workspaceMemoryFixture{svc: svc, repo: repo, keyring: keyring}
}

func (fx *workspaceMemoryFixture) put(ctx context.Context, baseURL string, enabled bool, apiKey *string) (*agentsv1.PutWorkspaceMemoryConfigResponse, error) {
	resp, err := fx.svc.PutWorkspaceMemoryConfig(ctx, connect.NewRequest(&agentsv1.PutWorkspaceMemoryConfigRequest{
		BaseUrl: baseURL, Enabled: enabled, ApiKey: apiKey,
	}))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

func (fx *workspaceMemoryFixture) storedKey(t *testing.T) string {
	t.Helper()
	conn, err := memoryconn.NewResolver(fx.repo, fx.keyring).Load(t.Context(), "ws-a")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return conn.APIKey
}

func TestWorkspaceMemoryGetWithoutConfigReturnsUnset(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	resp, err := fx.svc.GetWorkspaceMemoryConfig(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceMemoryConfigRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.Msg.GetConfig() != nil {
		t.Fatalf("config = %v, want unset", resp.Msg.GetConfig())
	}
}

func TestWorkspaceMemoryPutStoresEncryptedKeyAndMembersCanRead(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	mem0 := newFakeMem0(t, "m0sk_good")

	got, err := fx.put(ownerCtx(), mem0.srv.URL+"/", true, proto.String("m0sk_good"))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if got.GetWarning() != "" {
		t.Fatalf("warning = %q, want none", got.GetWarning())
	}
	cfg := got.GetConfig()
	if cfg.GetBaseUrl() != mem0.srv.URL || !cfg.GetEnabled() || !cfg.GetCredentialSet() || cfg.GetWorkspaceId() != "ws-a" {
		t.Fatalf("config = %v", cfg)
	}
	if key, calls := mem0.seen(); key != "m0sk_good" || calls != 1 {
		t.Fatalf("probe key = %q calls = %d", key, calls)
	}
	if fx.storedKey(t) != "m0sk_good" {
		t.Fatal("key not stored")
	}
	cred, err := fx.repo.GetCredential(t.Context(), "ws-a")
	if err != nil || cred.Ciphertext == "" || strings.Contains(cred.Ciphertext, "m0sk_good") {
		t.Fatalf("stored credential = %+v, %v; want ciphertext", cred, err)
	}

	resp, err := fx.svc.GetWorkspaceMemoryConfig(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceMemoryConfigRequest{}))
	if err != nil {
		t.Fatalf("member Get: %v", err)
	}
	if !resp.Msg.GetConfig().GetCredentialSet() || resp.Msg.GetConfig().GetBaseUrl() != mem0.srv.URL {
		t.Fatalf("member view = %v", resp.Msg.GetConfig())
	}
}

func TestWorkspaceMemoryPutRejectsRefusedKey(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	mem0 := newFakeMem0(t, "m0sk_good")

	_, err := fx.put(ownerCtx(), mem0.srv.URL, true, proto.String("m0sk_bad"))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, want FailedPrecondition", code)
	}
	if _, err := fx.repo.Get(t.Context(), "ws-a"); err == nil {
		t.Fatal("config saved despite a refused key")
	}
}

func TestWorkspaceMemoryPutUnreachableServerSavesWithWarning(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	mem0 := newFakeMem0(t, "")
	unreachable := mem0.srv.URL
	mem0.srv.Close()

	got, err := fx.put(ownerCtx(), unreachable, true, proto.String("m0sk_any"))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if got.GetWarning() == "" {
		t.Fatal("no warning for an unreachable server")
	}
	if !got.GetConfig().GetEnabled() || fx.storedKey(t) != "m0sk_any" {
		t.Fatalf("config not saved: %v", got.GetConfig())
	}
}

func TestWorkspaceMemoryPutDisabledSkipsProbe(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	mem0 := newFakeMem0(t, "m0sk_good")

	got, err := fx.put(ownerCtx(), mem0.srv.URL, false, proto.String("m0sk_bad"))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if got.GetConfig().GetEnabled() || got.GetWarning() != "" {
		t.Fatalf("resp = %v", got)
	}
	if _, calls := mem0.seen(); calls != 0 {
		t.Fatalf("disabled config probed %d times", calls)
	}
}

func TestWorkspaceMemoryPutKeepsRotatesAndClearsKey(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	mem0 := newFakeMem0(t, "")

	if _, err := fx.put(ownerCtx(), mem0.srv.URL, true, proto.String("key-1")); err != nil {
		t.Fatalf("set: %v", err)
	}
	// Absent keeps the stored key, and the probe uses it.
	got, err := fx.put(ownerCtx(), mem0.srv.URL, true, nil)
	if err != nil || !got.GetConfig().GetCredentialSet() || fx.storedKey(t) != "key-1" {
		t.Fatalf("keep: %v, %v", got, err)
	}
	if key, _ := mem0.seen(); key != "key-1" {
		t.Fatalf("keep probe used key %q, want the stored key", key)
	}
	// A value rotates.
	if _, err := fx.put(ownerCtx(), mem0.srv.URL, true, proto.String("key-2")); err != nil || fx.storedKey(t) != "key-2" {
		t.Fatalf("rotate: %v", err)
	}
	// Empty clears.
	got, err = fx.put(ownerCtx(), mem0.srv.URL, true, proto.String(""))
	if err != nil || got.GetConfig().GetCredentialSet() {
		t.Fatalf("clear: %v, %v", got, err)
	}
	if key, _ := mem0.seen(); key != "" {
		t.Fatalf("clear probe sent key %q", key)
	}
}

func TestWorkspaceMemoryPutValidatesBaseURL(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	for _, baseURL := range []string{"", "mem0.example.com", "ftp://mem0.example.com", "https://"} {
		_, err := fx.put(ownerCtx(), baseURL, false, nil)
		if code := connectCode(t, err); code != connect.CodeInvalidArgument {
			t.Fatalf("base_url %q: code = %v, want InvalidArgument", baseURL, code)
		}
	}
}

func TestWorkspaceMemoryMutationsRequireManageRole(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	mem0 := newFakeMem0(t, "")

	if _, err := fx.put(memberCtx(), mem0.srv.URL, true, nil); connectCode(t, err) != connect.CodePermissionDenied {
		t.Fatalf("member Put = %v, want PermissionDenied", err)
	}
	if _, err := fx.put(wsAdminCtx(), mem0.srv.URL, true, nil); err != nil {
		t.Fatalf("workspace admin Put: %v", err)
	}
	_, err := fx.svc.DeleteWorkspaceMemoryConfig(memberCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceMemoryConfigRequest{}))
	if code := connectCode(t, err); code != connect.CodePermissionDenied {
		t.Fatalf("member Delete code = %v, want PermissionDenied", code)
	}
	// A global admin bypasses membership.
	globalAdmin := ctxAs("root", "admin", "ws-a")
	if _, err := fx.svc.DeleteWorkspaceMemoryConfig(globalAdmin, connect.NewRequest(&agentsv1.DeleteWorkspaceMemoryConfigRequest{})); err != nil {
		t.Fatalf("global admin Delete: %v", err)
	}
	// Non-members see NotFound.
	if _, err := fx.put(ctxAs("stranger", "user", "ws-a"), mem0.srv.URL, true, nil); connectCode(t, err) != connect.CodeNotFound {
		t.Fatalf("non-member Put = %v, want NotFound", err)
	}
}

func TestWorkspaceMemoryDelete(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	mem0 := newFakeMem0(t, "")

	_, err := fx.svc.DeleteWorkspaceMemoryConfig(ownerCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceMemoryConfigRequest{}))
	if code := connectCode(t, err); code != connect.CodeNotFound {
		t.Fatalf("Delete without config: code = %v, want NotFound", code)
	}
	if _, err := fx.put(ownerCtx(), mem0.srv.URL, true, proto.String("k")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, err := fx.svc.DeleteWorkspaceMemoryConfig(ownerCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceMemoryConfigRequest{})); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := memoryconn.NewResolver(fx.repo, fx.keyring).Load(t.Context(), "ws-a"); err != memoryconn.ErrNotConfigured {
		t.Fatalf("Load after delete = %v, want ErrNotConfigured", err)
	}
}

func TestWorkspaceMemoryTestConnection(t *testing.T) {
	fx := newWorkspaceMemoryFixture(t)
	mem0 := newFakeMem0(t, "m0sk_good")

	_, err := fx.svc.TestWorkspaceMemoryConnection(memberCtx(), connect.NewRequest(&agentsv1.TestWorkspaceMemoryConnectionRequest{}))
	if code := connectCode(t, err); code != connect.CodeNotFound {
		t.Fatalf("without config: code = %v, want NotFound", code)
	}

	if _, err := fx.put(ownerCtx(), mem0.srv.URL, true, proto.String("m0sk_good")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	resp, err := fx.svc.TestWorkspaceMemoryConnection(memberCtx(), connect.NewRequest(&agentsv1.TestWorkspaceMemoryConnectionRequest{}))
	if err != nil || !resp.Msg.GetOk() || resp.Msg.GetError() != "" {
		t.Fatalf("Test = %v, %v; want ok", resp, err)
	}

	// The key is revoked on the server side: the probe reports it rather
	// than failing the RPC.
	mem0.setValidKey("m0sk_rotated")
	resp, err = fx.svc.TestWorkspaceMemoryConnection(memberCtx(), connect.NewRequest(&agentsv1.TestWorkspaceMemoryConnectionRequest{}))
	if err != nil || resp.Msg.GetOk() || !strings.Contains(resp.Msg.GetError(), "401") {
		t.Fatalf("Test with revoked key = %v, %v", resp, err)
	}
}
