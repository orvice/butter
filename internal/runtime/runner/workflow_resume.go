package runner

import (
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"
)

// pendingInterruptsInSession derives the session's unanswered Interrupts by
// scanning its events, oldest first (ADR 0002: no separate pending-interrupt
// store — session events are the single source of truth). A FunctionCall
// part named adk_request_input opens an Interrupt; a FunctionResponse part
// carrying the same ID answers it.
func pendingInterruptsInSession(sess session.Session) []PendingInput {
	if sess == nil {
		return nil
	}
	var ordered []PendingInput
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
				question, _ := fc.Args["message"].(string)
				ordered = append(ordered, PendingInput{InterruptID: fc.ID, Question: question})
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

// resumeParts implements the implicit-resume contract: when the session has
// a pending Interrupt and the inbound message is plain text, the text is
// taken as the answer to the oldest pending Interrupt and rewrapped as the
// FunctionResponse the workflow engine resumes on. Messages that already
// carry a FunctionResponse (a precisely-addressed reply) and turns with no
// pending Interrupt pass through unchanged.
func resumeParts(sess session.Session, parts []*genai.Part) []*genai.Part {
	for _, p := range parts {
		if p != nil && p.FunctionResponse != nil {
			return parts
		}
	}
	answer := joinTextParts(parts)
	if answer == "" {
		return parts
	}
	pending := pendingInterruptsInSession(sess)
	if len(pending) == 0 {
		return parts
	}
	return []*genai.Part{{
		FunctionResponse: &genai.FunctionResponse{
			ID:   pending[0].InterruptID,
			Name: workflow.WorkflowInputFunctionCallName,
			Response: map[string]any{
				"payload": answer,
			},
		},
	}}
}
