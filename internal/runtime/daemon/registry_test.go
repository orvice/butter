package daemon

import (
	"fmt"
	"sync"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRegistryRegisterAndFind(t *testing.T) {
	r := NewRegistry()
	conn := NewConnection(&agentsv1.DaemonInfo{
		WorkspaceId:  "ws-1",
		DaemonId:     "d1",
		Name:         "daemon-1",
		Capabilities: []string{"opencode", "shell"},
	})
	r.Register(conn)

	got := r.FindByCapability("ws-1", "opencode")
	if got != conn {
		t.Fatalf("expected conn, got %v", got)
	}

	got = r.FindByCapability("ws-1", "shell")
	if got != conn {
		t.Fatalf("expected conn for shell, got %v", got)
	}

	got = r.FindByCapability("ws-1", "unknown")
	if got != nil {
		t.Fatalf("expected nil for unknown capability, got %v", got)
	}

	got = r.FindByCapability("ws-2", "opencode")
	if got != nil {
		t.Fatalf("expected nil for other workspace, got %v", got)
	}
}

func TestRegistryUnregister(t *testing.T) {
	r := NewRegistry()
	conn := NewConnection(&agentsv1.DaemonInfo{
		WorkspaceId:  "ws-1",
		DaemonId:     "d1",
		Capabilities: []string{"opencode"},
	})
	r.Register(conn)
	r.Unregister("ws-1", "d1")

	if got := r.FindByCapability("ws-1", "opencode"); got != nil {
		t.Fatalf("expected nil after unregister, got %v", got)
	}
}

func TestRegistryListConnected(t *testing.T) {
	r := NewRegistry()
	r.Register(NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonId: "d1", Capabilities: []string{"a"}}))
	r.Register(NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-2", DaemonId: "d2", Capabilities: []string{"a"}}))

	list := r.ListConnected("")
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}

	list = r.ListConnected("ws-1")
	if len(list) != 1 || list[0].GetDaemonId() != "d1" {
		t.Fatalf("expected only ws-1 daemon, got %#v", list)
	}

	if got := r.FindByCapability("ws-2", "a"); got == nil || got.Info.GetDaemonId() != "d2" {
		t.Fatalf("expected ws-2 capability lookup to find d2, got %v", got)
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
			conn := NewConnection(&agentsv1.DaemonInfo{WorkspaceId: "ws-1", DaemonId: id, Capabilities: []string{"cap"}})
			r.Register(conn)
			r.FindByCapability("ws-1", "cap")
			r.ListConnected("ws-1")
			r.Unregister("ws-1", id)
		}(i)
	}

	wg.Wait()
}
