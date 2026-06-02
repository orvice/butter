package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type SessionServiceConnectAdapter struct {
	agentsv1connect.UnimplementedSessionServiceHandler
	svc *SessionServiceServer
}

func NewSessionServiceConnectAdapter(svc *SessionServiceServer) *SessionServiceConnectAdapter {
	return &SessionServiceConnectAdapter{svc: svc}
}

func (a *SessionServiceConnectAdapter) CreateSession(ctx context.Context, req *connect.Request[agentsv1.CreateSessionRequest]) (*connect.Response[agentsv1.CreateSessionResponse], error) {
	return connectx.WrapUnary(a.svc.CreateSession)(ctx, req)
}

func (a *SessionServiceConnectAdapter) GetSession(ctx context.Context, req *connect.Request[agentsv1.GetSessionRequest]) (*connect.Response[agentsv1.GetSessionResponse], error) {
	return connectx.WrapUnary(a.svc.GetSession)(ctx, req)
}

func (a *SessionServiceConnectAdapter) ListSessions(ctx context.Context, req *connect.Request[agentsv1.ListSessionsRequest]) (*connect.Response[agentsv1.ListSessionsResponse], error) {
	return connectx.WrapUnary(a.svc.ListSessions)(ctx, req)
}

func (a *SessionServiceConnectAdapter) DeleteSession(ctx context.Context, req *connect.Request[agentsv1.DeleteSessionRequest]) (*connect.Response[agentsv1.DeleteSessionResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteSession)(ctx, req)
}

func (a *SessionServiceConnectAdapter) ReplySession(ctx context.Context, req *connect.Request[agentsv1.ReplySessionRequest]) (*connect.Response[agentsv1.ReplySessionResponse], error) {
	return connectx.WrapUnary(a.svc.ReplySession)(ctx, req)
}
