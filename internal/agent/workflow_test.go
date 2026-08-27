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

// childAgent returns an independent LLM agent usable as a workflow child.
func childAgent(id string) *agentsv1.Agent {
	return &agentsv1.Agent{AgentId: id, Name: id, Config: &agentsv1.AgentConfig{Model: "m1"}}
}

// poolOf indexes children by agent_id, mirroring the runner's per-workspace
// pool construction.
func poolOf(children ...*agentsv1.Agent) AgentPool {
	pool := make(AgentPool, len(children))
	for _, c := range children {
		pool[c.GetAgentId()] = c
	}
	return pool
}

// linearWorkflowProto returns a WORKFLOW agent whose graph is a linear chain
// of two AGENT nodes referencing independent children by agent_id.
func linearWorkflowProto() (*agentsv1.Agent, AgentPool) {
	pb := &agentsv1.Agent{
		AgentId:       "wf",
		Name:          "wf",
		Description:   "linear two-step workflow",
		Type:          agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: []string{"step-a", "step-b"},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "step_a", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "step-a"},
					{Name: "step_b", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "step-b"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "step_a"},
					{From: "step_a", To: "step_b"},
				},
			},
		},
	}
	return pb, poolOf(childAgent("step-a"), childAgent("step-b"))
}

// branchingWorkflowProto returns a WORKFLOW agent with an approve/reject
// graph: an agent's answer feeds a Router that sends "approve" down one
// branch and everything else down the default branch. Conditional branches
// deliberately do not converge into a Join: the barrier waits for every
// declared predecessor, and a route-skipped predecessor never fires.
func branchingWorkflowProto() (*agentsv1.Agent, AgentPool) {
	pb := &agentsv1.Agent{
		AgentId:       "review",
		Name:          "review",
		Type:          agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: []string{"classify", "approver", "rejecter"},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "classify", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "classify"},
					{Name: "decide", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_ROUTER},
					{Name: "approver", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "approver"},
					{Name: "rejecter", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "rejecter"},
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
	return pb, poolOf(childAgent("classify"), childAgent("approver"), childAgent("rejecter"))
}

// fanOutJoinWorkflowProto returns a WORKFLOW agent that fans out from one
// node to two branches over unconditional edges and re-converges through a
// Join node.
func fanOutJoinWorkflowProto() (*agentsv1.Agent, AgentPool) {
	pb := &agentsv1.Agent{
		AgentId:       "fanout",
		Name:          "fanout",
		Type:          agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: []string{"seed", "b1", "b2", "summarize"},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "seed", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "seed"},
					{Name: "b1", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "b1"},
					{Name: "b2", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "b2"},
					{Name: "gather", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_JOIN},
					{Name: "summarize", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "summarize"},
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
	return pb, poolOf(childAgent("seed"), childAgent("b1"), childAgent("b2"), childAgent("summarize"))
}

// humanInputWorkflowProto returns a WORKFLOW agent that pauses on a Human
// Input node between two agent steps: draft -> ask (human) -> publish.
func humanInputWorkflowProto() (*agentsv1.Agent, AgentPool) {
	pb := &agentsv1.Agent{
		AgentId:       "approval",
		Name:          "approval",
		Type:          agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: []string{"draft", "publish"},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "draft", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "draft"},
					{Name: "ask", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Approve this draft?"},
					{Name: "publish", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "publish"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "draft"},
					{From: "draft", To: "ask"},
					{From: "ask", To: "publish"},
				},
			},
		},
	}
	return pb, poolOf(childAgent("draft"), childAgent("publish"))
}

func TestNewFromProto_WorkflowHumanInput(t *testing.T) {
	pb, pool := humanInputWorkflowProto()
	a, err := NewFromProtoWithToolsetFactory(context.Background(), pb, workflowProviders(), nil, nil, nil, nil, nil, nil, nil, pool)
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
	pb, pool := humanInputWorkflowProto()
	pb.Config.Workflow.Nodes[1].Question = ""
	assertGraphRejected(t, pb, pool, "question")
}

func TestNewFromProto_WorkflowRouterAndJoin(t *testing.T) {
	branching, branchingPool := branchingWorkflowProto()
	fanOut, fanOutPool := fanOutJoinWorkflowProto()
	for _, tc := range []struct {
		pb   *agentsv1.Agent
		pool AgentPool
	}{{branching, branchingPool}, {fanOut, fanOutPool}} {
		a, err := NewFromProtoWithToolsetFactory(context.Background(), tc.pb, workflowProviders(), nil, nil, nil, nil, nil, nil, nil, tc.pool)
		if err != nil {
			t.Fatalf("NewFromProto(%q): %v", tc.pb.GetName(), err)
		}
		if a.Name() != tc.pb.GetName() {
			t.Errorf("agent name = %q, want %q", a.Name(), tc.pb.GetName())
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
			name: "agent node without an agent_id reference",
			mutate: func(pb *agentsv1.Agent) {
				pb.Config.Workflow.Nodes[0].AgentId = ""
				pb.Config.Workflow.Nodes[0].Agent = "step_a" // legacy name ref is never resolved
			},
			wantErr: "agent_id",
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
			pb, pool := linearWorkflowProto()
			tt.mutate(pb)
			assertGraphRejected(t, pb, pool, tt.wantErr)
		})
	}
}

