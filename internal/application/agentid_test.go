package application

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// --- helpers ---

func seedAgentWithID(t *testing.T, store *memory.Store, wsID, name, agentID string) {
	t.Helper()
	if _, err := store.CreateAgent(context.Background(), wsID, &agentsv1.Agent{Name: name, AgentId: agentID}); err != nil {
		t.Fatalf("seed agent %s/%s: %v", wsID, name, err)
	}
}

func seedAgentFull(t *testing.T, store *memory.Store, wsID string, a *agentsv1.Agent) {
	t.Helper()
	if _, err := store.CreateAgent(context.Background(), wsID, a); err != nil {
		t.Fatalf("seed agent %s/%s: %v", wsID, a.GetName(), err)
	}
}

// --- tests ---

// TestMigrationRPCs_Retired locks the issue #241 cutover: the Agent ID
// rollout RPCs are gone for good, so every call answers Unimplemented no
// matter how well-formed the request is.
func TestMigrationRPCs_Retired(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	seedAgentWithID(t, store, wsTest, "my-agent", "my-agent")
	ctx := testCtx()

	t.Run("AssignAgentID", func(t *testing.T) {
		_, err := svc.AssignAgentID(ctx, connect.NewRequest(&agentsv1.AssignAgentIDRequest{
			Name:    "my-agent",
			AgentId: "my-agent",
		}))
		if connect.CodeOf(err) != connect.CodeUnimplemented {
			t.Fatalf("err = %v, want CodeUnimplemented", err)
		}
	})

	t.Run("GetMigrationReadiness", func(t *testing.T) {
		_, err := svc.GetMigrationReadiness(ctx, connect.NewRequest(&agentsv1.GetMigrationReadinessRequest{}))
		if connect.CodeOf(err) != connect.CodeUnimplemented {
			t.Fatalf("err = %v, want CodeUnimplemented", err)
		}
	})

	t.Run("MigrateAgentsV2", func(t *testing.T) {
		_, err := svc.MigrateAgentsV2(ctx, connect.NewRequest(&agentsv1.MigrateAgentsV2Request{}))
		if connect.CodeOf(err) != connect.CodeUnimplemented {
			t.Fatalf("err = %v, want CodeUnimplemented", err)
		}
	})
}

// TestCreateAgent_RequiresAgentID locks the V2 contract: the identity-less
// create path was removed, so every new agent must carry an agent_id.
func TestCreateAgent_RequiresAgentID(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())

	_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "legacy", Description: "no id"},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("err = %v, want CodeInvalidArgument (agent_id required)", err)
	}
}

// TestCreateAgent_RejectsEmbeddedSubAgents locks the V2 contract: the
// embedded sub_agents write path was removed in favor of child_agent_ids.
func TestCreateAgent_RejectsEmbeddedSubAgents(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())

	_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name:      "parent",
			AgentId:   "parent",
			SubAgents: []*agentsv1.Agent{{Name: "child"}},
		},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("err = %v, want CodeInvalidArgument (sub_agents not writable)", err)
	}
}

// TestCreateAgent_ValidationRejectsInvalidAgentID exercises the slug rules
// on the only remaining path that assigns an Agent ID: CreateAgent.
func TestCreateAgent_ValidationRejectsInvalidAgentID(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())
	ctx := testCtx()

	cases := []struct {
		agentID string
		desc    string
	}{
		{"MY-AGENT", "uppercase"},
		{"-leading", "leading hyphen"},
		{"trailing-", "trailing hyphen"},
		{"user", "reserved"},
		{"system", "reserved"},
		{"a b", "space"},
	}
	for _, tc := range cases {
		_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
			Agent: &agentsv1.Agent{Name: "my-agent", AgentId: tc.agentID},
		}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("[%s] agentID=%q: expected InvalidArgument, got %v", tc.desc, tc.agentID, err)
		}
	}
}

// TestCreateAgent_SameIDDifferentWorkspace: agent_id uniqueness is scoped to
// the workspace, so the same slug may exist in two workspaces.
func TestCreateAgent_SameIDDifferentWorkspace(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)

	seedAgentWithID(t, store, "ws-other", "agent-a", "shared-id")

	resp, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "agent-b", AgentId: "shared-id"},
	}))
	if err != nil {
		t.Fatalf("expected same ID in different workspace to succeed, got %v", err)
	}
	if resp.Msg.GetAgent().GetAgentId() != "shared-id" {
		t.Fatalf("expected shared-id, got %q", resp.Msg.GetAgent().GetAgentId())
	}
}

