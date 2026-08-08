// Package repocache stores the workspace-scoped repository cache produced
// by synchronization (issue #215). Each workspace points to one immutable
// tree snapshot containing tree entries and UTF-8 Markdown blobs. Reads are
// workspace-isolated: one workspace never sees another's cache, even when
// both bind the same repository location.
package repocache

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrNotFound means the workspace has no cached tree or the requested
	// path does not exist in the cache.
	ErrNotFound = errors.New("not found")
)

// CachedBlob holds the raw content of a cached Markdown file.
type CachedBlob struct {
	Path    string
	Content []byte
}

// SnapshotMetadata identifies the repository binding and revision used to
// build a cached snapshot. BindingKey prevents content from a previous
// binding being served after a workspace is rebound.
type SnapshotMetadata struct {
	// SnapshotID is an opaque storage version returned by GetMetadata. Callers
	// pass it back to reads so one response cannot mix concurrent snapshots.
	SnapshotID string
	BindingKey string
	CommitSHA  string
}

// Repository stores workspace-scoped repository cache entries.
type Repository interface {
	// EnsureIndexes creates storage indexes required by the implementation.
	EnsureIndexes(ctx context.Context) error

	// PutSnapshot replaces the workspace's current cached tree. Implementations
	// must publish entries, blobs, and metadata as one logical snapshot so
	// readers cannot combine data from concurrent synchronizations.
	PutSnapshot(ctx context.Context, workspaceID string, metadata SnapshotMetadata, entries []*agentsv1.RepoCacheEntry, blobs []CachedBlob) error

	// GetMetadata returns the identity and revision of the current snapshot,
	// or ErrNotFound if none exists.
	GetMetadata(ctx context.Context, workspaceID string) (SnapshotMetadata, error)

	// ListEntries returns cached tree entries whose parent directory matches
	// the given path (empty path lists root-level entries). Returns
	// ErrNotFound when the workspace has no cache.
	ListEntries(ctx context.Context, workspaceID, snapshotID, dirPath string) ([]*agentsv1.RepoCacheEntry, error)

	// GetEntry returns a single cached entry by exact path, or ErrNotFound.
	GetEntry(ctx context.Context, workspaceID, snapshotID, path string) (*agentsv1.RepoCacheEntry, error)

	// GetBlob returns the cached file content at path, or ErrNotFound.
	GetBlob(ctx context.Context, workspaceID, snapshotID, path string) ([]byte, error)

	// Delete removes all cached data for the workspace.
	Delete(ctx context.Context, workspaceID string) error
}
