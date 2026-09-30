// Package linearstate stores the short-lived, single-use states of Linear
// App install flows (ADR-0015). BeginLinearInstall issues a state bound to
// the workspace, App, initiating user and redirect URI; the public OAuth
// callback consumes it. The states live in the database so the callback
// works on any Pod, and only a hash of each state is stored.
package linearstate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

// ErrNotFound is returned when a state is unknown, expired, or consumed.
var ErrNotFound = errors.New("linear install state not found")

// Entry is what one install state is bound to.
type Entry struct {
	// State is the plaintext state; implementations store only its hash.
	State       string
	WorkspaceID string
	AppID       string
	UserID      string
	RedirectURI string
	ReturnURL   string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// Repository persists install states with a TTL.
type Repository interface {
	EnsureIndexes(ctx context.Context) error
	// Create stores an entry.
	Create(ctx context.Context, entry *Entry) error
	// Consume atomically loads and deletes the entry for state, returning
	// ErrNotFound when it is unknown, expired, or already consumed.
	Consume(ctx context.Context, state string, now time.Time) (*Entry, error)
}

// Hash is the storage key of a state.
func Hash(state string) string {
	sum := sha256.Sum256([]byte(state))
	return hex.EncodeToString(sum[:])
}
