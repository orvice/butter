package agent

import (
	"math"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func agentWithMemory(t agentsv1.AgentType, mc *agentsv1.MemoryConfig) *agentsv1.Agent {
	return &agentsv1.Agent{Name: "a", Type: t, Config: &agentsv1.AgentConfig{Memory: mc}}
}

func TestValidateMemoryConfig(t *testing.T) {
	cases := []struct {
		name    string
		agent   *agentsv1.Agent
		wantErr string
	}{
		{name: "no config", agent: &agentsv1.Agent{Name: "a"}},
		{name: "no memory", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, nil)},
		{name: "defaults", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.MemoryConfig{Enabled: true})},
		{name: "composite root accepted", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL, &agentsv1.MemoryConfig{Enabled: true})},
		{name: "bounds accepted", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.MemoryConfig{
			Enabled: true, TopK: proto.Int32(MaxMemoryTopK), Threshold: proto.Float32(1),
		})},
		{name: "zero threshold accepted", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.MemoryConfig{Threshold: proto.Float32(0)})},
		{name: "zero top_k", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.MemoryConfig{TopK: proto.Int32(0)}), wantErr: "top_k"},
		{name: "top_k too large", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.MemoryConfig{TopK: proto.Int32(MaxMemoryTopK + 1)}), wantErr: "top_k"},
		{name: "negative threshold", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.MemoryConfig{Threshold: proto.Float32(-0.1)}), wantErr: "threshold"},
		{name: "threshold above one", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.MemoryConfig{Threshold: proto.Float32(1.5)}), wantErr: "threshold"},
		{name: "NaN threshold", agent: agentWithMemory(agentsv1.AgentType_AGENT_TYPE_LLM, &agentsv1.MemoryConfig{Threshold: proto.Float32(float32(math.NaN()))}), wantErr: "threshold"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMemoryConfig(tc.agent)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want mention of %q", err, tc.wantErr)
			}
		})
	}
}
