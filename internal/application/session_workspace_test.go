package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeWorkspaceSessionStore stubs the WorkspaceSessionStore interface for tests.
type fakeWorkspaceSessionStore struct {
	sessions    map[string][]session.Session // workspaceID -> sessions
	wsIDByCoord map[string]string            // "app/user/session" -> workspace_id
}

func newFakeWSStore() *fakeWorkspaceSessionStore {
	return &fakeWorkspaceSessionStore{
		sessions:    make(map[string][]session.Session),
		wsIDByCoord: make(map[string]string),
	}
}

func (f *fakeWorkspaceSessionStore) ListByWorkspace(_ context.Context, workspaceID, userID string) ([]session.Session, error) {
	all := f.sessions[workspaceID]
	if userID == "" {
		return all, nil
	}
	var filtered []session.Session
	for _, s := range all {
		if s.UserID() == userID {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

func (f *fakeWorkspaceSessionStore) GetWorkspaceID(_ context.Context, appName, userID, sessionID string) (string, error) {
	key := appName + "/" + userID + "/" + sessionID
	ws, ok := f.wsIDByCoord[key]
	if !ok {
		return "", errors.New("session not found")
	}
	return ws, nil
}

func (f *fakeWorkspaceSessionStore) addSession(wsID string, sess session.Session) {
	f.sessions[wsID] = append(f.sessions[wsID], sess)
	key := sess.AppName() + "/" + sess.UserID() + "/" + sess.ID()
	f.wsIDByCoord[key] = wsID
}

// fakeWSSession implements session.Session with workspace identity.
type fakeWSSession struct {
	id, appName, userID, wsID, title string
	lastUpdate                        time.Time
}

func (s *fakeWSSession) ID() string                { return s.id }
func (s *fakeWSSession) AppName() string           { return s.appName }
func (s *fakeWSSession) UserID() string            { return s.userID }
func (s *fakeWSSession) State() session.State      { return &fakeState{data: nil} }
func (s *fakeWSSession) Events() session.Events    { return &emptyEvents{} }
func (s *fakeWSSession) LastUpdateTime() time.Time { return s.lastUpdate }
func (s *fakeWSSession) Title() string             { return s.title }
func (s *fakeWSSession) WorkspaceID() string       { return s.wsID }

// fakeListSessionService implements session.Service with scripted List/Get responses.
type fakeListSessionService struct {
	session.Service
	sessions []session.Session
}

func (f *fakeListSessionService) List(_ context.Context, _ *session.ListRequest) (*session.ListResponse, error) {
	return &session.ListResponse{Sessions: f.sessions}, nil
}

func (f *fakeListSessionService) Get(_ context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	for _, s := range f.sessions {
		if s.ID() == req.SessionID && s.AppName() == req.AppName && s.UserID() == req.UserID {
			return &session.GetResponse{Session: s}, nil
		}
	}
	return nil, errors.New("session not found")
}

func (f *fakeListSessionService) Delete(_ context.Context, _ *session.DeleteRequest) error {
	return nil
}

func TestListSessions_WorkspaceScoped_ReturnsOnlyWorkspaceSessions(t *testing.T) {
	wsStore := newFakeWSStore()
	now := time.Now()

	wsStore.addSession("ws-alpha", &fakeWSSession{
		id: "s1", appName: "web-chat", userID: "user-1", wsID: "ws-alpha",
		title: "Alpha Chat", lastUpdate: now,
	})
	wsStore.addSession("ws-beta", &fakeWSSession{
		id: "s2", appName: "web-chat", userID: "user-1", wsID: "ws-beta",
		title: "Beta Chat", lastUpdate: now.Add(-time.Hour),
	})
	wsStore.addSession("ws-alpha", &fakeWSSession{
		id: "s3", appName: "web-chat", userID: "user-2", wsID: "ws-alpha",
		title: "Other User", lastUpdate: now.Add(-2 * time.Hour),
	})

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1"}, nil)

	resp, err := svc.ListSessions(ctx, connect.NewRequest(&agentsv1.ListSessionsRequest{
		WorkspaceScoped: true,
	}))
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	if len(resp.Msg.GetSessions()) != 1 {
		t.Fatalf("expected 1 session (own), got %d", len(resp.Msg.GetSessions()))
	}
	if resp.Msg.GetSessions()[0].GetSessionId() != "s1" {
		t.Fatalf("expected session s1, got %q", resp.Msg.GetSessions()[0].GetSessionId())
	}
	if resp.Msg.GetSessions()[0].GetWorkspaceId() != "ws-alpha" {
		t.Fatalf("expected workspace_id ws-alpha, got %q", resp.Msg.GetSessions()[0].GetWorkspaceId())
	}
	if resp.Msg.GetTotal() != 1 {
		t.Fatalf("expected total=1, got %d", resp.Msg.GetTotal())
	}
}

func TestListSessions_WorkspaceScoped_AdminSeesAllUsers(t *testing.T) {
	wsStore := newFakeWSStore()
	now := time.Now()

	wsStore.addSession("ws-alpha", &fakeWSSession{
		id: "s1", appName: "web-chat", userID: "user-1", wsID: "ws-alpha",
		title: "User 1 Chat", lastUpdate: now,
	})
	wsStore.addSession("ws-alpha", &fakeWSSession{
		id: "s2", appName: "web-chat", userID: "user-2", wsID: "ws-alpha",
		title: "User 2 Chat", lastUpdate: now,
	})

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAdmin(ctx)

	resp, err := svc.ListSessions(ctx, connect.NewRequest(&agentsv1.ListSessionsRequest{
		WorkspaceScoped: true,
	}))
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	if len(resp.Msg.GetSessions()) != 2 {
		t.Fatalf("admin should see all workspace sessions, got %d", len(resp.Msg.GetSessions()))
	}
}

func TestListSessions_WorkspaceScoped_RequiresWorkspaceHeader(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(newFakeWSStore())

	ctx := context.Background()
	_, err := svc.ListSessions(ctx, connect.NewRequest(&agentsv1.ListSessionsRequest{
		WorkspaceScoped: true,
	}))
	if err == nil {
		t.Fatal("expected error for missing workspace header")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}
}

func TestListSessions_WorkspaceScoped_ExcludesLegacySessions(t *testing.T) {
	wsStore := newFakeWSStore()
	now := time.Now()

	// Only sessions with workspace_id are added to the store; legacy ones
	// are never stored with a workspace so they won't appear.
	wsStore.addSession("ws-alpha", &fakeWSSession{
		id: "s1", appName: "web-chat", userID: "user-1", wsID: "ws-alpha",
		title: "New Session", lastUpdate: now,
	})

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1"}, nil)

	resp, err := svc.ListSessions(ctx, connect.NewRequest(&agentsv1.ListSessionsRequest{
		WorkspaceScoped: true,
	}))
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(resp.Msg.GetSessions()) != 1 {
		t.Fatalf("expected 1 session, got %d", len(resp.Msg.GetSessions()))
	}
}

func TestListSessions_WorkspaceScoped_PaginationAfterFilter(t *testing.T) {
	wsStore := newFakeWSStore()
	now := time.Now()

	for i := 0; i < 5; i++ {
		wsStore.addSession("ws-alpha", &fakeWSSession{
			id: "s" + string(rune('a'+i)), appName: "web-chat", userID: "user-1", wsID: "ws-alpha",
			lastUpdate: now.Add(-time.Duration(i) * time.Hour),
		})
	}

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1"}, nil)

	resp, err := svc.ListSessions(ctx, connect.NewRequest(&agentsv1.ListSessionsRequest{
		WorkspaceScoped: true,
		PageSize:        2,
	}))
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if len(resp.Msg.GetSessions()) != 2 {
		t.Fatalf("expected 2 sessions on page 1, got %d", len(resp.Msg.GetSessions()))
	}
	if resp.Msg.GetTotal() != 5 {
		t.Fatalf("expected total=5, got %d", resp.Msg.GetTotal())
	}
	if resp.Msg.GetNextPageToken() == "" {
		t.Fatal("expected next_page_token on page 1")
	}

	// Page 2.
	resp2, err := svc.ListSessions(ctx, connect.NewRequest(&agentsv1.ListSessionsRequest{
		WorkspaceScoped: true,
		PageSize:        2,
		PageToken:       resp.Msg.GetNextPageToken(),
	}))
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(resp2.Msg.GetSessions()) != 2 {
		t.Fatalf("expected 2 sessions on page 2, got %d", len(resp2.Msg.GetSessions()))
	}
}

func TestGetSession_WrongWorkspace_ReturnsNotFound(t *testing.T) {
	wsStore := newFakeWSStore()
	now := time.Now()

	sess := &fakeWSSession{
		id: "s1", appName: "web-chat", userID: "user-1", wsID: "ws-beta",
		title: "Beta Session", lastUpdate: now,
	}
	wsStore.addSession("ws-beta", sess)

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{sessions: []session.Session{sess}})
	svc.SetWorkspaceSessionStore(wsStore)

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1"}, nil)

	_, err := svc.GetSession(ctx, connect.NewRequest(&agentsv1.GetSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "s1",
	}))
	if err == nil {
		t.Fatal("expected error for wrong workspace")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestGetSession_SameWorkspace_Succeeds(t *testing.T) {
	now := time.Now()
	sess := &fakeWSSession{
		id: "s1", appName: "web-chat", userID: "user-1", wsID: "ws-alpha",
		title: "Alpha Session", lastUpdate: now,
	}

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{sessions: []session.Session{sess}})

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1"}, nil)

	resp, err := svc.GetSession(ctx, connect.NewRequest(&agentsv1.GetSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if resp.Msg.GetSessionDetail().GetSession().GetWorkspaceId() != "ws-alpha" {
		t.Fatalf("expected workspace_id ws-alpha, got %q", resp.Msg.GetSessionDetail().GetSession().GetWorkspaceId())
	}
}

func TestGetSession_AdminBypassesWorkspaceCheck(t *testing.T) {
	now := time.Now()
	sess := &fakeWSSession{
		id: "s1", appName: "web-chat", userID: "user-1", wsID: "ws-beta",
		title: "Beta Session", lastUpdate: now,
	}

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{sessions: []session.Session{sess}})

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAdmin(ctx)

	resp, err := svc.GetSession(ctx, connect.NewRequest(&agentsv1.GetSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("admin GetSession: %v", err)
	}
	if resp.Msg.GetSessionDetail().GetSession().GetSessionId() != "s1" {
		t.Fatal("admin should bypass workspace check")
	}
}

func TestDeleteSession_WrongWorkspace_ReturnsNotFound(t *testing.T) {
	wsStore := newFakeWSStore()
	wsStore.wsIDByCoord["web-chat/user-1/s1"] = "ws-beta"

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1"}, nil)

	_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "s1",
	}))
	if err == nil {
		t.Fatal("expected error for wrong workspace delete")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestDeleteSession_SameWorkspace_Succeeds(t *testing.T) {
	wsStore := newFakeWSStore()
	wsStore.wsIDByCoord["web-chat/user-1/s1"] = "ws-alpha"

	stub := &stubSessionService{}
	svc := NewSessionServiceServer()
	svc.SetSessionService(stub)
	svc.SetWorkspaceSessionStore(wsStore)

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1"}, nil)

	_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if len(stub.deleted) != 1 {
		t.Fatalf("expected 1 delete call, got %d", len(stub.deleted))
	}
}

func TestGetSession_LegacySession_BlockedWhenWorkspaceScoped(t *testing.T) {
	now := time.Now()
	// Legacy session has empty workspace_id.
	sess := &fakeWSSession{
		id: "legacy-1", appName: "web-chat", userID: "user-1", wsID: "",
		title: "Old Session", lastUpdate: now,
	}

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{sessions: []session.Session{sess}})

	ctx := workspace.WithID(context.Background(), "ws-alpha")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1"}, nil)

	_, err := svc.GetSession(ctx, connect.NewRequest(&agentsv1.GetSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "legacy-1",
	}))
	if err == nil {
		t.Fatal("expected error for legacy session in workspace-scoped request")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestSessionToInfo_ExposesWorkspaceID(t *testing.T) {
	sess := &fakeWSSession{
		id: "s1", appName: "web-chat", userID: "user-1", wsID: "ws-alpha",
		title: "Chat", lastUpdate: time.Now(),
	}
	info := sessionToInfo(sess)
	if info.GetWorkspaceId() != "ws-alpha" {
		t.Fatalf("expected workspace_id ws-alpha, got %q", info.GetWorkspaceId())
	}
}

func TestSessionToInfo_EmptyWorkspaceForLegacy(t *testing.T) {
	sess := &fakeWSSession{
		id: "s1", appName: "web-chat", userID: "user-1", wsID: "",
		title: "Legacy", lastUpdate: time.Now(),
	}
	info := sessionToInfo(sess)
	if info.GetWorkspaceId() != "" {
		t.Fatalf("expected empty workspace_id for legacy session, got %q", info.GetWorkspaceId())
	}
}
