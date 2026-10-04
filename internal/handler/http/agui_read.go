package http

import (
	"context"
	"errors"
	"iter"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/gin-gonic/gin"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/runstate"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Thread reads (ADR-0016 decision 6): the thread history and the UI snapshot
// never wait for a run. They take no lease. A read during a run answers at
// once with the run (running), and leaves out what the run has produced so
// far: it shows the thread as the run found it, plus the turn the run started
// from (aguiCutAt). The rest of the run reaches a client through the run
// itself.
//
// A read holds no lock. It reads the run state before and after the session
// and starts over when the two differ, so a run that starts or ends while the
// session is read never shows as half a reply. A run that both starts and
// ends while the session is read leaves the run state as the read first found
// it, but it changes the thread's latest Invocation record, which the read
// compares as well.
//
// A Detached Run's state outlives the run, marked ended, as long as its Run
// Log (#404). A read takes an ended run for one that is not running.

// aguiReadAttempts bounds how many passes one thread read makes: it starts
// over after a run started or ended while it read the session, and waits for
// a run in flight that has not stored the turn it started from yet.
const aguiReadAttempts = 6

// aguiReadBackoff is the pause before a thread read's second pass; it doubles
// before each later one, so all the pauses add up to under a second.
const aguiReadBackoff = 25 * time.Millisecond

// aguiUserAuthor is the author ADK gives the events of a user's turn.
const aguiUserAuthor = "user"

// lastRun statuses: how the thread's latest run ended, when it did not
// succeed.
const (
	aguiLastRunFailed    = "failed"
	aguiLastRunCancelled = "cancelled"
)

// aguiRunning names the run in flight on a thread. Its output is left out of
// the reads until it ends.
type aguiRunning struct {
	RunID        string `json:"runId"`
	InvocationID string `json:"invocationId"`
}

// aguiLastRun reports the thread's latest run when it failed or was stopped,
// from its Invocation record, so a client can show the outcome after a reload
// and offer the input again.
type aguiLastRun struct {
	// Status is "failed" or "cancelled".
	Status string `json:"status"`
	// Error is why the run ended, as its record keeps it.
	Error string `json:"error"`
	// Input is the text the user sent.
	Input string `json:"input"`
}

// aguiThreadRead is one consistent read of a thread.
type aguiThreadRead struct {
	threadID string
	// sess is the thread's session as the read shows it — cut where the run
	// in flight started — or nil when the thread has no session the caller
	// may see.
	sess session.Session
	// running is the run in flight, nil when none.
	running *aguiRunning
	// lastRun is the thread's latest run when no run is in flight and that
	// one failed or was stopped; nil otherwise.
	lastRun *aguiLastRun
}

// aguiThreadTarget is the thread a thread endpoint names, resolved for its
// caller: the thread, its key, the read of its session, and the binding that
// session must carry for the caller to reach the thread.
type aguiThreadTarget struct {
	threadID    string
	workspaceID string
	// thread is the thread's key, as its lease and run state are kept.
	thread   string
	sessions session.Service
	// request reads the thread's session: the caller's, in the AG-UI app.
	request *session.GetRequest
	binding a2ui.Binding
}

// resolveThread resolves the thread a thread endpoint names, with the checks
// all of them share: the caller reaches the agent in the request's workspace
// (401, 404 or 503 otherwise), the request names a thread (400), and the
// session store is wired (503). It answers those errors itself (ok false).
// Whether the caller holds the thread is the binding's to say, which each
// endpoint checks against its own read of the session.
func (h *AGUIHandler) resolveThread(c *gin.Context) (aguiThreadTarget, bool) {
	target, ok := h.resolveAgent(c)
	if !ok {
		return aguiThreadTarget{}, false
	}
	ctx := c.Request.Context()
	threadID := strings.TrimSpace(c.Param("thread_id"))
	if threadID == "" {
		c.JSON(http.StatusBadRequest, aguiErrorResponse{Error: "threadId is required"})
		return aguiThreadTarget{}, false
	}
	svc := h.getSessionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "session service unavailable"})
		return aguiThreadTarget{}, false
	}
	ctxInfo := &agentsv1.ContextInfo{UserId: aguiUserID(ctx), SessionId: aguiSessionPrefix + threadID}
	return aguiThreadTarget{
		threadID:    threadID,
		workspaceID: target.workspaceID,
		thread:      aguiSessionKey(ctxInfo),
		sessions:    svc,
		request:     &session.GetRequest{AppName: aguiAppName, UserID: ctxInfo.GetUserId(), SessionID: ctxInfo.GetSessionId()},
		binding:     aguiBinding(ctx, target.workspaceID, target.agent.GetAgentId(), threadID),
	}, true
}

