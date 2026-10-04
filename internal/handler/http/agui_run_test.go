package http

import (
	"bufio"
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

	"github.com/gin-gonic/gin"
	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/runstate"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Detached Runs (#401, ADR-0016 decisions 1–3 and 6), through the HTTP
// contract with a real runner, a real session store and the fake model. A
// client opts in with forwardedProps.butterRun = {"detach": true}; its run
// then holds its thread's lease itself, owns its Invocation record and keeps
// going when the client leaves. Every run, detached or not, records its run
// state next to the lease.

const (
	detachOwner  = "pod-a"
	detachThread = "u1:agui-t-1"
)

// detachHarness is the A2UI harness with what a Detached Run needs: the
// in-process lease, counted; an Invocation store this process owns records
// in, which the runner records into as well, as in production; and a run
// state store.
type detachHarness struct {
	*a2uiHarness
	lease       *countingGuard
	invocations *statusLog
	runStates   runstate.Store
}

func newDetachHarness(t *testing.T, opts ...func(*detachHarness)) *detachHarness {
	t.Helper()
	d := &detachHarness{
		a2uiHarness: &a2uiHarness{t: t, backend: openaifake.New(t), sessions: newStoreLikeSessions(), guard: &fakeSessionGuard{}},
		lease:       newCountingGuard(sessionguard.NewMemory()),
		invocations: newStatusLog(invocationmemory.New().WithOwner(detachOwner)),
		runStates:   runstate.NewMemory(time.Minute),
	}
	for _, opt := range opts {
		opt(d)
	}
	d.recorder = d.invocations
	d.router = d.build([]agentsv1.Agent{cardAgent()}, []string{"card-model"})
	d.handler.SetSessionGuard(d.lease)
	d.handler.SetInvocationRepo(d.invocations)
	d.handler.SetRunStateStore(d.runStates)
	// Registered before any gate, so it runs after every gate has opened.
	t.Cleanup(d.waitForRuns)
	return d
}

// detachBody is a user turn from a client that asks for a Detached Run.
func detachBody(threadID, runID, text string) map[string]any {
	body := minimalAGUIBody(threadID, text)
	body["runId"] = runID
	body["forwardedProps"] = map[string]any{"butterRun": map[string]any{"detach": true}}
	return body
}

// waitForRuns waits until no Detached Run is in flight.
func (d *detachHarness) waitForRuns() {
	d.t.Helper()
	done := make(chan struct{})
	go func() {
		d.handler.waitForRuns()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		d.t.Fatal("detached runs still in flight after 10s")
	}
}

// record is the AG-UI-owned record of a run, found the way the handler
// checks for a repeated runId.
func (d *detachHarness) record(threadID, runID string) (*agentsv1.Invocation, bool) {
	inv, err := d.invocations.FindByRequestID(context.Background(), "ws-a", "agui:u1:agui-"+threadID+":"+runID)
	if err != nil {
		return nil, false
	}
	return inv, true
}

func (d *detachHarness) allRecords() []*agentsv1.Invocation {
	d.t.Helper()
	all, _, _, err := d.invocations.List(context.Background(), invocation.ListFilter{WorkspaceID: "ws-a"}, 100, "")
	if err != nil {
		d.t.Fatalf("list invocations: %v", err)
	}
	return all
}

func (d *detachHarness) runState(thread string) (runstate.State, bool) {
	d.t.Helper()
	st, ok, err := d.runStates.Get(context.Background(), thread)
	if err != nil {
		d.t.Fatalf("read run state: %v", err)
	}
	return st, ok
}

// pendingPost is a request in flight from a client that can disconnect.
type pendingPost struct {
	w          *httptest.ResponseRecorder
	disconnect context.CancelFunc
	done       chan struct{}
}

func (d *detachHarness) start(agentID string, body map[string]any, opts ...a2uiOpt) *pendingPost {
	d.t.Helper()
	var ro a2uiRequest
	for _, o := range opts {
		o(&ro)
	}
	payload, err := json.Marshal(body)
	if err != nil {
		d.t.Fatalf("marshal: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.t.Cleanup(cancel)
	req := httptest.NewRequest(http.MethodPost, "/api/agui/"+agentID, strings.NewReader(string(payload))).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	ro.apply(req)
	p := &pendingPost{w: httptest.NewRecorder(), disconnect: cancel, done: make(chan struct{})}
	go func() {
		defer close(p.done)
		d.router.ServeHTTP(p.w, req)
	}()
	return p
}

// wait returns the response once the handler is done with the request.
func (p *pendingPost) wait(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	select {
	case <-p.done:
	case <-time.After(10 * time.Second):
		t.Fatal("the request never completed")
	}
	return p.w
}

// modelGate holds every call to a model until it opens, or until the call is
// cancelled. Each call that starts is announced on started.
type modelGate struct {
	started   chan struct{}
	release   chan struct{}
	once      sync.Once
	cancelled atomic.Int32
}

func (d *detachHarness) gate(model, reply string) *modelGate {
	g := &modelGate{started: make(chan struct{}, 16), release: make(chan struct{})}
	d.backend.ScriptCall(model, func(w http.ResponseWriter, r *http.Request, req openaifake.ChatCompletionRequest) {
		g.started <- struct{}{}
		select {
		case <-g.release:
			openaifake.WriteReply(w, req, reply)
		case <-r.Context().Done():
			g.cancelled.Add(1)
		}
	})
	d.t.Cleanup(g.open)
	return g
}

func (g *modelGate) open() { g.once.Do(func() { close(g.release) }) }

func (g *modelGate) waitStarted(t *testing.T) {
	t.Helper()
	select {
	case <-g.started:
	case <-time.After(5 * time.Second):
		t.Fatal("the model was never called")
	}
}

// countingGuard wraps a guard, counting acquisitions and releases and
// telling which threads are held.
type countingGuard struct {
	inner sessionguard.Guard

	mu                 sync.Mutex
	held               map[string]int
	acquired, released int
}

func newCountingGuard(inner sessionguard.Guard) *countingGuard {
	return &countingGuard{inner: inner, held: map[string]int{}}
}

func (g *countingGuard) Acquire(ctx context.Context, key string) (context.Context, func(), bool, error) {
	leaseCtx, release, ok, err := g.inner.Acquire(ctx, key)
	if err != nil || !ok {
		return leaseCtx, release, ok, err
	}
	g.mu.Lock()
	g.acquired++
	g.held[key]++
	g.mu.Unlock()
	var once sync.Once
	return leaseCtx, func() {
		once.Do(func() {
			release()
			g.mu.Lock()
			g.released++
			g.held[key]--
			g.mu.Unlock()
		})
	}, true, nil
}

func (g *countingGuard) isHeld(key string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.held[key] > 0
}

func (g *countingGuard) counts() (acquired, released int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.acquired, g.released
}

// statusLog is an Invocation store that remembers every status each record
// was saved with.
type statusLog struct {
	invocation.Repository
	mu       sync.Mutex
	statuses map[string][]agentsv1.InvocationStatus
}

func newStatusLog(repo invocation.Repository) *statusLog {
	return &statusLog{Repository: repo, statuses: map[string][]agentsv1.InvocationStatus{}}
}

func (s *statusLog) Save(ctx context.Context, inv *agentsv1.Invocation) error {
	s.mu.Lock()
	s.statuses[inv.GetId()] = append(s.statuses[inv.GetId()], inv.GetStatus())
	s.mu.Unlock()
	return s.Repository.Save(ctx, inv)
}

func (s *statusLog) history(id string) []agentsv1.InvocationStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.statuses[id])
}

// agentSaid reports whether an agent-authored event of the session carries
// text.
func agentSaid(sess adksession.Session, text string) bool {
	if sess == nil {
		return false
	}
	events := sess.Events()
	for i := 0; i < events.Len(); i++ {
		ev := events.At(i)
		if ev == nil || ev.Author == "user" || ev.Content == nil {
			continue
		}
		for _, p := range ev.Content.Parts {
			if p != nil && strings.Contains(p.Text, text) {
				return true
			}
		}
	}
	return false
}

// eventually polls cond until it holds or a deadline passes.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting until %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func eventTypes(t *testing.T, body string) []string {
	t.Helper()
	var types []string
	for _, ev := range sseEvents(t, body) {
		typ, _ := ev["type"].(string)
		types = append(types, typ)
	}
	return types
}

