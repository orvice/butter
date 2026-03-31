package telegram

import (
	"bytes"

	tgmd "github.com/Mad-Pixels/goldmark-tgmd"
)

// markdownToTelegramMarkdownV2 converts standard Markdown (as produced by LLMs)
// into Telegram MarkdownV2 format using goldmark-tgmd.
//
// Telegram MarkdownV2 reference: https://core.telegram.org/bots/api#markdownv2-style
func markdownToTelegramMarkdownV2(text string) string {
	var buf bytes.Buffer
	md := tgmd.TGMD()
	if err := md.Convert([]byte(text), &buf); err != nil {
		// If conversion fails, return original text as-is.
		return text
	}
	return buf.String()
}
