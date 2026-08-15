package http

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/streamorch"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// aguiEventRecorder collects emitted AG-UI events so tests can assert on their
// type sequence, and keeps the typed values for field-level assertions.
type aguiEventRecorder struct {
	types  []string
	events []aguievents.Event
}

func (r *aguiEventRecorder) emit(ev aguievents.Event) error {
	r.types = append(r.types, string(ev.Type()))
	r.events = append(r.events, ev)
	return nil
}

func (r *aguiEventRecorder) runFinished(t *testing.T) *aguievents.RunFinishedEvent {
	t.Helper()
	for _, ev := range r.events {
		if f, ok := ev.(*aguievents.RunFinishedEvent); ok {
			return f
		}
	}
	t.Fatalf("no RUN_FINISHED event in %v", r.types)
	return nil
}

// runAGUISink drives the sink through the real streamorch.Run so the tests
// cover the actual frame dispatch (partial-vs-mixed classification included),
// not a hand-rolled imitation of it.
func runAGUISink(t *testing.T, events []*session.Event, response string, runErr error) *aguiEventRecorder {
	t.Helper()
	rec := &aguiEventRecorder{}
	sink := newAGUISink("thread-1", "run-1", "msg-1", rec.emit)
	mock := &mockRunner{
		runResult: response,
		runErr:    runErr,
		onEventFn: func(onEvent runner.EventCallback) {
			for _, evt := range events {
				onEvent(evt)
			}
		},
	}
	ctxInfo := &agentsv1.ContextInfo{Uuid: "inv-1", SessionId: "agui-thread-1"}

	err := streamorch.Run(context.Background(), mock, streamorch.AgentRef{Name: "a", ID: "agent-1"},
		[]*genai.Part{{Text: "hi"}}, "", ctxInfo, sink)
	if err != nil {
		if sinkErr := sink.Error(err); sinkErr != nil {
			t.Fatalf("sink.Error: %v", sinkErr)
		}
	}
	return rec
}

func assertAGUISequence(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("event sequence\n got: %v\nwant: %v", got, want)
	}
}

func TestAGUISink_TextRun(t *testing.T) {
	partial := &session.Event{ID: "e1", Author: "chat"}
	partial.Content = &genai.Content{Parts: []*genai.Part{{Text: "hel"}}}
	partial.Partial = true

	rec := runAGUISink(t, []*session.Event{partial}, "hello", nil)
	assertAGUISequence(t, rec.types, []string{
		"RUN_STARTED", "TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT",
		"TEXT_MESSAGE_END", "RUN_FINISHED",
	})
	if outcome := rec.runFinished(t).Outcome; outcome == nil ||
		outcome.Type != aguievents.RunFinishedOutcomeTypeSuccess {
		t.Fatalf("outcome = %+v, want success", outcome)
	}
}

// A non-streaming turn produces no partial events, so the whole answer arrives
// in Final and must still reach the client as a message.
func TestAGUISink_NonStreamingResponseBecomesMessage(t *testing.T) {
	rec := runAGUISink(t, nil, "the answer", nil)
	assertAGUISequence(t, rec.types, []string{
		"RUN_STARTED", "TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT",
		"TEXT_MESSAGE_END", "RUN_FINISHED",
	})
	content, ok := rec.events[2].(*aguievents.TextMessageContentEvent)
	if !ok || content.Delta != "the answer" {
		t.Fatalf("content event = %+v", rec.events[2])
	}
}

