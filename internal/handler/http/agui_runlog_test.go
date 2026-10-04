package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	aguisse "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/encoding/sse"
	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/runtime/runlog"
	"go.orx.me/apps/butter/internal/runtime/runstate"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// The Run Log and attaching to a run (#404, ADR-0016 decisions 5–7), through
// the HTTP contract with a real runner, a real session store and the fake
// model. A Detached Run's events go to its Run Log; its POST response and
// every GET …/threads/:thread_id/run replay the log from RUN_STARTED and
// follow it to the run's end, on any instance.

// tuneRunLog changes the handler's Run Log limits.
func (h *AGUIHandler) tuneRunLog(tune func(*aguiRunLogLimits)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	tune(&h.logLimits)
}

// attachRun attaches to the thread's run through router, as the dashboard
// does after a reload, and returns the response once its stream ended.
func attachRun(t *testing.T, router http.Handler, agentID, threadID string, opts ...a2uiOpt) *httptest.ResponseRecorder {
	t.Helper()
	var ro a2uiRequest
	for _, o := range opts {
		o(&ro)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/agui/"+agentID+"/threads/"+threadID+"/run", nil)
	ro.apply(req)
	w := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		router.ServeHTTP(w, req)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the attached stream never ended")
	}
	return w
}

// liveStream is an SSE response read as it arrives.
type liveStream struct {
	status int
	mu     sync.Mutex
	body   bytes.Buffer
	done   chan struct{}
}

// openStream sends GET path through srv and reads the body as it arrives.
func openStream(t *testing.T, srv *httptest.Server, path string) *liveStream {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	s := &liveStream{status: resp.StatusCode, done: make(chan struct{})}
	go func() {
		defer close(s.done)
		defer func() { _ = resp.Body.Close() }()
		buf := make([]byte, 32<<10)
		for {
			n, err := resp.Body.Read(buf)
			s.mu.Lock()
			s.body.Write(buf[:n])
			s.mu.Unlock()
			if err != nil {
				return
			}
		}
	}()
	return s
}

func (s *liveStream) text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.body.String()
}

// wait returns the whole body once the stream ended.
func (s *liveStream) wait(t *testing.T) string {
	t.Helper()
	select {
	case <-s.done:
	case <-time.After(10 * time.Second):
		t.Fatalf("the stream never ended; so far:\n%s", s.text())
	}
	return s.text()
}

// dataFrames lists the events of a stream as they were sent, in order:
// heartbeats and frame IDs left out.
func dataFrames(body string) []string {
	var out []string
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "data: ") {
			out = append(out, line)
		}
	}
	return out
}

// fallbackReason is the reason of the fallback marker a stream ended with;
// ok is false when it ended otherwise.
func fallbackReason(t *testing.T, body string) (reason string, ok bool) {
	t.Helper()
	end := runEnd(t, body)
	if end["type"] != "CUSTOM" || end["name"] != aguiFallbackEvent {
		return "", false
	}
	v, _ := end["value"].(map[string]any)
	reason, _ = v["reason"].(string)
	return reason, true
}

// requireWholeRun checks that a stream carries a whole run: RUN_STARTED under
// runID first, RUN_FINISHED last.
func requireWholeRun(t *testing.T, label, body, runID string) {
	t.Helper()
	events := sseEvents(t, body)
	if len(events) < 2 || events[0]["type"] != "RUN_STARTED" || events[0]["runId"] != runID ||
		events[len(events)-1]["type"] != "RUN_FINISHED" {
		t.Fatalf("%s: events = %v, want RUN_STARTED (%s) … RUN_FINISHED\n%s", label, eventTypes(t, body), runID, body)
	}
}

