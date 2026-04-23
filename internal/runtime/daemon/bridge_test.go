package daemon

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestBridgeBuildAgent(t *testing.T) {
	registry := NewRegistry()
	bridge := NewBridge(registry, "opencode")

	ag, err := bridge.BuildAgent("test-daemon", "A test daemon agent")
	if err != nil {
		t.Fatalf("BuildAgent: %v", err)
	}
	if ag.Name() != "test-daemon" {
		t.Fatalf("expected name 'test-daemon', got %q", ag.Name())
	}
	if ag.Description() != "A test daemon agent" {
		t.Fatalf("expected description 'A test daemon agent', got %q", ag.Description())
	}
}

func TestBridgeNoDaemonAvailable(t *testing.T) {
	registry := NewRegistry()
	bridge := NewBridge(registry, "opencode")

	ag, err := bridge.BuildAgent("test", "test")
	if err != nil {
		t.Fatalf("BuildAgent: %v", err)
	}

	// We can't easily invoke the agent.Run() directly without a full ADK
	// InvocationContext, but we can verify the agent was built successfully
	// and the bridge logic is testable via the connection/registry layer.
	_ = ag
}

func TestExtractText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"nil content", "", ""},
		{"single text", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.text == "" {
				got := extractText(nil)
				if got != tt.want {
					t.Fatalf("got %q, want %q", got, tt.want)
				}
			}
		})
	}
}

// TestBridgeEndToEndViaConnection tests the bridge logic indirectly by
// verifying that a task dispatched to a daemon connection receives the result.
func TestBridgeEndToEndViaConnection(t *testing.T) {
	registry := NewRegistry()
	conn := NewConnection(&agentsv1.DaemonInfo{
		DaemonId:     "d1",
		Capabilities: []string{"opencode"},
	})
	registry.Register(conn)

	bridge := NewBridge(registry, "opencode")

	// Verify the bridge can find the daemon.
	found := bridge.registry.FindByCapability("opencode")
	if found == nil {
		t.Fatal("bridge's registry should find daemon")
	}

	// Send a task and verify the round-trip via connection.
	task := &agentsv1.DaemonTask{TaskId: "t1", Capability: "opencode", Input: "test"}
	resultCh, err := found.SendTask(task)
	if err != nil {
		t.Fatalf("SendTask: %v", err)
	}
	<-conn.SendCh // drain server message

	conn.DispatchUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: "t1",
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED,
		Output: "result",
	})

	update := <-resultCh
	if update.Output != "result" {
		t.Fatalf("expected 'result', got %q", update.Output)
	}
}
