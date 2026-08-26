package agent

import (
	"fmt"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// isLLMAgentType reports whether an agent type constructs as an ADK LLM
// agent, the only type that owns an ADK model call.
func isLLMAgentType(t agentsv1.AgentType) bool {
	return t == agentsv1.AgentType_AGENT_TYPE_LLM || t == agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED
}

// ValidateContextGuard checks an agent's ContextGuard configuration. It is
// pure proto validation - no models, providers, or ADK plugins are
// constructed - so the service layer can reject a bad config at save time on
// every Agent write path (issue #322).
//
// Semantics enforced here (the "Agent Context Override" contract):
//
//   - ContextGuard applies only to ADK LLM Agents: AGENT_TYPE_LLM and
//     AGENT_TYPE_UNSPECIFIED (which constructs as an LLM agent). It is
//     rejected on Loop, Sequential, Parallel, Workflow, and box-backed
//     (AGENT_TYPE_PI) types because those records do not make the relevant
//     ADK model call; referenced child LLM Agents own their own context
//     policy.
//   - A present ContextGuard must select a concrete strategy; an unspecified
//     strategy is a silent no-op and is rejected.
//   - CONTEXT_GUARD_STRATEGY_THRESHOLD accepts max_tokens = 0 (inherit
//     Model/embedded metadata) or a positive Agent Context Override. It
//     rejects any non-zero max_turns because that field has no threshold
//     meaning.
//   - CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW accepts max_turns = 0 (the
//     existing default of 20 content entries) or a positive value. It
//     rejects any non-zero max_tokens because the ContextGuard dependency
//     ignores that option for sliding-window compaction - accepting it would
//     save a configuration that silently does nothing.
//   - Negative max_tokens and max_turns values are rejected for every
//     strategy.
func ValidateContextGuard(pb *agentsv1.Agent) error {
	if pb == nil {
		return nil
	}
	cfg := pb.GetConfig()
	if cfg == nil || cfg.GetContextGuard() == nil {
		return nil
	}

	if err := validateContextGuard(pb); err != nil {
		return fmt.Errorf("agent %q: %w", pb.GetName(), err)
	}
	return nil
}

func validateContextGuard(pb *agentsv1.Agent) error {
	cg := pb.GetConfig().GetContextGuard()

	if !isLLMAgentType(pb.GetType()) {
		return fmt.Errorf(
			"config.context_guard is not supported on %s agents: context management applies to ADK LLM agents - configure context_guard on the standalone LLM agent that makes the model call instead",
			pb.GetType().String())
	}

	if cg.GetMaxTokens() < 0 {
		return fmt.Errorf("config.context_guard.max_tokens must not be negative (it overrides the context window in tokens; unset means inherit the model's capacity)")
	}
	if cg.GetMaxTurns() < 0 {
		return fmt.Errorf("config.context_guard.max_turns must not be negative (it is the maximum conversation turns before sliding-window compaction; unset means the default of 20)")
	}

	switch cg.GetStrategy() {
	case agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_UNSPECIFIED:
		return fmt.Errorf("config.context_guard.strategy is required: choose CONTEXT_GUARD_STRATEGY_THRESHOLD or CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW")
	case agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD:
		if cg.GetMaxTurns() != 0 {
			return fmt.Errorf("config.context_guard.max_turns is not supported with the threshold strategy: threshold compaction is driven by the context window (config.context_guard.max_tokens), not by turn count")
		}
	case agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW:
		if cg.GetMaxTokens() != 0 {
			return fmt.Errorf("config.context_guard.max_tokens is not supported with the sliding window strategy: sliding-window compaction is driven by the content-entry limit (config.context_guard.max_turns); the model's context capacity is used for the post-compaction safety check instead")
		}
	default:
		return fmt.Errorf("config.context_guard.strategy %v is not a known strategy", cg.GetStrategy())
	}
	return nil
}
