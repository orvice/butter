package pibox

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/a2ui"
)

func TestHostWaitCancellationAbortsWithoutReceipt(t *testing.T) {
	f := newHostPi()
	f.blockCalls = true
	_, h := hostHarness(t, f)
	ctx, cancel := context.WithTimeout(uiContext(t), 100*time.Millisecond)
	defer cancel()
	if _, err := h.turn(ctx, textContent("show status")); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("turn error = %v", err)
	}
	if f.abortCount() != 1 || len(f.deliveries) != 0 {
		t.Fatalf("aborts=%d receipts=%d", f.abortCount(), len(f.deliveries))
	}
}

func TestHostCardsStopAfterConsecutiveValidationFailures(t *testing.T) {
	f := newHostPi()
	f.arguments = []string{`{"messages":[]}`, `{"messages":[]}`, `{"messages":[]}`, createCardArgs}
	_, h := hostHarness(t, f)
	if _, err := h.turn(uiContext(t), textContent("show status")); err != nil {
		t.Fatal(err)
	}
	if len(f.deliveries) != 4 || f.deliveries[3].GetError() == "" {
		t.Fatal("valid call bypassed the consecutive failure limit")
	}
	if len(a2ui.LiveCards(storedSession(t, h).State())) != 0 {
		t.Fatal("refused call wrote a card")
	}
}

func TestHostToolHistoryUsesSeparateModelAndUserRoles(t *testing.T) {
	f := newHostPi()
	_, h := hostHarness(t, f)
	if _, err := h.turn(uiContext(t), textContent("show status")); err != nil {
		t.Fatal(err)
	}
	calls, results := 0, 0
	events := storedSession(t, h).Events()
	for i := range events.Len() {
		event := events.At(i)
		if event.Content == nil {
			continue
		}
		for _, part := range event.Content.Parts {
			if part.FunctionCall != nil {
				calls++
				if event.Content.Role != "model" {
					t.Fatal("tool call has the wrong role")
				}
			}
			if part.FunctionResponse != nil {
				results++
				if event.Content.Role != "user" {
					t.Fatal("tool result has the wrong role")
				}
			}
		}
	}
	if calls != 1 || results != 1 {
		t.Fatalf("calls=%d results=%d", calls, results)
	}
}

func TestHostCardsDeleteAndKeepRevision(t *testing.T) {
	f := newHostPi()
	f.arguments = append(f.arguments, `{"surface_id":"$SURFACE","messages":[{"deleteSurface":{}}]}`)
	_, h := hostHarness(t, f)
	if _, err := h.turn(uiContext(t), textContent("create then delete")); err != nil {
		t.Fatal(err)
	}
	cards := a2ui.Cards(storedSession(t, h).State())
	if len(cards) != 1 {
		t.Fatalf("cards=%d", len(cards))
	}
	for _, card := range cards {
		if !card.Deleted || card.Revision != 2 {
			t.Fatalf("card=%+v", card)
		}
	}
}
