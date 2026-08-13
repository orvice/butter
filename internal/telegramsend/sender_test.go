package telegramsend

// Contract tests for the unified Telegram sender (issue #264/#266): Topic
// targeting, Markdown fallback, retry_after, verification bookkeeping, and
// the availability rules that must fail loudly rather than silently skip.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramapi/telegramtest"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	botToken  = "111111:bot-token"
	workspace = "ws-a"
)

type fixture struct {
	sender *Sender
	repo   *telegrammemory.Store
	bots   *telegramtest.Fake
	slept  []time.Duration
}

// newFixture builds a workspace with one outbound-enabled Channel and one
// Topic Destination.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	repo := telegrammemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	ciphertext, keyID, err := keyring.Encrypt(t.Context(), []byte(botToken))
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if _, err := repo.CreateChannel(t.Context(), workspace, &agentsv1.TelegramChannel{
		Id: "ch-1", Key: "main", BotId: "111111", OutboundEnabled: true,
	}, telegramrepo.ChannelCredentials{
		BotToken: telegramrepo.Credential{Ciphertext: ciphertext, KeyID: keyID},
	}); err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if _, err := repo.CreateDestination(t.Context(), workspace, &agentsv1.TelegramDestination{
		Id: "dest-1", Key: "incidents", ChannelId: "ch-1",
		ChatId: "-1001234567890", MessageThreadId: "42", OutboundEnabled: true,
	}); err != nil {
		t.Fatalf("create destination: %v", err)
	}

	bots := telegramtest.NewFake().Register(botToken, telegramtest.Identity(111111, "mainbot"))
	fx := &fixture{repo: repo, bots: bots}
	fx.sender = New(repo, keyring, bots.Factory())
	fx.sender.SetSleeper(func(_ context.Context, d time.Duration) error {
		fx.slept = append(fx.slept, d)
		return nil
	})
	return fx
}

func (fx *fixture) destination(t *testing.T) *agentsv1.TelegramDestination {
	t.Helper()
	dest, err := fx.repo.GetDestination(t.Context(), workspace, "dest-1")
	if err != nil {
		t.Fatalf("GetDestination: %v", err)
	}
	return dest
}

// A Topic Destination must always carry its thread ID. Falling back to the
// group's general chat would leak a reply out of the topic it belongs to.
func TestSendAlwaysTargetsTheDestinationTopic(t *testing.T) {
	fx := newFixture(t)

	if _, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "hello"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	sent, ok := fx.bots.LastSent()
	if !ok {
		t.Fatal("nothing was sent")
	}
	if sent.Params.ChatID != "-1001234567890" {
		t.Errorf("chat_id = %q", sent.Params.ChatID)
	}
	if sent.Params.MessageThreadID != "42" {
		t.Errorf("message_thread_id = %q, want the destination topic", sent.Params.MessageThreadID)
	}
	if sent.Token != botToken {
		t.Errorf("sent with token %q, want the channel credential", sent.Token)
	}
}

// A non-Topic Destination must omit the field entirely rather than send 0,
// which Telegram rejects.
func TestNonTopicDestinationOmitsTheThreadID(t *testing.T) {
	fx := newFixture(t)
	if _, err := fx.repo.CreateDestination(t.Context(), workspace, &agentsv1.TelegramDestination{
		Id: "dest-2", Key: "general", ChannelId: "ch-1",
		ChatId: "-1001234567890", OutboundEnabled: true,
	}); err != nil {
		t.Fatalf("create destination: %v", err)
	}

	if _, err := fx.sender.Send(t.Context(), workspace, "dest-2", Message{Text: "hello"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	sent, _ := fx.bots.LastSent()
	if sent.Params.MessageThreadID != "" {
		t.Errorf("message_thread_id = %q, want it omitted", sent.Params.MessageThreadID)
	}
}

func TestSendConvertsMarkdownAndFallsBackToPlainText(t *testing.T) {
	fx := newFixture(t)
	// Reject the first attempt the way Telegram rejects bad markup.
	fx.bots.OnSend(func(attempt int, _ telegramapi.SendMessageParams) error {
		if attempt == 1 {
			return telegramtest.MarkdownRejection()
		}
		return nil
	})

	result, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "a-b **bold**"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !result.PlainTextFallback {
		t.Error("expected the fallback to be reported")
	}
	sends := fx.bots.Sent()
	if len(sends) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(sends))
	}
	if sends[0].Params.ParseMode != telegramapi.ParseModeMarkdownV2 {
		t.Errorf("first attempt parse_mode = %q, want MarkdownV2", sends[0].Params.ParseMode)
	}
	if sends[1].Params.ParseMode != telegramapi.ParseModeNone {
		t.Errorf("fallback parse_mode = %q, want none", sends[1].Params.ParseMode)
	}
	if sends[1].Params.Text != "a-b **bold**" {
		t.Errorf("fallback text = %q, want the original text verbatim", sends[1].Params.Text)
	}
	// The topic is preserved on the fallback too.
	if sends[1].Params.MessageThreadID != "42" {
		t.Errorf("fallback lost the topic: %q", sends[1].Params.MessageThreadID)
	}
}

