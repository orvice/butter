package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// WorkspaceServiceServer implements the WorkspaceService Twirp interface.
type WorkspaceServiceServer struct {
	repo workspacerepo.Repository
}

func NewWorkspaceServiceServer(repo workspacerepo.Repository) *WorkspaceServiceServer {
	return &WorkspaceServiceServer{repo: repo}
}

func (s *WorkspaceServiceServer) SetRepo(repo workspacerepo.Repository) { s.repo = repo }

func (s *WorkspaceServiceServer) ListWorkspaces(ctx context.Context, _ *agentsv1.ListWorkspacesRequest) (*agentsv1.ListWorkspacesResponse, error) {
	if s.repo == nil {
		return &agentsv1.ListWorkspacesResponse{}, nil
	}
	user, hasUser := auth.UserFromContext(ctx)
	if hasUser && user.GetRole() != "admin" {
		members, err := s.repo.ListMembershipsForUser(ctx, user.GetId())
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
		}
		out := make([]*agentsv1.Workspace, 0, len(members))
		for _, m := range members {
			ws, err := s.repo.GetWorkspace(ctx, m.GetWorkspaceId())
			if err != nil {
				if errors.Is(err, workspacerepo.ErrNotFound) {
					continue
				}
				return nil, twirp.InternalErrorWith(err)
			}
			out = append(out, ws)
		}
		return &agentsv1.ListWorkspacesResponse{Workspaces: out}, nil
	}
	all, err := s.repo.ListWorkspaces(ctx)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ListWorkspacesResponse{Workspaces: all}, nil
}

func (s *WorkspaceServiceServer) GetWorkspace(ctx context.Context, req *agentsv1.GetWorkspaceRequest) (*agentsv1.GetWorkspaceResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "workspace store not available")
	}
	ws, err := s.repo.GetWorkspace(ctx, req.GetId())
	if err != nil {
		return nil, mapWorkspaceErr(err)
	}
	return &agentsv1.GetWorkspaceResponse{Workspace: ws}, nil
}

func (s *WorkspaceServiceServer) CreateWorkspace(ctx context.Context, req *agentsv1.CreateWorkspaceRequest) (*agentsv1.CreateWorkspaceResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "workspace store not available")
	}
	in := req.GetWorkspace()
	if in == nil {
		return nil, twirp.RequiredArgumentError("workspace")
	}
	name := strings.TrimSpace(in.GetName())
	slug := strings.TrimSpace(in.GetSlug())
	if name == "" {
		return nil, twirp.RequiredArgumentError("name")
	}
	if slug == "" {
		return nil, twirp.RequiredArgumentError("slug")
	}

	now := time.Now().UTC()
	ws := &agentsv1.Workspace{
		Id:          uuid.NewString(),
		Name:        name,
		Slug:        slug,
		Description: in.GetDescription(),
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}
	created, err := s.repo.CreateWorkspace(ctx, ws)
	if err != nil {
		return nil, mapWorkspaceErr(err)
	}

	// Add the caller as the initial owner.
	if user, ok := auth.UserFromContext(ctx); ok {
		_, _ = s.repo.AddMember(ctx, &agentsv1.WorkspaceMember{
			WorkspaceId: created.GetId(),
			UserId:      user.GetId(),
			Role:        "owner",
			CreatedAt:   timestamppb.New(now),
		})
	}

	return &agentsv1.CreateWorkspaceResponse{Workspace: created}, nil
}

func (s *WorkspaceServiceServer) UpdateWorkspace(ctx context.Context, req *agentsv1.UpdateWorkspaceRequest) (*agentsv1.UpdateWorkspaceResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "workspace store not available")
	}
	in := req.GetWorkspace()
	if in == nil || in.GetId() == "" {
		return nil, twirp.RequiredArgumentError("workspace.id")
	}
	in.UpdatedAt = timestamppb.New(time.Now().UTC())
	updated, err := s.repo.UpdateWorkspace(ctx, in)
	if err != nil {
		return nil, mapWorkspaceErr(err)
	}
	return &agentsv1.UpdateWorkspaceResponse{Workspace: updated}, nil
}

func (s *WorkspaceServiceServer) DeleteWorkspace(ctx context.Context, req *agentsv1.DeleteWorkspaceRequest) (*agentsv1.DeleteWorkspaceResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "workspace store not available")
	}
	if err := s.repo.DeleteWorkspace(ctx, req.GetId()); err != nil {
		return nil, mapWorkspaceErr(err)
	}
	return &agentsv1.DeleteWorkspaceResponse{}, nil
}

func (s *WorkspaceServiceServer) ListWorkspaceMembers(ctx context.Context, req *agentsv1.ListWorkspaceMembersRequest) (*agentsv1.ListWorkspaceMembersResponse, error) {
	if s.repo == nil {
		return &agentsv1.ListWorkspaceMembersResponse{}, nil
	}
	members, err := s.repo.ListMembers(ctx, req.GetWorkspaceId())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ListWorkspaceMembersResponse{Members: members}, nil
}

func (s *WorkspaceServiceServer) AddWorkspaceMember(ctx context.Context, req *agentsv1.AddWorkspaceMemberRequest) (*agentsv1.AddWorkspaceMemberResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "workspace store not available")
	}
	role := strings.TrimSpace(req.GetRole())
	if role == "" {
		role = "member"
	}
	m := &agentsv1.WorkspaceMember{
		WorkspaceId: req.GetWorkspaceId(),
		UserId:      req.GetUserId(),
		Role:        role,
		CreatedAt:   timestamppb.New(time.Now().UTC()),
	}
	created, err := s.repo.AddMember(ctx, m)
	if err != nil {
		return nil, mapWorkspaceErr(err)
	}
	return &agentsv1.AddWorkspaceMemberResponse{Member: created}, nil
}

func (s *WorkspaceServiceServer) UpdateWorkspaceMember(ctx context.Context, req *agentsv1.UpdateWorkspaceMemberRequest) (*agentsv1.UpdateWorkspaceMemberResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "workspace store not available")
	}
	m := &agentsv1.WorkspaceMember{
		WorkspaceId: req.GetWorkspaceId(),
		UserId:      req.GetUserId(),
		Role:        req.GetRole(),
	}
	updated, err := s.repo.UpdateMember(ctx, m)
	if err != nil {
		return nil, mapWorkspaceErr(err)
	}
	return &agentsv1.UpdateWorkspaceMemberResponse{Member: updated}, nil
}

func (s *WorkspaceServiceServer) RemoveWorkspaceMember(ctx context.Context, req *agentsv1.RemoveWorkspaceMemberRequest) (*agentsv1.RemoveWorkspaceMemberResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "workspace store not available")
	}
	if err := s.repo.RemoveMember(ctx, req.GetWorkspaceId(), req.GetUserId()); err != nil {
		return nil, mapWorkspaceErr(err)
	}
	return &agentsv1.RemoveWorkspaceMemberResponse{}, nil
}

func mapWorkspaceErr(err error) twirp.Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, workspacerepo.ErrNotFound) {
		return twirp.NotFoundError(err.Error())
	}
	if errors.Is(err, workspacerepo.ErrAlreadyExists) {
		return twirp.NewError(twirp.AlreadyExists, err.Error())
	}
	return twirp.InternalErrorWith(err)
}
