package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/a2ui"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Stop (ADR-0016 decisions 4 and 7): a person ends a Detached Run, on
// whichever Pod runs it. A Stop is accepted on the thread's run state, never
// on its lease, so it never waits behind the run it stops: in one step it
// marks the running run's state with the run's lease token and nudges the
// run (runstate.Store.Stop). The run cancels its turn — a Pi or Cursor Agent
// through its AbortSession path — and ends CANCELLED, its observers getting
// RUN_ERROR with the stop code. The Stop endpoint, CancelAgentInvocation on
// an AG-UI-owned Invocation, and deleting a thread all go through it. A run
// without the opt-in is never stopped: it still ends with its request.

// aguiDeleteLeaseWait bounds how long deleting a thread waits for its lease,
// once its Detached Run was told to stop.
const aguiDeleteLeaseWait = 10 * time.Second

// aguiDeleteRetryInterval caps the pause between two attempts to take a
// thread's lease for its delete.
const aguiDeleteRetryInterval = 250 * time.Millisecond

// aguiStopAccepted is the body of a Stop that reached a run: the run that
// will end CANCELLED.
type aguiStopAccepted struct {
	ThreadID     string `json:"threadId"`
	RunID        string `json:"runId"`
	InvocationID string `json:"invocationId"`
}

// StopRun handles POST /api/agui/:agent_id/threads/:thread_id/stop. It asks
// the thread's Detached Run to stop, on any Pod, and answers at once: 202
// with the run once the Stop is accepted, 204 when no Detached Run is in
// flight — an idle thread, a run without the opt-in, a run that already
// ended, or a thread the caller does not hold. It is idempotent: a second
// Stop before the run ends is accepted again, and one after it answers 204.
func (h *AGUIHandler) StopRun(c *gin.Context) {
	threadID, thread, ok := h.stopTarget(c)
	if !ok {
		return
	}
	if thread == "" {
		c.Status(http.StatusNoContent)
		return
	}
	store := h.getRunStateStore()
	if store == nil {
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "stop unavailable: run state store unavailable"})
		return
	}
	ctx := c.Request.Context()
	st, accepted, err := store.Stop(ctx, thread, "")
	if err != nil {
		log.FromContext(ctx).Error("agui stop failed", "thread_id", threadID, "err", err)
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "stop unavailable, retry later"})
		return
	}
	if !accepted {
		c.Status(http.StatusNoContent)
		return
	}
	log.FromContext(ctx).Info("agui run stop accepted",
		"thread_id", threadID, "run_id", st.RunID, "invocation_id", st.InvocationID)
	c.JSON(http.StatusAccepted, aguiStopAccepted{ThreadID: threadID, RunID: st.RunID, InvocationID: st.InvocationID})
}

// stopTarget resolves the thread a Stop names, with the checks of the thread
// reads (resolveThread): the caller reaches the agent in the request's
// workspace (401, 404, 503 otherwise), and the thread's session carries this
// caller's binding for that workspace and agent. thread is the thread's key,
// or "" when the caller does not hold the thread that way, which leaves them
// nothing to stop and reveals nothing about whose thread it is. It answers
// any error itself (ok false).
func (h *AGUIHandler) stopTarget(c *gin.Context) (threadID, thread string, ok bool) {
	target, ok := h.resolveThread(c)
	if !ok {
		return "", "", false
	}
	ctx := c.Request.Context()
	// The binding lives in the session's state: one recent event is enough.
	get := *target.request
	get.NumRecentEvents = 1
	resp, err := target.sessions.Get(ctx, &get)
	if err != nil {
		if aguiSessionMissing(err) {
			return target.threadID, "", true
		}
		log.FromContext(ctx).Error("agui stop could not read the thread", "thread_id", target.threadID, "err", err)
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "session store unavailable, retry later"})
		return "", "", false
	}
	if !a2ui.Bound(resp.Session, target.binding) {
		return target.threadID, "", true
	}
	return target.threadID, target.thread, true
}

// aguiSessionMissing reports whether a session read failed because the
// session does not exist, as both session stores word it.
func aguiSessionMissing(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}

// StopInvocation stops the AG-UI run with this Invocation, on any Pod, while
// it runs: CancelAgentInvocation routes AG-UI-owned Invocations here. Only
// that run is stopped, never a later one on its thread. stopped reports
// whether the Stop was accepted; the run then ends CANCELLED shortly after.
func (h *AGUIHandler) StopInvocation(ctx context.Context, userID, sessionID, invocationID string) (bool, error) {
	store := h.getRunStateStore()
	if store == nil {
		// Without a run state no run can detach, so none is running.
		return false, nil
	}
	_, accepted, err := store.Stop(ctx, aguiSessionKey(&agentsv1.ContextInfo{UserId: userID, SessionId: sessionID}), invocationID)
	if err != nil {
		return false, err
	}
	if accepted {
		log.FromContext(ctx).Info("agui run stop accepted", "session_id", sessionID, "invocation_id", invocationID)
	}
	return accepted, nil
}

// HoldThreadForDelete gets an AG-UI thread ready to be deleted (ADR-0016
// decision 7). It stops the thread's Detached Run, then takes the thread's
// lease itself, waiting at most a bound while the run lets it go, and drops
// the thread's run state under it. Holding the lease, the caller deletes the
// thread on holdCtx, so no run starts or writes in between on any Pod, then
// calls release.
//
// held is false when the lease did not come free in time — for instance
// under a run without the opt-in, which a Stop does not reach: nothing was
// changed, and the delete can be retried.
func (h *AGUIHandler) HoldThreadForDelete(ctx context.Context, userID, sessionID string) (holdCtx context.Context, release func(), held bool, err error) {
	thread := aguiSessionKey(&agentsv1.ContextInfo{UserId: userID, SessionId: sessionID})
	store := h.getRunStateStore()
	guard := h.getSessionGuard()
	if guard == nil {
		// Without a guard the handler serializes no runs, so there is no
		// lease to take and no run to wait for.
		return ctx, func() {}, true, nil
	}
	deadline := time.Now().Add(h.deleteLeaseWait())
	pause := 25 * time.Millisecond
	for {
		if store != nil {
			// On every attempt: a run that took the thread in between is
			// stopped too.
			if _, _, err := store.Stop(ctx, thread, ""); err != nil {
				return nil, nil, false, fmt.Errorf("stop the thread's run: %w", err)
			}
		}
		leaseCtx, releaseLease, acquired, err := guard.Acquire(ctx, thread)
		if err != nil {
			return nil, nil, false, fmt.Errorf("take the thread's lease: %w", err)
		}
		if acquired {
			if store != nil {
				// No run holds the thread: what is left belongs to a run that
				// ended or died, and the deleted thread must not read as
				// running.
				if err := store.Drop(leaseCtx, thread); err != nil {
					releaseLease()
					return nil, nil, false, fmt.Errorf("drop the thread's run state: %w", err)
				}
			}
			return leaseCtx, releaseLease, true, nil
		}
		wait := min(pause, time.Until(deadline))
		if wait <= 0 {
			log.FromContext(ctx).Warn("agui thread delete gave up waiting for its lease",
				"session_id", sessionID, "waited", h.deleteLeaseWait().String())
			return nil, nil, false, nil
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, nil, false, ctx.Err()
		case <-timer.C:
		}
		pause = min(pause*2, aguiDeleteRetryInterval)
	}
}
