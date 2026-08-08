package application

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

// --- fake workspace repo for role tests ---

type fakeWSRepo struct {
	members map[string]map[string]*agentsv1.WorkspaceMember
}

func newFakeWSRepo() *fakeWSRepo {
	return &fakeWSRepo{members: make(map[string]map[string]*agentsv1.WorkspaceMember)}
}

func (f *fakeWSRepo) addMember(wsID, userID, role string) {
	if f.members[wsID] == nil {
		f.members[wsID] = make(map[string]*agentsv1.WorkspaceMember)
	}
	f.members[wsID][userID] = &agentsv1.WorkspaceMember{
		WorkspaceId: wsID,
		UserId:      userID,
		Role:        role,
	}
}

func (f *fakeWSRepo) GetMember(_ context.Context, wsID, userID string) (*agentsv1.WorkspaceMember, error) {
	if m, ok := f.members[wsID][userID]; ok {
		return proto.Clone(m).(*agentsv1.WorkspaceMember), nil
	}
	return nil, workspacerepo.ErrNotFound
}

func (f *fakeWSRepo) EnsureIndexes(context.Context) error                   { return nil }
func (f *fakeWSRepo) ListWorkspaces(context.Context) ([]*agentsv1.Workspace, error) {
	return nil, nil
}
func (f *fakeWSRepo) GetWorkspace(context.Context, string) (*agentsv1.Workspace, error) {
	return nil, nil
}
func (f *fakeWSRepo) GetWorkspaceBySlug(context.Context, string) (*agentsv1.Workspace, error) {
	return nil, nil
}
func (f *fakeWSRepo) CreateWorkspace(context.Context, *agentsv1.Workspace) (*agentsv1.Workspace, error) {
	return nil, nil
}
func (f *fakeWSRepo) UpdateWorkspace(context.Context, *agentsv1.Workspace) (*agentsv1.Workspace, error) {
	return nil, nil
}
func (f *fakeWSRepo) DeleteWorkspace(context.Context, string) error { return nil }
func (f *fakeWSRepo) CountWorkspaces(context.Context) (int64, error) {
	return 0, nil
}
func (f *fakeWSRepo) ListMembers(context.Context, string) ([]*agentsv1.WorkspaceMember, error) {
	return nil, nil
}
func (f *fakeWSRepo) ListMembershipsForUser(context.Context, string) ([]*agentsv1.WorkspaceMember, error) {
	return nil, nil
}
func (f *fakeWSRepo) IsMember(context.Context, string, string) (bool, error) {
	return false, nil
}
func (f *fakeWSRepo) AddMember(context.Context, *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error) {
	return nil, nil
}
func (f *fakeWSRepo) UpdateMember(context.Context, *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error) {
	return nil, nil
}
func (f *fakeWSRepo) RemoveMember(context.Context, string, string) error { return nil }

// --- helpers ---

func ctxOwner(wsID, userID string) context.Context {
	ctx := workspace.WithID(context.Background(), wsID)
	return auth.WithAuthenticated(ctx, &agentsv1.User{Id: userID, Role: "user"}, nil)
}

func ctxAdmin(wsID string) context.Context {
	ctx := workspace.WithID(context.Background(), wsID)
	return auth.WithAdmin(ctx)
}

func seedAgent(t *testing.T, store *memory.Store, wsID, name string) {
	t.Helper()
	if _, err := store.CreateAgent(context.Background(), wsID, &agentsv1.Agent{Name: name}); err != nil {
		t.Fatalf("seed agent %s/%s: %v", wsID, name, err)
	}
}

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

func testAdminCtx() context.Context {
	return auth.WithAdmin(testCtx())
}

