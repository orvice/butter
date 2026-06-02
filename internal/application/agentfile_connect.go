package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type AgentFileServiceConnectAdapter struct {
	agentsv1connect.UnimplementedAgentFileServiceHandler
	svc *AgentFileServiceServer
}

func NewAgentFileServiceConnectAdapter(svc *AgentFileServiceServer) *AgentFileServiceConnectAdapter {
	return &AgentFileServiceConnectAdapter{svc: svc}
}

func (a *AgentFileServiceConnectAdapter) ListAgentFileSpaces(ctx context.Context, req *connect.Request[agentsv1.ListAgentFileSpacesRequest]) (*connect.Response[agentsv1.ListAgentFileSpacesResponse], error) {
	return connectx.WrapUnary(a.svc.ListAgentFileSpaces)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) GetAgentFileSpace(ctx context.Context, req *connect.Request[agentsv1.GetAgentFileSpaceRequest]) (*connect.Response[agentsv1.GetAgentFileSpaceResponse], error) {
	return connectx.WrapUnary(a.svc.GetAgentFileSpace)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) CreateAgentFileSpace(ctx context.Context, req *connect.Request[agentsv1.CreateAgentFileSpaceRequest]) (*connect.Response[agentsv1.CreateAgentFileSpaceResponse], error) {
	return connectx.WrapUnary(a.svc.CreateAgentFileSpace)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) UpdateAgentFileSpace(ctx context.Context, req *connect.Request[agentsv1.UpdateAgentFileSpaceRequest]) (*connect.Response[agentsv1.UpdateAgentFileSpaceResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateAgentFileSpace)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) DeleteAgentFileSpace(ctx context.Context, req *connect.Request[agentsv1.DeleteAgentFileSpaceRequest]) (*connect.Response[agentsv1.DeleteAgentFileSpaceResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteAgentFileSpace)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) ListAgentFiles(ctx context.Context, req *connect.Request[agentsv1.ListAgentFilesRequest]) (*connect.Response[agentsv1.ListAgentFilesResponse], error) {
	return connectx.WrapUnary(a.svc.ListAgentFiles)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) GetAgentFile(ctx context.Context, req *connect.Request[agentsv1.GetAgentFileRequest]) (*connect.Response[agentsv1.GetAgentFileResponse], error) {
	return connectx.WrapUnary(a.svc.GetAgentFile)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) WriteAgentFile(ctx context.Context, req *connect.Request[agentsv1.WriteAgentFileRequest]) (*connect.Response[agentsv1.WriteAgentFileResponse], error) {
	return connectx.WrapUnary(a.svc.WriteAgentFile)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) DeleteAgentFile(ctx context.Context, req *connect.Request[agentsv1.DeleteAgentFileRequest]) (*connect.Response[agentsv1.DeleteAgentFileResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteAgentFile)(ctx, req)
}

func (a *AgentFileServiceConnectAdapter) SearchAgentFiles(ctx context.Context, req *connect.Request[agentsv1.SearchAgentFilesRequest]) (*connect.Response[agentsv1.SearchAgentFilesResponse], error) {
	return connectx.WrapUnary(a.svc.SearchAgentFiles)(ctx, req)
}
