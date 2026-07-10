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

// branchingWorkflowProto returns a WORKFLOW agent with an approve/reject
// graph: an agent's answer feeds a Router that sends "approve" down one
// branch and everything else down the default branch. Conditional branches
// deliberately do not converge into a Join: the barrier waits for every
// declared predecessor, and a route-skipped predecessor never fires.
func branchingWorkflowProto() *agentsv1.Agent {
	return &agentsv1.Agent{
		Name: "review",
		Type: agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		SubAgents: []*agentsv1.Agent{
			{Name: "classify", Config: &agentsv1.AgentConfig{Model: "m1"}},
			{Name: "approver", Config: &agentsv1.AgentConfig{Model: "m1"}},
			{Name: "rejecter", Config: &agentsv1.AgentConfig{Model: "m1"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "classify", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "classify"},
					{Name: "decide", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_ROUTER},
					{Name: "approver", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "approver"},
					{Name: "rejecter", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "rejecter"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "classify"},
					{From: "classify", To: "decide"},
					{From: "decide", To: "approver", Route: "approve"},
					{From: "decide", To: "rejecter", IsDefault: true},
				},
			},
		},
	}
}

// fanOutJoinWorkflowProto returns a WORKFLOW agent that fans out from one
// node to two branches over unconditional edges and re-converges through a
// Join node.
func fanOutJoinWorkflowProto() *agentsv1.Agent {
	return &agentsv1.Agent{
		Name: "fanout",
		Type: agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		SubAgents: []*agentsv1.Agent{
			{Name: "seed", Config: &agentsv1.AgentConfig{Model: "m1"}},
			{Name: "b1", Config: &agentsv1.AgentConfig{Model: "m1"}},
			{Name: "b2", Config: &agentsv1.AgentConfig{Model: "m1"}},
			{Name: "summarize", Config: &agentsv1.AgentConfig{Model: "m1"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "seed", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "seed"},
					{Name: "b1", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "b1"},
					{Name: "b2", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "b2"},
					{Name: "gather", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_JOIN},
					{Name: "summarize", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "summarize"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "seed"},
					{From: "seed", To: "b1"},
					{From: "seed", To: "b2"},
					{From: "b1", To: "gather"},
					{From: "b2", To: "gather"},
					{From: "gather", To: "summarize"},
				},
			},
		},
	}
}

// humanInputWorkflowProto returns a WORKFLOW agent that pauses on a Human
// Input node between two agent steps: draft -> ask (human) -> publish.
func humanInputWorkflowProto() *agentsv1.Agent {
	return &agentsv1.Agent{
		Name: "approval",
		Type: agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		SubAgents: []*agentsv1.Agent{
			{Name: "draft", Config: &agentsv1.AgentConfig{Model: "m1"}},
			{Name: "publish", Config: &agentsv1.AgentConfig{Model: "m1"}},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "draft", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "draft"},
					{Name: "ask", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Approve this draft?"},
					{Name: "publish", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "publish"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "draft"},
					{From: "draft", To: "ask"},
					{From: "ask", To: "publish"},
				},
			},
		},
	}
}

func TestNewFromProto_WorkflowHumanInput(t *testing.T) {
	a, err := NewFromProto(context.Background(), humanInputWorkflowProto(), workflowProviders(), nil, nil, nil)
	if err != nil {
		t.Fatalf("NewFromProto: %v", err)
	}
	if a.Name() != "approval" {
		t.Errorf("agent name = %q, want %q", a.Name(), "approval")
	}
}

// TestValidateWorkflowAgent_HumanInputRequiresQuestion: a Human Input node
// without a question would pause the workflow with an empty prompt.
func TestValidateWorkflowAgent_HumanInputRequiresQuestion(t *testing.T) {
	pb := humanInputWorkflowProto()
	pb.Config.Workflow.Nodes[1].Question = ""
	assertGraphRejected(t, pb, "question")
}

