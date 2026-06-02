package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/auth/provider"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/oauthstate"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeAuthRepo is a minimal in-memory auth.Repository for OAuth tests. Only
// the methods CompleteOAuthFlow / upsertOAuthUser need are implemented; the
// rest panic so an unexpected use is loud.
type fakeAuthRepo struct {
	mu       sync.Mutex
	users    map[string]*agentsv1.User
	byExt    map[string]*agentsv1.User
	sessions map[string]*auth.Session
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		users:    make(map[string]*agentsv1.User),
		byExt:    make(map[string]*agentsv1.User),
		sessions: make(map[string]*auth.Session),
	}
}

func extKey(provider, externalID string) string { return provider + "::" + externalID }

func (f *fakeAuthRepo) CreateUser(_ context.Context, user *agentsv1.User, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.GetUsername() == user.GetUsername() {
			return auth.ErrUserAlreadyExists
		}
	}
	f.users[user.GetId()] = user
	if user.GetProvider() != "" && user.GetExternalId() != "" {
		f.byExt[extKey(user.GetProvider(), user.GetExternalId())] = user
	}
	return nil
}

func (f *fakeAuthRepo) FindUserByExternalID(_ context.Context, provider, externalID string) (*agentsv1.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byExt[extKey(provider, externalID)]
	if !ok {
		return nil, auth.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeAuthRepo) CreateSession(_ context.Context, session *auth.Session) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sessions[session.ID] = session
	return nil
}

func (f *fakeAuthRepo) GetUser(_ context.Context, id string) (*agentsv1.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return nil, auth.ErrUserNotFound
	}
	return u, nil
}

// Unused methods — panic so misuse is loud.
func (f *fakeAuthRepo) EnsureIndexes(context.Context) error { panic("not implemented") }
func (f *fakeAuthRepo) CountUsers(context.Context) (int64, error) {
	panic("not implemented")
}
func (f *fakeAuthRepo) ListUsers(context.Context) ([]*agentsv1.User, error) {
	panic("not implemented")
}
func (f *fakeAuthRepo) UpdateUserPassword(context.Context, string, string, time.Time) (*agentsv1.User, error) {
	panic("not implemented")
}
func (f *fakeAuthRepo) UpdateUserProfile(context.Context, string, string, *string, time.Time) (*agentsv1.User, error) {
	panic("not implemented")
}
func (f *fakeAuthRepo) SetUserDisabled(context.Context, string, bool, time.Time) (*agentsv1.User, error) {
	panic("not implemented")
}
func (f *fakeAuthRepo) FindUserByUsername(context.Context, string) (*agentsv1.User, string, error) {
	panic("not implemented")
}
func (f *fakeAuthRepo) FindUserByID(context.Context, string) (*agentsv1.User, string, error) {
	panic("not implemented")
}
func (f *fakeAuthRepo) LookupSession(context.Context, string, time.Time) (*auth.Session, *agentsv1.User, error) {
	panic("not implemented")
}
func (f *fakeAuthRepo) TouchSession(context.Context, string, time.Time) error {
	panic("not implemented")
}
func (f *fakeAuthRepo) RevokeSession(context.Context, string) error { panic("not implemented") }

// stubProvider returns canned claims.
type stubProvider struct {
	name        string
	displayName string
	authorize   string
	claims      *provider.Claims
	exchangeErr error
	gotCode     string
}

func (s *stubProvider) Name() string        { return s.name }
func (s *stubProvider) DisplayName() string { return s.displayName }
func (s *stubProvider) AuthorizeURL(state string) string {
	return s.authorize + "?state=" + state
}
func (s *stubProvider) Exchange(_ context.Context, code string) (*provider.Claims, error) {
	s.gotCode = code
	if s.exchangeErr != nil {
		return nil, s.exchangeErr
	}
	return s.claims, nil
}

