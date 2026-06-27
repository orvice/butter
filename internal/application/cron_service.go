package application

import (
	"context"
	"errors"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// cronJobMutationError maps scheduler errors to typed connect errors so
// that validation failures (e.g. agent does not exist in workspace) surface
// as InvalidArgument instead of being swallowed by toConnectError's
// catch-all InternalWith.
func cronJobMutationError(err error) error {
	if errors.Is(err, cron.ErrAgentNotInWorkspace) {
		return connectx.InvalidArgument("agent_name", err.Error())
	}
	return toConnectError(err)
}

// CronJobServiceServer implements the CronJobService ConnectRPC handler.
type CronJobServiceServer struct {
	scheduler *cron.Scheduler
	execRepo  cron.ExecutionRepo
}

// NewCronJobServiceServer creates a new CronJobServiceServer.
func NewCronJobServiceServer() *CronJobServiceServer {
	return &CronJobServiceServer{}
}

// SetScheduler sets the cron scheduler (called after bootstrap).
func (s *CronJobServiceServer) SetScheduler(scheduler *cron.Scheduler) {
	s.scheduler = scheduler
}

// SetExecutionRepo sets the execution repo (called after bootstrap).
func (s *CronJobServiceServer) SetExecutionRepo(repo cron.ExecutionRepo) {
	s.execRepo = repo
}

func (s *CronJobServiceServer) ListCronJobs(ctx context.Context, _ *connect.Request[agentsv1.ListCronJobsRequest]) (*connect.Response[agentsv1.ListCronJobsResponse], error) {
	if s.scheduler == nil {
		return connect.NewResponse(&agentsv1.ListCronJobsResponse{}), nil
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	jobs, err := s.scheduler.ListJobs(ctx, wsID)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.ListCronJobsResponse{CronJobs: jobs}), nil
}

func (s *CronJobServiceServer) GetCronJob(ctx context.Context, req *connect.Request[agentsv1.GetCronJobRequest]) (*connect.Response[agentsv1.GetCronJobResponse], error) {
	if s.scheduler == nil {
		return nil, connectx.NotFound("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	job, err := s.scheduler.GetJob(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.GetCronJobResponse{CronJob: job}), nil
}

func (s *CronJobServiceServer) CreateCronJob(ctx context.Context, req *connect.Request[agentsv1.CreateCronJobRequest]) (*connect.Response[agentsv1.CreateCronJobResponse], error) {
	if s.scheduler == nil {
		return nil, connectx.Internal("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	job := req.Msg.GetCronJob()
	job.WorkspaceId = wsID
	logger := log.FromContext(ctx)
	logger.Info("creating cron job", "workspace_id", wsID, "name", job.GetName(), "schedule", job.GetSchedule())
	if err := s.scheduler.AddJob(ctx, job); err != nil {
		logger.Error("create cron job failed", "workspace_id", wsID, "name", job.GetName(), "err", err)
		return nil, cronJobMutationError(err)
	}
	logger.Info("cron job created", "workspace_id", wsID, "name", job.GetName())
	return connect.NewResponse(&agentsv1.CreateCronJobResponse{CronJob: job}), nil
}

func (s *CronJobServiceServer) UpdateCronJob(ctx context.Context, req *connect.Request[agentsv1.UpdateCronJobRequest]) (*connect.Response[agentsv1.UpdateCronJobResponse], error) {
	if s.scheduler == nil {
		return nil, connectx.Internal("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	job := req.Msg.GetCronJob()
	job.WorkspaceId = wsID
	logger := log.FromContext(ctx)
	logger.Info("updating cron job", "workspace_id", wsID, "name", job.GetName(), "schedule", job.GetSchedule())
	if err := s.scheduler.UpdateJob(ctx, job); err != nil {
		logger.Error("update cron job failed", "workspace_id", wsID, "name", job.GetName(), "err", err)
		return nil, cronJobMutationError(err)
	}
	logger.Info("cron job updated", "workspace_id", wsID, "name", job.GetName())
	return connect.NewResponse(&agentsv1.UpdateCronJobResponse{CronJob: job}), nil
}

func (s *CronJobServiceServer) DeleteCronJob(ctx context.Context, req *connect.Request[agentsv1.DeleteCronJobRequest]) (*connect.Response[agentsv1.DeleteCronJobResponse], error) {
	if s.scheduler == nil {
		return nil, connectx.Internal("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	job, err := s.scheduler.GetJob(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	logger.Info("deleting cron job", "workspace_id", wsID, "name", req.Msg.GetName())
	if err := s.scheduler.RemoveJob(ctx, wsID, req.Msg.GetName()); err != nil {
		logger.Error("delete cron job failed", "workspace_id", wsID, "name", req.Msg.GetName(), "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("cron job deleted", "workspace_id", wsID, "name", req.Msg.GetName())
	return connect.NewResponse(&agentsv1.DeleteCronJobResponse{CronJob: job}), nil
}

func (s *CronJobServiceServer) RunCronJobNow(ctx context.Context, req *connect.Request[agentsv1.RunCronJobNowRequest]) (*connect.Response[agentsv1.RunCronJobNowResponse], error) {
	if s.scheduler == nil {
		return nil, connectx.Internal("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("running cron job manually", "workspace_id", wsID, "name", req.Msg.GetName())
	exec, err := s.scheduler.RunJobNow(ctx, wsID, req.Msg.GetName())
	if err != nil {
		logger.Error("run cron job manually failed", "workspace_id", wsID, "name", req.Msg.GetName(), "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("cron job manual run started", "workspace_id", wsID, "name", req.Msg.GetName(), "execution_id", exec.GetId())
	return connect.NewResponse(&agentsv1.RunCronJobNowResponse{Execution: exec}), nil
}

func (s *CronJobServiceServer) ListCronExecutions(ctx context.Context, req *connect.Request[agentsv1.ListCronExecutionsRequest]) (*connect.Response[agentsv1.ListCronExecutionsResponse], error) {
	if s.execRepo == nil {
		return connect.NewResponse(&agentsv1.ListCronExecutionsResponse{}), nil
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	executions, nextToken, err := s.execRepo.List(ctx, wsID, req.Msg.GetJobName(), req.Msg.GetPageSize(), req.Msg.GetPageToken())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.ListCronExecutionsResponse{
		Executions:    executions,
		NextPageToken: nextToken,
	}), nil
}
