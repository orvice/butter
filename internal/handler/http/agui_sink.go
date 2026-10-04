package http

import (
	"context"
	"encoding/json"
	"io"
	"sort"

	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	aguitypes "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/types"
	aguisse "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/encoding/sse"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/runtime/interrupt"
	"go.orx.me/apps/butter/internal/runtime/streamorch"
)

// aguiInterruptReason is the AG-UI Interrupt.Reason for a Workflow Agent
// pausing on a Human Input node. AG-UI leaves the vocabulary to the server;
// "human_input" names what the pause actually is.
const aguiInterruptReason = "human_input"

// aguiEmitter delivers one encoded AG-UI event. It is where a run's events
// leave the sink: straight to the response for a run without the opt-in, and
// to the run's in-process fan-out for a Detached Run (aguiFanout.emit), which
// the Run Log replaces later (ADR-0016 decision 5). Tests substitute a
// recorder so event ordering can be asserted without going through SSE.
type aguiEmitter func(aguievents.Event) error

// newAGUISSEEmitter returns an aguiEmitter writing events to w as SSE frames,
// flushing each one so the client sees it immediately.
//
// ctx must be non-nil: the SDK's JSON encoder calls ctx.Err() unguarded and
// panics on a nil context.
func newAGUISSEEmitter(ctx context.Context, w io.Writer, flush func()) aguiEmitter {
	writer := aguisse.NewSSEWriter()
	return func(ev aguievents.Event) error {
		if err := writer.WriteEvent(ctx, w, ev); err != nil {
			return err
		}
		if flush != nil {
			flush()
		}
		return nil
	}
}

// aguiSink implements streamorch.Sink by translating a run's frames into AG-UI
// protocol events.
//
// It lives beside the handler rather than in streamorch because streamorch is
// the protocol-neutral orchestration seam; every Sink implementation belongs
// with its own transport (compare streamAgentSink in internal/application and
// asyncrun's hubSink).
//
// Not concurrency-safe, which is what streamorch.Run guarantees: it calls the
// Sink serially from inside the ADK event loop.
type aguiSink struct {
	threadID  string
	runID     string
	messageID string
	emit      aguiEmitter

	// msgOpen tracks whether a TEXT_MESSAGE_START has been emitted without a
	// matching END, so text can stream lazily and still be closed exactly once.
	msgOpen bool
	// interrupts are the pauses observed in-stream via
	// session.Event.RequestedInput. Collecting them here is what lets Final
	// pick the run's outcome without reaching for runner.TurnResult — see
	// docs/research/ag-ui-integration.md.
	interrupts []aguitypes.Interrupt

	// state mirrors the client-visible session state as the stream advances:
	// seeded with the pre-run authoritative state, updated by every delta the
	// sink emits, and diffed against the post-run authoritative state in
	// Final. It is what makes each STATE_DELTA a correct patch on what the
	// client has already seen.
	state map[string]any
	// emitSnapshot requests a STATE_SNAPSHOT right after RUN_STARTED — set
	// when the client's mirror is absent or diverged from the authoritative
	// state (the conflict answer: the server wins, visibly).
	emitSnapshot bool

	// fetchFinal re-reads the authoritative session after the run. It closes
	// the gap for state deltas the runner never streams (agent output_key
	// writes land on final events, which only reach the callback in special
	// cases), and answers which Interrupts and forms are still open.
	fetchFinal   func() (session.Session, bool)
	finalFetched bool
	finalSess    session.Session
	finalOK      bool
	// reportPending adds every Interrupt still open after the run to the
	// interrupt outcome, not only those raised in-stream. Set for clients
	// that negotiated A2UI; others keep the in-stream-only outcome.
	reportPending bool

	// holdFinished makes Final keep its RUN_FINISHED in heldFinished instead
	// of emitting it, so the handler can settle the run (its record, its run
	// state) before observers learn it ended, and still end it otherwise.
	// releaseRunFinished emits the held event; Error drops it.
	holdFinished bool
	heldFinished aguievents.Event
	// response is the turn's final text, as Final received it.
	response string

	// ui is non-nil when A2UI is live for this run.
	ui *aguiSinkUI
}

// aguiSinkUI tracks what the client has been shown of the session's UI, so
// every persisted change goes out as the envelopes that move the client from
// the state it has to the state the session now holds.
type aguiSinkUI struct {
	// cards is the client's view of every card record, tombstones included.
	cards map[string]*a2ui.Card
	// openForms are the forms pending when the run started or raised during
	// it; each one answered by the end of the run is marked answered.
	openForms []a2ui.Form
}