func TestAssignAgentID_Success(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	seedAgent(t, store, wsTest, "my-agent")

	resp, err := svc.AssignAgentID(testAdminCtx(), connect.NewRequest(&agentsv1.AssignAgentIDRequest{
		Name:    "my-agent",
		AgentId: "my-agent",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetAgent().GetAgentId() != "my-agent" {
		t.Fatalf("expected agent_id=my-agent, got %q", resp.Msg.GetAgent().GetAgentId())
	}

	got, err := store.GetAgent(context.Background(), wsTest, "my-agent")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetAgentId() != "my-agent" {
		t.Fatalf("persisted agent_id: expected my-agent, got %q", got.GetAgentId())
	}
}

func TestAssignAgentID_Immutability(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	seedAgentWithID(t, store, wsTest, "my-agent", "existing-id")

	_, err := svc.AssignAgentID(testAdminCtx(), connect.NewRequest(&agentsv1.AssignAgentIDRequest{
		Name:    "my-agent",
		AgentId: "new-id",
	}))
	cerr, ok := err.(*connect.Error)
	if !ok || cerr.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition for immutable reassignment, got %v", err)
	}
}

func TestAssignAgentID_ValidationRejectsInvalid(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	seedAgent(t, store, wsTest, "my-agent")

	cases := []struct {
		agentID string
		desc    string
	}{
		{"", "empty"},
		{"MY-AGENT", "uppercase"},
		{"-leading", "leading hyphen"},
		{"trailing-", "trailing hyphen"},
		{"user", "reserved"},
		{"system", "reserved"},
		{"a b", "space"},
	}
	for _, tc := range cases {
		_, err := svc.AssignAgentID(testAdminCtx(), connect.NewRequest(&agentsv1.AssignAgentIDRequest{
			Name:    "my-agent",
			AgentId: tc.agentID,
		}))
		cerr, ok := err.(*connect.Error)
		if !ok || cerr.Code() != connect.CodeInvalidArgument {
			t.Errorf("[%s] agentID=%q: expected InvalidArgument, got %v", tc.desc, tc.agentID, err)
		}
	}
}

func TestAssignAgentID_DuplicateInSameWorkspace(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	seedAgentWithID(t, store, wsTest, "agent-a", "taken-id")
	seedAgent(t, store, wsTest, "agent-b")

	_, err := svc.AssignAgentID(testAdminCtx(), connect.NewRequest(&agentsv1.AssignAgentIDRequest{
		Name:    "agent-b",
		AgentId: "taken-id",
	}))
	cerr, ok := err.(*connect.Error)
	if !ok || cerr.Code() != connect.CodeAlreadyExists {
		t.Fatalf("expected AlreadyExists for duplicate agent_id, got %v", err)
	}
}

func TestAssignAgentID_SameIDDifferentWorkspace(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)

	seedAgentWithID(t, store, "ws-other", "agent-a", "shared-id")
	seedAgent(t, store, wsTest, "agent-b")

	resp, err := svc.AssignAgentID(testAdminCtx(), connect.NewRequest(&agentsv1.AssignAgentIDRequest{
		Name:    "agent-b",
		AgentId: "shared-id",
	}))
	if err != nil {
		t.Fatalf("expected same ID in different workspace to succeed, got %v", err)
	}
	if resp.Msg.GetAgent().GetAgentId() != "shared-id" {
		t.Fatalf("expected shared-id, got %q", resp.Msg.GetAgent().GetAgentId())
	}
}

func TestAssignAgentID_NotFound(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)

	_, err := svc.AssignAgentID(testAdminCtx(), connect.NewRequest(&agentsv1.AssignAgentIDRequest{
		Name:    "nonexistent",
		AgentId: "some-id",
	}))
	cerr, ok := err.(*connect.Error)
	if !ok || cerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestAssignAgentID_RequiresName(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())

	_, err := svc.AssignAgentID(testAdminCtx(), connect.NewRequest(&agentsv1.AssignAgentIDRequest{
		AgentId: "some-id",
	}))
	cerr, ok := err.(*connect.Error)
	if !ok || cerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for missing name, got %v", err)
	}
}

