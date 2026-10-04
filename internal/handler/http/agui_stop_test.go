package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/runlog"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/runstate"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// The Stop (#402, ADR-0016 decisions 3, 4 and 7), through the HTTP contract
// and the RPCs that reach it, with a real runner, a real session store and
// the fake model. Two instances of the endpoint stand for two Pods: each has
// its own runner and handler over the same sessions and Invocation records,
// and they share thread leases, run states and Run Logs — one in-process
// copy, or, with REDIS_ADDR set, one Redis that each reaches through its own
// clients.

// podStores are what one instance shares its threads through.
type podStores struct {
	guard sessionguard.Guard
	runs  runstate.Store
	logs  runlog.Store
}

// memoryPods share one in-process lease guard, run state store and run log
// store.
func memoryPods(*testing.T) func(pod string) podStores {
	guard, runs, logs := sessionguard.NewMemory(), runstate.NewMemory(time.Minute), runlog.NewMemory()
	return func(string) podStores { return podStores{guard: guard, runs: runs, logs: logs} }
}

// redisPods give each instance its own lease guard, run state store and run
// log store over one Redis, as two Pods have: a Stop's nudge, and a run's
// events, travel through Redis.
func redisPods(t *testing.T) func(pod string) podStores {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is required for the cross-Pod Stop over Redis")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	prefix := "butter:test:" + uuid.NewString() + ":agui:"
	t.Cleanup(func() {
		keys, _ := rdb.Keys(context.Background(), prefix+"*").Result()
		if len(keys) > 0 {
			_ = rdb.Del(context.Background(), keys...).Err()
		}
		_ = rdb.Close()
	})
	const ttl = 3 * time.Second
	return func(pod string) podStores {
		runs := runstate.NewRedis(rdb, prefix+"run:", ttl)
		logs := runlog.NewRedis(rdb, prefix+"log:")
		t.Cleanup(func() {
			_ = runs.Close()
			logs.Close()
		})
		return podStores{guard: sessionguard.NewRedis(rdb, pod, prefix+"lease:", ttl), runs: runs, logs: logs}
	}
}

// acrossPods runs test once for each way two instances share threads.
func acrossPods(t *testing.T, test func(t *testing.T, pods func(pod string) podStores)) {
	t.Run("memory", func(t *testing.T) { test(t, memoryPods(t)) })
	t.Run("redis", func(t *testing.T) { test(t, redisPods(t)) })
}

// onPod runs the harness's instance of the endpoint on these stores.
func onPod(stores podStores) func(*detachHarness) {
	return func(d *detachHarness) {
		d.lease = newCountingGuard(stores.guard)
		d.runStates = stores.runs
		d.runLogs = stores.logs
	}
}

// aguiPeer is a second instance of the endpoint, as on another Pod: its own
// runner and handler over the harness's sessions and Invocation records.
type aguiPeer struct {
	router  *gin.Engine
	handler *AGUIHandler
}

func (d *detachHarness) peer(stores podStores) *aguiPeer {
	d.t.Helper()
	first := d.handler
	p := &aguiPeer{router: d.build([]agentsv1.Agent{cardAgent()}, []string{"card-model"})}
	p.handler, d.handler = d.handler, first
	p.handler.SetSessionGuard(stores.guard)
	p.handler.SetInvocationRepo(d.invocations)
	p.handler.SetRunStateStore(stores.runs)
	p.handler.SetRunLogStore(stores.logs)
	d.t.Cleanup(func() {
		done := make(chan struct{})
		go func() {
			p.handler.waitForRuns()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			d.t.Error("the peer's detached runs still in flight after 10s")
		}
	})
	return p
}

