package memory

import (
	"context"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestStoreWriteReadVersions(t *testing.T) {
	ctx := context.Background()
	store := New()
	space, err := store.CreateSpace(ctx, "ws-1", &agentsv1.AgentFileSpace{Name: "Notes"})
	if err != nil {
		t.Fatalf("CreateSpace: %v", err)
	}

	first, err := store.WriteFile(ctx, "ws-1", space.GetId(), "/notes.md", "one", "", nil)
	if err != nil {
		t.Fatalf("WriteFile first: %v", err)
	}
	if first.GetVersion() != 1 {
		t.Fatalf("first version = %d, want 1", first.GetVersion())
	}

	second, err := store.WriteFile(ctx, "ws-1", space.GetId(), "/notes.md", "two", "", nil)
	if err != nil {
		t.Fatalf("WriteFile second: %v", err)
	}
	if second.GetVersion() != 2 {
		t.Fatalf("second version = %d, want 2", second.GetVersion())
	}

	_, content, err := store.ReadFile(ctx, "ws-1", space.GetId(), "/notes.md", 1)
	if err != nil {
		t.Fatalf("ReadFile v1: %v", err)
	}
	if content != "one" {
		t.Fatalf("v1 content = %q, want one", content)
	}

	_, content, err = store.ReadFile(ctx, "ws-1", space.GetId(), "/notes.md", 0)
	if err != nil {
		t.Fatalf("ReadFile latest: %v", err)
	}
	if content != "two" {
		t.Fatalf("latest content = %q, want two", content)
	}
}