func TestAGUISink_ToolCall(t *testing.T) {
	call := &session.Event{ID: "e1", Author: "chat"}
	call.Content = &genai.Content{Parts: []*genai.Part{{
		FunctionCall: &genai.FunctionCall{ID: "tc-1", Name: "search",
			Args: map[string]any{"q": "go"}},
	}}}
	resp := &session.Event{ID: "e2", Author: "chat"}
	resp.Content = &genai.Content{Parts: []*genai.Part{{
		FunctionResponse: &genai.FunctionResponse{ID: "tc-1", Name: "search",
			Response: map[string]any{"hits": 3}},
	}}}

	rec := runAGUISink(t, []*session.Event{call, resp}, "done", nil)
	assertAGUISequence(t, rec.types, []string{
		"RUN_STARTED",
		"TOOL_CALL_START", "TOOL_CALL_ARGS", "TOOL_CALL_END",
		"TOOL_CALL_RESULT",
		"TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT", "TEXT_MESSAGE_END",
		"RUN_FINISHED",
	})

	start, ok := rec.events[1].(*aguievents.ToolCallStartEvent)
	if !ok || start.ToolCallName != "search" || start.ToolCallID != "tc-1" {
		t.Fatalf("tool call start = %+v", rec.events[1])
	}
	args, ok := rec.events[2].(*aguievents.ToolCallArgsEvent)
	if !ok || args.Delta != `{"q":"go"}` {
		t.Fatalf("tool call args = %+v", rec.events[2])
	}
	result, ok := rec.events[4].(*aguievents.ToolCallResultEvent)
	if !ok || result.ToolCallID != "tc-1" || result.Content != `{"hits":3}` {
		t.Fatalf("tool call result = %+v", rec.events[4])
	}
}

// The decisive case: a Workflow Agent pausing on a Human Input node must reach
// the sink in-stream and become RUN_FINISHED{outcome:interrupt} — no access to
// runner.TurnResult required — with no phantom tool call for the
// adk_request_input FunctionCall that carries the pause.
func TestAGUISink_InterruptOutcome(t *testing.T) {
	rec := runAGUISink(t, []*session.Event{aguiPauseEvent("int-1", "Approve?")}, "Approve?", nil)

	for _, ty := range rec.types {
		if strings.HasPrefix(ty, "TOOL_CALL") {
			t.Fatalf("request-input leaked as a tool call: %v", rec.types)
		}
	}
	// The question travels as Interrupt.Message; it must not also be repeated
	// as assistant content, which runner.run's turn output would otherwise do.
	assertAGUISequence(t, rec.types, []string{"RUN_STARTED", "RUN_FINISHED"})

	outcome := rec.runFinished(t).Outcome
	if outcome == nil || outcome.Type != aguievents.RunFinishedOutcomeTypeInterrupt {
		t.Fatalf("outcome = %+v, want interrupt", outcome)
	}
	if len(outcome.Interrupts) != 1 {
		t.Fatalf("interrupts = %+v, want 1", outcome.Interrupts)
	}
	got := outcome.Interrupts[0]
	if got.ID != "int-1" || got.Message != "Approve?" || got.Reason != aguiInterruptReason {
		t.Fatalf("interrupt = %+v", got)
	}
}

// Text streamed before the pause is still delivered and closed; only the
// duplicated question is suppressed.
func TestAGUISink_InterruptAfterStreamedText(t *testing.T) {
	partial := &session.Event{ID: "e1", Author: "chat"}
	partial.Content = &genai.Content{Parts: []*genai.Part{{Text: "checking"}}}
	partial.Partial = true

	rec := runAGUISink(t,
		[]*session.Event{partial, aguiPauseEvent("int-1", "Approve?")}, "Approve?", nil)
	assertAGUISequence(t, rec.types, []string{
		"RUN_STARTED", "TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT",
		"TEXT_MESSAGE_END", "RUN_FINISHED",
	})
	if outcome := rec.runFinished(t).Outcome; outcome == nil ||
		outcome.Type != aguievents.RunFinishedOutcomeTypeInterrupt {
		t.Fatalf("outcome = %+v, want interrupt", outcome)
	}
}