// readThread reads the thread a read endpoint names and answers any error
// itself (ok false). It takes no lease. sess is nil when the thread has no
// session or is bound to another caller, workspace or agent, so the endpoint
// answers an empty body, with no run, without saying which.
func (h *AGUIHandler) readThread(c *gin.Context) (aguiThreadRead, bool) {
	target, ok := h.resolveThread(c)
	if !ok {
		return aguiThreadRead{}, false
	}
	ctx := c.Request.Context()
	reader := &aguiThreadReader{
		aguiThreadTarget: target,
		runStates:        h.getRunStateStore(),
		invocations:      h.getInvocations(),
	}
	attempts, backoff := h.readRetry()
	for attempt := 1; ; attempt++ {
		pass, err := reader.read(ctx)
		if err != nil {
			log.FromContext(ctx).Error("agui thread read: run state unavailable",
				"session_id", target.request.SessionID, "err", err)
			c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "run state unavailable, retry later"})
			return aguiThreadRead{}, false
		}
		last := attempt >= attempts
		// A run that has not stored its first turn yet is about to: the read
		// waits for it a little, then answers without it.
		if pass.stable && (!pass.startPending || last) {
			pass.read.threadID = target.threadID
			return pass.read, true
		}
		if last {
			log.FromContext(ctx).Warn("agui thread read: runs kept starting or ending while the thread was read",
				"session_id", target.request.SessionID, "attempts", attempts)
			c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "the thread changed while it was read, retry"})
			return aguiThreadRead{}, false
		}
		pause := time.NewTimer(backoff << min(attempt-1, 6))
		select {
		case <-ctx.Done():
			pause.Stop()
			c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "the read was cancelled"})
			return aguiThreadRead{}, false
		case <-pause.C:
		}
	}
}

// aguiThreadReader reads one caller's thread.
type aguiThreadReader struct {
	aguiThreadTarget
	// runStates is where runs record their run state; without it no run
	// records one, and a read never cuts.
	runStates runstate.Store
	// invocations holds the runs' Invocation records; without it a read
	// reports no lastRun.
	invocations invocation.Repository
}

// aguiReadPass is one pass of a read.
type aguiReadPass struct {
	read aguiThreadRead
	// stable reports that no run started or ended while the session was
	// read. An unstable pass shows nothing; the read starts over.
	stable bool
	// startPending reports a run in flight that has not stored the turn it
	// started from yet. It stores that turn first, so a later pass shows it.
	startPending bool
}

