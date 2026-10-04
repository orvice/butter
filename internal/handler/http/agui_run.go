package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"butterfly.orx.me/core/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/aguitool"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/runstate"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	"go.orx.me/apps/butter/internal/runtime/sessionshare"
	"go.orx.me/apps/butter/internal/runtime/streamorch"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// A run, from the moment it holds its thread to its terminal state
// (ADR-0016). Every run records its run state next to the thread lease. A run
// whose client opted in with forwardedProps.butterRun = {"detach": true} is a
// Detached Run: it holds the lease itself, on a context the request's end
// does not reach, owns its Invocation record, and runs in its own goroutine
// while the response only observes it. Any other run lives in its request, as
// it always has.

// aguiRunExtensionKey is the forwardedProps key of the Detached Run
// extension, next to butterA2UI.
const aguiRunExtensionKey = "butterRun"

// AGUIRunStateKeyPrefix namespaces the run states recorded next to the AG-UI
// session leases, keyed like them by caller and thread.
const AGUIRunStateKeyPrefix = "butter:agui:run:session:"

// AGUIDefaultMaxRunDuration bounds a Detached Run when no maximum is
// configured.
const AGUIDefaultMaxRunDuration = 30 * time.Minute

// aguiPostRunTimeout bounds the work a Detached Run does after its turn: the
// final session read behind the trailing STATE_DELTA, the answered forms and
// the interrupt outcome.
const aguiPostRunTimeout = 15 * time.Second

// aguiRecordTimeout bounds one write of a Detached Run's Invocation record.
// The writes run on contexts the run's end does not reach.
const aguiRecordTimeout = 10 * time.Second

// aguiRunStateEndTimeout bounds removing a run's state, as releasing its
// lease is bounded. A state that is not removed lapses on its own.
const aguiRunStateEndTimeout = 5 * time.Second

// aguiRecordTextLimit caps the input and output an Invocation record keeps,
// as the runner caps its own.
const aguiRecordTextLimit = 4096

// aguiCodeStopped is the RUN_ERROR code of a run a person stopped (ADR-0016
// decision 4): a client tells a Stop apart from a failure by it.
const aguiCodeStopped = "stopped"

// aguiRunError is how a run ended, as its RUN_ERROR reports it: the message,
// and the code a client tells the end apart by, when it has one.
type aguiRunError struct {
	message string
	code    string
}

func (e *aguiRunError) Error() string { return e.message }

var (
	// errAGUIRunStopped ends a Detached Run a person stopped, through the
	// Stop endpoint, CancelAgentInvocation or deleting its thread.
	errAGUIRunStopped error = &aguiRunError{message: "stopped by user", code: aguiCodeStopped}
	// errAGUIShutdown ends the Detached Runs in flight at a graceful
	// shutdown.
	errAGUIShutdown = errors.New("server shutting down")
	// errAGUIRunTimedOut ends a Detached Run past the maximum run duration.
	errAGUIRunTimedOut = errors.New("maximum run duration exceeded")
)

// aguiLeaseLostMessage names a run cancelled because it lost its thread's
// lease, instead of a bare context cancellation.
const aguiLeaseLostMessage = "session lease lost, the run was cancelled"

// aguiShutdownReason is the error a Detached Run that a graceful shutdown
// interrupted records and reports.
const aguiShutdownReason = "interrupted by a service shutdown" + invocation.NoReplaySuffix

func aguiDeadlineReason(maxRun time.Duration) string {
	return "exceeded the maximum run duration (" + maxRun.String() + ") and was stopped" + invocation.NoReplaySuffix
}

// Codes of the Detached Run refusals.
const (
	// aguiCodeRunExists: the thread already ran this runId.
	aguiCodeRunExists = "run_exists"
	// aguiCodeThreadUnbound: the thread has no UI Binding, so its runs
	// cannot detach.
	aguiCodeThreadUnbound = "thread_unbound"
)

