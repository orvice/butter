package account

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.orx.me/apps/butter/desktop/internal/authapi"
	"go.orx.me/apps/butter/desktop/internal/credentials"
	"go.orx.me/apps/butter/desktop/internal/endpoint"
)

type fakeAPI struct {
	login      func(context.Context) (authapi.Session, error)
	me         func(context.Context, string) (authapi.User, error)
	workspaces func(context.Context, string) ([]authapi.Workspace, error)
	logout     func(context.Context, string) error

	mu          sync.Mutex
	revoked     []string
	startedFlag atomic.Bool
}

func (f *fakeAPI) Login(ctx context.Context, _, _ string) (authapi.Session, error) {
	f.startedFlag.Store(true)
	if f.login == nil {
		return authapi.Session{}, nil
	}
	return f.login(ctx)
}

func (f *fakeAPI) started() bool { return f.startedFlag.Load() }

func (f *fakeAPI) Me(ctx context.Context, token string) (authapi.User, error) {
	if f.me == nil {
		return authapi.User{}, nil
	}
	return f.me(ctx, token)
}

func (f *fakeAPI) Workspaces(ctx context.Context, token string) ([]authapi.Workspace, error) {
	if f.workspaces == nil {
		return nil, nil
	}
	return f.workspaces(ctx, token)
}

func (f *fakeAPI) Logout(ctx context.Context, token string) error {
	f.mu.Lock()
	f.revoked = append(f.revoked, token)
	f.mu.Unlock()
	if f.logout != nil {
		return f.logout(ctx, token)
	}
	return nil
}

func (f *fakeAPI) revokedTokens() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.revoked...)
}

var testWorkspaces = []authapi.Workspace{
	{ID: "w1", Name: "Default", Slug: "default"},
	{ID: "w2", Name: "Team", Slug: "team"},
}

func signedInManager(t *testing.T) (*Manager, *credentials.Memory, *fakeAPI) {
	t.Helper()
	store := credentials.NewMemory()
	api := &fakeAPI{
		login: func(context.Context) (authapi.Session, error) {
			return authapi.Session{
				Token:      "tok-1",
				User:       authapi.User{ID: "u1", Username: "lion"},
				Workspaces: testWorkspaces,
				ExpiresAt:  time.Now().Add(time.Hour),
			}, nil
		},
		me: func(context.Context, string) (authapi.User, error) {
			return authapi.User{ID: "u1", Username: "lion"}, nil
		},
		workspaces: func(context.Context, string) ([]authapi.Workspace, error) {
			return testWorkspaces, nil
		},
	}
	m := New(store, func(string) API { return api })
	if _, err := m.UseServer("butter.example.com"); err != nil {
		t.Fatalf("UseServer: %v", err)
	}
	return m, store, api
}

func TestSignInStoresTokenAndState(t *testing.T) {
	m, store, _ := signedInManager(t)
	st, err := m.SignIn(context.Background(), " lion ", "secret")
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	if !st.SignedIn() || st.User.Username != "lion" || len(st.Workspaces) != 2 {
		t.Fatalf("state = %+v", st)
	}
	if st.WorkspaceID != "" {
		t.Fatalf("a two-workspace account must choose one, got %q", st.WorkspaceID)
	}
	tok, ok := m.Token()
	if !ok || tok != "tok-1" {
		t.Fatalf("token = %q, %v", tok, ok)
	}
	if keys := store.Keys(); len(keys) != 1 || !strings.HasPrefix(keys[0], "butter-session-") {
		t.Fatalf("stored keys = %v", keys)
	}
}

func TestSignInPicksTheOnlyWorkspace(t *testing.T) {
	store := credentials.NewMemory()
	api := &fakeAPI{
		login: func(context.Context) (authapi.Session, error) {
			return authapi.Session{
				Token:      "tok-1",
				User:       authapi.User{ID: "u1", Username: "lion"},
				Workspaces: testWorkspaces[:1],
			}, nil
		},
	}
	m := New(store, func(string) API { return api })
	if _, err := m.UseServer("butter.example.com"); err != nil {
		t.Fatal(err)
	}
	st, err := m.SignIn(context.Background(), "lion", "secret")
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	if st.WorkspaceID != "w1" {
		t.Fatalf("workspace = %q, want w1", st.WorkspaceID)
	}
}

