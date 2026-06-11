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

func TestConnectorRegistersAndPolls(t *testing.T) {
	registeredCh := make(chan struct{})
	pollCh := make(chan struct{})
	unregisterCh := make(chan struct{})
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	path, handler := agentsv1connect.NewDaemonConnectorServiceHandler(testDaemonConnectorHandler{
		register: func(_ context.Context, req *connect.Request[agentsv1.DaemonConnectorServiceRegisterRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceRegisterResponse], error) {
			if req.Msg.GetDaemon() == nil {
				err := connect.NewError(connect.CodeInvalidArgument, errors.New("daemon is required"))
				errCh <- err
				return nil, err
			}
			close(registeredCh)
			return connect.NewResponse(&agentsv1.DaemonConnectorServiceRegisterResponse{Daemon: req.Msg.GetDaemon()}), nil
		},
		poll: func(ctx context.Context, _ *connect.Request[agentsv1.DaemonConnectorServicePollRequest]) (*connect.Response[agentsv1.DaemonConnectorServicePollResponse], error) {
			close(pollCh)
			<-ctx.Done()
			return nil, ctx.Err()
		},
		unregister: func(context.Context, *connect.Request[agentsv1.DaemonConnectorServiceUnregisterRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceUnregisterResponse], error) {
			close(unregisterCh)
			return connect.NewResponse(&agentsv1.DaemonConnectorServiceUnregisterResponse{}), nil
		},
		reportTaskUpdate: func(context.Context, *connect.Request[agentsv1.DaemonConnectorServiceReportTaskUpdateRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceReportTaskUpdateResponse], error) {
			return connect.NewResponse(&agentsv1.DaemonConnectorServiceReportTaskUpdateResponse{}), nil
		},
		connect: func(context.Context, *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse]) error {
			err := connect.NewError(connect.CodeUnimplemented, errors.New("stream transport should not be used"))
			errCh <- err
			return err
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
	case <-registeredCh:
	case err := <-errCh:
		t.Fatalf("server register error: %v", err)
	case err := <-connectorErrCh:
		t.Fatalf("connector exited before register: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for register")
	}

	select {
	case <-pollCh:
	case err := <-errCh:
		t.Fatalf("server poll error: %v", err)
	case err := <-connectorErrCh:
		t.Fatalf("connector exited before poll: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for poll")
	}

	cancel()
	select {
	case <-unregisterCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for unregister")
	}
}

type testDaemonConnectorHandler struct {
	agentsv1connect.UnimplementedDaemonConnectorServiceHandler
	connect          func(context.Context, *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse]) error
	register         func(context.Context, *connect.Request[agentsv1.DaemonConnectorServiceRegisterRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceRegisterResponse], error)
	poll             func(context.Context, *connect.Request[agentsv1.DaemonConnectorServicePollRequest]) (*connect.Response[agentsv1.DaemonConnectorServicePollResponse], error)
	reportTaskUpdate func(context.Context, *connect.Request[agentsv1.DaemonConnectorServiceReportTaskUpdateRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceReportTaskUpdateResponse], error)
	unregister       func(context.Context, *connect.Request[agentsv1.DaemonConnectorServiceUnregisterRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceUnregisterResponse], error)
}

func (h testDaemonConnectorHandler) Connect(ctx context.Context, stream *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse]) error {
	return h.connect(ctx, stream)
}

func (h testDaemonConnectorHandler) Register(ctx context.Context, req *connect.Request[agentsv1.DaemonConnectorServiceRegisterRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceRegisterResponse], error) {
	return h.register(ctx, req)
}

func (h testDaemonConnectorHandler) Poll(ctx context.Context, req *connect.Request[agentsv1.DaemonConnectorServicePollRequest]) (*connect.Response[agentsv1.DaemonConnectorServicePollResponse], error) {
	return h.poll(ctx, req)
}

func (h testDaemonConnectorHandler) ReportTaskUpdate(ctx context.Context, req *connect.Request[agentsv1.DaemonConnectorServiceReportTaskUpdateRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceReportTaskUpdateResponse], error) {
	return h.reportTaskUpdate(ctx, req)
}

func (h testDaemonConnectorHandler) Unregister(ctx context.Context, req *connect.Request[agentsv1.DaemonConnectorServiceUnregisterRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceUnregisterResponse], error) {
	return h.unregister(ctx, req)
}

type testConnectorExecutor struct{}

func (testConnectorExecutor) Runtime() string { return "test" }

func (testConnectorExecutor) Execute(context.Context, *agentsv1.DaemonTask, func(*agentsv1.DaemonTaskUpdate)) error {
	return nil
}
