package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type AgentServiceConnectAdapter struct {
	agentsv1connect.UnimplementedAgentServiceHandler
	svc *AgentServiceServer
}

func NewAgentServiceConnectAdapter(svc *AgentServiceServer) *AgentServiceConnectAdapter {
	return &AgentServiceConnectAdapter{svc: svc}
}

func (a *AgentServiceConnectAdapter) ListAgents(ctx context.Context, req *connect.Request[agentsv1.ListAgentsRequest]) (*connect.Response[agentsv1.ListAgentsResponse], error) {
	return connectx.WrapUnary(a.svc.ListAgents)(ctx, req)
}

func (a *AgentServiceConnectAdapter) ReloadAgents(ctx context.Context, req *connect.Request[agentsv1.ReloadAgentsRequest]) (*connect.Response[agentsv1.ReloadAgentsResponse], error) {
	return connectx.WrapUnary(a.svc.ReloadAgents)(ctx, req)
}

func (a *AgentServiceConnectAdapter) GetAgent(ctx context.Context, req *connect.Request[agentsv1.GetAgentRequest]) (*connect.Response[agentsv1.GetAgentResponse], error) {
	return connectx.WrapUnary(a.svc.GetAgent)(ctx, req)
}

func (a *AgentServiceConnectAdapter) CreateAgent(ctx context.Context, req *connect.Request[agentsv1.CreateAgentRequest]) (*connect.Response[agentsv1.CreateAgentResponse], error) {
	return connectx.WrapUnary(a.svc.CreateAgent)(ctx, req)
}

func (a *AgentServiceConnectAdapter) UpdateAgent(ctx context.Context, req *connect.Request[agentsv1.UpdateAgentRequest]) (*connect.Response[agentsv1.UpdateAgentResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateAgent)(ctx, req)
}

func (a *AgentServiceConnectAdapter) DeleteAgent(ctx context.Context, req *connect.Request[agentsv1.DeleteAgentRequest]) (*connect.Response[agentsv1.DeleteAgentResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteAgent)(ctx, req)
}

func (a *AgentServiceConnectAdapter) InvokeAgent(ctx context.Context, req *connect.Request[agentsv1.InvokeAgentRequest]) (*connect.Response[agentsv1.InvokeAgentResponse], error) {
	return connectx.WrapUnary(a.svc.InvokeAgent)(ctx, req)
}

func (a *AgentServiceConnectAdapter) ListAgentInvocations(ctx context.Context, req *connect.Request[agentsv1.ListAgentInvocationsRequest]) (*connect.Response[agentsv1.ListAgentInvocationsResponse], error) {
	return connectx.WrapUnary(a.svc.ListAgentInvocations)(ctx, req)
}

func (a *AgentServiceConnectAdapter) CancelAgentInvocation(ctx context.Context, req *connect.Request[agentsv1.CancelAgentInvocationRequest]) (*connect.Response[agentsv1.CancelAgentInvocationResponse], error) {
	return connectx.WrapUnary(a.svc.CancelAgentInvocation)(ctx, req)
}

func (a *AgentServiceConnectAdapter) GetAgentRuntimeStatus(ctx context.Context, req *connect.Request[agentsv1.GetAgentRuntimeStatusRequest]) (*connect.Response[agentsv1.GetAgentRuntimeStatusResponse], error) {
	return connectx.WrapUnary(a.svc.GetAgentRuntimeStatus)(ctx, req)
}

func (a *AgentServiceConnectAdapter) ListAgentRuntimeStatuses(ctx context.Context, req *connect.Request[agentsv1.ListAgentRuntimeStatusesRequest]) (*connect.Response[agentsv1.ListAgentRuntimeStatusesResponse], error) {
	return connectx.WrapUnary(a.svc.ListAgentRuntimeStatuses)(ctx, req)
}