// The issue's headline: with the opt-in, the client leaves mid-run and the
// run still completes. Its events are persisted, its record ends SUCCEEDED
// and its lease is released. The work after the turn (here the thread's
// title) still happens, with nobody watching.
func TestAGUIDetached_RunOutlivesItsClient(t *testing.T) {
	titler := &fakeTitler{}
	d := newDetachHarness(t, func(d *detachHarness) { d.titler = titler })
	model := d.gate("card-model", "Here is the summary.")

	post := d.start("carder", detachBody("t-1", "run-1", "Summarize the deploy"))
	model.waitStarted(t)

	// The client goes away mid-run: that detaches its observer, nothing more.
	post.disconnect()
	w := post.wait(t)
	if w.Code != http.StatusOK || !streamed(w.Body.String()) {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if types := eventTypes(t, w.Body.String()); slices.Contains(types, "RUN_FINISHED") || slices.Contains(types, "RUN_ERROR") {
		t.Fatalf("the response ended with the run, not before it: %v", types)
	}

	// The run still holds its thread, and its record says it runs, owned by
	// this process.
	if !d.lease.isHeld(detachThread) {
		t.Fatal("the run gave its lease up with its request")
	}
	inv, ok := d.record("t-1", "run-1")
	if !ok || inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
		t.Fatalf("record mid-run = %+v (found %v), want RUNNING", inv, ok)
	}
	if owners, err := d.invocations.ActiveOwners(context.Background()); err != nil || !slices.Equal(owners, []string{detachOwner}) {
		t.Fatalf("active owners = %v, %v; want the record stamped %s", owners, err, detachOwner)
	}

	model.open()
	d.waitForRuns()

	// The reply was persisted although nobody watched it arrive.
	sess, _ := d.threadSession("t-1")
	if !agentSaid(sess, "Here is the summary.") {
		t.Fatal("the run's reply was not persisted")
	}
	if n := model.cancelled.Load(); n != 0 {
		t.Fatalf("the model call was cancelled %d times", n)
	}

	// The record went QUEUED → RUNNING → SUCCEEDED, and it is the AG-UI
	// path's: the runner recorded nothing of its own for the run.
	inv, _ = d.record("t-1", "run-1")
	if inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED ||
		inv.GetOutput() != "Here is the summary." || inv.GetInput() != "Summarize the deploy" ||
		inv.GetSource() != invocation.SourceAGUIDetached || inv.GetAppName() != aguiAppName ||
		inv.GetUserId() != "u1" || inv.GetSessionId() != "agui-t-1" || inv.GetAgentId() != "carder" ||
		inv.GetFinishedAt() == nil || inv.GetError() != "" {
		t.Fatalf("terminal record = %+v", inv)
	}
	if got, want := d.invocations.history(inv.GetId()), []agentsv1.InvocationStatus{
		agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
	}; !slices.Equal(got, want) {
		t.Fatalf("record statuses = %v, want %v", got, want)
	}
	if all := d.allRecords(); len(all) != 1 {
		t.Fatalf("records = %d, want only the AG-UI path's", len(all))
	}

	// The lease is free and the run state gone.
	if acquired, released := d.lease.counts(); d.lease.isHeld(detachThread) || acquired != released {
		t.Fatalf("lease held = %v, acquired %d, released %d", d.lease.isHeld(detachThread), acquired, released)
	}
	if st, ok := d.runState(detachThread); ok {
		t.Fatalf("run state after the run = %+v", st)
	}

	// The thread was titled after the run, with its client long gone.
	d.handler.titles.Wait()
	if calls := titler.recorded(); len(calls) != 1 || calls[0].session != "agui/u1/agui-t-1" {
		t.Fatalf("title calls = %+v, want one for the thread", calls)
	}
}

