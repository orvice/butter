package runner

import (
	"slices"

	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// The derivation of pending Interrupts from session events and the matching of
// a reply to the oldest unanswered one live in internal/runtime/interrupt —
// the single seam both the runner (implicit resume, below) and the cron
// scheduler (WAITING_INPUT finalization, via TurnResult.Pending) consume.

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
