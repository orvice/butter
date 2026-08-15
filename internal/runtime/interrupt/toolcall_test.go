package interrupt

import (
	"testing"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

// toolCallEvent builds the final event of a run that paused on a frontend
// tool: a FunctionCall listed in LongRunningToolIDs.
func toolCallEvent(id, name string) *session.Event {
	ev := &session.Event{}
	ev.LongRunningToolIDs = []string{id}
	ev.Content = &genai.Content{Parts: []*genai.Part{{
		FunctionCall: &genai.FunctionCall{Name: name, ID: id, Args: map[string]any{"q": "x"}},
	}}}
	return ev
}

// toolAnswerEvent builds the event carrying the client's FunctionResponse.
func toolAnswerEvent(id, name string) *session.Event {
	ev := &session.Event{}
	ev.Content = &genai.Content{Parts: []*genai.Part{{
		FunctionResponse: &genai.FunctionResponse{Name: name, ID: id, Response: map[string]any{"result": "ok"}},
	}}}
	return ev
}

func TestPendingToolCalls(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		if got := PendingToolCalls(nil); got != nil {
			t.Fatalf("PendingToolCalls(nil) = %v", got)
		}
	})

	t.Run("unanswered calls in order, answered ones excluded", func(t *testing.T) {
		sess := &fakeSession{events: []*session.Event{
			toolCallEvent("call-1", "confirm"),
			toolAnswerEvent("call-1", "confirm"),
			toolCallEvent("call-2", "pick_color"),
			toolCallEvent("call-3", "confirm"),
		}}
		got := PendingToolCalls(sess)
		if len(got) != 2 || got[0] != (ToolCall{ID: "call-2", Name: "pick_color"}) || got[1] != (ToolCall{ID: "call-3", Name: "confirm"}) {
			t.Fatalf("PendingToolCalls = %+v", got)
		}
	})

	t.Run("regular tool calls are not pending", func(t *testing.T) {
		// A server-side tool call has no LongRunningToolIDs entry: its
		// response arrives in the same run.
		ev := &session.Event{}
		ev.Content = &genai.Content{Parts: []*genai.Part{{
			FunctionCall: &genai.FunctionCall{Name: "search", ID: "call-9"},
		}}}
		if got := PendingToolCalls(&fakeSession{events: []*session.Event{ev}}); got != nil {
			t.Fatalf("PendingToolCalls = %+v, want none", got)
		}
	})

	t.Run("workflow human input is an Interrupt, not a tool call", func(t *testing.T) {
		ev := askEvent("int-1", "Approve?")
		ev.LongRunningToolIDs = []string{"int-1"}
		if got := PendingToolCalls(&fakeSession{events: []*session.Event{ev}}); got != nil {
			t.Fatalf("PendingToolCalls = %+v, want none", got)
		}
	})
}
