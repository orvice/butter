package telegramsend

import (
	"strings"

	"go.orx.me/apps/butter/internal/telegramapi"
)

// MaxSegmentUTF16Units is the per-message budget. Telegram counts UTF-16 code
// units for its 4096-unit text limit; the margin absorbs the MarkdownV2
// escaping the converter adds, which can only ever grow the string.
const MaxSegmentUTF16Units = 3800

// SplitMessage breaks text into ordered segments that each fit one Telegram
// message.
//
// Splitting happens on the largest natural boundary that fits — paragraph,
// then line, then word — and only falls back to a hard cut when a single
// token is itself longer than a whole message. That ordering matters because
// a reply cut mid-word reads as corrupted, while one cut between paragraphs
// reads as a continuation.
//
// Cuts are always on rune boundaries: slicing a multi-byte character in half
// produces bytes Telegram rejects outright, turning a long reply into no
// reply.
func SplitMessage(text string) []string {
	// Whitespace-only output is nothing to deliver; sending it would post an
	// empty-looking message Telegram would reject anyway.
	if strings.TrimSpace(text) == "" {
		return nil
	}
	trimmed := strings.TrimRight(text, "\n")
	if telegramapi.UTF16Len(trimmed) <= MaxSegmentUTF16Units {
		return []string{trimmed}
	}

	var segments []string
	remaining := trimmed
	for telegramapi.UTF16Len(remaining) > MaxSegmentUTF16Units {
		head, tail := splitOnce(remaining, MaxSegmentUTF16Units)
		segments = append(segments, head)
		remaining = tail
	}
	if strings.TrimSpace(remaining) != "" {
		segments = append(segments, remaining)
	}
	return segments
}

// splitOnce takes at most limit UTF-16 units off the front, preferring a
// natural boundary.
func splitOnce(text string, limit int) (head, tail string) {
	window, remainder := telegramapi.SplitUTF16(text, limit)

	for _, separator := range []string{"\n\n", "\n", " "} {
		if at := strings.LastIndex(window, separator); at > 0 {
			return strings.TrimRight(window[:at], "\n "),
				strings.TrimLeft(text[at+len(separator):], "\n ")
		}
	}
	// One unbroken token longer than a whole message: cut on a Unicode code
	// point boundary rather than dropping it.
	return window, remainder
}
