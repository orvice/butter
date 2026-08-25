// Package butterbox stores workspace-registered ButterBoxes: agent VMs
// running the butter-box server whose PiService hosts pi coding-agent
// sessions (ADR-0011). The box access token never enters the public model:
// it is stored as ciphertext in dedicated credential columns beside the
// resource, and reads only derive whether one exists.
package butterbox

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNoCredential  = errors.New("no credential")
)

// Credential is an encrypted box access token as produced by the secretbox
// keyring. The repository stores it verbatim and never sees plaintext.
type Credential struct {
	Ciphertext string
	KeyID      string
}

// Set reports whether the credential holds a token.
func (c Credential) Set() bool { return c.Ciphertext != "" }

// Repository stores ButterBoxes scoped to one workspace per call. IDs are
// caller-assigned UUIDs; box names are unique within a workspace (enforced
// by index, not read-then-write). The derived fields `credential_set`,
// `credential_updated_at`, and `workspace_id` are stamped on every read and
// ignored on writes.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	List(ctx context.Context, workspaceID string) ([]*agentsv1.ButterBox, error)
	Get(ctx context.Context, workspaceID, id string) (*agentsv1.ButterBox, error)
	// Create stores the box and, when cred.Set(), its credential atomically.
	Create(ctx context.Context, workspaceID string, box *agentsv1.ButterBox, cred Credential) (*agentsv1.ButterBox, error)
	// Update replaces name, base URL, and enabled; the credential columns are
	// left untouched.
	Update(ctx context.Context, workspaceID string, box *agentsv1.ButterBox) (*agentsv1.ButterBox, error)
	Delete(ctx context.Context, workspaceID, id string) error

	// SetCredential sets or rotates the stored credential; an unset cred
	// clears it. Returns the box with derived fields refreshed.
	SetCredential(ctx context.Context, workspaceID, id string, cred Credential) (*agentsv1.ButterBox, error)
	// GetCredential returns the stored ciphertext, or ErrNoCredential when
	// none is set.
	GetCredential(ctx context.Context, workspaceID, id string) (Credential, error)
}
