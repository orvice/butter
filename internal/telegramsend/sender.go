// Package telegramsend is the single outbound path to Telegram (issue #264).
//
// Every proactive and interactive send — Cron delivery, Notify Groups,
// Dashboard test messages, agent replies — goes through Sender.Send with a
// Destination ID. There is deliberately no exported way to send to a raw
// chat ID: raw addressing is what let Bot Tokens and chat IDs spread through
// Cron jobs and Notify Groups in the first place, and it is what allows a
// reply to escape a Forum Topic into the group's general conversation.
//
// The one runtime exception is the transport-level `/where` command, which
// must answer at addresses no Destination covers. It lives behind
// SendRaw, in this package, so the exception has exactly one call site.
package telegramsend

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrDestinationNotFound means no such Destination exists in the workspace.
	ErrDestinationNotFound = errors.New("telegram destination not found")
	// ErrDestinationUnavailable means the Destination or its Channel is not
	// currently allowed to send. Callers surface it rather than skipping:
	// a silently dropped alert is worse than a failed one.
	ErrDestinationUnavailable = errors.New("telegram destination is not available for outbound delivery")
	// ErrNotConfigured means the sender is missing a dependency.
	ErrNotConfigured = errors.New("telegram sender is not configured")
)

// defaultMaxAttempts bounds delivery retries. Only Telegram's own
// `retry_after` drives the wait; Butter does not add proactive rate limiting.
const defaultMaxAttempts = 3

// Message is one outbound message body.
type Message struct {
	// Text is standard Markdown as produced by agents and cron summaries. The
	// sender converts it to Telegram MarkdownV2 and falls back to plain text
	// if Telegram rejects the formatting.
	Text string
	// ReplyToMessageID quotes an inbound message. Empty sends standalone.
	// Forum Topic targeting is independent of this and always applied.
	ReplyToMessageID string
	// DisableNotification sends silently.
	DisableNotification bool
}

// Result reports what was delivered.
type Result struct {
	// MessageIDs are the Telegram message IDs produced, in delivery order.
	MessageIDs []string
	// PlainTextFallback is true when MarkdownV2 was rejected and the text was
	// resent verbatim.
	PlainTextFallback bool
}

// Sleeper lets tests run retry paths without real delays.
type Sleeper func(ctx context.Context, d time.Duration) error

// Sender resolves Destinations and delivers messages.
type Sender struct {
	repo       telegramrepo.Repository
	keyring    *secretbox.Keyring
	botFactory telegramapi.Factory
	sleep      Sleeper
	// maxAttempts bounds delivery retries per message.
	maxAttempts int
}

func New(repo telegramrepo.Repository, keyring *secretbox.Keyring, factory telegramapi.Factory) *Sender {
	if factory == nil {
		factory = telegramapi.NewFactory()
	}
	return &Sender{
		repo:        repo,
		keyring:     keyring,
		botFactory:  factory,
		sleep:       sleepContext,
		maxAttempts: defaultMaxAttempts,
	}
}

// SetSleeper overrides the retry delay. Used by tests.
func (s *Sender) SetSleeper(sleep Sleeper) {
	if sleep != nil {
		s.sleep = sleep
	}
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *Sender) ready() error {
	if s == nil || s.repo == nil || s.keyring == nil {
		return ErrNotConfigured
	}
	return nil
}

// Resolved is a Destination together with everything needed to reach it.
type Resolved struct {
	Destination *agentsv1.TelegramDestination
	Channel     *agentsv1.TelegramChannel
	Client      telegramapi.Client
}