func TestSendHonorsRetryAfter(t *testing.T) {
	fx := newFixture(t)
	fx.bots.OnSend(func(attempt int, _ telegramapi.SendMessageParams) error {
		if attempt == 1 {
			return telegramtest.RateLimited(3)
		}
		return nil
	})

	if _, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "hello"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(fx.slept) != 1 || fx.slept[0] != 3*time.Second {
		t.Fatalf("waits = %v, want one 3s wait from retry_after", fx.slept)
	}
}

// Butter adds no proactive rate limiting, so a persistently rate-limited
// address gives up rather than retrying forever.
func TestSendGivesUpAfterBoundedRetries(t *testing.T) {
	fx := newFixture(t)
	fx.bots.OnSend(func(int, telegramapi.SendMessageParams) error {
		return telegramtest.RateLimited(1)
	})

	if _, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "hello"}); err == nil {
		t.Fatal("expected the send to fail")
	}
	if got := len(fx.bots.Sent()); got != defaultMaxAttempts {
		t.Errorf("attempts = %d, want %d", got, defaultMaxAttempts)
	}
}

func TestSuccessfulSendVerifiesTheDestination(t *testing.T) {
	fx := newFixture(t)
	if fx.destination(t).GetVerification().GetVerified() {
		t.Fatal("a new destination must start unverified")
	}

	if _, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "hello"}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	verification := fx.destination(t).GetVerification()
	if !verification.GetVerified() {
		t.Error("a successful send must verify the destination")
	}
	if verification.GetLastOutboundAt() == nil {
		t.Error("expected last_outbound_at to be recorded")
	}
	if verification.GetLastOutboundError() != "" {
		t.Errorf("last_outbound_error = %q, want empty", verification.GetLastOutboundError())
	}
}

// A failure records a derived status but must not rewrite the configured
// address.
func TestFailedSendRecordsAnErrorWithoutChangingTheAddress(t *testing.T) {
	fx := newFixture(t)
	fx.bots.OnSend(func(int, telegramapi.SendMessageParams) error {
		return &telegramapi.APIError{Code: 403, Description: "Forbidden: bot was blocked by the user"}
	})

	if _, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "hello"}); err == nil {
		t.Fatal("expected the send to fail")
	}

	dest := fx.destination(t)
	if dest.GetChatId() != "-1001234567890" || dest.GetMessageThreadId() != "42" {
		t.Errorf("the address changed to chat %q thread %q", dest.GetChatId(), dest.GetMessageThreadId())
	}
	if !strings.Contains(dest.GetVerification().GetLastOutboundError(), "Forbidden") {
		t.Errorf("last_outbound_error = %q", dest.GetVerification().GetLastOutboundError())
	}
	if dest.GetVerification().GetVerified() {
		t.Error("a failed send must not verify the destination")
	}
}