// negotiateDetach reads the Detached Run extension:
// forwardedProps.butterRun = {"detach": true} asks for a run that outlives
// its request (ADR-0016 decision 1). Without the declaration, or with
// {"detach": false}, the run lives in its request as before. A declaration
// of any other shape is an error, so a client never believes its run
// detached when it did not.
func negotiateDetach(forwardedProps any) (bool, error) {
	props, isMap := forwardedProps.(map[string]any)
	if !isMap {
		return false, nil
	}
	raw, present := props[aguiRunExtensionKey]
	if !present || raw == nil {
		return false, nil
	}
	decl, isMap := raw.(map[string]any)
	if !isMap {
		return false, errors.New("forwardedProps.butterRun must be an object")
	}
	var unknown []string
	for key := range decl {
		if key != "detach" {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		slices.Sort(unknown)
		return false, fmt.Errorf("forwardedProps.butterRun.%s is not supported", unknown[0])
	}
	detach, isBool := decl["detach"].(bool)
	if !isBool {
		return false, errors.New("forwardedProps.butterRun.detach must be a boolean")
	}
	return detach, nil
}

// aguiRequestID is the request_id of a Detached Run's Invocation record: the
// client's runId, scoped to the thread, because the client chooses it.
func aguiRequestID(ctxInfo *agentsv1.ContextInfo, runID string) string {
	return "agui:" + aguiSessionKey(ctxInfo) + ":" + runID
}

// aguiRun is one run from the moment it holds its thread to its terminal
// state.
type aguiRun struct {
	h  *AGUIHandler
	rc *aguiRunContext
	// thread is the thread's key, under which its lease and run state live.
	thread    string
	messageID string

	// ctx is the lease context: cancelled when the lease is lost and, for a
	// Detached Run, by a Stop, a shutdown or the maximum run duration.
	ctx context.Context
	// reqCtx is the request's context.
	reqCtx  context.Context
	release func()
	// leaseToken names the run's hold on its thread: the token of its lease
	// acquisition, which a Stop's marker holds.
	leaseToken string

	ui   *aguiUIContext
	sink *aguiSink
	// runState is the run state kept next to the lease; nil without a store.
	runState *runstate.Kept

	// detached is what only a Detached Run has; nil for a run in its request.
	detached *aguiDetachedRun
}

// aguiDetachedRun is what a Detached Run has on top of every run: the bounds
// of its context, its place among the runs in flight, its Invocation record
// and its observers.
type aguiDetachedRun struct {
	maxRun time.Duration
	entry  *aguiRunEntry
	// cancel ends the run's context, which the request's end does not reach.
	cancel func()
	inv    *agentsv1.Invocation
	fanout *aguiFanout
}

// beginRun takes the thread's lease and, holding it, runs every check that
// reads the session and records the run's start: its run state and, for a
// Detached Run, its QUEUED Invocation record. A refusal is an HTTP error
// written here, before the stream opens and before anything is detached; ok
// false means nothing runs and the lease is free again.
func (h *AGUIHandler) beginRun(c *gin.Context, rc *aguiRunContext) (*aguiRun, bool) {
	run := &aguiRun{
		h:         h,
		rc:        rc,
		thread:    aguiSessionKey(rc.ctxInfo),
		messageID: uuid.NewString(),
		reqCtx:    c.Request.Context(),
	}
	parent := c.Request.Context()
	if rc.detach {
		// The run, not the request, holds the lease: it is taken on a context
		// the request's end does not reach, bounded by the maximum run
		// duration (ADR-0016 decision 2).
		d := &aguiDetachedRun{maxRun: h.maxRunDuration(), fanout: newAGUIFanout()}
		ctx, cancel := context.WithCancelCause(context.WithoutCancel(parent))
		ctx, cancelTimeout := context.WithTimeoutCause(ctx, d.maxRun, errAGUIRunTimedOut)
		d.cancel = func() {
			cancelTimeout()
			cancel(nil)
		}
		entry, ok := h.runs.register(rc.ctxInfo.GetUuid(), cancel)
		if !ok {
			d.cancel()
			c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "the server is shutting down, retry the run"})
			return nil, false
		}
		d.entry = entry
		run.detached = d
		parent = ctx
	}
	leaseCtx, release, ok := h.acquireThread(parent, c, rc.ctxInfo)
	if !ok {
		run.unregister()
		return nil, false
	}
	run.ctx, run.release = leaseCtx, release
	run.leaseToken = aguiLeaseToken(leaseCtx, rc.ctxInfo.GetUuid())
	// Until the stream opens the work is the request's: it ends with the
	// request even when the run would not.
	checkCtx, stopChecks := context.WithCancel(leaseCtx)
	defer stopChecks()
	defer context.AfterFunc(c.Request.Context(), stopChecks)()
	if refusal := run.prepare(checkCtx); refusal != nil {
		run.endRunState()
		run.release()
		run.unregister()
		refusal.write(c)
		return nil, false
	}
	return run, true
}

