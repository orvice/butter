package executor

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// OpenCodeExecutor invokes the opencode CLI to execute tasks.
type OpenCodeExecutor struct {
	WorkDir string
	Binary  string // defaults to "opencode"
}

func (e *OpenCodeExecutor) Capability() string { return "opencode" }

func (e *OpenCodeExecutor) Execute(ctx context.Context, task *agentsv1.DaemonTask, onUpdate func(*agentsv1.DaemonTaskUpdate)) error {
	binary := e.Binary
	if binary == "" {
		binary = "opencode"
	}

	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_ACCEPTED,
	})

	cmd := exec.CommandContext(ctx, binary, "--non-interactive", "--prompt", task.Input)
	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return sendFailed(task.TaskId, fmt.Sprintf("stdout pipe: %v", err), onUpdate)
	}

	if err := cmd.Start(); err != nil {
		return sendFailed(task.TaskId, fmt.Sprintf("start: %v", err), onUpdate)
	}

	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
	})

	var output string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		output += scanner.Text() + "\n"
		onUpdate(&agentsv1.DaemonTaskUpdate{
			TaskId: task.TaskId,
			Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
			Output: output,
		})
	}

	if err := cmd.Wait(); err != nil {
		return sendFailed(task.TaskId, fmt.Sprintf("exit: %v\n%s", err, output), onUpdate)
	}

	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED,
		Output: output,
	})
	return nil
}

func sendFailed(taskID, errMsg string, onUpdate func(*agentsv1.DaemonTaskUpdate)) error {
	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: taskID,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED,
		Error:  errMsg,
	})
	return fmt.Errorf("%s", errMsg)
}
