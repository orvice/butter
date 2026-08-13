package application

import (
	"context"
	"errors"
	"iter"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	invmem "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// --- fakes for session summary tests ---

// fakeReadStore implements SessionReadStore for tests.
type fakeReadStore struct {
	reads map[string]time.Time // "app/user/session" -> last_read_at
}

func newFakeReadStore() *fakeReadStore {
	return &fakeReadStore{reads: make(map[string]time.Time)}
}

func (f *fakeReadStore) MarkRead(_ context.Context, appName, userID, sessionID string, readAt time.Time) (SessionReadResult, error) {
	key := appName + "/" + userID + "/" + sessionID
	f.reads[key] = readAt
	return SessionReadResult{
		SessionID:      sessionID,
		AppName:        appName,
		UserID:         userID,
		LastUpdateTime: readAt, // simplified; in real impl comes from DB
		LastReadAt:     readAt,
	}, nil
}

// fakeReadableSession implements session.Session with LastReadAt support.
type fakeReadableSession struct {
	id, appName, userID, wsID, title string
	lastUpdate                       time.Time
	lastReadAt                       *time.Time
	events                           []*session.Event
}

func (s *fakeReadableSession) ID() string                { return s.id }
func (s *fakeReadableSession) AppName() string           { return s.appName }
func (s *fakeReadableSession) UserID() string            { return s.userID }
func (s *fakeReadableSession) State() session.State      { return &fakeState{data: nil} }
func (s *fakeReadableSession) LastUpdateTime() time.Time { return s.lastUpdate }
func (s *fakeReadableSession) Title() string             { return s.title }
func (s *fakeReadableSession) WorkspaceID() string       { return s.wsID }
func (s *fakeReadableSession) LastReadAt() *time.Time    { return s.lastReadAt }

func (s *fakeReadableSession) Events() session.Events {
	if len(s.events) > 0 {
		return &fakeEventsImpl{events: s.events}
	}
	return &emptyEvents{}
}

// fakeListSessionServiceWithEvents supports Get with events for interrupt enrichment.
type fakeListSessionServiceWithEvents struct {
	session.Service
	sessions []session.Session
}

func (f *fakeListSessionServiceWithEvents) List(_ context.Context, _ *session.ListRequest) (*session.ListResponse, error) {
	return &session.ListResponse{Sessions: f.sessions}, nil
}

func (f *fakeListSessionServiceWithEvents) Get(_ context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	for _, s := range f.sessions {
		if s.ID() == req.SessionID && s.AppName() == req.AppName && s.UserID() == req.UserID {
			return &session.GetResponse{Session: s}, nil
		}
	}
	return nil, errors.New("session not found")
}

func (f *fakeListSessionServiceWithEvents) Delete(_ context.Context, _ *session.DeleteRequest) error {
	return nil
}

// --- Summary derivation tests ---

func TestSessionToInfo_ExposesReadState_Unread(t *testing.T) {
	now := time.Now()
	sess := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: now, lastReadAt: nil,
	}
	info := sessionToInfo(sess)
	if !info.GetUnread() {
		t.Fatal("expected unread=true when lastReadAt is nil")
	}
	if info.GetLastReadAt() != nil {
		t.Fatal("expected last_read_at to be nil")
	}
}

func TestSessionToInfo_ExposesReadState_Read(t *testing.T) {
	now := time.Now()
	readAt := now.Add(time.Second)
	sess := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: now, lastReadAt: &readAt,
	}
	info := sessionToInfo(sess)
	if info.GetUnread() {
		t.Fatal("expected unread=false when lastReadAt > lastUpdateTime")
	}
	if info.GetLastReadAt() == nil {
		t.Fatal("expected last_read_at to be set")
	}
}

func TestSessionToInfo_Unread_WhenUpdateAfterRead(t *testing.T) {
	readAt := time.Now()
	updateAt := readAt.Add(time.Minute)
	sess := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: updateAt, lastReadAt: &readAt,
	}
	info := sessionToInfo(sess)
	if !info.GetUnread() {
		t.Fatal("expected unread=true when lastUpdateTime > lastReadAt")
	}
}