// prepare runs, under the lease, every check that reads the session — the
// thread checks, client tool results against the pending calls, the
// shared-state baseline and form submissions — then records the run's start.
// Nothing has run, and no record exists, when it refuses.
func (r *aguiRun) prepare(ctx context.Context) *aguiRefusal {
	rc := r.rc
	ui, refusal := r.h.prepareThread(ctx, rc)
	if refusal != nil {
		return refusal
	}
	r.ui = ui
	if r.detached != nil && !ui.bound {
		// Stop and attach are checked against the thread's UI Binding, so a
		// run on a thread without one could be neither stopped nor followed.
		return &aguiRefusal{status: http.StatusBadRequest, body: aguiCodedError{
			Error: "this thread cannot detach its runs: it has no UI binding; start a new thread, or run without forwardedProps.butterRun",
			Code:  aguiCodeThreadUnbound,
		}}
	}
	if r.h.getSessionService() != nil {
		rc.prepareSharedState(ui.sess)
	}
	if refusal := resolveFormSubmissions(rc, ui); refusal != nil {
		return refusal
	}
	if r.detached != nil {
		if refusal := r.checkRunID(ctx); refusal != nil {
			return refusal
		}
	}
	if err := r.keepRunState(aguiEventCount(ui.sess)); err != nil {
		log.FromContext(r.ctx).Error("agui run state unavailable", "session_id", rc.ctxInfo.GetSessionId(), "err", err)
		return refuseAGUI(http.StatusServiceUnavailable, errors.New("run state unavailable, retry later"))
	}
	if r.detached != nil {
		if err := r.recordQueued(); err != nil {
			log.FromContext(r.ctx).Error("agui invocation record unavailable", "session_id", rc.ctxInfo.GetSessionId(), "err", err)
			return refuseAGUI(http.StatusServiceUnavailable, errors.New("invocation store unavailable, retry later"))
		}
	}
	return nil
}

// prepareThread runs, under the thread's lease, the checks that read the
// thread's session, and creates the session of a thread that has none yet,
// carrying this caller's binding:
//   - a threadId another user holds, or one whose session lives in another
//     workspace or is bound to another agent, is refused (403);
//   - client tool results must answer calls the session is waiting on (400),
//     and their parts are added to the run's.
//
// A refused request creates nothing. A session created before A2UI existed
// has no binding: it runs with any agent, as text chat with no UI.
func (h *AGUIHandler) prepareThread(ctx context.Context, rc *aguiRunContext) (*aguiUIContext, *aguiRefusal) {
	svc := h.getSessionService()
	if svc == nil {
		if len(rc.formEntries) > 0 {
			return nil, refuseAGUI(http.StatusServiceUnavailable, errors.New("form submissions are not accepted: session service unavailable"))
		}
		return &aguiUIContext{}, nil
	}
	binding := aguiBinding(ctx, rc.ctxInfo.GetWorkspaceId(), rc.agent.GetAgentId(), rc.input.ThreadID)
	get := &session.GetRequest{
		AppName:   aguiAppName,
		UserID:    rc.ctxInfo.GetUserId(),
		SessionID: rc.ctxInfo.GetSessionId(),
	}
	var ui *aguiUIContext
	if resp, err := svc.Get(ctx, get); err == nil {
		existing, refusal := existingThread(resp.Session, rc, binding)
		if refusal != nil {
			return nil, refusal
		}
		ui = existing
	}
	if len(rc.toolResults) > 0 {
		var sess session.Session
		if ui != nil {
			sess = ui.sess
		}
		parts, err := toolResultParts(sess, rc.toolResults)
		if err != nil {
			return nil, refuseAGUI(http.StatusBadRequest, err)
		}
		rc.parts = append(rc.parts, parts...)
	}
	if ui != nil {
		return ui, nil
	}
	created, err := svc.Create(ctx, &session.CreateRequest{
		AppName:   aguiAppName,
		UserID:    rc.ctxInfo.GetUserId(),
		SessionID: rc.ctxInfo.GetSessionId(),
		State:     map[string]any{a2ui.BindingKey: binding.StateValue()},
	})
	if err != nil {
		if errors.Is(err, sessionshare.ErrIDTaken) {
			return nil, refuseAGUI(http.StatusForbidden, errThreadUnavailable)
		}
		// A concurrent first request may have created it in between.
		if resp, getErr := svc.Get(ctx, get); getErr == nil {
			return existingThread(resp.Session, rc, binding)
		}
		return nil, refuseAGUI(http.StatusServiceUnavailable, errors.New("session store unavailable, retry later"))
	}
	return &aguiUIContext{sess: created.Session, bound: true}, nil
}

