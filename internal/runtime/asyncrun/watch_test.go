package asyncrun

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/genai"

	inputpartmemory "go.orx.me/apps/butter/internal/repo/inputpart/memory"
	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestWatchHub_SlowObserverDroppedWithoutBlockingPublish(t *testing.T) {
	h := newWatchHub()
	slow, cancelSlow := h.subscribe("inv-1")
	defer cancelSlow()

	// Fill the slow observer's buffer and one more; publish must never block.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < observerBuffer+1; i++ {
			h.publish("inv-1", Frame{Kind: FrameTextDelta, Text: "x"})
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("publish blocked on a slow observer")
	}

	// The slow observer was dropped: after draining its buffered frames the
	// channel is closed without a terminal frame.
	var sawTerminal bool
	for f := range slow {
		if f.Terminal() {
			sawTerminal = true
		}
	}
	if sawTerminal {
		t.Fatal("lagged observer must not see a terminal frame")
	}
}

func TestWatchHub_HealthyObserverUnaffectedBySlowPeer(t *testing.T) {
	h := newWatchHub()
	slow, cancelSlow := h.subscribe("inv-1")
	defer cancelSlow()
	_ = slow // never read — lagged on purpose

	healthy, cancelHealthy := h.subscribe("inv-1")
	defer cancelHealthy()

	go func() {
		for i := 0; i < observerBuffer+1; i++ {
			h.publish("inv-1", Frame{Kind: FrameTextDelta, Text: "x"})
			// Keep the healthy observer drained by pacing under its buffer.
			if i%16 == 0 {
				time.Sleep(time.Millisecond)
			}
		}
		h.publish("inv-1", Frame{Kind: FrameState, Invocation: &agentsv1.Invocation{
			Id:     "inv-1",
			Status: agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		}})
		h.closeAll("inv-1")
	}()

	var deltas int
	var sawTerminal bool
	deadline := time.After(10 * time.Second)
	for {
		select {
		case f, ok := <-healthy:
			if !ok {
				if !sawTerminal {
					t.Fatalf("healthy observer closed without terminal after %d deltas", deltas)
				}
				if deltas == 0 {
					t.Fatal("healthy observer received no deltas")
				}
				return
			}
			switch f.Kind {
			case FrameTextDelta:
				deltas++
			case FrameState:
				if f.Terminal() {
					sawTerminal = true
				}
			}
		case <-deadline:
			t.Fatal("healthy observer starved")
		}
	}
}

// failingSaveRepo fails the first N Save calls, then delegates.
type failingSaveRepo struct {
	invocation.Repository
	failures int32
}

func (r *failingSaveRepo) Save(ctx context.Context, inv *agentsv1.Invocation) error {
	if atomic.AddInt32(&r.failures, -1) >= 0 {
		return errors.New("db down")
	}
	return r.Repository.Save(ctx, inv)
}

// noopRunner satisfies Runner; RunSSE must never be reached in the
// persist-failure test.
type noopRunner struct{ calls int32 }

func (r *noopRunner) RunSSE(context.Context, string, []*genai.Part, string, *agentsv1.ContextInfo, runner.EventCallback, runner.CompactionCallback) (string, error) {
	atomic.AddInt32(&r.calls, 1)
	return "", nil
}
func (r *noopRunner) ResolveAgentRef(_, id string) (string, bool)      { return id, id != "" }
func (r *noopRunner) GetAgentIdentity(n string) (string, string, bool) { return n, n, true }
func (r *noopRunner) CancelInvocation(string, string) bool             { return false }

func TestCoordinator_RunningPersistFailureStillEmitsTerminalFrame(t *testing.T) {
	repo := &failingSaveRepo{Repository: invocationmemory.New(), failures: 1}
	fake := &noopRunner{}
	c := New(repo, inputpartmemory.New(), fake, Config{})

	inv := &agentsv1.Invocation{
		Id:     "inv-persist-fail",
		Status: agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
	}
	frames, cancel := c.Watch(inv.GetId())
	defer cancel()

	c.Enqueue(inv, "agent", "")

	// The watcher must observe one FAILED terminal frame, then close —
	// close-without-terminal would misread as observer lag.
	var terminal *agentsv1.Invocation
	deadline := time.After(5 * time.Second)
	for terminal == nil {
		select {
		case f, ok := <-frames:
			if !ok {
				t.Fatal("channel closed without a terminal frame")
			}
			if f.Terminal() {
				terminal = f.Invocation
			}
		case <-deadline:
			t.Fatal("no terminal frame after RUNNING-persist failure")
		}
	}
	if terminal.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("terminal status = %v, want FAILED", terminal.GetStatus())
	}
	if atomic.LoadInt32(&fake.calls) != 0 {
		t.Fatal("runner must not run when the RUNNING transition never persisted")
	}
}

func TestWatchHub_CancelIsIdempotentAfterDrop(t *testing.T) {
	h := newWatchHub()
	ch, cancel := h.subscribe("inv-1")

	// Force the hub to drop the observer (buffer overflow).
	for i := 0; i < observerBuffer+1; i++ {
		h.publish("inv-1", Frame{Kind: FrameTextDelta})
	}
	for range ch {
	}

	// cancel after the hub already closed the channel must not panic.
	cancel()
	cancel()
}