// Without the opt-in nothing changes: a disconnect still cancels the run,
// which the runner records as before, and frees the thread.
func TestAGUIRun_WithoutOptInADisconnectCancelsTheRun(t *testing.T) {
	d := newDetachHarness(t)
	model := d.gate("card-model", "never delivered")

	post := d.start("carder", minimalAGUIBody("t-1", "hi"))
	model.waitStarted(t)
	post.disconnect()
	post.wait(t)

	eventually(t, "the model call is cancelled", func() bool { return model.cancelled.Load() == 1 })
	sess, _ := d.threadSession("t-1")
	if agentSaid(sess, "never delivered") {
		t.Fatal("a cancelled run's reply was persisted")
	}
	if d.lease.isHeld(detachThread) {
		t.Fatal("the lease outlived the request")
	}
	if _, ok := d.runState(detachThread); ok {
		t.Fatal("the run state outlived the run")
	}
	if _, ok := d.record("t-1", "run-1"); ok {
		t.Fatal("a run without the opt-in got an AG-UI-owned record")
	}
	all := d.allRecords()
	if len(all) != 1 || all[0].GetSource() == invocation.SourceAGUIDetached ||
		all[0].GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("records = %+v, want the runner's own, FAILED", all)
	}
}

// While a Detached Run holds its thread, a second POST — detached or not —
// answers 409 before its stream opens, also after the first client left.
func TestAGUIDetached_SecondPostWhileRunningIsConflict(t *testing.T) {
	d := newDetachHarness(t)
	model := d.gate("card-model", "done")

	first := d.start("carder", detachBody("t-1", "run-1", "first"))
	model.waitStarted(t)
	first.disconnect()
	first.wait(t)

	for name, body := range map[string]map[string]any{
		"detached":     detachBody("t-1", "run-2", "second"),
		"not detached": minimalAGUIBody("t-1", "third"),
	} {
		w := d.post("carder", body)
		if w.Code != http.StatusConflict || streamed(w.Body.String()) {
			t.Fatalf("%s: status = %d, body = %s; want a pre-stream 409", name, w.Code, w.Body.String())
		}
	}

	model.open()
	d.waitForRuns()
	w := d.post("carder", detachBody("t-1", "run-3", "fourth"))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
		t.Fatalf("after the run: status = %d, body = %s", w.Code, w.Body.String())
	}
	if n := d.backend.CallCount("card-model"); n != 2 {
		t.Fatalf("model calls = %d, want the first run and the last", n)
	}
}

