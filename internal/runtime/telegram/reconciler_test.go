package telegram

// Reconciliation tests (issue #264/#267): only the leader touches Telegram,
// registration is repaired rather than rewritten, and a disabled Channel has
// its registration removed.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	telegramsettingmemory "go.orx.me/apps/butter/internal/repo/telegramsetting/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeLeader hands out leadership on demand.
type fakeLeader struct {
	held     bool
	acquires int
	released bool
}

func (l *fakeLeader) Acquire(context.Context) (bool, error) { l.acquires++; return l.held, nil }
func (l *fakeLeader) Release(context.Context) error         { l.released = true; return nil }

// fakeWebhookClient records what the reconciler asked Telegram to do.
type fakeWebhookClient struct {
	mu       sync.Mutex
	info     telegramapi.WebhookInfo
	setCalls []telegramapi.SetWebhookParams
	deletes  int
	setErr   error
}

func (c *fakeWebhookClient) GetWebhookInfo(context.Context) (telegramapi.WebhookInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.info, nil
}

func (c *fakeWebhookClient) SetWebhook(_ context.Context, params telegramapi.SetWebhookParams) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.setErr != nil {
		return c.setErr
	}
	c.setCalls = append(c.setCalls, params)
	c.info.URL = params.URL
	return nil
}

func (c *fakeWebhookClient) DeleteWebhook(context.Context, bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deletes++
	c.info.URL = ""
	return nil
}

type reconcilerFixture struct {
	reconciler *Reconciler
	repo       *telegrammemory.Store
	settings   *telegramsettingmemory.Store
	client     *fakeWebhookClient
	leader     *fakeLeader
}

func newReconcilerFixture(t *testing.T, inbound bool, baseURL string) *reconcilerFixture {
	t.Helper()
	repo := telegrammemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())

	tokenCipher, tokenKey, err := keyring.Encrypt(t.Context(), []byte("111111:token"))
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if _, err := repo.CreateChannel(t.Context(), "ws-a", &agentsv1.TelegramChannel{
		Id: "ch-1", Key: "main", BotId: "111111",
		InboundEnabled: inbound, OutboundEnabled: true,
		ReceiveMode: agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK,
	}, telegramrepo.ChannelCredentials{
		BotToken: telegramrepo.Credential{Ciphertext: tokenCipher, KeyID: tokenKey},
	}); err != nil {
		t.Fatalf("create channel: %v", err)
	}
	secretCipher, secretKey, err := keyring.Encrypt(t.Context(), []byte(webhookSecret))
	if err != nil {
		t.Fatalf("encrypt secret: %v", err)
	}
	if err := repo.SetWebhookSecret(t.Context(), "ws-a", "ch-1",
		telegramrepo.Credential{Ciphertext: secretCipher, KeyID: secretKey}); err != nil {
		t.Fatalf("store secret: %v", err)
	}

	settings := telegramsettingmemory.New()
	if _, err := settings.Put(t.Context(), &agentsv1.TelegramSettings{WebhookBaseUrl: baseURL}); err != nil {
		t.Fatalf("store settings: %v", err)
	}

	client := &fakeWebhookClient{}
	leader := &fakeLeader{held: true}
	reconciler := NewReconciler(repo, settings, keyring,
		func(string) telegramapi.WebhookClient { return client }, leader, time.Minute)
	return &reconcilerFixture{
		reconciler: reconciler, repo: repo, settings: settings, client: client, leader: leader,
	}
}

func TestReconcilerRegistersTheDerivedCallbackURL(t *testing.T) {
	fx := newReconcilerFixture(t, true, "https://butter.test")

	fx.reconciler.tick(t.Context())

	if len(fx.client.setCalls) != 1 {
		t.Fatalf("setWebhook calls = %d, want 1", len(fx.client.setCalls))
	}
	call := fx.client.setCalls[0]
	if call.URL != "https://butter.test/api/telegram/webhook/ch-1" {
		t.Errorf("url = %q", call.URL)
	}
	if call.SecretToken != webhookSecret {
		t.Errorf("secret token was not the stored per-channel secret")
	}
	state, _ := fx.reconciler.State("ch-1")
	if state.State != agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_REGISTERED {
		t.Errorf("state = %v", state.State)
	}
}

// Steady state must issue no writes: reconciliation compares first.
func TestReconcilerIsANoOpWhenAlreadyRegistered(t *testing.T) {
	fx := newReconcilerFixture(t, true, "https://butter.test")
	fx.client.info.URL = "https://butter.test/api/telegram/webhook/ch-1"

	fx.reconciler.tick(t.Context())

	if len(fx.client.setCalls) != 0 || fx.client.deletes != 0 {
		t.Fatalf("expected no writes, got %d sets and %d deletes",
			len(fx.client.setCalls), fx.client.deletes)
	}
}

