// Package workspace contains the workspace and workspace-membership
// repositories shared by the application and auth layers.
package workspace

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrNotFound      = errors.New("workspace not found")
	ErrAlreadyExists = errors.New("workspace already exists")
)

// Repository persists workspaces and workspace memberships.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	ListWorkspaces(ctx context.Context) ([]*agentsv1.Workspace, error)
	GetWorkspace(ctx context.Context, id string) (*agentsv1.Workspace, error)
	GetWorkspaceBySlug(ctx context.Context, slug string) (*agentsv1.Workspace, error)
	CreateWorkspace(ctx context.Context, ws *agentsv1.Workspace) (*agentsv1.Workspace, error)
	UpdateWorkspace(ctx context.Context, ws *agentsv1.Workspace) (*agentsv1.Workspace, error)
	DeleteWorkspace(ctx context.Context, id string) error
	CountWorkspaces(ctx context.Context) (int64, error)

	ListMembers(ctx context.Context, workspaceID string) ([]*agentsv1.WorkspaceMember, error)
	ListMembershipsForUser(ctx context.Context, userID string) ([]*agentsv1.WorkspaceMember, error)
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
	GetMember(ctx context.Context, workspaceID, userID string) (*agentsv1.WorkspaceMember, error)
	AddMember(ctx context.Context, member *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error)
	UpdateMember(ctx context.Context, member *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error)
	RemoveMember(ctx context.Context, workspaceID, userID string) error
}
