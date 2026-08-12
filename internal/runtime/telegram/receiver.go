package telegram

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrUnauthorized means the presented secret did not match.
	ErrUnauthorized = errors.New("invalid telegram webhook secret")
	// ErrChannelNotFound means no Channel has that ID.
	ErrChannelNotFound = errors.New("unknown telegram channel")
	// ErrChannelNotReceiving means the Channel exists but is not accepting
	// updates right now.
	ErrChannelNotReceiving = errors.New("telegram channel is not receiving")
	// ErrMalformedUpdate means the payload can never be processed. Callers
	// acknowledge it rather than asking Telegram to retry.
	ErrMalformedUpdate = errors.New("malformed telegram update")
)

// AuthenticatedChannel is a Channel that presented a valid secret.
type AuthenticatedChannel struct {
	Channel *agentsv1.TelegramChannel
}

// Receiver authenticates and routes inbound Telegram callbacks.
//
// Every request reloads Channel state from the database. That costs a read
// per update, and it is the point: a Channel disabled seconds ago must stop
// accepting, and a cross-request configuration cache would keep accepting
// until it expired.
type Receiver struct {
	repo    telegramrepo.Repository
	keyring *secretbox.Keyring
	router  *Router
}

func NewReceiver(repo telegramrepo.Repository, keyring *secretbox.Keyring, router *Router) *Receiver {
	return &Receiver{repo: repo, keyring: keyring, router: router}
}

// Authenticate resolves the Channel and verifies the per-Channel secret in
// constant time before any payload is parsed.
func (r *Receiver) Authenticate(ctx context.Context, channelID, presented string) (AuthenticatedChannel, error) {
	if r == nil || r.repo == nil || r.keyring == nil || r.router == nil {
		return AuthenticatedChannel{}, errors.New("telegram receiver is not configured")
	}
	channel, err := r.repo.FindChannel(ctx, channelID)
	if err != nil {
		if errors.Is(err, telegramrepo.ErrNotFound) {
			return AuthenticatedChannel{}, ErrChannelNotFound
		}
		return AuthenticatedChannel{}, err
	}

	stored, err := r.repo.GetWebhookSecret(ctx, channel.GetWorkspaceId(), channel.GetId())
	if err != nil {
		if errors.Is(err, telegramrepo.ErrNoCredential) {
			// No secret means no authenticated caller is possible.
			return AuthenticatedChannel{}, ErrUnauthorized
		}
		return AuthenticatedChannel{}, err
	}
	expected, err := r.keyring.Decrypt(ctx, stored.Ciphertext, stored.KeyID)
	if err != nil {
		return AuthenticatedChannel{}, fmt.Errorf("decrypt webhook secret: %w", err)
	}
	if subtle.ConstantTimeCompare([]byte(presented), expected) != 1 {
		return AuthenticatedChannel{}, ErrUnauthorized
	}

	// Authentication succeeded, but a Channel that is not receiving must not
	// have its updates queued. This is checked after the secret so an
	// unauthenticated caller cannot probe which Channels are enabled.
	if !channel.GetInboundEnabled() ||
		channel.GetReceiveMode() != agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK {
		return AuthenticatedChannel{}, ErrChannelNotReceiving
	}
	return AuthenticatedChannel{Channel: channel}, nil
}

// Deliver routes one raw update for an authenticated Channel. The router
// distinguishes "this payload is unusable" (ErrMalformedUpdate, which the
// caller acknowledges) from "we could not process it" (which it retries).
func (r *Receiver) Deliver(ctx context.Context, channel AuthenticatedChannel, raw []byte) (Decision, error) {
	return r.router.Route(ctx, channel.Channel, raw)
}
