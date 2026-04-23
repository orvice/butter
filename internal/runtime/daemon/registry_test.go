package daemon

import (
	"sync"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRegistryRegisterAndFind(t *testing.T) {
	r := NewRegistry()
	conn := NewConnection(&agentsv1.DaemonInfo{
		DaemonId:     "d1",
		Name:         "daemon-1",
		Capabilities: []string{"opencode", "shell"},
	})
	r.Register(conn)

	got := r.FindByCapability("opencode")
	if got != conn {
		t.Fatalf("expected conn, got %v", got)
	}

	got = r.FindByCapability("shell")
	if got != conn {
		t.Fatalf("expected conn for shell, got %v", got)
	}

	got = r.FindByCapability("unknown")
	if got != nil {
		t.Fatalf("expected nil for unknown capability, got %v", got)
	}
}

func TestRegistryUnregister(t *testing.T) {
	r := NewRegistry()
	conn := NewConnection(&agentsv1.DaemonInfo{
		DaemonId:     "d1",
		Capabilities: []string{"opencode"},
	})
	r.Register(conn)
	r.Unregister("d1")

	if got := r.FindByCapability("opencode"); got != nil {
		t.Fatalf("expected nil after unregister, got %v", got)
	}
}

func TestRegistryListConnected(t *testing.T) {
	r := NewRegistry()
	r.Register(NewConnection(&agentsv1.DaemonInfo{DaemonId: "d1", Capabilities: []string{"a"}}))
	r.Register(NewConnection(&agentsv1.DaemonInfo{DaemonId: "d2", Capabilities: []string{"b"}}))

	list := r.ListConnected()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			conn := NewConnection(&agentsv1.DaemonInfo{DaemonId: id, Capabilities: []string{"cap"}})
			r.Register(conn)
			r.FindByCapability("cap")
			r.ListConnected()
			r.Unregister(id)
		}(string(rune('a' + i%26)))
	}

	wg.Wait()
}