// A registration pointing at a stale host is repaired, not left alone.
func TestReconcilerRepairsAStaleRegistration(t *testing.T) {
	fx := newReconcilerFixture(t, true, "https://butter.test")
	fx.client.info.URL = "https://old-host.test/api/telegram/webhook/ch-1"

	fx.reconciler.tick(t.Context())

	if len(fx.client.setCalls) != 1 {
		t.Fatalf("setWebhook calls = %d, want the stale URL repaired", len(fx.client.setCalls))
	}
}

func TestReconcilerRemovesTheRegistrationWhenInboundIsDisabled(t *testing.T) {
	fx := newReconcilerFixture(t, false, "https://butter.test")
	fx.client.info.URL = "https://butter.test/api/telegram/webhook/ch-1"

	fx.reconciler.tick(t.Context())

	if fx.client.deletes != 1 {
		t.Fatalf("deleteWebhook calls = %d, want 1", fx.client.deletes)
	}
}

// Without a base URL there is nowhere to register, so the Channel stays
// pending rather than being registered somewhere wrong.
func TestReconcilerWaitsForABaseURL(t *testing.T) {
	fx := newReconcilerFixture(t, true, "")

	fx.reconciler.tick(t.Context())

	if len(fx.client.setCalls) != 0 {
		t.Fatalf("expected no registration without a base URL")
	}
	state, _ := fx.reconciler.State("ch-1")
	if state.State != agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_PENDING {
		t.Errorf("state = %v, want PENDING", state.State)
	}
}

// Only the leader touches Telegram: several Pods calling setWebhook would
// overwrite each other, and Telegram keeps only the last.
func TestOnlyTheLeaderReconciles(t *testing.T) {
	fx := newReconcilerFixture(t, true, "https://butter.test")
	fx.leader.held = false

	fx.reconciler.tick(t.Context())

	if len(fx.client.setCalls) != 0 {
		t.Fatalf("a non-leader issued %d registrations", len(fx.client.setCalls))
	}
	if fx.reconciler.HoldsLease() {
		t.Error("HoldsLease must report the lease was not taken")
	}

	// Once it wins the lease, the same Pod reconciles.
	fx.leader.held = true
	fx.reconciler.tick(t.Context())
	if len(fx.client.setCalls) != 1 {
		t.Fatalf("the new leader issued %d registrations, want 1", len(fx.client.setCalls))
	}
	if !fx.reconciler.HoldsLease() {
		t.Error("HoldsLease must report the lease was taken")
	}
}

func TestReconcilerRecordsAFailureWithoutRetryingForever(t *testing.T) {
	fx := newReconcilerFixture(t, true, "https://butter.test")
	fx.client.setErr = errors.New("telegram error 400: bad webhook")

	fx.reconciler.tick(t.Context())

	state, ok := fx.reconciler.State("ch-1")
	if !ok {
		t.Fatal("expected a recorded state")
	}
	if state.State != agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_FAILED {
		t.Errorf("state = %v, want FAILED", state.State)
	}
	if state.Error == "" {
		t.Error("expected the failure to be reported")
	}
}

// A Long Polling channel is not the reconciler's business.
func TestReconcilerSkipsLongPollingChannels(t *testing.T) {
	fx := newReconcilerFixture(t, true, "https://butter.test")
	channel, err := fx.repo.GetChannel(t.Context(), "ws-a", "ch-1")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	channel.ReceiveMode = agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_LONG_POLLING
	if _, err := fx.repo.UpdateChannel(t.Context(), "ws-a", channel, channel.GetRevision()); err != nil {
		t.Fatalf("switch mode: %v", err)
	}

	fx.reconciler.tick(t.Context())

	if len(fx.client.setCalls) != 0 || fx.client.deletes != 0 {
		t.Fatal("long polling channels must not be registered")
	}
	state, _ := fx.reconciler.State("ch-1")
	if state.State != agentsv1.TelegramWebhookState_TELEGRAM_WEBHOOK_STATE_NOT_APPLICABLE {
		t.Errorf("state = %v, want NOT_APPLICABLE", state.State)
	}
}

func TestCallbackURLIsDerivedFromTheBaseURL(t *testing.T) {
	for _, tc := range []struct{ base, want string }{
		{"https://butter.test", "https://butter.test/api/telegram/webhook/ch-1"},
		{"https://butter.test/", "https://butter.test/api/telegram/webhook/ch-1"},
	} {
		if got := CallbackURL(tc.base, "ch-1"); got != tc.want {
			t.Errorf("CallbackURL(%q) = %q, want %q", tc.base, got, tc.want)
		}
	}
}
