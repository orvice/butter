package interrupt

import (
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
)

// ToolCall is one unanswered long-running (frontend) tool call: the agent
// declared the call, the run ended with it pending, and the tool executes
// outside the server — an AG-UI client answers it on a later request.
type ToolCall struct {
	ID   string
	Name string
}

// PendingToolCalls derives the session's unanswered long-running tool calls
// by scanning its events oldest-first, the same single-source-of-truth
// derivation as Pending (ADR-0002). A FunctionCall part whose ID is listed in
// the event's LongRunningToolIDs opens a pending call; a FunctionResponse
// with the same ID answers it. Workflow human-input calls are excluded — they
// are Interrupts, not tool calls. Returns nil when nothing is pending.
func PendingToolCalls(sess session.Session) []ToolCall {
	if sess == nil {
		return nil
	}
	var ordered []ToolCall
	answered := map[string]bool{}

	events := sess.Events()
	for i := 0; i < events.Len(); i++ {
		ev := events.At(i)
		if ev == nil || ev.Content == nil {
			continue
		}
		longRunning := map[string]bool{}
		for _, id := range ev.LongRunningToolIDs {
			longRunning[id] = true
		}
		for _, part := range ev.Content.Parts {
			if part == nil {
				continue
			}
			if fc := part.FunctionCall; fc != nil && fc.ID != "" &&
				longRunning[fc.ID] && fc.Name != workflow.WorkflowInputFunctionCallName {
				ordered = append(ordered, ToolCall{ID: fc.ID, Name: fc.Name})
			}
			if fr := part.FunctionResponse; fr != nil && fr.ID != "" {
				answered[fr.ID] = true
			}
		}
	}

	pending := ordered[:0]
	for _, c := range ordered {
		if !answered[c.ID] {
			pending = append(pending, c)
		}
	}
	if len(pending) == 0 {
		return nil
	}
	return pending
}