// A retried POST with a runId the thread already ran starts nothing: the
// Agent never runs twice for one request. The runId is scoped to its thread.
func TestAGUIDetached_RepeatedRunIDDoesNotRunAgain(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "first answer")

	if w := d.post("carder", detachBody("t-1", "run-1", "hello")); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
		t.Fatalf("first run: status = %d, body = %s", w.Code, w.Body.String())
	}
	// A client that stayed sees its stream end only once the run has
	// settled and freed the thread, so its next POST is never a 409.
	if d.lease.isHeld(detachThread) {
		t.Fatal("the stream ended while the run still held its lease")
	}
	if _, ok := d.runState(detachThread); ok {
		t.Fatal("the stream ended while the run state was still recorded")
	}
	if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("the stream ended before the record was terminal: %+v", inv)
	}
	d.waitForRuns()

	w := d.post("carder", detachBody("t-1", "run-1", "hello"))
	if w.Code != http.StatusConflict || streamed(w.Body.String()) {
		t.Fatalf("retry: status = %d, body = %s; want a pre-stream 409", w.Code, w.Body.String())
	}
	var refusal aguiCodedError
	if err := json.Unmarshal(w.Body.Bytes(), &refusal); err != nil || refusal.Code != aguiCodeRunExists {
		t.Fatalf("retry body = %s, want code %s", w.Body.String(), aguiCodeRunExists)
	}
	if n := d.backend.CallCount("card-model"); n != 1 {
		t.Fatalf("model calls = %d, want 1", n)
	}
	if all := d.allRecords(); len(all) != 1 {
		t.Fatalf("records = %d, want 1", len(all))
	}
	if d.lease.isHeld(detachThread) {
		t.Fatal("the refused retry kept the lease")
	}

	if w := d.post("carder", detachBody("t-2", "run-1", "hello")); w.Code != http.StatusOK {
		t.Fatalf("the same runId on another thread: status = %d, body = %s", w.Code, w.Body.String())
	}
}

// Past the maximum run duration a Detached Run is cancelled and ends FAILED,
// releasing its lease.
func TestAGUIDetached_MaxRunDurationEndsTheRunFailed(t *testing.T) {
	d := newDetachHarness(t)
	d.handler.SetMaxRunDuration(150 * time.Millisecond)
	model := d.gate("card-model", "too late")

	w := d.post("carder", detachBody("t-1", "run-1", "hi"))
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, `"type":"RUN_ERROR"`) ||
		!strings.Contains(body, "exceeded the maximum run duration (150ms)") {
		t.Fatalf("status = %d, body = %s", w.Code, body)
	}
	d.waitForRuns()
	inv, _ := d.record("t-1", "run-1")
	if inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED ||
		!strings.Contains(inv.GetError(), "exceeded the maximum run duration (150ms)") {
		t.Fatalf("record = %+v, want FAILED past the maximum duration", inv)
	}
	eventually(t, "the model call is cancelled", func() bool { return model.cancelled.Load() == 1 })
	if d.lease.isHeld(detachThread) {
		t.Fatal("the timed-out run kept its lease")
	}
}

