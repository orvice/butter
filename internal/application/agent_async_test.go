package application

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// asyncTestRunner is a controllable fake runner for async tests.
type asyncTestRunner struct {
	mu       sync.Mutex
	idToName map[string]string
	calls    int
	block    chan struct{} // blocks RunSSE when non-nil
	response string
	err      error
}

func (r *asyncTestRunner) IsReservedAgentName(string) bool { return false }
func (r *asyncTestRunner) Run(_ context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	return r.response, r.err
}
func (r *asyncTestRunner) RunSSE(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.mu.Lock()
	r.calls++
	block := r.block
	r.mu.Unlock()
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return r.response, r.err
}
func (r *asyncTestRunner) CancelInvocation(string, string) bool { return false }
func (r *asyncTestRunner) ResolveAgentRef(_, agentID string) (string, bool) {
	name, ok := r.idToName[agentID]
	return name, ok
}
func (r *asyncTestRunner) GetAgentIdentity(name string) (string, string, bool) {
	for id, n := range r.idToName {
		if n == name {
			return id, name, true
		}
	}
	return "", name, true
}

// fakeAsyncCoordinator captures Enqueue calls for testing without real
// goroutines.
type fakeAsyncCoordinator struct {
	mu              sync.Mutex
	enqueued        []*agentsv1.Invocation
	cancelled       bool
	cancelledID     string
	cancelWorkspace string
}

func (c *fakeAsyncCoordinator) Enqueue(inv *agentsv1.Invocation, _ string, _ []*genai.Part, _ string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enqueued = append(c.enqueued, inv)
}
func (c *fakeAsyncCoordinator) Cancel(invocationID, workspaceID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelledID = invocationID
	c.cancelWorkspace = workspaceID
	return c.cancelled
}

type delayedActiveLookupRepo struct {
	invocation.Repository
	delay time.Duration
}

func (r *delayedActiveLookupRepo) FindActiveBySession(ctx context.Context, workspaceID, sessionID string) (*agentsv1.Invocation, error) {
	time.Sleep(r.delay)
	return r.Repository.FindActiveBySession(ctx, workspaceID, sessionID)
}

func testContextWithUser(wsID, userID string) context.Context {
	ctx := workspace.WithID(context.Background(), wsID)
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: userID, Role: "member"}, nil)
	return ctx
}

func TestSubmitAgentInvocation_Basic(t *testing.T) {
	invRepo := invocationmemory.New()
	coord := &fakeAsyncCoordinator{}
	fake := &asyncTestRunner{idToName: map[string]string{"test-agent": "test-agent-name"}, response: "ok"}

	svc := &AgentServiceServer{
		runnerSvc:  fake,
		invRepo:    invRepo,
		asyncCoord: coord,
		sessionSvc: session.InMemoryService(),
	}

	ctx := testContextWithUser(wsTest, "user-1")

	resp, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "req-1",
		AgentId:   "test-agent",
		Message:   "hello world",
	}))
	if err != nil {
		t.Fatal(err)
	}

	if resp.Msg.GetSessionId() == "" {
		t.Fatal("expected session_id to be set")
	}
	if resp.Msg.GetInvocationId() == "" {
		t.Fatal("expected invocation_id to be set")
	}
	if resp.Msg.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED {
		t.Fatalf("status = %v, want QUEUED", resp.Msg.GetStatus())
	}
	if !resp.Msg.GetSessionCreated() {
		t.Fatal("expected session_created = true")
	}
	if len(coord.enqueued) != 1 {
		t.Fatalf("expected 1 enqueued invocation, got %d", len(coord.enqueued))
	}
}

