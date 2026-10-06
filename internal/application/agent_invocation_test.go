package application

import (
	"context"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/auth"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// GetAgentInvocation by ID and CancelAgentInvocation, as the RPCs serve them
// after the retired dashboard async chat (#410). The GetAgentInvocation tests
// and the helpers moved here unchanged from agent_async_test.go and
// agent_watch_test.go.

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

func testContextWithUser(wsID, userID string) context.Context {
	ctx := workspace.WithID(context.Background(), wsID)
	ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: userID, Role: "member"}, nil)
	return ctx
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

func TestGetAgentInvocation_PrivateOwnershipEnforced(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}

	inv := &agentsv1.Invocation{
		Id:          "inv-owned",
		AppName:     "web-chat",
		UserId:      "owner-user",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		Source:      "dashboard-async",
		WorkspaceId: wsTest,
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	// Another member of the same workspace is refused with NotFound.
	otherCtx := testContextWithUser(wsTest, "other-user")
	_, err := svc.GetAgentInvocation(otherCtx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{InvocationId: "inv-owned"}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("non-owner err = %v, want NotFound", err)
	}

	// The owner can read it.
	ownerCtx := testContextWithUser(wsTest, "owner-user")
	if _, err := svc.GetAgentInvocation(ownerCtx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{InvocationId: "inv-owned"})); err != nil {
		t.Fatalf("owner read failed: %v", err)
	}

	// A global admin retains the support path.
	adminCtx := workspace.WithID(context.Background(), wsTest)
	adminCtx = auth.WithAuthenticated(adminCtx, &agentsv1.User{Id: "admin-user", Role: "admin"}, nil)
	if _, err := svc.GetAgentInvocation(adminCtx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{InvocationId: "inv-owned"})); err != nil {
		t.Fatalf("admin read failed: %v", err)
	}
}

// Another member cannot cancel a private chat Invocation, and the refused
// cancel never reaches the runner.
func TestCancelAgentInvocation_RejectsOtherPrivateSessionOwner(t *testing.T) {
	invRepo := invocationmemory.New()
	runner := &cancelCountingRunner{}
	svc := &AgentServiceServer{
		runnerSvc: runner,
		invRepo:   invRepo,
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
	if runner.cancels != 0 {
		t.Fatalf("unauthorized cancellation reached the runner %d times", runner.cancels)
	}
}
