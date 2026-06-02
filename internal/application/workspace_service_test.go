package application

import (
	"context"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/auth"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeWorkspaceRepo is a minimal in-memory workspacerepo.Repository for tests.
type fakeWorkspaceRepo struct {
	mu       sync.Mutex
	wsByID   map[string]*agentsv1.Workspace
	wsBySlug map[string]string
	members  map[string]map[string]*agentsv1.WorkspaceMember // workspaceID → userID → member
}

func newFakeWorkspaceRepo() *fakeWorkspaceRepo {
	return &fakeWorkspaceRepo{
		wsByID:   map[string]*agentsv1.Workspace{},
		wsBySlug: map[string]string{},
		members:  map[string]map[string]*agentsv1.WorkspaceMember{},
	}
}

func (f *fakeWorkspaceRepo) EnsureIndexes(context.Context) error { return nil }

func (f *fakeWorkspaceRepo) ListWorkspaces(context.Context) ([]*agentsv1.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*agentsv1.Workspace, 0, len(f.wsByID))
	for _, ws := range f.wsByID {
		out = append(out, proto.Clone(ws).(*agentsv1.Workspace))
	}
	return out, nil
}

func (f *fakeWorkspaceRepo) GetWorkspace(_ context.Context, id string) (*agentsv1.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	ws, ok := f.wsByID[id]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	return proto.Clone(ws).(*agentsv1.Workspace), nil
}

func (f *fakeWorkspaceRepo) GetWorkspaceBySlug(_ context.Context, slug string) (*agentsv1.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.wsBySlug[slug]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	return proto.Clone(f.wsByID[id]).(*agentsv1.Workspace), nil
}

func (f *fakeWorkspaceRepo) CreateWorkspace(_ context.Context, ws *agentsv1.Workspace) (*agentsv1.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.wsBySlug[ws.GetSlug()]; ok {
		return nil, workspacerepo.ErrAlreadyExists
	}
	f.wsByID[ws.GetId()] = proto.Clone(ws).(*agentsv1.Workspace)
	f.wsBySlug[ws.GetSlug()] = ws.GetId()
	return proto.Clone(ws).(*agentsv1.Workspace), nil
}

func (f *fakeWorkspaceRepo) UpdateWorkspace(_ context.Context, ws *agentsv1.Workspace) (*agentsv1.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.wsByID[ws.GetId()]; !ok {
		return nil, workspacerepo.ErrNotFound
	}
	f.wsByID[ws.GetId()] = proto.Clone(ws).(*agentsv1.Workspace)
	return proto.Clone(ws).(*agentsv1.Workspace), nil
}

func (f *fakeWorkspaceRepo) DeleteWorkspace(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	ws, ok := f.wsByID[id]
	if !ok {
		return workspacerepo.ErrNotFound
	}
	delete(f.wsBySlug, ws.GetSlug())
	delete(f.wsByID, id)
	delete(f.members, id)
	return nil
}

func (f *fakeWorkspaceRepo) CountWorkspaces(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return int64(len(f.wsByID)), nil
}

func (f *fakeWorkspaceRepo) ListMembers(_ context.Context, workspaceID string) ([]*agentsv1.WorkspaceMember, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []*agentsv1.WorkspaceMember{}
	for _, m := range f.members[workspaceID] {
		out = append(out, proto.Clone(m).(*agentsv1.WorkspaceMember))
	}
	return out, nil
}

func (f *fakeWorkspaceRepo) ListMembershipsForUser(_ context.Context, userID string) ([]*agentsv1.WorkspaceMember, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*agentsv1.WorkspaceMember
	for _, byUser := range f.members {
		if m, ok := byUser[userID]; ok {
			out = append(out, proto.Clone(m).(*agentsv1.WorkspaceMember))
		}
	}
	return out, nil
}

func (f *fakeWorkspaceRepo) IsMember(_ context.Context, workspaceID, userID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.members[workspaceID][userID]
	return ok, nil
}

func (f *fakeWorkspaceRepo) GetMember(_ context.Context, workspaceID, userID string) (*agentsv1.WorkspaceMember, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.members[workspaceID][userID]
	if !ok {
		return nil, workspacerepo.ErrNotFound
	}
	return proto.Clone(m).(*agentsv1.WorkspaceMember), nil
}

func (f *fakeWorkspaceRepo) AddMember(_ context.Context, m *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.members[m.GetWorkspaceId()] == nil {
		f.members[m.GetWorkspaceId()] = map[string]*agentsv1.WorkspaceMember{}
	}
	f.members[m.GetWorkspaceId()][m.GetUserId()] = proto.Clone(m).(*agentsv1.WorkspaceMember)
	return proto.Clone(m).(*agentsv1.WorkspaceMember), nil
}

func (f *fakeWorkspaceRepo) UpdateMember(_ context.Context, m *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.members[m.GetWorkspaceId()][m.GetUserId()]; !ok {
		return nil, workspacerepo.ErrNotFound
	}
	f.members[m.GetWorkspaceId()][m.GetUserId()] = proto.Clone(m).(*agentsv1.WorkspaceMember)
	return proto.Clone(m).(*agentsv1.WorkspaceMember), nil
}

func (f *fakeWorkspaceRepo) RemoveMember(_ context.Context, workspaceID, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.members[workspaceID][userID]; !ok {
		return workspacerepo.ErrNotFound
	}
	delete(f.members[workspaceID], userID)
	return nil
}

