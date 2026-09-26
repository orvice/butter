package application

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func memoryAgent(id string, mc *agentsv1.MemoryConfig) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:    id,
		AgentId: id,
		Type:    agentsv1.AgentType_AGENT_TYPE_LLM,
		Config:  &agentsv1.AgentConfig{Model: "model-a", Memory: mc},
	}
}

func TestAgentService_MemoryConfigRoundTrips(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())
	created, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: memoryAgent("remembers", &agentsv1.MemoryConfig{
			Enabled: true, EnableTools: true, TopK: proto.Int32(8), Threshold: proto.Float32(0.5),
		}),
	}))
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	mc := created.Msg.GetAgent().GetConfig().GetMemory()
	if !mc.GetEnabled() || !mc.GetEnableTools() || mc.GetTopK() != 8 || mc.GetThreshold() != 0.5 {
		t.Fatalf("memory config = %v", mc)
	}
}

func TestAgentService_MemoryConfigValidationRunsOnEveryWritePath(t *testing.T) {
	invalid := func() *agentsv1.MemoryConfig {
		return &agentsv1.MemoryConfig{Enabled: true, TopK: proto.Int32(0)}
	}

	t.Run("create", func(t *testing.T) {
		store := memory.New()
		svc := NewAgentServiceServer(store)
		_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
			Agent: memoryAgent("invalid-create", invalid()),
		}))
		requireInvalidArgument(t, err, "config.memory.top_k")
		if _, getErr := store.GetAgent(testCtx(), wsTest, "invalid-create"); !errors.Is(getErr, configrepo.ErrNotFound) {
			t.Fatalf("invalid create was persisted: %v", getErr)
		}
	})

	t.Run("direct update", func(t *testing.T) {
		store := memory.New()
		if _, err := store.CreateAgent(testCtx(), wsTest, memoryAgent("invalid-update", nil)); err != nil {
			t.Fatalf("seed agent: %v", err)
		}
		svc := NewAgentServiceServer(store)
		_, err := svc.UpdateAgent(testCtx(), connect.NewRequest(&agentsv1.UpdateAgentRequest{
			Agent: memoryAgent("invalid-update", invalid()),
		}))
		requireInvalidArgument(t, err, "config.memory.top_k")
	})

	t.Run("repository-bound composite update", func(t *testing.T) {
		svc := NewAgentServiceServer(memory.New())
		_, err := svc.UpdateAgentConfiguration(auth.WithAdmin(testCtx()), connect.NewRequest(&agentsv1.UpdateAgentConfigurationRequest{
			AgentPatch: memoryAgent("invalid-composite", invalid()),
		}))
		requireInvalidArgument(t, err, "config.memory.top_k")
	})
}
