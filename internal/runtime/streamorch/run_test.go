package streamorch

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// spySink records every call it receives, in order, so tests can assert on
// frame ordering without caring about transport encoding. startErr, when
// set, is returned by Started to simulate the first frame failing to send.
type spySink struct {
	calls    []string
	startErr error
}

func (s *spySink) Started(id RunIdentity) error {
	s.calls = append(s.calls, "started:"+id.InvocationID+":"+id.SessionID+":"+id.AgentName)
	return s.startErr
}

func (s *spySink) TextDelta(id RunIdentity, text string) error {
	s.calls = append(s.calls, "delta:"+text)
	return nil
}

func (s *spySink) RunEvent(id RunIdentity, evt *session.Event) error {
	s.calls = append(s.calls, "event:"+evt.ID)
	return nil
}

func (s *spySink) Final(id RunIdentity, response string) error {
	s.calls = append(s.calls, "final:"+response)
	return nil
}

// fakeRunner emits a fixed sequence of events via onEvent, then returns a
// fixed response/error.
type fakeRunner struct {
	events   []*session.Event
	response string
	err      error
	called   bool
}

func (r *fakeRunner) RunSSE(_ context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, onEvent runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.called = true
	for _, evt := range r.events {
		onEvent(evt)
	}
	return r.response, r.err
}

func TestRun_TextOnlyPartialEventEmitsOnlyDelta(t *testing.T) {
	evt := &session.Event{ID: "evt-1"}
	evt.Content = &genai.Content{Parts: []*genai.Part{{Text: "hel"}}}
	evt.Partial = true

	fake := &fakeRunner{events: []*session.Event{evt}, response: "hello"}
	sink := &spySink{}
	ctxInfo := &agentsv1.ContextInfo{Uuid: "inv-1", SessionId: "sess-1"}

	if err := Run(context.Background(), fake, "chat-agent", []*genai.Part{{Text: "hi"}}, "", ctxInfo, sink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"started:inv-1:sess-1:chat-agent", "delta:hel", "final:hello"}
	if len(sink.calls) != len(want) {
		t.Fatalf("expected calls %v, got %v", want, sink.calls)
	}
	for i := range want {
		if sink.calls[i] != want[i] {
			t.Fatalf("expected calls %v, got %v", want, sink.calls)
		}
	}
}

func TestRun_MixedEventEmitsDeltaThenRunEvent(t *testing.T) {
	evt := &session.Event{ID: "evt-2"}
	evt.Content = &genai.Content{Parts: []*genai.Part{
		{Text: "partial text"},
		{FunctionCall: &genai.FunctionCall{Name: "lookup"}},
	}}
	evt.Partial = true

	fake := &fakeRunner{events: []*session.Event{evt}, response: "done"}
	sink := &spySink{}
	ctxInfo := &agentsv1.ContextInfo{Uuid: "inv-1", SessionId: "sess-1"}

	if err := Run(context.Background(), fake, "chat-agent", []*genai.Part{{Text: "hi"}}, "", ctxInfo, sink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"started:inv-1:sess-1:chat-agent", "delta:partial text", "event:evt-2", "final:done"}
	if len(sink.calls) != len(want) {
		t.Fatalf("expected calls %v, got %v", want, sink.calls)
	}
	for i := range want {
		if sink.calls[i] != want[i] {
			t.Fatalf("expected calls %v, got %v", want, sink.calls)
		}
	}
}

func TestRun_CancellationMidStreamStopsBeforeFinal(t *testing.T) {
	evt := &session.Event{ID: "evt-1"}
	evt.Content = &genai.Content{Parts: []*genai.Part{{Text: "partial"}}}
	evt.Partial = true

	fake := &fakeRunner{events: []*session.Event{evt}, err: context.Canceled}
	sink := &spySink{}
	ctxInfo := &agentsv1.ContextInfo{Uuid: "inv-1", SessionId: "sess-1"}

	err := Run(context.Background(), fake, "chat-agent", []*genai.Part{{Text: "hi"}}, "", ctxInfo, sink)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	want := []string{"started:inv-1:sess-1:chat-agent", "delta:partial"}
	if len(sink.calls) != len(want) {
		t.Fatalf("expected calls %v (no final on cancellation), got %v", want, sink.calls)
	}
	for i := range want {
		if sink.calls[i] != want[i] {
			t.Fatalf("expected calls %v, got %v", want, sink.calls)
		}
	}
}

func TestRun_ErrorPropagatesWithoutFinal(t *testing.T) {
	fake := &fakeRunner{err: errors.New("boom")}
	sink := &spySink{}
	ctxInfo := &agentsv1.ContextInfo{Uuid: "inv-1", SessionId: "sess-1"}

	err := Run(context.Background(), fake, "chat-agent", []*genai.Part{{Text: "hi"}}, "", ctxInfo, sink)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error containing %q, got %v", "boom", err)
	}
	for _, call := range sink.calls {
		if strings.HasPrefix(call, "final:") {
			t.Fatalf("expected no final frame on error, got calls %v", sink.calls)
		}
	}
}

func TestRun_StartedFailureShortCircuitsBeforeRunnerCalled(t *testing.T) {
	sinkErr := errors.New("client already gone")
	fake := &fakeRunner{response: "hello"}
	sink := &spySink{startErr: sinkErr}
	ctxInfo := &agentsv1.ContextInfo{Uuid: "inv-1", SessionId: "sess-1"}

	err := Run(context.Background(), fake, "chat-agent", []*genai.Part{{Text: "hi"}}, "", ctxInfo, sink)
	if !errors.Is(err, sinkErr) {
		t.Fatalf("expected the sink's Started error to propagate unchanged, got %v", err)
	}
	if fake.called {
		t.Fatal("expected the runner to never be invoked when Started fails")
	}
	if len(sink.calls) != 1 || !strings.HasPrefix(sink.calls[0], "started:") {
		t.Fatalf("expected only the Started call to be recorded, got %v", sink.calls)
	}
}

func TestRun_StartedThenFinalWithNoEvents(t *testing.T) {
	fake := &fakeRunner{response: "hello"}
	sink := &spySink{}
	ctxInfo := &agentsv1.ContextInfo{Uuid: "inv-1", SessionId: "sess-1"}

	err := Run(context.Background(), fake, "chat-agent", []*genai.Part{{Text: "hi"}}, "", ctxInfo, sink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"started:inv-1:sess-1:chat-agent", "final:hello"}
	if len(sink.calls) != len(want) {
		t.Fatalf("expected calls %v, got %v", want, sink.calls)
	}
	for i := range want {
		if sink.calls[i] != want[i] {
			t.Fatalf("expected calls %v, got %v", want, sink.calls)
		}
	}
}
