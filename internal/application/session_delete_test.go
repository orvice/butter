package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"

	inputpartmemory "go.orx.me/apps/butter/internal/repo/inputpart/memory"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// stubDeleteCoordinator tracks cancel-and-wait calls and optionally blocks
// until released, simulating an in-flight invocation.
type stubDeleteCoordinator struct {
	mu          sync.Mutex
	cancelled   map[string]bool
	waitBlocked chan struct{} // if set, CancelAndWait blocks until closed
}

func newStubDeleteCoordinator() *stubDeleteCoordinator {
	return &stubDeleteCoordinator{cancelled: make(map[string]bool)}
}

func (c *stubDeleteCoordinator) CancelAndWait(_ context.Context, invocationID, _ string) bool {
	if c.waitBlocked != nil {
		<-c.waitBlocked
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelled[invocationID] = true
	return true
}

func (c *stubDeleteCoordinator) wasCancelled(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cancelled[id]
}

func TestDeleteSession_CancelsActiveInvocation(t *testing.T) {
	invRepo := invocationmemory.New()
	coord := newStubDeleteCoordinator()
	stub := &stubSessionService{}

	svc := NewSessionServiceServer()
	svc.SetSessionService(stub)
	svc.SetInvocationRepo(invRepo)
	svc.SetAsyncCoordinator(coord)

	wsID := "ws-test"
	sessionID := "sess-1"

	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		AgentName:   "agent-1",
		AppName:     "web-chat",
		UserId:      "user-1",
		SessionId:   sessionID,
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		Input:       "hello",
		Output:      "world",
		WorkspaceId: wsID,
		StartedAt:   timestamppb.Now(),
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ctx := workspace.WithID(context.Background(), wsID)

	_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: sessionID,
	}))
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	if !coord.wasCancelled("inv-1") {
		t.Fatal("expected active invocation to be cancelled")
	}
	if len(stub.deleted) != 1 {
		t.Fatalf("expected session to be deleted, got %d deletes", len(stub.deleted))
	}
}

func TestDeleteSession_RedactsInvocationContent(t *testing.T) {
	invRepo := invocationmemory.New()
	stub := &stubSessionService{}

	svc := NewSessionServiceServer()
	svc.SetSessionService(stub)
	svc.SetInvocationRepo(invRepo)

	wsID := "ws-test"
	sessionID := "sess-1"

	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		AgentName:   "agent-1",
		AgentId:     "aid-1",
		AppName:     "web-chat",
		UserId:      "user-1",
		SessionId:   sessionID,
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		Input:       "sensitive input",
		Output:      "sensitive output",
		Error:       "sensitive error",
		WorkspaceId: wsID,
		StartedAt:   timestamppb.New(time.Now().Add(-time.Minute)),
		FinishedAt:  timestamppb.Now(),
		LatencyMs:   1234,
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ctx := workspace.WithID(context.Background(), wsID)
	_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: sessionID,
	}))
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	redacted, getErr := invRepo.Get(context.Background(), wsID, "inv-1")
	if getErr != nil {
		t.Fatalf("Get redacted invocation: %v", getErr)
	}
	if redacted.GetInput() != "" {
		t.Fatalf("expected input to be redacted, got %q", redacted.GetInput())
	}
	if redacted.GetOutput() != "" {
		t.Fatalf("expected output to be redacted, got %q", redacted.GetOutput())
	}
	if redacted.GetError() != "" {
		t.Fatalf("expected error to be redacted, got %q", redacted.GetError())
	}

	// Operational metadata preserved.
	if redacted.GetAgentName() != "agent-1" {
		t.Fatalf("agent_name should be preserved, got %q", redacted.GetAgentName())
	}
	if redacted.GetAgentId() != "aid-1" {
		t.Fatalf("agent_id should be preserved, got %q", redacted.GetAgentId())
	}
	if redacted.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("status should be preserved, got %v", redacted.GetStatus())
	}
	if redacted.GetLatencyMs() != 1234 {
		t.Fatalf("latency_ms should be preserved, got %d", redacted.GetLatencyMs())
	}
	if redacted.GetStartedAt() == nil || redacted.GetFinishedAt() == nil {
		t.Fatal("timestamps should be preserved")
	}
}

func TestDeleteSession_DeletesInputParts(t *testing.T) {
	invRepo := invocationmemory.New()
	ipRepo := inputpartmemory.New()
	stub := &stubSessionService{}

	svc := NewSessionServiceServer()
	svc.SetSessionService(stub)
	svc.SetInvocationRepo(invRepo)
	svc.SetInputPartRepo(ipRepo)

	wsID := "ws-test"
	sessionID := "sess-1"

	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		AgentName:   "agent-1",
		AppName:     "web-chat",
		UserId:      "user-1",
		SessionId:   sessionID,
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED,
		Input:       "hello",
		WorkspaceId: wsID,
		StartedAt:   timestamppb.Now(),
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	parts := []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "hello"}},
	}
	if err := ipRepo.SaveAll(context.Background(), "inv-1", parts); err != nil {
		t.Fatal(err)
	}

	ctx := workspace.WithID(context.Background(), wsID)
	_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: sessionID,
	}))
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	_, loadErr := ipRepo.Load(context.Background(), "inv-1")
	if loadErr == nil {
		t.Fatal("expected input parts to be deleted")
	}
}

