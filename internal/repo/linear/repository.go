// Package linear stores Linear Apps (ADR-0015).
//
// The App's client secret and webhook signing secret are handled through a
// credential seam, exactly as Telegram Bot Tokens and ButterBox tokens are
// (ADR-0005, ADR-0008): callers pass pre-encrypted ciphertext in and get
// ciphertext out, so implementations never see plaintext and a secret can
// never ride along on an App read into an API response or log line.
// Implementations derive the App proto's server-owned workspace_id,
// credential_state and *_secret_set fields from storage on every read.
package linear

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrNotFound means no such App exists in the workspace.
	ErrNotFound = errors.New("not found")
	// ErrClientIDExists means another App — in any workspace — already
	// registered that Linear OAuth client ID.
	ErrClientIDExists = errors.New("linear client id already registered")
	// ErrRevisionConflict means the caller's expected revision no longer
	// matches the stored one; the write was not applied.
	ErrRevisionConflict = errors.New("revision conflict")
)

// Credential is one encrypted secret plus the master key ID that sealed it.
type Credential struct {
	Ciphertext string
	KeyID      string
}

// Set reports whether a credential is actually present.
func (c Credential) Set() bool { return c.Ciphertext != "" }

// AppCredentials are an App's two write-only secrets.
type AppCredentials struct {
	ClientSecret  Credential
	WebhookSecret Credential
}

// CredentialChange updates an App's secrets. A nil field keeps the stored
// secret; a non-nil unset Credential clears it.
type CredentialChange struct {
	ClientSecret  *Credential
	WebhookSecret *Credential
}

// Repository persists Linear Apps.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	ListApps(ctx context.Context, workspaceID string) ([]*agentsv1.LinearApp, error)
	GetApp(ctx context.Context, workspaceID, id string) (*agentsv1.LinearApp, error)
	// FindApp resolves an App by ID without a workspace scope. The public
	// webhook route carries only the App ID, so the receive path reads the
	// workspace off the returned App.
	FindApp(ctx context.Context, id string) (*agentsv1.LinearApp, error)

	// CreateApp stores a new App with its secrets in one operation. It
	// returns ErrClientIDExists when any workspace already registered the
	// client ID.
	CreateApp(ctx context.Context, workspaceID string, app *agentsv1.LinearApp, creds AppCredentials) (*agentsv1.LinearApp, error)
	// UpdateApp replaces the mutable fields of an App when the stored
	// revision equals expectedRevision, returning ErrRevisionConflict
	// without writing otherwise. The ID and client ID are preserved
	// regardless of what the caller passes.
	UpdateApp(ctx context.Context, workspaceID string, app *agentsv1.LinearApp, expectedRevision int64) (*agentsv1.LinearApp, error)
	// SetAppCredentials applies change atomically and returns the App with
	// its derived credential fields refreshed.
	SetAppCredentials(ctx context.Context, workspaceID, id string, change CredentialChange) (*agentsv1.LinearApp, error)
	// GetAppCredentials returns the stored secrets' ciphertext; unset
	// fields are simply not Set().
	GetAppCredentials(ctx context.Context, workspaceID, id string) (AppCredentials, error)
	DeleteApp(ctx context.Context, workspaceID, id string) error
}

// StampCredentialState fills the App's derived credential fields from the
// stored secrets. Implementations call it on every read.
func StampCredentialState(app *agentsv1.LinearApp, creds AppCredentials) {
	app.ClientSecretSet = creds.ClientSecret.Set()
	app.WebhookSecretSet = creds.WebhookSecret.Set()
	if app.ClientSecretSet && app.WebhookSecretSet {
		app.CredentialState = agentsv1.LinearAppCredentialState_LINEAR_APP_CREDENTIAL_STATE_COMPLETE
	} else {
		app.CredentialState = agentsv1.LinearAppCredentialState_LINEAR_APP_CREDENTIAL_STATE_INCOMPLETE
	}
}

// StripDerived clears every server-derived field before storage, so stored
// specs can never contradict the credential columns or the platform URL.
func StripDerived(app *agentsv1.LinearApp) {
	app.WorkspaceId = ""
	app.CredentialState = agentsv1.LinearAppCredentialState_LINEAR_APP_CREDENTIAL_STATE_UNSPECIFIED
	app.ClientSecretSet = false
	app.WebhookSecretSet = false
	app.CallbackUrl = ""
	app.WebhookUrl = ""
}