// Resolve loads the current Destination, Channel, and credential and checks
// that outbound delivery is allowed. Configuration is read on every send —
// there is no cross-request cache — so a rotated credential or a disabled
// Destination takes effect immediately.
func (s *Sender) Resolve(ctx context.Context, workspaceID, destinationID string) (*Resolved, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	dest, err := s.repo.GetDestination(ctx, workspaceID, destinationID)
	if err != nil {
		if errors.Is(err, telegramrepo.ErrNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrDestinationNotFound, destinationID)
		}
		return nil, err
	}
	if !dest.GetOutboundEnabled() {
		return nil, fmt.Errorf("%w: destination %s has outbound delivery disabled",
			ErrDestinationUnavailable, destinationID)
	}
	channel, err := s.repo.GetChannel(ctx, workspaceID, dest.GetChannelId())
	if err != nil {
		if errors.Is(err, telegramrepo.ErrNotFound) {
			return nil, fmt.Errorf("%w: channel %s is missing", ErrDestinationUnavailable, dest.GetChannelId())
		}
		return nil, err
	}
	if !channel.GetOutboundEnabled() {
		return nil, fmt.Errorf("%w: channel %s has outbound delivery disabled",
			ErrDestinationUnavailable, channel.GetKey())
	}
	client, err := s.clientFor(ctx, workspaceID, channel.GetId())
	if err != nil {
		return nil, err
	}
	return &Resolved{Destination: dest, Channel: channel, Client: client}, nil
}

// clientFor decrypts the Channel's Bot Token and builds a client. The client
// is not cached: a credential rotation must take effect on the next send.
func (s *Sender) clientFor(ctx context.Context, workspaceID, channelID string) (telegramapi.Client, error) {
	cred, err := s.repo.GetChannelCredential(ctx, workspaceID, channelID)
	if err != nil {
		if errors.Is(err, telegramrepo.ErrNoCredential) {
			return nil, fmt.Errorf("%w: channel %s has no bot token", ErrDestinationUnavailable, channelID)
		}
		return nil, err
	}
	token, err := s.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
	if err != nil {
		return nil, fmt.Errorf("decrypt bot token for channel %s: %w", channelID, err)
	}
	return s.botFactory(string(token)), nil
}

// Send delivers one message to a Destination.
//
// On success it records outbound evidence on the Destination: an unverified
// address becomes verified, and the timestamp is refreshed. On failure it
// records a sanitized error without touching the configured address —
// verification is evidence, not configuration.
func (s *Sender) Send(ctx context.Context, workspaceID, destinationID string, msg Message) (*Result, error) {
	resolved, err := s.Resolve(ctx, workspaceID, destinationID)
	if err != nil {
		return nil, err
	}
	result, sendErr := s.deliver(ctx, resolved, msg)
	s.recordOutcome(ctx, workspaceID, resolved.Destination, sendErr)
	if sendErr != nil {
		return nil, sendErr
	}
	return result, nil
}

// SendToDestination satisfies the notify package's TelegramDelivery seam, so
// Notify Groups and Cron deliveries reach Telegram through the same path as
// everything else.
func (s *Sender) SendToDestination(ctx context.Context, workspaceID, destinationID, text string) error {
	_, err := s.Send(ctx, workspaceID, destinationID, Message{Text: text})
	return err
}

// SendResolved delivers to an already-resolved Destination. Callers that
// resolve once and send several segments use this so one message does not pay
// for repeated credential decryption.
func (s *Sender) SendResolved(ctx context.Context, workspaceID string, resolved *Resolved, msg Message) (*Result, error) {
	result, sendErr := s.deliver(ctx, resolved, msg)
	s.recordOutcome(ctx, workspaceID, resolved.Destination, sendErr)
	if sendErr != nil {
		return nil, sendErr
	}
	return result, nil
}

func (s *Sender) deliver(ctx context.Context, resolved *Resolved, msg Message) (*Result, error) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil, errors.New("message text is empty")
	}

	params := telegramapi.SendMessageParams{
		ChatID: resolved.Destination.GetChatId(),
		// Always carry the Destination's Topic. A Destination never falls
		// back from a Topic to the group's general chat.
		MessageThreadID:     resolved.Destination.GetMessageThreadId(),
		Text:                ToTelegramMarkdownV2(text),
		ParseMode:           telegramapi.ParseModeMarkdownV2,
		ReplyToMessageID:    msg.ReplyToMessageID,
		DisableNotification: msg.DisableNotification,
	}

	sent, err := s.sendWithRetry(ctx, resolved.Client, params)
	if err == nil {
		return &Result{MessageIDs: []string{telegramapi.FormatMessageID(sent.ID)}}, nil
	}
	if !isFormattingRejection(err) {
		return nil, err
	}

	// Telegram rejected the formatting, not the address. Resend verbatim so a
	// markup edge case cannot drop a response entirely.
	log.FromContext(ctx).Warn("telegram rejected MarkdownV2, retrying as plain text",
		"destination_id", resolved.Destination.GetId(), "err", err)
	params.Text = text
	params.ParseMode = telegramapi.ParseModeNone
	sent, err = s.sendWithRetry(ctx, resolved.Client, params)
	if err != nil {
		return nil, err
	}
	return &Result{
		MessageIDs:        []string{telegramapi.FormatMessageID(sent.ID)},
		PlainTextFallback: true,
	}, nil
}

