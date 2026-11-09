package plane

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/eventqueue"
	processingrepo "go.orx.me/apps/butter/internal/repo/planeprocessing"
	processingmem "go.orx.me/apps/butter/internal/repo/planeprocessing/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type unavailableProjection struct {
	processingrepo.Repository
	unavailable atomic.Bool
	cause       error
	entered     chan struct{}
	once        sync.Once
}

func (s *unavailableProjection) Project(ctx context.Context, row *agentsv1.PlaneProcessingRecord) error {
	unavailable := s.unavailable.Load()
	s.once.Do(func() { close(s.entered) })
	if unavailable {
		if s.cause != nil {
			return s.cause
		}
		return errors.New("projection unavailable")
	}
	return s.Repository.Project(ctx, row)
}

func TestProjectionFailureCannotDropAcceptedWorkAsTerminalInput(t *testing.T) {
	f := newReceiverFixture(t)
	if _, err := f.deliver(t, officialEnvelope(t)); err != nil {
		t.Fatal(err)
	}
	projection := &unavailableProjection{Repository: processingmem.New(), entered: make(chan struct{}), cause: errors.Join(ErrMalformed, errors.New("private record values"))}
	projection.unavailable.Store(true)
	model := &fakeRunner{out: "answer"}
	p := NewProcessor(model, &fakePublisher{})
	p.SetProcessingRepository(projection)
	worker := &Worker{processor: p, resolver: &planeTestResolver{app: f.app}}
	payload, err := f.queue.events[0].Encode()
	if err != nil {
		t.Fatal(err)
	}
	err = worker.handle(t.Context(), eventqueue.Entry{ID: "1-0", Payload: payload})
	if err == nil || terminalPlaneEvent(err) || model.calls != 0 || strings.Contains(err.Error(), "private") {
		t.Fatalf("projection failure turned into a terminal ACK: %v", err)
	}
}

func TestRedisAdmissionACKDoesNotWaitForProjectionAndWorkerRetries(t *testing.T) {
	q, _ := planeQueueFixture(t)
	f := newReceiverFixture(t)
	f.receiver.queue = q
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		keys, _ := q.client.Keys(ctx, controlPrefix+"ws:app:"+f.installation.GetId()+":*").Result()
		seen, _ := q.client.Keys(ctx, "butter:plane:seen:"+f.installation.GetId()+":*").Result()
		keys = append(keys, seen...)
		if len(keys) > 0 {
			_ = q.client.Del(ctx, keys...).Err()
		}
	})
	records := processingmem.New()
	projection := &unavailableProjection{Repository: records, entered: make(chan struct{})}
	projection.unavailable.Store(true)
	model := &fakeRunner{out: "answer"}
	p := batchProcessor(q, model, records, "projection-worker")
	p.SetProcessingRepository(projection)
	worker := NewWorker(q, p, &planeTestResolver{app: f.app}, "projection-worker")
	if err := worker.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(worker.Stop)
	if decision, err := f.deliver(t, officialEnvelope(t)); err != nil || decision != DecisionAccepted {
		t.Fatalf("durable ACK waited on projection: %s %v", decision, err)
	}
	awaitPlane(t, projection.entered, "worker projection attempt")
	pending, err := q.client.XPending(t.Context(), StreamKey, ConsumerGroup).Result()
	if err != nil || pending.Count != 1 {
		t.Fatalf("projection failure lost pending work: %+v %v", pending, err)
	}
	model.mu.Lock()
	calls := model.calls
	model.mu.Unlock()
	if calls != 0 {
		t.Fatal("model ran before its processing projection was persisted")
	}
	projection.unavailable.Store(false)
	waitBatch(t, func() bool {
		rows, _ := records.List(t.Context(), processingrepo.Filter{WorkspaceID: "ws", AppID: "app"})
		pending, err := q.client.XPending(t.Context(), StreamKey, ConsumerGroup).Result()
		return err == nil && pending.Count == 0 && len(rows) == 1 && rows[0].GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED && rows[0].GetReplyDeliveryStatus() == "delivered"
	}, "worker did not recover projection and finish the accepted turn")
	model.mu.Lock()
	defer model.mu.Unlock()
	if model.calls != 1 {
		t.Fatalf("projection retry reran the model %d times", model.calls)
	}
}