func TestEnrichSessionInfos_SetsInvocationStatus_Running(t *testing.T) {
	invRepo := invmem.New()
	ctx := context.Background()

	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		SessionId:   "s1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		WorkspaceId: "ws1",
		StartedAt:   timestamppb.Now(),
	}
	_ = invRepo.Save(ctx, inv)

	svc := NewSessionServiceServer()
	svc.SetInvocationRepo(invRepo)

	infos := []*agentsv1.SessionInfo{
		{SessionId: "s1", AppName: "web-chat", UserId: "u1"},
	}
	svc.enrichSessionInfos(ctx, "ws1", infos)

	if infos[0].GetLatestInvocationId() != "inv-1" {
		t.Fatalf("expected invocation_id inv-1, got %q", infos[0].GetLatestInvocationId())
	}
	if infos[0].GetLatestInvocationStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
		t.Fatalf("expected RUNNING, got %v", infos[0].GetLatestInvocationStatus())
	}
}

func TestEnrichSessionInfos_SetsInvocationStatus_Failed(t *testing.T) {
	invRepo := invmem.New()
	ctx := context.Background()

	inv := &agentsv1.Invocation{
		Id:          "inv-2",
		SessionId:   "s1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED,
		WorkspaceId: "ws1",
		StartedAt:   timestamppb.Now(),
	}
	_ = invRepo.Save(ctx, inv)

	svc := NewSessionServiceServer()
	svc.SetInvocationRepo(invRepo)

	infos := []*agentsv1.SessionInfo{
		{SessionId: "s1", AppName: "web-chat", UserId: "u1"},
	}
	svc.enrichSessionInfos(ctx, "ws1", infos)

	if infos[0].GetLatestInvocationId() != "inv-2" {
		t.Fatalf("expected invocation_id inv-2, got %q", infos[0].GetLatestInvocationId())
	}
	if infos[0].GetLatestInvocationStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("expected FAILED, got %v", infos[0].GetLatestInvocationStatus())
	}
}

func TestEnrichSessionInfos_NoInvocation_LeavesZero(t *testing.T) {
	invRepo := invmem.New()
	svc := NewSessionServiceServer()
	svc.SetInvocationRepo(invRepo)

	infos := []*agentsv1.SessionInfo{
		{SessionId: "s-no-inv", AppName: "web-chat", UserId: "u1"},
	}
	svc.enrichSessionInfos(context.Background(), "ws1", infos)

	if infos[0].GetLatestInvocationId() != "" {
		t.Fatalf("expected empty invocation_id, got %q", infos[0].GetLatestInvocationId())
	}
}

func TestEnrichSessionInfos_PendingInterrupt(t *testing.T) {
	sess := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: time.Now(),
		events: []*session.Event{
			{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{
							FunctionCall: &genai.FunctionCall{
								ID:   "int-1",
								Name: workflow.WorkflowInputFunctionCallName,
								Args: map[string]any{"message": "What should I do?"},
							},
						}},
					},
				},
				Author: "agent",
			},
		},
	}

	sessSvc := &fakeListSessionServiceWithEvents{sessions: []session.Session{sess}}

	svc := NewSessionServiceServer()
	svc.SetSessionService(sessSvc)

	infos := []*agentsv1.SessionInfo{
		{SessionId: "s1", AppName: "web-chat", UserId: "u1"},
	}
	svc.enrichSessionInfos(context.Background(), "ws1", infos)

	if !infos[0].GetHasPendingInterrupt() {
		t.Fatal("expected has_pending_interrupt=true")
	}
}

func TestEnrichSessionInfos_AnsweredInterrupt_NotPending(t *testing.T) {
	sess := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: time.Now(),
		events: []*session.Event{
			{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{
							FunctionCall: &genai.FunctionCall{
								ID:   "int-1",
								Name: workflow.WorkflowInputFunctionCallName,
								Args: map[string]any{"message": "Approve?"},
							},
						}},
					},
				},
				Author: "agent",
			},
			{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{
							FunctionResponse: &genai.FunctionResponse{
								ID:   "int-1",
								Name: workflow.WorkflowInputFunctionCallName,
								Response: map[string]any{
									"payload": "yes",
								},
							},
						}},
					},
				},
				Author: "user",
			},
		},
	}

	sessSvc := &fakeListSessionServiceWithEvents{sessions: []session.Session{sess}}

	svc := NewSessionServiceServer()
	svc.SetSessionService(sessSvc)

	infos := []*agentsv1.SessionInfo{
		{SessionId: "s1", AppName: "web-chat", UserId: "u1"},
	}
	svc.enrichSessionInfos(context.Background(), "ws1", infos)

	if infos[0].GetHasPendingInterrupt() {
		t.Fatal("expected has_pending_interrupt=false when interrupt is answered")
	}
}

// --- MarkSessionRead tests ---