// A graceful shutdown ends the Detached Runs in flight FAILED with a shutdown
// reason, releases their leases and waits for them; new detached runs are
// refused afterwards.
func TestAGUIDetached_ShutdownEndsRunsFailedAndReleasesTheirLeases(t *testing.T) {
	d := newDetachHarness(t)
	model := d.gate("card-model", "too late")
	post := d.start("carder", detachBody("t-1", "run-1", "hi"))
	model.waitStarted(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.handler.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	// Shutdown returned once the run was done with its thread.
	if d.lease.isHeld(detachThread) {
		t.Fatal("Shutdown returned while the run still held its lease")
	}
	inv, _ := d.record("t-1", "run-1")
	if inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED ||
		inv.GetError() != aguiShutdownReason {
		t.Fatalf("record = %+v, want FAILED with the shutdown reason", inv)
	}
	w := post.wait(t)
	if body := w.Body.String(); !strings.Contains(body, `"type":"RUN_ERROR"`) || !strings.Contains(body, "interrupted by a service shutdown") {
		t.Fatalf("the observer was not told why the run ended: %s", body)
	}

	if w := d.post("carder", detachBody("t-2", "run-2", "hi")); w.Code != http.StatusServiceUnavailable || streamed(w.Body.String()) {
		t.Fatalf("after shutdown: status = %d, body = %s; want a pre-stream 503", w.Code, w.Body.String())
	}
}

// Shutdown's wait is bounded: a run that does not finish in time leaves it
// to the stale sweep, and Shutdown returns.
func TestAGUIRuns_ShutdownWaitIsBounded(t *testing.T) {
	runs := newAGUIRuns()
	var cause error
	entry, ok := runs.register("inv-1", func(err error) { cause = err })
	if !ok {
		t.Fatal("register refused before shutdown")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := runs.shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("shutdown = %v, want it to give up at its deadline", err)
	}
	if !errors.Is(cause, errAGUIShutdown) {
		t.Fatalf("the run was cancelled with %v, want the shutdown", cause)
	}
	if claim := runs.claim(entry); !claim.shutdown || claim.stopped {
		t.Fatalf("claim = %+v, want a shutdown", claim)
	}
	runs.done(entry)
	if _, ok := runs.register("inv-2", func(error) {}); ok {
		t.Fatal("a run registered after shutdown")
	}
}

// flakyLease is a lease nobody contends for, whose renewals fail as a Redis
// hiccup would: the first failFirst ones, or every one.
type flakyLease struct {
	mu          sync.Mutex
	failFirst   int
	failForever bool
	renewals    int
	failures    int
}

func (l *flakyLease) Acquire(context.Context) (bool, error) { return true, nil }

func (l *flakyLease) Renew(context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.renewals++
	if l.failForever || l.failures < l.failFirst {
		l.failures++
		return false, errors.New("redis: connection reset by peer")
	}
	return true, nil
}

func (l *flakyLease) Release(context.Context) error { return nil }

func (l *flakyLease) counts() (renewals, failures int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.renewals, l.failures
}

// A renewal that errors does not cancel a run: it is retried until the lease
// would really have lapsed. The run state is renewed with the lease all
// along. Only renewals that keep failing past one TTL end the run.
func TestAGUIDetached_TransientRenewalErrorKeepsTheRun(t *testing.T) {
	const ttl = 150 * time.Millisecond
	withLease := func(lease *flakyLease) func(*detachHarness) {
		return func(d *detachHarness) {
			d.lease = newCountingGuard(sessionguard.NewLeased(detachOwner, ttl,
				func(string, string) sessionguard.Lease { return lease }))
			d.runStates = runstate.NewMemory(ttl)
		}
	}

	t.Run("one failed renewal", func(t *testing.T) {
		lease := &flakyLease{failFirst: 1}
		d := newDetachHarness(t, withLease(lease))
		model := d.gate("card-model", "made it")
		post := d.start("carder", detachBody("t-1", "run-1", "hi"))
		model.waitStarted(t)

		// Well past one TTL, with the first renewal failed and retried.
		eventually(t, "the lease is renewed several times", func() bool {
			renewals, _ := lease.counts()
			return renewals >= 6
		})
		if _, failures := lease.counts(); failures != 1 {
			t.Fatalf("failed renewals = %d, want 1", failures)
		}
		if st, ok := d.runState(detachThread); !ok || st.RunID != "run-1" {
			t.Fatal("the run state lapsed while the run held its lease")
		}
		model.open()
		w := post.wait(t)
		if !strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
			t.Fatalf("body = %s", w.Body.String())
		}
		d.waitForRuns()
		if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			t.Fatalf("record = %+v, want SUCCEEDED", inv)
		}
	})

	t.Run("renewals that keep failing", func(t *testing.T) {
		lease := &flakyLease{failForever: true}
		d := newDetachHarness(t, withLease(lease))
		model := d.gate("card-model", "too late")
		w := d.post("carder", detachBody("t-1", "run-1", "hi"))
		if body := w.Body.String(); !strings.Contains(body, `"type":"RUN_ERROR"`) || !strings.Contains(body, "session lease lost") {
			t.Fatalf("body = %s, want RUN_ERROR for the lost lease", body)
		}
		d.waitForRuns()
		if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED ||
			!strings.Contains(inv.GetError(), "session lease lost") {
			t.Fatalf("record = %+v, want FAILED with the lost lease", inv)
		}
		if renewals, failures := lease.counts(); failures < 2 || renewals != failures {
			t.Fatalf("renewals %d, failures %d: want the failing renewal retried before the run gave up", renewals, failures)
		}
		eventually(t, "the model call is cancelled", func() bool { return model.cancelled.Load() == 1 })
	})
}

