package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/types/known/emptypb"

	"go.orx.me/apps/butter/cmd/butter-daemon/executor"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

var daemonClientHeartbeatInterval = 20 * time.Second

type connectStream = connect.BidiStreamForClient[agentsv1.ConnectRequest, agentsv1.ConnectResponse]

// Connector manages the ConnectRPC connection to the butter server.
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
		execMap[e.Runtime()] = e
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
	baseURL := normalizeBaseURL(c.cfg.Server)
	client := agentsv1connect.NewDaemonConnectorServiceClient(newHTTPClient(baseURL), baseURL)

	stream := client.Connect(ctx)
	if c.cfg.Credential != "" {
		stream.RequestHeader().Set("Authorization", "Bearer "+c.cfg.Credential)
	}

	// Send registration.
	runtimes := make([]string, 0, len(c.executors))
	for runtime := range c.executors {
		runtimes = append(runtimes, runtime)
	}

	err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{
				Name:        c.cfg.Name,
				AcpRuntimes: runtimes,
				Labels:      c.cfg.Labels,
				Version:     daemonBuildVersion(),
				Os:          runtime.GOOS + "-" + runtime.GOARCH,
				Executors:   runtimes,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	slog.Info("registered with server",
		"acp_runtimes", runtimes,
	)

	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	defer stopHeartbeat()
	go c.sendHeartbeat(heartbeatCtx, stream)

	// Receive loop.
	for {
		msg, err := stream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("server closed stream")
			}
			return fmt.Errorf("recv: %w", err)
		}

		switch m := msg.Message.(type) {
		case *agentsv1.ConnectResponse_Task:
			go c.handleTask(ctx, stream, m.Task)
		case *agentsv1.ConnectResponse_Cancel:
			c.handleCancel(m.Cancel.TaskId)
		case *agentsv1.ConnectResponse_Heartbeat:
			// Heartbeat response keeps the downstream side of the stream active.
		}
	}
}

func (c *Connector) sendHeartbeat(ctx context.Context, stream *connectStream) {
	ticker := time.NewTicker(daemonClientHeartbeatInterval)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ticker.C:
			c.sendMu.Lock()
			err := stream.Send(&agentsv1.ConnectRequest{
				Message: &agentsv1.ConnectRequest_Heartbeat{Heartbeat: &emptypb.Empty{}},
			})
			c.sendMu.Unlock()
			if err != nil {
				slog.Warn("failed to send daemon heartbeat", "err", err)
				return
			}
			count++
			if count <= 3 {
				slog.Info("daemon heartbeat sent", "count", count)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Connector) handleTask(ctx context.Context, stream *connectStream, task *agentsv1.DaemonTask) {
	slog.Info("task received", "task_id", task.TaskId, "acp_runtime", task.AcpRuntime, "work_dir", task.WorkDir, "input_len", len(task.Input))

	exec, ok := c.executors[task.AcpRuntime]
	if !ok {
		c.sendUpdate(stream, &agentsv1.DaemonTaskUpdate{
			TaskId: task.TaskId,
			Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED,
			Error:  fmt.Sprintf("unsupported acp_runtime: %s", task.AcpRuntime),
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

func (c *Connector) sendUpdate(stream *connectStream, update *agentsv1.DaemonTaskUpdate) {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_TaskUpdate{TaskUpdate: update},
	}); err != nil {
		slog.Error("failed to send update", "task_id", update.TaskId, "err", err)
	}
}

// normalizeBaseURL turns a bare host/path into a ConnectRPC base URL,
// defaulting to cleartext HTTP.
func normalizeBaseURL(server string) string {
	if strings.HasPrefix(server, "http://") || strings.HasPrefix(server, "https://") {
		return server
	}
	return "http://" + server
}

// newHTTPClient builds an HTTP/2 client. For https URLs it relies on ALPN; for
// http URLs it speaks h2c (prior-knowledge cleartext HTTP/2) to match the
// daemon connect server.
func newHTTPClient(baseURL string) *http.Client {
	if strings.HasPrefix(baseURL, "https://") {
		return &http.Client{Transport: &http2.Transport{}}
	}
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}
}
