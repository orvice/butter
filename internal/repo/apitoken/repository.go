package apitoken

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ErrNotFound is returned when a token lookup misses.
var ErrNotFound = errors.New("api token not found")

// Repository persists API token records.
//
// Tokens are stored hashed; the plaintext secret is only known to the caller
// at create time. Authentication is performed via Lookup using the SHA-256
// hash of the bearer secret.
type Repository interface {
	List(ctx context.Context) ([]*agentsv1.APIToken, error)
	Get(ctx context.Context, id string) (*agentsv1.APIToken, error)
	Create(ctx context.Context, token *agentsv1.APIToken, secretHash string) error
	Revoke(ctx context.Context, id string) (*agentsv1.APIToken, error)
	// Lookup returns the token whose secret hash matches the argument or
	// ErrNotFound. Revoked tokens are filtered out by the implementation.
	Lookup(ctx context.Context, secretHash string) (*agentsv1.APIToken, error)
	// TouchLastUsed updates the last_used_at timestamp for the given token.
	// Failures should be logged but not propagated to the caller of auth.
	TouchLastUsed(ctx context.Context, id string) error
}
