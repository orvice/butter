package agent

import (
	"testing"

	adkagent "google.golang.org/adk/v2/agent"

	"go.orx.me/apps/butter/internal/a2ui"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestPiBuilderReceivesInheritedCardPolicy(t *testing.T) {
	child := validPiAgent()
	root := &agentsv1.Agent{
		AgentId: "root", Name: "root", Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
		ChildAgentIds: []string{child.AgentId},
		Config: &agentsv1.AgentConfig{ResultCards: &agentsv1.ResultCardConfig{
			Generation: agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED,
		}},
	}
	var policies []a2ui.CardPolicy
	builder := func(pb *agentsv1.Agent, p a2ui.CardPolicy) (adkagent.Agent, error) {
		policies = append(policies, p)
		return adkagent.New(adkagent.Config{Name: pb.GetName()})
	}
	builders := &BoxAgentBuilders{Pi: builder}
	if _, err := NewFromProtoWithToolsetFactory(t.Context(), root, nil, nil, nil, nil, nil, nil, builders, AgentPool{child.AgentId: child}); err != nil {
		t.Fatal(err)
	}
	if len(policies) != 1 || policies[0].Allowed() {
		t.Fatal("Pi leaf escaped its parent's disabled policy")
	}
	if _, err := NewFromProtoWithToolsetFactory(t.Context(), child, nil, nil, nil, nil, nil, nil, builders); err != nil {
		t.Fatal(err)
	}
	if len(policies) != 2 || !policies[1].Allowed() {
		t.Fatal("direct Pi invocation inherited another tree's policy")
	}
}
