package asyncrun

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	inputpartmemory "go.orx.me/apps/butter/internal/repo/inputpart/memory"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeRunner struct {
	mu           sync.Mutex
	block        chan struct{}
	ignoreCancel bool
	response     string
	err          error
	calls        int
}

type blockingStatusSaveRepo struct {
	invocation.Repository
	status  agentsv1.InvocationStatus
	started chan struct{}
	release chan struct{}
}

func (r *blockingStatusSaveRepo) Save(ctx context.Context, inv *agentsv1.Invocation) error {
	if inv.GetStatus() == r.status {
		close(r.started)
		<-r.release
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return r.Repository.Save(ctx, inv)
}

func (r *fakeRunner) RunSSE(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.mu.Lock()
	r.calls++
	block := r.block
	r.mu.Unlock()
	if block != nil {
		if r.ignoreCancel {
			<-block
			return r.response, r.err
		}
		select {
		case <-block:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return r.response, r.err
}

func TestCoordinator_ExplicitCancelWinsWhenRunnerReturnsNormally(t *testing.T) {
	repo := memory.New()
	block := make(chan struct{})
	ipRepo := inputpartmemory.New()
	fr := &fakeRunner{block: block, ignoreCancel: true, response: "late output"}
	coord := New(repo, ipRepo, fr, Config{MaxRunDuration: 5 * time.Second})

	inv := &agentsv1.Invocation{
		Id:          "inv-cancel-normal-return",
		AgentName:   "test",
		SessionId:   "sess-1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		StartedAt:   timestamppb.Now(),
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ipRepo.SaveAll(context.Background(), "inv-cancel-normal-return", []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "hello"}},
	})
	coord.Enqueue(inv, "test", "")
	deadline := time.After(2 * time.Second)
	for {
		got, _ := repo.GetAcrossWorkspaces(context.Background(), inv.GetId())
		if got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for RUNNING")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	if !coord.Cancel(inv.GetId(), "ws-1") {
		t.Fatal("Cancel returned false")
	}
	close(block)

	deadline = time.After(2 * time.Second)
	for {
		got, _ := repo.GetAcrossWorkspaces(context.Background(), inv.GetId())
		if got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED {
			if got.GetOutput() != "" {
				t.Fatalf("cancelled invocation output = %q, want empty", got.GetOutput())
			}
			break
		}
		if got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			t.Fatal("explicitly cancelled invocation became SUCCEEDED")
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for CANCELLED")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestCoordinator_QueuedCancelPersistsCancelledWithDetachedContext(t *testing.T) {
	base := memory.New()
	repo := &blockingStatusSaveRepo{
		Repository: base,
		status:     agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		started:    make(chan struct{}),
		release:    make(chan struct{}),
	}
	ipRepo := inputpartmemory.New()
	fr := &fakeRunner{response: "must not run"}
	coord := New(repo, ipRepo, fr, Config{MaxRunDuration: 5 * time.Second})

	inv := &agentsv1.Invocation{
		Id:          "inv-queued-cancel",
		AgentName:   "test",
		SessionId:   "sess-1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		StartedAt:   timestamppb.Now(),
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ipRepo.SaveAll(context.Background(), "inv-queued-cancel", []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "hello"}},
	})
	coord.Enqueue(inv, "test", "")
	select {
	case <-repo.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for RUNNING save")
	}
	if !coord.Cancel(inv.GetId(), "ws-1") {
		t.Fatal("Cancel returned false")
	}
	close(repo.release)

	deadline := time.After(2 * time.Second)
	for {
		got, _ := base.GetAcrossWorkspaces(context.Background(), inv.GetId())
		if got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("status = %v, want CANCELLED", got.GetStatus())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	if fr.calls != 0 {
		t.Fatalf("runner calls = %d, want 0", fr.calls)
	}
}

func TestCoordinator_CancelAfterTerminalClaimReturnsFalse(t *testing.T) {
	base := memory.New()
	repo := &blockingStatusSaveRepo{
		Repository: base,
		status:     agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		started:    make(chan struct{}),
		release:    make(chan struct{}),
	}
	ipRepo := inputpartmemory.New()
	fr := &fakeRunner{response: "done"}
	coord := New(repo, ipRepo, fr, Config{MaxRunDuration: 5 * time.Second})

	inv := &agentsv1.Invocation{
		Id:          "inv-finished-before-stop",
		AgentName:   "test",
		SessionId:   "sess-1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		StartedAt:   timestamppb.Now(),
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ipRepo.SaveAll(context.Background(), "inv-finished-before-stop", []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "hello"}},
	})
	coord.Enqueue(inv, "test", "")
	select {
	case <-repo.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for terminal save")
	}
	if coord.Cancel(inv.GetId(), "ws-1") {
		t.Fatal("Cancel returned true after terminal state was claimed")
	}
	close(repo.release)

	deadline := time.After(2 * time.Second)
	for {
		got, _ := base.GetAcrossWorkspaces(context.Background(), inv.GetId())
		if got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("status = %v, want SUCCEEDED", got.GetStatus())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
func (r *fakeRunner) ResolveAgentRef(_, agentID string) (string, bool) { return agentID, true }
func (r *fakeRunner) GetAgentIdentity(name string) (string, string, bool) {
	return name, name, true
}
func (r *fakeRunner) CancelInvocation(string, string) bool { return false }

func TestCoordinator_EnqueueAndComplete(t *testing.T) {
	repo := memory.New()
	ipRepo := inputpartmemory.New()
	fr := &fakeRunner{response: "done"}
	coord := New(repo, ipRepo, fr, Config{MaxRunDuration: 5 * time.Second})

	var completedMu sync.Mutex
	var completedInv *agentsv1.Invocation
	coord.SetTurnCompleteCallback(func(_ context.Context, inv *agentsv1.Invocation) {
		completedMu.Lock()
		completedInv = inv
		completedMu.Unlock()
	})

	inv := &agentsv1.Invocation{
		Id:          "inv-1",
		AgentName:   "test",
		SessionId:   "sess-1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		StartedAt:   timestamppb.Now(),
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	// Persist input parts before enqueue (as the RPC handler does).
	ipRepo.SaveAll(context.Background(), "inv-1", []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "hello"}},
	})
	coord.Enqueue(inv, "test", "")

	// Wait for completion.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for invocation to complete")
		default:
		}
		got, err := repo.GetAcrossWorkspaces(context.Background(), "inv-1")
		if err != nil {
			t.Fatal(err)
		}
		if got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			if got.GetOutput() != "done" {
				t.Fatalf("output = %q, want done", got.GetOutput())
			}
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Verify turn complete callback was called.
	time.Sleep(50 * time.Millisecond)
	completedMu.Lock()
	if completedInv == nil {
		t.Fatal("turn complete callback not called")
	}
	completedMu.Unlock()

	// Verify input parts were cleaned up after successful execution.
	_, loadErr := ipRepo.Load(context.Background(), "inv-1")
	if loadErr == nil {
		t.Fatal("expected input parts to be deleted after success")
	}
}

