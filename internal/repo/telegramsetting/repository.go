// Package telegramsetting stores the platform-level Telegram settings
// (issue #264): currently just the public Webhook base URL.
//
// These are deliberately not workspace-scoped. The base URL names the public
// address of this deployment behind its load balancer — a platform fact, not
// a tenant choice — and letting workspace input set it would let a tenant
// redirect another tenant's Telegram callbacks.
package telegramsetting

import (
	"context"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Repository persists the singleton settings document.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Get returns the current settings. A deployment that has never
	// configured them gets a zero-valued message rather than an error: "not
	// configured yet" is a normal state the Dashboard renders.
	Get(ctx context.Context) (*agentsv1.TelegramSettings, error)

	// Put replaces the settings and returns what was stored.
	Put(ctx context.Context, settings *agentsv1.TelegramSettings) (*agentsv1.TelegramSettings, error)
}
