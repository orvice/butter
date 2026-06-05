package executor

import (
	"context"
	"fmt"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Executor executes daemon tasks for a specific ACP runtime.
type Executor interface {
	// Execute runs the task and reports progress via onUpdate.
	// It should respect ctx cancellation.
	Execute(ctx context.Context, task *agentsv1.DaemonTask, onUpdate func(*agentsv1.DaemonTaskUpdate)) error
	// Runtime returns the ACP runtime string this executor handles.
	Runtime() string
}

func sendFailed(taskID, errMsg string, onUpdate func(*agentsv1.DaemonTaskUpdate)) error {
	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: taskID,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED,
		Error:  errMsg,
	})
	return fmt.Errorf("%s", errMsg)
}
