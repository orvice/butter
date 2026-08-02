package application

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/auth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/v2/session"
)

// stubTitleStore implements SessionTitleStore for tests.
type stubTitleStore struct {
	setErr    error
	setCalled int
	lastTitle string
	notFound  bool
}

func (s *stubTitleStore) SetSessionTitle(_ context.Context, appName, userID, sessionID, title string) (*agentsv1.SessionInfo, error) {
	s.setCalled++
	s.lastTitle = title
	if s.setErr != nil {
		return nil, s.setErr
	}
	if s.notFound {
		return nil, fmt.Errorf("%w: %s/%s/%s", ErrSessionNotFound, appName, userID, sessionID)
	}
	return &agentsv1.SessionInfo{
		SessionId: sessionID,
		AppName:   appName,
		UserId:    userID,
		Title:     title,
	}, nil
}

func newTitleTestService(store *stubTitleStore) *SessionServiceServer {
	svc := NewSessionServiceServer()
	svc.titleStore = store
	return svc
}

func titleAdminCtx() context.Context {
	return auth.WithAdmin(context.Background())
}

func titleUserCtx(id, role string) context.Context {
	user := &agentsv1.User{Id: id, Role: role}
	return auth.WithAuthenticated(context.Background(), user, &auth.Session{})
}

// --- Validation tests ---

func TestUpdateSessionTitle_RequiredFields(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	tests := []struct {
		name string
		req  *agentsv1.UpdateSessionTitleRequest
		want string
	}{
		{"missing app_name", &agentsv1.UpdateSessionTitleRequest{UserId: "u1", SessionId: "s1", Title: "hi"}, "app_name"},
		{"missing user_id", &agentsv1.UpdateSessionTitleRequest{AppName: "web", SessionId: "s1", Title: "hi"}, "user_id"},
		{"missing session_id", &agentsv1.UpdateSessionTitleRequest{AppName: "web", UserId: "u1", Title: "hi"}, "session_id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpdateSessionTitle(titleAdminCtx(), connect.NewRequest(tt.req))
			assertInvalidArgument(t, err)
			if store.setCalled > 0 {
				t.Fatal("store must not be called on validation failure")
			}
		})
	}
}

func TestUpdateSessionTitle_BlankTitleRejected(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	tests := []string{"", "   ", "\n\t\r"}
	for _, title := range tests {
		t.Run(fmt.Sprintf("title=%q", title), func(t *testing.T) {
			_, err := svc.UpdateSessionTitle(titleAdminCtx(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
				AppName: "web", UserId: "u1", SessionId: "s1", Title: title,
			}))
			assertInvalidArgument(t, err)
		})
	}
}

func TestUpdateSessionTitle_TitleTooLong(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	longTitle := strings.Repeat("あ", 101) // 101 code points
	_, err := svc.UpdateSessionTitle(titleAdminCtx(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "s1", Title: longTitle,
	}))
	assertInvalidArgument(t, err)
}

func TestUpdateSessionTitle_TitleExactly100CodePoints(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	exactTitle := strings.Repeat("x", 100)
	resp, err := svc.UpdateSessionTitle(titleAdminCtx(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "s1", Title: exactTitle,
	}))
	if err != nil {
		t.Fatalf("100-code-point title should be accepted: %v", err)
	}
	if resp.Msg.GetSession().GetTitle() != exactTitle {
		t.Fatalf("title mismatch: got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestUpdateSessionTitle_TitleNormalized(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	_, err := svc.UpdateSessionTitle(titleAdminCtx(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "s1", Title: "  hello\nworld\r\nfoo  ",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.lastTitle != "hello world foo" {
		t.Fatalf("expected normalized title %q, got %q", "hello world foo", store.lastTitle)
	}
}

// --- Authorization tests ---

func TestUpdateSessionTitle_OwnerCanUpdate(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	ctx := titleUserCtx("user-42", "user")
	resp, err := svc.UpdateSessionTitle(ctx, connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "user-42", SessionId: "s1", Title: "My Chat",
	}))
	if err != nil {
		t.Fatalf("owner should be allowed: %v", err)
	}
	if resp.Msg.GetSession().GetTitle() != "My Chat" {
		t.Fatalf("title mismatch")
	}
}

func TestUpdateSessionTitle_NonOwnerRejected(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	ctx := titleUserCtx("user-42", "user")
	_, err := svc.UpdateSessionTitle(ctx, connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "other-user", SessionId: "s1", Title: "hijack",
	}))
	if err == nil {
		t.Fatal("expected PermissionDenied for non-owner")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodePermissionDenied {
		t.Fatalf("expected CodePermissionDenied, got %v", err)
	}
}

func TestUpdateSessionTitle_AdminCanUpdateAnyone(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	ctx := titleAdminCtx()
	resp, err := svc.UpdateSessionTitle(ctx, connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "someone-else", SessionId: "s1", Title: "Admin Title",
	}))
	if err != nil {
		t.Fatalf("admin should be allowed: %v", err)
	}
	if resp.Msg.GetSession().GetTitle() != "Admin Title" {
		t.Fatalf("title mismatch")
	}
}

func TestUpdateSessionTitle_UnauthenticatedRejected(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)

	_, err := svc.UpdateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "s1", Title: "sneaky",
	}))
	if err == nil {
		t.Fatal("expected Unauthenticated error")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeUnauthenticated {
		t.Fatalf("expected CodeUnauthenticated, got %v", err)
	}
}

// --- Session not found ---

