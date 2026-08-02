package application

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// titleSeamSessionService overrides Get/List with canned sessions so title
// reads can be asserted through the RPC seam.
type titleSeamSessionService struct {
	session.Service
	sessions []session.Session
}

func (s *titleSeamSessionService) List(context.Context, *session.ListRequest) (*session.ListResponse, error) {
	return &session.ListResponse{Sessions: s.sessions}, nil
}

func (s *titleSeamSessionService) Get(_ context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	for _, sess := range s.sessions {
		if sess.ID() == req.SessionID {
			return &session.GetResponse{Session: sess}, nil
		}
	}
	return nil, fmt.Errorf("session not found: %s", req.SessionID)
}

// newSessionSeamClient serves the real SessionService Connect handler over
// httptest, so requests cross routing, codec negotiation, and error mapping.
// authorize decorates each request context the way the production auth
// middleware does.
func newSessionSeamClient(t *testing.T, svc *SessionServiceServer, authorize func(context.Context) context.Context) agentsv1connect.SessionServiceClient {
	t.Helper()
	path, handler := agentsv1connect.NewSessionServiceHandler(svc, connectx.HandlerOptions()...)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req.WithContext(authorize(req.Context())))
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return agentsv1connect.NewSessionServiceClient(srv.Client(), srv.URL)
}

func asAdmin(ctx context.Context) context.Context { return auth.WithAdmin(ctx) }

func asUser(id string) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return auth.WithAuthenticated(ctx, &agentsv1.User{Id: id, Role: "user"}, &auth.Session{})
	}
}

func TestUpdateSessionTitleSeam_SuccessNormalizes(t *testing.T) {
	store := &stubTitleStore{}
	svc := newTitleTestService(store)
	client := newSessionSeamClient(t, svc, asUser("u1"))

	resp, err := client.UpdateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "s1", Title: "  hello\nworld  ",
	}))
	if err != nil {
		t.Fatalf("UpdateSessionTitle over seam: %v", err)
	}
	if got := resp.Msg.GetSession().GetTitle(); got != "hello world" {
		t.Fatalf("expected normalized title over the wire, got %q", got)
	}
	if store.lastTitle != "hello world" {
		t.Fatalf("store received %q", store.lastTitle)
	}
}

func TestUpdateSessionTitleSeam_NonOwnerPermissionDenied(t *testing.T) {
	svc := newTitleTestService(&stubTitleStore{})
	client := newSessionSeamClient(t, svc, asUser("u1"))

	_, err := client.UpdateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "someone-else", SessionId: "s1", Title: "hijack",
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("expected PermissionDenied over seam, got %v", err)
	}
}

func TestUpdateSessionTitleSeam_UnauthenticatedRejected(t *testing.T) {
	svc := newTitleTestService(&stubTitleStore{})
	client := newSessionSeamClient(t, svc, func(ctx context.Context) context.Context { return ctx })

	_, err := client.UpdateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "s1", Title: "x",
	}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("expected Unauthenticated over seam, got %v", err)
	}
}

func TestUpdateSessionTitleSeam_NotFound(t *testing.T) {
	svc := newTitleTestService(&stubTitleStore{notFound: true})
	client := newSessionSeamClient(t, svc, asAdmin)

	_, err := client.UpdateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "missing", Title: "x",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("expected NotFound over seam, got %v", err)
	}
}

func TestUpdateSessionTitleSeam_BlankInvalidArgument(t *testing.T) {
	svc := newTitleTestService(&stubTitleStore{})
	client := newSessionSeamClient(t, svc, asAdmin)

	_, err := client.UpdateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.UpdateSessionTitleRequest{
		AppName: "web", UserId: "u1", SessionId: "s1", Title: "   \n  ",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument over seam, got %v", err)
	}
}

func TestListSessionsSeam_EffectiveTitles(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetSessionService(&titleSeamSessionService{sessions: []session.Session{
		&fakeSession{id: "s1", title: "First Class", state: &fakeState{data: map[string]any{"title": "Legacy"}}},
		&fakeSession{id: "s2", title: "", state: &fakeState{data: map[string]any{"title": " Legacy Title "}}},
		&fakeSession{id: "s3", title: "", state: &fakeState{data: map[string]any{}}},
	}})
	client := newSessionSeamClient(t, svc, asUser("u1"))

	resp, err := client.ListSessions(context.Background(), connect.NewRequest(&agentsv1.ListSessionsRequest{
		AppName: "web", UserId: "u1",
	}))
	if err != nil {
		t.Fatalf("ListSessions over seam: %v", err)
	}
	titles := make(map[string]string)
	for _, info := range resp.Msg.GetSessions() {
		titles[info.GetSessionId()] = info.GetTitle()
	}
	want := map[string]string{"s1": "First Class", "s2": "Legacy Title", "s3": ""}
	for id, expected := range want {
		if titles[id] != expected {
			t.Fatalf("session %s: expected title %q, got %q", id, expected, titles[id])
		}
	}
}

func TestGenerateSessionTitleSeam_Success(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Seam test question"))},
	}
	store := &stubTitleStore{}
	svc := NewSessionServiceServer()
	svc.SetSessionService(&titleSeamSessionService{sessions: []session.Session{sess}})
	svc.SetTitleStore(store)
	client := newSessionSeamClient(t, svc, asUser("u1"))

	resp, err := client.GenerateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.GenerateSessionTitleRequest{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("GenerateSessionTitle over seam: %v", err)
	}
	if !resp.Msg.GetGenerated() {
		t.Fatal("expected generated=true")
	}
	if got := resp.Msg.GetSession().GetTitle(); got != "Seam test question" {
		t.Fatalf("expected derived title over the wire, got %q", got)
	}
}

func TestGenerateSessionTitleSeam_NonOwnerDenied(t *testing.T) {
	store := &stubTitleStore{}
	svc := NewSessionServiceServer()
	svc.SetTitleStore(store)
	svc.SetSessionService(&titleSeamSessionService{})
	client := newSessionSeamClient(t, svc, asUser("u1"))

	_, err := client.GenerateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.GenerateSessionTitleRequest{
		AppName: "web", UserId: "someone-else", SessionId: "s1",
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("expected PermissionDenied over seam, got %v", err)
	}
}

func TestGenerateSessionTitleSeam_ExistingTitleNoOp(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "Already Set", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Hello"))},
	}
	store := &stubTitleStore{}
	svc := NewSessionServiceServer()
	svc.SetSessionService(&titleSeamSessionService{sessions: []session.Session{sess}})
	svc.SetTitleStore(store)
	client := newSessionSeamClient(t, svc, asAdmin)

	resp, err := client.GenerateSessionTitle(context.Background(), connect.NewRequest(&agentsv1.GenerateSessionTitleRequest{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("expected generated=false when title exists")
	}
	if got := resp.Msg.GetSession().GetTitle(); got != "Already Set" {
		t.Fatalf("expected existing title, got %q", got)
	}
}

func TestGetSessionSeam_LegacyTitleFallback(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetSessionService(&titleSeamSessionService{sessions: []session.Session{
		&fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{"title": "Legacy Title"}}},
	}})
	client := newSessionSeamClient(t, svc, asUser("u1"))

	resp, err := client.GetSession(context.Background(), connect.NewRequest(&agentsv1.GetSessionRequest{
		AppName: "web", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("GetSession over seam: %v", err)
	}
	if got := resp.Msg.GetSessionDetail().GetSession().GetTitle(); got != "Legacy Title" {
		t.Fatalf("expected legacy fallback title, got %q", got)
	}
}
