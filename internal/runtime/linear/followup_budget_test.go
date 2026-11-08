package linear

import (
	"strings"
	"testing"
)

func TestPromptBatchKeepsAcceptedMessagesWholeAndOrdered(t *testing.T) {
	first := strings.Repeat("a", maxPromptBatchBytes-2)
	prompts := []*prompt{{text: first}, {text: "second"}, {text: "third"}}

	batch, remaining := promptBatch(prompts)
	if len(batch) != 1 || len(remaining) != 2 {
		t.Fatalf("batch sizes = %d/%d, want 1/2", len(batch), len(remaining))
	}
	if batch[0].text != first || remaining[0].text != "second" || remaining[1].text != "third" {
		t.Fatalf("prompt order changed: batch=%q remaining=%q,%q", batch[0].text, remaining[0].text, remaining[1].text)
	}

	over := strings.Repeat("x", maxPromptBatchBytes+1)
	batch, remaining = promptBatch([]*prompt{{text: over}, {text: "after"}})
	if len(batch) != 1 || batch[0].text != over || len(remaining) != 1 || remaining[0].text != "after" {
		t.Fatalf("oversized prompt was truncated or dropped: %d/%d", len(batch), len(remaining))
	}
}
