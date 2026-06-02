package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/transport/connectx"
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

func (s *WorkspaceServiceServer) ListWorkspaces(ctx context.Context, _ *connect.Request[agentsv1.ListWorkspacesRequest]) (*connect.Response[agentsv1.ListWorkspacesResponse], error) {
	if s.repo == nil {
		return connect.NewResponse(&agentsv1.ListWorkspacesResponse{}), nil
	}
	if !auth.IsAdmin(ctx) {
		user, hasUser := auth.UserFromContext(ctx)
		if !hasUser {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
		}
		members, err := s.repo.ListMembershipsForUser(ctx, user.GetId())
		if err != nil {
			return nil, connectx.InternalWith(err)
		}
		out := make([]*agentsv1.Workspace, 0, len(members))
		for _, m := range members {
			ws, err := s.repo.GetWorkspace(ctx, m.GetWorkspaceId())
			if err != nil {
				if errors.Is(err, workspacerepo.ErrNotFound) {
					continue
				}
				return nil, connectx.InternalWith(err)
			}
			out = append(out, ws)
		}
		return connect.NewResponse(&agentsv1.ListWorkspacesResponse{Workspaces: out}), nil
	}
	all, err := s.repo.ListWorkspaces(ctx)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.ListWorkspacesResponse{Workspaces: all}), nil
}

func (s *WorkspaceServiceServer) GetWorkspace(ctx context.Context, req *connect.Request[agentsv1.GetWorkspaceRequest]) (*connect.Response[agentsv1.GetWorkspaceResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace store not available"))
	}
	if req.Msg.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	if err := s.requireMembership(ctx, req.Msg.GetId()); err != nil {
		return nil, err
	}
	ws, err := s.repo.GetWorkspace(ctx, req.Msg.GetId())
	if err != nil {
		return nil, mapWorkspaceErr(err)
	}
	return connect.NewResponse(&agentsv1.GetWorkspaceResponse{Workspace: ws}), nil
}

func (s *WorkspaceServiceServer) CreateWorkspace(ctx context.Context, req *connect.Request[agentsv1.CreateWorkspaceRequest]) (*connect.Response[agentsv1.CreateWorkspaceResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace store not available"))
	}
	in := req.Msg.GetWorkspace()
	if in == nil {
		return nil, connectx.RequiredArgument("workspace")
	}
	name := strings.TrimSpace(in.GetName())
	slug := strings.TrimSpace(in.GetSlug())
	if name == "" {
		return nil, connectx.RequiredArgument("name")
	}
	if slug == "" {
		return nil, connectx.RequiredArgument("slug")
	}

	logger := log.FromContext(ctx)
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
		logger.Error("create workspace failed", "name", name, "slug", slug, "err", err)
		return nil, mapWorkspaceErr(err)
	}

	// Add the caller as the initial owner.
	if user, ok := auth.UserFromContext(ctx); ok {
		if _, err := s.repo.AddMember(ctx, &agentsv1.WorkspaceMember{
			WorkspaceId: created.GetId(),
			UserId:      user.GetId(),
			Role:        "owner",
			CreatedAt:   timestamppb.New(now),
		}); err != nil {
			logger.Warn("create workspace: failed to add caller as owner",
				"workspace_id", created.GetId(), "user_id", user.GetId(), "err", err)
		}
	}

	logger.Info("workspace created", "workspace_id", created.GetId(), "name", created.GetName(), "slug", created.GetSlug())
	return connect.NewResponse(&agentsv1.CreateWorkspaceResponse{Workspace: created}), nil
}

