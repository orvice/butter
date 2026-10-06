package application

import (
	"errors"
	"testing"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func cardPolicyAgent(id string, typ agentsv1.AgentType, rc *agentsv1.ResultCardConfig) *agentsv1.Agent {
	return &agentsv1.Agent{Name: id, AgentId: id, Type: typ, Config: &agentsv1.AgentConfig{ResultCards: rc}}
}

func cardsDisabled() *agentsv1.ResultCardConfig {
	return &agentsv1.ResultCardConfig{Generation: agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED}
}

// LLM and composite agents both keep a Card Policy; on a composite it
// narrows the subtree.
func TestAgentService_CardPolicyRoundTrips(t *testing.T) {
	for id, typ := range map[string]agentsv1.AgentType{
		"cards-llm":        agentsv1.AgentType_AGENT_TYPE_LLM,
		"cards-sequential": agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
	} {
		t.Run(id, func(t *testing.T) {
			svc := NewAgentServiceServer(memory.New())
			created, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
				Agent: cardPolicyAgent(id, typ, cardsDisabled()),
			}))
			if err != nil {
				t.Fatalf("CreateAgent: %v", err)
			}
			if got := created.Msg.GetAgent().GetConfig().GetResultCards().GetGeneration(); got != agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED {
				t.Fatalf("generation = %v, want DISABLED", got)
			}
		})
	}
}

func TestAgentService_CardPolicyValidationRunsOnEveryWritePath(t *testing.T) {
	unknown := func() *agentsv1.ResultCardConfig {
		return &agentsv1.ResultCardConfig{Generation: agentsv1.ResultCardGeneration(7)}
	}
	llm := agentsv1.AgentType_AGENT_TYPE_LLM

	t.Run("create", func(t *testing.T) {
		store := memory.New()
		svc := NewAgentServiceServer(store)
		_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
			Agent: cardPolicyAgent("invalid-create", llm, unknown()),
		}))
		requireInvalidArgument(t, err, "config.result_cards.generation")
		if _, getErr := store.GetAgent(testCtx(), wsTest, "invalid-create"); !errors.Is(getErr, configrepo.ErrNotFound) {
			t.Fatalf("invalid create was persisted: %v", getErr)
		}
	})

	t.Run("direct update", func(t *testing.T) {
		store := memory.New()
		if _, err := store.CreateAgent(testCtx(), wsTest, cardPolicyAgent("invalid-update", llm, nil)); err != nil {
			t.Fatalf("seed agent: %v", err)
		}
		svc := NewAgentServiceServer(store)
		_, err := svc.UpdateAgent(testCtx(), connect.NewRequest(&agentsv1.UpdateAgentRequest{
			Agent: cardPolicyAgent("invalid-update", llm, unknown()),
		}))
		requireInvalidArgument(t, err, "config.result_cards.generation")
	})

	t.Run("repository-bound composite update", func(t *testing.T) {
		svc := NewAgentServiceServer(memory.New())
		_, err := svc.UpdateAgentConfiguration(auth.WithAdmin(testCtx()), connect.NewRequest(&agentsv1.UpdateAgentConfigurationRequest{
			AgentPatch: cardPolicyAgent("invalid-composite", llm, unknown()),
		}))
		requireInvalidArgument(t, err, "config.result_cards.generation")
	})
}

// PI and CURSOR agents keep their behavior on the ButterBox, so a write that
// gives one a Card Policy is refused.
func TestAgentService_CardPolicyRejectedOnBoxAgents(t *testing.T) {
	for name, agent := range map[string]*agentsv1.Agent{
		"pi":     testPiAgent("pi-cards", "box-1"),
		"cursor": testCursorAgent("cursor-cards", "box-1"),
	} {
		t.Run(name, func(t *testing.T) {
			agent.Config.ResultCards = cardsDisabled()
			svc := NewAgentServiceServer(memory.New())
			_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: agent}))
			requireInvalidArgument(t, err, "result_cards")
		})
	}
}
