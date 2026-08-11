package asyncrun

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeRunner struct {
	mu       sync.Mutex
	block    chan struct{}
	response string
	err      error
	calls    int
}

func (r *fakeRunner) RunSSE(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
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
func (r *fakeRunner) ResolveAgentRef(_, agentID string) (string, bool) { return agentID, true }
func (r *fakeRunner) GetAgentIdentity(name string) (string, string, bool) {
	return name, name, true
}
func (r *fakeRunner) CancelInvocation(string, string) bool { return false }

func TestCoordinator_EnqueueAndComplete(t *testing.T) {
	repo := memory.New()
	fr := &fakeRunner{response: "done"}
	coord := New(repo, fr, Config{MaxRunDuration: 5 * time.Second})

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

	parts := []*genai.Part{genai.NewPartFromText("hello")}
	coord.Enqueue(inv, "test", parts, "")

	// Wait for completion.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for invocation to complete")
		default:
		}
		got, err := repo.Get(context.Background(), "inv-1")
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
}

func TestCoordinator_Cancel(t *testing.T) {
	repo := memory.New()
	block := make(chan struct{})
	fr := &fakeRunner{block: block, response: "done"}
	coord := New(repo, fr, Config{MaxRunDuration: 5 * time.Second})

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

	parts := []*genai.Part{genai.NewPartFromText("hello")}
	coord.Enqueue(inv, "test", parts, "")

	// Wait for it to become RUNNING.
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for RUNNING")
		default:
		}
		got, _ := repo.Get(context.Background(), "inv-cancel")
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
		got, _ := repo.Get(context.Background(), "inv-cancel")
		if got != nil && got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCoordinator_RunError(t *testing.T) {
	repo := memory.New()
	fr := &fakeRunner{err: errors.New("model error")}
	coord := New(repo, fr, Config{MaxRunDuration: 5 * time.Second})

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

	parts := []*genai.Part{genai.NewPartFromText("hello")}
	coord.Enqueue(inv, "test", parts, "")

	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for FAILED")
		default:
		}
		got, _ := repo.Get(context.Background(), "inv-fail")
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

	got1, _ := repo.Get(context.Background(), "stale-1")
	if got1.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("stale-1 status = %v, want FAILED", got1.GetStatus())
	}
	got2, _ := repo.Get(context.Background(), "stale-2")
	if got2.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("stale-2 status = %v, want FAILED", got2.GetStatus())
	}
	got3, _ := repo.Get(context.Background(), "ok-1")
	if got3.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("ok-1 status = %v, want SUCCEEDED", got3.GetStatus())
	}
}
