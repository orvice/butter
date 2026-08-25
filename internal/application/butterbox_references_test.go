package application

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeAgentRepoForButterBoxGuard struct {
	configrepo.AgentRepository
	agents []*agentsv1.Agent
}

func (f *fakeAgentRepoForButterBoxGuard) ListAgents(_ context.Context, _ string) ([]*agentsv1.Agent, error) {
	return f.agents, nil
}

func piAgent(agentID, butterboxID string) *agentsv1.Agent {
	return &agentsv1.Agent{
		AgentId: agentID,
		Type:    agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{
			Pi: &agentsv1.PiAgentConfig{
				ButterboxId: butterboxID,
			},
		},
	}
}

func TestButterBoxReferenceGuard_NilGuard(t *testing.T) {
	var guard *ButterBoxReferenceGuard
	if err := guard.CheckRemovable(t.Context(), "ws-a", "bb-1"); err != nil {
		t.Fatalf("nil guard blocked delete: %v", err)
	}
}

func TestButterBoxReferenceGuard_NoReferences(t *testing.T) {
	repo := &fakeAgentRepoForButterBoxGuard{agents: nil}
	guard := NewButterBoxReferenceGuard(repo)
	if err := guard.CheckRemovable(t.Context(), "ws-a", "bb-1"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestButterBoxReferenceGuard_NonPiAgentsIgnored(t *testing.T) {
	repo := &fakeAgentRepoForButterBoxGuard{
		agents: []*agentsv1.Agent{
			{
				AgentId: "llm-1",
				Type:    agentsv1.AgentType_AGENT_TYPE_LLM,
				Config: &agentsv1.AgentConfig{
					Pi: &agentsv1.PiAgentConfig{ButterboxId: "bb-1"},
				},
			},
		},
	}
	guard := NewButterBoxReferenceGuard(repo)
	if err := guard.CheckRemovable(t.Context(), "ws-a", "bb-1"); err != nil {
		t.Fatalf("LLM agent with pi config should not block: %v", err)
	}
}

func TestButterBoxReferenceGuard_SingleReference(t *testing.T) {
	repo := &fakeAgentRepoForButterBoxGuard{
		agents: []*agentsv1.Agent{piAgent("pi-1", "bb-1")},
	}
	guard := NewButterBoxReferenceGuard(repo)
	err := guard.CheckRemovable(t.Context(), "ws-a", "bb-1")
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, want FailedPrecondition", code)
	}
	if !strings.Contains(err.Error(), "pi-1") {
		t.Errorf("error should list agent_id, got %q", err.Error())
	}
}

func TestButterBoxReferenceGuard_MultipleReferences(t *testing.T) {
	repo := &fakeAgentRepoForButterBoxGuard{
		agents: []*agentsv1.Agent{
			piAgent("pi-1", "bb-1"),
			piAgent("pi-2", "bb-1"),
		},
	}
	guard := NewButterBoxReferenceGuard(repo)
	err := guard.CheckRemovable(t.Context(), "ws-a", "bb-1")
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, want FailedPrecondition", code)
	}
	if !strings.Contains(err.Error(), "pi-1") || !strings.Contains(err.Error(), "pi-2") {
		t.Errorf("error should list both agent_ids, got %q", err.Error())
	}
}

func TestButterBoxReferenceGuard_DifferentButterbox(t *testing.T) {
	repo := &fakeAgentRepoForButterBoxGuard{
		agents: []*agentsv1.Agent{piAgent("pi-1", "bb-other")},
	}
	guard := NewButterBoxReferenceGuard(repo)
	if err := guard.CheckRemovable(t.Context(), "ws-a", "bb-1"); err != nil {
		t.Fatalf("different butterbox should not block: %v", err)
	}
}
