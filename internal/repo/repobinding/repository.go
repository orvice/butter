// Package repobinding stores each workspace's repository binding (issue
// #214): the zero-or-one association between a workspace and a repository
// location on an admin-configured Git host.
//
// The binding's PAT is handled through a separate credential seam and is
// deliberately not a field of the WorkspaceRepoBinding proto: callers pass
// pre-encrypted ciphertext in and get ciphertext out, so implementations
// never see plaintext and the credential can never ride along on a binding
// read into an API response or log line. Implementations derive the proto's
// server-owned credential_set / credential_updated_at fields from the stored
// credential state on every read.
package repobinding

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrNotFound means the workspace has no binding.
	ErrNotFound = errors.New("not found")
	// ErrNoCredential means the binding exists but no PAT has been stored.
	ErrNoCredential = errors.New("no credential")
)

// Repository stores workspace repository bindings, keyed by workspace: at
// most one binding per workspace, enforced by construction (Put upserts).
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Get returns the workspace's binding or ErrNotFound.
	Get(ctx context.Context, workspaceID string) (*agentsv1.WorkspaceRepoBinding, error)
	// Put creates or replaces the workspace's binding. It stamps
	// workspace_id and timestamps (preserving created_at on upsert), keeps
	// any stored credential, and derives the credential fields on the
	// returned proto. All other fields are stored as given.
	Put(ctx context.Context, workspaceID string, binding *agentsv1.WorkspaceRepoBinding) (*agentsv1.WorkspaceRepoBinding, error)
	// Delete removes the binding and its stored credential.
	Delete(ctx context.Context, workspaceID string) error

	// SetCredential stores the encrypted PAT for an existing binding
	// (ErrNotFound when the workspace has no binding). An empty ciphertext
	// clears the stored credential and its timestamp.
	SetCredential(ctx context.Context, workspaceID, ciphertext string) error
	// GetCredential returns the stored ciphertext, ErrNotFound when the
	// workspace has no binding, or ErrNoCredential when none was set.
	GetCredential(ctx context.Context, workspaceID string) (string, error)

	// SetWebhookSecret stores the encrypted webhook secret for an existing
	// binding (ErrNotFound when the workspace has no binding). An empty
	// ciphertext clears the stored secret.
	SetWebhookSecret(ctx context.Context, workspaceID, ciphertext string) error
	// GetWebhookSecret returns the stored webhook secret ciphertext,
	// ErrNotFound when the workspace has no binding, or ErrNoCredential
	// when none was set.
	GetWebhookSecret(ctx context.Context, workspaceID string) (string, error)

	// ListAcrossWorkspaces returns every workspace's binding — the flat view
	// used to surface intentional repository overlap between workspaces.
	ListAcrossWorkspaces(ctx context.Context) ([]*agentsv1.WorkspaceRepoBinding, error)
}
