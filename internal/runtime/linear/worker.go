package linear

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"butterfly.orx.me/core/log"
)

const (
	// DefaultConcurrency bounds how many events one Pod handles at once.
	DefaultConcurrency = 32
	// workerBlock is how long a read waits for work before looping.
	workerBlock = 5 * time.Second
	// reclaimIdle is how long an entry may sit unacknowledged before another
	// Pod takes it over — the recovery path for a crashed worker.
	reclaimIdle = 2 * time.Minute
)

// EventHandler processes one claimed event. Returning nil acknowledges it.
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
}

type workerQueue interface {
	EnsureGroup(ctx context.Context) error
	ReadPending(ctx context.Context, consumer string, count int64) ([]Delivery, error)
	Claim(ctx context.Context, consumer string, minIdle time.Duration, count int64) ([]Delivery, error)
	Read(ctx context.Context, consumer string, count int64, block time.Duration) ([]Delivery, error)
	Touch(ctx context.Context, consumer, id string) error
	Ack(ctx context.Context, ids ...string) error
}

// Worker drains the Linear Stream on one Pod.
//
// Unlike the Telegram worker it never waits for claimed work to finish
// before reading again: an Agent turn can take half an hour, and a created
// session queued behind it must still be acknowledged within Linear's
// 10-second deadline. Each event runs in its own goroutine; reads claim only
// as many entries as there are free slots, so a busy Pod leaves the rest of
// the Stream to other Pods.
type Worker struct {
	queue             workerQueue
	handler           EventHandler
	consumer          string
	slots             chan struct{}
	heartbeatInterval time.Duration
	reclaimIdle       time.Duration
	block             time.Duration

	stop     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
	inflight sync.WaitGroup
}

func NewWorker(queue workerQueue, handler EventHandler, consumer string, concurrency int) *Worker {
	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}
	return &Worker{
		queue:             queue,
		handler:           handler,
		consumer:          consumer,
		slots:             make(chan struct{}, concurrency),
		heartbeatInterval: reclaimIdle / 3,
		reclaimIdle:       reclaimIdle,
		block:             workerBlock,
		stop:              make(chan struct{}),
		stopped:           make(chan struct{}),
	}
}

// Start begins consuming. It first resumes this consumer's own pending
// entries — the work it held when the process last died.
func (w *Worker) Start(ctx context.Context) error {
	if w.queue == nil {
		return errors.New("linear worker requires a queue")
	}
	if err := w.queue.EnsureGroup(ctx); err != nil {
		return err
	}
	go w.run(ctx)
	return nil
}

// Stop halts reading and waits for in-flight events.
func (w *Worker) Stop() {
	w.stopOnce.Do(func() { close(w.stop) })
	<-w.stopped
	w.inflight.Wait()
}

func (w *Worker) free() int64 { return int64(cap(w.slots) - len(w.slots)) }

func (w *Worker) stopping(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	case <-w.stop:
		return true
	default:
		return false
	}
}

func (w *Worker) run(ctx context.Context) {
	defer close(w.stopped)
	logger := log.FromContext(ctx)

	if pending, err := w.queue.ReadPending(ctx, w.consumer, int64(cap(w.slots))); err != nil {
		logger.Warn("could not recover pending linear work", "err", err)
	} else {
		w.dispatch(ctx, pending)
	}

	for !w.stopping(ctx) {
		if w.free() == 0 {
			// Every slot is busy: wait for one rather than claiming work
			// this Pod cannot start.
			select {
			case <-ctx.Done():
				return
			case <-w.stop:
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}
		// Take over anything a dead Pod abandoned before claiming new work.
		if reclaimed, err := w.queue.Claim(ctx, w.consumer, w.reclaimIdle, w.free()); err != nil {
			logger.Debug("linear reclaim failed", "err", err)
		} else {
			w.dispatch(ctx, reclaimed)
		}
		if w.free() == 0 {
			continue
		}
		deliveries, err := w.queue.Read(ctx, w.consumer, w.free(), w.block)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Warn("linear worker read failed", "err", err)
			select {
			case <-ctx.Done():
				return
			case <-w.stop:
				return
			case <-time.After(time.Second):
			}
			continue
		}
		w.dispatch(ctx, deliveries)
	}
}

// dispatch starts each delivery in its own slot.
func (w *Worker) dispatch(ctx context.Context, deliveries []Delivery) {
	for _, delivery := range deliveries {
		w.slots <- struct{}{}
		w.inflight.Add(1)
		go func() {
			defer func() {
				<-w.slots
				w.inflight.Done()
			}()
			w.processOne(ctx, delivery)
		}()
	}
}

func (w *Worker) processOne(ctx context.Context, delivery Delivery) {
	logger := log.FromContext(ctx).With("stream_id", delivery.ID,
		"app_id", delivery.Event.AppID, "delivery_id", delivery.Event.DeliveryID)
	handleCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	var leaseLost atomic.Bool
	go func() {
		defer close(done)
		ticker := time.NewTicker(w.heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-handleCtx.Done():
				return
			case <-ticker.C:
				if err := w.queue.Touch(handleCtx, w.consumer, delivery.ID); err != nil {
					logger.Error("linear queue delivery lease lost", "err", err)
					leaseLost.Store(true)
					cancel()
					return
				}
			}
		}
	}()

	err := w.handle(handleCtx, delivery.Event)
	cancel()
	<-done
	if err != nil {
		// Leave it unacknowledged: it is reclaimed after reclaimIdle.
		if errors.Is(err, ErrSessionBusy) {
			logger.Debug("linear event deferred", "err", err)
		} else {
			logger.Error("linear event handling failed", "err", err)
		}
		return
	}
	if leaseLost.Load() {
		// Never acknowledge work another consumer may be running.
		return
	}
	if err := w.queue.Ack(ctx, delivery.ID); err != nil {
		logger.Warn("could not acknowledge linear event", "err", err)
	}
}

// handle runs the handler, turning a panic into an error so one bad event
// cannot take the Pod down.
func (w *Worker) handle(ctx context.Context, event *Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("linear handler panic: %v", r)
		}
	}()
	return w.handler.Handle(ctx, event)
}