func newServerWithOAuth(t *testing.T, stub provider.Provider) (*AuthServiceServer, *fakeAuthRepo) {
	t.Helper()
	repo := newFakeAuthRepo()
	srv := NewAuthServiceServer(repo, time.Hour)
	reg := provider.NewRegistry()
	reg.Register(stub)
	srv.SetProviderRegistry(reg)
	srv.SetOAuthStateRepo(oauthstateMemory())
	return srv, repo
}

// oauthstateMemory wraps the in-memory store from the repo package so tests
// don't depend on the mongo backend.
func oauthstateMemory() oauthstate.Repository {
	return &memoryStateRepo{m: make(map[string]oauthstate.Entry)}
}

type memoryStateRepo struct {
	mu sync.Mutex
	m  map[string]oauthstate.Entry
}

func (s *memoryStateRepo) EnsureIndexes(context.Context) error { return nil }
func (s *memoryStateRepo) Create(_ context.Context, e *oauthstate.Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[e.State] = *e
	return nil
}
func (s *memoryStateRepo) Consume(_ context.Context, state string, now time.Time) (*oauthstate.Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.m[state]
	if !ok {
		return nil, oauthstate.ErrNotFound
	}
	delete(s.m, state)
	if !e.ExpiresAt.After(now) {
		return nil, oauthstate.ErrNotFound
	}
	return &e, nil
}

func TestAuthService_OAuth_Flow_CreatesAndReusesUser(t *testing.T) {
	stub := &stubProvider{
		name:        "github",
		displayName: "GitHub",
		authorize:   "https://example.test/authorize",
		claims: &provider.Claims{
			Provider:   "github",
			ExternalID: "42",
			Login:      "alice",
			Name:       "Alice",
			Email:      "alice@example.com",
			AvatarURL:  "https://av/alice",
		},
	}
	srv, repo := newServerWithOAuth(t, stub)
	ctx := context.Background()

	// Begin: state should be stored and authorize URL echo state back.
	beg, err := srv.BeginOAuthFlow(ctx, connect.NewRequest(&agentsv1.BeginOAuthFlowRequest{Provider: "github"}))
	if err != nil {
		t.Fatalf("BeginOAuthFlow: %v", err)
	}
	if beg.Msg.GetState() == "" {
		t.Fatal("BeginOAuthFlow returned empty state")
	}
	if !strings.Contains(beg.Msg.GetAuthorizeUrl(), beg.Msg.GetState()) {
		t.Errorf("authorize_url missing state: %s", beg.Msg.GetAuthorizeUrl())
	}

	// Complete: exchanges code, creates user, issues session.
	resp, err := srv.CompleteOAuthFlow(ctx, connect.NewRequest(&agentsv1.CompleteOAuthFlowRequest{
		Provider: "github",
		Code:     "the-code",
		State:    beg.Msg.GetState(),
	}))
	if err != nil {
		t.Fatalf("CompleteOAuthFlow: %v", err)
	}
	if resp.Msg.GetToken() == "" {
		t.Error("LoginResponse missing token")
	}
	if resp.Msg.GetUser().GetExternalId() != "42" {
		t.Errorf("user external id = %q, want 42", resp.Msg.GetUser().GetExternalId())
	}
	if resp.Msg.GetUser().GetProvider() != "github" {
		t.Errorf("user provider = %q, want github", resp.Msg.GetUser().GetProvider())
	}
	if !strings.HasPrefix(resp.Msg.GetUser().GetUsername(), "gith_alice") {
		t.Errorf("username = %q, want prefix gith_alice", resp.Msg.GetUser().GetUsername())
	}
	if stub.gotCode != "the-code" {
		t.Errorf("provider got code %q, want the-code", stub.gotCode)
	}
	firstUserID := resp.Msg.GetUser().GetId()

	// Re-using the same provider identity must return the SAME user record.
	beg2, err := srv.BeginOAuthFlow(ctx, connect.NewRequest(&agentsv1.BeginOAuthFlowRequest{Provider: "github"}))
	if err != nil {
		t.Fatalf("BeginOAuthFlow second: %v", err)
	}
	resp2, err := srv.CompleteOAuthFlow(ctx, connect.NewRequest(&agentsv1.CompleteOAuthFlowRequest{
		Provider: "github",
		Code:     "another-code",
		State:    beg2.Msg.GetState(),
	}))
	if err != nil {
		t.Fatalf("CompleteOAuthFlow second: %v", err)
	}
	if resp2.Msg.GetUser().GetId() != firstUserID {
		t.Errorf("second login created new user (%s vs %s) — should reuse existing", resp2.Msg.GetUser().GetId(), firstUserID)
	}
	if got := len(repo.users); got != 1 {
		t.Errorf("expected exactly 1 user after two logins, got %d", got)
	}
}

