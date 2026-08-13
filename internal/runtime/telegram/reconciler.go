package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/repo/telegramsetting"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// WebhookPathPrefix is the public callback path, minus the Channel ID.
const WebhookPathPrefix = "/api/telegram/webhook/"

// Leader is the leadership token the reconciler holds. Declaring it as an
// interface keeps reconciliation testable without Redis: what matters is what
// the leader does with Telegram, not how leadership is decided.
type Leader interface {
	Acquire(ctx context.Context) (bool, error)
	Release(ctx context.Context) error
}

// ReconcileState is the derived Webhook registration status for one Channel.
// It is observed, never persisted as configuration.
type ReconcileState struct {
	State        agentsv1.TelegramWebhookState
	URL          string
	Error        string
	ReconciledAt time.Time
}

// Reconciler keeps Telegram's registration in step with desired Channel
// state.
//
// Exactly one Pod runs it at a time, elected by a Redis lease: several Pods
// calling setWebhook would overwrite each other's registrations, and Telegram
// keeps only the last one. Reconciliation is a comparison, not a write —
// getWebhookInfo first, and setWebhook/deleteWebhook only when observed and
// desired actually differ — so a steady-state fleet issues no writes at all.
type Reconciler struct {
	repo       telegramrepo.Repository
	settings   telegramsetting.Repository
	keyring    *secretbox.Keyring
	newClient  func(token string) telegramapi.WebhookClient
	lease      Leader
	interval   time.Duration
	stop       chan struct{}
	stopped    chan struct{}
	stopOnce   sync.Once
	mu         sync.RWMutex
	states     map[string]ReconcileState
	holdsLease bool
}

// NewReconciler builds the reconciler. `newClient` is injectable so tests can
// substitute a fake Bot API.
func NewReconciler(
	repo telegramrepo.Repository,
	settings telegramsetting.Repository,
	keyring *secretbox.Keyring,
	newClient func(token string) telegramapi.WebhookClient,
	lease Leader,
	interval time.Duration,
) *Reconciler {
	if newClient == nil {
		newClient = func(token string) telegramapi.WebhookClient { return telegramapi.New(token) }
	}
	if interval <= 0 {
		interval = time.Minute
	}
	return &Reconciler{
		repo:      repo,
		settings:  settings,
		keyring:   keyring,
		newClient: newClient,
		lease:     lease,
		interval:  interval,
		stop:      make(chan struct{}),
		stopped:   make(chan struct{}),
		states:    make(map[string]ReconcileState),
	}
}

// Start runs the reconcile loop until Stop.
func (r *Reconciler) Start(ctx context.Context) {
	go func() {
		defer close(r.stopped)
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		r.tick(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.stop:
				// Give the lease up immediately so the next leader does not
				// wait out the TTL.
				releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
				_ = r.lease.Release(releaseCtx)
				cancel()
				return
			case <-ticker.C:
				r.tick(ctx)
			}
		}
	}()
}

// Stop halts the loop and releases leadership.
func (r *Reconciler) Stop() {
	r.stopOnce.Do(func() { close(r.stop) })
	<-r.stopped
}

// State returns the last observed registration state for a Channel.
func (r *Reconciler) State(channelID string) (ReconcileState, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	state, ok := r.states[channelID]
	return state, ok
}

func (r *Reconciler) tick(ctx context.Context) {
	logger := log.FromContext(ctx)
	held, err := r.lease.Acquire(ctx)
	if err != nil {
		logger.Warn("telegram webhook reconciler could not evaluate leadership", "err", err)
		return
	}
	r.mu.Lock()
	r.holdsLease = held
	r.mu.Unlock()
	if !held {
		return
	}

	channels, err := r.repo.ListChannelsAcrossWorkspaces(ctx)
	if err != nil {
		logger.Error("telegram webhook reconciler could not list channels", "err", err)
		return
	}
	settings, err := r.settings.Get(ctx)
	if err != nil {
		logger.Error("telegram webhook reconciler could not read settings", "err", err)
		return
	}
	baseURL := strings.TrimRight(strings.TrimSpace(settings.GetWebhookBaseUrl()), "/")

	for _, channel := range channels {
		r.reconcileChannel(ctx, channel, baseURL)
	}
}

// HoldsLease reports whether this Pod is currently the reconciler leader.
func (r *Reconciler) HoldsLease() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.holdsLease
}