func newAGUISink(threadID, runID, messageID string, emit aguiEmitter) *aguiSink {
	return &aguiSink{threadID: threadID, runID: runID, messageID: messageID, emit: emit}
}

// setSharedState arms the state mapping for this run. initial must already be
// the client-visible (filtered, normalized) view.
func (s *aguiSink) setSharedState(initial map[string]any, emitSnapshot bool) {
	s.state = initial
	s.emitSnapshot = emitSnapshot
}

// setFinalSession wires the post-run session read.
func (s *aguiSink) setFinalSession(fetch func() (session.Session, bool)) {
	s.fetchFinal = fetch
}

// reportPendingInterrupts makes the outcome list every open Interrupt.
func (s *aguiSink) reportPendingInterrupts() {
	s.reportPending = true
}

// holdRunFinished makes Final hold its RUN_FINISHED until
// releaseRunFinished.
func (s *aguiSink) holdRunFinished() {
	s.holdFinished = true
}

// releaseRunFinished emits the RUN_FINISHED that Final held, if any.
func (s *aguiSink) releaseRunFinished() error {
	ev := s.heldFinished
	s.heldFinished = nil
	if ev == nil {
		return nil
	}
	return s.emit(ev)
}

// setA2UI makes A2UI live for this run. sess is the session as it stood
// before the run; the client is assumed to hold its cards already (from
// earlier runs or the UI snapshot), so only changes are sent.
func (s *aguiSink) setA2UI(sess session.Session) {
	ui := &aguiSinkUI{cards: map[string]*a2ui.Card{}}
	if sess != nil {
		ui.cards = a2ui.Cards(sess.State())
		ui.openForms = a2ui.PendingForms(sess)
	}
	s.ui = ui
}

// finalSession returns the post-run session, fetched at most once.
func (s *aguiSink) finalSession() (session.Session, bool) {
	if !s.finalFetched && s.fetchFinal != nil {
		s.finalFetched = true
		s.finalSess, s.finalOK = s.fetchFinal()
	}
	return s.finalSess, s.finalOK
}

func (s *aguiSink) Started(streamorch.RunIdentity) error {
	if err := s.emit(aguievents.NewRunStartedEvent(s.threadID, s.runID)); err != nil {
		return err
	}
	if s.emitSnapshot {
		return s.emit(aguievents.NewStateSnapshotEvent(s.state))
	}
	return nil
}

func (s *aguiSink) TextDelta(_ streamorch.RunIdentity, text string) error {
	if err := s.openMessage(); err != nil {
		return err
	}
	return s.emit(aguievents.NewTextMessageContentEvent(s.messageID, text))
}

func (s *aguiSink) RunEvent(_ streamorch.RunIdentity, evt *session.Event) error {
	if evt == nil {
		return nil
	}

	// A Workflow Agent pausing for human input is signalled in-stream: the
	// runner forwards request-input events to the callback even though they
	// count as final responses (internal/runtime/runner/runner.go).
	if req := evt.RequestedInput; req != nil {
		s.interrupts = append(s.interrupts, aguitypes.Interrupt{
			ID:      req.InterruptID,
			Reason:  aguiInterruptReason,
			Message: req.Message,
		})
	}

	// Tool-written state changes stream with the events that carry them.
	if s.state != nil && len(evt.Actions.StateDelta) > 0 {
		if err := s.emitStateDelta(aguiVisibleState(evt.Actions.StateDelta)); err != nil {
			return err
		}
	}

	if evt.Content != nil {
		if err := s.emitContent(evt.Content.Parts); err != nil {
			return err
		}
	}
	if s.ui != nil {
		return s.emitUI(evt)
	}
	return nil
}

// emitContent maps an event's function calls and responses onto AG-UI tool
// events.
func (s *aguiSink) emitContent(parts []*genai.Part) error {
	for _, part := range parts {
		if part == nil {
			continue
		}
		// The request-input FunctionCall/Response pair *is* the interrupt
		// handshake, already reported above and below as an Interrupt. Emitting
		// TOOL_CALL_* for it too would make clients render a tool the user
		// never called.
		if fc := part.FunctionCall; fc != nil && fc.Name != workflow.WorkflowInputFunctionCallName {
			if err := s.emitToolCall(fc.ID, fc.Name, fc.Args); err != nil {
				return err
			}
		}
		if fr := part.FunctionResponse; fr != nil && fr.Name != workflow.WorkflowInputFunctionCallName {
			out, err := json.Marshal(fr.Response)
			if err != nil {
				out = []byte("{}")
			}
			if err := s.emit(aguievents.NewToolCallResultEvent(s.messageID, fr.ID, string(out))); err != nil {
				return err
			}
		}
	}
	return nil
}

