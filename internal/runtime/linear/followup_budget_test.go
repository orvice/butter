package linear

import (
	"strings"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
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

func TestOversizedFollowUpsAllExecuteInSeparateTurns(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	release, done := startBlockedTurn(t, fx)
	defer release()
	followUps := []*Event{}
	for _, label := range []string{"first:", "second:", "third:"} {
		e := fx.prompted(label + strings.Repeat("x", maxPromptBatchBytes/2))
		e.DeliveryID = label
		followUps = append(followUps, e)
		fx.handle(t, e)
	}
	release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	turns := fx.runner.turns()
	if len(turns) != 4 {
		t.Fatalf("turns=%d, want original turn plus all three budgeted follow-ups", len(turns))
	}
	ids := map[string]bool{}
	for i, e := range followUps {
		if turns[i+1].text != e.PromptText {
			t.Fatalf("turn %d was truncated or reordered", i+1)
		}
		record := recordFor(t, records, e.DeliveryID)
		if record.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED {
			t.Fatalf("follow-up %d status=%v", i, record.GetStatus())
		}
		id := record.GetInvocationId()
		if id == "" || ids[id] {
			t.Fatalf("separate batch reused invocation %q", id)
		}
		ids[id] = true
	}
}