// aguiEventCount is how many events the session held before the run.
func aguiEventCount(sess session.Session) int {
	if sess == nil || sess.Events() == nil {
		return 0
	}
	return sess.Events().Len()
}

// uiLive reports whether A2UI is live for the run: the client negotiated it
// and the session's binding matches.
func (r *aguiRun) uiLive() bool { return r.rc.a2uiNegotiated && r.ui.bound }

// openSink builds the run's sink over emit.
func (r *aguiRun) openSink(emit aguiEmitter) {
	rc := r.rc
	sink := newAGUISink(rc.input.ThreadID, rc.input.RunID, r.messageID, emit)
	// The run settles its record and run state before observers learn that
	// it ended.
	sink.holdRunFinished()
	if svc := r.h.getSessionService(); svc != nil {
		sink.setFinalSession(r.finalSession(svc))
	}
	if rc.stateArmed {
		sink.setSharedState(rc.stateInitial, rc.stateSnapshot)
	}
	if rc.a2uiNegotiated {
		// An A2UI client's forms track the thread's open Interrupts, so its
		// runs end by reporting every one still open — not only those they
		// raised — however the others were answered.
		sink.reportPendingInterrupts()
	}
	if r.uiLive() {
		sink.setA2UI(r.ui.sess)
	}
	r.sink = sink
}

// finalSession reads the session after the run. A run in its request reads on
// the request's context, which ends with the only observer; a Detached Run
// reads on a context its own end does not reach, with a timeout.
func (r *aguiRun) finalSession(svc session.Service) func() (session.Session, bool) {
	get := &session.GetRequest{
		AppName:   aguiAppName,
		UserID:    r.rc.ctxInfo.GetUserId(),
		SessionID: r.rc.ctxInfo.GetSessionId(),
	}
	return func() (session.Session, bool) {
		ctx := r.reqCtx
		if r.detached != nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.WithoutCancel(r.ctx), aguiPostRunTimeout)
			defer cancel()
		}
		resp, err := svc.Get(ctx, get)
		if err != nil {
			return nil, false
		}
		return resp.Session, true
	}
}

// execute runs the turn and settles how it ended. From here on the run owns
// its thread: the run state ends and the lease is released before execute
// returns, whether or not anyone still observes the run.
func (r *aguiRun) execute() {
	defer r.close()
	rc := r.rc
	ctx := r.ctx
	if r.uiLive() {
		ctx = a2ui.WithRun(ctx, &a2ui.Run{ThreadID: rc.input.ThreadID, RunID: rc.input.RunID, MessageID: r.messageID})
	}
	// Client-declared frontend tools ride the run context: the aguitool
	// toolset resolves them per invocation, so the agent sees them for this
	// run only.
	if len(rc.clientTools) > 0 {
		ctx = aguitool.WithClientTools(ctx, rc.clientTools)
	}
	if r.detached == nil {
		r.settle(r.runTurn(ctx))
		return
	}
	// The AG-UI path owns a Detached Run's record; the runner records
	// nothing for it.
	runErr := r.markRunning()
	if runErr == nil {
		runErr = r.runTurn(runner.WithoutInvocationRecording(ctx))
	}
	r.settleDetached(runErr)
}

