package cron

import (
	"context"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// JobRepo persists cron job configurations, scoped per-workspace.
type JobRepo interface {
	List(ctx context.Context, workspaceID string) ([]*agentsv1.CronJob, error)
	ListAll(ctx context.Context) ([]*agentsv1.CronJob, error)
	Get(ctx context.Context, workspaceID, name string) (*agentsv1.CronJob, error)
	Create(ctx context.Context, job *agentsv1.CronJob) error
	Update(ctx context.Context, job *agentsv1.CronJob) error
	Delete(ctx context.Context, workspaceID, name string) error
}

// ExecutionRepo persists cron job execution records.
type ExecutionRepo interface {
	Save(ctx context.Context, exec *agentsv1.CronExecution) error
	List(ctx context.Context, workspaceID, jobName string, pageSize int32, pageToken string) ([]*agentsv1.CronExecution, string, error)
	GetByID(ctx context.Context, id string) (*agentsv1.CronExecution, error)
	// ListWaitingBySession returns the WAITING_INPUT executions whose stored
	// session coordinates match, ordered oldest first. Used to close paused
	// executions when a human reply completes the workflow on that session.
	ListWaitingBySessionAcrossWorkspaces(ctx context.Context, appName, userID, sessionID string) ([]*agentsv1.CronExecution, error)
	// ListByTimeRange returns executions whose started_at falls within
	// [start, end). Optional workspace/jobName filters. Implementations should
	// return results in ascending order by started_at; callers may bucket as
	// needed.
	ListByTimeRange(ctx context.Context, workspaceID, jobName string, start, end time.Time) ([]*agentsv1.CronExecution, error)
}
