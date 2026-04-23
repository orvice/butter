package daemon

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func startTestServer(t *testing.T, registry *Registry, apiToken string) (agentsv1.DaemonConnectorServiceClient, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	agentsv1.RegisterDaemonConnectorServiceServer(srv, NewGRPCHandler(registry, apiToken))
	go srv.Serve(lis)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		srv.Stop()
		t.Fatalf("dial: %v", err)
	}

	client := agentsv1.NewDaemonConnectorServiceClient(conn)
	cleanup := func() {
		conn.Close()
		srv.Stop()
	}
	return client, cleanup
}

func TestGRPCHandlerConnectAndTask(t *testing.T) {
	registry := NewRegistry()
	client, cleanup := startTestServer(t, registry, "")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Register.
	err = stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{
				DaemonId:     "test-daemon",
				Name:         "Test",
				Capabilities: []string{"opencode"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Send register: %v", err)
	}

	// Wait for registration to propagate.
	time.Sleep(100 * time.Millisecond)

	conn := registry.FindByCapability("opencode")
	if conn == nil {
		t.Fatal("daemon not found in registry after register")
	}

	// Send a task via the registry connection.
	task := &agentsv1.DaemonTask{TaskId: "t1", AgentName: "coder", Input: "hello", Capability: "opencode"}
	resultCh, err := conn.SendTask(task)
	if err != nil {
		t.Fatalf("SendTask: %v", err)
	}

	// Daemon receives the task.
	msg, err := stream.Recv()
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
	client, cleanup := startTestServer(t, registry, "secret-token")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Connect without token.
	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	err = stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{DaemonId: "d1"},
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	// Should get an error on recv.
	_, err = stream.Recv()
	if err == nil {
		t.Fatal("expected auth error")
	}
}

func TestGRPCHandlerAuthAcceptsValidToken(t *testing.T) {
	registry := NewRegistry()
	client, cleanup := startTestServer(t, registry, "secret-token")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Connect with valid token.
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer secret-token")
	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	err = stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{DaemonId: "d1", Capabilities: []string{"test"}},
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if conn := registry.FindByCapability("test"); conn == nil {
		t.Fatal("expected daemon to be registered with valid token")
	}
}