// A butterRun declaration that is not {"detach": <bool>} is a 400 before
// anything runs.
func TestAGUIRun_MalformedButterRunIsBadRequest(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "ran")
	for name, butterRun := range map[string]any{
		"not an object":        "detach",
		"a list":               []any{true},
		"no detach":            map[string]any{},
		"detach not a boolean": map[string]any{"detach": "true"},
		"an unknown option":    map[string]any{"detach": true, "replay": true},
	} {
		body := minimalAGUIBody("t-1", "hi")
		body["forwardedProps"] = map[string]any{"butterRun": butterRun}
		w := d.post("carder", body)
		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "butterRun") || streamed(w.Body.String()) {
			t.Errorf("%s: status = %d, body = %s; want a pre-stream 400 naming butterRun", name, w.Code, w.Body.String())
		}
	}
	if n := d.backend.CallCount("card-model"); n != 0 {
		t.Fatalf("model calls = %d, want none", n)
	}

	// An explicit {"detach": false}, or null, is the default run.
	for i, butterRun := range []any{map[string]any{"detach": false}, nil} {
		body := minimalAGUIBody("t-1", "hi")
		body["runId"] = "run-" + string(rune('a'+i))
		body["forwardedProps"] = map[string]any{"butterRun": butterRun}
		if w := d.post("carder", body); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
			t.Fatalf("butterRun %v: status = %d, body = %s", butterRun, w.Code, w.Body.String())
		}
	}
	if _, ok := d.record("t-1", "run-a"); ok {
		t.Fatal("a run that did not detach got an AG-UI-owned record")
	}
}

// Stop and attach are checked against the thread's UI Binding, so a thread
// without one cannot detach its runs: refused before the stream opens. The
// same thread still runs as before without the opt-in.
func TestAGUIDetached_UnboundThreadIsRefused(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "plain text")
	// A session from before A2UI: it carries no binding.
	if _, err := d.sessions.Create(wsctx.WithID(context.Background(), "ws-a"), &adksession.CreateRequest{
		AppName: aguiAppName, UserID: "u1", SessionID: "agui-t-old",
	}); err != nil {
		t.Fatalf("create legacy session: %v", err)
	}

	w := d.post("carder", detachBody("t-old", "run-1", "hi"))
	if w.Code != http.StatusBadRequest || streamed(w.Body.String()) {
		t.Fatalf("status = %d, body = %s; want a pre-stream 400", w.Code, w.Body.String())
	}
	var refusal aguiCodedError
	if err := json.Unmarshal(w.Body.Bytes(), &refusal); err != nil || refusal.Code != aguiCodeThreadUnbound {
		t.Fatalf("body = %s, want code %s", w.Body.String(), aguiCodeThreadUnbound)
	}
	if n := d.backend.CallCount("card-model"); n != 0 {
		t.Fatalf("model calls = %d, want none", n)
	}
	if all := d.allRecords(); len(all) != 0 {
		t.Fatalf("records = %+v, want none", all)
	}
	if d.lease.isHeld("u1:agui-t-old") {
		t.Fatal("the refused run kept its lease")
	}
	if _, ok := d.runState("u1:agui-t-old"); ok {
		t.Fatal("the refused run left a run state")
	}

	if w := d.post("carder", minimalAGUIBody("t-old", "hi")); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
		t.Fatalf("without the opt-in: status = %d, body = %s", w.Code, w.Body.String())
	}
}