// stopRun asks router to stop the thread's Detached Run, as the dashboard's
// Stop does.
func stopRun(t *testing.T, router http.Handler, agentID, threadID string, opts ...a2uiOpt) *httptest.ResponseRecorder {
	t.Helper()
	var ro a2uiRequest
	for _, o := range opts {
		o(&ro)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/agui/"+agentID+"/threads/"+threadID+"/stop", nil)
	ro.apply(req)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// runEnd is the event a run's stream ended with.
func runEnd(t *testing.T, body string) map[string]any {
	t.Helper()
	events := sseEvents(t, body)
	if len(events) == 0 {
		t.Fatalf("no events in the stream: %q", body)
	}
	return events[len(events)-1]
}

// stoppedEnd reports whether a stream ended the way a stopped run's does.
func stoppedEnd(end map[string]any) bool {
	return end["type"] == "RUN_ERROR" && end["code"] == aguiCodeStopped && end["message"] == "stopped by user"
}

// threadFree reports whether the thread's lease can be taken, through the
// guard of yet another instance.
func threadFree(t *testing.T, guard sessionguard.Guard, thread string) bool {
	t.Helper()
	_, release, ok, err := guard.Acquire(t.Context(), thread)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if ok {
		release()
	}
	return ok
}

// memberContext is a request of a signed-in member of workspace ws-a.
func memberContext(userID string) context.Context {
	ctx := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: userID, Role: "member"}, nil)
	return wsctx.WithID(ctx, "ws-a")
}

var (
	queuedRunningCancelled = []agentsv1.InvocationStatus{
		agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED,
	}
	statusCancelled = agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED
	statusSucceeded = agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED
	statusRunning   = agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
)

// The issue's headline: a Detached Run on instance A is stopped by a request
// to instance B. The Stop is accepted at once; the run cancels its turn, ends
// CANCELLED and frees its thread, and its observer is told with the stop
// code instead of a lost lease.
func TestAGUIStop_ARunIsStoppedThroughAnotherInstance(t *testing.T) {
	acrossPods(t, func(t *testing.T, pods func(string) podStores) {
		d := newDetachHarness(t, onPod(pods("pod-a")))
		b := d.peer(pods("pod-b"))
		model := d.gate("card-model", "never delivered")

		post := d.start("carder", detachBody("t-1", "run-1", "hi"))
		model.waitStarted(t)
		inv, _ := d.record("t-1", "run-1")

		w := stopRun(t, b.router, "carder", "t-1")
		if w.Code != http.StatusAccepted {
			t.Fatalf("Stop through B: status = %d, body = %s; want 202", w.Code, w.Body.String())
		}
		var accepted aguiStopAccepted
		if err := json.Unmarshal(w.Body.Bytes(), &accepted); err != nil ||
			accepted != (aguiStopAccepted{ThreadID: "t-1", RunID: "run-1", InvocationID: inv.GetId()}) {
			t.Fatalf("Stop body = %s, want the run it stopped (%s)", w.Body.String(), inv.GetId())
		}

		if end := runEnd(t, post.wait(t).Body.String()); !stoppedEnd(end) {
			t.Fatalf("A's observer saw the run end with %v, want RUN_ERROR with code %q", end, aguiCodeStopped)
		}
		d.waitForRuns()

		inv, _ = d.record("t-1", "run-1")
		if inv.GetStatus() != statusCancelled || inv.GetError() != "stopped by user" {
			t.Fatalf("record = %+v, want CANCELLED by the Stop", inv)
		}
		if got := d.invocations.history(inv.GetId()); !slices.Equal(got, queuedRunningCancelled) {
			t.Fatalf("record statuses = %v, want %v", got, queuedRunningCancelled)
		}
		eventually(t, "the model call is cancelled", func() bool { return model.cancelled.Load() == 1 })
		if sess, _ := d.threadSession("t-1"); agentSaid(sess, "never delivered") {
			t.Fatal("the stopped run's reply was persisted")
		}
		if !threadFree(t, pods("pod-c").guard, detachThread) {
			t.Fatal("the stopped run kept its thread's lease")
		}
		if st, ok := d.runState(detachThread); ok && !st.Ended {
			t.Fatalf("run state after the Stop = %+v; want it ended", st)
		}
	})
}

