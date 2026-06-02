package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type ModelProviderServiceConnectAdapter struct {
	agentsv1connect.UnimplementedModelProviderServiceHandler
	svc *ModelProviderServiceServer
}

func NewModelProviderServiceConnectAdapter(svc *ModelProviderServiceServer) *ModelProviderServiceConnectAdapter {
	return &ModelProviderServiceConnectAdapter{svc: svc}
}

func (a *ModelProviderServiceConnectAdapter) ListModelProviders(ctx context.Context, req *connect.Request[agentsv1.ListModelProvidersRequest]) (*connect.Response[agentsv1.ListModelProvidersResponse], error) {
	return connectx.WrapUnary(a.svc.ListModelProviders)(ctx, req)
}

func (a *ModelProviderServiceConnectAdapter) GetModelProvider(ctx context.Context, req *connect.Request[agentsv1.GetModelProviderRequest]) (*connect.Response[agentsv1.GetModelProviderResponse], error) {
	return connectx.WrapUnary(a.svc.GetModelProvider)(ctx, req)
}

func (a *ModelProviderServiceConnectAdapter) CreateModelProvider(ctx context.Context, req *connect.Request[agentsv1.CreateModelProviderRequest]) (*connect.Response[agentsv1.CreateModelProviderResponse], error) {
	return connectx.WrapUnary(a.svc.CreateModelProvider)(ctx, req)
}

func (a *ModelProviderServiceConnectAdapter) UpdateModelProvider(ctx context.Context, req *connect.Request[agentsv1.UpdateModelProviderRequest]) (*connect.Response[agentsv1.UpdateModelProviderResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateModelProvider)(ctx, req)
}

func (a *ModelProviderServiceConnectAdapter) DeleteModelProvider(ctx context.Context, req *connect.Request[agentsv1.DeleteModelProviderRequest]) (*connect.Response[agentsv1.DeleteModelProviderResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteModelProvider)(ctx, req)
}
