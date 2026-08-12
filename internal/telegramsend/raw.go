package telegramsend

import (
	"context"
	"fmt"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/telegramapi"
)

// SendRaw delivers to a Telegram address that no Destination covers.
//
// This is the single sanctioned exception to Destination-only addressing, and
// it exists for exactly one caller: the transport-level `/where` command,
// which has to answer in chats and Topics an administrator has not configured
// yet — that is the whole point of the command. Everything else must go
// through Send, so that a Bot Token or chat ID can never be re-introduced into
// downstream configuration.
//
// It deliberately takes a Channel ID rather than a credential: the caller
// still cannot supply its own token, and the send is still attributable to a
// managed transport. No verification metadata is recorded, because there is no
// Destination to record it on, and no Destination is created as a side effect.
func (s *Sender) SendRaw(ctx context.Context, workspaceID, channelID, chatID, messageThreadID, text string) (telegramapi.Message, error) {
	if err := s.ready(); err != nil {
		return telegramapi.Message{}, err
	}
	channel, err := s.repo.GetChannel(ctx, workspaceID, channelID)
	if err != nil {
		return telegramapi.Message{}, err
	}
	if !channel.GetOutboundEnabled() {
		return telegramapi.Message{}, fmt.Errorf("%w: channel %s has outbound delivery disabled",
			ErrDestinationUnavailable, channel.GetKey())
	}
	client, err := s.clientFor(ctx, workspaceID, channelID)
	if err != nil {
		return telegramapi.Message{}, err
	}

	log.FromContext(ctx).Info("telegram raw send",
		"audit", "telegram_raw_send", "workspace_id", workspaceID,
		"channel_id", channelID, "chat_id", chatID, "message_thread_id", messageThreadID)

	// `/where` output is identifiers, not agent prose: send it verbatim so a
	// chat ID containing a hyphen cannot be mangled by Markdown escaping.
	return s.sendWithRetry(ctx, client, telegramapi.SendMessageParams{
		ChatID:          chatID,
		MessageThreadID: messageThreadID,
		Text:            text,
		ParseMode:       telegramapi.ParseModeNone,
	})
}
