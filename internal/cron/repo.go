package cron

import (
	"context"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ExecutionRepo persists cron job execution records.
type ExecutionRepo interface {
	Save(ctx context.Context, exec *agentsv1.CronExecution) error
	List(ctx context.Context, jobName string, pageSize int32, pageToken string) ([]*agentsv1.CronExecution, string, error)
	GetByID(ctx context.Context, id string) (*agentsv1.CronExecution, error)
}