func TestAuthService_OAuth_RejectsReplayedState(t *testing.T) {
	stub := &stubProvider{
		name:      "github",
		authorize: "https://example.test/authorize",
		claims:    &provider.Claims{Provider: "github", ExternalID: "1", Login: "u"},
	}
	srv, _ := newServerWithOAuth(t, stub)
	ctx := context.Background()

	beg, err := srv.BeginOAuthFlow(ctx, connect.NewRequest(&agentsv1.BeginOAuthFlowRequest{Provider: "github"}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.CompleteOAuthFlow(ctx, connect.NewRequest(&agentsv1.CompleteOAuthFlowRequest{Provider: "github", Code: "c", State: beg.Msg.GetState()})); err != nil {
		t.Fatalf("first complete: %v", err)
	}
	// Replay the same state — must be rejected.
	_, err = srv.CompleteOAuthFlow(ctx, connect.NewRequest(&agentsv1.CompleteOAuthFlowRequest{Provider: "github", Code: "c", State: beg.Msg.GetState()}))
	if err == nil {
		t.Fatal("expected error on state replay, got nil")
	}
}

func TestAuthService_OAuth_RejectsMismatchedProvider(t *testing.T) {
	stub := &stubProvider{
		name:      "github",
		authorize: "https://example.test/a",
		claims:    &provider.Claims{Provider: "github", ExternalID: "1", Login: "u"},
	}
	srv, _ := newServerWithOAuth(t, stub)
	ctx := context.Background()
	beg, err := srv.BeginOAuthFlow(ctx, connect.NewRequest(&agentsv1.BeginOAuthFlowRequest{Provider: "github"}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = srv.CompleteOAuthFlow(ctx, connect.NewRequest(&agentsv1.CompleteOAuthFlowRequest{Provider: "google", Code: "c", State: beg.Msg.GetState()}))
	if err == nil {
		t.Fatal("expected provider-mismatch rejection, got nil")
	}
}

func TestAuthService_OAuth_DisabledWhenUnconfigured(t *testing.T) {
	srv := NewAuthServiceServer(newFakeAuthRepo(), time.Hour)
	_, err := srv.BeginOAuthFlow(context.Background(), connect.NewRequest(&agentsv1.BeginOAuthFlowRequest{Provider: "github"}))
	if err == nil {
		t.Fatal("expected error when oauth not configured")
	}
}

func TestAuthService_OAuth_PropagatesExchangeError(t *testing.T) {
	stub := &stubProvider{
		name:        "github",
		authorize:   "https://example.test/a",
		exchangeErr: errors.New("github boom"),
	}
	srv, _ := newServerWithOAuth(t, stub)
	ctx := context.Background()
	beg, err := srv.BeginOAuthFlow(ctx, connect.NewRequest(&agentsv1.BeginOAuthFlowRequest{Provider: "github"}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = srv.CompleteOAuthFlow(ctx, connect.NewRequest(&agentsv1.CompleteOAuthFlowRequest{Provider: "github", Code: "c", State: beg.Msg.GetState()}))
	if err == nil {
		t.Fatal("expected exchange error to surface")
	}
}