func (s *WorkspaceServiceServer) UpdateWorkspace(ctx context.Context, req *connect.Request[agentsv1.UpdateWorkspaceRequest]) (*connect.Response[agentsv1.UpdateWorkspaceResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace store not available"))
	}
	in := req.Msg.GetWorkspace()
	if in == nil || in.GetId() == "" {
		return nil, connectx.RequiredArgument("workspace.id")
	}
	if err := s.requireRole(ctx, in.GetId(), "owner"); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	in.UpdatedAt = timestamppb.New(time.Now().UTC())
	updated, err := s.repo.UpdateWorkspace(ctx, in)
	if err != nil {
		logger.Error("update workspace failed", "workspace_id", in.GetId(), "err", err)
		return nil, mapWorkspaceErr(err)
	}
	logger.Info("workspace updated", "workspace_id", updated.GetId(), "name", updated.GetName())
	return connect.NewResponse(&agentsv1.UpdateWorkspaceResponse{Workspace: updated}), nil
}

func (s *WorkspaceServiceServer) DeleteWorkspace(ctx context.Context, req *connect.Request[agentsv1.DeleteWorkspaceRequest]) (*connect.Response[agentsv1.DeleteWorkspaceResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace store not available"))
	}
	if req.Msg.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	if err := s.requireRole(ctx, req.Msg.GetId(), "owner"); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	if err := s.repo.DeleteWorkspace(ctx, req.Msg.GetId()); err != nil {
		logger.Error("delete workspace failed", "workspace_id", req.Msg.GetId(), "err", err)
		return nil, mapWorkspaceErr(err)
	}
	logger.Info("workspace deleted", "workspace_id", req.Msg.GetId())
	return connect.NewResponse(&agentsv1.DeleteWorkspaceResponse{}), nil
}

func (s *WorkspaceServiceServer) ListWorkspaceMembers(ctx context.Context, req *connect.Request[agentsv1.ListWorkspaceMembersRequest]) (*connect.Response[agentsv1.ListWorkspaceMembersResponse], error) {
	if s.repo == nil {
		return connect.NewResponse(&agentsv1.ListWorkspaceMembersResponse{}), nil
	}
	if req.Msg.GetWorkspaceId() == "" {
		return nil, connectx.RequiredArgument("workspace_id")
	}
	if err := s.requireMembership(ctx, req.Msg.GetWorkspaceId()); err != nil {
		return nil, err
	}
	members, err := s.repo.ListMembers(ctx, req.Msg.GetWorkspaceId())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.ListWorkspaceMembersResponse{Members: members}), nil
}

func (s *WorkspaceServiceServer) AddWorkspaceMember(ctx context.Context, req *connect.Request[agentsv1.AddWorkspaceMemberRequest]) (*connect.Response[agentsv1.AddWorkspaceMemberResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace store not available"))
	}
	if req.Msg.GetWorkspaceId() == "" {
		return nil, connectx.RequiredArgument("workspace_id")
	}
	if req.Msg.GetUserId() == "" {
		return nil, connectx.RequiredArgument("user_id")
	}
	if err := s.requireRole(ctx, req.Msg.GetWorkspaceId(), "owner"); err != nil {
		return nil, err
	}
	role := strings.TrimSpace(req.Msg.GetRole())
	if role == "" {
		role = "member"
	}
	m := &agentsv1.WorkspaceMember{
		WorkspaceId: req.Msg.GetWorkspaceId(),
		UserId:      req.Msg.GetUserId(),
		Role:        role,
		CreatedAt:   timestamppb.New(time.Now().UTC()),
	}
	logger := log.FromContext(ctx)
	created, err := s.repo.AddMember(ctx, m)
	if err != nil {
		logger.Error("add workspace member failed", "workspace_id", req.Msg.GetWorkspaceId(), "user_id", req.Msg.GetUserId(), "role", role, "err", err)
		return nil, mapWorkspaceErr(err)
	}
	logger.Info("workspace member added", "workspace_id", req.Msg.GetWorkspaceId(), "user_id", req.Msg.GetUserId(), "role", role)
	return connect.NewResponse(&agentsv1.AddWorkspaceMemberResponse{Member: created}), nil
}