// With no Detached Run in flight a Stop answers 204: on a thread nobody ran,
// after a run ended, and while a run without the opt-in holds the thread. A
// Stop never reaches such a run: it ends only with its request.
func TestAGUIStop_NothingInFlightAnswersNoContent(t *testing.T) {
	d := newDetachHarness(t)
	if w := stopRun(t, d.router, "carder", "t-never"); w.Code != http.StatusNoContent || w.Body.Len() != 0 {
		t.Fatalf("a thread nobody ran: status = %d, body = %q; want an empty 204", w.Code, w.Body.String())
	}

	d.answer("card-model", "first reply")
	if w := d.post("carder", detachBody("t-1", "run-1", "hi")); runEnd(t, w.Body.String())["type"] != "RUN_FINISHED" {
		t.Fatalf("first run: %s", w.Body.String())
	}
	d.waitForRuns()
	for i := range 2 {
		if w := stopRun(t, d.router, "carder", "t-1"); w.Code != http.StatusNoContent {
			t.Fatalf("Stop %d after the run ended: status = %d, body = %s", i+1, w.Code, w.Body.String())
		}
	}

	model := d.gate("card-model", "in its request")
	body := minimalAGUIBody("t-1", "again")
	body["runId"] = "run-2"
	post := d.start("carder", body)
	model.waitStarted(t)
	if w := stopRun(t, d.router, "carder", "t-1"); w.Code != http.StatusNoContent {
		t.Fatalf("under a run without the opt-in: status = %d, body = %s; want 204", w.Code, w.Body.String())
	}
	model.open()
	w := post.wait(t)
	if runEnd(t, w.Body.String())["type"] != "RUN_FINISHED" || !strings.Contains(w.Body.String(), "in its request") {
		t.Fatalf("the run without the opt-in did not run to its end: %s", w.Body.String())
	}
}

// A Stop never reaches a later run: once a run ended (here, stopped), a Stop
// finds nothing running, and the next run on the thread runs to its end.
func TestAGUIStop_AfterARunEndsTheNextRunIsUntouched(t *testing.T) {
	acrossPods(t, func(t *testing.T, pods func(string) podStores) {
		d := newDetachHarness(t, onPod(pods("pod-a")))
		b := d.peer(pods("pod-b"))
		model := d.gate("card-model", "never delivered")
		post := d.start("carder", detachBody("t-1", "run-1", "first"))
		model.waitStarted(t)
		if w := stopRun(t, b.router, "carder", "t-1"); w.Code != http.StatusAccepted {
			t.Fatalf("Stop: status = %d, body = %s", w.Code, w.Body.String())
		}
		post.wait(t)
		d.waitForRuns()
		if w := stopRun(t, b.router, "carder", "t-1"); w.Code != http.StatusNoContent {
			t.Fatalf("Stop after the run ended: status = %d, body = %s; want 204", w.Code, w.Body.String())
		}

		d.answer("card-model", "second reply")
		w := d.post("carder", detachBody("t-1", "run-2", "second"))
		if end := runEnd(t, w.Body.String()); end["type"] != "RUN_FINISHED" || !strings.Contains(w.Body.String(), "second reply") {
			t.Fatalf("the next run ended with %v:\n%s", end, w.Body.String())
		}
		d.waitForRuns()
		if inv, _ := d.record("t-1", "run-2"); inv.GetStatus() != statusSucceeded {
			t.Fatalf("next run's record = %+v, want SUCCEEDED", inv)
		}
	})
}

// claimHook runs a step right before, or right after, a run claims its end on
// the run state.
type claimHook struct {
	runstate.Store
	before, after func()
}

func (h *claimHook) Claim(ctx context.Context, thread, invocationID string) (runstate.Claim, error) {
	if h.before != nil {
		h.before()
	}
	claim, err := h.Store.Claim(ctx, thread, invocationID)
	if h.after != nil {
		h.after()
	}
	return claim, err
}