func TestMarkSessionRead_Success(t *testing.T) {
	wsStore := newFakeWSStore()
	wsStore.wsIDByCoord["web-chat/u1/s1"] = "ws1"

	readStore := newFakeReadStore()

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)
	svc.SetReadStore(readStore)

	ctx := workspace.WithID(context.Background(), "ws1")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "u1"}, nil)

	resp, err := svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		AppName:         "web-chat",
		UserId:          "u1",
		SessionId:       "s1",
		LastReadEventId: "evt-10",
	}))
	if err != nil {
		t.Fatalf("MarkSessionRead: %v", err)
	}
	if resp.Msg.GetSession().GetSessionId() != "s1" {
		t.Fatalf("expected session_id s1, got %q", resp.Msg.GetSession().GetSessionId())
	}
	if resp.Msg.GetSession().GetLastReadAt() == nil {
		t.Fatal("expected last_read_at to be set")
	}
}

func TestMarkSessionRead_AnotherUserDenied(t *testing.T) {
	wsStore := newFakeWSStore()
	wsStore.wsIDByCoord["web-chat/u2/s1"] = "ws1"

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)
	svc.SetReadStore(newFakeReadStore())

	ctx := workspace.WithID(context.Background(), "ws1")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "u1"}, nil)

	_, err := svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		AppName:   "web-chat",
		UserId:    "u2",
		SessionId: "s1",
	}))
	if err == nil {
		t.Fatal("expected error when another user marks read")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestMarkSessionRead_WrongWorkspaceDenied(t *testing.T) {
	wsStore := newFakeWSStore()
	wsStore.wsIDByCoord["web-chat/u1/s1"] = "ws-other"

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)
	svc.SetReadStore(newFakeReadStore())

	ctx := workspace.WithID(context.Background(), "ws1")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "u1"}, nil)

	_, err := svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		AppName:   "web-chat",
		UserId:    "u1",
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

func TestMarkSessionRead_AdminBypassesOwnerCheck(t *testing.T) {
	wsStore := newFakeWSStore()
	wsStore.wsIDByCoord["web-chat/u2/s1"] = "ws1"

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)
	svc.SetReadStore(newFakeReadStore())

	ctx := workspace.WithID(context.Background(), "ws1")
	ctx = auth.WithAdmin(ctx)

	resp, err := svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		AppName:   "web-chat",
		UserId:    "u2",
		SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("admin MarkSessionRead: %v", err)
	}
	if resp.Msg.GetSession().GetSessionId() != "s1" {
		t.Fatal("admin should be able to mark any session read")
	}
}

func TestMarkSessionRead_RequiresAppName(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetReadStore(newFakeReadStore())

	ctx := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "u1"}, nil)
	_, err := svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		UserId:    "u1",
		SessionId: "s1",
	}))
	if err == nil {
		t.Fatal("expected error for missing app_name")
	}
}

func TestMarkSessionRead_RequiresUserId(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetReadStore(newFakeReadStore())

	ctx := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "u1"}, nil)
	_, err := svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		AppName:   "web-chat",
		SessionId: "s1",
	}))
	if err == nil {
		t.Fatal("expected error for missing user_id")
	}
}

func TestMarkSessionRead_RequiresSessionId(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetReadStore(newFakeReadStore())

	ctx := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "u1"}, nil)
	_, err := svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		AppName: "web-chat",
		UserId:  "u1",
	}))
	if err == nil {
		t.Fatal("expected error for missing session_id")
	}
}

func TestMarkSessionRead_MultiClient_SecondOverwritesFirst(t *testing.T) {
	wsStore := newFakeWSStore()
	wsStore.wsIDByCoord["web-chat/u1/s1"] = "ws1"

	readStore := newFakeReadStore()

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)
	svc.SetReadStore(readStore)

	ctx := workspace.WithID(context.Background(), "ws1")
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "u1"}, nil)

	// First mark.
	_, err := svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		AppName:         "web-chat",
		UserId:          "u1",
		SessionId:       "s1",
		LastReadEventId: "evt-5",
	}))
	if err != nil {
		t.Fatalf("first MarkSessionRead: %v", err)
	}

	first := readStore.reads["web-chat/u1/s1"]

	// Slight delay then second mark.
	time.Sleep(time.Millisecond)

	_, err = svc.MarkSessionRead(ctx, connect.NewRequest(&agentsv1.MarkSessionReadRequest{
		AppName:         "web-chat",
		UserId:          "u1",
		SessionId:       "s1",
		LastReadEventId: "evt-10",
	}))
	if err != nil {
		t.Fatalf("second MarkSessionRead: %v", err)
	}

	second := readStore.reads["web-chat/u1/s1"]
	if !second.After(first) {
		t.Fatalf("expected second read (%v) after first read (%v)", second, first)
	}
}

