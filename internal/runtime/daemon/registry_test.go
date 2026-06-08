package daemon

import (
	"errors"
	"fmt"
	"sync"
	"testing"

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
