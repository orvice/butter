package application

import (
	"strings"
	"testing"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestMigrateV2_DryRunReportsExpandable(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:    "root",
		AgentId: "root",
		SubAgents: []*agentsv1.Agent{
			{Name: "sub1", AgentId: "sub1"},
			{Name: "sub2", AgentId: "sub2"},
		},
	})

	resp, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_DRY_RUN,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetMode() != agentsv1.MigrateMode_MIGRATE_MODE_DRY_RUN {
		t.Fatalf("expected DRY_RUN mode, got %v", resp.Msg.GetMode())
	}
	found := false
	for _, r := range resp.Msg.GetResults() {
		if r.GetName() == "root" && r.GetAction() == "expandable" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected root agent to be reported as expandable")
	}
}

func TestMigrateV2_DryRunReportsMissingID(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name: "no-id-root",
		SubAgents: []*agentsv1.Agent{
			{Name: "child-no-id"},
		},
	})

	resp, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_DRY_RUN,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetErrors() == 0 {
		t.Fatal("expected errors for missing IDs")
	}
	hasMissingID := false
	for _, r := range resp.Msg.GetResults() {
		if r.GetAction() == "missing_id" {
			hasMissingID = true
		}
	}
	if !hasMissingID {
		t.Fatal("expected missing_id action in results")
	}
}

func TestMigrateV2_ApplyExpandsTree(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:    "root",
		AgentId: "root",
		SubAgents: []*agentsv1.Agent{
			{Name: "sub1", AgentId: "sub1", Description: "first child"},
			{Name: "sub2", AgentId: "sub2", Description: "second child"},
		},
	})

	resp, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_APPLY,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetMigrated() == 0 {
		t.Fatal("expected at least one migrated agent")
	}

	getResp, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "root"}))
	if err != nil {
		t.Fatal(err)
	}
	root := getResp.Msg.GetAgent()
	if len(root.GetSubAgents()) != 0 {
		t.Fatalf("expected sub_agents cleared, got %d", len(root.GetSubAgents()))
	}
	if len(root.GetChildAgentIds()) != 2 {
		t.Fatalf("expected 2 child_agent_ids, got %d", len(root.GetChildAgentIds()))
	}
	if root.GetChildAgentIds()[0] != "sub1" || root.GetChildAgentIds()[1] != "sub2" {
		t.Fatalf("unexpected child_agent_ids: %v", root.GetChildAgentIds())
	}

	for _, childName := range []string{"sub1", "sub2"} {
		childResp, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: childName}))
		if err != nil {
			t.Fatalf("child %q should exist as independent: %v", childName, err)
		}
		child := childResp.Msg.GetAgent()
		if child.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
			t.Fatalf("child %q expected ACTIVE, got %v", childName, child.GetLifecycleStatus())
		}
		if child.GetDisplayName() != childName {
			t.Fatalf("child %q expected display_name=%q, got %q", childName, childName, child.GetDisplayName())
		}
	}
}

func TestMigrateV2_ApplyConvertsWorkflowRefs(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:    "wf-root",
		AgentId: "wf-root",
		Type:    agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		SubAgents: []*agentsv1.Agent{
			{Name: "step", AgentId: "step"},
		},
		Config: &agentsv1.AgentConfig{
			Workflow: &agentsv1.WorkflowConfig{
				Nodes: []*agentsv1.WorkflowNode{
					{Name: "step-node", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, Agent: "step"},
				},
				Edges: []*agentsv1.WorkflowEdge{
					{From: "START", To: "step-node"},
				},
			},
		},
	})

	_, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_APPLY,
	}))
	if err != nil {
		t.Fatal(err)
	}

	getResp, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "wf-root"}))
	if err != nil {
		t.Fatal(err)
	}
	wf := getResp.Msg.GetAgent().GetConfig().GetWorkflow()
	for _, node := range wf.GetNodes() {
		if node.GetKind() == agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT {
			if node.GetAgentId() != "step" {
				t.Fatalf("expected workflow node agent_id=step, got %q", node.GetAgentId())
			}
		}
	}
}

func TestMigrateV2_VerifyCleanState(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentWithID(t, store, wsTest, "clean", "clean")

	resp, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_VERIFY,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetErrors() != 0 {
		t.Fatalf("expected 0 errors for clean state, got %d", resp.Msg.GetErrors())
	}
}