// Accepting a Stop and claiming the run's end are ordered on the run state
// (decision 3). A Stop accepted just before the claim ends the run CANCELLED
// although its turn had already finished; one just after it finds nothing
// running, and the run stays SUCCEEDED.
func TestAGUIStop_AnAcceptedStopAlwaysEndsTheRunCancelled(t *testing.T) {
	for _, tc := range []struct {
		name       string
		beforeEnd  bool
		wantStop   int
		wantStatus agentsv1.InvocationStatus
		wantEnd    string
	}{
		{"accepted before the claim", true, http.StatusAccepted, statusCancelled, "RUN_ERROR"},
		{"sent after the claim", false, http.StatusNoContent, statusSucceeded, "RUN_FINISHED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			acrossPods(t, func(t *testing.T, pods func(string) podStores) {
				a := pods("pod-a")
				hook := &claimHook{Store: a.runs}
				d := newDetachHarness(t, onPod(podStores{guard: a.guard, runs: hook, logs: a.logs}))
				b := d.peer(pods("pod-b"))
				stopStatus := 0
				stop := func() { stopStatus = stopRun(t, b.router, "carder", "t-1").Code }
				if tc.beforeEnd {
					hook.before = stop
				} else {
					hook.after = stop
				}
				d.answer("card-model", "finished turn")

				w := d.post("carder", detachBody("t-1", "run-1", "hi"))
				d.waitForRuns()
				if stopStatus != tc.wantStop {
					t.Fatalf("Stop status = %d, want %d", stopStatus, tc.wantStop)
				}
				if end := runEnd(t, w.Body.String()); end["type"] != tc.wantEnd {
					t.Fatalf("the stream ended with %v, want %s", end, tc.wantEnd)
				}
				if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != tc.wantStatus {
					t.Fatalf("record = %+v, want %v", inv, tc.wantStatus)
				}
			})
		})
	}
}

// beginHook runs a step the moment a run's state exists.
type beginHook struct {
	runstate.Store
	after func()
}

func (h *beginHook) Begin(ctx context.Context, thread string, st runstate.State) error {
	if err := h.Store.Begin(ctx, thread, st); err != nil {
		return err
	}
	if h.after != nil {
		h.after()
	}
	return nil
}

// A Stop that races the run's start is not lost: the run subscribed to its
// nudges before its state existed, so a Stop accepted the moment it does,
// before the stream even opened, still ends the run CANCELLED.
func TestAGUIStop_AStopRacingTheRunsStartIsNotLost(t *testing.T) {
	acrossPods(t, func(t *testing.T, pods func(string) podStores) {
		a := pods("pod-a")
		hook := &beginHook{Store: a.runs}
		d := newDetachHarness(t, onPod(podStores{guard: a.guard, runs: hook, logs: a.logs}))
		b := d.peer(pods("pod-b"))
		stopStatus := 0
		hook.after = func() { stopStatus = stopRun(t, b.router, "carder", "t-1").Code }
		d.gate("card-model", "never delivered")

		w := d.post("carder", detachBody("t-1", "run-1", "hi"))
		d.waitForRuns()
		if stopStatus != http.StatusAccepted {
			t.Fatalf("Stop as the run started: status = %d, want 202", stopStatus)
		}
		if end := runEnd(t, w.Body.String()); !stoppedEnd(end) {
			t.Fatalf("the stream ended with %v, want the stop", end)
		}
		if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != statusCancelled {
			t.Fatalf("record = %+v, want CANCELLED", inv)
		}
	})
}

// The Stop answers to the checks of the thread reads: it reaches only the
// caller's own thread, bound to the request's workspace and the route's
// agent. Anyone else's Stop reaches nothing.
func TestAGUIStop_ReachesOnlyTheCallersBoundThread(t *testing.T) {
	d := newDetachHarness(t)
	d.router = d.build([]agentsv1.Agent{cardAgent(), plainAgent("plain", "plain-model")}, []string{"card-model", "plain-model"})
	d.handler.SetSessionGuard(d.lease)
	d.handler.SetInvocationRepo(d.invocations)
	d.handler.SetRunStateStore(d.runStates)
	d.handler.SetRunLogStore(d.runLogs)
	model := d.gate("card-model", "never delivered")
	post := d.start("carder", detachBody("t-1", "run-1", "hi"))
	model.waitStarted(t)

	for _, tc := range []struct {
		name  string
		agent string
		opts  []a2uiOpt
		want  int
	}{
		{"another user, the same threadId", "carder", []a2uiOpt{asUser("u2")}, http.StatusNoContent},
		{"the thread through another agent's route", "plain", nil, http.StatusNoContent},
		{"another workspace", "carder", []a2uiOpt{inWorkspace("ws-b")}, http.StatusNotFound},
		{"an API token, the agent without enable_agui", "carder", []a2uiOpt{asAPIToken()}, http.StatusNotFound},
	} {
		if w := stopRun(t, d.router, tc.agent, "t-1", tc.opts...); w.Code != tc.want {
			t.Errorf("%s: status = %d, body = %s; want %d", tc.name, w.Code, w.Body.String(), tc.want)
		}
	}
	if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != statusRunning || model.cancelled.Load() != 0 {
		t.Fatalf("record = %+v, model cancelled %d times: another caller's Stop reached the run", inv, model.cancelled.Load())
	}

	if w := stopRun(t, d.router, "carder", "t-1"); w.Code != http.StatusAccepted {
		t.Fatalf("the caller's own Stop: status = %d, body = %s; want 202", w.Code, w.Body.String())
	}
	if end := runEnd(t, post.wait(t).Body.String()); !stoppedEnd(end) {
		t.Fatalf("the stream ended with %v, want the stop", end)
	}
}