// --- ListSessions workspace-scoped enrichment tests ---

func TestListSessions_WorkspaceScoped_EnrichesWithInvocationAndInterrupt(t *testing.T) {
	invRepo := invmem.New()
	ctx := context.Background()

	_ = invRepo.Save(ctx, &agentsv1.Invocation{
		Id:          "inv-run",
		SessionId:   "s1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		WorkspaceId: "ws1",
		StartedAt:   timestamppb.Now(),
	})
	_ = invRepo.Save(ctx, &agentsv1.Invocation{
		Id:          "inv-done",
		SessionId:   "s2",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		WorkspaceId: "ws1",
		StartedAt:   timestamppb.Now(),
	})

	now := time.Now()
	readAt := now.Add(-time.Hour)

	s1 := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: now, lastReadAt: nil,
	}
	s2 := &fakeReadableSession{
		id: "s2", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: now, lastReadAt: &readAt,
	}

	wsStore := newFakeWSStore()
	wsStore.addSession("ws1", s1)
	wsStore.addSession("ws1", s2)

	sessSvc := &fakeListSessionServiceWithEvents{
		sessions: []session.Session{s1, s2},
	}

	svc := NewSessionServiceServer()
	svc.SetSessionService(sessSvc)
	svc.SetWorkspaceSessionStore(wsStore)
	svc.SetInvocationRepo(invRepo)

	rctx := workspace.WithID(context.Background(), "ws1")
	rctx = auth.WithAuthenticated(rctx, &agentsv1.User{Id: "u1"}, nil)

	resp, err := svc.ListSessions(rctx, connect.NewRequest(&agentsv1.ListSessionsRequest{
		WorkspaceScoped: true,
	}))
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	if len(resp.Msg.GetSessions()) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(resp.Msg.GetSessions()))
	}

	// Find sessions by ID for deterministic checks.
	byID := make(map[string]*agentsv1.SessionInfo)
	for _, s := range resp.Msg.GetSessions() {
		byID[s.GetSessionId()] = s
	}

	// s1: RUNNING invocation, unread.
	s1Info := byID["s1"]
	if s1Info.GetLatestInvocationId() != "inv-run" {
		t.Fatalf("s1: expected inv-run, got %q", s1Info.GetLatestInvocationId())
	}
	if s1Info.GetLatestInvocationStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
		t.Fatalf("s1: expected RUNNING, got %v", s1Info.GetLatestInvocationStatus())
	}
	if !s1Info.GetUnread() {
		t.Fatal("s1: expected unread=true")
	}

	// s2: SUCCEEDED invocation, unread (update after read).
	s2Info := byID["s2"]
	if s2Info.GetLatestInvocationId() != "inv-done" {
		t.Fatalf("s2: expected inv-done, got %q", s2Info.GetLatestInvocationId())
	}
	if s2Info.GetLatestInvocationStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("s2: expected SUCCEEDED, got %v", s2Info.GetLatestInvocationStatus())
	}
	if !s2Info.GetUnread() {
		t.Fatal("s2: expected unread=true (update after read)")
	}
}

// --- GetSession enrichment tests ---