func TestNewFromProto_WorkflowRouterAndJoin(t *testing.T) {
	for _, pb := range []*agentsv1.Agent{branchingWorkflowProto(), fanOutJoinWorkflowProto()} {
		a, err := NewFromProto(context.Background(), pb, workflowProviders(), nil, nil, nil)
		if err != nil {
			t.Fatalf("NewFromProto(%q): %v", pb.GetName(), err)
		}
		if a.Name() != pb.GetName() {
			t.Errorf("agent name = %q, want %q", a.Name(), pb.GetName())
		}
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
		{
			name: "parallel worker on a non-agent node",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow.Nodes = append(pb.Config.Workflow.Nodes, &agentsv1.WorkflowNode{
					Name:           "gather",
					Kind:           agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_JOIN,
					ParallelWorker: true,
				})
				pb.Config.Workflow.Edges = append(pb.Config.Workflow.Edges,
					&agentsv1.WorkflowEdge{From: "step_b", To: "gather"})
			},
			wantErr: "parallel_worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := linearWorkflowProto()
			tt.mutate(pb)
			assertGraphRejected(t, pb, tt.wantErr)
		})
	}
}

// TestValidateWorkflowAgent_RejectsConditionalEdgeIntoJoin: a Join barrier
// waits for every declared predecessor, and a route-skipped predecessor
// never fires — a routed or default edge into a JOIN node produces a graph
// that hangs at runtime, so validation must reject it at save time.
func TestValidateWorkflowAgent_RejectsConditionalEdgeIntoJoin(t *testing.T) {
	t.Run("routed edge into join", func(t *testing.T) {
		pb := fanOutJoinWorkflowProto()
		pb.Config.Workflow.Edges[3].Route = "left" // b1 -> gather
		assertGraphRejected(t, pb, "gather")
	})
	t.Run("default edge into join", func(t *testing.T) {
		pb := fanOutJoinWorkflowProto()
		pb.Config.Workflow.Edges[4].IsDefault = true // b2 -> gather
		assertGraphRejected(t, pb, "gather")
	})
}

// TestValidateWorkflowAgent_RejectsNearDuplicateRouteLabels: route matching
// is trimmed and case-insensitive, so two outgoing labels that differ only
// by case or whitespace can never both be reachable — only the first would
// ever fire. Validation must reject the ambiguity.
func TestValidateWorkflowAgent_RejectsNearDuplicateRouteLabels(t *testing.T) {
	pb := branchingWorkflowProto()
	pb.Config.Workflow.Edges = append(pb.Config.Workflow.Edges,
		&agentsv1.WorkflowEdge{From: "decide", To: "rejecter", Route: " Approve "})
	assertGraphRejected(t, pb, "approve")
}

// TestValidateWorkflowAgent_RouterRequiresDefaultEdge: an unmatched Router
// with no default edge dead-ends silently in the ADK engine, so validation
// must require one.
func TestValidateWorkflowAgent_RouterRequiresDefaultEdge(t *testing.T) {
	pb := branchingWorkflowProto()
	// Drop the default edge, keeping the router reachable and the graph
	// otherwise valid.
	pb.Config.Workflow.Edges = []*agentsv1.WorkflowEdge{
		{From: "START", To: "classify"},
		{From: "classify", To: "decide"},
		{From: "decide", To: "approver", Route: "approve"},
		{From: "decide", To: "rejecter", Route: "reject"},
	}
	assertGraphRejected(t, pb, "default")
}

func assertGraphRejected(t *testing.T, pb *agentsv1.Agent, wantErr string) {
	t.Helper()

	err := ValidateWorkflowAgent(pb)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("error %q does not mention %q", err.Error(), wantErr)
	}

	// The factory must reject the same graph.
	if _, err := NewFromProto(context.Background(), pb, workflowProviders(), nil, nil, nil); err == nil {
		t.Error("NewFromProto accepted an invalid graph")
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

// TestMatchRouteLabel: the Router matches its input text against outgoing
// edge labels with a trimmed, case-insensitive exact match, and stamps the
// label as configured (the engine compares route tags to labels verbatim).
func TestMatchRouteLabel(t *testing.T) {
	labels := []string{"approve", "REJECT "}
	tests := []struct {
		input     string
		wantLabel string
		wantOK    bool
	}{
		{input: "approve", wantLabel: "approve", wantOK: true},
		{input: " APPROVE ", wantLabel: "approve", wantOK: true},
		{input: "\tReject\n", wantLabel: "REJECT ", wantOK: true},
		{input: "approved", wantOK: false}, // exact match, not prefix
		{input: "", wantOK: false},
	}
	for _, tt := range tests {
		got, ok := matchRouteLabel(tt.input, labels)
		if ok != tt.wantOK || got != tt.wantLabel {
			t.Errorf("matchRouteLabel(%q) = (%q, %v), want (%q, %v)",
				tt.input, got, ok, tt.wantLabel, tt.wantOK)
		}
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
