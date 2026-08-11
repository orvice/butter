package asyncrun

import (
	"context"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	inputpartmemory "go.orx.me/apps/butter/internal/repo/inputpart/memory"
	"go.orx.me/apps/butter/internal/repo/invocation/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func queuedInvocation(id string) *agentsv1.Invocation {
	return &agentsv1.Invocation{
		Id:          id,
		AgentName:   "test",
		SessionId:   "sess-1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		StartedAt:   timestamppb.Now(),
	}
}

func seedQueued(t *testing.T, repo *memory.Store, ipRepo *inputpartmemory.Store, id string) *agentsv1.Invocation {
	t.Helper()
	inv := queuedInvocation(id)
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}
	if err := ipRepo.SaveAll(context.Background(), id, []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "hello"}},
	}); err != nil {
		t.Fatal(err)
	}
	return inv
}

// waitForStatus polls inside a synctest bubble until the invocation reaches
// the wanted status. synctest.Wait blocks until every other goroutine in the
// bubble is durably idle, so each iteration observes a settled state.
func waitForStatus(t *testing.T, repo *memory.Store, id string, want agentsv1.InvocationStatus) *agentsv1.Invocation {
	t.Helper()
	for range 1000 {
		synctest.Wait()
		got, err := repo.GetAcrossWorkspaces(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		if got.GetStatus() == want {
			return got
		}
	}
	got, _ := repo.GetAcrossWorkspaces(context.Background(), id)
	t.Fatalf("status = %v, want %v", got.GetStatus(), want)
	return nil
}

func TestCoordinator_TimeoutFailsWithActionableReason(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New()
		ipRepo := inputpartmemory.New()
		fr := &fakeRunner{block: make(chan struct{})} // blocks until ctx deadline
		coord := New(repo, ipRepo, fr, Config{})      // default 30-minute maximum

		inv := seedQueued(t, repo, ipRepo, "inv-timeout")
		coord.Enqueue(inv, "test", "")

		// The runner never unblocks. Once every goroutine in the bubble is
		// idle the fake clock advances past MaxRunDuration and the run's
		// context deadline fires.
		time.Sleep(30*time.Minute + time.Second)
		got := waitForStatus(t, repo, "inv-timeout", agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED)

		if !strings.Contains(got.GetError(), "maximum run duration (30m0s)") {
			t.Fatalf("error = %q, want the configured duration in an actionable reason", got.GetError())
		}
		if !strings.Contains(got.GetError(), "resubmit") {
			t.Fatalf("error = %q, want a resubmit hint", got.GetError())
		}
		if got.GetFinishedAt() == nil {
			t.Fatal("finished_at not set")
		}
		if fr.calls != 1 {
			t.Fatalf("runner calls = %d, want 1 (timeout must not retry)", fr.calls)
		}
		// Input parts stay retained so the user can restore and resubmit.
		if _, err := ipRepo.Load(context.Background(), "inv-timeout"); err != nil {
			t.Fatalf("expected input parts retained after timeout, got: %v", err)
		}
	})
}

func TestCoordinator_ShutdownFailsInFlightRunHonestly(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New()
		ipRepo := inputpartmemory.New()
		fr := &fakeRunner{block: make(chan struct{})}
		coord := New(repo, ipRepo, fr, Config{})

		inv := seedQueued(t, repo, ipRepo, "inv-shutdown")
		coord.Enqueue(inv, "test", "")
		waitForStatus(t, repo, "inv-shutdown", agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING)

		// Observers must receive the terminal frame before their stream closes.
		frames, unsubscribe := coord.Watch("inv-shutdown")
		defer unsubscribe()

		if err := coord.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown returned %v", err)
		}

		got, err := repo.GetAcrossWorkspaces(context.Background(), "inv-shutdown")
		if err != nil {
			t.Fatal(err)
		}
		if got.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
			t.Fatalf("status = %v, want FAILED (shutdown is a system failure, not user cancellation)", got.GetStatus())
		}
		if !strings.Contains(got.GetError(), "shutdown") || !strings.Contains(got.GetError(), "resubmit") {
			t.Fatalf("error = %q, want an honest shutdown reason with a resubmit hint", got.GetError())
		}

		var sawTerminal bool
		for f := range frames {
			if f.Terminal() {
				sawTerminal = true
				if f.Invocation.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
					t.Fatalf("terminal frame status = %v, want FAILED", f.Invocation.GetStatus())
				}
			}
		}
		if !sawTerminal {
			t.Fatal("observer stream closed without a terminal state frame")
		}
	})
}