// read reads the thread once: the run state, the session, and the run state
// again. A run that started or ended in between changed the run state, and
// the pass is unstable. With a run in flight, the pass shows the session as
// that run found it.
//
// With no run in flight it also reads the session's latest Invocation record,
// before the run state it compares and again after the session. Runs record
// their Invocation before their first event: a Detached Run always, the
// runner unless the write fails. So a run that both started and ended while
// the session was read changed the latest record, which the run state alone
// cannot show, and the pass is unstable. Otherwise the record gives lastRun.
//
// Only a run state that cannot be read is an error.
func (r *aguiThreadReader) read(ctx context.Context) (aguiReadPass, error) {
	stateBefore, wasRunning, err := r.runState(ctx)
	if err != nil {
		return aguiReadPass{}, err
	}
	var recordBefore aguiRunRecord
	if !wasRunning {
		recordBefore = r.latestRun(ctx)
		if stateBefore, wasRunning, err = r.runState(ctx); err != nil {
			return aguiReadPass{}, err
		}
	}
	resp, err := r.sessions.Get(ctx, r.request)
	if err != nil || !a2ui.Bound(resp.Session, r.binding) {
		// Nothing the caller may see, and so no run to report either.
		return aguiReadPass{stable: true}, nil
	}
	stateAfter, isRunning, err := r.runState(ctx)
	if err != nil {
		return aguiReadPass{}, err
	}
	if wasRunning != isRunning || stateBefore != stateAfter {
		return aguiReadPass{}, nil
	}
	sess := resp.Session
	if isRunning {
		cut, startStored := aguiCutAt(sess, stateBefore.EventCount)
		return aguiReadPass{
			read: aguiThreadRead{
				sess:    cut,
				running: &aguiRunning{RunID: stateBefore.RunID, InvocationID: stateBefore.InvocationID},
			},
			stable:       true,
			startPending: !startStored && sess.Events().Len() <= stateBefore.EventCount,
		}, nil
	}
	recordAfter := r.latestRun(ctx)
	if recordBefore.known && recordAfter.known && recordBefore.inv.GetId() != recordAfter.inv.GetId() {
		return aguiReadPass{}, nil
	}
	var lastRun *aguiLastRun
	if recordAfter.known {
		lastRun = aguiLastRunOf(r.threadRun(ctx, recordAfter.inv))
	}
	return aguiReadPass{read: aguiThreadRead{sess: sess, lastRun: lastRun}, stable: true}, nil
}

// runState reads the thread's run state, and whether it is of a run in
// flight: one that ended is kept, but no longer runs.
func (r *aguiThreadReader) runState(ctx context.Context) (st runstate.State, running bool, err error) {
	if r.runStates == nil {
		return runstate.State{}, false, nil
	}
	st, held, err := r.runStates.Get(ctx, r.thread)
	if err != nil {
		return runstate.State{}, false, err
	}
	return st, held && !st.Ended, nil
}

// aguiRunRecord is the session's latest Invocation record, as one lookup
// found it.
type aguiRunRecord struct {
	// inv is nil when the session has no record.
	inv *agentsv1.Invocation
	// known is false when there is no store, or the lookup failed.
	known bool
}

// latestRun looks up the latest Invocation record under the thread's session
// ID. A failed lookup only costs the read its lastRun.
func (r *aguiThreadReader) latestRun(ctx context.Context) aguiRunRecord {
	if r.invocations == nil {
		return aguiRunRecord{}
	}
	inv, err := r.invocations.FindLatestBySession(ctx, r.workspaceID, r.request.SessionID)
	switch {
	case err == nil:
		return aguiRunRecord{inv: inv, known: true}
	case errors.Is(err, invocation.ErrNotFound):
		return aguiRunRecord{known: true}
	default:
		log.FromContext(ctx).Warn("agui thread read: latest run record unavailable",
			"session_id", r.request.SessionID, "err", err)
		return aguiRunRecord{}
	}
}

// threadRun is the thread's latest run record, given the latest record under
// its session ID. That one is the thread's unless another app or caller keeps
// records under the same session ID, as data from before a session ID had one
// holder may; then the thread's latest is looked for among them all.
func (r *aguiThreadReader) threadRun(ctx context.Context, latest *agentsv1.Invocation) *agentsv1.Invocation {
	if latest == nil || r.ownRun(latest) {
		return latest
	}
	all, err := r.invocations.ListBySession(ctx, r.workspaceID, r.request.SessionID)
	if err != nil {
		log.FromContext(ctx).Warn("agui thread read: run records unavailable",
			"session_id", r.request.SessionID, "err", err)
		return nil
	}
	for _, inv := range all {
		if r.ownRun(inv) {
			return inv
		}
	}
	return nil
}