// runTurn drives the turn through the shared streaming orchestration. A
// Detached Run runs outside any request, where nothing but the process
// itself would stop a panic on its way out: the run ends FAILED instead, and
// still releases its thread.
func (r *aguiRun) runTurn(ctx context.Context) (runErr error) {
	rc := r.rc
	if r.detached != nil {
		defer func() {
			if p := recover(); p != nil {
				log.FromContext(ctx).Error("agui detached run panicked",
					"invocation_id", rc.ctxInfo.GetUuid(), "panic", p, "stack", string(debug.Stack()))
				runErr = errors.New("the run failed with an internal error")
			}
		}()
	}
	return streamorch.Run(ctx, rc.svc,
		streamorch.AgentRef{Name: rc.agentName, ID: rc.agent.GetAgentId()},
		rc.parts, "", rc.ctxInfo, r.sink)
}

// settle ends a run that lives in its request. The runner recorded its
// Invocation.
func (r *aguiRun) settle(runErr error) {
	r.endRunState()
	if runErr == nil {
		// The emitter writes to the response; it fails only once the client
		// has gone, which a run in its request expects.
		if err := r.sink.releaseRunFinished(); err != nil {
			log.FromContext(r.reqCtx).Debug("agui failed to emit RUN_FINISHED", "err", err)
		}
		r.titleAfterSuccess()
		return
	}
	// A cancelled lease context with a live request means the lease was lost
	// (fenced out by expiry or takeover); name it instead of reporting a bare
	// context cancellation.
	if errors.Is(runErr, context.Canceled) && r.reqCtx.Err() == nil {
		runErr = errors.New(aguiLeaseLostMessage)
	}
	r.logFailure(runErr.Error())
	// The status and headers are already committed, so the failure is
	// reported in-band as RUN_ERROR rather than as an HTTP error.
	if err := r.sink.Error(runErr); err != nil {
		log.FromContext(r.reqCtx).Error("agui failed to emit RUN_ERROR", "err", err)
	}
}

// aguiOutcome is how a Detached Run ended, as its record keeps it.
type aguiOutcome struct {
	status agentsv1.InvocationStatus
	// reason is why a FAILED or CANCELLED run ended, as observers see it.
	reason string
	// code is the RUN_ERROR code of the end, if it has one.
	code string
}

// err is the end as the run's RUN_ERROR reports it.
func (o aguiOutcome) err() error {
	return &aguiRunError{message: o.reason, code: o.code}
}

// outcome claims the run's terminal state, with asyncrun's precedence: an
// accepted Stop wins; otherwise a run whose runner returned cleanly is
// SUCCEEDED, even when a shutdown or the deadline raced it; after that come
// the shutdown, the deadline, a lost lease and the run's own error. Every
// operational end is FAILED; only a person's Stop is CANCELLED.
//
// The claim is made twice: in this process, which closes the window of a
// shutdown, and on the run state, in one step with which a Stop is accepted
// on any Pod (ADR-0016 decision 3). A Stop accepted before that claim always
// ends the run CANCELLED, whether or not the run had heard of it yet; one
// after it finds nothing running.
func (r *aguiRun) outcome(runErr error) aguiOutcome {
	d := r.detached
	claim := r.h.runs.claim(d.entry)
	// Claimed in every case, so no Stop is accepted for the run past here.
	stoppedOnAnyPod := r.claimRunState()
	stopped := claim.stopped || stoppedOnAnyPod
	switch {
	case stopped:
		return aguiOutcome{status: agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED,
			reason: errAGUIRunStopped.Error(), code: aguiCodeStopped}
	case runErr == nil:
		return aguiOutcome{status: agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED}
	case claim.shutdown:
		return aguiOutcome{status: agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED, reason: aguiShutdownReason}
	case errors.Is(context.Cause(r.ctx), errAGUIRunTimedOut):
		return aguiOutcome{status: agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED, reason: aguiDeadlineReason(d.maxRun)}
	case r.ctx.Err() != nil:
		// Nothing else ends a Detached Run's context: its lease was lost,
		// with sessionguard.ErrLeaseLost as the cause when the guard names
		// one.
		return aguiOutcome{status: agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED, reason: aguiLeaseLostMessage + invocation.NoReplaySuffix}
	default:
		return aguiOutcome{status: agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED, reason: runErr.Error()}
	}
}

