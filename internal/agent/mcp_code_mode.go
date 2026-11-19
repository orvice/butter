package agent

import (
	"fmt"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ValidateMCPCodeMode checks where config.mcp_code_mode may be set. Only an
// LLM agent offers MCP tools to a model, so the flag means nothing on a
// composite or workflow agent and is refused there instead of being silently
// ignored. PI and CURSOR agents reject it through rejectBoxOwnedFields. A
// partial patch that does not name a type is not checked: the stored agent's
// type decides, and the runtime ignores the flag off an LLM agent.
func ValidateMCPCodeMode(pb *agentsv1.Agent) error {
	if !pb.GetConfig().GetMcpCodeMode() {
		return nil
	}
	switch pb.GetType() {
	case agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED, agentsv1.AgentType_AGENT_TYPE_LLM:
		return nil
	case agentsv1.AgentType_AGENT_TYPE_PI, agentsv1.AgentType_AGENT_TYPE_CURSOR:
		// Reported by the box-agent validation, with its box-side hint.
		return nil
	default:
		return fmt.Errorf("agent %q: config.mcp_code_mode is only supported on LLM agents, not %v", pb.GetName(), pb.GetType())
	}
}
