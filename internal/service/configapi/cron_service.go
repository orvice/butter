package configapi

import (
	"context"

	"go.orx.me/apps/butter/internal/cron"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// CronJobServiceServer implements the CronJobService Twirp interface.
type CronJobServiceServer struct {
	scheduler *cron.Scheduler
	repo      cron.ExecutionRepo
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
	s.repo = repo
}

func (s *CronJobServiceServer) ListCronJobs(_ context.Context, _ *agentsv1.ListCronJobsRequest) (*agentsv1.ListCronJobsResponse, error) {
	if s.scheduler == nil {
		return &agentsv1.ListCronJobsResponse{}, nil
	}
	jobs := s.scheduler.Jobs()
	ptrs := make([]*agentsv1.CronJob, len(jobs))
	for i := range jobs {
		ptrs[i] = &jobs[i]
	}
	return &agentsv1.ListCronJobsResponse{CronJobs: ptrs}, nil
}

func (s *CronJobServiceServer) ListCronExecutions(ctx context.Context, req *agentsv1.ListCronExecutionsRequest) (*agentsv1.ListCronExecutionsResponse, error) {
	if s.repo == nil {
		return &agentsv1.ListCronExecutionsResponse{}, nil
	}
	executions, nextToken, err := s.repo.List(ctx, req.GetJobName(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, err
	}
	return &agentsv1.ListCronExecutionsResponse{
		Executions:    executions,
		NextPageToken: nextToken,
	}, nil
}