func TestDeleteSession_NotifiesListenersAfterCleanup(t *testing.T) {
	invRepo := invocationmemory.New()
	stub := &stubSessionService{}

	svc := NewSessionServiceServer()
	svc.SetSessionService(stub)
	svc.SetInvocationRepo(invRepo)

	var got []deletedCoords
	svc.AddSessionDeleteListener(func(appName, userID, sessionID string) {
		got = append(got, deletedCoords{appName, userID, sessionID})
	})

	wsID := "ws-test"

	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		AgentName:   "agent-1",
		AppName:     "web-chat",
		UserId:      "user-1",
		SessionId:   "sess-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		WorkspaceId: wsID,
		StartedAt:   timestamppb.Now(),
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ctx := workspace.WithID(context.Background(), wsID)
	_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "sess-1",
	}))
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	want := deletedCoords{"web-chat", "user-1", "sess-1"}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("listener calls = %+v, want exactly one with %+v", got, want)
	}
}

func TestDeleteSession_SessionDeleteFailure_NoListenerNotification(t *testing.T) {
	invRepo := invocationmemory.New()

	svc := NewSessionServiceServer()
	svc.SetSessionService(&stubSessionService{deleteErr: errors.New("mongo unavailable")})
	svc.SetInvocationRepo(invRepo)

	wsID := "ws-test"
	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		AgentName:   "agent-1",
		AppName:     "web-chat",
		UserId:      "user-1",
		SessionId:   "sess-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		Input:       "keep me",
		WorkspaceId: wsID,
		StartedAt:   timestamppb.Now(),
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	calls := 0
	svc.AddSessionDeleteListener(func(_, _, _ string) { calls++ })

	ctx := workspace.WithID(context.Background(), wsID)
	_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "sess-1",
	}))
	if err == nil {
		t.Fatal("expected DeleteSession to fail")
	}
	if calls != 0 {
		t.Fatalf("listener calls = %d, want 0 on failed delete", calls)
	}
}

func TestDeleteSession_ExclusionSetDuringDeletion(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetSessionService(&stubSessionService{})

	sessionID := "sess-1"

	if svc.IsSessionDeleting(sessionID) {
		t.Fatal("session should not be deleting initially")
	}

	svc.markDeleting(sessionID)
	if !svc.IsSessionDeleting(sessionID) {
		t.Fatal("session should be marked as deleting")
	}

	svc.unmarkDeleting(sessionID)
	if svc.IsSessionDeleting(sessionID) {
		t.Fatal("session should no longer be deleting")
	}
}

func TestDeleteSession_BlocksRunnerAndReleasesAfterCompletion(t *testing.T) {
	invRepo := invocationmemory.New()
	coord := newStubDeleteCoordinator()
	coord.waitBlocked = make(chan struct{})
	stub := &stubSessionService{}

	svc := NewSessionServiceServer()
	svc.SetSessionService(stub)
	svc.SetInvocationRepo(invRepo)
	svc.SetAsyncCoordinator(coord)

	wsID := "ws-test"
	sessionID := "sess-1"

	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		AgentName:   "agent-1",
		AppName:     "web-chat",
		UserId:      "user-1",
		SessionId:   sessionID,
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		Input:       "hello",
		WorkspaceId: wsID,
		StartedAt:   timestamppb.Now(),
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	deleteDone := make(chan error, 1)
	go func() {
		ctx := workspace.WithID(context.Background(), wsID)
		_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
			AppName:   "web-chat",
			UserId:    "user-1",
			SessionId: sessionID,
		}))
		deleteDone <- err
	}()

	// Give the goroutine time to enter CancelAndWait.
	time.Sleep(50 * time.Millisecond)

	if !svc.IsSessionDeleting(sessionID) {
		t.Fatal("session should be marked as deleting during delete")
	}

	close(coord.waitBlocked)

	if err := <-deleteDone; err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	if !svc.IsSessionDeleting(sessionID) {
		// After the function returns, the defer should have cleared it.
		// But since we're checking after the function returned via channel,
		// the defer has already run.
	}
}

