package plane

import (
	"context"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"sync"
	"testing"
)

type progressRecorder struct {
	mu     sync.Mutex
	events []ProgressEvent
}

func (r *progressRecorder) EmitProgress(_ context.Context, _ *Event, e ProgressEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
	return nil
}
func TestProcessorProgressOrderedAndTerminal(t *testing.T) {
	r := &fakeRunner{out: "ok"}
	sink := &progressRecorder{}
	p := NewProcessor(r, &fakePublisher{})
	p.SetProgressSink(sink)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", Prompt: "x"}
	if _, err := p.Process(context.Background(), &agentsv1.PlaneApp{AgentId: "a"}, e); err != nil {
		t.Fatal(err)
	}
	if len(sink.events) != 2 || sink.events[0].Sequence != 1 || sink.events[0].State != RunStateRunning || sink.events[1].Sequence != 2 || sink.events[1].State != RunStateSucceeded {
		t.Fatalf("events=%+v", sink.events)
	}
}