func TestCoordinator_ShutdownRacingCleanFinishStaysSucceeded(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New()
		ipRepo := inputpartmemory.New()
		block := make(chan struct{})
		// The runner ignores cancellation and completes with full output —
		// the shutdown signal must not relabel a finished run as FAILED.
		fr := &fakeRunner{block: block, ignoreCancel: true, response: "complete answer"}
		coord := New(repo, ipRepo, fr, Config{})

		inv := seedQueued(t, repo, ipRepo, "inv-shutdown-vs-finish")
		coord.Enqueue(inv, "test", "")
		waitForStatus(t, repo, "inv-shutdown-vs-finish", agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING)

		shutdownDone := make(chan struct{})
		go func() {
			defer close(shutdownDone)
			_ = coord.Shutdown(context.Background())
		}()
		// Shutdown has marked the entry and is now waiting on the in-flight
		// run; release the runner so it finishes cleanly.
		synctest.Wait()
		close(block)
		<-shutdownDone

		got, _ := repo.GetAcrossWorkspaces(context.Background(), "inv-shutdown-vs-finish")
		if got.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			t.Fatalf("status = %v, want SUCCEEDED (clean finish must beat the shutdown label)", got.GetStatus())
		}
		if got.GetOutput() != "complete answer" {
			t.Fatalf("output = %q, want the complete response", got.GetOutput())
		}
	})
}

func TestCoordinator_UserCancelBeatsShutdown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New()
		ipRepo := inputpartmemory.New()
		fr := &fakeRunner{block: make(chan struct{})}
		coord := New(repo, ipRepo, fr, Config{})

		inv := seedQueued(t, repo, ipRepo, "inv-cancel-then-shutdown")
		coord.Enqueue(inv, "test", "")
		waitForStatus(t, repo, "inv-cancel-then-shutdown", agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING)

		if !coord.Cancel("inv-cancel-then-shutdown", "ws-1") {
			t.Fatal("Cancel returned false")
		}
		if err := coord.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown returned %v", err)
		}

		got, _ := repo.GetAcrossWorkspaces(context.Background(), "inv-cancel-then-shutdown")
		if got.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED {
			t.Fatalf("status = %v, want CANCELLED (explicit Stop must not be reported as a failure)", got.GetStatus())
		}
	})
}

func TestCoordinator_EnqueueAfterShutdownFailsImmediately(t *testing.T) {
	repo := memory.New()
	ipRepo := inputpartmemory.New()
	fr := &fakeRunner{response: "must not run"}
	coord := New(repo, ipRepo, fr, Config{})

	if err := coord.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned %v", err)
	}

	inv := seedQueued(t, repo, ipRepo, "inv-after-shutdown")
	coord.Enqueue(inv, "test", "")

	got, err := repo.GetAcrossWorkspaces(context.Background(), "inv-after-shutdown")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("status = %v, want FAILED", got.GetStatus())
	}
	if !strings.Contains(got.GetError(), "shutdown") {
		t.Fatalf("error = %q, want shutdown reason", got.GetError())
	}
	if fr.calls != 0 {
		t.Fatalf("runner calls = %d, want 0", fr.calls)
	}
}

func TestReconcileStale_RecordsHonestReasonAndNeverReplays(t *testing.T) {
	repo := memory.New()
	repo.Save(context.Background(), &agentsv1.Invocation{
		Id:          "stale-running",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		WorkspaceId: "ws-1",
		StartedAt:   timestamppb.Now(),
	})

	n, err := ReconcileStale(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 stale, got %d", n)
	}

	got, _ := repo.GetAcrossWorkspaces(context.Background(), "stale-running")
	if got.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("status = %v, want FAILED", got.GetStatus())
	}
	if !strings.Contains(got.GetError(), "restart") || !strings.Contains(got.GetError(), "resubmit") {
		t.Fatalf("error = %q, want an honest restart reason with a resubmit hint", got.GetError())
	}
	if got.GetFinishedAt() == nil {
		t.Fatal("finished_at not set on reconciled record")
	}
	// ReconcileStale takes only the repository: reconciliation marks records
	// and, by construction, cannot re-invoke the Agent or repeat tool side
	// effects.
}