func TestCreateAgent_PreservesAgentID(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	resp, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "v2-leaf", AgentId: "v2-leaf"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetAgent().GetAgentId() != "v2-leaf" {
		t.Fatalf("expected agent_id preserved on V2 create, got %q", resp.Msg.GetAgent().GetAgentId())
	}
}

// TestUpdateAgent_RequiresAgentID locks the issue #241 contract: agent_id is
// the only lookup key, so a name-only update is rejected instead of falling
// back to name resolution.
func TestUpdateAgent_RequiresAgentID(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "no-id-in-request", "some-id")

	_, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "no-id-in-request", Description: "updated"},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("err = %v, want CodeInvalidArgument (agent.agent_id required)", err)
	}
}

// TestUpdateAgent_UnknownAgentIDNotFound: the record is selected by agent_id
// alone; an unknown ID is NotFound even when the name matches an existing
// agent.
func TestUpdateAgent_UnknownAgentIDNotFound(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "locked", "stable")

	_, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "locked", AgentId: "different"},
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("err = %v, want CodeNotFound (agent_id is the only lookup key)", err)
	}
}

// TestUpdateAgent_ServerPreservesName: the runtime name is server-controlled;
// a client-sent name never renames the record and never selects it.
func TestUpdateAgent_ServerPreservesName(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "canonical-name", "stable-slug")

	updated, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "client-rename", AgentId: "stable-slug", Description: "new desc"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Msg.GetAgent().GetName() != "canonical-name" {
		t.Fatalf("expected server-preserved name canonical-name, got %q", updated.Msg.GetAgent().GetName())
	}
	if updated.Msg.GetAgent().GetDescription() != "new desc" {
		t.Fatalf("expected description updated, got %q", updated.Msg.GetAgent().GetDescription())
	}
}

// TestUpdateAgent_SubAgentsReadOnly: a historical record with an embedded
// tree round-trips unchanged through UpdateAgent, mutating the embedded tree
// is rejected, and the record is deletable by its agent_id.
func TestUpdateAgent_SubAgentsReadOnly(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentFull(t, store, wsTest, &agentsv1.Agent{
		Name:        "historical",
		AgentId:     "historical",
		Description: "embedded tree",
		SubAgents:   []*agentsv1.Agent{{Name: "embedded-child"}},
	})

	updated, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{
			AgentId:     "historical",
			Description: "updated",
			SubAgents:   []*agentsv1.Agent{{Name: "embedded-child"}}, // unchanged round-trip
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Msg.GetAgent().GetDescription() != "updated" {
		t.Fatalf("expected description updated, got %q", updated.Msg.GetAgent().GetDescription())
	}

	// Mutating the embedded tree is a removed write path.
	_, err = svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{
			AgentId:     "historical",
			Description: "updated",
			SubAgents:   []*agentsv1.Agent{{Name: "embedded-child"}, {Name: "new-child"}},
		},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("err = %v, want CodeInvalidArgument (sub_agents not writable)", err)
	}

	if _, err := svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{AgentId: "historical"})); err != nil {
		t.Fatal(err)
	}
}

// TestDeleteAgent_RequiresAgentID: name-only deletion was removed with the
// name fallback (issue #241).
func TestDeleteAgent_RequiresAgentID(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "keep-me", "keep-me")

	_, err := svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{Name: "keep-me"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("err = %v, want CodeInvalidArgument (agent_id required)", err)
	}
}

// TestGetAgent_RequiresAgentID: name-only reads were removed with the name
// fallback (issue #241).
func TestGetAgent_RequiresAgentID(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "readable", "readable")

	_, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "readable"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("err = %v, want CodeInvalidArgument (agent_id required)", err)
	}
}

func TestUpdateAgent_PreservesAgentID(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "keep-id", "stable-slug")

	updated, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "keep-id", Description: "new desc", AgentId: "stable-slug"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Msg.GetAgent().GetAgentId() != "stable-slug" {
		t.Fatalf("expected agent_id preserved, got %q", updated.Msg.GetAgent().GetAgentId())
	}
	if updated.Msg.GetAgent().GetDescription() != "new desc" {
		t.Fatalf("expected description updated, got %q", updated.Msg.GetAgent().GetDescription())
	}
}
