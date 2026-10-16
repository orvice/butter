package application

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// GetAgentInvocation returns the authoritative state of one invocation by its
// ID, scoped by workspace and the session access policy.
func (s *AgentServiceServer) GetAgentInvocation(ctx context.Context, req *connect.Request[agentsv1.GetAgentInvocationRequest]) (*connect.Response[agentsv1.GetAgentInvocationResponse], error) {
	if s.invRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("invocation repository not available"))
	}
	if req.Msg.GetInvocationId() == "" {
		return nil, connectx.RequiredArgument("invocation_id")
	}

	wsID, hasWorkspace := workspace.FromContext(ctx)
	if !hasWorkspace && !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace required (set X-Workspace-ID header)"))
	}

	inv, err := getInvocation(ctx, s.invRepo, wsID, req.Msg.GetInvocationId())
	if err != nil {
		if errors.Is(err, invocation.ErrNotFound) {
			return nil, connectx.NotFound("invocation not found")
		}
		return nil, connectx.InternalWith(err)
	}
	if inv == nil {
		return nil, connectx.NotFound("invocation not found")
	}

	if err := s.authorizeInvocationAccess(ctx, wsID, inv); err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.GetAgentInvocationResponse{Invocation: inv}), nil
}

func getInvocation(ctx context.Context, repo invocation.Repository, workspaceID, invocationID string) (*agentsv1.Invocation, error) {
	if workspaceID == "" {
		return repo.GetAcrossWorkspaces(ctx, invocationID)
	}
	return repo.Get(ctx, workspaceID, invocationID)
}

// authorizeInvocationAccess decides whether the caller may reach an
// invocation: one in the request's workspace (when it names one), and then
// exactly when its session is (the session access policy). Unwired, only
// global admins pass. A denied invocation answers NotFound.
func (s *AgentServiceServer) authorizeInvocationAccess(ctx context.Context, workspaceID string, inv *agentsv1.Invocation) error {
	if workspaceID != "" && inv.GetWorkspaceId() != workspaceID {
		return connectx.NotFound("invocation not found")
	}
	if s.sessionAuth == nil {
		if auth.IsAdmin(ctx) {
			return nil
		}
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("session access policy not available"))
	}
	return s.sessionAuth.AuthorizeInvocation(ctx, inv.GetAppName(), inv.GetUserId(), inv.GetWorkspaceId())
}
