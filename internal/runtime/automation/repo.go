package automation

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrAutomationNotFound      = errors.New("automation not found")
	ErrAutomationAlreadyExists = errors.New("automation already exists")
	ErrRunNotFound             = errors.New("automation run not found")
	ErrStepRunNotFound         = errors.New("automation step run not found")
)

// DefinitionRepo persists workspace-scoped automation definitions.
type DefinitionRepo interface {
	EnsureIndexes(ctx context.Context) error
	List(ctx context.Context, workspaceID string) ([]*agentsv1.Automation, error)
	ListAll(ctx context.Context) ([]*agentsv1.Automation, error)
	Get(ctx context.Context, workspaceID, name string) (*agentsv1.Automation, error)
	Create(ctx context.Context, automation *agentsv1.Automation) error
	Update(ctx context.Context, automation *agentsv1.Automation) error
	Delete(ctx context.Context, workspaceID, name string) (*agentsv1.Automation, error)
}

// RunRepo persists automation run records.
type RunRepo interface {
	EnsureIndexes(ctx context.Context) error
	Save(ctx context.Context, run *agentsv1.AutomationRun) error
	Get(ctx context.Context, workspaceID, id string) (*agentsv1.AutomationRun, error)
	List(ctx context.Context, workspaceID, automationName string, pageSize int32, pageToken string) ([]*agentsv1.AutomationRun, string, error)
}

// StepRunRepo persists step-level run records.
type StepRunRepo interface {
	EnsureIndexes(ctx context.Context) error
	Save(ctx context.Context, stepRun *agentsv1.AutomationStepRun) error
	Get(ctx context.Context, workspaceID, id string) (*agentsv1.AutomationStepRun, error)
	ListByRun(ctx context.Context, workspaceID, runID string) ([]*agentsv1.AutomationStepRun, error)
}