func TestCoordinator_FailedRunRetainsInputParts(t *testing.T) {
	repo := memory.New()
	ipRepo := inputpartmemory.New()
	fr := &fakeRunner{err: errors.New("boom")}
	coord := New(repo, ipRepo, fr, Config{MaxRunDuration: 5 * time.Second})

	inv := &agentsv1.Invocation{
		Id:          "inv-retain",
		AgentName:   "test",
		SessionId:   "sess-1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		StartedAt:   timestamppb.Now(),
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}
	ipRepo.SaveAll(context.Background(), "inv-retain", []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "keep me"}},
	})
	coord.Enqueue(inv, "test", "")

	// Wait for FAILED status.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for FAILED")
		default:
		}
		got, _ := repo.Get(context.Background(), "ws-1", "inv-retain")
		if got != nil && got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Input parts should still be loadable for recovery.
	parts, loadErr := ipRepo.Load(context.Background(), "inv-retain")
	if loadErr != nil {
		t.Fatalf("expected input parts to be retained after failure, got: %v", loadErr)
	}
	if len(parts) != 1 || parts[0].GetText() != "keep me" {
		t.Fatalf("unexpected parts: %+v", parts)
	}
}

func TestCoordinator_Cancel(t *testing.T) {
	repo := memory.New()
	ipRepo := inputpartmemory.New()
	block := make(chan struct{})
	fr := &fakeRunner{block: block, response: "done"}
	coord := New(repo, ipRepo, fr, Config{MaxRunDuration: 5 * time.Second})

	inv := &agentsv1.Invocation{
		Id:          "inv-cancel",
		AgentName:   "test",
		SessionId:   "sess-1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		StartedAt:   timestamppb.Now(),
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ipRepo.SaveAll(context.Background(), "inv-cancel", []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "hello"}},
	})
	coord.Enqueue(inv, "test", "")

	// Wait for it to become RUNNING.
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for RUNNING")
		default:
		}
		got, _ := repo.GetAcrossWorkspaces(context.Background(), "inv-cancel")
		if got != nil && got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Cancel it.
	if !coord.Cancel("inv-cancel", "ws-1") {
		t.Fatal("Cancel returned false")
	}

	// Wait for terminal status.
	deadline = time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for CANCELLED")
		default:
		}
		got, _ := repo.GetAcrossWorkspaces(context.Background(), "inv-cancel")
		if got != nil && got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCoordinator_RunError(t *testing.T) {
	repo := memory.New()
	ipRepo := inputpartmemory.New()
	fr := &fakeRunner{err: errors.New("model error")}
	coord := New(repo, ipRepo, fr, Config{MaxRunDuration: 5 * time.Second})

	inv := &agentsv1.Invocation{
		Id:          "inv-fail",
		AgentName:   "test",
		SessionId:   "sess-1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		StartedAt:   timestamppb.Now(),
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ipRepo.SaveAll(context.Background(), "inv-fail", []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "hello"}},
	})
	coord.Enqueue(inv, "test", "")

	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for FAILED")
		default:
		}
		got, _ := repo.GetAcrossWorkspaces(context.Background(), "inv-fail")
		if got != nil && got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
			if got.GetError() != "model error" {
				t.Fatalf("error = %q, want model error", got.GetError())
			}
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestReconcileStale(t *testing.T) {
	repo := memory.New()
	repo.Save(context.Background(), &agentsv1.Invocation{
		Id:          "stale-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		WorkspaceId: "ws-1",
		StartedAt:   timestamppb.Now(),
	})
	repo.Save(context.Background(), &agentsv1.Invocation{
		Id:          "stale-2",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		WorkspaceId: "ws-1",
		StartedAt:   timestamppb.Now(),
	})
	repo.Save(context.Background(), &agentsv1.Invocation{
		Id:          "ok-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		WorkspaceId: "ws-1",
		StartedAt:   timestamppb.Now(),
	})

	n, err := ReconcileStale(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 stale, got %d", n)
	}

	got1, _ := repo.GetAcrossWorkspaces(context.Background(), "stale-1")
	if got1.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("stale-1 status = %v, want FAILED", got1.GetStatus())
	}
	got2, _ := repo.GetAcrossWorkspaces(context.Background(), "stale-2")
	if got2.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("stale-2 status = %v, want FAILED", got2.GetStatus())
	}
	got3, _ := repo.GetAcrossWorkspaces(context.Background(), "ok-1")
	if got3.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("ok-1 status = %v, want SUCCEEDED", got3.GetStatus())
	}
}
