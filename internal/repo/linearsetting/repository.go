// Package linearsetting stores the platform-level Linear settings
// (ADR-0015): currently just the public base URL that Linear callback and
// webhook URLs are derived from.
//
// These are deliberately not workspace-scoped. The base URL names the public
// address of this deployment — a platform fact, not a tenant choice — and
// letting workspace input set it would let a tenant redirect another
// tenant's Linear webhooks and OAuth callbacks.
package linearsetting

import (
	"context"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Repository persists the singleton settings document.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Get returns the current settings. A deployment that has never
	// configured them gets a zero-valued message rather than an error.
	Get(ctx context.Context) (*agentsv1.LinearSettings, error)

	// Put replaces the settings and returns what was stored.
	Put(ctx context.Context, settings *agentsv1.LinearSettings) (*agentsv1.LinearSettings, error)
}
