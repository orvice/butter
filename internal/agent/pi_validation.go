package agent

import (
	"fmt"
	"strings"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ValidatePiAgent checks that a PI-type agent's config is valid:
//   - config.pi is required with a non-empty butterbox_id
//   - child_agent_ids are forbidden (PI is a leaf)
//   - instruction, global_instruction, mcp_servers, mcp_server_ids, skills,
//     file_mounts, context_guard, and remote_agent_ids are rejected because
//     behavior is configured on the box, not in butter (ADR-0011)
func ValidatePiAgent(a *agentsv1.Agent) error {
	if a.GetType() != agentsv1.AgentType_AGENT_TYPE_PI {
		return nil
	}

	piCfg := a.GetConfig().GetPi()
	if piCfg == nil {
		return fmt.Errorf("type PI requires config.pi")
	}
	if strings.TrimSpace(piCfg.GetButterboxId()) == "" {
		return fmt.Errorf("config.pi.butterbox_id is required")
	}

	if len(a.GetChildAgentIds()) > 0 {
		return fmt.Errorf("PI agents are leaves and cannot have child_agent_ids")
	}

	cfg := a.GetConfig()

	var rejected []string
	if strings.TrimSpace(cfg.GetInstruction()) != "" {
		rejected = append(rejected, "instruction")
	}
	if strings.TrimSpace(cfg.GetGlobalInstruction()) != "" {
		rejected = append(rejected, "global_instruction")
	}
	if len(cfg.GetMcpServers()) > 0 {
		rejected = append(rejected, "mcp_servers")
	}
	if len(cfg.GetMcpServerIds()) > 0 {
		rejected = append(rejected, "mcp_server_ids")
	}
	if len(cfg.GetSkills()) > 0 {
		rejected = append(rejected, "skills")
	}
	if len(cfg.GetFileMounts()) > 0 {
		rejected = append(rejected, "file_mounts")
	}
	if cfg.GetContextGuard() != nil && cfg.GetContextGuard().GetStrategy() != agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_UNSPECIFIED {
		rejected = append(rejected, "context_guard")
	}
	if len(cfg.GetRemoteAgentIds()) > 0 {
		rejected = append(rejected, "remote_agent_ids")
	}

	if len(rejected) > 0 {
		return fmt.Errorf("PI agents do not support %s; configure behavior on the box (AGENTS.md / .pi/)",
			strings.Join(rejected, ", "))
	}

	return nil
}
