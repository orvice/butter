package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/types/known/durationpb"

	"go.orx.me/apps/butter/cmd/butter-daemon/executor"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

var daemonPollWaitTimeout = 25 * time.Second

// Connector manages the ConnectRPC connection to the butter server.
type Connector struct {
	cfg       *Config
	executors map[string]executor.Executor

	mu          sync.Mutex
	cancelFuncs map[string]context.CancelFunc // task_id → cancel
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
	serveCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	runtimes := make([]string, 0, len(c.executors))
	for runtime := range c.executors {
		runtimes = append(runtimes, runtime)
	}

	registerReq := connect.NewRequest(&agentsv1.DaemonConnectorServiceRegisterRequest{
		Daemon: &agentsv1.DaemonInfo{
			Name:        c.cfg.Name,
			AcpRuntimes: runtimes,
			Labels:      c.cfg.Labels,
			Version:     daemonBuildVersion(),
			Os:          runtime.GOOS + "-" + runtime.GOARCH,
			Executors:   runtimes,
		},
	})
	c.authorize(registerReq)
	if _, err := client.Register(serveCtx, registerReq); err != nil {
		return fmt.Errorf("register: %w", err)
	}
	defer c.unregister(client)

	slog.Info("registered with server", "acp_runtimes", runtimes, "transport", "poll")

	for {
		pollReq := connect.NewRequest(&agentsv1.DaemonConnectorServicePollRequest{
			WaitTimeout: durationpb.New(daemonPollWaitTimeout),
		})
		c.authorize(pollReq)
		resp, err := client.Poll(serveCtx, pollReq)
		if err != nil {
			return fmt.Errorf("poll: %w", err)
		}
		for _, msg := range resp.Msg.GetMessages() {
			c.handleMessage(serveCtx, client, msg)
		}
	}
}

func (c *Connector) authorize(req interface{ Header() http.Header }) {
	if c.cfg.Credential != "" {
		req.Header().Set("Authorization", "Bearer "+c.cfg.Credential)
	}
}

func (c *Connector) unregister(client agentsv1connect.DaemonConnectorServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(&agentsv1.DaemonConnectorServiceUnregisterRequest{})
	c.authorize(req)
	if _, err := client.Unregister(ctx, req); err != nil {
		slog.Debug("failed to unregister daemon", "err", err)
	}
}

func (c *Connector) handleMessage(ctx context.Context, client agentsv1connect.DaemonConnectorServiceClient, msg *agentsv1.ConnectResponse) {
	switch m := msg.GetMessage().(type) {
	case *agentsv1.ConnectResponse_Task:
		go c.handleTask(ctx, client, m.Task)
	case *agentsv1.ConnectResponse_Cancel:
		c.handleCancel(m.Cancel.TaskId)
	}
}

func (c *Connector) handleTask(ctx context.Context, client agentsv1connect.DaemonConnectorServiceClient, task *agentsv1.DaemonTask) {
	slog.Info("task received", "task_id", task.TaskId, "acp_runtime", task.AcpRuntime, "work_dir", task.WorkDir, "input_len", len(task.Input))

	exec, ok := c.executors[task.AcpRuntime]
	if !ok {
		c.sendUpdate(ctx, client, &agentsv1.DaemonTaskUpdate{
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
		c.sendUpdate(ctx, client, update)
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

func (c *Connector) sendUpdate(ctx context.Context, client agentsv1connect.DaemonConnectorServiceClient, update *agentsv1.DaemonTaskUpdate) {
	req := connect.NewRequest(&agentsv1.DaemonConnectorServiceReportTaskUpdateRequest{Update: update})
	c.authorize(req)
	if _, err := client.ReportTaskUpdate(ctx, req); err != nil {
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
