package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type WorkspaceServiceConnectAdapter struct {
	agentsv1connect.UnimplementedWorkspaceServiceHandler
	svc *WorkspaceServiceServer
}

func NewWorkspaceServiceConnectAdapter(svc *WorkspaceServiceServer) *WorkspaceServiceConnectAdapter {
	return &WorkspaceServiceConnectAdapter{svc: svc}
}

func (a *WorkspaceServiceConnectAdapter) ListWorkspaces(ctx context.Context, req *connect.Request[agentsv1.ListWorkspacesRequest]) (*connect.Response[agentsv1.ListWorkspacesResponse], error) {
	return connectx.WrapUnary(a.svc.ListWorkspaces)(ctx, req)
}

func (a *WorkspaceServiceConnectAdapter) GetWorkspace(ctx context.Context, req *connect.Request[agentsv1.GetWorkspaceRequest]) (*connect.Response[agentsv1.GetWorkspaceResponse], error) {
	return connectx.WrapUnary(a.svc.GetWorkspace)(ctx, req)
}

func (a *WorkspaceServiceConnectAdapter) CreateWorkspace(ctx context.Context, req *connect.Request[agentsv1.CreateWorkspaceRequest]) (*connect.Response[agentsv1.CreateWorkspaceResponse], error) {
	return connectx.WrapUnary(a.svc.CreateWorkspace)(ctx, req)
}

func (a *WorkspaceServiceConnectAdapter) UpdateWorkspace(ctx context.Context, req *connect.Request[agentsv1.UpdateWorkspaceRequest]) (*connect.Response[agentsv1.UpdateWorkspaceResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateWorkspace)(ctx, req)
}

func (a *WorkspaceServiceConnectAdapter) DeleteWorkspace(ctx context.Context, req *connect.Request[agentsv1.DeleteWorkspaceRequest]) (*connect.Response[agentsv1.DeleteWorkspaceResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteWorkspace)(ctx, req)
}

func (a *WorkspaceServiceConnectAdapter) ListWorkspaceMembers(ctx context.Context, req *connect.Request[agentsv1.ListWorkspaceMembersRequest]) (*connect.Response[agentsv1.ListWorkspaceMembersResponse], error) {
	return connectx.WrapUnary(a.svc.ListWorkspaceMembers)(ctx, req)
}

func (a *WorkspaceServiceConnectAdapter) AddWorkspaceMember(ctx context.Context, req *connect.Request[agentsv1.AddWorkspaceMemberRequest]) (*connect.Response[agentsv1.AddWorkspaceMemberResponse], error) {
	return connectx.WrapUnary(a.svc.AddWorkspaceMember)(ctx, req)
}

func (a *WorkspaceServiceConnectAdapter) UpdateWorkspaceMember(ctx context.Context, req *connect.Request[agentsv1.UpdateWorkspaceMemberRequest]) (*connect.Response[agentsv1.UpdateWorkspaceMemberResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateWorkspaceMember)(ctx, req)
}

func (a *WorkspaceServiceConnectAdapter) RemoveWorkspaceMember(ctx context.Context, req *connect.Request[agentsv1.RemoveWorkspaceMemberRequest]) (*connect.Response[agentsv1.RemoveWorkspaceMemberResponse], error) {
	return connectx.WrapUnary(a.svc.RemoveWorkspaceMember)(ctx, req)
}
