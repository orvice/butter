package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type NotifyGroupServiceConnectAdapter struct {
	agentsv1connect.UnimplementedNotifyGroupServiceHandler
	svc *NotifyGroupServiceServer
}

func NewNotifyGroupServiceConnectAdapter(svc *NotifyGroupServiceServer) *NotifyGroupServiceConnectAdapter {
	return &NotifyGroupServiceConnectAdapter{svc: svc}
}

func (a *NotifyGroupServiceConnectAdapter) ListNotifyGroups(ctx context.Context, req *connect.Request[agentsv1.ListNotifyGroupsRequest]) (*connect.Response[agentsv1.ListNotifyGroupsResponse], error) {
	return connectx.WrapUnary(a.svc.ListNotifyGroups)(ctx, req)
}

func (a *NotifyGroupServiceConnectAdapter) GetNotifyGroup(ctx context.Context, req *connect.Request[agentsv1.GetNotifyGroupRequest]) (*connect.Response[agentsv1.GetNotifyGroupResponse], error) {
	return connectx.WrapUnary(a.svc.GetNotifyGroup)(ctx, req)
}

func (a *NotifyGroupServiceConnectAdapter) CreateNotifyGroup(ctx context.Context, req *connect.Request[agentsv1.CreateNotifyGroupRequest]) (*connect.Response[agentsv1.CreateNotifyGroupResponse], error) {
	return connectx.WrapUnary(a.svc.CreateNotifyGroup)(ctx, req)
}

func (a *NotifyGroupServiceConnectAdapter) UpdateNotifyGroup(ctx context.Context, req *connect.Request[agentsv1.UpdateNotifyGroupRequest]) (*connect.Response[agentsv1.UpdateNotifyGroupResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateNotifyGroup)(ctx, req)
}

func (a *NotifyGroupServiceConnectAdapter) DeleteNotifyGroup(ctx context.Context, req *connect.Request[agentsv1.DeleteNotifyGroupRequest]) (*connect.Response[agentsv1.DeleteNotifyGroupResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteNotifyGroup)(ctx, req)
}