func TestSignInFailureKeepsSignedOut(t *testing.T) {
	store := credentials.NewMemory()
	api := &fakeAPI{
		login: func(context.Context) (authapi.Session, error) {
			return authapi.Session{}, &authapi.Error{Kind: authapi.KindInvalidCredentials, Op: "login", Err: errors.New("401")}
		},
	}
	m := New(store, func(string) API { return api })
	if _, err := m.UseServer("butter.example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.SignIn(context.Background(), "lion", "wrong"); authapi.KindOf(err) != authapi.KindInvalidCredentials {
		t.Fatalf("err = %v", err)
	}
	if m.SignedIn() {
		t.Fatal("signed in after a rejected password")
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("stored keys = %v", keys)
	}
}

func TestUseServerClearsPreviousToken(t *testing.T) {
	m, store, _ := signedInManager(t)
	if _, err := m.SignIn(context.Background(), "lion", "secret"); err != nil {
		t.Fatal(err)
	}
	st, err := m.UseServer("other.example.com")
	if err != nil {
		t.Fatalf("UseServer: %v", err)
	}
	if st.SignedIn() {
		t.Fatalf("still signed in after switching server: %+v", st)
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("a token survived the server switch: %v", keys)
	}
	if m.SignedIn() {
		t.Fatal("token still in memory after switching server")
	}
	if st.Server != "https://other.example.com" {
		t.Fatalf("server = %q", st.Server)
	}
}

func TestUseServerRejectsInvalidAddress(t *testing.T) {
	m := New(credentials.NewMemory(), nil)
	if _, err := m.UseServer("ftp://x"); err == nil {
		t.Fatal("accepted a non-http scheme")
	}
	if _, ok := m.Server(); ok {
		t.Fatal("a rejected address became the current server")
	}
}

func TestLateSignInFromAnotherContextIsDropped(t *testing.T) {
	store := credentials.NewMemory()
	gate := make(chan struct{})
	api := &fakeAPI{
		login: func(context.Context) (authapi.Session, error) {
			<-gate
			return authapi.Session{Token: "tok-late", User: authapi.User{ID: "u1", Username: "lion"}}, nil
		},
	}
	m := New(store, func(string) API { return api })
	if _, err := m.UseServer("first.example.com"); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := m.SignIn(context.Background(), "lion", "secret")
		done <- err
	}()
	// Let the login start before the context moves.
	for !api.started() {
		time.Sleep(time.Millisecond)
	}
	if _, err := m.UseServer("second.example.com"); err != nil {
		t.Fatal(err)
	}
	close(gate)
	if err := <-done; !errors.Is(err, ErrStale) {
		t.Fatalf("err = %v, want ErrStale", err)
	}
	if m.SignedIn() {
		t.Fatal("a session from the previous server was applied")
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("stale token stored: %v", keys)
	}
	if got := api.revokedTokens(); len(got) != 1 || got[0] != "tok-late" {
		t.Fatalf("revoked = %v, want the unused token revoked", got)
	}
}

func TestRestoreWithoutSession(t *testing.T) {
	m := New(credentials.NewMemory(), nil)
	if _, err := m.UseServer("butter.example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Restore(context.Background()); !errors.Is(err, ErrNoSession) {
		t.Fatalf("err = %v, want ErrNoSession", err)
	}
}

