package daemon

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	tokenmem "go.orx.me/apps/butter/internal/repo/apitoken/memory"
	configmem "go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	testWorkspaceID = "ws-1"
	testSecret      = "daemon-secret"
)

func startTestServer(t *testing.T, registry *Registry, seed bool) (agentsv1connect.DaemonConnectorServiceClient, func()) {
	t.Helper()

	tokenRepo := tokenmem.New()
	runtimeRepo := configmem.New()
	if seed {
		_, err := runtimeRepo.CreateDaemonRuntime(context.Background(), testWorkspaceID, &agentsv1.DaemonRuntime{
			Id:          "test-daemon",
			Name:        "Test",
			WorkspaceId: testWorkspaceID,
		})
		if err != nil {
			t.Fatalf("seed daemon runtime: %v", err)
		}
		err = tokenRepo.Create(context.Background(), &agentsv1.APIToken{
			Id:              "tok-1",
			Name:            "daemon",
			WorkspaceId:     testWorkspaceID,
			Kind:            agentsv1.APITokenKind_API_TOKEN_KIND_DAEMON,
			Scopes:          []string{"daemon:connect"},
			DaemonRuntimeId: "test-daemon",
			CreatedAt:       timestamppb.New(time.Now().UTC()),
		}, hashTestSecret(testSecret))
		if err != nil {
			t.Fatalf("seed daemon credential: %v", err)
		}
	}

	mux := http.NewServeMux()
	path, h := agentsv1connect.NewDaemonConnectorServiceHandler(NewGRPCHandler(registry, tokenRepo, runtimeRepo))
	mux.Handle(path, h)

	srv := httptest.NewServer(h2c.NewHandler(mux, &http2.Server{}))

	httpClient := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}

	client := agentsv1connect.NewDaemonConnectorServiceClient(httpClient, srv.URL)
	cleanup := func() {
		srv.Close()
	}
	return client, cleanup
}

func hashTestSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func TestGRPCHandlerConnectAndTask(t *testing.T) {
	registry := NewRegistry()
	client, cleanup := startTestServer(t, registry, true)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := client.Connect(ctx)
	stream.RequestHeader().Set("Authorization", "Bearer "+testSecret)

	// Register.
	err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{
				DaemonRuntimeId: "test-daemon",
				Name:            "Test",
				AcpRuntimes:     []string{"opencode"},
				WorkspaceId:     testWorkspaceID,
			},
		},
	})
	if err != nil {
		t.Fatalf("Send register: %v", err)
	}

	// Wait for registration to propagate.
	conn := waitForConn(registry, testWorkspaceID, "test-daemon", 2*time.Second)
	if conn == nil {
		t.Fatal("daemon not found in registry after register")
	}

	// Send a task via the registry connection.
	task := &agentsv1.DaemonTask{TaskId: "t1", AgentName: "coder", Input: "hello", AcpRuntime: "opencode", WorkspaceId: testWorkspaceID}
	resultCh, err := conn.SendTask(task)
	if err != nil {
		t.Fatalf("SendTask: %v", err)
	}

	// Daemon receives the task.
	msg, err := stream.Receive()
	if err != nil {
		t.Fatalf("Recv task: %v", err)
	}
	if msg.GetTask().GetTaskId() != "t1" {
		t.Fatalf("expected task t1, got %v", msg)
	}

	// Daemon sends back a completed update.
	err = stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_TaskUpdate{
			TaskUpdate: &agentsv1.DaemonTaskUpdate{
				TaskId: "t1",
				Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED,
				Output: "result text",
			},
		},
	})
	if err != nil {
		t.Fatalf("Send task update: %v", err)
	}

	// Server side receives the result.
	select {
	case update := <-resultCh:
		if update.Status != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED {
			t.Fatalf("expected COMPLETED, got %v", update.Status)
		}
		if update.Output != "result text" {
			t.Fatalf("expected 'result text', got %q", update.Output)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for task result")
	}
}

func TestGRPCHandlerAuthRejectsInvalidToken(t *testing.T) {
	registry := NewRegistry()
	client, cleanup := startTestServer(t, registry, true)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Connect without token.
	stream := client.Connect(ctx)
	err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{DaemonRuntimeId: "d1"},
		},
	})
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
		t.Fatalf("Send: %v", err)
	}

	// Should get an error on recv.
	_, err = stream.Receive()
	if err == nil {
		t.Fatal("expected auth error")
	}
	if code := connect.CodeOf(err); code != connect.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated error, got %v: %v", code, err)
	}
}

func TestGRPCHandlerAuthAcceptsValidToken(t *testing.T) {
	registry := NewRegistry()
	client, cleanup := startTestServer(t, registry, true)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Connect with valid token.
	stream := client.Connect(ctx)
	stream.RequestHeader().Set("Authorization", "Bearer "+testSecret)

	err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{DaemonRuntimeId: "test-daemon", AcpRuntimes: []string{"test"}},
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if conn := waitForConn(registry, testWorkspaceID, "test-daemon", 2*time.Second); conn == nil {
		t.Fatal("expected daemon to be registered with valid token")
	}
}

func TestGRPCHandlerSendsHeartbeat(t *testing.T) {
	oldInterval := daemonHeartbeatInterval
	daemonHeartbeatInterval = 20 * time.Millisecond
	defer func() {
		daemonHeartbeatInterval = oldInterval
	}()

	registry := NewRegistry()
	client, cleanup := startTestServer(t, registry, true)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream := client.Connect(ctx)
	stream.RequestHeader().Set("Authorization", "Bearer "+testSecret)

	if err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{DaemonRuntimeId: "test-daemon", AcpRuntimes: []string{"opencode"}},
		},
	}); err != nil {
		t.Fatalf("Send register: %v", err)
	}

	if waitForConn(registry, testWorkspaceID, "test-daemon", 2*time.Second) == nil {
		t.Fatal("daemon was not registered")
	}

	msg, err := stream.Receive()
	if err != nil {
		t.Fatalf("Receive heartbeat: %v", err)
	}
	if msg.Message != nil {
		t.Fatalf("expected empty heartbeat response, got %T", msg.Message)
	}

	if err := stream.CloseRequest(); err != nil {
		t.Fatalf("CloseRequest: %v", err)
	}
}

func TestGRPCHandlerUnregistersOnCleanClientClose(t *testing.T) {
	registry := NewRegistry()
	client, cleanup := startTestServer(t, registry, true)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stream := client.Connect(ctx)
	stream.RequestHeader().Set("Authorization", "Bearer "+testSecret)
	if err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{DaemonRuntimeId: "test-daemon", AcpRuntimes: []string{"opencode"}},
		},
	}); err != nil {
		t.Fatalf("Send register: %v", err)
	}

	if waitForConn(registry, testWorkspaceID, "test-daemon", 2*time.Second) == nil {
		t.Fatal("daemon was not registered")
	}

	if err := stream.CloseRequest(); err != nil {
		t.Fatalf("CloseRequest: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for registry.Get(testWorkspaceID, "test-daemon") != nil {
		select {
		case <-deadline:
			t.Fatal("daemon remained registered after clean client close")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func waitForConn(registry *Registry, workspaceID, runtimeID string, timeout time.Duration) *Connection {
	deadline := time.After(timeout)
	for {
		if conn := registry.Get(workspaceID, runtimeID); conn != nil {
			return conn
		}
		select {
		case <-deadline:
			return nil
		case <-time.After(10 * time.Millisecond):
		}
	}
}