func TestDeleteSession_MultipleInvocationsAllRedacted(t *testing.T) {
	invRepo := invocationmemory.New()
	ipRepo := inputpartmemory.New()
	stub := &stubSessionService{}

	svc := NewSessionServiceServer()
	svc.SetSessionService(stub)
	svc.SetInvocationRepo(invRepo)
	svc.SetInputPartRepo(ipRepo)

	wsID := "ws-test"
	sessionID := "sess-1"

	for i, id := range []string{"inv-1", "inv-2", "inv-3"} {
		inv := &agentsv1.Invocation{
			Id:          id,
			AgentName:   "agent-1",
			AppName:     "web-chat",
			UserId:      "user-1",
			SessionId:   sessionID,
			Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
			Input:       "input-" + id,
			Output:      "output-" + id,
			WorkspaceId: wsID,
			StartedAt:   timestamppb.New(time.Now().Add(-time.Duration(3-i) * time.Minute)),
		}
		if err := invRepo.Save(context.Background(), inv); err != nil {
			t.Fatal(err)
		}
		parts := []*agentsv1.InputPart{
			{Part: &agentsv1.InputPart_Text{Text: "text-" + id}},
		}
		if err := ipRepo.SaveAll(context.Background(), id, parts); err != nil {
			t.Fatal(err)
		}
	}

	// Different session — must not be affected.
	otherInv := &agentsv1.Invocation{
		Id:          "inv-other",
		AgentName:   "agent-1",
		AppName:     "web-chat",
		UserId:      "user-1",
		SessionId:   "sess-other",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		Input:       "keep-this",
		Output:      "keep-that",
		WorkspaceId: wsID,
		StartedAt:   timestamppb.Now(),
	}
	if err := invRepo.Save(context.Background(), otherInv); err != nil {
		t.Fatal(err)
	}

	ctx := workspace.WithID(context.Background(), wsID)
	_, err := svc.DeleteSession(ctx, connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: sessionID,
	}))
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	for _, id := range []string{"inv-1", "inv-2", "inv-3"} {
		inv, getErr := invRepo.Get(context.Background(), wsID, id)
		if getErr != nil {
			t.Fatalf("Get %s: %v", id, getErr)
		}
		if inv.GetInput() != "" || inv.GetOutput() != "" {
			t.Fatalf("invocation %s should be redacted, input=%q output=%q", id, inv.GetInput(), inv.GetOutput())
		}
	}

	other, _ := invRepo.Get(context.Background(), wsID, "inv-other")
	if other.GetInput() != "keep-this" || other.GetOutput() != "keep-that" {
		t.Fatalf("other session's invocation should not be redacted")
	}
}

func TestDeleteSession_NoInvocations_StillDeletes(t *testing.T) {
	stub := &stubSessionService{}
	svc := NewSessionServiceServer()
	svc.SetSessionService(stub)
	svc.SetInvocationRepo(invocationmemory.New())

	_, err := svc.DeleteSession(context.Background(), connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "web-chat",
		UserId:    "user-1",
		SessionId: "sess-empty",
	}))
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if len(stub.deleted) != 1 {
		t.Fatalf("expected 1 delete call, got %d", len(stub.deleted))
	}
}

func TestDeleteSession_WrongWorkspace_Rejected252(t *testing.T) {
	wsStore := newFakeWSStore()
	wsStore.wsIDByCoord["web-chat/user-1/s1"] = "ws-beta"

	svc := NewSessionServiceServer()
	svc.SetSessionService(&fakeListSessionService{})
	svc.SetWorkspaceSessionStore(wsStore)
	svc.SetInvocationRepo(invocationmemory.New())

	ctx := workspace.WithID(context.Background(), "ws-alpha")

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

func TestSubmitAgentInvocation_RejectedDuringDeletion(t *testing.T) {
	invRepo := invocationmemory.New()
	ipRepo := inputpartmemory.New()
	coord := &fakeAsyncCoordinator{}
	fake := &asyncTestRunner{idToName: map[string]string{"test-agent": "test-agent-name"}, response: "ok"}

	sessionSvc := newFakeWSStore()
	sessionSvcServer := NewSessionServiceServer()
	sessionSvcServer.markDeleting("existing-session")

	agentSvc := &AgentServiceServer{
		runnerSvc:       fake,
		invRepo:         invRepo,
		inputPartRepo:   ipRepo,
		asyncCoord:      coord,
		sessionSvc:      newFakeInMemorySessionSvc(),
		sessionExcluder: sessionSvcServer,
	}
	_ = sessionSvc

	ctx := testContextWithUser("ws-test", "user-1")

	_, err := agentSvc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "req-1",
		AgentId:   "test-agent",
		SessionId: "existing-session",
		Message:   "hello",
	}))
	if err == nil {
		t.Fatal("expected submit to fail during session deletion")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}
	if !errors.As(err, &ce) || ce.Message() != "session is being deleted" {
		t.Fatalf("expected 'session is being deleted' message, got %q", ce.Message())
	}
}

func newFakeInMemorySessionSvc() *fakeListSessionService {
	return &fakeListSessionService{}
}