func (r *Reconciler) reconcileChannel(ctx context.Context, channel *agentsv1.TelegramChannel, baseURL string) {
	logger := log.FromContext(ctx)
	if channel.GetReceiveMode() != agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK {
		// A Long Polling Channel is not reconciled as an active Webhook, but
		// a *stale* registration still has to go: Telegram refuses getUpdates
		// with a 409 while a webhook is set, so leaving one behind would make
		// a freshly switched Channel silently receive nothing.
		r.clearStaleWebhook(ctx, channel)
		return
	}

	desiredURL := ""
	if channel.GetInboundEnabled() && baseURL != "" {
		desiredURL = CallbackURL(baseURL, channel.GetId())
	}

	client, secret, err := r.clientFor(ctx, channel)
	if err != nil {
		r.fail(channel.GetId(), desiredURL, err)
		return
	}

	info, err := client.GetWebhookInfo(ctx)
	if err != nil {
		r.fail(channel.GetId(), desiredURL, err)
		return
	}

	switch {
	case desiredURL == "" && info.URL == "":
		// Nothing registered, nothing wanted.
	case desiredURL == "":
		if err := client.DeleteWebhook(ctx, false); err != nil {
			r.fail(channel.GetId(), desiredURL, err)
			return
		}
		logger.Info("removed telegram webhook registration", "channel_id", channel.GetId())
	case info.URL != desiredURL:
		if secret == "" {
			r.fail(channel.GetId(), desiredURL, errors.New("no webhook secret is stored for this channel"))
			return
		}
		if err := client.SetWebhook(ctx, telegramapi.SetWebhookParams{
			URL:         desiredURL,
			SecretToken: secret,
		}); err != nil {
			r.fail(channel.GetId(), desiredURL, err)
			return
		}
		logger.Info("registered telegram webhook", "channel_id", channel.GetId(), "url", desiredURL)
	}

	state := agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_REGISTERED
	if desiredURL == "" {
		state = agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_PENDING
	}
	r.record(channel.GetId(), ReconcileState{
		State:        state,
		URL:          desiredURL,
		Error:        info.LastErrorMessage,
		ReconciledAt: time.Now().UTC(),
	})
}

// clearStaleWebhook removes a registration left over from Webhook mode.
func (r *Reconciler) clearStaleWebhook(ctx context.Context, channel *agentsv1.TelegramChannel) {
	state := ReconcileState{
		State:        agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_NOT_APPLICABLE,
		ReconciledAt: time.Now().UTC(),
	}
	client, _, err := r.clientFor(ctx, channel)
	if err != nil {
		// Without a usable credential there is nothing to clear and nothing
		// to poll either; the enablement preflight reports that separately.
		r.record(channel.GetId(), state)
		return
	}
	info, err := client.GetWebhookInfo(ctx)
	if err != nil {
		state.Error = err.Error()
		r.record(channel.GetId(), state)
		return
	}
	if info.URL == "" {
		r.record(channel.GetId(), state)
		return
	}
	if err := client.DeleteWebhook(ctx, false); err != nil {
		state.State = agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_FAILED
		state.Error = err.Error()
		r.record(channel.GetId(), state)
		return
	}
	log.FromContext(ctx).Info("removed a stale telegram webhook so long polling can run",
		"channel_id", channel.GetId())
	r.record(channel.GetId(), state)
}

// clientFor decrypts the Channel's Bot Token and Webhook secret.
func (r *Reconciler) clientFor(ctx context.Context, channel *agentsv1.TelegramChannel) (telegramapi.WebhookClient, string, error) {
	cred, err := r.repo.GetChannelCredential(ctx, channel.GetWorkspaceId(), channel.GetId())
	if err != nil {
		return nil, "", fmt.Errorf("read channel credential: %w", err)
	}
	token, err := r.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
	if err != nil {
		return nil, "", fmt.Errorf("decrypt channel credential: %w", err)
	}
	secret := ""
	stored, err := r.repo.GetWebhookSecret(ctx, channel.GetWorkspaceId(), channel.GetId())
	switch {
	case err == nil:
		plaintext, decErr := r.keyring.Decrypt(ctx, stored.Ciphertext, stored.KeyID)
		if decErr != nil {
			return nil, "", fmt.Errorf("decrypt webhook secret: %w", decErr)
		}
		secret = string(plaintext)
	case errors.Is(err, telegramrepo.ErrNoCredential):
		// Handled by the caller: a registration needs one, a removal does not.
	default:
		return nil, "", fmt.Errorf("read webhook secret: %w", err)
	}
	return r.newClient(string(token)), secret, nil
}

func (r *Reconciler) record(channelID string, state ReconcileState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[channelID] = state
}

func (r *Reconciler) fail(channelID, url string, err error) {
	r.record(channelID, ReconcileState{
		State:        agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_FAILED,
		URL:          url,
		Error:        err.Error(),
		ReconciledAt: time.Now().UTC(),
	})
}

// CallbackURL derives a Channel's public callback URL. It is derived rather
// than stored so changing the deployment's public host does not require
// touching any Channel.
func CallbackURL(baseURL, channelID string) string {
	return strings.TrimRight(baseURL, "/") + WebhookPathPrefix + channelID
}
