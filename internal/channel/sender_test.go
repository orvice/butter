package channel

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitDiscordOutboundMessagePreservesUTF8(t *testing.T) {
	text := strings.Repeat("ab你", 4) + "🙂tail"
	chunks := splitDiscordOutboundMessage(text, 7)
	if len(chunks) <= 1 {
		t.Fatalf("chunks len = %d, want multiple chunks", len(chunks))
	}
	var joined strings.Builder
	for i, chunk := range chunks {
		if !utf8.ValidString(chunk) {
			t.Fatalf("chunk %d is invalid UTF-8: %q", i, chunk)
		}
		if len(chunk) > 7 {
			t.Fatalf("chunk %d len = %d, want <= 7", i, len(chunk))
		}
		joined.WriteString(chunk)
	}
	if joined.String() != text {
		t.Fatalf("joined chunks = %q, want %q", joined.String(), text)
	}
}
