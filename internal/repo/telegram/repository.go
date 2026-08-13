// Package telegram stores Telegram Channels and Destinations (issue #264).
//
// Channels and Destinations live behind one repository because their
// invariants are joint: a Destination's address is only unique relative to a
// Channel, and a Channel may not be deleted while a Destination references
// it. Splitting them would move those checks into the service layer, where
// concurrent callers could interleave past them.
//
// Bot Tokens and Webhook secrets are handled through a separate credential
// seam, exactly as WorkspaceRepoBinding handles PATs (ADR-0005): callers pass
// pre-encrypted ciphertext in and get ciphertext out, so implementations never
// see plaintext and a credential can never ride along on a Channel read into
// an API response or log line. Implementations derive the Channel proto's
// server-owned credential_state / credential_updated_at / webhook_secret_set
// fields from the stored credential on every read.
package telegram

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrNotFound means no such Channel or Destination exists.
	ErrNotFound = errors.New("not found")
	// ErrKeyExists means the workspace already has a resource with that key.
	ErrKeyExists = errors.New("key already exists")
	// ErrBotExists means another Channel — in any workspace — is already
	// pinned to that Telegram Bot ID.
	ErrBotExists = errors.New("bot already registered")
	// ErrAddressExists means another Destination already covers that exact
	// (channel, chat, thread) address.
	ErrAddressExists = errors.New("address already registered")
	// ErrRevisionConflict means the caller's expected revision no longer
	// matches the stored one; the write was not applied.
	ErrRevisionConflict = errors.New("revision conflict")
	// ErrNoCredential means the Channel exists but no credential is stored.
	ErrNoCredential = errors.New("no credential")
	// ErrChannelInUse means a Destination still references the Channel.
	ErrChannelInUse = errors.New("channel is referenced by a destination")
)

// Credential is one encrypted secret plus the master key ID that sealed it.
// Storing the key ID alongside the ciphertext is what makes a future key
// rotation or Secret Manager migration possible.
type Credential struct {
	Ciphertext string
	KeyID      string
}

// Set reports whether a credential is actually present.
func (c Credential) Set() bool { return c.Ciphertext != "" }

// ChannelCredentials groups the write-only secrets that must be committed
// atomically with a new Channel. A Webhook Channel is never observable without
// its callback secret, even when persistence fails.
type ChannelCredentials struct {
	BotToken      Credential
	WebhookSecret Credential
}

// Repository persists Telegram Channels and Destinations.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// --- Channels -------------------------------------------------------

	ListChannels(ctx context.Context, workspaceID string) ([]*agentsv1.TelegramChannel, error)
	GetChannel(ctx context.Context, workspaceID, id string) (*agentsv1.TelegramChannel, error)

	// CreateChannel stores a new Channel together with its credentials in one
	// operation. It returns ErrKeyExists on a workspace key collision and
	// ErrBotExists when any workspace already owns that Bot ID — the check
	// that stops two Channels from consuming the same Bot's updates.
	CreateChannel(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, credentials ChannelCredentials) (*agentsv1.TelegramChannel, error)

	// UpdateChannel replaces the mutable fields of an existing Channel. It
	// applies only when the stored revision equals expectedRevision, and
	// returns ErrRevisionConflict without writing otherwise. The stored key
	// and Bot identity are preserved regardless of what the caller passes.
	UpdateChannel(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, expectedRevision int64) (*agentsv1.TelegramChannel, error)
	// RotateChannelCredential atomically replaces the mutable Channel metadata
	// and encrypted Bot Token under the same optimistic revision check. A
	// revision conflict or persistence failure leaves both unchanged.
	RotateChannelCredential(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, cred Credential, expectedRevision int64) (*agentsv1.TelegramChannel, error)

	// DeleteChannel removes a Channel and its stored credentials. It returns
	// ErrChannelInUse while any Destination still references it.
	DeleteChannel(ctx context.Context, workspaceID, id string) error

	// SetChannelCredential replaces the stored Bot Token. An empty
	// ciphertext clears it.
	SetChannelCredential(ctx context.Context, workspaceID, id string, cred Credential) error
	// GetChannelCredential returns the stored Bot Token ciphertext,
	// ErrNotFound when there is no such Channel, or ErrNoCredential when
	// none was stored.
	GetChannelCredential(ctx context.Context, workspaceID, id string) (Credential, error)

	// SetWebhookSecret replaces the per-Channel Telegram secret token used to
	// authenticate callback requests. An empty ciphertext clears it.
	SetWebhookSecret(ctx context.Context, workspaceID, id string, cred Credential) error
	// GetWebhookSecret returns the stored Webhook secret ciphertext.
	GetWebhookSecret(ctx context.Context, workspaceID, id string) (Credential, error)

	// FindChannel resolves a Channel by ID without a workspace scope. The
	// Telegram callback route is public and carries only the Channel ID, so
	// the receive path has no workspace to scope by; it reads the workspace
	// off the returned Channel.
	FindChannel(ctx context.Context, id string) (*agentsv1.TelegramChannel, error)

	// ListChannelsAcrossWorkspaces returns the flat global view used by
	// runtime layers (Webhook reconciler, Long Polling leader).
	ListChannelsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.TelegramChannel, error)

	// --- Destinations ---------------------------------------------------

	// ListDestinations returns the workspace's Destinations, narrowed to one
	// Channel when channelID is non-empty.
	ListDestinations(ctx context.Context, workspaceID, channelID string) ([]*agentsv1.TelegramDestination, error)
	GetDestination(ctx context.Context, workspaceID, id string) (*agentsv1.TelegramDestination, error)

	// CreateDestination stores a new Destination. It returns ErrKeyExists on
	// a workspace key collision and ErrAddressExists when the exact
	// (channel, chat, thread) address is already covered — the check that
	// guarantees one inbound update can never match two Destinations.
	CreateDestination(ctx context.Context, workspaceID string, dest *agentsv1.TelegramDestination) (*agentsv1.TelegramDestination, error)

	// UpdateDestination replaces the mutable fields of an existing
	// Destination under the same optimistic revision rule as UpdateChannel.
	// The stored key and address are preserved regardless of input.
	UpdateDestination(ctx context.Context, workspaceID string, dest *agentsv1.TelegramDestination, expectedRevision int64) (*agentsv1.TelegramDestination, error)

	DeleteDestination(ctx context.Context, workspaceID, id string) error

	// FindDestinationByAddress resolves the single Destination covering an
	// exact Telegram address. threadID is empty for non-Topic addresses.
	// Returns ErrNotFound for unknown addresses, which the receive path
	// treats as "ignore", not "error".
	FindDestinationByAddress(ctx context.Context, channelID, chatID, threadID string) (*agentsv1.TelegramDestination, error)

	// ListDestinationsAcrossWorkspaces returns the flat global view.
	ListDestinationsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.TelegramDestination, error)
}
