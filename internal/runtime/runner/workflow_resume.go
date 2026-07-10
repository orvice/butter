package runner

import (
	"slices"

	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Keys of the request-input wire format. ADK builds the FunctionCall args
// and reads the FunctionResponse payload with these literal keys but exports
// no constants for them (see adk/v2/workflow/request_input.go and
// decodeWorkflowInputResponse).
const (
	requestInputMessageKey = "message"
	requestInputPayloadKey = "payload"
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
				question, _ := fc.Args[requestInputMessageKey].(string)
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
// a pending Interrupt and the inbound message carries plain text, the text
// is taken as the answer to the oldest pending Interrupt and rewrapped as
// the FunctionResponse the workflow engine resumes on. Non-text parts (for
// example images sent alongside the answer) are preserved after the
// response part. The boolean reports whether the rewrap happened.
//
// Messages that already carry a FunctionResponse (a precisely-addressed
// reply) and turns with no pending Interrupt pass through unchanged.
func resumeParts(sess session.Session, parts []*genai.Part) ([]*genai.Part, bool) {
	for _, p := range parts {
		if p != nil && p.FunctionResponse != nil {
			return parts, false
		}
	}
	answer := joinTextParts(parts)
	if answer == "" {
		return parts, false
	}
	pending := pendingInterruptsInSession(sess)
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

// containsWorkflowAgent reports whether the agent config tree holds a
// Workflow Agent anywhere. Implicit resume applies only to such agents: a
// session may be reused with a different agent_name, and a message meant
// for an unrelated agent must not be rewrapped as an Interrupt's answer.
func containsWorkflowAgent(pb *agentsv1.Agent) bool {
	if pb == nil {
		return false
	}
	if pb.GetType() == agentsv1.AgentType_AGENT_TYPE_WORKFLOW {
		return true
	}
	return slices.ContainsFunc(pb.GetSubAgents(), containsWorkflowAgent)
}

// partsCarryFunctionResponse reports whether the message carries any
// FunctionResponse part — the shape of a resume turn.
func partsCarryFunctionResponse(parts []*genai.Part) bool {
	return slices.ContainsFunc(parts, func(p *genai.Part) bool {
		return p != nil && p.FunctionResponse != nil
	})
}
