package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type MCPServerServiceConnectAdapter struct {
	agentsv1connect.UnimplementedMCPServerServiceHandler
	svc *MCPServerServiceServer
}

func NewMCPServerServiceConnectAdapter(svc *MCPServerServiceServer) *MCPServerServiceConnectAdapter {
	return &MCPServerServiceConnectAdapter{svc: svc}
}

func (a *MCPServerServiceConnectAdapter) ListMCPServers(ctx context.Context, req *connect.Request[agentsv1.ListMCPServersRequest]) (*connect.Response[agentsv1.ListMCPServersResponse], error) {
	return connectx.WrapUnary(a.svc.ListMCPServers)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) GetMCPServer(ctx context.Context, req *connect.Request[agentsv1.GetMCPServerRequest]) (*connect.Response[agentsv1.GetMCPServerResponse], error) {
	return connectx.WrapUnary(a.svc.GetMCPServer)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) CreateMCPServer(ctx context.Context, req *connect.Request[agentsv1.CreateMCPServerRequest]) (*connect.Response[agentsv1.CreateMCPServerResponse], error) {
	return connectx.WrapUnary(a.svc.CreateMCPServer)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) UpdateMCPServer(ctx context.Context, req *connect.Request[agentsv1.UpdateMCPServerRequest]) (*connect.Response[agentsv1.UpdateMCPServerResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateMCPServer)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) DeleteMCPServer(ctx context.Context, req *connect.Request[agentsv1.DeleteMCPServerRequest]) (*connect.Response[agentsv1.DeleteMCPServerResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteMCPServer)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) GetMCPServerStatus(ctx context.Context, req *connect.Request[agentsv1.GetMCPServerStatusRequest]) (*connect.Response[agentsv1.GetMCPServerStatusResponse], error) {
	return connectx.WrapUnary(a.svc.GetMCPServerStatus)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) ListMCPTools(ctx context.Context, req *connect.Request[agentsv1.ListMCPToolsRequest]) (*connect.Response[agentsv1.ListMCPToolsResponse], error) {
	return connectx.WrapUnary(a.svc.ListMCPTools)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) StartMCPServerOAuth(ctx context.Context, req *connect.Request[agentsv1.StartMCPServerOAuthRequest]) (*connect.Response[agentsv1.StartMCPServerOAuthResponse], error) {
	return connectx.WrapUnary(a.svc.StartMCPServerOAuth)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) CompleteMCPServerOAuth(ctx context.Context, req *connect.Request[agentsv1.CompleteMCPServerOAuthRequest]) (*connect.Response[agentsv1.CompleteMCPServerOAuthResponse], error) {
	return connectx.WrapUnary(a.svc.CompleteMCPServerOAuth)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) GetMCPServerOAuthStatus(ctx context.Context, req *connect.Request[agentsv1.GetMCPServerOAuthStatusRequest]) (*connect.Response[agentsv1.GetMCPServerOAuthStatusResponse], error) {
	return connectx.WrapUnary(a.svc.GetMCPServerOAuthStatus)(ctx, req)
}

func (a *MCPServerServiceConnectAdapter) DisconnectMCPServerOAuth(ctx context.Context, req *connect.Request[agentsv1.DisconnectMCPServerOAuthRequest]) (*connect.Response[agentsv1.DisconnectMCPServerOAuthResponse], error) {
	return connectx.WrapUnary(a.svc.DisconnectMCPServerOAuth)(ctx, req)
}
