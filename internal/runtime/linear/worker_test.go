package linear

import (
	"context"
	"sync"
	"testing"
	"time"
)

// scriptedQueue hands out deliveries pushed by the test, one read at a time,
// and records acknowledgements.
type scriptedQueue struct {
	incoming chan Delivery
	mu       sync.Mutex
	acked    []string
}

func (q *scriptedQueue) EnsureGroup(context.Context) error { return nil }
func (q *scriptedQueue) ReadPending(context.Context, string, int64) ([]Delivery, error) {
	return nil, nil
}
func (q *scriptedQueue) Claim(context.Context, string, time.Duration, int64) ([]Delivery, error) {
	return nil, nil
}
func (q *scriptedQueue) Touch(context.Context, string, string) error { return nil }

func (q *scriptedQueue) Read(ctx context.Context, _ string, _ int64, block time.Duration) ([]Delivery, error) {
	select {
	case d := <-q.incoming:
		return []Delivery{d}, nil
	case <-time.After(block):
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (q *scriptedQueue) Ack(_ context.Context, ids ...string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.acked = append(q.acked, ids...)
	return nil
}

func (q *scriptedQueue) ackedIDs() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]string(nil), q.acked...)
}

// gatedHandler blocks events for session "slow" until released.
type gatedHandler struct {
	release chan struct{}
	handled chan string
}

func (h *gatedHandler) Handle(ctx context.Context, event *Event) error {
	if event.AgentSessionID == "slow" {
		select {
		case <-h.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	h.handled <- event.AgentSessionID
	return nil
}

func TestABlockedTurnDoesNotDelayAnotherSessionsEvent(t *testing.T) {
	queue := &scriptedQueue{incoming: make(chan Delivery, 2)}
	handler := &gatedHandler{release: make(chan struct{}), handled: make(chan string, 2)}
	worker := NewWorker(queue, handler, "pod-a", 4)
	worker.block = 20 * time.Millisecond
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if err := worker.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	queue.incoming <- Delivery{ID: "1-0", Event: &Event{AgentSessionID: "slow"}}
	queue.incoming <- Delivery{ID: "2-0", Event: &Event{AgentSessionID: "fast"}}

	select {
	case got := <-handler.handled:
		if got != "fast" {
			t.Fatalf("first handled = %q, want the fast session while the slow one is blocked", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the second session's event waited behind the blocked turn")
	}
	close(handler.release)
	if got := <-handler.handled; got != "slow" {
		t.Fatalf("second handled = %q", got)
	}
	cancel()
	worker.Stop()
	if acked := queue.ackedIDs(); len(acked) != 2 {
		t.Fatalf("acked = %v, want both entries", acked)
	}
}

func TestAWorkerClaimsNoMoreThanItsFreeSlots(t *testing.T) {
	queue := &scriptedQueue{incoming: make(chan Delivery, 3)}
	handler := &gatedHandler{release: make(chan struct{}), handled: make(chan string, 3)}
	worker := NewWorker(queue, handler, "pod-a", 1)
	worker.block = 20 * time.Millisecond
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if err := worker.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	queue.incoming <- Delivery{ID: "1-0", Event: &Event{AgentSessionID: "slow"}}
	queue.incoming <- Delivery{ID: "2-0", Event: &Event{AgentSessionID: "fast"}}

	// With its one slot busy, the worker leaves the second entry in the
	// Stream for other Pods instead of claiming it.
	time.Sleep(150 * time.Millisecond)
	if n := len(queue.incoming); n != 1 {
		t.Fatalf("entries left in the stream = %d, want 1", n)
	}
	close(handler.release)
	for range 2 {
		select {
		case <-handler.handled:
		case <-time.After(2 * time.Second):
			t.Fatal("events were not handled after the slot freed")
		}
	}
	cancel()
	worker.Stop()
}