// ownRun reports whether inv records one of this thread's runs: one of this
// caller's, in the AG-UI app.
func (r *aguiThreadReader) ownRun(inv *agentsv1.Invocation) bool {
	return inv.GetAppName() == aguiAppName && inv.GetUserId() == r.request.UserID
}

// aguiLastRunOf is lastRun for the thread's latest run record: set when that
// run failed or was stopped, and omitted once a later run succeeds.
func aguiLastRunOf(inv *agentsv1.Invocation) *aguiLastRun {
	if inv == nil {
		return nil
	}
	var status string
	switch inv.GetStatus() {
	case agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED:
		status = aguiLastRunFailed
	case agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED:
		status = aguiLastRunCancelled
	default:
		return nil
	}
	return &aguiLastRun{Status: status, Error: inv.GetError(), Input: inv.GetInput()}
}

// --- The cut -------------------------------------------------------------------

// aguiCutAt is sess as a read shows it while a run is in flight that found
// eventCount events in it (ADR-0016 decision 6): the events before the run,
// then the turn the run started from, and the UI state those events left.
// startStored reports whether that turn is stored yet.
//
// The turn is the run's first event: the user's message, a Human Input
// answer or tool results. It is kept because a client already shows it, a
// resumed reply hangs under the history's last message, and an Interrupt the
// run answered must not read as open. Everything else the run stored is left
// out, its cards included.
func aguiCutAt(sess session.Session, eventCount int) (cut session.Session, startStored bool) {
	events := sess.Events()
	n := min(max(eventCount, 0), events.Len())
	kept := make(aguiEventList, 0, n+1)
	for i := range n {
		kept = append(kept, events.At(i))
	}
	if n == eventCount && n < events.Len() {
		if ev := events.At(n); ev != nil && ev.Author == aguiUserAuthor {
			kept = append(kept, ev)
			startStored = true
		}
	}
	return &aguiCutSession{Session: sess, events: kept, state: aguiStateAt(sess.State(), kept)}, startStored
}

// aguiStateAt is the session's state with its cards as events left them.
// Cards live in session state (ADR-0014 §3), which already holds every card
// write of a run in flight; a card is only ever written through an event's
// state delta, so the kept events say what each card was before the run.
// Every other key, the UI binding among them, is the session's own.
func aguiStateAt(st session.State, events []*session.Event) aguiStateMap {
	out := aguiStateMap{}
	if st != nil {
		for key, v := range st.All() {
			if _, isCard := a2ui.CardIDFromKey(key); !isCard {
				out[key] = v
			}
		}
	}
	for _, ev := range events {
		if ev == nil {
			continue
		}
		for key, v := range ev.Actions.StateDelta {
			if _, isCard := a2ui.CardIDFromKey(key); isCard {
				out[key] = v
			}
		}
	}
	return out
}

// aguiCutSession is a session cut where a run started: its own identity, the
// kept events and the state they left.
type aguiCutSession struct {
	session.Session
	events aguiEventList
	state  aguiStateMap
}

func (s *aguiCutSession) Events() session.Events { return s.events }
func (s *aguiCutSession) State() session.State   { return s.state }

// aguiEventList is a fixed list of session events.
type aguiEventList []*session.Event

func (e aguiEventList) All() iter.Seq[*session.Event] { return slices.Values(e) }
func (e aguiEventList) Len() int                      { return len(e) }
func (e aguiEventList) At(i int) *session.Event       { return e[i] }

// aguiStateMap is a read-only session state.
type aguiStateMap map[string]any

func (s aguiStateMap) Get(key string) (any, error) {
	v, ok := s[key]
	if !ok {
		return nil, session.ErrStateKeyNotExist
	}
	return v, nil
}

func (s aguiStateMap) Set(string, any) error {
	return errors.New("a thread read's session state is read-only")
}

func (s aguiStateMap) All() iter.Seq2[string, any] { return maps.All(s) }