// settleDetached ends a Detached Run: it records the terminal state, ends
// the run state, and only then tells its observers how the run ended.
func (r *aguiRun) settleDetached(runErr error) {
	outcome := r.outcome(runErr)
	r.recordTerminal(outcome)
	r.endRunState()
	if outcome.status == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		// The fan-out never fails on an observer; an error here is an event
		// that could not be encoded.
		if err := r.sink.releaseRunFinished(); err != nil {
			log.FromContext(r.ctx).Error("agui failed to emit RUN_FINISHED", "err", err)
		}
		r.titleAfterSuccess()
		return
	}
	r.logFailure(outcome.reason)
	if err := r.sink.Error(outcome.err()); err != nil {
		log.FromContext(r.ctx).Error("agui failed to emit RUN_ERROR", "err", err)
	}
}

func (r *aguiRun) logFailure(reason string) {
	log.FromContext(r.reqCtx).Error("agui run failed",
		"workspace_id", r.rc.ctxInfo.GetWorkspaceId(),
		"agent_id", r.rc.agent.GetAgentId(),
		"thread_id", r.rc.input.ThreadID,
		"run_id", r.rc.input.RunID,
		"invocation_id", r.rc.ctxInfo.GetUuid(),
		"detached", r.detached != nil,
		"err", reason,
	)
}

// titleAfterSuccess titles a thread that had no title when the run started,
// in the background: no observer need stay for it, and the next run on the
// thread need not wait for it.
func (r *aguiRun) titleAfterSuccess() {
	if !aguiThreadTitled(r.ui.sess) {
		r.h.titleThread(r.ctx, r.rc.ctxInfo)
	}
}

// close releases the thread: the run state ends (settling already ended it,
// unless a panic cut the run short) and the lease is released. For a
// Detached Run it then ends the observers' streams, so a client that stayed
// sees its stream end only once the thread is free for its next run.
func (r *aguiRun) close() {
	r.endRunState()
	r.release()
	if r.detached != nil {
		r.detached.fanout.close()
		r.unregister()
	}
}

// unregister drops a Detached Run from the runs in flight and frees its
// context.
func (r *aguiRun) unregister() {
	d := r.detached
	if d == nil {
		return
	}
	if d.entry != nil {
		r.h.runs.done(d.entry)
		d.entry = nil
	}
	d.cancel()
}

// --- The run state -----------------------------------------------------------------

// aguiLeaseToken is the token of the lease acquisition that made leaseCtx.
// Without a guard that names one, the run's Invocation ID, unique to the run,
// stands in for it.
func aguiLeaseToken(leaseCtx context.Context, invocationID string) string {
	if token, ok := sessionguard.Token(leaseCtx); ok {
		return token
	}
	return "invocation:" + invocationID
}

// keepRunState records the run's state next to its lease and keeps it
// renewed for as long as the run holds the lease (ADR-0016 decision 6). A
// Detached Run also learns from it that a Stop was accepted for it, on any
// Pod, and is cancelled then (decision 4).
func (r *aguiRun) keepRunState(eventCount int) error {
	store := r.h.getRunStateStore()
	if store == nil {
		return nil
	}
	kept, err := runstate.Keep(r.ctx, store, r.thread, runstate.State{
		RunID:        r.rc.input.RunID,
		InvocationID: r.rc.ctxInfo.GetUuid(),
		EventCount:   eventCount,
		Detached:     r.detached != nil,
		LeaseToken:   r.leaseToken,
	})
	if err != nil {
		return err
	}
	r.runState = kept
	if d := r.detached; d != nil {
		entry, done := d.entry, r.ctx.Done()
		go func() {
			select {
			case <-kept.Stopped():
				r.h.runs.observeStop(entry)
			case <-done:
				// Whatever ended the run, a Stop accepted before its claim
				// still decides its end there.
			}
		}()
	}
	return nil
}

// claimRunState claims the run's end on its run state, in one step with which
// a Stop is accepted on any Pod, and reports whether one was. A claim that
// cannot be made leaves the end to what this process saw.
func (r *aguiRun) claimRunState() bool {
	if r.runState == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.ctx), aguiRunStateEndTimeout)
	defer cancel()
	claim, err := r.runState.Claim(ctx)
	if err != nil {
		log.FromContext(r.ctx).Warn("agui run could not claim its end on the run state; a Stop from another Pod may be missed",
			"session_id", r.rc.ctxInfo.GetSessionId(), "err", err)
		return false
	}
	return claim.Stopped
}

