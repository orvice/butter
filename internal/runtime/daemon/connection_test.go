package daemon

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestConnectionSendAndDispatch(t *testing.T) {
	conn := NewConnection(&agentsv1.DaemonInfo{DaemonId: "d1", Capabilities: []string{"opencode"}})

	task := &agentsv1.DaemonTask{TaskId: "t1", AgentName: "coder", Input: "fix bug", Capability: "opencode"}
	resultCh, err := conn.SendTask(task)
	if err != nil {
		t.Fatalf("SendTask: %v", err)
	}

	// Drain the server message that was sent.
	msg := <-conn.SendCh
	if msg.GetTask().GetTaskId() != "t1" {
		t.Fatalf("expected task t1, got %v", msg)
	}

	// Dispatch a running update.
	conn.DispatchUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: "t1",
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
		Output: "progress",
	})

	update := <-resultCh
	if update.Status != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING {
		t.Fatalf("expected RUNNING, got %v", update.Status)
	}

	// Dispatch completed.
	conn.DispatchUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: "t1",
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED,
		Output: "done",
	})

	update = <-resultCh
	if update.Status != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED {
		t.Fatalf("expected COMPLETED, got %v", update.Status)
	}
	if update.Output != "done" {
		t.Fatalf("expected 'done', got %q", update.Output)
	}

	// Channel should be closed after terminal status.
	_, ok := <-resultCh
	if ok {
		t.Fatal("expected channel to be closed")
	}
}

func TestConnectionCloseNotifiesWaiters(t *testing.T) {
	conn := NewConnection(&agentsv1.DaemonInfo{DaemonId: "d1"})
	task := &agentsv1.DaemonTask{TaskId: "t1"}
	resultCh, err := conn.SendTask(task)
	if err != nil {
		t.Fatalf("SendTask: %v", err)
	}
	<-conn.SendCh // drain

	conn.Close()

	update := <-resultCh
	if update.Status != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED {
		t.Fatalf("expected FAILED on disconnect, got %v", update.Status)
	}
	if update.Error != "daemon disconnected" {
		t.Fatalf("expected disconnect error, got %q", update.Error)
	}
}

func TestConnectionSendTaskAfterClose(t *testing.T) {
	conn := NewConnection(&agentsv1.DaemonInfo{DaemonId: "d1"})
	conn.Close()

	_, err := conn.SendTask(&agentsv1.DaemonTask{TaskId: "t1"})
	if err != ErrDaemonDisconnected {
		t.Fatalf("expected ErrDaemonDisconnected, got %v", err)
	}
}

func TestConnectionDispatchUnknownTask(t *testing.T) {
	conn := NewConnection(&agentsv1.DaemonInfo{DaemonId: "d1"})
	// Should not panic when dispatching for an unknown task.
	conn.DispatchUpdate(&agentsv1.DaemonTaskUpdate{TaskId: "unknown", Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED})
}