func TestSubmitAgentInvocation_Idempotency(t *testing.T) {
	invRepo := invocationmemory.New()
	coord := &fakeAsyncCoordinator{}
	fake := &asyncTestRunner{idToName: map[string]string{"test-agent": "test-agent-name"}, response: "ok"}

	svc := &AgentServiceServer{
		runnerSvc:  fake,
		invRepo:    invRepo,
		asyncCoord: coord,
		sessionSvc: session.InMemoryService(),
	}

	ctx := testContextWithUser(wsTest, "user-1")

	resp1, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "req-dup",
		AgentId:   "test-agent",
		Message:   "hello",
	}))
	if err != nil {
		t.Fatal(err)
	}

	// Submit with the same request_id again.
	resp2, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "req-dup",
		AgentId:   "test-agent",
		Message:   "hello",
	}))
	if err != nil {
		t.Fatal(err)
	}

	if resp1.Msg.GetInvocationId() != resp2.Msg.GetInvocationId() {
		t.Fatalf("idempotent retry returned different invocation_id: %q vs %q",
			resp1.Msg.GetInvocationId(), resp2.Msg.GetInvocationId())
	}
	if resp1.Msg.GetSessionId() != resp2.Msg.GetSessionId() {
		t.Fatalf("idempotent retry returned different session_id: %q vs %q",
			resp1.Msg.GetSessionId(), resp2.Msg.GetSessionId())
	}
	// Only one actual enqueue should have happened.
	if len(coord.enqueued) != 1 {
		t.Fatalf("expected 1 enqueue, got %d", len(coord.enqueued))
	}
}

func TestSubmitAgentInvocation_SingleActivePerSession(t *testing.T) {
	invRepo := invocationmemory.New()
	coord := &fakeAsyncCoordinator{}
	fake := &asyncTestRunner{idToName: map[string]string{"test-agent": "test-agent-name"}, response: "ok"}

	svc := &AgentServiceServer{
		runnerSvc:  fake,
		invRepo:    invRepo,
		asyncCoord: coord,
		sessionSvc: session.InMemoryService(),
	}

	ctx := testContextWithUser(wsTest, "user-1")

	// First submission creates a session.
	resp, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "first",
		AgentId:   "test-agent",
		Message:   "hello",
	}))
	if err != nil {
		t.Fatal(err)
	}
	sessionID := resp.Msg.GetSessionId()

	// Second submission to the same session while the first is still QUEUED.
	_, err = svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "second",
		AgentId:   "test-agent",
		SessionId: sessionID,
		Message:   "blocked",
	}))
	if err == nil {
		t.Fatal("expected error for concurrent submission to same session")
	}
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v: %v", connect.CodeOf(err), err)
	}
	if !strings.Contains(err.Error(), resp.Msg.GetInvocationId()) {
		t.Fatalf("error %q does not include active invocation id %q", err, resp.Msg.GetInvocationId())
	}
	if got := err.(*connect.Error).Meta().Get("active-invocation-id"); got != resp.Msg.GetInvocationId() {
		t.Fatalf("active-invocation-id metadata = %q, want %q", got, resp.Msg.GetInvocationId())
	}
}

func TestSubmitAgentInvocation_SimultaneousSubmitsCannotBypassSingleActiveRule(t *testing.T) {
	baseRepo := invocationmemory.New()
	invRepo := &delayedActiveLookupRepo{Repository: baseRepo, delay: 40 * time.Millisecond}
	coord := &fakeAsyncCoordinator{}
	fake := &asyncTestRunner{idToName: map[string]string{"test-agent": "test-agent-name"}}
	sessionSvc := session.InMemoryService()
	ctx := testContextWithUser(wsTest, "user-1")

	created, err := sessionSvc.Create(ctx, &session.CreateRequest{
		AppName:   "web-chat",
		UserID:    "user-1",
		SessionID: "shared-session",
		State:     map[string]any{"workspace_id": wsTest, "agent_name": "test-agent-name"},
	})
	if err != nil {
		t.Fatal(err)
	}

	svc := &AgentServiceServer{
		runnerSvc:  fake,
		invRepo:    invRepo,
		asyncCoord: coord,
		sessionSvc: sessionSvc,
	}

	start := make(chan struct{})
	type result struct {
		resp *connect.Response[agentsv1.SubmitAgentInvocationResponse]
		err  error
	}
	results := make(chan result, 2)
	for _, requestID := range []string{"simultaneous-1", "simultaneous-2"} {
		requestID := requestID
		go func() {
			<-start
			resp, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
				RequestId: requestID,
				AgentId:   "test-agent",
				SessionId: created.Session.ID(),
				Message:   requestID,
			}))
			results <- result{resp: resp, err: err}
		}()
	}
	close(start)

	var successfulID string
	var failed error
	for range 2 {
		result := <-results
		if result.err == nil {
			if successfulID != "" {
				t.Fatal("both simultaneous submissions succeeded")
			}
			successfulID = result.resp.Msg.GetInvocationId()
			continue
		}
		failed = result.err
	}
	if successfulID == "" {
		t.Fatal("neither simultaneous submission succeeded")
	}
	if connect.CodeOf(failed) != connect.CodeFailedPrecondition {
		t.Fatalf("losing submission error = %v, want FailedPrecondition", failed)
	}
	if !strings.Contains(failed.Error(), successfulID) {
		t.Fatalf("losing submission error %q does not include active invocation id %q", failed, successfulID)
	}
	if len(coord.enqueued) != 1 {
		t.Fatalf("enqueued %d invocations, want 1", len(coord.enqueued))
	}
}