func TestMigrateV2_VerifyReportsEmbeddedSubAgents(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:    "unmigrated",
		AgentId: "unmigrated",
		SubAgents: []*agentsv1.Agent{
			{Name: "embedded", AgentId: "embedded"},
		},
	})

	resp, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_VERIFY,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetErrors() == 0 {
		t.Fatal("expected errors for unmigrated agent")
	}
	hasSubErr := false
	for _, r := range resp.Msg.GetResults() {
		if r.GetAction() == "error" && strings.Contains(r.GetDetail(), "sub_agents") {
			hasSubErr = true
		}
	}
	if !hasSubErr {
		t.Fatal("expected error about embedded sub_agents")
	}
}

func TestMigrateV2_InvalidMode(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())
	ctx := testAdminCtx()

	_, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{}))
	if err == nil {
		t.Fatal("expected error for unspecified mode")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument, got: %v", err)
	}
}

func TestMigrateV2_ApplyExpandsNestedTree(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:    "root",
		AgentId: "root",
		SubAgents: []*agentsv1.Agent{
			{
				Name:    "child",
				AgentId: "child",
				SubAgents: []*agentsv1.Agent{
					{Name: "grandchild", AgentId: "grandchild", Description: "leaf"},
				},
			},
		},
	})

	resp, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_APPLY,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetMigrated() == 0 {
		t.Fatal("expected root to be migrated")
	}
	if resp.Msg.GetErrors() != 0 {
		t.Fatalf("expected no errors, got %d", resp.Msg.GetErrors())
	}

	gcResp, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "grandchild"}))
	if err != nil {
		t.Fatal("grandchild should exist as independent agent:", err)
	}
	gc := gcResp.Msg.GetAgent()
	if gc.GetAgentId() != "grandchild" {
		t.Fatalf("expected grandchild agent_id, got %q", gc.GetAgentId())
	}
	if len(gc.GetSubAgents()) != 0 {
		t.Fatalf("expected grandchild sub_agents cleared, got %d", len(gc.GetSubAgents()))
	}

	childResp, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "child"}))
	if err != nil {
		t.Fatal("child should exist as independent agent:", err)
	}
	child := childResp.Msg.GetAgent()
	if len(child.GetChildAgentIds()) != 1 || child.GetChildAgentIds()[0] != "grandchild" {
		t.Fatalf("expected child to have child_agent_ids=[grandchild], got %v", child.GetChildAgentIds())
	}
	if len(child.GetSubAgents()) != 0 {
		t.Fatalf("expected child sub_agents cleared, got %d", len(child.GetSubAgents()))
	}

	rootResp, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "root"}))
	if err != nil {
		t.Fatal(err)
	}
	root := rootResp.Msg.GetAgent()
	if len(root.GetChildAgentIds()) != 1 || root.GetChildAgentIds()[0] != "child" {
		t.Fatalf("expected root child_agent_ids=[child], got %v", root.GetChildAgentIds())
	}
}

func TestMigrateV2_DryRunReportsNestedMissingID(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:    "root",
		AgentId: "root",
		SubAgents: []*agentsv1.Agent{
			{
				Name:    "child",
				AgentId: "child",
				SubAgents: []*agentsv1.Agent{
					{Name: "grandchild-no-id"},
				},
			},
		},
	})

	resp, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_DRY_RUN,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetErrors() == 0 {
		t.Fatal("expected errors for nested missing ID")
	}
	hasMissing := false
	for _, r := range resp.Msg.GetResults() {
		if r.GetName() == "grandchild-no-id" && r.GetAction() == "missing_id" {
			hasMissing = true
		}
	}
	if !hasMissing {
		t.Fatal("expected missing_id for grandchild-no-id")
	}
}

func TestMigrateV2_ApplyChildMissingIDBlocksParent(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testAdminCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:    "root",
		AgentId: "root",
		SubAgents: []*agentsv1.Agent{
			{Name: "no-id-child"},
		},
	})

	resp, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{
		Mode: agentsv1.MigrateMode_MIGRATE_MODE_APPLY,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetErrors() == 0 {
		t.Fatal("expected errors for missing child ID")
	}

	rootResp, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "root"}))
	if err != nil {
		t.Fatal(err)
	}
	root := rootResp.Msg.GetAgent()
	if len(root.GetSubAgents()) == 0 {
		t.Fatal("root sub_agents should be preserved when migration is blocked")
	}
}
