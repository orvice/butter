package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type APITokenServiceConnectAdapter struct {
	agentsv1connect.UnimplementedAPITokenServiceHandler
	svc *APITokenServiceServer
}

func NewAPITokenServiceConnectAdapter(svc *APITokenServiceServer) *APITokenServiceConnectAdapter {
	return &APITokenServiceConnectAdapter{svc: svc}
}

func (a *APITokenServiceConnectAdapter) ListAPITokens(ctx context.Context, req *connect.Request[agentsv1.ListAPITokensRequest]) (*connect.Response[agentsv1.ListAPITokensResponse], error) {
	return connectx.WrapUnary(a.svc.ListAPITokens)(ctx, req)
}

func (a *APITokenServiceConnectAdapter) CreateAPIToken(ctx context.Context, req *connect.Request[agentsv1.CreateAPITokenRequest]) (*connect.Response[agentsv1.CreateAPITokenResponse], error) {
	return connectx.WrapUnary(a.svc.CreateAPIToken)(ctx, req)
}

func (a *APITokenServiceConnectAdapter) RevokeAPIToken(ctx context.Context, req *connect.Request[agentsv1.RevokeAPITokenRequest]) (*connect.Response[agentsv1.RevokeAPITokenResponse], error) {
	return connectx.WrapUnary(a.svc.RevokeAPIToken)(ctx, req)
}
