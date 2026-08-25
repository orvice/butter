package agent

import (
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// PiAgentBuilder builds the ADK agent for one AGENT_TYPE_PI leaf. It is
// implemented by the runtime layer (internal/runtime/pibox), which owns the
// ButterBox resolution and the PiService bridge; nil means PI agents cannot
// be built in this deployment.
type PiAgentBuilder func(pb *agentsv1.Agent) (agent.Agent, error)

// ValidatePiAgent checks an AGENT_TYPE_PI config. It is pure proto
// validation — no ButterBox lookup happens here — so the service layer can
// reject a bad config at save time; the box-existence check lives with the
// repositories. Non-PI agents pass through.
//
// A PI agent is a leaf whose whole behavior surface lives on the box
// (ADR-0011): tools come from pi's built-ins, extensions, and skills loaded
// from the working directory, and instructions from its AGENTS.md /
// SYSTEM.md. Every butter-side behavior field is therefore rejected on
// write — a dangling binding that never takes effect is worse than a
// refusal.
func ValidatePiAgent(pb *agentsv1.Agent) error {
	if pb == nil || pb.GetType() != agentsv1.AgentType_AGENT_TYPE_PI {
		return nil
	}
	if err := validatePiConfig(pb); err != nil {
		return fmt.Errorf("agent %q: %w", pb.GetName(), err)
	}
	return nil
}

func validatePiConfig(pb *agentsv1.Agent) error {
	pi := pb.GetConfig().GetPi()
	if pi == nil {
		return fmt.Errorf("a pi agent requires config.pi with the ButterBox binding")
	}
	if strings.TrimSpace(pi.GetButterboxId()) == "" {
		return fmt.Errorf("config.pi.butterbox_id is required: a pi agent binds one ButterBox")
	}
	if pi.MaxRunSeconds != nil && pi.GetMaxRunSeconds() < 0 {
		return fmt.Errorf("config.pi.max_run_seconds must not be negative (unset defaults to 1800, 0 means unlimited)")
	}

	if len(pb.GetChildAgentIds()) > 0 {
		return fmt.Errorf("child_agent_ids is not supported: a pi agent is a leaf — compose it as a child or workflow node of another agent instead")
	}

	cfg := pb.GetConfig()
	boxOwned := []struct {
		set   bool
		field string
	}{
		{cfg.GetInstruction() != "", "instruction"},
		{cfg.GetGlobalInstruction() != "", "global_instruction"},
		{len(cfg.GetMcpServers()) > 0, "mcp_servers"},
		{len(cfg.GetMcpServerIds()) > 0, "mcp_server_ids"},
		{len(cfg.GetSkills()) > 0, "skills"},
		{len(cfg.GetFileMounts()) > 0, "file_mounts"},
		{cfg.GetContextGuard() != nil, "context_guard"},
		{len(cfg.GetRemoteAgentIds()) > 0, "remote_agent_ids"},
	}
	for _, f := range boxOwned {
		if f.set {
			return fmt.Errorf("%s is not supported on a pi agent: pi's tools and instructions are configured on the box, in the working directory's AGENTS.md and .pi/ (ADR-0011)", f.field)
		}
	}
	return nil
}

// newPiAgent validates the PI config and delegates construction to the
// runtime-provided builder.
func newPiAgent(pb *agentsv1.Agent, builder PiAgentBuilder) (agent.Agent, error) {
	if err := ValidatePiAgent(pb); err != nil {
		return nil, err
	}
	if builder == nil {
		return nil, fmt.Errorf("agent %q: pi agents are not available: no ButterBox bridge is wired", pb.GetName())
	}
	return builder(pb)
}
