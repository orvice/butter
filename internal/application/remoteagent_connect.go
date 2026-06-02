package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type RemoteAgentServiceConnectAdapter struct {
	agentsv1connect.UnimplementedRemoteAgentServiceHandler
	svc *RemoteAgentServiceServer
}

func NewRemoteAgentServiceConnectAdapter(svc *RemoteAgentServiceServer) *RemoteAgentServiceConnectAdapter {
	return &RemoteAgentServiceConnectAdapter{svc: svc}
}

func (a *RemoteAgentServiceConnectAdapter) ListRemoteAgents(ctx context.Context, req *connect.Request[agentsv1.ListRemoteAgentsRequest]) (*connect.Response[agentsv1.ListRemoteAgentsResponse], error) {
	return connectx.WrapUnary(a.svc.ListRemoteAgents)(ctx, req)
}

func (a *RemoteAgentServiceConnectAdapter) GetRemoteAgent(ctx context.Context, req *connect.Request[agentsv1.GetRemoteAgentRequest]) (*connect.Response[agentsv1.GetRemoteAgentResponse], error) {
	return connectx.WrapUnary(a.svc.GetRemoteAgent)(ctx, req)
}

func (a *RemoteAgentServiceConnectAdapter) CreateRemoteAgent(ctx context.Context, req *connect.Request[agentsv1.CreateRemoteAgentRequest]) (*connect.Response[agentsv1.CreateRemoteAgentResponse], error) {
	return connectx.WrapUnary(a.svc.CreateRemoteAgent)(ctx, req)
}

func (a *RemoteAgentServiceConnectAdapter) UpdateRemoteAgent(ctx context.Context, req *connect.Request[agentsv1.UpdateRemoteAgentRequest]) (*connect.Response[agentsv1.UpdateRemoteAgentResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateRemoteAgent)(ctx, req)
}

func (a *RemoteAgentServiceConnectAdapter) DeleteRemoteAgent(ctx context.Context, req *connect.Request[agentsv1.DeleteRemoteAgentRequest]) (*connect.Response[agentsv1.DeleteRemoteAgentResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteRemoteAgent)(ctx, req)
}

func (a *RemoteAgentServiceConnectAdapter) GetRemoteAgentStatus(ctx context.Context, req *connect.Request[agentsv1.GetRemoteAgentStatusRequest]) (*connect.Response[agentsv1.GetRemoteAgentStatusResponse], error) {
	return connectx.WrapUnary(a.svc.GetRemoteAgentStatus)(ctx, req)
}