func TestSubmitAgentInvocation_DifferentSessionsMayRunConcurrently(t *testing.T) {
	invRepo := invocationmemory.New()
	coord := &fakeAsyncCoordinator{}
	fake := &asyncTestRunner{idToName: map[string]string{"test-agent": "test-agent-name"}}
	sessionSvc := session.InMemoryService()
	ctx := testContextWithUser(wsTest, "user-1")

	for _, sessionID := range []string{"session-a", "session-b"} {
		if _, err := sessionSvc.Create(ctx, &session.CreateRequest{
			AppName:   "web-chat",
			UserID:    "user-1",
			SessionID: sessionID,
			State: map[string]any{
				"workspace_id": wsTest,
				"agent_id":     "test-agent",
				"agent_name":   "test-agent-name",
			},
		}); err != nil {
			t.Fatal(err)
		}
	}

	svc := &AgentServiceServer{
		runnerSvc:  fake,
		invRepo:    invRepo,
		asyncCoord: coord,
		sessionSvc: sessionSvc,
	}
	for index, sessionID := range []string{"session-a", "session-b"} {
		if _, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
			RequestId: "cross-session-" + sessionID,
			AgentId:   "test-agent",
			SessionId: sessionID,
			Message:   "message",
		})); err != nil {
			t.Fatalf("submit %d to %s: %v", index, sessionID, err)
		}
	}
	if len(coord.enqueued) != 2 {
		t.Fatalf("enqueued %d invocations, want 2", len(coord.enqueued))
	}
}

func TestSubmitAgentInvocation_RequiresRequestID(t *testing.T) {
	svc := &AgentServiceServer{
		runnerSvc:  &asyncTestRunner{idToName: map[string]string{"a": "a"}},
		invRepo:    invocationmemory.New(),
		asyncCoord: &fakeAsyncCoordinator{},
	}
	ctx := testContextWithUser(wsTest, "user-1")

	_, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		AgentId: "a",
		Message: "hi",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for missing request_id, got %v: %v", connect.CodeOf(err), err)
	}
}

func TestSubmitAgentInvocation_RequiresAgentID(t *testing.T) {
	svc := &AgentServiceServer{
		runnerSvc:  &asyncTestRunner{idToName: map[string]string{"a": "a"}},
		invRepo:    invocationmemory.New(),
		asyncCoord: &fakeAsyncCoordinator{},
	}
	ctx := testContextWithUser(wsTest, "user-1")

	_, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "r1",
		Message:   "hi",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for missing agent_id, got %v: %v", connect.CodeOf(err), err)
	}
}

func TestSubmitAgentInvocation_RequiresAuth(t *testing.T) {
	svc := &AgentServiceServer{
		runnerSvc:  &asyncTestRunner{idToName: map[string]string{"a": "a"}},
		invRepo:    invocationmemory.New(),
		asyncCoord: &fakeAsyncCoordinator{},
	}
	// Context with workspace but no user.
	ctx := workspace.WithID(context.Background(), wsTest)

	_, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "r1",
		AgentId:   "a",
		Message:   "hi",
	}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("expected Unauthenticated, got %v: %v", connect.CodeOf(err), err)
	}
}

