// Package cryptokey stores the database-backed master key that encrypts
// credential material owned by other repositories (issue #264: Telegram Bot
// Tokens and per-Channel Webhook secrets).
//
// The key is transitional. It lives in the same database as the ciphertext it
// protects, so it does not defend against complete database compromise — that
// is the explicit trade-off recorded in #264 until Secret Manager/KMS support
// lands. What it does buy is that no YAML secret is required to configure
// credentials in this release, and that every ciphertext record carries the
// `key_id` a future rotation or migration will need.
//
// Initialization is atomic across Pods: EnsureActive inserts the pointer
// document and treats a duplicate-key error as "another Pod won", reading the
// winner back rather than overwriting it. Two Pods starting simultaneously
// therefore converge on one key instead of each encrypting under its own.
package cryptokey

import (
	"context"
	"errors"
)

// ErrNotFound means no key with the requested ID exists.
var ErrNotFound = errors.New("not found")

// MasterKey is one versioned encryption key.
type MasterKey struct {
	// ID identifies this key in ciphertext records.
	ID string
	// Material is the raw key (32 bytes for the AES-256 keys this package
	// generates).
	Material []byte
}

// Repository persists master keys.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// EnsureActive returns the active master key, creating it on first use.
	// Concurrent callers — including callers in different processes — all
	// return the same key: exactly one insert wins and the losers read the
	// winner back. `generate` is only consulted when no key exists yet, and
	// may be called even by a caller that ultimately loses the race.
	EnsureActive(ctx context.Context, generate func() (MasterKey, error)) (MasterKey, error)

	// Get returns the key with the given ID, or ErrNotFound. Used to decrypt
	// ciphertext written under a key that is no longer active.
	Get(ctx context.Context, id string) (MasterKey, error)
}
