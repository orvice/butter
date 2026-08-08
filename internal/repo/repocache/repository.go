// Package repocache stores the workspace-scoped repository cache produced
// by synchronization (issue #215). Each workspace has at most one cached
// tree snapshot (keyed by commit SHA) containing tree entries and UTF-8
// Markdown blobs. Reads are workspace-isolated: one workspace never sees
// another's cache, even when both bind the same repository location.
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

// Repository stores workspace-scoped repository cache entries.
type Repository interface {
	// PutSnapshot atomically replaces the workspace's cached tree entries
	// and blobs for the given commit SHA. Existing entries for the workspace
	// are removed before the new ones are written.
	PutSnapshot(ctx context.Context, workspaceID, commitSHA string, entries []*agentsv1.RepoCacheEntry, blobs []CachedBlob) error

	// GetCommitSHA returns the commit SHA of the currently cached snapshot
	// for the workspace, or ErrNotFound if none exists.
	GetCommitSHA(ctx context.Context, workspaceID string) (string, error)

	// ListEntries returns cached tree entries whose parent directory matches
	// the given path (empty path lists root-level entries). Returns
	// ErrNotFound when the workspace has no cache.
	ListEntries(ctx context.Context, workspaceID, dirPath string) ([]*agentsv1.RepoCacheEntry, error)

	// GetEntry returns a single cached entry by exact path, or ErrNotFound.
	GetEntry(ctx context.Context, workspaceID, path string) (*agentsv1.RepoCacheEntry, error)

	// GetBlob returns the cached file content at path, or ErrNotFound.
	GetBlob(ctx context.Context, workspaceID, path string) ([]byte, error)

	// Delete removes all cached data for the workspace.
	Delete(ctx context.Context, workspaceID string) error
}