func TestRestoreExpiredSessionForgetsTheToken(t *testing.T) {
	store := credentials.NewMemory()
	api := &fakeAPI{
		me: func(context.Context, string) (authapi.User, error) {
			return authapi.User{}, &authapi.Error{Kind: authapi.KindSessionExpired, Op: "me", Err: errors.New("401")}
		},
	}
	m := New(store, func(string) API { return api })
	if _, err := m.UseServer("butter.example.com"); err != nil {
		t.Fatal(err)
	}
	st := m.Snapshot()
	if err := store.Set("butter-session-"+serverKey(t, st.Server), []byte("stale")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Restore(context.Background()); authapi.KindOf(err) != authapi.KindSessionExpired {
		t.Fatalf("err = %v, want session expired", err)
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("an expired token survived: %v", keys)
	}
	if m.SignedIn() {
		t.Fatal("signed in with an expired token")
	}
}

func TestRestoreUsesStoredToken(t *testing.T) {
	m, store, api := signedInManager(t)
	st0 := m.Snapshot()
	if err := store.Set("butter-session-"+serverKey(t, st0.Server), []byte("tok-restored")); err != nil {
		t.Fatal(err)
	}
	var seen string
	api.me = func(_ context.Context, token string) (authapi.User, error) {
		seen = token
		return authapi.User{ID: "u1", Username: "lion"}, nil
	}
	st, err := m.Restore(context.Background())
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if seen != "tok-restored" {
		t.Fatalf("token = %q", seen)
	}
	if !st.SignedIn() || len(st.Workspaces) != 2 {
		t.Fatalf("state = %+v", st)
	}
}

func TestSelectWorkspace(t *testing.T) {
	m, _, _ := signedInManager(t)
	if _, err := m.SignIn(context.Background(), "lion", "secret"); err != nil {
		t.Fatal(err)
	}
	st, err := m.SelectWorkspace("w2")
	if err != nil {
		t.Fatalf("SelectWorkspace: %v", err)
	}
	if st.WorkspaceID != "w2" {
		t.Fatalf("workspace = %q", st.WorkspaceID)
	}
	if _, err := m.SelectWorkspace("nope"); !errors.Is(err, ErrUnknownWorkspace) {
		t.Fatalf("err = %v, want ErrUnknownWorkspace", err)
	}
	if got := m.Snapshot().WorkspaceID; got != "w2" {
		t.Fatalf("rejected selection changed the workspace to %q", got)
	}
}

func TestSignOutRevokesAndForgets(t *testing.T) {
	m, store, api := signedInManager(t)
	if _, err := m.SignIn(context.Background(), "lion", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := m.SignOut(context.Background()); err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	if m.SignedIn() {
		t.Fatal("still signed in after SignOut")
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("token left in the store: %v", keys)
	}
	if got := api.revokedTokens(); len(got) != 1 || got[0] != "tok-1" {
		t.Fatalf("revoked = %v", got)
	}
	if st := m.Snapshot(); st.Server == "" || st.SignedIn() {
		t.Fatalf("state = %+v, want the server kept and no account", st)
	}
}

func TestSignOutSurvivesAnUnreachableServer(t *testing.T) {
	m, store, api := signedInManager(t)
	if _, err := m.SignIn(context.Background(), "lion", "secret"); err != nil {
		t.Fatal(err)
	}
	api.logout = func(context.Context, string) error {
		return &authapi.Error{Kind: authapi.KindNetwork, Op: "logout", Err: errors.New("offline")}
	}
	if err := m.SignOut(context.Background()); err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("token left in the store: %v", keys)
	}
}

func serverKey(t *testing.T, raw string) string {
	t.Helper()
	srv, err := endpoint.Parse(raw)
	if err != nil {
		t.Fatalf("Parse(%q): %v", raw, err)
	}
	_ = srv
	return srv.Key()
}

func TestSignInFillsWorkspacesForAdminsWithoutMemberships(t *testing.T) {
	store := credentials.NewMemory()
	api := &fakeAPI{
		login: func(context.Context) (authapi.Session, error) {
			return authapi.Session{Token: "tok-1", User: authapi.User{ID: "u1", Username: "root", Role: "admin"}}, nil
		},
		workspaces: func(context.Context, string) ([]authapi.Workspace, error) {
			return testWorkspaces, nil
		},
	}
	m := New(store, func(string) API { return api })
	if _, err := m.UseServer("butter.example.com"); err != nil {
		t.Fatal(err)
	}
	st, err := m.SignIn(context.Background(), "root", "secret")
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	if len(st.Workspaces) != 2 {
		t.Fatalf("workspaces = %+v, want the server's list", st.Workspaces)
	}
}

func TestSignInKeepsGoingWhenTheWorkspaceListFails(t *testing.T) {
	store := credentials.NewMemory()
	api := &fakeAPI{
		login: func(context.Context) (authapi.Session, error) {
			return authapi.Session{Token: "tok-1", User: authapi.User{ID: "u1", Username: "root"}}, nil
		},
		workspaces: func(context.Context, string) ([]authapi.Workspace, error) {
			return nil, &authapi.Error{Kind: authapi.KindServer, Op: "list workspaces", Err: errors.New("500")}
		},
	}
	m := New(store, func(string) API { return api })
	if _, err := m.UseServer("butter.example.com"); err != nil {
		t.Fatal(err)
	}
	st, err := m.SignIn(context.Background(), "root", "secret")
	if err != nil {
		t.Fatalf("SignIn blocked by a workspace listing: %v", err)
	}
	if !st.SignedIn() {
		t.Fatalf("state = %+v", st)
	}
}
