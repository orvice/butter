package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/a2ui"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/runlog"
	"go.orx.me/apps/butter/internal/runtime/runstate"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Thread reads during a run (#403, ADR-0016 decision 6), through the HTTP
// contract with a real runner, a real session store and the fake model. The
// thread history and the UI snapshot take no lease: during a run they answer
// at once with running, show the thread as the run found it plus the turn
// the run started from, and leave out what the run has stored so far. After
// a run that failed or was stopped, the history reports it as lastRun.

// newReadHarness is the Detached Run harness over the given agents.
func newReadHarness(t *testing.T, agents []agentsv1.Agent, models ...string) *detachHarness {
	t.Helper()
	d := &detachHarness{
		a2uiHarness: &a2uiHarness{t: t, backend: openaifake.New(t), sessions: newStoreLikeSessions(), guard: &fakeSessionGuard{}},
		lease:       newCountingGuard(sessionguard.NewMemory()),
		invocations: newStatusLog(invocationmemory.New().WithOwner(detachOwner)),
		runStates:   runstate.NewMemory(time.Minute),
		runLogs:     runlog.NewMemory(),
	}
	d.recorder = d.invocations
	d.router = d.build(agents, models)
	d.handler.SetSessionGuard(d.lease)
	d.handler.SetInvocationRepo(d.invocations)
	d.handler.SetRunStateStore(d.runStates)
	d.handler.SetRunLogStore(d.runLogs)
	t.Cleanup(d.waitForRuns)
	return d
}

// liveBody is a user turn from an A2UI client, AG-UI Chat's kind, asking
// for a Detached Run when detach is set.
func liveBody(threadID, runID, text string, detach bool) map[string]any {
	body := updateBody(threadID, runID, text)
	if detach {
		body["forwardedProps"].(map[string]any)["butterRun"] = map[string]any{"detach": true}
	}
	return body
}

// renderPreamble opens each reply renderThenWait scripts.
const renderPreamble = "Rendering the summary."

// renderThenWait makes model answer each run's first call with
// renderPreamble and a card, and hold the call that writes the rest of the
// run's reply until the gate opens. While it holds, the run has stored the
// turn it started from, the preamble with its render_ui call, and the card:
// part of a reply, and nothing more.
func (d *detachHarness) renderThenWait(model, reply string) *modelGate {
	g := &modelGate{started: make(chan struct{}, 16), release: make(chan struct{})}
	args, err := json.Marshal(deployCardArgs())
	if err != nil {
		d.t.Fatalf("marshal card: %v", err)
	}
	var renders atomic.Int32
	d.backend.ScriptCall(model, func(w http.ResponseWriter, r *http.Request, req openaifake.ChatCompletionRequest) {
		if req.LastRole() != "tool" {
			id := fmt.Sprintf("render-%d", renders.Add(1))
			openaifake.WriteReply(w, req, renderPreamble, openaifake.ToolCall{ID: id, Name: "render_ui", Arguments: string(args)})
			return
		}
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

// afterPreamble is how the history shows a reply renderThenWait scripted,
// once the run has stored all of it.
func afterPreamble(reply string) string { return renderPreamble + "\n\n" + reply }

// runToEnd runs one detached A2UI turn on the thread to its end, and waits
// until the run has released the thread.
func (d *detachHarness) runToEnd(threadID, runID, text string) {
	d.t.Helper()
	if w := d.post("carder", liveBody(threadID, runID, text, true)); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
		d.t.Fatalf("run %s: status = %d, body = %s", runID, w.Code, w.Body.String())
	}
	d.waitForRuns()
}

// failModel makes every call to model fail.
func (d *detachHarness) failModel(model string) {
	d.backend.ScriptCall(model, func(w http.ResponseWriter, _ *http.Request, _ openaifake.ChatCompletionRequest) {
		http.Error(w, `{"error": {"message": "the model is unavailable"}}`, http.StatusBadRequest)
	})
}

// raceSessions runs hooks around the next session read it is armed for, so
// a test can start or end a run while a thread read is reading the session.
// Every other call goes straight to the store.
type raceSessions struct {
	adksession.Service
	mu   sync.Mutex
	hook *sessionReadHook
}

type sessionReadHook struct{ before, after func() }

// raceReads puts a raceSessions between the handler and its session store.
// The runner keeps writing to the store directly.
func (d *detachHarness) raceReads() *raceSessions {
	s := &raceSessions{Service: d.sessions}
	d.handler.SetSessionService(s)
	return s
}

// arm runs before and after (either may be nil) around the next Get.
func (s *raceSessions) arm(before, after func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hook = &sessionReadHook{before: before, after: after}
}

func (s *raceSessions) Get(ctx context.Context, req *adksession.GetRequest) (*adksession.GetResponse, error) {
	s.mu.Lock()
	hook := s.hook
	s.hook = nil
	s.mu.Unlock()
	if hook != nil && hook.before != nil {
		hook.before()
	}
	resp, err := s.Service.Get(ctx, req)
	if hook != nil && hook.after != nil {
		hook.after()
	}
	return resp, err
}

func (d *detachHarness) leaseAcquisitions() int {
	acquired, _ := d.lease.counts()
	return acquired
}

// requireRunning checks a read's running against the run state of the
// thread's run in flight.
func requireRunning(t *testing.T, label string, got *aguiRunning, want runstate.State) {
	t.Helper()
	if got == nil || got.RunID != want.RunID || got.InvocationID != want.InvocationID {
		t.Fatalf("%s: running = %+v, want runId %s, invocationId %s", label, got, want.RunID, want.InvocationID)
	}
}

func textOf(t *testing.T, msg any) string {
	t.Helper()
	s, ok := msg.(string)
	if !ok {
		t.Fatalf("message content = %#v, want text", msg)
	}
	return s
}

// The issue's headline. While a run holds its thread, detached or not, the
// history and the UI snapshot answer at once, without the lease: the thread
// as the run found it, the turn the run started from, and running. What the
// run has already stored — here a new card, its render_ui call and result —
// is left out, the card included although session state already holds it.
// Once the run ends, the reads show everything and no run.
func TestAGUIRead_DuringARunShowsTheThreadAsTheRunFoundIt(t *testing.T) {
	for _, detached := range []bool{true, false} {
		name := "detached"
		if !detached {
			name = "in its request"
		}
		t.Run(name, func(t *testing.T) {
			d := newReadHarness(t, []agentsv1.Agent{cardAgent(), plainAgent("other", "card-model")}, "card-model")
			// The thread before the run: a first run's turn, its reply and its card.
			d.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Deployed.")
			if w := d.post("carder", liveBody("t-1", "run-1", "deploy", detached)); w.Code != http.StatusOK ||
				!strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
				t.Fatalf("first run: status = %d, body = %s", w.Code, w.Body.String())
			}
			d.waitForRuns()
			_, before := d.history("carder", "t-1")
			beforeCards := d.snapshotSurfaces("carder", "t-1")
			if len(before.Messages) != 2 || len(beforeCards) != 1 {
				t.Fatalf("setup: history %+v, cards %+v", before.Messages, beforeCards)
			}

			model := d.renderThenWait("card-model", "Rolled back.")
			post := d.start("carder", liveBody("t-1", "run-2", "roll it back", detached))
			model.waitStarted(t)
			st, ok := d.runState(detachThread)
			if !ok || st.RunID != "run-2" {
				t.Fatalf("run state = %+v, %v; want run-2's", st, ok)
			}
			sess, _ := d.threadSession("t-1")
			if cards := a2ui.LiveCards(sess.State()); len(cards) != 2 || !agentSaid(sess, renderPreamble) {
				t.Fatalf("setup: the session holds %d cards; want the run's new card and the start of its reply stored", len(cards))
			}
			if !d.lease.isHeld(detachThread) {
				t.Fatal("setup: the run does not hold its thread")
			}
			leases := d.leaseAcquisitions()

			code, during := d.history("carder", "t-1")
			if code != http.StatusOK {
				t.Fatalf("history during the run: status %d, want 200", code)
			}
			requireRunning(t, "history", during.Running, st)
			if during.LastRun != nil {
				t.Errorf("history during the run reports a last run: %+v", during.LastRun)
			}
			if len(during.Messages) != 3 {
				t.Fatalf("history during the run = %+v, want the thread before it and the turn it started from", during.Messages)
			}
			requireJSON(t, "messages before the run", during.Messages[:2], before.Messages)
			if turn := during.Messages[2]; turn.Role != "user" || textOf(t, turn.Content) != "roll it back" {
				t.Fatalf("last message = %+v, want the user's turn the run started from", turn)
			}
			requireJSON(t, "surfaces during the run", during.Surfaces, before.Surfaces)
			requireJSON(t, "interrupts during the run", during.Interrupts, []any{})

			code, snap := d.snapshot("carder", "t-1")
			if code != http.StatusOK {
				t.Fatalf("snapshot during the run: status %d, want 200", code)
			}
			running, _ := snap["running"].(map[string]any)
			if running["runId"] != st.RunID || running["invocationId"] != st.InvocationID {
				t.Fatalf("snapshot running = %+v, want run-2's", snap["running"])
			}
			surfaces, _ := snap["surfaces"].([]any)
			if len(surfaces) != 1 || surfaces[0].(map[string]any)["surfaceId"] != beforeCards[0]["surfaceId"] {
				t.Fatalf("snapshot surfaces during the run = %+v, want only the card from before it", surfaces)
			}

			// The binding checks are unchanged: the thread through another
			// agent's route, or another caller's thread of the same ID, reads
			// empty, and names no run.
			for label, read := range map[string]func() (int, aguiThreadHistory){
				"another agent": func() (int, aguiThreadHistory) { return d.history("other", "t-1") },
				"another user":  func() (int, aguiThreadHistory) { return d.history("carder", "t-1", asUser("u2")) },
			} {
				if code, h := read(); code != http.StatusOK || len(h.Messages) != 0 || h.Running != nil || h.LastRun != nil {
					t.Errorf("%s: status %d, history %+v; want an empty one", label, code, h)
				}
			}
			if code, other := d.snapshot("other", "t-1"); code != http.StatusOK || other["running"] != nil || len(other["surfaces"].([]any)) != 0 {
				t.Errorf("another agent's snapshot: status %d, body %+v; want an empty one", code, other)
			}
			if got := d.leaseAcquisitions(); got != leases {
				t.Fatalf("the reads took the lease %d times", got-leases)
			}

			model.open()
			post.wait(t)
			d.waitForRuns()

			_, after := d.history("carder", "t-1")
			if after.Running != nil || after.LastRun != nil {
				t.Fatalf("history after the run: running %+v, lastRun %+v; want neither", after.Running, after.LastRun)
			}
			if len(after.Messages) != 4 || textOf(t, after.Messages[3].Content) != afterPreamble("Rolled back.") || len(after.Surfaces) != 2 {
				t.Fatalf("history after the run = %+v, surfaces %+v; want the run's reply and card", after.Messages, after.Surfaces)
			}
			_, snapAfter := d.snapshot("carder", "t-1")
			if snapAfter["running"] != nil || len(snapAfter["surfaces"].([]any)) != 2 {
				t.Fatalf("snapshot after the run = %+v, want both cards and no run", snapAfter)
			}
		})
	}
}

// An Interrupt the run answers does not read as open while the run goes on:
// the history shows the question and the form's answer, reports no open
// Interrupt, and the snapshot no longer offers the form.
func TestAGUIRead_AnInterruptTheRunAnswersIsNotOpen(t *testing.T) {
	d := newReadHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
	d.echoModels("drafter")
	fx := d.pauseWithForm("t-form")
	if _, paused := d.history("approval", "t-form"); len(paused.Interrupts) != 1 {
		t.Fatalf("setup: interrupts = %+v, want the form's", paused.Interrupts)
	}

	publisher := d.gate("publisher", "Published.")
	body := resumeBody("t-form", "run-2", fx.interruptID, formSubmission(fx.form, fx.surfaceID, validDeployValues()))
	body["forwardedProps"].(map[string]any)["butterRun"] = map[string]any{"detach": true}
	post := d.start("approval", body)
	publisher.waitStarted(t)
	st, ok := d.runState("u1:agui-t-form")
	if !ok {
		t.Fatal("no run state while the run runs")
	}

	code, during := d.history("approval", "t-form")
	if code != http.StatusOK {
		t.Fatalf("history during the run: status %d", code)
	}
	requireRunning(t, "history", during.Running, st)
	requireJSON(t, "interrupts during the run", during.Interrupts, []any{})
	last := during.Messages[len(during.Messages)-1]
	if last.Role != "user" || !strings.HasPrefix(textOf(t, last.Content), "Deploy approval\nEnvironment: Staging\nReason: weekly") {
		t.Fatalf("last message = %+v, want the form's answer", last)
	}
	question := during.Messages[len(during.Messages)-2]
	if question.Role != "assistant" || !strings.Contains(textOf(t, question.Content), "Approve this deploy?") {
		t.Fatalf("message before the answer = %+v, want the question it answered", question)
	}
	if surfaces := d.snapshotSurfaces("approval", "t-form"); len(surfaces) != 0 {
		t.Fatalf("snapshot during the run = %+v, want the answered form gone", surfaces)
	}

	publisher.open()
	post.wait(t)
	d.waitForRuns()
	_, after := d.history("approval", "t-form")
	if after.Running != nil || len(after.Interrupts) != 0 ||
		!strings.Contains(textOf(t, after.Messages[len(after.Messages)-1].Content), "Published.") {
		t.Fatalf("history after the run = %+v", after)
	}
}

// A run that starts from client tool results reads, while it goes on, as the
// call those results answer, the results, and nothing the run has produced
// since. The thread and the run are written to the stores directly: the
// first turn is whatever a run stores first, as the tests above show.
func TestAGUIRead_ARunStartedFromToolResultsShowsThem(t *testing.T) {
	d := newDetachHarness(t)
	ctx := wsctx.WithID(context.Background(), "ws-a")
	binding := a2ui.Binding{Principal: "u1", WorkspaceID: "ws-a", AgentID: "carder", ThreadID: "t-1"}
	created, err := d.sessions.Create(ctx, &adksession.CreateRequest{
		AppName: aguiAppName, UserID: "u1", SessionID: "agui-t-1",
		State: map[string]any{a2ui.BindingKey: binding.StateValue()},
	})
	if err != nil {
		t.Fatalf("create the thread: %v", err)
	}
	sess := created.Session
	appendEvents := func(events ...*adksession.Event) {
		t.Helper()
		for _, ev := range events {
			if err := d.sessions.AppendEvent(ctx, sess, ev); err != nil {
				t.Fatalf("append: %v", err)
			}
		}
	}

	// The thread's last run ended on a client tool call, still pending.
	ask := userEvent(&genai.Part{Text: "ship the release"})
	call := agentEvent(callPart("confirm-1", "confirm", map[string]any{"q": "Ship it?"}))
	call.LongRunningToolIDs = []string{"confirm-1"}
	appendEvents(ask, call)
	// A run starts from the client's result, and has stored the start of its
	// reply.
	if err := d.runStates.Begin(ctx, detachThread, runstate.State{RunID: "run-2", InvocationID: "inv-2", EventCount: 2, Detached: true}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	result := userEvent(responsePart("confirm-1", "confirm", map[string]any{"approved": true}))
	reply := agentEvent(&genai.Part{Text: "Shipping now."})
	appendEvents(result, reply)

	code, during := d.history("carder", "t-1")
	if code != http.StatusOK || during.Running == nil || during.Running.RunID != "run-2" {
		t.Fatalf("status %d, history %+v; want run-2 running", code, during)
	}
	requireJSON(t, "messages during the run", during.Messages, []map[string]any{
		{"id": ask.ID, "role": "user", "content": "ship the release"},
		{"id": call.ID, "role": "assistant", "toolCalls": []map[string]any{
			{"id": "confirm-1", "type": "function", "function": map[string]any{"name": "confirm", "arguments": `{"q":"Ship it?"}`}},
		}},
		{"id": "result:confirm-1", "role": "tool", "content": `{"approved":true}`, "toolCallId": "confirm-1"},
	})

	if err := d.runStates.End(ctx, detachThread, "inv-2", 0); err != nil {
		t.Fatalf("End: %v", err)
	}
	if _, after := d.history("carder", "t-1"); after.Running != nil || len(after.Messages) != 4 ||
		textOf(t, after.Messages[3].Content) != "Shipping now." {
		t.Fatalf("history after the run = %+v, want its reply", after)
	}
}

// A run that ends while a read reads the session never shows as half a
// reply: the read saw the start of the run's reply and its card stored but
// not the rest, saw the run ended when it looked at the run state again, and
// read the thread again.
func TestAGUIRead_ARunEndingDuringAReadNeverShowsHalfAReply(t *testing.T) {
	for _, endpoint := range []string{"history", "snapshot"} {
		t.Run(endpoint, func(t *testing.T) {
			d := newDetachHarness(t)
			model := d.renderThenWait("card-model", "Deployed.")
			post := d.start("carder", liveBody("t-1", "run-1", "deploy", true))
			model.waitStarted(t)

			sessions := d.raceReads()
			raced := false
			sessions.arm(nil, func() {
				// The read has the session as the run left it so far: half a
				// reply and the card. Now the run ends.
				sess, _ := d.threadSession("t-1")
				if !agentSaid(sess, renderPreamble) || agentSaid(sess, "Deployed.") || len(a2ui.LiveCards(sess.State())) != 1 {
					t.Error("setup: the read did not catch the run half way")
				}
				model.open()
				d.waitForRuns()
				raced = true
			})

			if endpoint == "history" {
				code, got := d.history("carder", "t-1")
				if code != http.StatusOK || got.Running != nil || len(got.Messages) != 2 ||
					textOf(t, got.Messages[1].Content) != afterPreamble("Deployed.") || len(got.Surfaces) != 1 {
					t.Fatalf("status %d, history %+v; want the whole run and no run in flight", code, got)
				}
			} else {
				code, snap := d.snapshot("carder", "t-1")
				if code != http.StatusOK || snap["running"] != nil || len(snap["surfaces"].([]any)) != 1 {
					t.Fatalf("status %d, snapshot %+v; want the run's card and no run in flight", code, snap)
				}
			}
			if !raced {
				t.Fatal("the read never reached the session")
			}
			post.wait(t)
		})
	}
}

// A run that starts while a read reads the session never shows as half a
// reply: the read saw no run, then the run's first turn, half its reply and
// its card stored, then the run, and read the thread again, cut where the run
// started.
func TestAGUIRead_ARunStartingDuringAReadNeverShowsHalfAReply(t *testing.T) {
	for _, endpoint := range []string{"history", "snapshot"} {
		t.Run(endpoint, func(t *testing.T) {
			d := newDetachHarness(t)
			d.answer("card-model", "Hello.")
			d.runToEnd("t-1", "run-1", "hi")

			model := d.renderThenWait("card-model", "Deployed.")
			sessions := d.raceReads()
			var post *pendingPost
			sessions.arm(func() {
				post = d.start("carder", liveBody("t-1", "run-2", "deploy", true))
				model.waitStarted(t)
			}, nil)

			if endpoint == "history" {
				code, got := d.history("carder", "t-1")
				if code != http.StatusOK || got.Running == nil || got.Running.RunID != "run-2" {
					t.Fatalf("status %d, history %+v; want run-2 running", code, got)
				}
				if len(got.Messages) != 3 || textOf(t, got.Messages[2].Content) != "deploy" || len(got.Surfaces) != 0 {
					t.Fatalf("history = %+v, surfaces %+v; want the thread before the run and its first turn, nothing more", got.Messages, got.Surfaces)
				}
			} else {
				code, snap := d.snapshot("carder", "t-1")
				running, _ := snap["running"].(map[string]any)
				if code != http.StatusOK || running["runId"] != "run-2" || len(snap["surfaces"].([]any)) != 0 {
					t.Fatalf("status %d, snapshot %+v; want run-2 running and not its card", code, snap)
				}
			}
			if post == nil {
				t.Fatal("the read never reached the session")
			}
			model.open()
			post.wait(t)
		})
	}
}

// A run that both starts and ends while a read reads the session leaves no
// run state either time the read looks. Its Invocation record, written before
// its first event, gives it away, and the read starts over instead of showing
// half a reply.
func TestAGUIRead_ARunStartingAndEndingDuringAReadNeverShowsHalfAReply(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "Hello.")
	d.runToEnd("t-1", "run-1", "hi")

	model := d.renderThenWait("card-model", "Deployed.")
	sessions := d.raceReads()
	var post *pendingPost
	sessions.arm(func() {
		post = d.start("carder", liveBody("t-1", "run-2", "deploy", true))
		model.waitStarted(t)
	}, func() {
		model.open()
		d.waitForRuns()
	})

	code, got := d.history("carder", "t-1")
	if post == nil {
		t.Fatal("the read never reached the session")
	}
	if code != http.StatusOK || got.Running != nil || len(got.Messages) != 4 ||
		textOf(t, got.Messages[3].Content) != afterPreamble("Deployed.") || len(got.Surfaces) != 1 {
		t.Fatalf("status %d, history %+v; want the whole of run-2", code, got)
	}
	post.wait(t)
}

// A read that comes between a run's start and the run's first write waits a
// little for it: the turn the run started from belongs in the read. A run
// that has not stored it by the read's last attempt reads without it.
func TestAGUIRead_WaitsBrieflyForTheTurnARunStartedFrom(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "Hello.")
	d.runToEnd("t-1", "run-1", "hi")
	sess, _ := d.threadSession("t-1")
	stored := sess.Events().Len()

	// A run holds the thread and has not stored its first turn yet; it does
	// so just after the read's first look at the session.
	ctx := context.Background()
	if err := d.runStates.Begin(ctx, detachThread, runstate.State{RunID: "run-2", InvocationID: "inv-2", EventCount: stored}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	sessions := d.raceReads()
	sessions.arm(nil, func() {
		if err := d.sessions.AppendEvent(ctx, sess, userEvent(&genai.Part{Text: "next"})); err != nil {
			t.Errorf("append the run's first turn: %v", err)
		}
	})
	code, got := d.history("carder", "t-1")
	if code != http.StatusOK || got.Running == nil || got.Running.RunID != "run-2" ||
		len(got.Messages) != 3 || textOf(t, got.Messages[2].Content) != "next" {
		t.Fatalf("status %d, history %+v; want run-2 running and its first turn", code, got)
	}

	// The next run never stores its first turn while the read waits.
	if err := d.runStates.Begin(ctx, detachThread, runstate.State{RunID: "run-3", InvocationID: "inv-3", EventCount: stored + 1}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	d.handler.readBackoff = time.Millisecond
	code, got = d.history("carder", "t-1")
	if code != http.StatusOK || got.Running == nil || got.Running.RunID != "run-3" || len(got.Messages) != 3 {
		t.Fatalf("status %d, history %+v; want run-3 running and the thread as it found it", code, got)
	}
}

// churnSessions starts a new run on the thread every time the session is
// read, as a thread under a storm of short runs would look.
type churnSessions struct {
	adksession.Service
	runStates runstate.Store
	reads     atomic.Int32
}

func (s *churnSessions) Get(ctx context.Context, req *adksession.GetRequest) (*adksession.GetResponse, error) {
	n := s.reads.Add(1)
	_ = s.runStates.Begin(ctx, req.UserID+":"+req.SessionID,
		runstate.State{RunID: fmt.Sprintf("run-%d", n), InvocationID: fmt.Sprintf("inv-%d", n)})
	return s.Service.Get(ctx, req)
}

// A thread whose runs keep starting while it is read answers a retryable 503
// once the read's attempts run out, never a half-read thread; so does a
// thread whose run state cannot be read.
func TestAGUIRead_AnUnsettledThreadIsRetryable(t *testing.T) {
	t.Run("runs keep starting", func(t *testing.T) {
		d := newDetachHarness(t)
		d.answer("card-model", "Hello.")
		d.runToEnd("t-1", "run-1", "hi")
		churn := &churnSessions{Service: d.sessions, runStates: d.runStates}
		d.handler.SetSessionService(churn)
		d.handler.readBackoff = time.Millisecond

		if code, _ := d.history("carder", "t-1"); code != http.StatusServiceUnavailable {
			t.Fatalf("status %d, want 503", code)
		}
		if reads := churn.reads.Load(); reads != aguiReadAttempts {
			t.Fatalf("the read looked at the session %d times, want %d", reads, aguiReadAttempts)
		}
		if code, _ := d.snapshot("carder", "t-1"); code != http.StatusServiceUnavailable {
			t.Fatalf("snapshot: status %d, want 503", code)
		}
	})

	t.Run("run state unavailable", func(t *testing.T) {
		d := newDetachHarness(t)
		d.answer("card-model", "Hello.")
		d.runToEnd("t-1", "run-1", "hi")
		d.handler.SetRunStateStore(brokenRunStates{Store: d.runStates})
		if code, _ := d.history("carder", "t-1"); code != http.StatusServiceUnavailable {
			t.Errorf("history: status %d, want 503", code)
		}
		if code, _ := d.snapshot("carder", "t-1"); code != http.StatusServiceUnavailable {
			t.Errorf("snapshot: status %d, want 503", code)
		}
	})
}

// lastRun comes only from this thread's own runs. A record that another
// caller or another app keeps under the same session ID in the workspace, as
// data from before a session ID had one holder may, is never reported, even
// when it is the newest: the thread's own latest run is.
func TestAGUIRead_LastRunIsOnlyTheThreadsOwn(t *testing.T) {
	for name, foreign := range map[string]func(inv *agentsv1.Invocation){
		"another caller": func(inv *agentsv1.Invocation) { inv.UserId = "u2" },
		"another app":    func(inv *agentsv1.Invocation) { inv.AppName = "web-chat" },
	} {
		for _, ownFailed := range []bool{false, true} {
			label := name + "/own run succeeded"
			if ownFailed {
				label = name + "/own run failed"
			}
			t.Run(label, func(t *testing.T) {
				d := newDetachHarness(t)
				if ownFailed {
					d.failModel("card-model")
				} else {
					d.answer("card-model", "Hello.")
				}
				if w := d.post("carder", liveBody("t-1", "run-1", "hi", true)); w.Code != http.StatusOK {
					t.Fatalf("first run: status = %d", w.Code)
				}
				d.waitForRuns()
				own, _ := d.record("t-1", "run-1")
				other := &agentsv1.Invocation{
					Id: "inv-other", AppName: aguiAppName, UserId: "u1", SessionId: "agui-t-1", WorkspaceId: "ws-a",
					Status: agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED, Error: "boom", Input: "someone else's words",
					StartedAt: timestamppb.New(own.GetStartedAt().AsTime().Add(time.Second)),
				}
				foreign(other)
				if err := d.invocations.Save(context.Background(), other); err != nil {
					t.Fatalf("save: %v", err)
				}

				code, got := d.history("carder", "t-1")
				if code != http.StatusOK {
					t.Fatalf("status %d", code)
				}
				if !ownFailed {
					if got.LastRun != nil {
						t.Fatalf("lastRun = %+v, want none: the thread's own latest run succeeded", got.LastRun)
					}
					return
				}
				requireJSON(t, "lastRun", got.LastRun, aguiLastRun{Status: "failed", Error: own.GetError(), Input: "hi"})
			})
		}
	}
}

// A read whose client gives up while the read waits to start over still
// answers, with a 503, and stops reading.
func TestAGUIRead_ACancelledReadAnswers(t *testing.T) {
	d := newDetachHarness(t)
	d.answer("card-model", "Hello.")
	d.runToEnd("t-1", "run-1", "hi")
	churn := &churnSessions{Service: d.sessions, runStates: d.runStates}
	d.handler.SetSessionService(churn)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/agui/carder/threads/t-1/messages", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "cancelled") {
		t.Fatalf("status = %d, body = %s; want a 503", w.Code, w.Body.String())
	}
	if reads := churn.reads.Load(); reads != 1 {
		t.Fatalf("the read looked at the session %d times after its client left, want once", reads)
	}
}

// brokenRunStates is a run state store that cannot be read.
type brokenRunStates struct{ runstate.Store }

func (brokenRunStates) Get(context.Context, string) (runstate.State, bool, error) {
	return runstate.State{}, false, errors.New("redis: connection refused")
}

// After a run that failed or was stopped, the history reports it as lastRun
// with what the user sent, from the run's Invocation record, so the turn can
// be restored after a reload. The report waits while a run is in flight, and
// a later run that succeeds clears it.
func TestAGUIRead_HistoryReportsAFailedOrStoppedLastRun(t *testing.T) {
	cases := map[string]struct {
		detach bool
		// end ends the run, once its model was called.
		end    func(d *detachHarness, post *pendingPost)
		status string
	}{
		"failed": {
			detach: true,
			end:    func(d *detachHarness, post *pendingPost) { post.wait(d.t) },
			status: "failed",
		},
		"failed in its request": {
			end:    func(d *detachHarness, post *pendingPost) { post.wait(d.t) },
			status: "failed",
		},
		"cancelled": {
			detach: true,
			end: func(d *detachHarness, post *pendingPost) {
				if w := stopRun(d.t, d.router, "carder", "t-1"); w.Code != http.StatusAccepted {
					d.t.Fatalf("Stop: status = %d, body = %s; want 202", w.Code, w.Body.String())
				}
				post.wait(d.t)
			},
			status: "cancelled",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := newDetachHarness(t)
			var model *modelGate
			if tc.status == "failed" {
				d.failModel("card-model")
			} else {
				model = d.gate("card-model", "too late")
			}
			post := d.start("carder", liveBody("t-1", "run-1", "Summarize the deploy", tc.detach))
			if model != nil {
				model.waitStarted(t)
			}
			tc.end(d, post)
			d.waitForRuns()

			all := d.allRecords()
			if len(all) != 1 {
				t.Fatalf("records = %+v, want the run's", all)
			}
			inv := all[0]
			code, got := d.history("carder", "t-1")
			if code != http.StatusOK || got.Running != nil {
				t.Fatalf("status %d, history %+v", code, got)
			}
			requireJSON(t, "lastRun", got.LastRun, aguiLastRun{Status: tc.status, Error: inv.GetError(), Input: "Summarize the deploy"})
			if inv.GetError() == "" {
				t.Fatalf("the %s run's record has no reason: %+v", name, inv)
			}
			if tc.status == "cancelled" && inv.GetError() != errAGUIRunStopped.Error() {
				t.Fatalf("a stopped run's reason = %q", inv.GetError())
			}
			// The turn itself stays in the history.
			if len(got.Messages) == 0 || textOf(t, got.Messages[0].Content) != "Summarize the deploy" {
				t.Fatalf("messages = %+v, want the failed turn", got.Messages)
			}
			// Nobody else learns of it.
			if _, other := d.history("carder", "t-1", asUser("u2")); other.LastRun != nil {
				t.Fatalf("another user's read reports the run: %+v", other.LastRun)
			}

			// While the next run goes on, the history reports it, not the
			// one before.
			next := d.gate("card-model", "All done.")
			retry := d.start("carder", liveBody("t-1", "run-2", "Summarize the deploy", true))
			next.waitStarted(t)
			if _, during := d.history("carder", "t-1"); during.Running == nil || during.LastRun != nil {
				t.Fatalf("during the next run: running %+v, lastRun %+v; want the run and no last run", during.Running, during.LastRun)
			}
			next.open()
			retry.wait(t)
			d.waitForRuns()
			if _, after := d.history("carder", "t-1"); after.LastRun != nil || after.Running != nil {
				t.Fatalf("after a run that succeeded: lastRun %+v, running %+v; want neither", after.LastRun, after.Running)
			}
		})
	}
}
