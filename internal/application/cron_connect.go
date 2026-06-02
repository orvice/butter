package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type CronJobServiceConnectAdapter struct {
	agentsv1connect.UnimplementedCronJobServiceHandler
	svc *CronJobServiceServer
}

func NewCronJobServiceConnectAdapter(svc *CronJobServiceServer) *CronJobServiceConnectAdapter {
	return &CronJobServiceConnectAdapter{svc: svc}
}

func (a *CronJobServiceConnectAdapter) ListCronJobs(ctx context.Context, req *connect.Request[agentsv1.ListCronJobsRequest]) (*connect.Response[agentsv1.ListCronJobsResponse], error) {
	return connectx.WrapUnary(a.svc.ListCronJobs)(ctx, req)
}

func (a *CronJobServiceConnectAdapter) GetCronJob(ctx context.Context, req *connect.Request[agentsv1.GetCronJobRequest]) (*connect.Response[agentsv1.GetCronJobResponse], error) {
	return connectx.WrapUnary(a.svc.GetCronJob)(ctx, req)
}

func (a *CronJobServiceConnectAdapter) CreateCronJob(ctx context.Context, req *connect.Request[agentsv1.CreateCronJobRequest]) (*connect.Response[agentsv1.CreateCronJobResponse], error) {
	return connectx.WrapUnary(a.svc.CreateCronJob)(ctx, req)
}

func (a *CronJobServiceConnectAdapter) UpdateCronJob(ctx context.Context, req *connect.Request[agentsv1.UpdateCronJobRequest]) (*connect.Response[agentsv1.UpdateCronJobResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateCronJob)(ctx, req)
}

func (a *CronJobServiceConnectAdapter) DeleteCronJob(ctx context.Context, req *connect.Request[agentsv1.DeleteCronJobRequest]) (*connect.Response[agentsv1.DeleteCronJobResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteCronJob)(ctx, req)
}

func (a *CronJobServiceConnectAdapter) RunCronJobNow(ctx context.Context, req *connect.Request[agentsv1.RunCronJobNowRequest]) (*connect.Response[agentsv1.RunCronJobNowResponse], error) {
	return connectx.WrapUnary(a.svc.RunCronJobNow)(ctx, req)
}

func (a *CronJobServiceConnectAdapter) ListCronExecutions(ctx context.Context, req *connect.Request[agentsv1.ListCronExecutionsRequest]) (*connect.Response[agentsv1.ListCronExecutionsResponse], error) {
	return connectx.WrapUnary(a.svc.ListCronExecutions)(ctx, req)
}
