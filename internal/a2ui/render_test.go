package a2ui_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go.orx.me/apps/butter/internal/a2ui"
	"google.golang.org/adk/v2/session"
)

func cardBatch(surfaceID, text string) a2ui.Batch {
	return a2ui.Batch{SurfaceID: surfaceID, Fallback: text, Messages: []map[string]any{
		{"updateComponents": map[string]any{"components": []any{map[string]any{"id": "root", "component": "Text", "text": text}}}},
	}}
}

func TestCardReservationsShareLimitsAndMonotonicPersistence(t *testing.T) {
	base := session.InMemoryService()
	created, err := base.Create(t.Context(), &session.CreateRequest{AppName: "app", UserID: "u", SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	run := &a2ui.Run{ThreadID: "t", RunID: "r", MessageID: "m"}
	ctx := a2ui.WithRun(t.Context(), run)
	makeEvent := func(batch a2ui.Batch) (a2ui.Result, *session.Event) {
		event := session.NewEvent(ctx, "inv")
		result, err := run.Render(created.Session.State(), batch, "inv", func(card *a2ui.Card) error {
			event.Actions.StateDelta[a2ui.CardKey(card.ID)] = card.StateValue()
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return result, event
	}
	first, event1 := makeEvent(cardBatch("", "one"))
	_, event2 := makeEvent(cardBatch(first.Card.ID, "two"))
	svc := a2ui.PersistCards(base)
	// Parallel forwarding may invert the reservation order.
	for _, event := range []*session.Event{event2, event1} {
		if err := svc.AppendEvent(ctx, created.Session, event); err != nil {
			t.Fatal(err)
		}
	}
	stored, err := base.Get(ctx, &session.GetRequest{AppName: "app", UserID: "u", SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	cards := a2ui.LiveCards(stored.Session.State())
	if len(cards) != 1 || cards[0].Revision != 2 {
		t.Fatalf("stored cards=%+v", cards)
	}
	if len(event1.Actions.StateDelta) != 0 {
		t.Fatal("stale delta was not removed from the yielded event")
	}
	// A sibling's stale session state still shares the same run's quota.
	for i := 1; i < a2ui.DefaultLimits.MaxSurfaces; i++ {
		makeEvent(cardBatch("", fmt.Sprintf("card %d", i)))
	}
	if _, err := run.Render(created.Session.State(), cardBatch("", "excess"), "inv", func(*a2ui.Card) error { return nil }); err == nil {
		t.Fatal("parallel reservations bypassed the surface limit")
	}
}

func TestFailedStageDoesNotReserveACard(t *testing.T) {
	base := session.InMemoryService()
	created, err := base.Create(t.Context(), &session.CreateRequest{AppName: "a", UserID: "u", SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	run := &a2ui.Run{}
	failure := errors.New("stage failed")
	if _, err := run.Render(created.Session.State(), cardBatch("", "one"), "inv", func(*a2ui.Card) error { return failure }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	for range a2ui.DefaultLimits.MaxSurfaces {
		if _, err := run.Render(created.Session.State(), cardBatch("", "ok"), "inv", func(*a2ui.Card) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
}

// Guard the adapter's plain-context no-op behavior for unrelated entry points.
func TestCardPersistenceWithoutUIRunIsTransparent(t *testing.T) {
	base := session.InMemoryService()
	created, err := base.Create(t.Context(), &session.CreateRequest{AppName: "a", UserID: "u", SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	event := session.NewEvent(context.Background(), "inv")
	event.Actions.StateDelta["ordinary"] = "ok"
	if err := a2ui.PersistCards(base).AppendEvent(t.Context(), created.Session, event); err != nil {
		t.Fatal(err)
	}
	stored, err := base.Get(t.Context(), &session.GetRequest{AppName: "a", UserID: "u", SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := stored.Session.State().Get("ordinary"); v != "ok" {
		t.Fatalf("state=%v", v)
	}
}