// The issue's headline. A client attaches through instance B to a run on
// instance A, mid-run, and receives the whole run in order, from RUN_STARTED
// under the original runId to RUN_FINISHED: the very frames A's POST
// response carried. Once the run ended, attaching again replays it whole.
func TestAGUIAttach_FollowsARunOnAnotherInstance(t *testing.T) {
	acrossPods(t, func(t *testing.T, pods func(string) podStores) {
		d := newDetachHarness(t, onPod(pods("pod-a")))
		b := d.peer(pods("pod-b"))
		srvB := httptest.NewServer(b.router)
		defer srvB.Close()
		model := d.gate("card-model", "Here is the summary.")

		post := d.start("carder", detachBody("t-1", "run-1", "Summarize the deploy"))
		model.waitStarted(t)
		watch := openStream(t, srvB, "/api/agui/carder/threads/t-1/run")
		if watch.status != http.StatusOK {
			t.Fatalf("attach through B: status = %d", watch.status)
		}
		// B replays what the run sent so far, then follows it live.
		eventually(t, "B replays the run's start", func() bool {
			return strings.Contains(watch.text(), `"type":"RUN_STARTED"`)
		})
		if strings.Contains(watch.text(), `"type":"RUN_FINISHED"`) {
			t.Fatal("setup: the run finished before the gate opened")
		}
		model.open()

		live := watch.wait(t)
		postBody := post.wait(t).Body.String()
		requireWholeRun(t, "A's POST", postBody, "run-1")
		requireWholeRun(t, "B's attach", live, "run-1")
		if got, want := dataFrames(live), dataFrames(postBody); !slices.Equal(got, want) {
			t.Fatalf("B's attach carried\n%s\nA's POST carried\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
		if !strings.Contains(live, `"delta":"Here is the summary."`) {
			t.Fatalf("the attach missed the reply:\n%s", live)
		}
		d.waitForRuns()

		// Within the retention, a late attach replays the whole run.
		late := attachRun(t, b.router, "carder", "t-1")
		if late.Code != http.StatusOK {
			t.Fatalf("late attach: status = %d, body = %s", late.Code, late.Body.String())
		}
		if got, want := dataFrames(late.Body.String()), dataFrames(postBody); !slices.Equal(got, want) {
			t.Fatalf("the late attach carried\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	})
}

// A run's log outlives the run only for its retention. Attaching after the
// run within it returns the whole run; after it, or on a thread nobody ran,
// it answers 204.
func TestAGUIAttach_AfterTheRunReplaysItForItsRetention(t *testing.T) {
	d := newDetachHarness(t)
	d.handler.tuneRunLog(func(l *aguiRunLogLimits) { l.retention = 500 * time.Millisecond })
	if w := attachRun(t, d.router, "carder", "t-never"); w.Code != http.StatusNoContent || w.Body.Len() != 0 {
		t.Fatalf("a thread nobody ran: status = %d, body = %q; want an empty 204", w.Code, w.Body.String())
	}

	d.answer("card-model", "Done.")
	post := d.post("carder", detachBody("t-1", "run-1", "hi"))
	requireWholeRun(t, "the POST", post.Body.String(), "run-1")
	d.waitForRuns()

	w := attachRun(t, d.router, "carder", "t-1")
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("attach after the run: status = %d, headers %v", w.Code, w.Header())
	}
	if got, want := dataFrames(w.Body.String()), dataFrames(post.Body.String()); !slices.Equal(got, want) {
		t.Fatalf("the attach carried\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	eventually(t, "the retention passes", func() bool {
		return attachRun(t, d.router, "carder", "t-1").Code == http.StatusNoContent
	})
	if st, ok := d.runState(detachThread); ok {
		t.Fatalf("the ended run state outlived its log: %+v", st)
	}
}

// A run without the opt-in has no log: attaching answers 204 while it runs
// and after it.
func TestAGUIAttach_ARunWithoutTheOptInHasNoLog(t *testing.T) {
	d := newDetachHarness(t)
	model := d.gate("card-model", "in its request")
	body := minimalAGUIBody("t-1", "hi")
	body["runId"] = "run-1"
	post := d.start("carder", body)
	model.waitStarted(t)
	if w := attachRun(t, d.router, "carder", "t-1"); w.Code != http.StatusNoContent {
		t.Fatalf("during the run: status = %d, body = %s; want 204", w.Code, w.Body.String())
	}
	model.open()
	post.wait(t)
	if w := attachRun(t, d.router, "carder", "t-1"); w.Code != http.StatusNoContent {
		t.Fatalf("after the run: status = %d, body = %s; want 204", w.Code, w.Body.String())
	}
}

// Attaching answers to the checks of the thread reads: it reaches only the
// caller's own thread, bound to the request's workspace and the route's
// agent. Anyone else reaches nothing, and learns nothing.
func TestAGUIAttach_ReachesOnlyTheCallersBoundThread(t *testing.T) {
	d := newDetachHarness(t)
	d.router = d.build([]agentsv1.Agent{cardAgent(), plainAgent("plain", "plain-model")}, []string{"card-model", "plain-model"})
	d.handler.SetSessionGuard(d.lease)
	d.handler.SetInvocationRepo(d.invocations)
	d.handler.SetRunStateStore(d.runStates)
	d.handler.SetRunLogStore(d.runLogs)
	d.answer("card-model", "mine")
	if w := d.post("carder", detachBody("t-1", "run-1", "hi")); w.Code != http.StatusOK {
		t.Fatalf("run: status = %d", w.Code)
	}
	d.waitForRuns()

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
		w := attachRun(t, d.router, tc.agent, "t-1", tc.opts...)
		if w.Code != tc.want || streamed(w.Body.String()) {
			t.Errorf("%s: status = %d, body = %s; want %d", tc.name, w.Code, w.Body.String(), tc.want)
		}
	}
	if w := attachRun(t, d.router, "carder", "t-1"); w.Code != http.StatusOK || !streamed(w.Body.String()) {
		t.Fatalf("the caller's own attach: status = %d, body = %s", w.Code, w.Body.String())
	}
}

// Deleting a thread drops its run's log with its run state: attaching then
// answers 204, and the log is gone.
func TestAGUIAttach_AfterADeleteAnswersNoContent(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "Done.")
	if w := d.post("carder", detachBody("t-1", "run-1", "hi")); w.Code != http.StatusOK {
		t.Fatalf("run: status = %d", w.Code)
	}
	d.waitForRuns()
	if w := attachRun(t, d.router, "carder", "t-1"); w.Code != http.StatusOK {
		t.Fatalf("attach before the delete: status = %d", w.Code)
	}
	inv, _ := d.record("t-1", "run-1")

	if err := d.aguiThreadDeleter(d.handler)(); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if w := attachRun(t, d.router, "carder", "t-1"); w.Code != http.StatusNoContent || w.Body.Len() != 0 {
		t.Fatalf("attach after the delete: status = %d, body = %q; want an empty 204", w.Code, w.Body.String())
	}
	if _, err := d.runLogs.Read(context.Background(), aguiRunLogKey(detachThread, inv.GetId()), "", 1); !errors.Is(err, runlog.ErrGone) {
		t.Fatalf("the run's log after the delete: %v; want it gone", err)
	}
	if st, ok := d.runState(detachThread); ok {
		t.Fatalf("run state after the delete = %+v", st)
	}
}

// A run whose events outgrow its log still finishes, its reply stored and its
// record SUCCEEDED. Its observers, the POST response and a late attach alike,
// follow it as far as the log goes, then end with the fallback marker.
func TestAGUIRunLog_ARunOverTheCapStillFinishes(t *testing.T) {
	reply := strings.Repeat("A long reply. ", 60)
	for name, tune := range map[string]func(*aguiRunLogLimits){
		"entries": func(l *aguiRunLogLimits) { l.maxEntries = 4 },
		"bytes":   func(l *aguiRunLogLimits) { l.maxBytes = 600 },
	} {
		t.Run(name, func(t *testing.T) {
			d := newDetachHarness(t)
			d.handler.tuneRunLog(tune)
			d.answer("card-model", reply)

			w := d.post("carder", detachBody("t-1", "run-1", "hi"))
			body := w.Body.String()
			if reason, ok := fallbackReason(t, body); w.Code != http.StatusOK || !ok || reason != aguiFallbackTruncated {
				t.Fatalf("the POST: status = %d, ended with %v; want the fallback marker\n%s", w.Code, runEnd(t, body), body)
			}
			if types := eventTypes(t, body); types[0] != "RUN_STARTED" || slices.Contains(types, "RUN_FINISHED") {
				t.Fatalf("the POST carried %v", types)
			}
			d.waitForRuns()

			if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != statusSucceeded {
				t.Fatalf("record = %+v, want SUCCEEDED", inv)
			}
			if sess, _ := d.threadSession("t-1"); !agentSaid(sess, "A long reply.") {
				t.Fatal("the run's reply was not stored")
			}
			if d.lease.isHeld(detachThread) {
				t.Fatal("the run kept its lease")
			}
			late := attachRun(t, d.router, "carder", "t-1")
			if reason, ok := fallbackReason(t, late.Body.String()); !ok || reason != aguiFallbackTruncated {
				t.Fatalf("the late attach ended with %v, want the fallback marker", runEnd(t, late.Body.String()))
			}
			// Each observer writes its own marker; what precedes it is the
			// log's.
			got, want := dataFrames(late.Body.String()), dataFrames(body)
			if !slices.Equal(got[:len(got)-1], want[:len(want)-1]) {
				t.Fatalf("the late attach carried\n%s\nthe POST\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
			}
			// The reads have the whole run.
			if _, read := d.history("carder", "t-1"); read.Running != nil || len(read.Messages) != 2 ||
				textOf(t, read.Messages[1].Content) != reply {
				t.Fatalf("history = %+v, want the whole run", read)
			}
		})
	}
}

// heldAppends is a log store whose appends after the first few wait until
// released, as a Redis that has stalled would.
type heldAppends struct {
	runlog.Store
	free     int32
	appends  atomic.Int32
	release  chan struct{}
	once     sync.Once
	released atomic.Bool
}

func newHeldAppends(store runlog.Store, free int32) *heldAppends {
	return &heldAppends{Store: store, free: free, release: make(chan struct{})}
}

func (s *heldAppends) Append(ctx context.Context, key string, at int, entries []runlog.Entry, ttl time.Duration) error {
	if s.appends.Add(1) > s.free {
		select {
		case <-s.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return s.Store.Append(ctx, key, at, entries, ttl)
}

func (s *heldAppends) open() {
	s.once.Do(func() {
		s.released.Store(true)
		close(s.release)
	})
}

// A slow store never slows the run. The writer only queues, bounded: while
// the store holds its appends, the run's turn goes on and stores its reply,
// the run ends SUCCEEDED and lets its thread go after a bounded wait for its
// last events, and its observer ends with the fallback marker.
func TestAGUIRunLog_ASlowStoreDoesNotSlowTheRun(t *testing.T) {
	acrossPods(t, func(t *testing.T, pods func(string) podStores) {
		stores := pods("pod-a")
		slow := newHeldAppends(stores.logs, 1)
		t.Cleanup(slow.open)
		stores.logs = slow
		d := newDetachHarness(t, onPod(stores))
		d.handler.tuneRunLog(func(l *aguiRunLogLimits) {
			l.maxPending = 256
			l.closeWait = 300 * time.Millisecond
			l.check = 20 * time.Millisecond
			l.endGrace = 200 * time.Millisecond
		})
		reply := strings.Repeat("A reply the store cannot keep up with. ", 20)
		d.answer("card-model", reply)

		post := d.start("carder", detachBody("t-1", "run-1", "hi"))
		eventually(t, "the run stores its reply while the store holds its appends", func() bool {
			sess, _ := d.threadSession("t-1")
			return agentSaid(sess, "A reply the store cannot keep up with.")
		})
		if slow.released.Load() {
			t.Fatal("setup: the store was released")
		}
		w := post.wait(t)
		if reason, ok := fallbackReason(t, w.Body.String()); !ok || reason != aguiFallbackTruncated {
			t.Fatalf("the POST ended with %v, want the fallback marker\n%s", runEnd(t, w.Body.String()), w.Body.String())
		}
		if types := eventTypes(t, w.Body.String()); types[0] != "RUN_STARTED" {
			t.Fatalf("the POST carried %v, want it to open with RUN_STARTED", types)
		}
		d.waitForRuns()
		if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != statusSucceeded {
			t.Fatalf("record = %+v, want SUCCEEDED", inv)
		}
		if d.lease.isHeld(detachThread) {
			t.Fatal("the run kept its lease")
		}
		// A late observer finds the log stopped short of the run's end.
		slow.open()
		late := attachRun(t, d.router, "carder", "t-1")
		if reason, ok := fallbackReason(t, late.Body.String()); !ok || reason != aguiFallbackTruncated {
			t.Fatalf("the late attach ended with %v, want the fallback marker", runEnd(t, late.Body.String()))
		}
	})
}

// A Detached Run always opens with a STATE_SNAPSHOT, even when the client's
// mirror matches the session, so a replay restores the state on its own.
func TestAGUIRunLog_ADetachedRunOpensWithAStateSnapshot(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "Done.")
	for _, detach := range []bool{false, true} {
		body := minimalAGUIBody("t-1", "hi")
		body["runId"] = "run-plain"
		if detach {
			body = detachBody("t-1", "run-detached", "hi")
		}
		types := eventTypes(t, d.post("carder", body).Body.String())
		if got := slices.Contains(types, "STATE_SNAPSHOT"); got != detach {
			t.Fatalf("detach %v: events %v; want a STATE_SNAPSHOT only for the Detached Run", detach, types)
		}
		if detach && types[1] != "STATE_SNAPSHOT" {
			t.Fatalf("events %v, want the snapshot right after RUN_STARTED", types)
		}
		d.waitForRuns()
	}
}

// A run longer than its log's TTL keeps its log: the writer renews it while
// the run lives.
func TestAGUIRunLog_ARunKeepsItsLogWhileItLives(t *testing.T) {
	d := newDetachHarness(t)
	d.handler.tuneRunLog(func(l *aguiRunLogLimits) { l.ttl = 150 * time.Millisecond })
	model := d.gate("card-model", "Late, but whole.")
	post := d.start("carder", detachBody("t-1", "run-1", "hi"))
	model.waitStarted(t)
	time.Sleep(600 * time.Millisecond)
	srv := httptest.NewServer(d.router)
	defer srv.Close()
	watch := openStream(t, srv, "/api/agui/carder/threads/t-1/run")
	if watch.status != http.StatusOK {
		t.Fatalf("attach after four TTLs: status = %d", watch.status)
	}
	model.open()
	requireWholeRun(t, "the attach", watch.wait(t), "run-1")
	post.wait(t)
}

// An observer that cannot follow its run to the end ends its stream with the
// fallback marker, which tells the client to read the thread instead: when
// the run's server died and its run state lapsed, when the log lapsed, and
// when the run ended without its end reaching the log.
func TestAGUIAttach_ARunThatCannotBeFollowedEndsWithTheFallback(t *testing.T) {
	for _, tc := range []struct {
		name   string
		vanish func(d *detachHarness, logKey string) error
		reason string
	}{
		{"its server died", func(d *detachHarness, _ string) error {
			return d.runStates.Drop(context.Background(), detachThread)
		}, aguiFallbackLost},
		{"its log lapsed", func(d *detachHarness, logKey string) error {
			return d.runLogs.Drop(context.Background(), logKey)
		}, aguiFallbackExpired},
		{"it ended without its end in the log", func(d *detachHarness, _ string) error {
			return d.runStates.End(context.Background(), detachThread, "inv-x", time.Minute)
		}, aguiFallbackTruncated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := newDetachHarness(t)
			d.handler.heartbeat = 30 * time.Millisecond
			d.handler.tuneRunLog(func(l *aguiRunLogLimits) {
				l.check = 20 * time.Millisecond
				l.endGrace = 150 * time.Millisecond
			})
			// A run in flight on another server: its thread, its run state,
			// and its log so far.
			ctx := wsctx.WithID(context.Background(), "ws-a")
			binding := a2ui.Binding{Principal: "u1", WorkspaceID: "ws-a", AgentID: "carder", ThreadID: "t-1"}
			if _, err := d.sessions.Create(ctx, &adksession.CreateRequest{
				AppName: aguiAppName, UserID: "u1", SessionID: "agui-t-1",
				State: map[string]any{a2ui.BindingKey: binding.StateValue()},
			}); err != nil {
				t.Fatalf("create the thread: %v", err)
			}
			logKey := aguiRunLogKey(detachThread, "inv-x")
			if err := d.runLogs.Open(ctx, logKey, time.Minute); err != nil {
				t.Fatalf("Open: %v", err)
			}
			if err := d.runStates.Begin(ctx, detachThread, runstate.State{RunID: "run-x", InvocationID: "inv-x", Detached: true}); err != nil {
				t.Fatalf("Begin: %v", err)
			}
			enc := aguisse.NewSSEWriter()
			var entries []runlog.Entry
			for _, ev := range []aguievents.Event{
				aguievents.NewRunStartedEvent("t-1", "run-x"),
				aguievents.NewTextMessageStartEvent("m-1", aguievents.WithRole("assistant")),
				aguievents.NewTextMessageContentEvent("m-1", "Half a"),
			} {
				frame, err := aguiSSEFrame(enc, ev)
				if err != nil {
					t.Fatalf("encode: %v", err)
				}
				entries = append(entries, runlog.Entry{Kind: runlog.KindEvent, Frame: frame})
			}
			if err := d.runLogs.Append(ctx, logKey, 0, entries, time.Minute); err != nil {
				t.Fatalf("Append: %v", err)
			}

			srv := httptest.NewServer(d.router)
			defer srv.Close()
			watch := openStream(t, srv, "/api/agui/carder/threads/t-1/run")
			if watch.status != http.StatusOK {
				t.Fatalf("attach: status = %d", watch.status)
			}
			eventually(t, "the attach replays the run so far", func() bool {
				return strings.Contains(watch.text(), `"delta":"Half a"`)
			})
			// The run is quiet for a while, then vanishes.
			time.Sleep(100 * time.Millisecond)
			if err := tc.vanish(d, logKey); err != nil {
				t.Fatalf("vanish: %v", err)
			}
			body := watch.wait(t)
			if reason, ok := fallbackReason(t, body); !ok || reason != tc.reason {
				t.Fatalf("the stream ended with %v, want the fallback marker for %q\n%s", runEnd(t, body), tc.reason, body)
			}
			events := sseEvents(t, body)
			value, _ := events[len(events)-1]["value"].(map[string]any)
			if value["threadId"] != "t-1" || value["runId"] != "run-x" {
				t.Fatalf("the marker names %v, want run-x of t-1", value)
			}
			if types := eventTypes(t, body); !slices.Equal(types, []string{"RUN_STARTED", "TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT", "CUSTOM"}) {
				t.Fatalf("events = %v", types)
			}
			if !strings.Contains(body, string(aguiHeartbeat)) {
				t.Fatalf("no heartbeat while the run was quiet:\n%s", body)
			}
		})
	}
}

// failingReads is a log store whose reads fail once told to, as when its
// Redis went away.
type failingReads struct {
	runlog.Store
	fail  atomic.Bool
	reads atomic.Int32
}

func (s *failingReads) Read(ctx context.Context, key, after string, count int) ([]runlog.Record, error) {
	if s.fail.Load() {
		s.reads.Add(1)
		return nil, errors.New("redis: connection refused")
	}
	return s.Store.Read(ctx, key, after, count)
}

// An observer whose log cannot be read any more retries at a measured pace,
// and once the reads have failed for its grace, ends with the fallback
// marker. (Both of the run's observers read the failing store here: its
// POST response and the attach.)
func TestAGUIAttach_AnUnreadableLogEndsWithTheFallback(t *testing.T) {
	reads := &failingReads{Store: runlog.NewMemory()}
	d := newDetachHarness(t, func(d *detachHarness) { d.runLogs = reads })
	d.handler.tuneRunLog(func(l *aguiRunLogLimits) {
		l.check = 300 * time.Millisecond
		l.endGrace = 600 * time.Millisecond
	})
	model := d.gate("card-model", "never seen")
	post := d.start("carder", detachBody("t-1", "run-1", "hi"))
	model.waitStarted(t)
	srv := httptest.NewServer(d.router)
	defer srv.Close()
	watch := openStream(t, srv, "/api/agui/carder/threads/t-1/run")
	eventually(t, "the attach replays the run's start", func() bool {
		return strings.Contains(watch.text(), `"type":"RUN_STARTED"`)
	})

	reads.fail.Store(true)
	body := watch.wait(t)
	if reason, ok := fallbackReason(t, body); !ok || reason != aguiFallbackLost {
		t.Fatalf("the stream ended with %v, want the fallback marker for a lost run\n%s", runEnd(t, body), body)
	}
	if n := reads.reads.Load(); n > 12 {
		t.Fatalf("the observers read %d times while the store was down; want them paced", n)
	}
	reads.fail.Store(false)
	model.open()
	post.wait(t)
}

// Once a Detached Run ended, its run state is kept, ended, as long as its
// log; reads take it for a run that is not running.
func TestAGUIRead_AnEndedRunIsNotRunning(t *testing.T) {
	d := newDetachHarness(t)
	d.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Deployed.")
	if w := d.post("carder", liveBody("t-1", "run-1", "deploy", true)); w.Code != http.StatusOK {
		t.Fatalf("run: status = %d", w.Code)
	}
	d.waitForRuns()
	if st, ok := d.runState(detachThread); !ok || !st.Ended {
		t.Fatalf("run state = %+v, %v; want it kept, ended", st, ok)
	}
	code, read := d.history("carder", "t-1")
	if code != http.StatusOK || read.Running != nil || read.LastRun != nil || len(read.Messages) != 2 || len(read.Surfaces) != 1 {
		t.Fatalf("history: status %d, %+v; want the whole run and no run in flight", code, read)
	}
	code, snap := d.snapshot("carder", "t-1")
	if code != http.StatusOK || snap["running"] != nil || len(snap["surfaces"].([]any)) != 1 {
		t.Fatalf("snapshot: status %d, %+v; want the run's card and no run in flight", code, snap)
	}
	if w := stopRun(t, d.router, "carder", "t-1"); w.Code != http.StatusNoContent {
		t.Fatalf("a Stop of the ended run: status = %d, want 204", w.Code)
	}
}

// The writer joins consecutive text deltas of one message into one entry,
// and keeps every other event, and the order, as the run sent them.
func TestAGUILogWriter_CoalescesConsecutiveTextDeltas(t *testing.T) {
	store := runlog.NewMemory()
	if err := store.Open(t.Context(), "l", time.Minute); err != nil {
		t.Fatalf("Open: %v", err)
	}
	limits := defaultAGUIRunLogLimits()
	limits.linger = time.Second
	w := newAGUILogWriter(store, "l", limits)
	for _, ev := range []aguievents.Event{
		aguievents.NewRunStartedEvent("t-1", "run-1"),
		aguievents.NewTextMessageStartEvent("m-1", aguievents.WithRole("assistant")),
		aguievents.NewTextMessageContentEvent("m-1", "Hel"),
		aguievents.NewTextMessageContentEvent("m-1", "lo, "),
		aguievents.NewTextMessageContentEvent("m-1", "world"),
		aguievents.NewToolCallStartEvent("call-1", "lookup"),
		aguievents.NewToolCallEndEvent("call-1"),
		aguievents.NewTextMessageContentEvent("m-1", "!"),
		aguievents.NewTextMessageContentEvent("m-2", "other"),
		aguievents.NewTextMessageEndEvent("m-1"),
		aguievents.NewRunFinishedEventWithOptions("t-1", "run-1", aguievents.WithSuccessOutcome()),
	} {
		if err := w.emit(ev); err != nil {
			t.Fatalf("emit %s: %v", ev.Type(), err)
		}
	}
	start := time.Now()
	if !w.close() {
		t.Fatal("the log does not hold the run's end")
	}
	if took := time.Since(start); took > 3*time.Second {
		t.Fatalf("closing took %v", took)
	}
	recs, err := store.Read(t.Context(), "l", "", 100)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	var got []string
	for _, rec := range recs {
		var ev map[string]any
		data := strings.TrimSuffix(strings.SplitN(string(rec.Frame), "data: ", 2)[1], "\n\n")
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			t.Fatalf("decode %q: %v", rec.Frame, err)
		}
		entry, _ := ev["type"].(string)
		if delta, ok := ev["delta"].(string); ok {
			entry += " " + delta
		}
		got = append(got, entry)
	}
	want := []string{
		"RUN_STARTED", "TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT Hello, world", "TOOL_CALL_START", "TOOL_CALL_END",
		"TEXT_MESSAGE_CONTENT !", "TEXT_MESSAGE_CONTENT other", "TEXT_MESSAGE_END", "RUN_FINISHED",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("log = %q\nwant  %q", got, want)
	}
	if last := recs[len(recs)-1]; last.Kind != runlog.KindEnd {
		t.Fatalf("last entry kind = %q, want the run's end", last.Kind)
	}
	// After the end the writer takes nothing more.
	if err := w.emit(aguievents.NewRunStartedEvent("t-1", "run-2")); err != nil {
		t.Fatalf("emit after close: %v", err)
	}
	if recs, _ := store.Read(t.Context(), "l", "", 100); len(recs) != len(want) {
		t.Fatalf("the log grew after the run's end: %d entries", len(recs))
	}
}

// lostReplies is a log store whose first appends go through but answer with
// an error, as when Redis applied a write and its reply was lost.
type lostReplies struct {
	runlog.Store
	lose atomic.Int32
}

func (s *lostReplies) Append(ctx context.Context, key string, at int, entries []runlog.Entry, ttl time.Duration) error {
	if err := s.Store.Append(ctx, key, at, entries, ttl); err != nil {
		return err
	}
	if s.lose.Add(-1) >= 0 {
		return errors.New("i/o timeout")
	}
	return nil
}

// An append that is retried after its reply was lost never adds its events
// twice, and new events wait for it rather than join it.
func TestAGUILogWriter_ARetriedAppendNeverDuplicates(t *testing.T) {
	store := &lostReplies{Store: runlog.NewMemory()}
	store.lose.Store(2)
	if err := store.Open(t.Context(), "l", time.Minute); err != nil {
		t.Fatalf("Open: %v", err)
	}
	limits := defaultAGUIRunLogLimits()
	limits.linger = 0
	w := newAGUILogWriter(store, "l", limits)
	events := []aguievents.Event{
		aguievents.NewRunStartedEvent("t-1", "run-1"),
		aguievents.NewStateSnapshotEvent(map[string]any{}),
		aguievents.NewTextMessageStartEvent("m-1", aguievents.WithRole("assistant")),
		aguievents.NewTextMessageContentEvent("m-1", "Hi"),
		aguievents.NewTextMessageEndEvent("m-1"),
		aguievents.NewRunFinishedEventWithOptions("t-1", "run-1", aguievents.WithSuccessOutcome()),
	}
	for _, ev := range events {
		if err := w.emit(ev); err != nil {
			t.Fatalf("emit: %v", err)
		}
	}
	if !w.close() {
		t.Fatal("the log does not hold the run's end")
	}
	recs, err := store.Read(t.Context(), "l", "", 100)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	var types []string
	for _, rec := range recs {
		types = append(types, eventTypes(t, string(rec.Frame))...)
	}
	if want := []string{"RUN_STARTED", "STATE_SNAPSHOT", "TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT", "TEXT_MESSAGE_END", "RUN_FINISHED"}; !slices.Equal(types, want) {
		t.Fatalf("log = %v, want %v", types, want)
	}
}

// A writer whose log is gone lets the run go on: emitting never fails, and
// closing reports that the log does not hold the run's end.
func TestAGUILogWriter_AGoneLogNeverStopsTheRun(t *testing.T) {
	store := runlog.NewMemory()
	if err := store.Open(t.Context(), "l", time.Minute); err != nil {
		t.Fatalf("Open: %v", err)
	}
	limits := defaultAGUIRunLogLimits()
	limits.linger = 0
	w := newAGUILogWriter(store, "l", limits)
	if err := store.Drop(t.Context(), "l"); err != nil {
		t.Fatalf("Drop: %v", err)
	}
	for _, ev := range []aguievents.Event{
		aguievents.NewRunStartedEvent("t-1", "run-1"),
		aguievents.NewRunFinishedEventWithOptions("t-1", "run-1", aguievents.WithSuccessOutcome()),
	} {
		if err := w.emit(ev); err != nil {
			t.Fatalf("emit: %v", err)
		}
	}
	if w.close() {
		t.Fatal("close reported the run's end in a log that is gone")
	}
	if _, err := store.Read(t.Context(), "l", "", 1); !errors.Is(err, runlog.ErrGone) {
		t.Fatalf("Read = %v; the writer brought the log back", err)
	}
}
