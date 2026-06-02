package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type DashboardServiceConnectAdapter struct {
	agentsv1connect.UnimplementedDashboardServiceHandler
	svc *DashboardServiceServer
}

func NewDashboardServiceConnectAdapter(svc *DashboardServiceServer) *DashboardServiceConnectAdapter {
	return &DashboardServiceConnectAdapter{svc: svc}
}

func (a *DashboardServiceConnectAdapter) GetOverview(ctx context.Context, req *connect.Request[agentsv1.GetOverviewRequest]) (*connect.Response[agentsv1.GetOverviewResponse], error) {
	return connectx.WrapUnary(a.svc.GetOverview)(ctx, req)
}

func (a *DashboardServiceConnectAdapter) GetActivityFeed(ctx context.Context, req *connect.Request[agentsv1.GetActivityFeedRequest]) (*connect.Response[agentsv1.GetActivityFeedResponse], error) {
	return connectx.WrapUnary(a.svc.GetActivityFeed)(ctx, req)
}

func (a *DashboardServiceConnectAdapter) GetCronExecutionTimeseries(ctx context.Context, req *connect.Request[agentsv1.GetCronExecutionTimeseriesRequest]) (*connect.Response[agentsv1.GetCronExecutionTimeseriesResponse], error) {
	return connectx.WrapUnary(a.svc.GetCronExecutionTimeseries)(ctx, req)
}
