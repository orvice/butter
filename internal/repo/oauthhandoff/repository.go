// Package oauthhandoff stores the one-time handoff codes of the native
// (mobile) OAuth login. After the provider returns to the server's fixed
// callback and the code exchange succeeds, the server records who signed in
// here and gives the browser only an opaque handoff code. The App redeems it
// once, together with its PKCE code_verifier, and only then is a session
// minted — so no session token ever travels through a URL.
package oauthhandoff

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a handoff code is unknown, expired or already
// redeemed.
var ErrNotFound = errors.New("oauth handoff not found")

// Entry is what a handoff code stands for. CodeHash is the SHA-256 (hex) of
// the code that was handed to the browser: the store never holds a code that
// could be redeemed from a database read alone.
type Entry struct {
	CodeHash string
	Provider string
	// State is the OAuth state of the flow this handoff completes; redemption
	// must present the same one.
	State string
	// UserID is the signed-in user the redeemed session will belong to.
	UserID string
	// CodeChallenge is the PKCE S256 challenge recorded when the flow began.
	CodeChallenge string
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

// Repository persists handoff entries with a TTL.
type Repository interface {
	// EnsureIndexes creates any required storage indexes.
	EnsureIndexes(ctx context.Context) error
	// Create stores an entry. Implementations must reject duplicates.
	Create(ctx context.Context, entry *Entry) error
	// Consume atomically loads and deletes the entry. Returns ErrNotFound if
	// the code is unknown, expired, or already consumed.
	Consume(ctx context.Context, codeHash string, now time.Time) (*Entry, error)
}
