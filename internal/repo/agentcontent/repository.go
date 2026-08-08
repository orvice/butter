// Package agentcontent stores validated Agent Content snapshots produced by
// the publication pipeline (issue #216). Each workspace has at most one
// active snapshot representing the last revision that passed full validation.
package agentcontent

import (
	"context"
	"errors"

	"go.orx.me/apps/butter/internal/agentcontent"
)

// ErrNotFound means the workspace has no active content snapshot.
var ErrNotFound = errors.New("not found")

// Repository stores workspace-scoped Agent Content snapshots.
type Repository interface {
	// PutSnapshot stores the validated content for the given workspace,
	// replacing any prior snapshot.
	PutSnapshot(ctx context.Context, workspaceID string, snapshot agentcontent.Snapshot) error

	// GetSnapshot returns the active snapshot for the workspace, or
	// ErrNotFound if none exists.
	GetSnapshot(ctx context.Context, workspaceID string) (agentcontent.Snapshot, error)

	// Delete removes the workspace's snapshot.
	Delete(ctx context.Context, workspaceID string) error
}
