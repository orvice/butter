package plane

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/eventqueue"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type PlaneAppResolver interface {
	FindScopedApp(context.Context, string, string, string) (*agentsv1.PlaneApp, error)
}
type planeWorkerQueue interface {
	EnsureGroup(context.Context) error
	Read(context.Context, string, int64, time.Duration) ([]eventqueue.Entry, error)
	ReadPending(context.Context, string, int64) ([]eventqueue.Entry, error)
	Claim(context.Context, string, time.Duration, int64) ([]eventqueue.Entry, error)
	Touch(context.Context, string, string) error
	Ack(context.Context, ...string) error
}

// Worker has separate bounded pools for model work and Stop confirmations.
// Cancellation itself was accepted atomically by the webhook; even a full
// control pool cannot delay the holder learning its Stop marker.
type Worker struct {
	queue, controls               planeWorkerQueue
	processor                     *Processor
	resolver                      PlaneAppResolver
	consumer                      string
	concurrency                   int
	block, reclaimIdle, heartbeat time.Duration
	stop                          context.CancelFunc
	wg, inflight                  sync.WaitGroup
}

func NewWorker(q *StreamQueue, p *Processor, r PlaneAppResolver, consumer string) *Worker {
	w := &Worker{processor: p, resolver: r, consumer: consumer, concurrency: 32,
		block: time.Second, reclaimIdle: 2 * time.Minute, heartbeat: 30 * time.Second}
	if q != nil {
		w.queue, w.controls = q.inner, q.controls
	}
	return w
}
func (w *Worker) Start(ctx context.Context) error {
	if w == nil || w.queue == nil || w.controls == nil || w.processor == nil || w.resolver == nil {
		return errors.New("plane worker dependencies are not configured")
	}
	if err := w.queue.EnsureGroup(ctx); err != nil {
		return err
	}
	if err := w.controls.EnsureGroup(ctx); err != nil {
		return err
	}
	run, stop := context.WithCancel(ctx)
	w.stop = stop
	w.wg.Add(2)
	go func() { defer w.wg.Done(); w.loop(run, w.queue, max(w.concurrency, 1)) }()
	go func() { defer w.wg.Done(); w.loop(run, w.controls, 4) }()
	return nil
}
func (w *Worker) Stop() {
	if w.stop != nil {
		w.stop()
		w.wg.Wait()
		w.inflight.Wait()
	}
}
func (w *Worker) loop(ctx context.Context, queue planeWorkerQueue, concurrency int) {
	slots := make(chan struct{}, concurrency)
	dispatch := func(entries []eventqueue.Entry) {
		for _, entry := range entries {
			slots <- struct{}{}
			w.inflight.Add(1)
			go func() { defer func() { <-slots; w.inflight.Done() }(); w.processEntry(ctx, queue, entry) }()
		}
	}
	if pending, err := queue.ReadPending(ctx, w.consumer, int64(concurrency)); err == nil {
		dispatch(pending)
	}
	for ctx.Err() == nil {
		free := int64(cap(slots) - len(slots))
		if free == 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(25 * time.Millisecond):
				continue
			}
		}
		if pending, err := queue.Claim(ctx, w.consumer, w.reclaimIdle, free); err == nil {
			dispatch(pending)
		}
		free = int64(cap(slots) - len(slots))
		if free == 0 {
			continue
		}
		entries, err := queue.Read(ctx, w.consumer, free, w.block)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}
		dispatch(entries)
	}
}
func (w *Worker) processEntry(ctx context.Context, queue planeWorkerQueue, entry eventqueue.Entry) {
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	beatDone := make(chan struct{})
	go func() {
		defer close(beatDone)
		tick := time.NewTicker(w.heartbeat)
		defer tick.Stop()
		for {
			select {
			case <-workCtx.Done():
				return
			case <-tick.C:
				if err := queue.Touch(workCtx, w.consumer, entry.ID); err != nil {
					cancel()
					return
				}
			}
		}
	}()
	err := w.handle(workCtx, entry)
	cancel()
	<-beatDone
	if err != nil && !terminalPlaneEvent(err) {
		log.FromContext(ctx).Warn("plane event deferred", "stream_id", entry.ID, "err", err)
		return
	}
	// XACK alone is not owner-aware. Check ownership after execution too.
	if ctx.Err() == nil && queue.Touch(ctx, w.consumer, entry.ID) == nil {
		if err := queue.Ack(ctx, entry.ID); err != nil {
			log.FromContext(ctx).Warn("plane event acknowledgment failed", "stream_id", entry.ID, "err", err)
		}
	}
}
func terminalPlaneEvent(err error) bool {
	return errors.Is(err, ErrTurnConsumed) || errors.Is(err, ErrPromptDiscarded) ||
		errors.Is(err, ErrDuplicate) || errors.Is(err, ErrPromptRequired) ||
		errors.Is(err, ErrIdentity) || errors.Is(err, ErrMalformed) ||
		errors.Is(err, ErrDeliveryAlreadyDone) || errors.Is(err, ErrPermanentDelivery)
}
func (w *Worker) handle(ctx context.Context, entry eventqueue.Entry) (result error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = fmt.Errorf("plane event handler panicked: %v", recovered)
		}
	}()
	e, err := DecodeEvent(entry.Payload)
	if err != nil {
		return fmt.Errorf("%w: invalid queue payload", ErrMalformed)
	}
	if e.WorkspaceID == "" || e.AppID == "" || e.InstallationID == "" {
		return ErrIdentity
	}
	app, err := w.resolver.FindScopedApp(ctx, e.WorkspaceID, e.AppID, e.InstallationID)
	if err != nil {
		return err
	}
	if e.Stop {
		if !allowedStopUser(app.GetAllowedUserIds(), e.ActorUserID) {
			return ErrIdentity
		}
		err := w.confirmStop(ctx, e)
		w.updateStatus(ctx, e, err, true)
		return err
	}
	w.updateStatus(ctx, e, nil, false)
	_, err = w.processor.Process(ctx, app, e)
	w.updateStatus(ctx, e, err, false)
	return err
}

