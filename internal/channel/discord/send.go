package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// SendText sends text directly to a Discord channel, splitting messages at
// Discord's 2000-character message limit.
func SendText(ctx context.Context, botToken, channelID, text string) error {
	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return fmt.Errorf("creating discord session: %w", err)
	}
	for _, chunk := range splitMessage(text, maxDiscordMessageLen) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if _, err := session.ChannelMessageSend(channelID, chunk); err != nil {
			return fmt.Errorf("sending discord message: %w", err)
		}
	}
	return nil
}
