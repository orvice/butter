package executor

import (
	"context"
	"fmt"
	"os/exec"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ShellExecutor runs arbitrary shell commands.
type ShellExecutor struct {
	WorkDir string
}

func (e *ShellExecutor) Runtime() string { return "shell" }

func (e *ShellExecutor) Execute(ctx context.Context, task *agentsv1.DaemonTask, onUpdate func(*agentsv1.DaemonTaskUpdate)) error {
	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_ACCEPTED,
	})

	cmd := exec.CommandContext(ctx, "sh", "-c", task.Input)
	workDir := task.GetWorkDir()
	if workDir == "" {
		workDir = e.WorkDir
	}
	if workDir != "" {
		cmd.Dir = workDir
	}

	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
	})

	output, err := cmd.CombinedOutput()
	if err != nil {
		return sendFailed(task.TaskId, fmt.Sprintf("shell: %v\n%s", err, string(output)), onUpdate)
	}

	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED,
		Output: string(output),
	})
	return nil
}