func (w *Worker) updateStatus(ctx context.Context, e *Event, err error, stop bool) {
	if w.processor == nil || w.processor.processing == nil {
		return
	}
	status := agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED
	failure := ""
	if !stop {
		status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_PROCESSING
		if err != nil {
			if errors.Is(err, ErrPromptDiscarded) {
				status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_CANCELLED
			} else {
				status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED
			}
			failure = err.Error()
		}
	} else if err != nil {
		status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED
		failure = err.Error()
	} else if e.StopTarget != "" {
		status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_CANCELLED
	}
	_ = w.processor.processing.UpdateStatus(ctx, e.WorkspaceID, e.AppID, e.InstallationID, e.AgentRunID, e.ActivityID, e.Action, status, failure)
}
func (w *Worker) confirmStop(ctx context.Context, e *Event) error {
	if w.processor.control == nil {
		return ErrRuntimeDependency
	}
	message := "Nothing is running, so there is nothing to stop."
	if e.AcceptedOrder == 0 {
		// Old queued controls have no admission ordering or holder token.
		// Replaying one must never stop a newer holder.
		message = "This Stop could not be confirmed. Send Stop again if the run is still active."
	} else if e.StopTarget != "" {
		waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		for {
			state, done, err := w.processor.control.Outcome(ctx, e, e.StopTarget)
			if err != nil {
				if !errors.Is(err, ErrRunLeaseLost) {
					return err
				}
				message = "The run lost its holder; stopping could not be confirmed."
			} else if !done {
				select {
				case <-waitCtx.Done():
					return errors.Join(ErrStopPending, waitCtx.Err())
				case <-time.After(100 * time.Millisecond):
					continue
				}
			} else if state == RunStateStopped {
				message = "Stopped — send a message to continue."
			} else {
				message = "The run completed before stopping was confirmed."
			}
			break
		}
	}
	return w.processor.publisher.PublishReply(ctx, e, message)
}
