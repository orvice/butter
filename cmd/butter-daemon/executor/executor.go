package executor

import (
	"context"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Executor executes daemon tasks for a specific capability.
type Executor interface {
	// Execute runs the task and reports progress via onUpdate.
	// It should respect ctx cancellation.
	Execute(ctx context.Context, task *agentsv1.DaemonTask, onUpdate func(*agentsv1.DaemonTaskUpdate)) error
	// Capability returns the capability string this executor handles.
	Capability() string
}