// emitUI sends the A2UI envelopes a persisted event implies. The runner
// hands the callback an event only after the session stored it, so a client
// never sees a card version or form that is not durable.
func (s *aguiSink) emitUI(evt *session.Event) error {
	// Card writes: each record in the delta moves the client from the
	// version it holds to the stored one.
	var ids []string
	for key := range evt.Actions.StateDelta {
		if id, ok := a2ui.CardIDFromKey(key); ok {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		next, ok := a2ui.DecodeCard(evt.Actions.StateDelta[a2ui.CardKey(id)])
		if !ok {
			continue
		}
		prev := s.ui.cards[id]
		if prev != nil && next.Revision <= prev.Revision {
			continue
		}
		fallback := next.Fallback
		for seq, env := range a2ui.Transition(id, prev, next) {
			if err := s.emitA2UI(a2ui.EventValue{
				SurfaceID: id, Kind: a2ui.KindCard, Revision: next.Revision, Seq: seq,
				Envelope: env, Fallback: fallback,
			}); err != nil {
				return err
			}
		}
		s.ui.cards[id] = next
	}

	// A Human Input node with a form: the server builds the form surface
	// from the binding persisted on the request-input event itself.
	if evt.RequestedInput != nil {
		if form, ok := a2ui.FormOf(evt); ok {
			view := form.View()
			for seq, env := range form.Envelopes() {
				if err := s.emitA2UI(a2ui.EventValue{
					SurfaceID: form.SurfaceID, Kind: a2ui.KindForm, Revision: form.Revision, Seq: seq,
					Envelope: env, Fallback: form.Fallback(), Form: view,
				}); err != nil {
					return err
				}
			}
			s.ui.openForms = append(s.ui.openForms, form)
		}
	}
	return nil
}

func (s *aguiSink) emitA2UI(v a2ui.EventValue) error {
	v.Version = a2ui.Version
	v.ThreadID, v.RunID, v.MessageID = s.threadID, s.runID, s.messageID
	return s.emit(aguievents.NewCustomEvent(a2ui.EventName, aguievents.WithValue(v)))
}

// emitAnsweredForms marks every form whose Interrupt the run answered — by
// form submission, plain resume, or implicit text reply — so the client
// disables it. Answered-ness is read from the stored session, not assumed
// from the request.
func (s *aguiSink) emitAnsweredForms() error {
	if s.ui == nil || len(s.ui.openForms) == 0 {
		return nil
	}
	final, ok := s.finalSession()
	if !ok {
		return nil
	}
	still := interrupt.PendingIDs(final)
	var remaining []a2ui.Form
	for _, form := range s.ui.openForms {
		if still[form.InterruptID] {
			remaining = append(remaining, form)
			continue
		}
		if err := s.emitA2UI(a2ui.EventValue{
			SurfaceID: form.SurfaceID, Kind: a2ui.KindForm, Revision: a2ui.AnsweredRevision,
			Envelope: form.AnsweredEnvelope(), Form: form.View(),
		}); err != nil {
			return err
		}
	}
	s.ui.openForms = remaining
	return nil
}

// outcomeInterrupts is the interrupt list of RUN_FINISHED: the Interrupts
// raised in-stream, plus — when reportPending is set — every other one still
// open, so the client's pending set matches the session.
func (s *aguiSink) outcomeInterrupts() []aguitypes.Interrupt {
	out := append([]aguitypes.Interrupt(nil), s.interrupts...)
	if !s.reportPending {
		return out
	}
	final, ok := s.finalSession()
	if !ok {
		return out
	}
	seen := make(map[string]bool, len(out))
	for _, it := range out {
		seen[it.ID] = true
	}
	for _, p := range interrupt.Pending(final) {
		if seen[p.InterruptID] {
			continue
		}
		out = append(out, aguitypes.Interrupt{ID: p.InterruptID, Reason: aguiInterruptReason, Message: p.Question})
	}
	return out
}

func (s *aguiSink) Final(_ streamorch.RunIdentity, response string) error {
	s.response = response
	// streamorch streams TextDelta only for *partial* events, so a
	// non-streaming turn carries its whole answer in response. Emit it so the
	// client is not left with an empty message.
	//
	// The exception is a paused workflow: runner.run appends each pending
	// question to the turn output, so response is the question that is already
	// travelling as Interrupt.Message. Emitting it again would show the client
	// the same prompt twice.
	if !s.msgOpen && response != "" && len(s.interrupts) == 0 {
		if err := s.openMessage(); err != nil {
			return err
		}
		if err := s.emit(aguievents.NewTextMessageContentEvent(s.messageID, response)); err != nil {
			return err
		}
	}
	if err := s.closeMessage(); err != nil {
		return err
	}
	// Deltas the runner never streamed (output_key and callback writes land
	// on final events) surface here by re-reading the authoritative state.
	if s.state != nil {
		if sess, ok := s.finalSession(); ok {
			final := aguiVisibleState(sessionStateMap(sess))
			if ops := aguiStateDiffOps(s.state, final); len(ops) > 0 {
				if err := s.emit(aguievents.NewStateDeltaEvent(ops)); err != nil {
					return err
				}
				s.state = final
			}
		}
	}
	if err := s.emitAnsweredForms(); err != nil {
		return err
	}
	var finished aguievents.Event
	if interrupts := s.outcomeInterrupts(); len(interrupts) > 0 {
		finished = aguievents.NewRunFinishedEventWithOptions(
			s.threadID, s.runID, aguievents.WithInterruptOutcome(interrupts))
	} else {
		finished = aguievents.NewRunFinishedEventWithOptions(
			s.threadID, s.runID, aguievents.WithSuccessOutcome())
	}
	if s.holdFinished {
		s.heldFinished = finished
		return nil
	}
	return s.emit(finished)
}

// emitStateDelta translates one batch of state changes into a STATE_DELTA
// patch on the client's mirror and folds it into the tracked state. delta
// must already be client-visible (filtered, normalized).
func (s *aguiSink) emitStateDelta(delta map[string]any) error {
	if len(delta) == 0 {
		return nil
	}
	next := make(map[string]any, len(s.state)+len(delta))
	for k, v := range s.state {
		next[k] = v
	}
	for k, v := range delta {
		next[k] = v
	}
	ops := aguiStateDiffOps(s.state, next)
	if len(ops) == 0 {
		return nil
	}
	if err := s.emit(aguievents.NewStateDeltaEvent(ops)); err != nil {
		return err
	}
	s.state = next
	return nil
}

// Error emits RUN_ERROR. streamorch.Sink has no error frame — streamorch.Run
// returns the run error to its caller — so the handler calls this after a
// failed run, or instead of a held RUN_FINISHED for a run that settled as
// failed or stopped. Any open message is closed first so the client is not
// left waiting on a TEXT_MESSAGE_END that never arrives.
func (s *aguiSink) Error(runErr error) error {
	s.heldFinished = nil
	if err := s.closeMessage(); err != nil {
		return err
	}
	// A failed run may still have consumed an Interrupt (the answer is
	// stored before the workflow resumes); its form must not stay open.
	if err := s.emitAnsweredForms(); err != nil {
		return err
	}
	return s.emit(aguievents.NewRunErrorEvent(runErr.Error(), aguievents.WithRunID(s.runID)))
}

func (s *aguiSink) openMessage() error {
	if s.msgOpen {
		return nil
	}
	if err := s.emit(aguievents.NewTextMessageStartEvent(
		s.messageID, aguievents.WithRole("assistant"))); err != nil {
		return err
	}
	s.msgOpen = true
	return nil
}

func (s *aguiSink) closeMessage() error {
	if !s.msgOpen {
		return nil
	}
	if err := s.emit(aguievents.NewTextMessageEndEvent(s.messageID)); err != nil {
		return err
	}
	s.msgOpen = false
	return nil
}

// emitToolCall renders one ADK FunctionCall as the AG-UI three-event sequence.
// ADK hands over complete arguments rather than a token stream, so the whole
// JSON object goes out as a single TOOL_CALL_ARGS delta.
func (s *aguiSink) emitToolCall(id, name string, args map[string]any) error {
	if err := s.emit(aguievents.NewToolCallStartEvent(id, name)); err != nil {
		return err
	}
	encoded, err := json.Marshal(args)
	if err != nil {
		encoded = []byte("{}")
	}
	if err := s.emit(aguievents.NewToolCallArgsEvent(id, string(encoded))); err != nil {
		return err
	}
	return s.emit(aguievents.NewToolCallEndEvent(id))
}

var _ streamorch.Sink = (*aguiSink)(nil)
