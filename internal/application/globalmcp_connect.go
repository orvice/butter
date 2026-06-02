package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type GlobalMCPServerServiceConnectAdapter struct {
	agentsv1connect.UnimplementedGlobalMCPServerServiceHandler
	svc *GlobalMCPServerServiceServer
}

func NewGlobalMCPServerServiceConnectAdapter(svc *GlobalMCPServerServiceServer) *GlobalMCPServerServiceConnectAdapter {
	return &GlobalMCPServerServiceConnectAdapter{svc: svc}
}

func (a *GlobalMCPServerServiceConnectAdapter) ListGlobalMCPServers(ctx context.Context, req *connect.Request[agentsv1.ListGlobalMCPServersRequest]) (*connect.Response[agentsv1.ListGlobalMCPServersResponse], error) {
	return connectx.WrapUnary(a.svc.ListGlobalMCPServers)(ctx, req)
}

func (a *GlobalMCPServerServiceConnectAdapter) CreateGlobalMCPServer(ctx context.Context, req *connect.Request[agentsv1.CreateGlobalMCPServerRequest]) (*connect.Response[agentsv1.CreateGlobalMCPServerResponse], error) {
	return connectx.WrapUnary(a.svc.CreateGlobalMCPServer)(ctx, req)
}

func (a *GlobalMCPServerServiceConnectAdapter) UpdateGlobalMCPServer(ctx context.Context, req *connect.Request[agentsv1.UpdateGlobalMCPServerRequest]) (*connect.Response[agentsv1.UpdateGlobalMCPServerResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateGlobalMCPServer)(ctx, req)
}

func (a *GlobalMCPServerServiceConnectAdapter) DeleteGlobalMCPServer(ctx context.Context, req *connect.Request[agentsv1.DeleteGlobalMCPServerRequest]) (*connect.Response[agentsv1.DeleteGlobalMCPServerResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteGlobalMCPServer)(ctx, req)
}

func (a *GlobalMCPServerServiceConnectAdapter) InstallGlobalMCPServer(ctx context.Context, req *connect.Request[agentsv1.InstallGlobalMCPServerRequest]) (*connect.Response[agentsv1.InstallGlobalMCPServerResponse], error) {
	return connectx.WrapUnary(a.svc.InstallGlobalMCPServer)(ctx, req)
}
