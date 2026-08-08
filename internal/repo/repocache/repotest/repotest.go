package repotest

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"go.orx.me/apps/butter/internal/repo/repocache"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Run exercises behavior shared by every repository cache implementation.
func Run(t *testing.T, newStore func(t *testing.T) repocache.Repository) {
	t.Helper()
	store := newStore(t)
	ctx := context.Background()
	if err := store.EnsureIndexes(ctx); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}

	metadata := repocache.SnapshotMetadata{BindingKey: "binding-a", CommitSHA: "sha-1"}
	entries := []*agentsv1.RepoCacheEntry{
		{Path: "agents", Kind: agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_DIRECTORY},
		{Path: "agents/known", Kind: agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_DIRECTORY, Claimed: true},
		{Path: "agents/known/prompt.md", Kind: agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_FILE, Size: 5, ContentHash: "hash-1", Claimed: true},
	}
	if err := store.PutSnapshot(ctx, "ws-a", metadata, entries, []repocache.CachedBlob{{
		Path: "agents/known/prompt.md", Content: []byte("hello"),
	}}); err != nil {
		t.Fatalf("PutSnapshot: %v", err)
	}

	gotMetadata, err := store.GetMetadata(ctx, "ws-a")
	if err != nil || gotMetadata.BindingKey != metadata.BindingKey || gotMetadata.CommitSHA != metadata.CommitSHA || gotMetadata.SnapshotID == "" {
		t.Fatalf("GetMetadata = %#v, %v", gotMetadata, err)
	}
	root, err := store.ListEntries(ctx, "ws-a", gotMetadata.SnapshotID, "")
	if err != nil || len(root) != 1 || root[0].GetPath() != "agents" {
		t.Fatalf("ListEntries root = %v, %v", root, err)
	}
	agents, err := store.ListEntries(ctx, "ws-a", gotMetadata.SnapshotID, "agents")
	if err != nil || len(agents) != 1 || !agents[0].GetClaimed() {
		t.Fatalf("ListEntries agents = %v, %v", agents, err)
	}
	entry, err := store.GetEntry(ctx, "ws-a", gotMetadata.SnapshotID, "agents/known/prompt.md")
	if err != nil || entry.GetContentHash() != "hash-1" || !entry.GetClaimed() {
		t.Fatalf("GetEntry = %v, %v", entry, err)
	}
	blob, err := store.GetBlob(ctx, "ws-a", gotMetadata.SnapshotID, "agents/known/prompt.md")
	if err != nil || !bytes.Equal(blob, []byte("hello")) {
		t.Fatalf("GetBlob = %q, %v", blob, err)
	}

	replacement := repocache.SnapshotMetadata{BindingKey: "binding-b", CommitSHA: "sha-2"}
	if err := store.PutSnapshot(ctx, "ws-a", replacement, []*agentsv1.RepoCacheEntry{
		{Path: "README.md", Kind: agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_FILE},
	}, []repocache.CachedBlob{{Path: "README.md", Content: nil}}); err != nil {
		t.Fatalf("replace PutSnapshot: %v", err)
	}
	gotMetadata, err = store.GetMetadata(ctx, "ws-a")
	if err != nil || gotMetadata.BindingKey != replacement.BindingKey || gotMetadata.CommitSHA != replacement.CommitSHA || gotMetadata.SnapshotID == "" {
		t.Fatalf("replacement metadata = %#v, %v", gotMetadata, err)
	}
	if _, err := store.GetEntry(ctx, "ws-a", gotMetadata.SnapshotID, "agents/known/prompt.md"); !errors.Is(err, repocache.ErrNotFound) {
		t.Fatalf("old entry survived replacement: %v", err)
	}
	empty, err := store.GetBlob(ctx, "ws-a", gotMetadata.SnapshotID, "README.md")
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty blob = %q, %v", empty, err)
	}
	if _, err := store.GetMetadata(ctx, "ws-b"); !errors.Is(err, repocache.ErrNotFound) {
		t.Fatalf("workspace isolation: %v", err)
	}

	if err := store.Delete(ctx, "ws-a"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.GetMetadata(ctx, "ws-a"); !errors.Is(err, repocache.ErrNotFound) {
		t.Fatalf("metadata survived delete: %v", err)
	}
}