// CancelAgentInvocation on an AG-UI-owned Invocation goes through the same
// Stop, so it reaches the run on another instance. An Invocation that ended
// stops nothing, not even a later run on its thread.
func TestCancelAgentInvocation_StopsAnAGUIRunOnAnyInstance(t *testing.T) {
	acrossPods(t, func(t *testing.T, pods func(string) podStores) {
		d := newDetachHarness(t, onPod(pods("pod-a")))
		b := d.peer(pods("pod-b"))
		agents := application.NewAgentServiceServer(nil)
		agents.SetRunnerService(b.handler.getRunner().(*runner.Service))
		agents.SetInvocationRepo(d.invocations)
		agents.SetAGUIRunStopper(b.handler)
		cancel := func(id string) bool {
			t.Helper()
			resp, err := agents.CancelAgentInvocation(memberContext("u1"),
				connect.NewRequest(&agentsv1.CancelAgentInvocationRequest{InvocationId: id}))
			if err != nil {
				t.Fatalf("CancelAgentInvocation(%s): %v", id, err)
			}
			return resp.Msg.GetCancelled()
		}

		model := d.gate("card-model", "never delivered")
		post := d.start("carder", detachBody("t-1", "run-1", "hi"))
		model.waitStarted(t)
		first, _ := d.record("t-1", "run-1")
		if !cancel(first.GetId()) {
			t.Fatal("cancelled = false for a running AG-UI run")
		}
		if end := runEnd(t, post.wait(t).Body.String()); !stoppedEnd(end) {
			t.Fatalf("the stream ended with %v, want the stop", end)
		}
		d.waitForRuns()
		if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != statusCancelled {
			t.Fatalf("record = %+v, want CANCELLED", inv)
		}

		next := d.gate("card-model", "second reply")
		post = d.start("carder", detachBody("t-1", "run-2", "again"))
		next.waitStarted(t)
		if cancel(first.GetId()) {
			t.Fatal("cancelling the ended Invocation reported a cancel")
		}
		next.open()
		if end := runEnd(t, post.wait(t).Body.String()); end["type"] != "RUN_FINISHED" {
			t.Fatalf("the later run ended with %v: the first run's Invocation reached it", end)
		}
		d.waitForRuns()
		if inv, _ := d.record("t-1", "run-2"); inv.GetStatus() != statusSucceeded {
			t.Fatalf("later run's record = %+v, want SUCCEEDED", inv)
		}
	})
}

// sessionLog records, in order, the events appended to sessions and the
// sessions deleted.
type sessionLog struct {
	adksession.Service
	log *orderLog
}

func (s sessionLog) AppendEvent(ctx context.Context, sess adksession.Session, evt *adksession.Event) error {
	s.log.add("append")
	return s.Service.AppendEvent(ctx, sess, evt)
}

func (s sessionLog) Delete(ctx context.Context, req *adksession.DeleteRequest) error {
	err := s.Service.Delete(ctx, req)
	s.log.add("delete")
	return err
}

// recordLog records, in the same order, every status an Invocation record is
// saved with.
type recordLog struct {
	invocation.Repository
	log *orderLog
}

func (r recordLog) Save(ctx context.Context, inv *agentsv1.Invocation) error {
	r.log.add("record " + inv.GetStatus().String())
	return r.Repository.Save(ctx, inv)
}