// endRunState removes the run's state, unless another run's replaced it.
//
// ADR-0016 decision 6 keeps an ended Detached Run's state, marked ended, as
// long as its Run Log, so a late attach still finds the run. The Run Log
// arrives with #404, and with it that retention; until then nothing reads an
// ended run's state, so it is removed.
func (r *aguiRun) endRunState() {
	if r.runState == nil {
		return
	}
	kept := r.runState
	r.runState = nil
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.ctx), aguiRunStateEndTimeout)
	defer cancel()
	if err := kept.End(ctx); err != nil {
		log.FromContext(r.ctx).Warn("agui run state not removed; it lapses on its own",
			"session_id", r.rc.ctxInfo.GetSessionId(), "err", err)
	}
}

// --- The Invocation record of a Detached Run ---------------------------------

// checkRunID refuses a runId the thread already ran, so the Agent never runs
// twice for one request (ADR-0009).
func (r *aguiRun) checkRunID(ctx context.Context) *aguiRefusal {
	rc := r.rc
	_, err := r.h.getInvocations().FindByRequestID(ctx, rc.ctxInfo.GetWorkspaceId(), aguiRequestID(rc.ctxInfo, rc.input.RunID))
	switch {
	case err == nil:
		return &aguiRefusal{status: http.StatusConflict, body: aguiCodedError{
			Error: "runId " + rc.input.RunID + " already ran on this thread; send a new runId to run again",
			Code:  aguiCodeRunExists,
		}}
	case errors.Is(err, invocation.ErrNotFound):
		return nil
	default:
		log.FromContext(r.ctx).Error("agui invocation lookup failed", "session_id", rc.ctxInfo.GetSessionId(), "err", err)
		return refuseAGUI(http.StatusServiceUnavailable, errors.New("invocation store unavailable, retry later"))
	}
}

// recordQueued writes the run's Invocation record, QUEUED. The store stamps
// it with this process as its owner.
func (r *aguiRun) recordQueued() error {
	rc := r.rc
	displayName := rc.agent.GetDisplayName()
	if displayName == "" {
		displayName = rc.agent.GetName()
	}
	r.detached.inv = &agentsv1.Invocation{
		Id:               rc.ctxInfo.GetUuid(),
		AgentName:        rc.agentName,
		AgentId:          rc.agent.GetAgentId(),
		AgentDisplayName: displayName,
		AppName:          aguiAppName,
		UserId:           rc.ctxInfo.GetUserId(),
		SessionId:        rc.ctxInfo.GetSessionId(),
		Status:           agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		Input:            aguiTruncate(aguiTurnText(rc.parts), aguiRecordTextLimit),
		StartedAt:        timestamppb.Now(),
		Source:           invocation.SourceAGUIDetached,
		RequestId:        aguiRequestID(rc.ctxInfo, rc.input.RunID),
		WorkspaceId:      rc.ctxInfo.GetWorkspaceId(),
	}
	return r.saveRecord()
}

// markRunning moves the record to RUNNING as the turn starts.
func (r *aguiRun) markRunning() error {
	inv := r.detached.inv
	inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
	inv.StartedAt = timestamppb.Now()
	if err := r.saveRecord(); err != nil {
		return fmt.Errorf("failed to persist RUNNING status: %w", err)
	}
	return nil
}

// recordTerminal writes the run's terminal state.
func (r *aguiRun) recordTerminal(outcome aguiOutcome) {
	inv := r.detached.inv
	now := timestamppb.Now()
	inv.Status = outcome.status
	inv.Error = aguiTruncate(outcome.reason, aguiRecordTextLimit)
	inv.Output = ""
	if outcome.status == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		inv.Output = aguiTruncate(r.sink.response, aguiRecordTextLimit)
	}
	inv.FinishedAt = now
	if started := inv.GetStartedAt(); started != nil {
		inv.LatencyMs = now.AsTime().Sub(started.AsTime()).Milliseconds()
	}
	if err := r.saveRecord(); err != nil {
		log.FromContext(r.ctx).Error("agui failed to persist the terminal status of its run",
			"invocation_id", inv.GetId(), "status", outcome.status.String(), "err", err)
	}
}

