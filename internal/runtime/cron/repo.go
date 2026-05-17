package cron

import (
	"context"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// JobRepo persists cron job configurations.
type JobRepo interface {
	List(ctx context.Context) ([]*agentsv1.CronJob, error)
	Get(ctx context.Context, name string) (*agentsv1.CronJob, error)
	Create(ctx context.Context, job *agentsv1.CronJob) error
	Update(ctx context.Context, job *agentsv1.CronJob) error
	Delete(ctx context.Context, name string) error
}

// ExecutionRepo persists cron job execution records.
type ExecutionRepo interface {
	Save(ctx context.Context, exec *agentsv1.CronExecution) error
	List(ctx context.Context, jobName string, pageSize int32, pageToken string) ([]*agentsv1.CronExecution, string, error)
	GetByID(ctx context.Context, id string) (*agentsv1.CronExecution, error)
	// ListByTimeRange returns executions whose started_at falls within
	// [start, end). Optional jobName filter. Implementations should return
	// results in ascending order by started_at; callers may bucket as needed.
	ListByTimeRange(ctx context.Context, jobName string, start, end time.Time) ([]*agentsv1.CronExecution, error)
}
