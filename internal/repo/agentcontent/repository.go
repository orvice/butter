// Package agentcontent stores validated Agent Content snapshots produced by
// the publication pipeline (issue #216). Snapshots are addressed by repository
// revision; the binding's active_commit_sha selects the effective one.
package agentcontent

import (
	"context"
	"errors"

	"go.orx.me/apps/butter/internal/agentcontent"
)

// ErrNotFound means the requested revision has no complete content snapshot.
var ErrNotFound = errors.New("not found")

// Repository stores workspace-scoped Agent Content snapshots.
type Repository interface {
	// EnsureIndexes creates storage indexes required by the implementation.
	EnsureIndexes(ctx context.Context) error

	// PutSnapshot stores the validated content for the given workspace,
	// keyed by its immutable commit SHA.
	PutSnapshot(ctx context.Context, workspaceID string, snapshot agentcontent.Snapshot) error

	// GetSnapshot returns the snapshot for an exact repository revision, or
	// ErrNotFound if that revision is not stored completely.
	GetSnapshot(ctx context.Context, workspaceID, commitSHA string) (agentcontent.Snapshot, error)

	// PruneSnapshots removes every workspace snapshot except keepCommitSHA.
	// Publication calls this only after the binding's Active Revision advances.
	PruneSnapshots(ctx context.Context, workspaceID, keepCommitSHA string) error

	// Delete removes every snapshot for the workspace.
	Delete(ctx context.Context, workspaceID string) error
}