// sendWithRetry honors Telegram's own `retry_after` and gives up after
// maxAttempts. Any other error is returned immediately: retrying a rejected
// address or a revoked token only delays the report.
func (s *Sender) sendWithRetry(ctx context.Context, client telegramapi.Client, params telegramapi.SendMessageParams) (telegramapi.Message, error) {
	var lastErr error
	for attempt := 1; attempt <= s.maxAttempts; attempt++ {
		sent, err := client.SendMessage(ctx, params)
		if err == nil {
			return sent, nil
		}
		lastErr = err
		wait, ok := telegramapi.RetryAfter(err)
		if !ok || attempt == s.maxAttempts {
			return telegramapi.Message{}, err
		}
		if sleepErr := s.sleep(ctx, wait); sleepErr != nil {
			return telegramapi.Message{}, sleepErr
		}
	}
	return telegramapi.Message{}, lastErr
}

// isFormattingRejection reports whether Telegram refused the markup rather
// than the request. Telegram signals this as a 400 whose description mentions
// the entity parsing failure.
func isFormattingRejection(err error) bool {
	apiErr, ok := errors.AsType[*telegramapi.APIError](err)
	if !ok || apiErr.Code != 400 {
		return false
	}
	description := strings.ToLower(apiErr.Description)
	return strings.Contains(description, "parse entities") ||
		strings.Contains(description, "entity") ||
		strings.Contains(description, "can't parse")
}

// recordOutcome persists delivery evidence. It never fails the send: the
// message already reached Telegram, and losing the bookkeeping is strictly
// better than reporting a delivered message as failed.
func (s *Sender) recordOutcome(ctx context.Context, workspaceID string, dest *agentsv1.TelegramDestination, sendErr error) {
	verification := dest.GetVerification()
	next := &agentsv1.TelegramDestinationVerification{
		Verified:       verification.GetVerified(),
		VerifiedAt:     verification.GetVerifiedAt(),
		LastInboundAt:  verification.GetLastInboundAt(),
		LastOutboundAt: verification.GetLastOutboundAt(),
	}
	now := timestamppb.New(time.Now().UTC())
	if sendErr == nil {
		next.LastOutboundAt = now
		next.LastOutboundError = ""
		if !next.Verified {
			next.Verified = true
			next.VerifiedAt = now
		}
	} else {
		next.LastOutboundError = sanitizeError(sendErr)
	}

	updated := cloneWithVerification(dest, next)
	if _, err := s.repo.UpdateDestination(ctx, workspaceID, updated, dest.GetRevision()); err != nil {
		// A concurrent configuration edit wins the revision race. The next
		// send re-reads and records again.
		log.FromContext(ctx).Debug("could not record telegram delivery outcome",
			"destination_id", dest.GetId(), "err", err)
	}
}

func cloneWithVerification(dest *agentsv1.TelegramDestination, verification *agentsv1.TelegramDestinationVerification) *agentsv1.TelegramDestination {
	return &agentsv1.TelegramDestination{
		Id:              dest.GetId(),
		Key:             dest.GetKey(),
		Name:            dest.GetName(),
		ChannelId:       dest.GetChannelId(),
		ChatId:          dest.GetChatId(),
		MessageThreadId: dest.GetMessageThreadId(),
		InboundEnabled:  dest.GetInboundEnabled(),
		OutboundEnabled: dest.GetOutboundEnabled(),
		Config:          dest.GetConfig(),
		Verification:    verification,
		WorkspaceId:     dest.GetWorkspaceId(),
	}
}

// sanitizeError keeps the operator-visible reason and drops anything that
// could carry credential material.
func sanitizeError(err error) string {
	apiErr, ok := errors.AsType[*telegramapi.APIError](err)
	if ok {
		return fmt.Sprintf("telegram error %d: %s", apiErr.Code, apiErr.Description)
	}
	return err.Error()
}
