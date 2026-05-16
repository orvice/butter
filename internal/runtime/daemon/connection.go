package daemon

import (
	"errors"
	"sync"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ErrDaemonDisconnected is returned when the daemon connection drops while a
// task is in progress.
var ErrDaemonDisconnected = errors.New("daemon disconnected")

// Connection represents a single connected daemon and its bidirectional
// communication channels.
type Connection struct {
	Info        *agentsv1.DaemonInfo
	SendCh      chan *agentsv1.ConnectResponse // server → daemon
	ConnectedAt time.Time
	// RemoteAddr is the peer address captured at handshake; empty if the gRPC
	// transport did not surface peer info.
	RemoteAddr string

	mu          sync.Mutex
	activeTasks map[string]chan *agentsv1.DaemonTaskUpdate // task_id → result channel
	closed      bool
}

// NewConnection creates a connection for the given daemon info.
func NewConnection(info *agentsv1.DaemonInfo) *Connection {
	return &Connection{
		Info:        info,
		SendCh:      make(chan *agentsv1.ConnectResponse, 16),
		ConnectedAt: time.Now(),
		activeTasks: make(map[string]chan *agentsv1.DaemonTaskUpdate),
	}
}

// SendTask pushes a task to the daemon and returns a channel that will receive
// status updates. The caller must read from the channel until a terminal status.
func (c *Connection) SendTask(task *agentsv1.DaemonTask) (<-chan *agentsv1.DaemonTaskUpdate, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, ErrDaemonDisconnected
	}
	resultCh := make(chan *agentsv1.DaemonTaskUpdate, 16)
	c.activeTasks[task.TaskId] = resultCh
	c.mu.Unlock()

	c.SendCh <- &agentsv1.ConnectResponse{
		Message: &agentsv1.ConnectResponse_Task{Task: task},
	}
	return resultCh, nil
}

// CancelTask sends a cancel request to the daemon for the given task.
func (c *Connection) CancelTask(taskID string) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrDaemonDisconnected
	}
	c.mu.Unlock()

	c.SendCh <- &agentsv1.ConnectResponse{
		Message: &agentsv1.ConnectResponse_Cancel{Cancel: &agentsv1.CancelTask{TaskId: taskID}},
	}
	return nil
}

// DispatchUpdate routes a daemon task update to the corresponding task's
// result channel. If the task is not found (already completed or cancelled),
// the update is silently dropped.
func (c *Connection) DispatchUpdate(update *agentsv1.DaemonTaskUpdate) {
	c.mu.Lock()
	ch, ok := c.activeTasks[update.TaskId]
	// Remove the task on terminal status.
	if ok && isTerminal(update.Status) {
		delete(c.activeTasks, update.TaskId)
	}
	c.mu.Unlock()

	if ok {
		ch <- update
		if isTerminal(update.Status) {
			close(ch)
		}
	}
}

// ActiveTaskCount returns the number of tasks currently in flight on this
// connection. Safe to call concurrently.
func (c *Connection) ActiveTaskCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.activeTasks)
}

// HasTask reports whether this connection is currently tracking the given task.
func (c *Connection) HasTask(taskID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.activeTasks[taskID]
	return ok
}

// ActiveTaskIDs returns the ids of all in-flight tasks on this connection.
// Used by the dashboard to surface active work.
func (c *Connection) ActiveTaskIDs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, 0, len(c.activeTasks))
	for id := range c.activeTasks {
		out = append(out, id)
	}
	return out
}

// Close marks the connection as closed and notifies all active task waiters
// that the daemon has disconnected.
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true

	for id, ch := range c.activeTasks {
		ch <- &agentsv1.DaemonTaskUpdate{
			TaskId: id,
			Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED,
			Error:  "daemon disconnected",
		}
		close(ch)
		delete(c.activeTasks, id)
	}
}

func isTerminal(s agentsv1.DaemonTaskStatus) bool {
	switch s {
	case agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED,
		agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED,
		agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_CANCELLED:
		return true
	}
	return false
}