// The thread checks of #388 still refuse with 403 before the stream opens,
// before anything is detached or recorded.
func TestAGUIDetached_ThreadChecksRefuseBeforeDetaching(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "mine")
	if w := d.post("carder", minimalAGUIBody("t-1", "hi")); w.Code != http.StatusOK {
		t.Fatalf("setup run: status = %d", w.Code)
	}
	before := len(d.allRecords())

	w := d.post("carder", detachBody("t-1", "run-9", "hi"), asUser("u2"))
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), errThreadUnavailable.Error()) || streamed(w.Body.String()) {
		t.Fatalf("status = %d, body = %s; want a pre-stream 403", w.Code, w.Body.String())
	}
	if after := len(d.allRecords()); after != before {
		t.Fatalf("records %d → %d: the refused run was recorded", before, after)
	}
	if d.lease.isHeld("u2:agui-t-1") {
		t.Fatal("the refused run kept its lease")
	}
	d.handler.runs.mu.Lock()
	inFlight := len(d.handler.runs.entries)
	d.handler.runs.mu.Unlock()
	if inFlight != 0 {
		t.Fatalf("%d detached runs in flight after a refusal", inFlight)
	}
}

// Every run, detached or not, records its run state next to its lease while
// it runs: its runId, its Invocation ID and the session's event count before
// it. It is gone once the run ends.
func TestAGUIRun_RecordsRunStateWhileItRuns(t *testing.T) {
	for _, detached := range []bool{false, true} {
		name := "in its request"
		if detached {
			name = "detached"
		}
		t.Run(name, func(t *testing.T) {
			d := newDetachHarness(t)
			d.answer("card-model", "first reply")
			if w := d.post("carder", minimalAGUIBody("t-1", "first")); w.Code != http.StatusOK {
				t.Fatalf("first run: status = %d", w.Code)
			}
			sess, _ := d.threadSession("t-1")
			before := sess.Events().Len()
			if before == 0 {
				t.Fatal("the first run left no events")
			}

			model := d.gate("card-model", "second reply")
			body := minimalAGUIBody("t-1", "second")
			body["runId"] = "run-2"
			if detached {
				body = detachBody("t-1", "run-2", "second")
			}
			post := d.start("carder", body)
			model.waitStarted(t)

			st, ok := d.runState(detachThread)
			if !ok {
				t.Fatal("no run state while the run runs")
			}
			if st.RunID != "run-2" || st.EventCount != before || st.Detached != detached {
				t.Fatalf("run state = %+v, want run-2 after %d events, detached %v", st, before, detached)
			}
			// The Invocation ID names the run's record, whoever writes it.
			inv, err := d.invocations.Get(context.Background(), "ws-a", st.InvocationID)
			if err != nil || inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
				t.Fatalf("record %s = %+v, %v; want it RUNNING", st.InvocationID, inv, err)
			}
			if (inv.GetSource() == invocation.SourceAGUIDetached) != detached {
				t.Fatalf("record source = %q for a run with detached %v", inv.GetSource(), detached)
			}

			model.open()
			post.wait(t)
			d.waitForRuns()
			if st, ok := d.runState(detachThread); ok {
				t.Fatalf("run state after the run = %+v", st)
			}
		})
	}
}

// A Stop is representable: an accepted one ends the run CANCELLED and tells
// its observers, and one that comes after the run claimed its end finds
// nothing running. (The Stop endpoint itself is #402.)
func TestAGUIDetached_StopEndsTheRunCancelled(t *testing.T) {
	d := newDetachHarness(t)
	model := d.gate("card-model", "too late")
	post := d.start("carder", detachBody("t-1", "run-1", "hi"))
	model.waitStarted(t)

	st, ok := d.runState(detachThread)
	if !ok {
		t.Fatal("no run state")
	}
	if !d.handler.runs.stop(st.InvocationID) {
		t.Fatal("the Stop was not accepted")
	}
	w := post.wait(t)
	if body := w.Body.String(); !strings.Contains(body, `"type":"RUN_ERROR"`) || !strings.Contains(body, "stopped by user") {
		t.Fatalf("body = %s, want RUN_ERROR for the Stop", body)
	}
	d.waitForRuns()
	inv, _ := d.record("t-1", "run-1")
	if inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED {
		t.Fatalf("record = %+v, want CANCELLED", inv)
	}
	if d.lease.isHeld(detachThread) {
		t.Fatal("the stopped run kept its lease")
	}
	if d.handler.runs.stop(st.InvocationID) {
		t.Fatal("a Stop after the run ended was accepted")
	}
}