func TestUnavailableDestinationsFailExplicitly(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, fx *fixture) string
	}{
		{
			name:  "missing",
			setup: func(*testing.T, *fixture) string { return "no-such-destination" },
		},
		{
			name: "outbound disabled",
			setup: func(t *testing.T, fx *fixture) string {
				dest := fx.destination(t)
				dest.OutboundEnabled = false
				if _, err := fx.repo.UpdateDestination(t.Context(), workspace, dest, dest.GetRevision()); err != nil {
					t.Fatalf("disable destination: %v", err)
				}
				return "dest-1"
			},
		},
		{
			name: "channel outbound disabled",
			setup: func(t *testing.T, fx *fixture) string {
				channel, err := fx.repo.GetChannel(t.Context(), workspace, "ch-1")
				if err != nil {
					t.Fatalf("GetChannel: %v", err)
				}
				channel.OutboundEnabled = false
				if _, err := fx.repo.UpdateChannel(t.Context(), workspace, channel, channel.GetRevision()); err != nil {
					t.Fatalf("disable channel: %v", err)
				}
				return "dest-1"
			},
		},
		{
			name: "credential cleared",
			setup: func(t *testing.T, fx *fixture) string {
				if err := fx.repo.SetChannelCredential(t.Context(), workspace, "ch-1", telegramrepo.Credential{}); err != nil {
					t.Fatalf("clear credential: %v", err)
				}
				return "dest-1"
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fx := newFixture(t)
			id := tc.setup(t, fx)
			_, err := fx.sender.Send(t.Context(), workspace, id, Message{Text: "hello"})
			if err == nil {
				t.Fatal("expected an explicit failure, not a silent skip")
			}
			if !errors.Is(err, ErrDestinationUnavailable) && !errors.Is(err, ErrDestinationNotFound) {
				t.Fatalf("err = %v, want an availability error", err)
			}
			if len(fx.bots.Sent()) != 0 {
				t.Error("nothing may be sent when the destination is unavailable")
			}
		})
	}
}

// A Destination in another workspace must not be reachable, even by ID.
func TestSendIsWorkspaceScoped(t *testing.T) {
	fx := newFixture(t)

	_, err := fx.sender.Send(t.Context(), "ws-other", "dest-1", Message{Text: "hello"})
	if !errors.Is(err, ErrDestinationNotFound) {
		t.Fatalf("err = %v, want ErrDestinationNotFound", err)
	}
}

// A rotated credential must take effect on the very next send: the sender
// deliberately keeps no client cache.
func TestSendUsesTheCurrentCredential(t *testing.T) {
	fx := newFixture(t)
	if _, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "first"}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	const rotated = "111111:rotated-token"
	fx.bots.Register(rotated, telegramtest.Identity(111111, "mainbot"))
	ciphertext, keyID, err := fx.sender.keyring.Encrypt(t.Context(), []byte(rotated))
	if err != nil {
		t.Fatalf("encrypt rotated token: %v", err)
	}
	if err := fx.repo.SetChannelCredential(t.Context(), workspace, "ch-1",
		telegramrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}); err != nil {
		t.Fatalf("rotate credential: %v", err)
	}

	if _, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "second"}); err != nil {
		t.Fatalf("Send after rotation: %v", err)
	}
	sent, _ := fx.bots.LastSent()
	if sent.Token != rotated {
		t.Errorf("sent with token %q, want the rotated credential", sent.Token)
	}
}

// `/where` is the only path allowed to answer an address no Destination
// covers, and it never creates one.
func TestSendRawReachesAnUnconfiguredAddress(t *testing.T) {
	fx := newFixture(t)

	if _, err := fx.sender.SendRaw(t.Context(), workspace, "ch-1", "-100999", "7", "channel ch-1"); err != nil {
		t.Fatalf("SendRaw: %v", err)
	}
	sent, _ := fx.bots.LastSent()
	if sent.Params.ChatID != "-100999" || sent.Params.MessageThreadID != "7" {
		t.Errorf("raw send addressed chat %q thread %q", sent.Params.ChatID, sent.Params.MessageThreadID)
	}
	if sent.Params.ParseMode != telegramapi.ParseModeNone {
		t.Errorf("parse_mode = %q; identifiers must not be markdown-escaped", sent.Params.ParseMode)
	}
	if _, err := fx.repo.FindDestinationByAddress(t.Context(), "ch-1", "-100999", "7"); !errors.Is(err, telegramrepo.ErrNotFound) {
		t.Error("a raw send must not create a destination")
	}
}

func TestSendRejectsEmptyText(t *testing.T) {
	fx := newFixture(t)
	if _, err := fx.sender.Send(t.Context(), workspace, "dest-1", Message{Text: "   "}); err == nil {
		t.Fatal("expected an empty message to be rejected")
	}
}
