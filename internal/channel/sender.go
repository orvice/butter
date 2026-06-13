package channel

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/go-telegram/bot"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const maxDiscordOutboundMessageLen = 2000

// Sender sends outbound messages through configured AgentChannels.
type Sender struct{}

func NewSender() *Sender {
	return &Sender{}
}

func (s *Sender) Send(ctx context.Context, channel *agentsv1.AgentChannel, chatID, text string) error {
	if channel == nil {
		return errors.New("channel is required")
	}
	if chatID == "" {
		return errors.New("chat_id is required")
	}
	switch channel.GetPlatform() {
	case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM:
		return sendTelegramMessage(ctx, channel, chatID, text)
	case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD:
		return sendDiscordMessage(channel, chatID, text)
	default:
		return fmt.Errorf("unsupported channel platform %s", channel.GetPlatform())
	}
}

func sendTelegramMessage(ctx context.Context, channel *agentsv1.AgentChannel, chatID, text string) error {
	token := channel.GetTelegram().GetBotToken()
	if token == "" {
		return fmt.Errorf("telegram channel %q has empty bot token", channel.GetName())
	}
	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat_id %q: %w", chatID, err)
	}
	b, err := bot.New(token)
	if err != nil {
		return fmt.Errorf("create telegram bot: %w", err)
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: id,
		Text:   text,
	})
	if err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}
	return nil
}

func sendDiscordMessage(channel *agentsv1.AgentChannel, chatID, text string) error {
	token := channel.GetDiscord().GetBotToken()
	if token == "" {
		return fmt.Errorf("discord channel %q has empty bot token", channel.GetName())
	}
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("create discord session: %w", err)
	}
	for _, chunk := range splitDiscordOutboundMessage(text, maxDiscordOutboundMessageLen) {
		if _, err := dg.ChannelMessageSend(chatID, chunk); err != nil {
			return fmt.Errorf("send discord message: %w", err)
		}
	}
	return nil
}

func splitDiscordOutboundMessage(text string, maxLen int) []string {
	if maxLen <= 0 || len(text) <= maxLen {
		return []string{text}
	}
	chunks := make([]string, 0, len(text)/maxLen+1)
	for len(text) > maxLen {
		chunks = append(chunks, text[:maxLen])
		text = text[maxLen:]
	}
	if text != "" {
		chunks = append(chunks, text)
	}
	return chunks
}