func TestAssignAgentID_Authorization(t *testing.T) {
	store := memory.New()
	wsRepo := newFakeWSRepo()
	wsRepo.addMember(wsTest, "u-owner", "owner")
	wsRepo.addMember(wsTest, "u-admin", "admin")
	wsRepo.addMember(wsTest, "u-member", "member")

	svc := NewAgentServiceServer(store)
	svc.SetWorkspaceRepo(wsRepo)

	t.Run("member denied", func(t *testing.T) {
		seedAgent(t, store, wsTest, "a-member")
		_, err := svc.AssignAgentID(
			ctxOwner(wsTest, "u-member"),
			connect.NewRequest(&agentsv1.AssignAgentIDRequest{
				Name:    "a-member",
				AgentId: "member-id",
			}),
		)
		cerr, ok := err.(*connect.Error)
		if !ok || cerr.Code() != connect.CodePermissionDenied {
			t.Fatalf("expected PermissionDenied for member, got %v", err)
		}
	})

	t.Run("owner allowed", func(t *testing.T) {
		seedAgent(t, store, wsTest, "a-owner")
		_, err := svc.AssignAgentID(
			ctxOwner(wsTest, "u-owner"),
			connect.NewRequest(&agentsv1.AssignAgentIDRequest{
				Name:    "a-owner",
				AgentId: "owner-id",
			}),
		)
		if err != nil {
			t.Fatalf("expected owner to succeed, got %v", err)
		}
	})

	t.Run("admin allowed", func(t *testing.T) {
		seedAgent(t, store, wsTest, "a-admin")
		_, err := svc.AssignAgentID(
			ctxOwner(wsTest, "u-admin"),
			connect.NewRequest(&agentsv1.AssignAgentIDRequest{
				Name:    "a-admin",
				AgentId: "admin-id",
			}),
		)
		if err != nil {
			t.Fatalf("expected admin to succeed, got %v", err)
		}
	})

	t.Run("global admin allowed", func(t *testing.T) {
		seedAgent(t, store, wsTest, "a-global")
		_, err := svc.AssignAgentID(
			ctxAdmin(wsTest),
			connect.NewRequest(&agentsv1.AssignAgentIDRequest{
				Name:    "a-global",
				AgentId: "global-id",
			}),
		)
		if err != nil {
			t.Fatalf("expected global admin to succeed, got %v", err)
		}
	})
}

func TestAssignAgentID_ReloadErrorRollsBack(t *testing.T) {
	store := memory.New()
	seedAgent(t, store, wsTest, "rollback-agent")

	svc := NewAgentServiceServer(store)
	svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

	_, err := svc.AssignAgentID(testAdminCtx(), connect.NewRequest(&agentsv1.AssignAgentIDRequest{
		Name:    "rollback-agent",
		AgentId: "rollback-id",
	}))
	cerr, ok := err.(*connect.Error)
	if !ok || cerr.Code() != connect.CodeInternal {
		t.Fatalf("expected Internal, got %v", err)
	}

	agent, err := store.GetAgent(context.Background(), wsTest, "rollback-agent")
	if err != nil {
		t.Fatal(err)
	}
	if agent.GetAgentId() != "" {
		t.Fatalf("expected rollback to clear agent_id, got %q", agent.GetAgentId())
	}
}

