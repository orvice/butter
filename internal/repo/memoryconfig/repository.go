// Package memoryconfig stores each workspace's WorkspaceMemoryConfig: the
// connection to the mem0 OSS server holding its Workspace Memory and Agent
// Memory (ADR-0013). A workspace has zero or one config. The mem0 API key
// never enters the public model: it is stored as ciphertext in dedicated
// credential columns beside the config, and reads only derive whether one
// exists.
package memoryconfig

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrNoCredential = errors.New("no credential")
)

// Credential is an encrypted mem0 API key as produced by the secretbox
// keyring. The repository stores it verbatim and never sees plaintext.
type Credential struct {
	Ciphertext string
	KeyID      string
}

// Set reports whether the credential holds a key.
func (c Credential) Set() bool { return c.Ciphertext != "" }

// Repository stores one WorkspaceMemoryConfig per workspace. The derived
// fields `workspace_id`, `credential_set`, and `credential_updated_at` are
// stamped on every read and ignored on writes.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Get returns the workspace's config, or ErrNotFound.
	Get(ctx context.Context, workspaceID string) (*agentsv1.WorkspaceMemoryConfig, error)
	// Put creates or replaces the base URL and enabled flag. A nil cred
	// keeps the stored credential; a non-nil unset cred clears it; a set
	// cred sets or rotates it. Config and credential change atomically.
	Put(ctx context.Context, workspaceID string, cfg *agentsv1.WorkspaceMemoryConfig, cred *Credential) (*agentsv1.WorkspaceMemoryConfig, error)
	// Delete removes the config and its credential, or returns ErrNotFound.
	Delete(ctx context.Context, workspaceID string) error
	// GetCredential returns the stored ciphertext: ErrNotFound without a
	// config, ErrNoCredential when the config has no key.
	GetCredential(ctx context.Context, workspaceID string) (Credential, error)
}
