// Package interrupt is the single seam for a paused workflow's human-input
// state. A Workflow Agent's Human Input node pauses by emitting an
// adk_request_input FunctionCall; the next reply resumes it by answering that
// call with a matching FunctionResponse. Both the runner (implicit resume) and
// the cron scheduler (WAITING_INPUT finalization) reason about that state
// through this package instead of scanning session events themselves.
//
// Per ADR-0002 the session's events are the single source of truth — there is
// no butter-owned pending-interrupt store — so every function here derives its
// answer by scanning events and holds no state of its own.
package interrupt

import (
	"strings"

	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"
)

// Keys of the request-input wire format. ADK builds the FunctionCall args and
// reads the FunctionResponse payload with these literal keys but exports no
// constants for them (see adk/v2/workflow/request_input.go and
// decodeWorkflowInputResponse).
const (
	requestInputMessageKey = "message"
	requestInputPayloadKey = "payload"
)

// Interrupt is one unanswered request for human input: a workflow paused on a
// Human Input node waiting for a reply.
type Interrupt struct {
	InterruptID string
	Question    string
}

// Pending derives the session's unanswered Interrupts by scanning its events
// oldest-first (ADR-0002). A FunctionCall part named adk_request_input opens an
// Interrupt; a FunctionResponse part carrying the same ID answers it. The
// result preserves event order, so the oldest unanswered Interrupt is first
// (FIFO). Returns nil when the session is nil or nothing is pending.
func Pending(sess session.Session) []Interrupt {
	if sess == nil {
		return nil
	}
	var ordered []Interrupt
	answered := map[string]bool{}

	events := sess.Events()
	for i := 0; i < events.Len(); i++ {
		ev := events.At(i)
		if ev == nil || ev.Content == nil {
			continue
		}
		for _, part := range ev.Content.Parts {
			if part == nil {
				continue
			}
			if fc := part.FunctionCall; fc != nil && fc.Name == workflow.WorkflowInputFunctionCallName && fc.ID != "" {
				question, _ := fc.Args[requestInputMessageKey].(string)
				ordered = append(ordered, Interrupt{InterruptID: fc.ID, Question: question})
			}
			if fr := part.FunctionResponse; fr != nil && fr.Name == workflow.WorkflowInputFunctionCallName && fr.ID != "" {
				answered[fr.ID] = true
			}
		}
	}

	pending := ordered[:0]
	for _, p := range ordered {
		if !answered[p.InterruptID] {
			pending = append(pending, p)
		}
	}
	if len(pending) == 0 {
		return nil
	}
	return pending
}

// Resume implements the implicit-resume contract: when the session has a
// pending Interrupt and the inbound message carries plain text, the text is
// taken as the answer to the oldest pending Interrupt (Pending's FIFO order)
// and rewrapped as the FunctionResponse the workflow engine resumes on.
// Non-text parts (for example an image sent alongside the answer) are preserved
// after the response part. The boolean reports whether the rewrap happened.
//
// Messages that already carry a FunctionResponse (a precisely-addressed reply)
// and turns with no pending Interrupt pass through unchanged.
func Resume(sess session.Session, parts []*genai.Part) ([]*genai.Part, bool) {
	for _, p := range parts {
		if p != nil && p.FunctionResponse != nil {
			return parts, false
		}
	}
	answer := joinTextParts(parts)
	if answer == "" {
		return parts, false
	}
	pending := Pending(sess)
	if len(pending) == 0 {
		return parts, false
	}
	resumed := []*genai.Part{{
		FunctionResponse: &genai.FunctionResponse{
			ID:   pending[0].InterruptID,
			Name: workflow.WorkflowInputFunctionCallName,
			Response: map[string]any{
				requestInputPayloadKey: answer,
			},
		},
	}}
	for _, p := range parts {
		if p != nil && p.Text == "" {
			resumed = append(resumed, p)
		}
	}
	return resumed, true
}

// joinTextParts concatenates the text of every text part, space-separated.
func joinTextParts(parts []*genai.Part) string {
	var b strings.Builder
	for _, p := range parts {
		if p == nil || p.Text == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(p.Text)
	}
	return b.String()
}
