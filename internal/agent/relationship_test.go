package agent

import (
	"strings"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func agentProto(id, ws string, children ...string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:          id,
		AgentId:       id,
		WorkspaceId:   ws,
		ChildAgentIds: children,
	}
}

func TestValidateAgentRelationships_ValidTree(t *testing.T) {
	pool := []*agentsv1.Agent{
		agentProto("parent", "ws1", "child-a", "child-b"),
		agentProto("child-a", "ws1"),
		agentProto("child-b", "ws1"),
	}
	if err := ValidateAgentRelationships(pool[0], pool); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidateAgentRelationships_MissingChild(t *testing.T) {
	pool := []*agentsv1.Agent{
		agentProto("parent", "ws1", "missing"),
	}
	err := ValidateAgentRelationships(pool[0], pool)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got: %v", err)
	}
}

func TestValidateAgentRelationships_CrossWorkspace(t *testing.T) {
	pool := []*agentsv1.Agent{
		agentProto("parent", "ws1", "child"),
		agentProto("child", "ws2"),
	}
	err := ValidateAgentRelationships(pool[0], pool)
	if err == nil || !strings.Contains(err.Error(), "workspace") {
		t.Fatalf("expected cross-workspace error, got: %v", err)
	}
}

func TestValidateAgentRelationships_SelfReference(t *testing.T) {
	pool := []*agentsv1.Agent{
		agentProto("loop", "ws1", "loop"),
	}
	err := ValidateAgentRelationships(pool[0], pool)
	if err == nil || !strings.Contains(err.Error(), "self-reference") {
		t.Fatalf("expected self-reference error, got: %v", err)
	}
}

func TestValidateAgentRelationships_Cycle(t *testing.T) {
	pool := []*agentsv1.Agent{
		agentProto("a", "ws1", "b"),
		agentProto("b", "ws1", "c"),
		agentProto("c", "ws1", "a"),
	}
	err := ValidateAgentRelationships(pool[0], pool)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got: %v", err)
	}
}

func TestValidateAgentRelationships_MultiParent(t *testing.T) {
	shared := agentProto("shared", "ws1")
	parentA := agentProto("parent-a", "ws1", "shared")
	parentB := agentProto("parent-b", "ws1", "shared")
	pool := []*agentsv1.Agent{parentA, parentB, shared}

	err := ValidateAgentRelationships(parentB, pool)
	if err == nil || !strings.Contains(err.Error(), "single-parent") {
		t.Fatalf("expected single-parent error, got: %v", err)
	}
}

func TestValidateAgentRelationships_WorkflowAgentIDRef(t *testing.T) {
	child := agentProto("step", "ws1")
	parent := &agentsv1.Agent{
		Name:        "wf",
		AgentId:     "wf",
		WorkspaceId: "ws1",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "step-node", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "step"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "step-node"},
				},
			},
		},
	}
	pool := []*agentsv1.Agent{parent, child}
	if err := ValidateAgentRelationships(parent, pool); err != nil {
		t.Fatalf("expected valid workflow ref, got: %v", err)
	}
}

func TestValidateAgentRelationships_WorkflowMissingRef(t *testing.T) {
	parent := &agentsv1.Agent{
		Name:        "wf",
		AgentId:     "wf",
		WorkspaceId: "ws1",
		Type:        agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "step-node", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "ghost"},
				},
			},
		},
	}
	pool := []*agentsv1.Agent{parent}
	err := ValidateAgentRelationships(parent, pool)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error for workflow ref, got: %v", err)
	}
}

func TestValidateAgentRelationships_NoChildren(t *testing.T) {
	agent := agentProto("leaf", "ws1")
	if err := ValidateAgentRelationships(agent, []*agentsv1.Agent{agent}); err != nil {
		t.Fatalf("expected valid for leaf agent, got: %v", err)
	}
}

func TestValidateNoOrphanedReferences_Blocked(t *testing.T) {
	pool := []*agentsv1.Agent{
		agentProto("parent", "ws1", "child"),
		agentProto("child", "ws1"),
	}
	err := ValidateNoOrphanedReferences("child", pool)
	if err == nil || !strings.Contains(err.Error(), "referenced by") {
		t.Fatalf("expected blocked-by-reference error, got: %v", err)
	}
}

func TestValidateNoOrphanedReferences_Safe(t *testing.T) {
	pool := []*agentsv1.Agent{
		agentProto("parent", "ws1"),
		agentProto("child", "ws1"),
	}
	if err := ValidateNoOrphanedReferences("child", pool); err != nil {
		t.Fatalf("expected safe deletion, got: %v", err)
	}
}
