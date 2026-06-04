package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"go.orx.me/apps/butter/cmd/butter-daemon/executor"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// daemonVersion is the semantic version of the daemon client binary. It is
// surfaced at registration so the server-side dashboard can display it.
const daemonVersion = "v0.1.0"

// Connector manages the gRPC connection to the butter server.
type Connector struct {
	cfg       *Config
	executors map[string]executor.Executor

	mu          sync.Mutex
	cancelFuncs map[string]context.CancelFunc // task_id → cancel

	sendMu sync.Mutex // serializes stream.Send across concurrent tasks/callbacks
}

// NewConnector creates a new daemon connector.
func NewConnector(cfg *Config, executors []executor.Executor) *Connector {
	execMap := make(map[string]executor.Executor, len(executors))
	for _, e := range executors {
		execMap[e.Capability()] = e
	}
	return &Connector{
		cfg:         cfg,
		executors:   execMap,
		cancelFuncs: make(map[string]context.CancelFunc),
	}
}

// Run connects to the server and processes tasks. It reconnects on failure
// with exponential backoff.
func (c *Connector) Run(ctx context.Context) error {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		err := c.connectAndServe(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		slog.Error("connection lost", "err", err)

		slog.Info("reconnecting", "backoff", backoff)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (c *Connector) connectAndServe(ctx context.Context) error {
	conn, err := grpc.NewClient(c.cfg.Server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	// Attach auth token.
	streamCtx := ctx
	if c.cfg.Token != "" {
		streamCtx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.cfg.Token)
	}

	client := agentsv1.NewDaemonConnectorServiceClient(conn)
	stream, err := client.Connect(streamCtx)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	// Send registration.
	capabilities := make([]string, 0, len(c.executors))
	for cap := range c.executors {
		capabilities = append(capabilities, cap)
	}

	err = stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{
				DaemonId:     c.cfg.DaemonID,
				Name:         c.cfg.Name,
				Capabilities: capabilities,
				Labels:       c.cfg.Labels,
				Version:      daemonVersion,
				Os:           runtime.GOOS + "-" + runtime.GOARCH,
				Executors:    capabilities,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	slog.Info("registered with server",
		"daemon_id", c.cfg.DaemonID,
		"capabilities", capabilities,
	)

	// Receive loop.
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("server closed stream")
			}
			return fmt.Errorf("recv: %w", err)
		}

		switch m := msg.Message.(type) {
		case *agentsv1.ConnectResponse_Task:
			go c.handleTask(ctx, stream, m.Task)
		case *agentsv1.ConnectResponse_Cancel:
			c.handleCancel(m.Cancel.TaskId)
		}
	}
}

func (c *Connector) handleTask(ctx context.Context, stream agentsv1.DaemonConnectorService_ConnectClient, task *agentsv1.DaemonTask) {
	slog.Info("task received", "task_id", task.TaskId, "capability", task.Capability, "input_len", len(task.Input))

	exec, ok := c.executors[task.Capability]
	if !ok {
		c.sendUpdate(stream, &agentsv1.DaemonTaskUpdate{
			TaskId: task.TaskId,
			Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED,
			Error:  fmt.Sprintf("unsupported capability: %s", task.Capability),
		})
		return
	}

	taskCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancelFuncs[task.TaskId] = cancel
	c.mu.Unlock()

	defer func() {
		cancel()
		c.mu.Lock()
		delete(c.cancelFuncs, task.TaskId)
		c.mu.Unlock()
	}()

	onUpdate := func(update *agentsv1.DaemonTaskUpdate) {
		c.sendUpdate(stream, update)
	}

	if err := exec.Execute(taskCtx, task, onUpdate); err != nil {
		slog.Error("task execution error", "task_id", task.TaskId, "err", err)
	}
}

func (c *Connector) handleCancel(taskID string) {
	c.mu.Lock()
	cancel, ok := c.cancelFuncs[taskID]
	c.mu.Unlock()

	if ok {
		slog.Info("cancelling task", "task_id", taskID)
		cancel()
	}
}

func (c *Connector) sendUpdate(stream agentsv1.DaemonConnectorService_ConnectClient, update *agentsv1.DaemonTaskUpdate) {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_TaskUpdate{TaskUpdate: update},
	}); err != nil {
		slog.Error("failed to send update", "task_id", update.TaskId, "err", err)
	}
}
