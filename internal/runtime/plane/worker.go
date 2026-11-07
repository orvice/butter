package plane

import (
	"context"
	"errors"
	"go.orx.me/apps/butter/internal/eventqueue"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"sync"
	"time"
)

type PlaneAppResolver interface {
	FindApp(context.Context, string) (*agentsv1.PlaneApp, error)
}
type Worker struct {
	queue     *StreamQueue
	processor *Processor
	resolver  PlaneAppResolver
	consumer  string
	batch     int
	block     time.Duration
	stop      context.CancelFunc
	wg        sync.WaitGroup
}

func NewWorker(q *StreamQueue, p *Processor, r PlaneAppResolver, consumer string) *Worker {
	return &Worker{queue: q, processor: p, resolver: r, consumer: consumer, batch: 10, block: time.Second}
}
func (w *Worker) Start(ctx context.Context) error {
	if w == nil || w.queue == nil || !w.queue.Available() {
		return errors.New("plane worker queue is not configured")
	}
	if err := w.queue.EnsureGroup(ctx); err != nil {
		return err
	}
	run, stop := context.WithCancel(ctx)
	w.stop = stop
	w.wg.Add(1)
	go func() { defer w.wg.Done(); w.loop(run) }()
	return nil
}
func (w *Worker) Stop() {
	if w.stop != nil {
		w.stop()
		w.wg.Wait()
	}
}
func (w *Worker) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		entries, err := w.queue.Read(ctx, w.consumer, int64(w.batch), w.block)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		for _, entry := range entries {
			_ = w.handle(ctx, entry)
		}
	}
}
func (w *Worker) handle(ctx context.Context, entry eventqueue.Entry) error {
	e, err := DecodeEvent(entry.Payload)
	if err != nil {
		return err
	}
	app, err := w.resolver.FindApp(ctx, e.AppID)
	if err != nil {
		return err
	}
	_, runErr := w.processor.Process(ctx, app, e)
	if ackErr := w.queue.Ack(ctx, entry.ID); ackErr != nil {
		return ackErr
	}
	return runErr
}
