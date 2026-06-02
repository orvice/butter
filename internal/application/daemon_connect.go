package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type DaemonServiceConnectAdapter struct {
	agentsv1connect.UnimplementedDaemonServiceHandler
	svc *DaemonServiceServer
}

func NewDaemonServiceConnectAdapter(svc *DaemonServiceServer) *DaemonServiceConnectAdapter {
	return &DaemonServiceConnectAdapter{svc: svc}
}

func (a *DaemonServiceConnectAdapter) ListDaemons(ctx context.Context, req *connect.Request[agentsv1.ListDaemonsRequest]) (*connect.Response[agentsv1.ListDaemonsResponse], error) {
	return connectx.WrapUnary(a.svc.ListDaemons)(ctx, req)
}

func (a *DaemonServiceConnectAdapter) GetDaemon(ctx context.Context, req *connect.Request[agentsv1.GetDaemonRequest]) (*connect.Response[agentsv1.GetDaemonResponse], error) {
	return connectx.WrapUnary(a.svc.GetDaemon)(ctx, req)
}

func (a *DaemonServiceConnectAdapter) CancelDaemonTask(ctx context.Context, req *connect.Request[agentsv1.CancelDaemonTaskRequest]) (*connect.Response[agentsv1.CancelDaemonTaskResponse], error) {
	return connectx.WrapUnary(a.svc.CancelDaemonTask)(ctx, req)
}

func (a *DaemonServiceConnectAdapter) ListDaemonTasks(ctx context.Context, req *connect.Request[agentsv1.ListDaemonTasksRequest]) (*connect.Response[agentsv1.ListDaemonTasksResponse], error) {
	return connectx.WrapUnary(a.svc.ListDaemonTasks)(ctx, req)
}

func (a *DaemonServiceConnectAdapter) GetBridgeDiagnostics(ctx context.Context, req *connect.Request[agentsv1.GetBridgeDiagnosticsRequest]) (*connect.Response[agentsv1.GetBridgeDiagnosticsResponse], error) {
	return connectx.WrapUnary(a.svc.GetBridgeDiagnostics)(ctx, req)
}
