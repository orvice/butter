package telegram

// Boundary tests for the authenticated Telegram callback (issue #264/#267):
// secret validation, the enablement gate, and what each failure means for
// Telegram's retry behavior.

import (
	"errors"
	"testing"

	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const webhookSecret = "s3cr3t-callback-token"

type receiverFixture struct {
	receiver *Receiver
	repo     *telegrammemory.Store
	queue    *fakeQueue
}

func newReceiverFixture(t *testing.T) *receiverFixture {
	t.Helper()
	repo := telegrammemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())

	if _, err := repo.CreateChannel(t.Context(), "ws-a", &agentsv1.TelegramChannel{
		Id: "ch-1", Key: "main", BotId: "111111", BotUsername: "opsbot",
		InboundEnabled: true, OutboundEnabled: true,
		ReceiveMode: agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK,
	}, telegramrepo.Credential{Ciphertext: "cipher", KeyID: "k1"}); err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if _, err := repo.CreateDestination(t.Context(), "ws-a", &agentsv1.TelegramDestination{
		Id: "dest-1", Key: "ops", ChannelId: "ch-1", ChatId: "-100", InboundEnabled: true,
		Config: &agentsv1.TelegramDestinationConfig{AgentId: "support"},
	}); err != nil {
		t.Fatalf("create destination: %v", err)
	}

	ciphertext, keyID, err := keyring.Encrypt(t.Context(), []byte(webhookSecret))
	if err != nil {
		t.Fatalf("encrypt secret: %v", err)
	}
	if err := repo.SetWebhookSecret(t.Context(), "ws-a", "ch-1",
		telegramrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}); err != nil {
		t.Fatalf("store secret: %v", err)
	}

	queue := newFakeQueue()
	return &receiverFixture{
		receiver: NewReceiver(repo, keyring, NewRouter(repo, queue)),
		repo:     repo,
		queue:    queue,
	}
}

func TestAuthenticateAcceptsTheStoredSecret(t *testing.T) {
	fx := newReceiverFixture(t)

	channel, err := fx.receiver.Authenticate(t.Context(), "ch-1", webhookSecret)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if channel.Channel.GetId() != "ch-1" {
		t.Errorf("channel = %q", channel.Channel.GetId())
	}
	// The workspace comes off the Channel, never from the caller.
	if channel.Channel.GetWorkspaceId() != "ws-a" {
		t.Errorf("workspace = %q", channel.Channel.GetWorkspaceId())
	}
}

func TestAuthenticateRejectsWrongSecrets(t *testing.T) {
	fx := newReceiverFixture(t)

	for _, presented := range []string{"", "wrong", webhookSecret + "x", webhookSecret[:5]} {
		if _, err := fx.receiver.Authenticate(t.Context(), "ch-1", presented); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("secret %q: err = %v, want ErrUnauthorized", presented, err)
		}
	}
}

func TestAuthenticateReportsUnknownChannels(t *testing.T) {
	fx := newReceiverFixture(t)

	if _, err := fx.receiver.Authenticate(t.Context(), "no-such-channel", webhookSecret); !errors.Is(err, ErrChannelNotFound) {
		t.Fatalf("err = %v, want ErrChannelNotFound", err)
	}
}

// The enablement check runs after the secret so an unauthenticated caller
// cannot probe which Channels are switched on.
func TestAuthenticateRejectsANonReceivingChannel(t *testing.T) {
	fx := newReceiverFixture(t)
	channel, err := fx.repo.GetChannel(t.Context(), "ws-a", "ch-1")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	channel.InboundEnabled = false
	if _, err := fx.repo.UpdateChannel(t.Context(), "ws-a", channel, channel.GetRevision()); err != nil {
		t.Fatalf("disable channel: %v", err)
	}

	if _, err := fx.receiver.Authenticate(t.Context(), "ch-1", webhookSecret); !errors.Is(err, ErrChannelNotReceiving) {
		t.Fatalf("err = %v, want ErrChannelNotReceiving", err)
	}
	// A wrong secret on a disabled channel is still reported as unauthorized.
	if _, err := fx.receiver.Authenticate(t.Context(), "ch-1", "wrong"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized to take precedence", err)
	}
}

// A Long Polling channel must not accept webhook callbacks: its updates come
// from the poller, and accepting both would double-process every message.
func TestAuthenticateRejectsLongPollingChannels(t *testing.T) {
	fx := newReceiverFixture(t)
	channel, err := fx.repo.GetChannel(t.Context(), "ws-a", "ch-1")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	channel.ReceiveMode = agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_LONG_POLLING
	if _, err := fx.repo.UpdateChannel(t.Context(), "ws-a", channel, channel.GetRevision()); err != nil {
		t.Fatalf("switch mode: %v", err)
	}

	if _, err := fx.receiver.Authenticate(t.Context(), "ch-1", webhookSecret); !errors.Is(err, ErrChannelNotReceiving) {
		t.Fatalf("err = %v, want ErrChannelNotReceiving", err)
	}
}

// A Channel with no stored secret has no authentic caller.
func TestAuthenticateRejectsWhenNoSecretIsStored(t *testing.T) {
	fx := newReceiverFixture(t)
	if err := fx.repo.SetWebhookSecret(t.Context(), "ws-a", "ch-1", telegramrepo.Credential{}); err != nil {
		t.Fatalf("clear secret: %v", err)
	}

	if _, err := fx.receiver.Authenticate(t.Context(), "ch-1", webhookSecret); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestDeliverRoutesAnAuthenticatedUpdate(t *testing.T) {
	fx := newReceiverFixture(t)
	channel, err := fx.receiver.Authenticate(t.Context(), "ch-1", webhookSecret)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	decision, err := fx.receiver.Deliver(t.Context(), channel, textUpdate(1, -100, 0, "hello", ""))
	if err != nil {
		t.Fatalf("Deliver: %v", err)
	}
	if decision != DecisionAccepted {
		t.Fatalf("decision = %s", decision)
	}
	if len(fx.queue.accepted) != 1 {
		t.Fatalf("accepted %d events", len(fx.queue.accepted))
	}
}
