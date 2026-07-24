package interrupt

import (
	"iter"
	"time"

	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"
)

// fakeSession is a minimal session.Session carrying a fixed event list, so the
// derivation seam can be exercised on synthetic events without spinning up an
// ADK runner.
type fakeSession struct {
	events []*session.Event
}

func (f *fakeSession) ID() string                { return "s1" }
func (f *fakeSession) AppName() string           { return "app" }
func (f *fakeSession) UserID() string            { return "u1" }
func (f *fakeSession) State() session.State      { return nil }
func (f *fakeSession) LastUpdateTime() time.Time { return time.Time{} }
func (f *fakeSession) Events() session.Events    { return fakeEvents(f.events) }

type fakeEvents []*session.Event

func (e fakeEvents) Len() int                { return len(e) }
func (e fakeEvents) At(i int) *session.Event { return e[i] }
func (e fakeEvents) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		for _, ev := range e {
			if !yield(ev) {
				return
			}
		}
	}
}

// askEvent builds an event carrying a request-input FunctionCall — a workflow
// pausing on a Human Input node with the given interrupt ID and question.
func askEvent(id, question string) *session.Event {
	ev := &session.Event{}
	ev.Content = &genai.Content{Parts: []*genai.Part{{
		FunctionCall: &genai.FunctionCall{
			Name: workflow.WorkflowInputFunctionCallName,
			ID:   id,
			Args: map[string]any{"message": question},
		},
	}}}
	return ev
}

// answerEvent builds an event carrying the FunctionResponse that answers the
// interrupt with the given ID.
func answerEvent(id string) *session.Event {
	ev := &session.Event{}
	ev.Content = &genai.Content{Parts: []*genai.Part{{
		FunctionResponse: &genai.FunctionResponse{
			Name:     workflow.WorkflowInputFunctionCallName,
			ID:       id,
			Response: map[string]any{"payload": "some answer"},
		},
	}}}
	return ev
}