func TestAGUISink_RunError(t *testing.T) {
	rec := runAGUISink(t, nil, "", errors.New("boom"))
	assertAGUISequence(t, rec.types, []string{"RUN_STARTED", "RUN_ERROR"})
	runErr, ok := rec.events[1].(*aguievents.RunErrorEvent)
	if !ok || runErr.Message != "boom" {
		t.Fatalf("run error = %+v", rec.events[1])
	}
}

// An error mid-message must not leave the client waiting on a
// TEXT_MESSAGE_END that never arrives.
func TestAGUISink_RunErrorClosesOpenMessage(t *testing.T) {
	partial := &session.Event{ID: "e1", Author: "chat"}
	partial.Content = &genai.Content{Parts: []*genai.Part{{Text: "par"}}}
	partial.Partial = true

	rec := runAGUISink(t, []*session.Event{partial}, "", errors.New("boom"))
	assertAGUISequence(t, rec.types, []string{
		"RUN_STARTED", "TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT",
		"TEXT_MESSAGE_END", "RUN_ERROR",
	})
}

// Thought parts are model reasoning, not answer text; they must not surface as
// assistant content (mirrors the dashboard stream and the OpenAI adapter).
func TestAGUISink_ThoughtPartsAreNotContent(t *testing.T) {
	partial := &session.Event{ID: "e1", Author: "chat"}
	partial.Content = &genai.Content{Parts: []*genai.Part{{Text: "thinking", Thought: true}}}
	partial.Partial = true

	rec := runAGUISink(t, []*session.Event{partial}, "answer", nil)
	for _, ev := range rec.events {
		if c, ok := ev.(*aguievents.TextMessageContentEvent); ok && c.Delta == "thinking" {
			t.Fatalf("thought leaked as content: %v", rec.types)
		}
	}
}

// TestAGUISink_SSEEncoding proves the events serialize as real AG-UI SSE frames.
func TestAGUISink_SSEEncoding(t *testing.T) {
	var buf bytes.Buffer
	flushed := 0
	sink := newAGUISink("thread-1", "run-1", "msg-1",
		newAGUISSEEmitter(context.Background(), &buf, func() { flushed++ }))

	for _, step := range []func() error{
		func() error { return sink.Started(streamorch.RunIdentity{}) },
		func() error { return sink.TextDelta(streamorch.RunIdentity{}, "hi") },
		func() error { return sink.Final(streamorch.RunIdentity{}, "hi") },
	} {
		if err := step(); err != nil {
			t.Fatalf("sink step: %v", err)
		}
	}

	out := buf.String()
	for _, want := range []string{
		`"type":"RUN_STARTED"`, `"threadId":"thread-1"`, `"runId":"run-1"`,
		`"type":"TEXT_MESSAGE_CONTENT"`, `"delta":"hi"`,
		`"type":"RUN_FINISHED"`, `"outcome":{"type":"success"}`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("SSE output missing %s\n---\n%s", want, out)
		}
	}
	if !strings.Contains(out, "data: ") || !strings.Contains(out, "\n\n") {
		t.Errorf("not SSE-framed:\n%s", out)
	}
	// Every event must be flushed, or an SSE client sees nothing until the
	// response body closes.
	if flushed != 5 {
		t.Errorf("flushed %d times, want 5 (one per event)", flushed)
	}
}

// aguiPauseEvent builds the event ADK emits when a Workflow Agent's Human
// Input node pauses: the RequestedInput signal plus the adk_request_input
// FunctionCall that carries the handshake.
func aguiPauseEvent(interruptID, question string) *session.Event {
	evt := &session.Event{ID: "pause-" + interruptID, Author: "approval-node"}
	evt.RequestedInput = &session.RequestInput{InterruptID: interruptID, Message: question}
	evt.Content = &genai.Content{Parts: []*genai.Part{{
		FunctionCall: &genai.FunctionCall{
			ID:   interruptID,
			Name: workflow.WorkflowInputFunctionCallName,
			Args: map[string]any{"message": question},
		},
	}}}
	return evt
}
