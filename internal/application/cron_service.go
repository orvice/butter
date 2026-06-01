package application

import (
	"context"
	"errors"

	"butterfly.orx.me/core/log"
	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// cronJobMutationError maps scheduler errors to typed twirp errors so that
// validation failures (e.g. agent does not exist in workspace) surface as
// InvalidArgument instead of being swallowed by toTwirpError's catch-all
// InternalErrorWith.
func cronJobMutationError(err error) error {
	if errors.Is(err, cron.ErrAgentNotInWorkspace) {
		return twirp.InvalidArgumentError("agent_name", err.Error())
	}
	return toTwirpError(err)
}

// CronJobServiceServer implements the CronJobService Twirp interface.
type CronJobServiceServer struct {
	agentsv1.UnimplementedCronJobServiceServer

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

func requireWorkspace(ctx context.Context) (string, error) {
	id, ok := workspace.FromContext(ctx)
	if !ok {
		return "", twirp.NewError(twirp.FailedPrecondition, "workspace required (set X-Workspace-ID header)")
	}
	return id, nil
}

func (s *CronJobServiceServer) ListCronJobs(ctx context.Context, _ *agentsv1.ListCronJobsRequest) (*agentsv1.ListCronJobsResponse, error) {
	if s.scheduler == nil {
		return &agentsv1.ListCronJobsResponse{}, nil
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	jobs, err := s.scheduler.ListJobs(ctx, wsID)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ListCronJobsResponse{CronJobs: jobs}, nil
}

func (s *CronJobServiceServer) GetCronJob(ctx context.Context, req *agentsv1.GetCronJobRequest) (*agentsv1.GetCronJobResponse, error) {
	if s.scheduler == nil {
		return nil, twirp.NotFoundError("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	job, err := s.scheduler.GetJob(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetCronJobResponse{CronJob: job}, nil
}

func (s *CronJobServiceServer) CreateCronJob(ctx context.Context, req *agentsv1.CreateCronJobRequest) (*agentsv1.CreateCronJobResponse, error) {
	if s.scheduler == nil {
		return nil, twirp.InternalError("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	job := req.GetCronJob()
	job.WorkspaceId = wsID
	logger := log.FromContext(ctx)
	logger.Info("creating cron job", "workspace_id", wsID, "name", job.GetName(), "schedule", job.GetSchedule())
	if err := s.scheduler.AddJob(ctx, job); err != nil {
		logger.Error("create cron job failed", "workspace_id", wsID, "name", job.GetName(), "err", err)
		return nil, cronJobMutationError(err)
	}
	logger.Info("cron job created", "workspace_id", wsID, "name", job.GetName())
	return &agentsv1.CreateCronJobResponse{CronJob: job}, nil
}

func (s *CronJobServiceServer) UpdateCronJob(ctx context.Context, req *agentsv1.UpdateCronJobRequest) (*agentsv1.UpdateCronJobResponse, error) {
	if s.scheduler == nil {
		return nil, twirp.InternalError("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	job := req.GetCronJob()
	job.WorkspaceId = wsID
	logger := log.FromContext(ctx)
	logger.Info("updating cron job", "workspace_id", wsID, "name", job.GetName(), "schedule", job.GetSchedule())
	if err := s.scheduler.UpdateJob(ctx, job); err != nil {
		logger.Error("update cron job failed", "workspace_id", wsID, "name", job.GetName(), "err", err)
		return nil, cronJobMutationError(err)
	}
	logger.Info("cron job updated", "workspace_id", wsID, "name", job.GetName())
	return &agentsv1.UpdateCronJobResponse{CronJob: job}, nil
}

func (s *CronJobServiceServer) DeleteCronJob(ctx context.Context, req *agentsv1.DeleteCronJobRequest) (*agentsv1.DeleteCronJobResponse, error) {
	if s.scheduler == nil {
		return nil, twirp.InternalError("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	job, err := s.scheduler.GetJob(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("deleting cron job", "workspace_id", wsID, "name", req.GetName())
	if err := s.scheduler.RemoveJob(ctx, wsID, req.GetName()); err != nil {
		logger.Error("delete cron job failed", "workspace_id", wsID, "name", req.GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("cron job deleted", "workspace_id", wsID, "name", req.GetName())
	return &agentsv1.DeleteCronJobResponse{CronJob: job}, nil
}

func (s *CronJobServiceServer) RunCronJobNow(ctx context.Context, req *agentsv1.RunCronJobNowRequest) (*agentsv1.RunCronJobNowResponse, error) {
	if s.scheduler == nil {
		return nil, twirp.InternalError("cron scheduler not initialized")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("running cron job manually", "workspace_id", wsID, "name", req.GetName())
	exec, err := s.scheduler.RunJobNow(ctx, wsID, req.GetName())
	if err != nil {
		logger.Error("run cron job manually failed", "workspace_id", wsID, "name", req.GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("cron job manual run started", "workspace_id", wsID, "name", req.GetName(), "execution_id", exec.GetId())
	return &agentsv1.RunCronJobNowResponse{Execution: exec}, nil
}

func (s *CronJobServiceServer) ListCronExecutions(ctx context.Context, req *agentsv1.ListCronExecutionsRequest) (*agentsv1.ListCronExecutionsResponse, error) {
	if s.execRepo == nil {
		return &agentsv1.ListCronExecutionsResponse{}, nil
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	executions, nextToken, err := s.execRepo.List(ctx, wsID, req.GetJobName(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ListCronExecutionsResponse{
		Executions:    executions,
		NextPageToken: nextToken,
	}, nil
}