func (s *WorkspaceServiceServer) UpdateWorkspaceMember(ctx context.Context, req *connect.Request[agentsv1.UpdateWorkspaceMemberRequest]) (*connect.Response[agentsv1.UpdateWorkspaceMemberResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace store not available"))
	}
	if req.Msg.GetWorkspaceId() == "" {
		return nil, connectx.RequiredArgument("workspace_id")
	}
	if req.Msg.GetUserId() == "" {
		return nil, connectx.RequiredArgument("user_id")
	}
	if err := s.requireRole(ctx, req.Msg.GetWorkspaceId(), "owner"); err != nil {
		return nil, err
	}
	m := &agentsv1.WorkspaceMember{
		WorkspaceId: req.Msg.GetWorkspaceId(),
		UserId:      req.Msg.GetUserId(),
		Role:        req.Msg.GetRole(),
	}
	logger := log.FromContext(ctx)
	updated, err := s.repo.UpdateMember(ctx, m)
	if err != nil {
		logger.Error("update workspace member failed", "workspace_id", req.Msg.GetWorkspaceId(), "user_id", req.Msg.GetUserId(), "role", req.Msg.GetRole(), "err", err)
		return nil, mapWorkspaceErr(err)
	}
	logger.Info("workspace member updated", "workspace_id", req.Msg.GetWorkspaceId(), "user_id", req.Msg.GetUserId(), "role", updated.GetRole())
	return connect.NewResponse(&agentsv1.UpdateWorkspaceMemberResponse{Member: updated}), nil
}

func (s *WorkspaceServiceServer) RemoveWorkspaceMember(ctx context.Context, req *connect.Request[agentsv1.RemoveWorkspaceMemberRequest]) (*connect.Response[agentsv1.RemoveWorkspaceMemberResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace store not available"))
	}
	if req.Msg.GetWorkspaceId() == "" {
		return nil, connectx.RequiredArgument("workspace_id")
	}
	if req.Msg.GetUserId() == "" {
		return nil, connectx.RequiredArgument("user_id")
	}
	if err := s.requireRole(ctx, req.Msg.GetWorkspaceId(), "owner"); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	if err := s.repo.RemoveMember(ctx, req.Msg.GetWorkspaceId(), req.Msg.GetUserId()); err != nil {
		logger.Error("remove workspace member failed", "workspace_id", req.Msg.GetWorkspaceId(), "user_id", req.Msg.GetUserId(), "err", err)
		return nil, mapWorkspaceErr(err)
	}
	logger.Info("workspace member removed", "workspace_id", req.Msg.GetWorkspaceId(), "user_id", req.Msg.GetUserId())
	return connect.NewResponse(&agentsv1.RemoveWorkspaceMemberResponse{}), nil
}

// requireMembership returns nil if the caller is an admin or a member of the
// workspace; otherwise NotFound to avoid leaking the existence of workspaces.
func (s *WorkspaceServiceServer) requireMembership(ctx context.Context, workspaceID string) error {
	if auth.IsAdmin(ctx) {
		return nil
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	member, err := s.repo.IsMember(ctx, workspaceID, user.GetId())
	if err != nil {
		return connectx.InternalWith(err)
	}
	if !member {
		return connectx.NotFound("workspace not found")
	}
	return nil
}

// requireRole returns nil if the caller is an admin or a member of the
// workspace with one of the given roles. Members lacking the required role
// receive PermissionDenied; non-members receive NotFound.
func (s *WorkspaceServiceServer) requireRole(ctx context.Context, workspaceID string, roles ...string) error {
	if auth.IsAdmin(ctx) {
		return nil
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	member, err := s.repo.GetMember(ctx, workspaceID, user.GetId())
	if err != nil {
		if errors.Is(err, workspacerepo.ErrNotFound) {
			return connectx.NotFound("workspace not found")
		}
		return connectx.InternalWith(err)
	}
	if slices.Contains(roles, member.GetRole()) {
		return nil
	}
	return connect.NewError(connect.CodePermissionDenied, errors.New("insufficient workspace role"))
}

func mapWorkspaceErr(err error) *connect.Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, workspacerepo.ErrNotFound) {
		return connectx.NotFound(err.Error())
	}
	if errors.Is(err, workspacerepo.ErrAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, errors.New(err.Error()))
	}
	return connectx.InternalWith(err)
}
