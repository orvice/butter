package memory

import (
	"context"
	"testing"

	"go.orx.me/apps/butter/internal/agentcontent"
	agentcontentrepo "go.orx.me/apps/butter/internal/repo/agentcontent"
)

func TestRoundTrip(t *testing.T) {
	s := New()
	ctx := context.Background()

	snap := agentcontent.Snapshot{
		CommitSHA: "abc123",
		Entries: map[string]agentcontent.AgentContent{
			"agent-1": {AgentID: "agent-1", Instruction: "do things"},
		},
	}
	if err := s.PutSnapshot(ctx, "ws1", snap); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetSnapshot(ctx, "ws1", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if got.CommitSHA != "abc123" {
		t.Errorf("commit = %q", got.CommitSHA)
	}
	if got.Entries["agent-1"].Instruction != "do things" {
		t.Errorf("instruction = %q", got.Entries["agent-1"].Instruction)
	}
}

func TestGetNotFound(t *testing.T) {
	s := New()
	_, err := s.GetSnapshot(context.Background(), "missing", "abc123")
	if err != agentcontentrepo.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := New()
	ctx := context.Background()

	snap := agentcontent.Snapshot{CommitSHA: "x", Entries: map[string]agentcontent.AgentContent{}}
	_ = s.PutSnapshot(ctx, "ws1", snap)
	_ = s.Delete(ctx, "ws1")

	_, err := s.GetSnapshot(ctx, "ws1", "x")
	if err != agentcontentrepo.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMutationIsolation(t *testing.T) {
	s := New()
	ctx := context.Background()

	entries := map[string]agentcontent.AgentContent{
		"a": {Instruction: "original"},
	}
	_ = s.PutSnapshot(ctx, "ws1", agentcontent.Snapshot{CommitSHA: "1", Entries: entries})

	entries["a"] = agentcontent.AgentContent{Instruction: "mutated"}

	got, _ := s.GetSnapshot(ctx, "ws1", "1")
	if got.Entries["a"].Instruction != "original" {
		t.Errorf("store should not be affected by external mutation")
	}
}

func TestSnapshotsAreRevisionAddressedAndPruned(t *testing.T) {
	s := New()
	ctx := context.Background()

	_ = s.PutSnapshot(ctx, "ws1", agentcontent.Snapshot{
		CommitSHA: "v1",
		Entries:   map[string]agentcontent.AgentContent{"a": {Instruction: "old"}},
	})
	_ = s.PutSnapshot(ctx, "ws1", agentcontent.Snapshot{
		CommitSHA: "v2",
		Entries:   map[string]agentcontent.AgentContent{"a": {Instruction: "new"}},
	})

	old, err := s.GetSnapshot(ctx, "ws1", "v1")
	if err != nil || old.Entries["a"].Instruction != "old" {
		t.Fatalf("old snapshot = %#v, %v", old, err)
	}
	got, err := s.GetSnapshot(ctx, "ws1", "v2")
	if err != nil || got.Entries["a"].Instruction != "new" {
		t.Fatalf("new snapshot = %#v, %v", got, err)
	}
	if err := s.PruneSnapshots(ctx, "ws1", "v2"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetSnapshot(ctx, "ws1", "v1"); err != agentcontentrepo.ErrNotFound {
		t.Fatalf("old snapshot after prune: %v", err)
	}
}
