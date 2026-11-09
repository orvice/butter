package plane

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/eventqueue"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
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
	if publisher, ok := w.processor.publisher.(*DurablePublisher); ok {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				recovery, cancel := context.WithTimeout(run, 30*time.Second)
				_ = publisher.Recover(recovery)
				cancel()
				select {
				case <-run.Done():
					return
				case <-ticker.C:
				}
			}
		}()
	}
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
	var mu sync.Mutex
	active := map[string]bool{}
	dispatch := func(entries []eventqueue.Entry) {
		for _, entry := range entries {
			mu.Lock()
			if active[entry.ID] {
				mu.Unlock()
				continue
			}
			select {
			case slots <- struct{}{}:
				active[entry.ID] = true
			default:
				mu.Unlock()
				return
			}
			mu.Unlock()
			w.inflight.Add(1)
			go func() {
				defer func() {
					mu.Lock()
					delete(active, entry.ID)
					mu.Unlock()
					<-slots
					w.inflight.Done()
				}()
				w.processEntry(ctx, queue, entry)
			}()
		}
	}
	pendingCursor, reclaimCursor := "0", "0"
	for ctx.Err() == nil {
		// Retry our own deferred FIFO entries promptly; reclaim handles other
		// consumers only after their heartbeat has lapsed. Scan beyond the
		// active pool so a running head cannot hide another Run's pending work.
		if paged, ok := queue.(interface {
			ReadPendingAfter(context.Context, string, string, int64) ([]eventqueue.Entry, error)
		}); ok {
			if pending, err := paged.ReadPendingAfter(ctx, w.consumer, pendingCursor, int64(concurrency*2)); err == nil {
				pendingCursor = "0"
				if len(pending) > 0 {
					pendingCursor = pending[len(pending)-1].ID
				}
				dispatch(pending)
			}
		} else if pending, err := queue.ReadPending(ctx, w.consumer, int64(concurrency)); err == nil {
			dispatch(pending)
		}
		free := int64(cap(slots) - len(slots))
		if free == 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(25 * time.Millisecond):
				continue
			}
		}
		if paged, ok := queue.(interface {
			ClaimAfter(context.Context, string, time.Duration, string, int64) ([]eventqueue.Entry, string, error)
		}); ok {
			if pending, next, err := paged.ClaimAfter(ctx, w.consumer, w.reclaimIdle, reclaimCursor, free); err == nil {
				reclaimCursor = next
				dispatch(pending)
			}
		} else if pending, err := queue.Claim(ctx, w.consumer, w.reclaimIdle, free); err == nil {
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
	if ctx.Err() == nil {
		var err error
		if owned, ok := queue.(interface {
			AckOwned(context.Context, string, ...string) error
		}); ok {
			err = owned.AckOwned(ctx, w.consumer, entry.ID)
		} else if err = queue.Touch(ctx, w.consumer, entry.ID); err == nil {
			err = queue.Ack(ctx, entry.ID)
		}
		if err != nil {
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
	installationID := e.PlaneInstallationID
	if installationID == "" {
		installationID = e.InstallationID
	} // legacy Stream entries
	app, err := w.resolver.FindScopedApp(ctx, e.WorkspaceID, e.AppID, installationID)
	if err != nil {
		return err
	}
	// Admission owns only the Redis append. Rebuild this projection here,
	// before execution; failures remain pending instead of stalling the ACK.
	if err := w.processor.project(ctx, e, agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_RECEIVED, "", "", ""); err != nil {
		// This is a projection failure, never a terminal input decision. Do
		// not expose its cause to terminalPlaneEvent and ACK unexecuted work;
		// database errors can also echo record values into logs.
		return errors.New("plane receive projection unavailable")
	}
	if e.Stop {
		if !allowedStopUser(app.GetAllowedUserIds(), e.ActorUserID) {
			return ErrIdentity
		}
		return w.confirmStop(ctx, e)
	}
	log.FromContext(ctx).Info("plane event processing", planeEventLogFields(e)...)
	_, err = w.processor.Process(ctx, app, e)
	if err == nil {
		return nil
	}
	if errors.Is(err, errTurnInProgress) {
		return err
	}
	if errors.Is(err, ErrRunBusy) || errors.Is(err, planeprocessing.ErrClaimConflict) {
		return w.queued(ctx, e)
	}
	if errors.Is(err, ErrPromptDiscarded) {
		return w.reject(ctx, e, agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_CANCELLED, err.Error())
	}
	if errors.Is(err, ErrPromptRequired) || errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrPromptRejected) {
		return w.reject(ctx, e, agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED, err.Error())
	}
	if errors.Is(err, ErrTurnConsumed) && w.processor.batches != nil {
		// Re-read the persistent execution outcome. A generic consumed error
		// alone must never acknowledge an event whose projection is missing.
		var consumed *consumedTurn
		if !errors.As(err, &consumed) {
			err = w.processor.batches.prepare(ctx, e)
		}
		if errors.As(err, &consumed) {
			e.InvocationID = consumed.invocation
			e.Members = nil
			reason := ""
			if consumed.state == "uncertain" {
				reason = "execution started but its holder disappeared; not rerun"
				if w.processor.invocations != nil {
					if err := w.processor.invocations.FailActive(ctx, e.WorkspaceID, e.InvocationID, "plane", reason); err != nil {
						return err
					}
				}
			}
			if err := w.processor.project(ctx, e, processingState(consumed.state), reason, "", ""); err != nil {
				return err
			}
			return w.processor.recoverDeliveryProjection(ctx, e, consumed.state)
		}
		return fmt.Errorf("Plane execution outcome is not yet recoverable: %v", err)
	}
	return err
}

func (w *Worker) reject(ctx context.Context, e *Event, status agentsv1.PlaneProcessingStatus, reason string) error {
	e.Members = nil
	if w.processor.batches != nil {
		if err := w.processor.batches.retire(ctx, e); err != nil {
			return err
		}
	}
	return w.processor.project(ctx, e, status, reason, "", "")
}

func (w *Worker) confirmStop(ctx context.Context, e *Event) error {
	if w.processor.control == nil {
		return ErrRuntimeDependency
	}
	message := "Nothing is running, so there is nothing to stop."
	status := agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED
	reason := ""
	if e.AcceptedOrder == 0 {
		// Old queued controls have no admission ordering or holder token.
		// Replaying one must never stop a newer holder.
		message = "This Stop could not be confirmed. Send Stop again if the run is still active."
		status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED_UNCERTAIN
		reason = message
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
				status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED_UNCERTAIN
				reason = message
			} else if !done {
				select {
				case <-waitCtx.Done():
					return errors.Join(ErrStopPending, waitCtx.Err())
				case <-time.After(100 * time.Millisecond):
					continue
				}
			} else if state == RunStateStopped {
				message = "Stopped — send a message to continue."
				status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_CANCELLED
			} else {
				message = "The run completed before stopping was confirmed."
			}
			break
		}
	}
	err := w.processor.publisher.PublishReply(ctx, e, message)
	delivery := "delivered"
	if err != nil && !errors.Is(err, ErrDeliveryAlreadyDone) {
		delivery = "unconfirmed"
	}
	if saveErr := w.processor.project(ctx, e, status, reason, "", w.processor.deliveryStatus(ctx, e, delivery)); saveErr != nil {
		return saveErr
	}
	return err
}
