package telegram

import (
	"strings"
	"testing"
)

func TestSplitMessageHonorsTelegramLimit(t *testing.T) {
	text := strings.Repeat("a", maxTelegramMessageLen+120)

	chunks := splitMessage(text, maxTelegramMessageLen)
	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}
	for i, chunk := range chunks {
		if len(chunk) > maxTelegramMessageLen {
			t.Fatalf("chunk %d length = %d, want <= %d", i, len(chunk), maxTelegramMessageLen)
		}
	}
	if got := strings.Join(chunks, ""); got != text {
		t.Fatal("splitMessage lost content")
	}
}

func TestSplitMessagePrefersNewline(t *testing.T) {
	text := strings.Repeat("a", 20) + "\n" + strings.Repeat("b", 20)

	chunks := splitMessage(text, 25)
	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}
	if !strings.HasSuffix(chunks[0], "\n") {
		t.Fatalf("first chunk = %q, want newline suffix", chunks[0])
	}
	if got := strings.Join(chunks, ""); got != text {
		t.Fatal("splitMessage lost content")
	}
}

func TestSplitMessagePreservesMultibyteRunes(t *testing.T) {
	text := strings.Repeat("界", maxTelegramMessageLen+1)

	chunks := splitMessage(text, maxTelegramMessageLen)
	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}
	if got := strings.Join(chunks, ""); got != text {
		t.Fatal("splitMessage lost multibyte content")
	}
	for i, chunk := range chunks {
		if strings.ContainsRune(chunk, '\uFFFD') {
			t.Fatalf("chunk %d contains replacement rune after split", i)
		}
	}
}
