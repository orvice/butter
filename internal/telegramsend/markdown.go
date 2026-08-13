package telegramsend

import (
	"bytes"

	tgmd "github.com/Mad-Pixels/goldmark-tgmd"
)

// ToTelegramMarkdownV2 converts standard Markdown — the dialect agents and
// cron summaries actually produce — into Telegram MarkdownV2.
//
// Conversion is centralized here rather than at each call site so every
// outbound path formats identically, and so the plain-text fallback in
// Sender.deliver is the single place that decides what happens when Telegram
// rejects the result. A conversion failure returns the input unchanged: the
// send then either succeeds as-is or falls back to plain text, which is
// strictly better than dropping the message.
//
// Reference: https://core.telegram.org/bots/api#markdownv2-style
func ToTelegramMarkdownV2(text string) string {
	var buf bytes.Buffer
	if err := tgmd.TGMD().Convert([]byte(text), &buf); err != nil {
		return text
	}
	return buf.String()
}
