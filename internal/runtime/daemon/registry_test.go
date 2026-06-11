package daemon

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	conn := NewConnection(&agentsv1.DaemonInfo{
		WorkspaceId:     "ws-1",
		DaemonRuntimeId: "d1",
		Name:            "daemon-1",
		AcpRuntimes:     []string{"opencode", "codex"},
	})
	if err := r.Register(conn); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got := r.Get("ws-1", "d1")
	if got != conn {
		t.Fatalf("expected conn, got %v", got)
	}

	if got := r.Get("ws-2", "d1"); got != nil {
		t.Fatalf("expected nil for other workspace, got %v", got)
	}
}

func TestRegistryRejectsDuplicateRuntime(t *testing.T) {
	r := NewRegistry()
	info := &agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonRuntimeId: "d1"}
	if err := r.Register(NewConnection(info)); err != nil {
		t.Fatalf("Register first: %v", err)
	}
	err := r.Register(NewConnection(info))
	if !errors.Is(err, ErrRuntimeAlreadyConnected) {
		t.Fatalf("expected ErrRuntimeAlreadyConnected, got %v", err)
	}
}

func TestRegistryUnregister(t *testing.T) {
	r := NewRegistry()
	conn := NewConnection(&agentsv1.DaemonInfo{
		WorkspaceId:     "ws-1",
		DaemonRuntimeId: "d1",
		AcpRuntimes:     []string{"opencode"},
	})
	if err := r.Register(conn); err != nil {
		t.Fatalf("Register: %v", err)
	}
	r.Unregister("ws-1", "d1")

	if got := r.Get("ws-1", "d1"); got != nil {
		t.Fatalf("expected nil after unregister, got %v", got)
	}
}

func TestRegistryRegisterOrReplaceClosesPreviousConnection(t *testing.T) {
	r := NewRegistry()
	first := NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonRuntimeId: "d1", AcpRuntimes: []string{"opencode"}})
	replaced := r.RegisterOrReplace(first)
	if replaced {
		t.Fatal("first register should not replace an existing connection")
	}

	resultCh, err := first.SendTask(&agentsv1.DaemonTask{TaskId: "t1", WorkspaceId: "ws-1", AcpRuntime: "opencode"})
	if err != nil {
		t.Fatalf("SendTask: %v", err)
	}

	second := NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonRuntimeId: "d1", AcpRuntimes: []string{"opencode"}})
	replaced = r.RegisterOrReplace(second)
	if !replaced {
		t.Fatal("second register should replace existing connection")
	}
	if got := r.Get("ws-1", "d1"); got != second {
		t.Fatalf("expected replacement connection, got %v", got)
	}

	select {
	case update := <-resultCh:
		if update.GetStatus() != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED {
			t.Fatalf("expected FAILED update from closed connection, got %v", update.GetStatus())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for disconnect update")
	}
}

func TestRegistryPruneStalePollConnections(t *testing.T) {
	r := NewRegistry()

	pollConn := NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonRuntimeId: "poll", AcpRuntimes: []string{"opencode"}})
	pollConn.MarkPollMode()
	pollConn.mu.Lock()
	pollConn.LastSeen = time.Now().Add(-2 * time.Minute)
	pollConn.mu.Unlock()
	r.RegisterOrReplace(pollConn)

	streamConn := NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonRuntimeId: "stream", AcpRuntimes: []string{"opencode"}})
	streamConn.mu.Lock()
	streamConn.LastSeen = time.Now().Add(-2 * time.Minute)
	streamConn.mu.Unlock()
	if err := r.Register(streamConn); err != nil {
		t.Fatalf("Register stream conn: %v", err)
	}

	r.PruneStalePollConnections(time.Minute)

	if got := r.Get("ws-1", "poll"); got != nil {
		t.Fatalf("expected stale poll connection to be pruned, got %v", got)
	}
	if got := r.Get("ws-1", "stream"); got != streamConn {
		t.Fatalf("expected stream connection to remain, got %v", got)
	}
	if _, err := pollConn.SendTask(&agentsv1.DaemonTask{TaskId: "after-prune"}); !errors.Is(err, ErrDaemonDisconnected) {
		t.Fatalf("expected pruned connection to be closed, got %v", err)
	}
}

func TestRegistryListConnected(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonRuntimeId: "d1", AcpRuntimes: []string{"a"}}))
	_ = r.Register(NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-2", DaemonRuntimeId: "d2", AcpRuntimes: []string{"a"}}))

	list := r.ListConnected("")
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}

	list = r.ListConnected("ws-1")
	if len(list) != 1 || list[0].GetDaemonRuntimeId() != "d1" {
		t.Fatalf("expected only ws-1 runtime, got %#v", list)
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("d-%d", i)
			conn := NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonRuntimeId: id, AcpRuntimes: []string{"opencode"}})
			_ = r.Register(conn)
			_ = r.Get("ws-1", id)
			r.ListConnected("ws-1")
			r.Unregister("ws-1", id)
		}(i)
	}

	wg.Wait()
}
