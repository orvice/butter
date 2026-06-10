package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"go.orx.me/apps/butter/cmd/butter-daemon/executor"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

func TestConnectorSendsHeartbeat(t *testing.T) {
	oldInterval := daemonClientHeartbeatInterval
	daemonClientHeartbeatInterval = 20 * time.Millisecond
	defer func() {
		daemonClientHeartbeatInterval = oldInterval
	}()

	heartbeatCh := make(chan struct{})
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	path, handler := agentsv1connect.NewDaemonConnectorServiceHandler(testDaemonConnectorHandler{
		connect: func(_ context.Context, stream *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse]) error {
			msg, err := stream.Receive()
			if err != nil {
				errCh <- err
				return err
			}
			if msg.GetRegister() == nil {
				errCh <- errExpectedRegister
				return errExpectedRegister
			}
			for {
				msg, err := stream.Receive()
				if err != nil {
					errCh <- err
					return err
				}
				if msg.GetHeartbeat() != nil {
					close(heartbeatCh)
					return nil
				}
			}
		},
	})
	mux.Handle(path, handler)
	server := httptest.NewServer(h2c.NewHandler(mux, &http2.Server{}))
	defer server.Close()

	connector := NewConnector(&Config{Server: server.URL}, []executor.Executor{testConnectorExecutor{}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connectorErrCh := make(chan error, 1)
	go func() {
		connectorErrCh <- connector.connectAndServe(ctx)
	}()

	select {
	case <-heartbeatCh:
	case err := <-errCh:
		t.Fatalf("server receive error: %v", err)
	case err := <-connectorErrCh:
		t.Fatalf("connector exited before heartbeat: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for heartbeat")
	}
}

var errExpectedRegister = errors.New("expected register")

type testDaemonConnectorHandler struct {
	agentsv1connect.UnimplementedDaemonConnectorServiceHandler
	connect func(context.Context, *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse]) error
}

func (h testDaemonConnectorHandler) Connect(ctx context.Context, stream *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse]) error {
	return h.connect(ctx, stream)
}

type testConnectorExecutor struct{}

func (testConnectorExecutor) Runtime() string { return "test" }

func (testConnectorExecutor) Execute(context.Context, *agentsv1.DaemonTask, func(*agentsv1.DaemonTaskUpdate)) error {
	return nil
}