// aguiThreadDeleter is DeleteSession over the harness's stores, deleting
// AG-UI threads through handler.
func (d *detachHarness) aguiThreadDeleter(handler *AGUIHandler) func() error {
	sessions := application.NewSessionServiceServer()
	sessions.SetSessionService(d.sessions)
	sessions.SetInvocationRepo(d.invocations)
	sessions.SetAGUIThreads(handler)
	return func() error {
		_, err := sessions.DeleteSession(memberContext("u1"), connect.NewRequest(&agentsv1.DeleteSessionRequest{
			AppName: aguiAppName, UserId: "u1", SessionId: aguiSessionPrefix + "t-1",
		}))
		return err
	}
}

// Deleting a thread whose run goes on on another instance stops the run
// first: by the time the session is deleted the run has ended CANCELLED,
// recorded and all, and nothing is written to the thread afterwards. The run
// state goes with the thread.
func TestDeleteSession_StopsTheThreadsRunFirstAcrossInstances(t *testing.T) {
	acrossPods(t, func(t *testing.T, pods func(string) podStores) {
		order := &orderLog{}
		d := newDetachHarness(t, onPod(pods("pod-a")), func(d *detachHarness) {
			d.sessions = sessionLog{Service: d.sessions, log: order}
			d.invocations = newStatusLog(recordLog{Repository: d.invocations.Repository, log: order})
		})
		b := d.peer(pods("pod-b"))
		deleteThread := d.aguiThreadDeleter(b.handler)
		model := d.gate("card-model", "never delivered")

		post := d.start("carder", detachBody("t-1", "run-1", "hi"))
		model.waitStarted(t)
		if err := deleteThread(); err != nil {
			t.Fatalf("DeleteSession: %v", err)
		}

		if end := runEnd(t, post.wait(t).Body.String()); !stoppedEnd(end) {
			t.Fatalf("the stream ended with %v, want the stop", end)
		}
		if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != statusCancelled {
			t.Fatalf("record = %+v, want CANCELLED", inv)
		}
		if _, ok := d.threadSession("t-1"); ok {
			t.Fatal("the session survived its delete")
		}
		if st, ok := d.runState(detachThread); ok {
			t.Fatalf("run state after the delete = %+v", st)
		}
		if !threadFree(t, pods("pod-c").guard, detachThread) {
			t.Fatal("the delete kept the thread's lease")
		}

		model.open()
		d.waitForRuns()
		got := order.list()
		deleted := slices.Index(got, "delete")
		cancelled := slices.Index(got, "record "+statusCancelled.String())
		if deleted < 0 || cancelled < 0 || cancelled > deleted || slices.Contains(got[deleted+1:], "append") {
			t.Fatalf("order = %v; want the run CANCELLED before the delete and nothing appended after it", got)
		}
	})
}

// A delete that cannot take the thread's lease in time — here under a run
// without the opt-in, which no Stop reaches — fails with a retryable error
// and deletes nothing. Once the run ended, the delete goes through.
func TestDeleteSession_FailsRetryablyWhileARunHoldsTheThread(t *testing.T) {
	d := newDetachHarness(t)
	d.handler.deleteWait = 150 * time.Millisecond
	deleteThread := d.aguiThreadDeleter(d.handler)
	model := d.gate("card-model", "still running")
	post := d.start("carder", minimalAGUIBody("t-1", "hi"))
	model.waitStarted(t)

	if err := deleteThread(); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("DeleteSession under a live run = %v, want unavailable", err)
	}
	if _, ok := d.threadSession("t-1"); !ok {
		t.Fatal("the session was deleted under a live run")
	}
	if _, ok := d.runState(detachThread); !ok {
		t.Fatal("the run state was dropped under a live run")
	}
	if all := d.allRecords(); len(all) != 1 || all[0].GetInput() != "hi" {
		t.Fatalf("records = %+v, want the run's own, unredacted", all)
	}

	model.open()
	if end := runEnd(t, post.wait(t).Body.String()); end["type"] != "RUN_FINISHED" {
		t.Fatalf("the run ended with %v, want it to finish", end)
	}
	if err := deleteThread(); err != nil {
		t.Fatalf("DeleteSession once the run ended: %v", err)
	}
	if _, ok := d.threadSession("t-1"); ok {
		t.Fatal("the session survived its delete")
	}
	if all := d.allRecords(); len(all) != 1 || all[0].GetInput() != "" {
		t.Fatalf("records = %+v, want the run's record redacted", all)
	}
}