// seed prepares one workspace ("ws-a") with one owner ("u-owner") and one
// regular member ("u-member"), plus an outsider workspace ("ws-other").
func (f *fakeWorkspaceRepo) seed() {
	ctx := context.Background()
	_, _ = f.CreateWorkspace(ctx, &agentsv1.Workspace{Id: "ws-a", Slug: "ws-a", Name: "A"})
	_, _ = f.CreateWorkspace(ctx, &agentsv1.Workspace{Id: "ws-other", Slug: "ws-other", Name: "Other"})
	_, _ = f.AddMember(ctx, &agentsv1.WorkspaceMember{WorkspaceId: "ws-a", UserId: "u-owner", Role: "owner"})
	_, _ = f.AddMember(ctx, &agentsv1.WorkspaceMember{WorkspaceId: "ws-a", UserId: "u-member", Role: "member"})
}

func ctxAsUser(role, id string) context.Context {
	ctx := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: id, Role: role}, &auth.Session{})
	if role == "admin" {
		ctx = auth.WithAdmin(ctx)
	}
	return ctx
}

func TestWorkspaceService_GetWorkspace_NonMemberDenied(t *testing.T) {
	repo := newFakeWorkspaceRepo()
	repo.seed()
	svc := NewWorkspaceServiceServer(repo)

	// Outsider user is not a member of ws-a.
	ctx := ctxAsUser("user", "u-outsider")
	_, err := svc.GetWorkspace(ctx, &agentsv1.GetWorkspaceRequest{Id: "ws-a"})
	twerr, ok := err.(*connect.Error)
	if !ok || twerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound for non-member, got %v", err)
	}
}

func TestWorkspaceService_GetWorkspace_MemberAllowed(t *testing.T) {
	repo := newFakeWorkspaceRepo()
	repo.seed()
	svc := NewWorkspaceServiceServer(repo)

	resp, err := svc.GetWorkspace(ctxAsUser("user", "u-member"), &agentsv1.GetWorkspaceRequest{Id: "ws-a"})
	if err != nil {
		t.Fatalf("member should be allowed: %v", err)
	}
	if resp.GetWorkspace().GetId() != "ws-a" {
		t.Fatalf("expected ws-a, got %s", resp.GetWorkspace().GetId())
	}
}

func TestWorkspaceService_DeleteWorkspace_MemberDenied(t *testing.T) {
	repo := newFakeWorkspaceRepo()
	repo.seed()
	svc := NewWorkspaceServiceServer(repo)

	// "member" role is in the workspace but lacks owner privileges.
	_, err := svc.DeleteWorkspace(ctxAsUser("user", "u-member"), &agentsv1.DeleteWorkspaceRequest{Id: "ws-a"})
	twerr, ok := err.(*connect.Error)
	if !ok || twerr.Code() != connect.CodePermissionDenied {
		t.Fatalf("expected PermissionDenied for non-owner, got %v", err)
	}

	// Outsider — never a member — should not be told the workspace exists.
	_, err = svc.DeleteWorkspace(ctxAsUser("user", "u-outsider"), &agentsv1.DeleteWorkspaceRequest{Id: "ws-a"})
	twerr, ok = err.(*connect.Error)
	if !ok || twerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound for outsider, got %v", err)
	}

	// Verify workspace still exists.
	if _, err := repo.GetWorkspace(context.Background(), "ws-a"); err != nil {
		t.Fatalf("workspace was deleted despite denial: %v", err)
	}
}

func TestWorkspaceService_AddMember_NonOwnerDenied(t *testing.T) {
	repo := newFakeWorkspaceRepo()
	repo.seed()
	svc := NewWorkspaceServiceServer(repo)

	_, err := svc.AddWorkspaceMember(ctxAsUser("user", "u-member"), &agentsv1.AddWorkspaceMemberRequest{
		WorkspaceId: "ws-a",
		UserId:      "u-attacker",
		Role:        "owner",
	})
	twerr, ok := err.(*connect.Error)
	if !ok || twerr.Code() != connect.CodePermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
	if got, _ := repo.GetMember(context.Background(), "ws-a", "u-attacker"); got != nil {
		t.Fatalf("attacker was added despite denial: %+v", got)
	}
}

func TestWorkspaceService_OwnerCanManage(t *testing.T) {
	repo := newFakeWorkspaceRepo()
	repo.seed()
	svc := NewWorkspaceServiceServer(repo)

	_, err := svc.AddWorkspaceMember(ctxAsUser("user", "u-owner"), &agentsv1.AddWorkspaceMemberRequest{
		WorkspaceId: "ws-a",
		UserId:      "u-new",
		Role:        "member",
	})
	if err != nil {
		t.Fatalf("owner add: %v", err)
	}
	if got, err := repo.GetMember(context.Background(), "ws-a", "u-new"); err != nil || got == nil {
		t.Fatalf("owner failed to add member: %v", err)
	}
}

func TestWorkspaceService_AdminBypassesMembership(t *testing.T) {
	repo := newFakeWorkspaceRepo()
	repo.seed()
	svc := NewWorkspaceServiceServer(repo)

	// Admin has no membership in ws-a but should still be allowed.
	_, err := svc.GetWorkspace(ctxAsUser("admin", "u-admin"), &agentsv1.GetWorkspaceRequest{Id: "ws-a"})
	if err != nil {
		t.Fatalf("admin GetWorkspace: %v", err)
	}
	_, err = svc.DeleteWorkspace(ctxAsUser("admin", "u-admin"), &agentsv1.DeleteWorkspaceRequest{Id: "ws-other"})
	if err != nil {
		t.Fatalf("admin DeleteWorkspace: %v", err)
	}
}