func TestUpdateSessionTitle_SessionNotFound(t *testing.T) {
	store := &stubTitleStore{notFound: true}
	svc := newTitleTestService(store)

	_, err := svc.UpdateSessionTitle(titleAdminCtx(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "nonexistent", Title: "hi",
	}))
	if err == nil {
		t.Fatal("expected NotFound error")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", err)
	}
}

// --- Title store not available ---

func TestUpdateSessionTitle_StoreNotAvailable(t *testing.T) {
	svc := NewSessionServiceServer() // no title store set

	_, err := svc.UpdateSessionTitle(titleAdminCtx(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "s1", Title: "hi",
	}))
	if err == nil {
		t.Fatal("expected FailedPrecondition error")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("expected CodeFailedPrecondition, got %v", err)
	}
}

// --- Effective title precedence tests ---

// fakeSession implements session.Session for testing effectiveTitle.
type fakeSession struct {
	id    string
	state *fakeState
	title string
}

func (s *fakeSession) ID() string             { return s.id }
func (s *fakeSession) AppName() string        { return "test" }
func (s *fakeSession) UserID() string         { return "u1" }
func (s *fakeSession) State() session.State   { return s.state }
func (s *fakeSession) Events() session.Events { return &emptyEvents{} }

type emptyEvents struct{}

func (e *emptyEvents) All() iter.Seq[*session.Event] {
	return func(func(*session.Event) bool) {}
}
func (e *emptyEvents) Len() int                  { return 0 }
func (e *emptyEvents) At(int) *session.Event     { return nil }
func (s *fakeSession) LastUpdateTime() time.Time { return time.Time{} }
func (s *fakeSession) Title() string             { return s.title }

type fakeState struct {
	data map[string]any
}

func (s *fakeState) Get(key string) (any, error) {
	v, ok := s.data[key]
	if !ok {
		return nil, session.ErrStateKeyNotExist
	}
	return v, nil
}

func (s *fakeState) Set(string, any) error { return nil }

func (s *fakeState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range s.data {
			if !yield(k, v) {
				return
			}
		}
	}
}

func TestEffectiveTitle_FirstClassWins(t *testing.T) {
	sess := &fakeSession{
		id:    "s1",
		title: "First Class",
		state: &fakeState{data: map[string]any{"title": "Legacy"}},
	}
	got := effectiveTitle(sess)
	if got != "First Class" {
		t.Fatalf("expected first-class title, got %q", got)
	}
}

func TestEffectiveTitle_LegacyFallback(t *testing.T) {
	sess := &fakeSession{
		id:    "s1",
		title: "",
		state: &fakeState{data: map[string]any{"title": "Legacy Title"}},
	}
	got := effectiveTitle(sess)
	if got != "Legacy Title" {
		t.Fatalf("expected legacy title, got %q", got)
	}
}

func TestEffectiveTitle_LegacyTrimmed(t *testing.T) {
	sess := &fakeSession{
		id:    "s1",
		title: "",
		state: &fakeState{data: map[string]any{"title": "  Old Title  "}},
	}
	got := effectiveTitle(sess)
	if got != "Old Title" {
		t.Fatalf("expected trimmed legacy title, got %q", got)
	}
}

func TestEffectiveTitle_Empty(t *testing.T) {
	sess := &fakeSession{
		id:    "s1",
		title: "",
		state: &fakeState{data: map[string]any{}},
	}
	got := effectiveTitle(sess)
	if got != "" {
		t.Fatalf("expected empty title, got %q", got)
	}
}

func TestEffectiveTitle_BlankLegacyIgnored(t *testing.T) {
	sess := &fakeSession{
		id:    "s1",
		title: "",
		state: &fakeState{data: map[string]any{"title": "   "}},
	}
	got := effectiveTitle(sess)
	if got != "" {
		t.Fatalf("expected empty for blank legacy title, got %q", got)
	}
}

func TestSessionToInfo_IncludesTitle(t *testing.T) {
	sess := &fakeSession{
		id:    "s1",
		title: "My Chat",
		state: &fakeState{data: map[string]any{}},
	}
	info := sessionToInfo(sess)
	if info.GetTitle() != "My Chat" {
		t.Fatalf("expected title in SessionInfo, got %q", info.GetTitle())
	}
}

// Ensure the generated proto SessionInfo.Title field exists and roundtrips.
func TestSessionInfo_TitleFieldExists(t *testing.T) {
	info := &agentsv1.SessionInfo{
		SessionId: "s1",
		Title:     "Test Title",
	}
	if info.GetTitle() != "Test Title" {
		t.Fatalf("expected Title field on SessionInfo, got %q", info.GetTitle())
	}
}

// Ensure UpdateSessionTitleRequest and Response proto types exist.
func TestUpdateSessionTitle_ProtoTypes(t *testing.T) {
	req := &agentsv1.UpdateSessionTitleRequest{
		AppName:   "web",
		UserId:    "u1",
		SessionId: "s1",
		Title:     "Test",
	}
	if req.GetTitle() != "Test" {
		t.Fatal("request title roundtrip failed")
	}

	resp := &agentsv1.UpdateSessionTitleResponse{
		Session: &agentsv1.SessionInfo{Title: "Test"},
	}
	if resp.GetSession().GetTitle() != "Test" {
		t.Fatal("response title roundtrip failed")
	}
}

// Verify sessionToInfo populates LastUpdateTime (regression guard).
func TestSessionToInfo_LastUpdateTimeSet(t *testing.T) {
	sess := &fakeSession{
		id:    "s1",
		title: "x",
		state: &fakeState{data: map[string]any{}},
	}
	info := sessionToInfo(sess)
	if info.GetLastUpdateTime() == nil {
		t.Fatal("expected LastUpdateTime to be set")
	}
}
