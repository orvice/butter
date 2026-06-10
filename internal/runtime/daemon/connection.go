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

// ErrRuntimeAlreadyConnected is returned when a second connection tries to
// register the same workspace/runtime pair.
var ErrRuntimeAlreadyConnected = errors.New("daemon runtime already connected")

// taskState tracks a single in-flight task on a daemon connection.
type taskState struct {
	mu          sync.Mutex
	resultCh    chan *agentsv1.DaemonTaskUpdate
	closed      bool
	startedAt   time.Time
	agentName   string
	acpRuntime  string
	workspaceID string
	currentStep string
	progress    int32
}

func (t *taskState) deliver(update *agentsv1.DaemonTaskUpdate, closeAfter bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return
	}
	select {
	case t.resultCh <- update:
	default:
		// Do not let a stopped or slow consumer wedge the connection. Drop one
		// stale buffered progress update so the latest update (including terminal
		// or disconnect) can be observed.
		select {
		case <-t.resultCh:
		default:
		}
		select {
		case t.resultCh <- update:
		default:
		}
	}
	if closeAfter {
		t.closed = true
		close(t.resultCh)
	}
}

func (t *taskState) closeWith(update *agentsv1.DaemonTaskUpdate) {
	t.deliver(update, true)
}

// TaskSnapshot is a point-in-time view of a single in-flight task suitable
// for the dashboard.
type TaskSnapshot struct {
	TaskID      string
	AgentName   string
	AcpRuntime  string
	WorkspaceID string
	StartedAt   time.Time
	CurrentStep string
	Progress    int32
}

// Connection represents a single connected daemon and its bidirectional
// communication channels.
type Connection struct {
	Info        *agentsv1.DaemonInfo
	WorkspaceID string
	SendCh      chan *agentsv1.ConnectResponse // server → daemon
	ConnectedAt time.Time
	LastSeen    time.Time
	PollMode    bool
	// RemoteAddr is the peer address captured at handshake; empty if the gRPC
	// transport did not surface peer info.
	RemoteAddr string

	mu          sync.Mutex
	activeTasks map[string]*taskState
	closed      bool
}

// NewConnection creates a connection for the given daemon info.
func NewConnection(info *agentsv1.DaemonInfo) *Connection {
	now := time.Now()
	return &Connection{
		Info:        info,
		WorkspaceID: info.GetWorkspaceId(),
		SendCh:      make(chan *agentsv1.ConnectResponse, 16),
		ConnectedAt: now,
		LastSeen:    now,
		activeTasks: make(map[string]*taskState),
	}
}

// MarkPollMode identifies this connection as backed by unary long polling.
func (c *Connection) MarkPollMode() {
	c.mu.Lock()
	c.PollMode = true
	c.LastSeen = time.Now()
	c.mu.Unlock()
}

// Touch records daemon activity for poll-mode liveness checks.
func (c *Connection) Touch() {
	c.mu.Lock()
	if !c.closed {
		c.LastSeen = time.Now()
	}
	c.mu.Unlock()
}

func (c *Connection) stalePollConnection(now time.Time, maxIdle time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.PollMode && !c.closed && maxIdle > 0 && now.Sub(c.LastSeen) > maxIdle
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
	c.activeTasks[task.TaskId] = &taskState{
		resultCh:    resultCh,
		startedAt:   time.Now(),
		agentName:   task.GetAgentName(),
		acpRuntime:  task.GetAcpRuntime(),
		workspaceID: task.GetWorkspaceId(),
	}
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
	terminal := isTerminal(update.Status)

	c.mu.Lock()
	t, ok := c.activeTasks[update.TaskId]
	if ok {
		if step := update.GetCurrentStep(); step != "" {
			t.currentStep = step
		}
		if p := update.GetProgress(); p > 0 {
			t.progress = p
		}
		if terminal {
			delete(c.activeTasks, update.TaskId)
		}
	}
	c.mu.Unlock()

	if ok {
		t.deliver(update, terminal)
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

// ActiveTaskSnapshots returns a copy of all in-flight tasks with their
// latest progress fields for the dashboard.
func (c *Connection) ActiveTaskSnapshots() []TaskSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]TaskSnapshot, 0, len(c.activeTasks))
	for id, t := range c.activeTasks {
		out = append(out, TaskSnapshot{
			TaskID:      id,
			AgentName:   t.agentName,
			AcpRuntime:  t.acpRuntime,
			WorkspaceID: t.workspaceID,
			StartedAt:   t.startedAt,
			CurrentStep: t.currentStep,
			Progress:    t.progress,
		})
	}
	return out
}

// Close marks the connection as closed and notifies all active task waiters
// that the daemon has disconnected.
func (c *Connection) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true

	tasks := make(map[string]*taskState, len(c.activeTasks))
	for id, t := range c.activeTasks {
		tasks[id] = t
		delete(c.activeTasks, id)
	}
	c.mu.Unlock()

	for id, t := range tasks {
		t.closeWith(&agentsv1.DaemonTaskUpdate{
			TaskId: id,
			Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED,
			Error:  "daemon disconnected",
		})
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
