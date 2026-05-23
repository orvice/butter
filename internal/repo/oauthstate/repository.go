// Package oauthstate stores short-lived CSRF state tokens used by the OAuth
// login flow. AuthService.BeginOAuthFlow generates a random state and stores
// it here together with the provider name and redirect URI; CompleteOAuthFlow
// looks it up and consumes it (single-use).
package oauthstate

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a state token is unknown or has already been
// consumed / expired.
var ErrNotFound = errors.New("oauth state not found")

// Entry is the data stored against a state token.
type Entry struct {
	State       string
	Provider    string
	RedirectURI string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// Repository persists OAuth state tokens with a TTL.
type Repository interface {
	// EnsureIndexes creates any required storage indexes.
	EnsureIndexes(ctx context.Context) error
	// Create stores an entry. Implementations must reject duplicates.
	Create(ctx context.Context, entry *Entry) error
	// Consume atomically loads and deletes the entry for state. Returns
	// ErrNotFound if state is unknown, expired, or already consumed.
	Consume(ctx context.Context, state string, now time.Time) (*Entry, error)
}
