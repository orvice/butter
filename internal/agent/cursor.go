package agent

import (
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// CursorAgentBuilder builds the ADK agent for one AGENT_TYPE_CURSOR leaf. It
// is implemented by the runtime layer (internal/runtime/cursorbox), which owns
// the ButterBox resolution and the CursorService bridge; nil means Cursor
// agents cannot be built in this deployment.
type CursorAgentBuilder func(pb *agentsv1.Agent) (agent.Agent, error)

// ValidateCursorAgent checks an AGENT_TYPE_CURSOR config. It is pure proto
// validation — no ButterBox lookup happens here — so the service layer can
// reject a bad config at save time; the box-existence check lives with the
// repositories.
//
// A Cursor agent is a leaf whose whole behavior surface lives on the box
// (mirroring PI, ADR-0011): tools come from .cursor/rules, mcp.json, and
// hooks, and model inference goes through Cursor's API. Every butter-side
// behavior field is therefore rejected on write.
func ValidateCursorAgent(pb *agentsv1.Agent) error {
	if pb == nil || pb.GetType() != agentsv1.AgentType_AGENT_TYPE_CURSOR {
		return nil
	}
	if err := validateCursorConfig(pb); err != nil {
		return fmt.Errorf("agent %q: %w", pb.GetName(), err)
	}
	return nil
}

func validateCursorConfig(pb *agentsv1.Agent) error {
	cur := pb.GetConfig().GetCursor()
	if cur == nil {
		return fmt.Errorf("a cursor agent requires config.cursor with the ButterBox binding")
	}
	if strings.TrimSpace(cur.GetButterboxId()) == "" {
		return fmt.Errorf("config.cursor.butterbox_id is required: a cursor agent binds one ButterBox")
	}
	if cur.MaxRunSeconds != nil && cur.GetMaxRunSeconds() < 0 {
		return fmt.Errorf("config.cursor.max_run_seconds must not be negative (unset defaults to 1800, 0 means unlimited)")
	}
	if m := strings.TrimSpace(cur.GetMode()); m != "" && m != "agent" && m != "plan" {
		return fmt.Errorf("config.cursor.mode must be \"agent\", \"plan\", or empty (defaults to \"agent\")")
	}

	if len(pb.GetChildAgentIds()) > 0 {
		return fmt.Errorf("child_agent_ids is not supported: a cursor agent is a leaf — compose it as a child or workflow node of another agent instead")
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
			return fmt.Errorf("%s is not supported on a cursor agent: Cursor's tools and instructions are configured on the box, in .cursor/rules and mcp.json", f.field)
		}
	}
	return nil
}

// newCursorAgent validates the Cursor config and delegates construction to the
// runtime-provided builder.
func newCursorAgent(pb *agentsv1.Agent, builder CursorAgentBuilder) (agent.Agent, error) {
	if err := ValidateCursorAgent(pb); err != nil {
		return nil, err
	}
	if builder == nil {
		return nil, fmt.Errorf("agent %q: cursor agents are not available: no ButterBox bridge is wired", pb.GetName())
	}
	return builder(pb)
}
