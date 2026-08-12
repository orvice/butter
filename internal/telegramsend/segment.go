package telegramsend

import (
	"strings"
	"unicode/utf8"
)

// MaxSegmentRunes is the per-message budget. Telegram's hard limit is 4096
// units for `text`; the margin absorbs the MarkdownV2 escaping the converter
// adds, which can only ever grow the string.
const MaxSegmentRunes = 3800

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
	if utf8.RuneCountInString(trimmed) <= MaxSegmentRunes {
		return []string{trimmed}
	}

	var segments []string
	remaining := trimmed
	for utf8.RuneCountInString(remaining) > MaxSegmentRunes {
		head, tail := splitOnce(remaining, MaxSegmentRunes)
		segments = append(segments, head)
		remaining = tail
	}
	if strings.TrimSpace(remaining) != "" {
		segments = append(segments, remaining)
	}
	return segments
}

// splitOnce takes at most limit runes off the front, preferring a natural
// boundary.
func splitOnce(text string, limit int) (head, tail string) {
	runes := []rune(text)
	window := string(runes[:limit])

	for _, separator := range []string{"\n\n", "\n", " "} {
		if at := strings.LastIndex(window, separator); at > 0 {
			return strings.TrimRight(window[:at], "\n "),
				strings.TrimLeft(text[at+len(separator):], "\n ")
		}
	}
	// One unbroken token longer than a whole message: cut on the rune
	// boundary rather than dropping it.
	return window, string(runes[limit:])
}