func TestGetAgentInvocation_Basic(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}

	ctx := testContextWithUser(wsTest, "user-1")

	// Save an invocation.
	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		WorkspaceId: wsTest,
		AgentName:   "test",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		Output:      "result",
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	resp, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		InvocationId: "inv-1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetInvocation().GetId() != "inv-1" {
		t.Fatalf("got id %q, want inv-1", resp.Msg.GetInvocation().GetId())
	}
	if resp.Msg.GetInvocation().GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("got status %v, want SUCCEEDED", resp.Msg.GetInvocation().GetStatus())
	}
}

func TestGetAgentInvocation_NotFound(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}
	ctx := testContextWithUser(wsTest, "user-1")

	_, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		InvocationId: "nonexistent",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v: %v", connect.CodeOf(err), err)
	}
}

func TestGetAgentInvocation_WorkspaceIsolation(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}

	// Invocation belongs to "other-workspace".
	inv := &agentsv1.Invocation{
		Id:          "inv-other",
		WorkspaceId: "other-workspace",
		AgentName:   "test",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ctx := testContextWithUser(wsTest, "user-1")
	_, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		InvocationId: "inv-other",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("expected NotFound for wrong workspace, got %v: %v", connect.CodeOf(err), err)
	}
}

func TestGetAgentInvocation_PrivateSessionOwnership(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}
	if err := invRepo.Save(context.Background(), &agentsv1.Invocation{
		Id:          "inv-private",
		WorkspaceId: wsTest,
		UserId:      "user-1",
		AppName:     "web-chat",
		Source:      "dashboard-async",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
	}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.GetAgentInvocation(testContextWithUser(wsTest, "user-2"), connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		InvocationId: "inv-private",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("other user lookup error = %v, want NotFound", err)
	}
}

func TestCancelAgentInvocation_QueuedDashboardInvocation(t *testing.T) {
	invRepo := invocationmemory.New()
	coord := &fakeAsyncCoordinator{cancelled: true}
	svc := &AgentServiceServer{
		runnerSvc:  &asyncTestRunner{},
		invRepo:    invRepo,
		asyncCoord: coord,
	}
	if err := invRepo.Save(context.Background(), &agentsv1.Invocation{
		Id:          "inv-queued",
		WorkspaceId: wsTest,
		UserId:      "user-1",
		AppName:     "web-chat",
		Source:      "dashboard-async",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
	}); err != nil {
		t.Fatal(err)
	}

	resp, err := svc.CancelAgentInvocation(testContextWithUser(wsTest, "user-1"), connect.NewRequest(&agentsv1.CancelAgentInvocationRequest{
		InvocationId: "inv-queued",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Msg.GetCancelled() {
		t.Fatal("cancelled = false, want true")
	}
	if coord.cancelledID != "inv-queued" || coord.cancelWorkspace != wsTest {
		t.Fatalf("coordinator cancel = (%q, %q), want (inv-queued, %s)", coord.cancelledID, coord.cancelWorkspace, wsTest)
	}
}

func TestCancelAgentInvocation_RejectsOtherPrivateSessionOwner(t *testing.T) {
	invRepo := invocationmemory.New()
	coord := &fakeAsyncCoordinator{cancelled: true}
	svc := &AgentServiceServer{
		runnerSvc:  &asyncTestRunner{},
		invRepo:    invRepo,
		asyncCoord: coord,
	}
	if err := invRepo.Save(context.Background(), &agentsv1.Invocation{
		Id:          "inv-owned",
		WorkspaceId: wsTest,
		UserId:      "user-1",
		AppName:     "web-chat",
		Source:      "dashboard-async",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
	}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.CancelAgentInvocation(testContextWithUser(wsTest, "user-2"), connect.NewRequest(&agentsv1.CancelAgentInvocationRequest{
		InvocationId: "inv-owned",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("other user cancellation error = %v, want NotFound", err)
	}
	if coord.cancelledID != "" {
		t.Fatalf("unauthorized cancellation reached coordinator for %q", coord.cancelledID)
	}
}
