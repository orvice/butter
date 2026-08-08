package application

import (
	"strings"
	"testing"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestV2_CreateWithChildAgentIDs(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "child-a", "child-a")
	seedAgentWithID(t, store, wsTest, "child-b", "child-b")

	resp, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name:          "parent",
			AgentId:       "parent",
			ChildAgentIds: []string{"child-a", "child-b"},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	a := resp.Msg.GetAgent()
	if a.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
		t.Fatalf("expected ACTIVE, got %v", a.GetLifecycleStatus())
	}
	if len(a.GetChildAgentIds()) != 2 {
		t.Fatalf("expected 2 child_agent_ids, got %d", len(a.GetChildAgentIds()))
	}
	if len(a.GetSubAgents()) != 0 {
		t.Fatalf("expected sub_agents to be stripped, got %d", len(a.GetSubAgents()))
	}
}

func TestV2_CreateRequiresAgentIDWhenChildrenSet(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "child", "child")

	_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name:          "parent",
			ChildAgentIds: []string{"child"},
		},
	}))
	if err == nil {
		t.Fatal("expected error when agent_id missing with child_agent_ids")
	}
	if !strings.Contains(err.Error(), "agent_id") {
		t.Fatalf("expected agent_id error, got: %v", err)
	}
}

func TestV2_CreateWithAgentIDNoChildren(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	resp, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "leaf", AgentId: "leaf", Description: "v2 leaf"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetAgent().GetAgentId() != "leaf" {
		t.Fatalf("expected agent_id preserved, got %q", resp.Msg.GetAgent().GetAgentId())
	}
}

func TestV2_CreateRejectsDuplicateAgentID(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "existing", "taken-id")

	_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "new-agent", AgentId: "taken-id"},
	}))
	if err == nil {
		t.Fatal("expected error for duplicate agent_id")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodeAlreadyExists {
		t.Fatalf("expected AlreadyExists, got: %v", err)
	}
}

func TestV2_CreateRejectsMissingChild(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name:          "parent",
			AgentId:       "parent",
			ChildAgentIds: []string{"nonexistent"},
		},
	}))
	if err == nil {
		t.Fatal("expected error for missing child reference")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got: %v", err)
	}
}

func TestV2_DeleteBlockedByReference(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "child", "child")
	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:          "parent",
		AgentId:       "parent",
		ChildAgentIds: []string{"child"},
	})

	_, err := svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{Name: "child"}))
	if err == nil {
		t.Fatal("expected error when deleting referenced child")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got: %v", err)
	}
}

func TestV2_DeleteUnreferencedSucceeds(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "leaf", "leaf")

	if _, err := svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{Name: "leaf"})); err != nil {
		t.Fatal(err)
	}
}

func TestV2_UpdatePreservesImmutableFields(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:            "a1",
		AgentId:         "a1",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
		LegacyName:      "old-name",
	})

	resp, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1", DisplayName: "New Display"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	updated := resp.Msg.GetAgent()
	if updated.GetAgentId() != "a1" {
		t.Fatalf("expected agent_id=a1 preserved, got %q", updated.GetAgentId())
	}
	if updated.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
		t.Fatalf("expected lifecycle_status preserved, got %v", updated.GetLifecycleStatus())
	}
	if updated.GetLegacyName() != "old-name" {
		t.Fatalf("expected legacy_name preserved, got %q", updated.GetLegacyName())
	}
}

func TestV2_UpdateValidatesRelationships(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "child", "child")
	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:          "parent",
		AgentId:       "parent",
		ChildAgentIds: []string{"child"},
	})

	_, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{
			Name:          "parent",
			ChildAgentIds: []string{"nonexistent"},
		},
	}))
	if err == nil {
		t.Fatal("expected error for invalid child reference on update")
	}
}