func TestGetSession_EnrichesWithInvocationStatus(t *testing.T) {
	invRepo := invmem.New()
	ctx := context.Background()

	_ = invRepo.Save(ctx, &agentsv1.Invocation{
		Id:          "inv-fail",
		SessionId:   "s1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED,
		WorkspaceId: "ws1",
		StartedAt:   timestamppb.Now(),
	})

	now := time.Now()
	sess := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: now,
	}

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionServiceWithEvents{sessions: []session.Session{sess}})
	svc.SetInvocationRepo(invRepo)

	rctx := workspace.WithID(context.Background(), "ws1")
	rctx = auth.WithAuthenticated(rctx, &agentsv1.User{Id: "u1"}, nil)

	resp, err := svc.GetSession(rctx, connect.NewRequest(&agentsv1.GetSessionRequest{
		AppName:   "web-chat",
		UserId:    "u1",
		SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	info := resp.Msg.GetSessionDetail().GetSession()
	if info.GetLatestInvocationId() != "inv-fail" {
		t.Fatalf("expected inv-fail, got %q", info.GetLatestInvocationId())
	}
	if info.GetLatestInvocationStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("expected FAILED, got %v", info.GetLatestInvocationStatus())
	}
}

func TestGetSession_PendingInterrupt(t *testing.T) {
	now := time.Now()
	sess := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: now,
		events: []*session.Event{
			{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{
							FunctionCall: &genai.FunctionCall{
								ID:   "int-1",
								Name: workflow.WorkflowInputFunctionCallName,
								Args: map[string]any{"message": "Approve deploy?"},
							},
						}},
					},
				},
				Author: "agent",
			},
		},
	}

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionServiceWithEvents{sessions: []session.Session{sess}})

	rctx := workspace.WithID(context.Background(), "ws1")
	rctx = auth.WithAuthenticated(rctx, &agentsv1.User{Id: "u1"}, nil)

	resp, err := svc.GetSession(rctx, connect.NewRequest(&agentsv1.GetSessionRequest{
		AppName:   "web-chat",
		UserId:    "u1",
		SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	info := resp.Msg.GetSessionDetail().GetSession()
	if !info.GetHasPendingInterrupt() {
		t.Fatal("expected has_pending_interrupt=true for unanswered interrupt")
	}
}

// --- Running/Failure/Unread transitions ---

func TestInvocationTransitions_QueuedToRunningToFailed(t *testing.T) {
	invRepo := invmem.New()
	ctx := context.Background()

	// QUEUED.
	inv := &agentsv1.Invocation{
		Id:          "inv-trans",
		SessionId:   "s1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		WorkspaceId: "ws1",
		StartedAt:   timestamppb.Now(),
	}
	_ = invRepo.Save(ctx, inv)

	svc := NewSessionServiceServer()
	svc.SetInvocationRepo(invRepo)

	infos := []*agentsv1.SessionInfo{{SessionId: "s1"}}
	svc.enrichSessionInfos(ctx, "ws1", infos)
	if infos[0].GetLatestInvocationStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED {
		t.Fatalf("expected QUEUED, got %v", infos[0].GetLatestInvocationStatus())
	}

	// Transition to RUNNING.
	inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
	_ = invRepo.Save(ctx, inv)

	infos = []*agentsv1.SessionInfo{{SessionId: "s1"}}
	svc.enrichSessionInfos(ctx, "ws1", infos)
	if infos[0].GetLatestInvocationStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
		t.Fatalf("expected RUNNING, got %v", infos[0].GetLatestInvocationStatus())
	}

	// Transition to FAILED.
	inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
	_ = invRepo.Save(ctx, inv)

	infos = []*agentsv1.SessionInfo{{SessionId: "s1"}}
	svc.enrichSessionInfos(ctx, "ws1", infos)
	if infos[0].GetLatestInvocationStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("expected FAILED, got %v", infos[0].GetLatestInvocationStatus())
	}
}

// --- FIFO Interrupt resume tests ---

func TestEnrichSessionInfos_FIFOInterruptResume(t *testing.T) {
	sess := &fakeReadableSession{
		id: "s1", appName: "web-chat", userID: "u1", wsID: "ws1",
		lastUpdate: time.Now(),
		events: []*session.Event{
			{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{
							FunctionCall: &genai.FunctionCall{
								ID:   "int-1",
								Name: workflow.WorkflowInputFunctionCallName,
								Args: map[string]any{"message": "First question"},
							},
						}},
					},
				},
				Author: "agent",
			},
			{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{
							FunctionCall: &genai.FunctionCall{
								ID:   "int-2",
								Name: workflow.WorkflowInputFunctionCallName,
								Args: map[string]any{"message": "Second question"},
							},
						}},
					},
				},
				Author: "agent",
			},
			// Answer first question.
			{
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Parts: []*genai.Part{{
							FunctionResponse: &genai.FunctionResponse{
								ID:       "int-1",
								Name:     workflow.WorkflowInputFunctionCallName,
								Response: map[string]any{"payload": "yes"},
							},
						}},
					},
				},
				Author: "user",
			},
		},
	}

	sessSvc := &fakeListSessionServiceWithEvents{sessions: []session.Session{sess}}
	svc := NewSessionServiceServer()
	svc.SetSessionService(sessSvc)

	infos := []*agentsv1.SessionInfo{
		{SessionId: "s1", AppName: "web-chat", UserId: "u1"},
	}
	svc.enrichSessionInfos(context.Background(), "ws1", infos)

	if !infos[0].GetHasPendingInterrupt() {
		t.Fatal("expected has_pending_interrupt=true: second interrupt still unanswered")
	}
}

// fakeEventsImpl is declared in session_generate_title_test.go — we reuse it
// via the fakeReadableSession.Events() method above which references it.
// If the build fails due to duplicate declaration, this type is already
// available from the existing test file in the same package.

// Ensure iter import is used (fakeEventsImpl uses it).
var _ iter.Seq[*session.Event] = nil