// TestNewFromProto_RejectsUnresolvedWorkflowNodeID: an AGENT node whose
// agent_id does not resolve in the pool must fail at build time.
func TestNewFromProto_RejectsUnresolvedWorkflowNodeID(t *testing.T) {
	pb, pool := linearWorkflowProto()
	pb.Config.Workflow.Nodes[0].AgentId = "no-such-agent"
	if _, err := NewFromProtoWithToolsetFactory(context.Background(), pb, workflowProviders(), nil, nil, nil, nil, nil, nil, nil, pool); err == nil {
		t.Fatal("NewFromProto accepted a workflow node with an unresolved agent_id")
	} else if !strings.Contains(err.Error(), "no-such-agent") {
		t.Errorf("error %q does not mention the unresolved id", err.Error())
	}
}

// TestValidateWorkflowAgent_RejectsConditionalEdgeIntoJoin: a Join barrier
// waits for every declared predecessor, and a route-skipped predecessor
// never fires — a routed or default edge into a JOIN node produces a graph
// that hangs at runtime, so validation must reject it at save time.
func TestValidateWorkflowAgent_RejectsConditionalEdgeIntoJoin(t *testing.T) {
	t.Run("routed edge into join", func(t *testing.T) {
		pb, pool := fanOutJoinWorkflowProto()
		pb.Config.Workflow.Edges[3].Route = "left" // b1 -> gather
		assertGraphRejected(t, pb, pool, "gather")
	})
	t.Run("default edge into join", func(t *testing.T) {
		pb, pool := fanOutJoinWorkflowProto()
		pb.Config.Workflow.Edges[4].IsDefault = true // b2 -> gather
		assertGraphRejected(t, pb, pool, "gather")
	})
}

// TestValidateWorkflowAgent_RejectsNearDuplicateRouteLabels: route matching
// is trimmed and case-insensitive, so two outgoing labels that differ only
// by case or whitespace can never both be reachable — only the first would
// ever fire. Validation must reject the ambiguity.
func TestValidateWorkflowAgent_RejectsNearDuplicateRouteLabels(t *testing.T) {
	pb, pool := branchingWorkflowProto()
	pb.Config.Workflow.Edges = append(pb.Config.Workflow.Edges,
		&agentsv1.WorkflowEdge{From: "decide", To: "rejecter", Route: " Approve "})
	assertGraphRejected(t, pb, pool, "approve")
}

// TestValidateWorkflowAgent_RouterRequiresDefaultEdge: an unmatched Router
// with no default edge dead-ends silently in the ADK engine, so validation
// must require one.
func TestValidateWorkflowAgent_RouterRequiresDefaultEdge(t *testing.T) {
	pb, pool := branchingWorkflowProto()
	// Drop the default edge, keeping the router reachable and the graph
	// otherwise valid.
	pb.Config.Workflow.Edges = []*agentsv1.WorkflowEdge{
		{From: "START", To: "classify"},
		{From: "classify", To: "decide"},
		{From: "decide", To: "approver", Route: "approve"},
		{From: "decide", To: "rejecter", Route: "reject"},
	}
	assertGraphRejected(t, pb, pool, "default")
}

func assertGraphRejected(t *testing.T, pb *agentsv1.Agent, pool AgentPool, wantErr string) {
	t.Helper()

	err := ValidateWorkflowAgent(pb)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("error %q does not mention %q", err.Error(), wantErr)
	}

	// The factory must reject the same graph.
	if _, err := NewFromProtoWithToolsetFactory(context.Background(), pb, workflowProviders(), nil, nil, nil, nil, nil, nil, nil, pool); err == nil {
		t.Error("NewFromProto accepted an invalid graph")
	}
}

func TestValidateWorkflowAgent_AcceptsValidGraph(t *testing.T) {
	pb, _ := linearWorkflowProto()
	if err := ValidateWorkflowAgent(pb); err != nil {
		t.Fatalf("valid graph rejected: %v", err)
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
	pb, pool := linearWorkflowProto()

	a, err := NewFromProtoWithToolsetFactory(context.Background(), pb, workflowProviders(), nil, nil, nil, nil, nil, nil, nil, pool)
	if err != nil {
		t.Fatalf("NewFromProto: %v", err)
	}
	if a.Name() != "wf" {
		t.Errorf("agent name = %q, want %q", a.Name(), "wf")
	}
}
