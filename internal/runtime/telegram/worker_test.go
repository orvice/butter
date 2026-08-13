package telegram

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/telegramqueue"
)

type fakeWorkerQueue struct {
	mu       sync.Mutex
	touches  int
	acked    []string
	touchErr error
}

func (q *fakeWorkerQueue) Available() bool                   { return true }
func (q *fakeWorkerQueue) EnsureGroup(context.Context) error { return nil }
func (q *fakeWorkerQueue) ReadPending(context.Context, string, int64) ([]telegramqueue.Delivery, error) {
	return nil, nil
}
func (q *fakeWorkerQueue) Claim(context.Context, string, time.Duration, int64) ([]telegramqueue.Delivery, error) {
	return nil, nil
}
func (q *fakeWorkerQueue) Read(context.Context, string, int64, time.Duration) ([]telegramqueue.Delivery, error) {
	return nil, nil
}
func (q *fakeWorkerQueue) Touch(context.Context, string, string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.touches++
	return q.touchErr
}
func (q *fakeWorkerQueue) Ack(_ context.Context, ids ...string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.acked = append(q.acked, ids...)
	return nil
}
func (q *fakeWorkerQueue) snapshot() (int, []string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.touches, append([]string(nil), q.acked...)
}

type handlerFunc func(context.Context, *telegramqueue.Event) error

func (f handlerFunc) Handle(ctx context.Context, event *telegramqueue.Event) error {
	return f(ctx, event)
}

func TestWorkerProcessesBatchInParallel(t *testing.T) {
	queue := &fakeWorkerQueue{}
	started := make(chan int64, 2)
	release := make(chan struct{})
	worker := NewWorker(queue, handlerFunc(func(_ context.Context, event *telegramqueue.Event) error {
		started <- event.UpdateID
		<-release
		return nil
	}), "worker-a")

	done := make(chan struct{})
	go func() {
		worker.process(t.Context(), []telegramqueue.Delivery{
			{ID: "1-0", Event: &telegramqueue.Event{UpdateID: 1}},
			{ID: "2-0", Event: &telegramqueue.Event{UpdateID: 2}},
		})
		close(done)
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("batch entries did not start concurrently")
		}
	}
	close(release)
	<-done
	_, acked := queue.snapshot()
	if len(acked) != 2 {
		t.Fatalf("acked = %v, want both deliveries", acked)
	}
}

func TestWorkerHeartbeatsLongDelivery(t *testing.T) {
	queue := &fakeWorkerQueue{}
	release := make(chan struct{})
	worker := NewWorker(queue, handlerFunc(func(context.Context, *telegramqueue.Event) error {
		<-release
		return nil
	}), "worker-a")
	worker.heartbeatInterval = 5 * time.Millisecond

	done := make(chan struct{})
	go func() {
		worker.processOne(t.Context(), telegramqueue.Delivery{ID: "1-0", Event: &telegramqueue.Event{UpdateID: 1}})
		close(done)
	}()
	deadline := time.After(time.Second)
	for {
		touches, _ := queue.snapshot()
		if touches > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("worker never refreshed the pending entry")
		case <-time.After(time.Millisecond):
		}
	}
	close(release)
	<-done
	_, acked := queue.snapshot()
	if len(acked) != 1 {
		t.Fatalf("acked = %v, want the completed delivery", acked)
	}
}

func TestWorkerDoesNotAckAfterQueueLeaseLoss(t *testing.T) {
	queue := &fakeWorkerQueue{touchErr: errors.New("entry moved to another consumer")}
	worker := NewWorker(queue, handlerFunc(func(ctx context.Context, _ *telegramqueue.Event) error {
		<-ctx.Done()
		return nil
	}), "worker-a")
	worker.heartbeatInterval = 5 * time.Millisecond

	worker.processOne(t.Context(), telegramqueue.Delivery{ID: "1-0", Event: &telegramqueue.Event{UpdateID: 1}})
	_, acked := queue.snapshot()
	if len(acked) != 0 {
		t.Fatalf("acked = %v after lease loss", acked)
	}
}