func TestGetMigrationReadiness_AllStatuses(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)

	seedAgentWithID(t, store, wsTest, "ready-agent", "ready-agent")
	seedAgent(t, store, wsTest, "missing-agent")
	seedAgentWithID(t, store, wsTest, "parent-agent", "parent-agent")

	ctx := testCtx()

	// Create parent-agent with sub-agent reference to missing-agent
	parent, _ := store.GetAgent(ctx, wsTest, "parent-agent")
	parent.SubAgents = []*agentsv1.Agent{{Name: "missing-agent"}}
	if _, err := store.UpdateAgent(ctx, wsTest, parent); err != nil {
		t.Fatal(err)
	}

	resp, err := svc.GetMigrationReadiness(ctx, connect.NewRequest(&agentsv1.GetMigrationReadinessRequest{}))
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]*agentsv1.AgentMigrationStatus)
	for _, s := range resp.Msg.GetStatuses() {
		byName[s.GetName()] = s
	}

	if s := byName["ready-agent"]; s.GetReadiness() != agentsv1.MigrationReadiness_MIGRATION_READINESS_READY {
		t.Errorf("ready-agent: expected READY, got %v", s.GetReadiness())
	}
	if s := byName["missing-agent"]; s.GetReadiness() != agentsv1.MigrationReadiness_MIGRATION_READINESS_MISSING_ID {
		t.Errorf("missing-agent: expected MISSING_ID, got %v", s.GetReadiness())
	}
	if s := byName["parent-agent"]; s.GetReadiness() != agentsv1.MigrationReadiness_MIGRATION_READINESS_INCOMPLETE_DEPS {
		t.Errorf("parent-agent: expected INCOMPLETE_DEPS, got %v (%s)", s.GetReadiness(), s.GetDetail())
	}
}

func TestGetMigrationReadiness_Empty(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())

	resp, err := svc.GetMigrationReadiness(testCtx(), connect.NewRequest(&agentsv1.GetMigrationReadinessRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Msg.GetStatuses()) != 0 {
		t.Fatalf("expected 0 statuses for empty workspace, got %d", len(resp.Msg.GetStatuses()))
	}
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

// TestUpdateAgent_LegacyRecordsStayEditable ensures a pre-migration record
// (stored without an agent_id, with an embedded tree) can still round-trip
// through UpdateAgent unchanged and be deleted by name, while mutating the
// embedded tree is rejected.
func TestUpdateAgent_LegacyRecordsStayEditable(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	if _, err := store.CreateAgent(context.Background(), wsTest, &agentsv1.Agent{
		Name:        "legacy",
		Description: "no id",
		SubAgents:   []*agentsv1.Agent{{Name: "embedded-child"}},
	}); err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{
			Name:        "legacy",
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
			Name:        "legacy",
			Description: "updated",
			SubAgents:   []*agentsv1.Agent{{Name: "embedded-child"}, {Name: "new-child"}},
		},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("err = %v, want CodeInvalidArgument (sub_agents not writable)", err)
	}

	if _, err := svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{Name: "legacy"})); err != nil {
		t.Fatal(err)
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

func TestUpdateAgent_RejectsAgentIDChange(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgentWithID(t, store, wsTest, "locked", "stable")

	_, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "locked", AgentId: "different"},
	}))
	cerr, ok := err.(*connect.Error)
	if !ok || cerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for changing agent_id, got %v", err)
	}
}

func TestUpdateAgent_CannotSetAgentIDViaCRUD(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	seedAgent(t, store, wsTest, "no-id")

	_, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "no-id", AgentId: "sneaky", Description: "updated"},
	}))
	cerr, ok := err.(*connect.Error)
	if !ok || cerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument when setting agent_id via UpdateAgent, got %v", err)
	}
}

func TestGetMigrationReadiness_Conflict(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)

	// Directly seed two agents with the same agent_id to simulate conflict
	seedAgentWithID(t, store, wsTest, "conflict-a", "same-id")
	seedAgentWithID(t, store, wsTest, "conflict-b", "same-id")

	resp, err := svc.GetMigrationReadiness(testCtx(), connect.NewRequest(&agentsv1.GetMigrationReadinessRequest{}))
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range resp.Msg.GetStatuses() {
		if s.GetReadiness() != agentsv1.MigrationReadiness_MIGRATION_READINESS_CONFLICT {
			t.Errorf("%s: expected CONFLICT, got %v", s.GetName(), s.GetReadiness())
		}
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