// saveRecord writes the record on a context the run's end does not reach.
func (r *aguiRun) saveRecord() error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.ctx), aguiRecordTimeout)
	defer cancel()
	return r.h.getInvocations().Save(ctx, r.detached.inv)
}

// aguiTurnText is the text of the turn a run starts from, as its record
// keeps it: its text parts joined as the runner joins them for its records.
func aguiTurnText(parts []*genai.Part) string {
	var text []string
	for _, p := range parts {
		if p != nil && p.Text != "" && !p.Thought {
			text = append(text, p.Text)
		}
	}
	return strings.Join(text, " ")
}

// aguiTruncate cuts s to at most limit bytes on a rune boundary.
func aguiTruncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
}

// --- The Detached Runs in flight ------------------------------------------------

// aguiRuns tracks this process's Detached Runs in flight: what ends them
// (a Stop they heard of, a shutdown) and this process's claim on their
// terminal state. A Stop itself is accepted on the run state, on any Pod
// (runstate.Store.Stop); a run hears of it there.
type aguiRuns struct {
	mu           sync.Mutex
	entries      map[string]*aguiRunEntry // by invocation ID
	shuttingDown bool
	wg           sync.WaitGroup
}

// aguiRunEntry is one Detached Run in flight.
type aguiRunEntry struct {
	invocationID string
	cancel       context.CancelCauseFunc
	// stopRequested marks a person's Stop the run heard of; the run ends
	// CANCELLED.
	stopRequested bool
	// shutdownRequested marks a graceful shutdown; the run ends FAILED with
	// a shutdown reason, never CANCELLED, so a person's intent stays
	// distinguishable from an operational end.
	shutdownRequested bool
	// finished is set when the run claims its terminal state; nothing
	// changes it afterwards.
	finished bool
}

func newAGUIRuns() *aguiRuns {
	return &aguiRuns{entries: map[string]*aguiRunEntry{}}
}

// register adds a run about to take its thread; false once shutting down.
func (rs *aguiRuns) register(invocationID string, cancel context.CancelCauseFunc) (*aguiRunEntry, bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.shuttingDown {
		return nil, false
	}
	e := &aguiRunEntry{invocationID: invocationID, cancel: cancel}
	rs.entries[invocationID] = e
	rs.wg.Add(1)
	return e, true
}

// done drops a run that released its thread.
func (rs *aguiRuns) done(e *aguiRunEntry) {
	rs.mu.Lock()
	if rs.entries[e.invocationID] == e {
		delete(rs.entries, e.invocationID)
	}
	rs.mu.Unlock()
	rs.wg.Done()
}

// aguiTerminalClaim reports why a run's context was interrupted, if it was.
type aguiTerminalClaim struct {
	stopped  bool
	shutdown bool
}

// claim closes the window in which a shutdown, or a Stop the run heard of,
// can still decide the run's end within this process. The run then claims
// its end on the run state too, where a Stop from any Pod was accepted or
// not (aguiRun.outcome).
func (rs *aguiRuns) claim(e *aguiRunEntry) aguiTerminalClaim {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	e.finished = true
	return aguiTerminalClaim{
		stopped:  e.stopRequested,
		shutdown: e.shutdownRequested && !e.stopRequested,
	}
}

// observeStop cancels a run that heard a Stop was accepted for it, unless it
// already claimed its end. The Stop decides the run's end CANCELLED.
func (rs *aguiRuns) observeStop(e *aguiRunEntry) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if e.finished {
		return
	}
	e.stopRequested = true
	e.cancel(errAGUIRunStopped)
}

// waitForRuns blocks until no Detached Run is in flight. Tests use it.
func (h *AGUIHandler) waitForRuns() { h.runs.wg.Wait() }

// shutdown ends every run in flight, FAILED with a shutdown reason, refuses
// new ones, and waits for the runs to release their threads until ctx ends.
func (rs *aguiRuns) shutdown(ctx context.Context) error {
	rs.mu.Lock()
	rs.shuttingDown = true
	for _, e := range rs.entries {
		if !e.finished {
			e.shutdownRequested = true
			e.cancel(errAGUIShutdown)
		}
	}
	rs.mu.Unlock()

	done := make(chan struct{})
	go func() {
		rs.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