// An observer learns that a Detached Run ended only once the end is settled:
// when RUN_FINISHED reaches the client, the record is terminal and the run
// state gone, so a read right after it already sees the finished thread.
func TestAGUIDetached_ObserversLearnTheEndOnceItIsRecorded(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "done")
	srv := httptest.NewServer(d.router)
	defer srv.Close()

	payload, err := json.Marshal(detachBody("t-1", "run-1", "hi"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := srv.Client().Post(srv.URL+"/api/agui/carder", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	lines := bufio.NewScanner(resp.Body)
	for lines.Scan() {
		if !strings.Contains(lines.Text(), `"type":"RUN_FINISHED"`) {
			continue
		}
		if inv, _ := d.record("t-1", "run-1"); inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			t.Fatalf("RUN_FINISHED arrived before the record was terminal: %+v", inv)
		}
		if st, ok := d.runState(detachThread); ok {
			t.Fatalf("RUN_FINISHED arrived while the run state was still recorded: %+v", st)
		}
		return
	}
	t.Fatalf("the stream ended without RUN_FINISHED (%v)", lines.Err())
}

// While a Detached Run is quiet, its observer's stream carries SSE comment
// heartbeats, which AG-UI clients skip.
func TestAGUIDetached_ObserverGetsHeartbeats(t *testing.T) {
	d := newDetachHarness(t)
	d.handler.heartbeat = 20 * time.Millisecond
	model := d.gate("card-model", "after a while")
	post := d.start("carder", detachBody("t-1", "run-1", "hi"))
	model.waitStarted(t)
	time.Sleep(150 * time.Millisecond)
	model.open()

	body := post.wait(t).Body.String()
	if !strings.Contains(body, string(aguiHeartbeat)) {
		t.Fatalf("no heartbeat in a quiet stream:\n%s", body)
	}
	types := eventTypes(t, body)
	if len(types) == 0 || types[0] != "RUN_STARTED" || types[len(types)-1] != "RUN_FINISHED" ||
		!strings.Contains(body, `"delta":"after a while"`) {
		t.Fatalf("events = %v\n%s", types, body)
	}
}

// orderLog records, in order, when the lease is taken and when the session
// is read.
type orderLog struct {
	mu      sync.Mutex
	entries []string
}

func (l *orderLog) add(entry string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
}

func (l *orderLog) list() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.entries)
}

type orderedGuard struct {
	sessionguard.Guard
	log *orderLog
}

func (g orderedGuard) Acquire(ctx context.Context, key string) (context.Context, func(), bool, error) {
	g.log.add("acquire")
	return g.Guard.Acquire(ctx, key)
}

type orderedSessions struct {
	adksession.Service
	log *orderLog
}

func (s orderedSessions) Get(ctx context.Context, req *adksession.GetRequest) (*adksession.GetResponse, error) {
	s.log.add("get")
	return s.Service.Get(ctx, req)
}

// Every check that reads the session — client tool results against the
// pending calls, the shared-state baseline, the thread checks — runs under
// the lease the run then keeps: a busy thread is refused before its session
// is read at all.
func TestAGUIRun_SessionChecksRunUnderTheLease(t *testing.T) {
	setup := func(guard sessionguard.Guard, log *orderLog) *gin.Engine {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), "test-workspace"))
			c.Next()
		})
		h := NewAGUIHandler(aguiEnabledRepo())
		h.SetRunnerService(&mockRunner{runResult: "done"})
		h.SetSessionService(orderedSessions{Service: &fakeSessionService{sess: pendingToolCallSession()}, log: log})
		h.SetSessionGuard(orderedGuard{Guard: guard, log: log})
		h.Register(r)
		return r
	}
	body := withAGUIField(toolResultBody("t-1", "call-1", `{"approved":true}`), "state", map[string]any{"draft": "mine"})

	busy := &orderLog{}
	if w := postAGUI(t, setup(&fakeSessionGuard{busy: true}, busy), "writer", body); w.Code != http.StatusConflict {
		t.Fatalf("busy thread: status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := busy.list(); !slices.Equal(got, []string{"acquire"}) {
		t.Fatalf("busy thread: %v, want the lease tried and the session never read", got)
	}

	free := &orderLog{}
	if w := postAGUI(t, setup(sessionguard.NewMemory(), free), "writer", body); w.Code != http.StatusOK {
		t.Fatalf("free thread: status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := free.list(); len(got) < 2 || got[0] != "acquire" || slices.Contains(got[1:], "acquire") {
		t.Fatalf("free thread: %v, want every session read after the one acquisition", got)
	}
}
