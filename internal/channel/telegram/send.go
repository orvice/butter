package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const maxTelegramMessageLen = 4096

// SendText sends text to a Telegram chat, splitting messages at Telegram's
// documented 4096-character sendMessage limit.
func SendText(ctx context.Context, botToken, chatID, text string) error {
	parsedChatID, err := strconv.ParseInt(strings.TrimSpace(chatID), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat_id %q: %w", chatID, err)
	}
	b, err := bot.New(botToken)
	if err != nil {
		return fmt.Errorf("creating telegram bot: %w", err)
	}

	for _, chunk := range splitMessage(text, maxTelegramMessageLen) {
		params := &bot.SendMessageParams{
			ChatID:    parsedChatID,
			Text:      markdownToTelegramMarkdownV2(chunk),
			ParseMode: models.ParseModeMarkdown,
		}
		if _, err := b.SendMessage(ctx, params); err != nil {
			params.Text = chunk
			params.ParseMode = ""
			if _, err2 := b.SendMessage(ctx, params); err2 != nil {
				return fmt.Errorf("sending telegram message: %w", err2)
			}
		}
	}
	return nil
}

func splitMessage(text string, maxLen int) []string {
	if maxLen <= 0 {
		return []string{text}
	}
	if runeCount(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for text != "" {
		if runeCount(text) <= maxLen {
			chunks = append(chunks, text)
			break
		}

		splitAt := byteIndexAfterRunes(text, maxLen)
		if idx := strings.LastIndex(text[:splitAt], "\n"); idx > 0 {
			splitAt = idx + 1
		}
		chunks = append(chunks, text[:splitAt])
		text = text[splitAt:]
	}
	return chunks
}

func runeCount(text string) int {
	count := 0
	for range text {
		count++
	}
	return count
}

func byteIndexAfterRunes(text string, maxRunes int) int {
	count := 0
	for idx := range text {
		if count == maxRunes {
			return idx
		}
		count++
	}
	return len(text)
}
