package agent

import (
	"strings"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func llmAgentWithGuard(t agentsv1.AgentType, cg *agentsv1.ContextGuardConfig) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name: "agent",
		Type: t,
		Config: &agentsv1.AgentConfig{
			ContextGuard: cg,
		},
	}
}

func TestValidateContextGuard_LLMAgentAccepts(t *testing.T) {
	cases := []struct {
		name    string
		agent   *agentsv1.Agent
		wantErr string // empty means accepted
	}{
		{
			name:  "no context guard is accepted",
			agent: &agentsv1.Agent{Name: "a", Type: agentsv1.AgentType_AGENT_TYPE_LLM},
		},
		{
			name: "threshold with override",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTokens: 32000,
			}),
		},
		{
			name: "threshold inherits model metadata (max_tokens unset)",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
			}),
		},
		{
			name: "sliding window with turn limit",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
				MaxTurns: 10,
			}),
		},
		{
			name: "sliding window retains default turn limit (unset)",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
			}),
		},
		{
			name: "legacy unspecified type constructs as LLM and accepts threshold",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED, &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTokens: 64000,
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateContextGuard(tc.agent)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("ValidateContextGuard: unexpected error %v", err)
			}
			if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("ValidateContextGuard error = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestValidateContextGuard_RejectsNonLLMTypes(t *testing.T) {
	cases := []agentsv1.AgentType{
		agentsv1.AgentType_AGENT_TYPE_LOOP,
		agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
		agentsv1.AgentType_AGENT_TYPE_PARALLEL,
		agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		agentsv1.AgentType_AGENT_TYPE_PI,
	}
	for _, typ := range cases {
		t.Run(typ.String(), func(t *testing.T) {
			err := ValidateContextGuard(llmAgentWithGuard(typ, &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTokens: 32000,
			}))
			if err == nil {
				t.Fatalf("expected context_guard rejection for %s", typ)
			}
			if !strings.Contains(err.Error(), "config.context_guard") {
				t.Fatalf("error %q must name the config.context_guard field", err)
			}
		})
	}
}

func TestValidateContextGuard_RejectsInvalidConfigs(t *testing.T) {
	cases := []struct {
		name    string
		agent   *agentsv1.Agent
		wantErr string
	}{
		{
			name: "unspecified strategy is rejected",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_UNSPECIFIED,
			}),
			wantErr: "config.context_guard.strategy is required",
		},
		{
			name: "threshold rejects non-zero max_turns",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTurns: 5,
			}),
			wantErr: "config.context_guard.max_turns is not supported with the threshold strategy",
		},
		{
			name: "threshold rejects negative max_turns",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTurns: -1,
			}),
			wantErr: "config.context_guard.max_turns must not be negative",
		},
		{
			name: "threshold rejects negative max_tokens",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTokens: -1,
			}),
			wantErr: "config.context_guard.max_tokens must not be negative",
		},
		{
			name: "sliding window rejects non-zero max_tokens",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
				MaxTokens: 32000,
			}),
			wantErr: "config.context_guard.max_tokens is not supported with the sliding window strategy",
		},
		{
			name: "sliding window rejects negative max_tokens",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
				MaxTokens: -1,
			}),
			wantErr: "config.context_guard.max_tokens must not be negative",
		},
		{
			name: "sliding window rejects negative max_turns",
			agent: llmAgentWithGuard(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
				MaxTurns: -2,
			}),
			wantErr: "config.context_guard.max_turns must not be negative",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateContextGuard(tc.agent)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestValidateContextGuard_NilAndEmptyPassThrough(t *testing.T) {
	if err := ValidateContextGuard(nil); err != nil {
		t.Fatalf("nil agent: unexpected error %v", err)
	}
	if err := ValidateContextGuard(&agentsv1.Agent{Name: "a", Type: agentsv1.AgentType_AGENT_TYPE_WORKFLOW}); err != nil {
		t.Fatalf("agent without context_guard: unexpected error %v", err)
	}
}
