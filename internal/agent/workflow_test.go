package agent

import (
	"context"
	"strings"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// workflowProviders returns a provider list whose models can be constructed
// without network access (the openai client only dials on generation).
func workflowProviders() []agentsv1.ModelProvider {
	return []agentsv1.ModelProvider{
		{
			Name:   "openai",
			Type:   "openai",
			Models: []*agentsv1.ModelConfig{{Name: "m1"}},
		},
	}
}

// linearWorkflowProto returns a WORKFLOW agent whose graph is a linear chain
// of two AGENT nodes referencing the agent's sub-agents by name.
func linearWorkflowProto() *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:        "wf",
		Description: "linear two-step workflow",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		SubAgents: []*agentsv1.Agent{
			{Name: "step_a", Config: &agentsv1.AgentConfig{Model: "m1"}},
			{Name: "step_b", Config: &agentsv1.AgentConfig{Model: "m1"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "step_a", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "step_a"},
					{Name: "step_b", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "step_b"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "step_a"},
					{From: "step_a", To: "step_b"},
				},
			},
		},
	}
}

func TestValidateWorkflowAgent_RejectsInvalidGraphs(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(pb *agentsv1.Agent)
		wantErr string
	}{
		{
			name: "duplicate node names",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow.Nodes[1].Name = "step_a"
			},
			wantErr: "duplicate node name",
		},
		{
			name: "edge references unknown node",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow.Edges[1].To = "missing"
			},
			wantErr: "unknown node",
		},
		{
			name: "agent node references missing sub-agent",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow.Nodes[0].Agent = "no_such_agent"
			},
			wantErr: "no_such_agent",
		},
		{
			name: "no entry edge from START",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow.Edges[0].From = "step_b"
			},
			wantErr: "START",
		},
		{
			name: "missing workflow config",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow = nil
			},
			wantErr: "at least one node",
		},
		{
			name: "reserved START node name",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow.Nodes[0].Name = "START"
			},
			wantErr: "reserved",
		},
		{
			name: "unspecified node kind",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow.Nodes[0].Kind = agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_UNSPECIFIED
			},
			wantErr: "kind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := linearWorkflowProto()
			tt.mutate(pb)

			err := ValidateWorkflowAgent(pb)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not mention %q", err.Error(), tt.wantErr)
			}

			// The factory must reject the same graph.
			if _, err := NewFromProto(context.Background(), pb, workflowProviders(), nil, nil, nil); err == nil {
				t.Error("NewFromProto accepted an invalid graph")
			}
		})
	}
}

func TestValidateWorkflowAgent_AcceptsValidGraph(t *testing.T) {
	if err := ValidateWorkflowAgent(linearWorkflowProto()); err != nil {
		t.Fatalf("valid graph rejected: %v", err)
	}
}

func TestValidateWorkflowAgent_RecursesIntoSubAgents(t *testing.T) {
	// A non-workflow root with an invalid workflow sub-agent must be rejected.
	inner := linearWorkflowProto()
	inner.Config.Workflow.Nodes[1].Name = "step_a" // duplicate
	root := &agentsv1.Agent{
		Name:      "root",
		Config:    &agentsv1.AgentConfig{Model: "m1"},
		SubAgents: []*agentsv1.Agent{inner},
	}

	err := ValidateWorkflowAgent(root)
	if err == nil {
		t.Fatal("expected validation error from nested workflow sub-agent")
	}
	if !strings.Contains(err.Error(), "duplicate node name") {
		t.Errorf("error %q does not mention the duplicate node", err.Error())
	}
}

func TestValidateWorkflowAgent_IgnoresNonWorkflowAgents(t *testing.T) {
	pb := &agentsv1.Agent{Name: "llm", Config: &agentsv1.AgentConfig{Model: "m1"}}
	if err := ValidateWorkflowAgent(pb); err != nil {
		t.Fatalf("non-workflow agent rejected: %v", err)
	}
}

func TestNewFromProto_WorkflowLinearChain(t *testing.T) {
	pb := linearWorkflowProto()

	a, err := NewFromProto(context.Background(), pb, workflowProviders(), nil, nil, nil)
	if err != nil {
		t.Fatalf("NewFromProto: %v", err)
	}
	if a.Name() != "wf" {
		t.Errorf("agent name = %q, want %q", a.Name(), "wf")
	}
}
