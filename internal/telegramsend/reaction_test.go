package telegramsend

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramapi/telegramtest"
)

func TestReactUsesDestinationAndBoundsAcknowledgement(t *testing.T) {
	fx := newFixture(t)
	fx.bots.OnReaction(func(ctx context.Context, _ telegramapi.ReactionParams) error {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 3*time.Second {
			t.Error("reaction must have a short deadline")
		}
		return nil
	})
	if err := fx.sender.React(t.Context(), workspace, "dest-1", "9", "👀"); err != nil {
		t.Fatal(err)
	}
	reactions := fx.bots.Reactions()
	if len(reactions) != 1 {
		t.Fatalf("reactions = %+v", reactions)
	}
	got := reactions[0]
	if got.Token != botToken || got.Params.ChatID != "-1001234567890" || got.Params.MessageID != "9" || got.Params.Emoji != "👀" {
		t.Fatalf("wrong reaction address or credential: %+v", got)
	}
	if len(fx.bots.Sent()) != 0 {
		t.Fatal("reaction must not send a message")
	}
}

func TestReactHonorsRetryAfter(t *testing.T) {
	fx := newFixture(t)
	attempts := 0
	fx.bots.OnReaction(func(context.Context, telegramapi.ReactionParams) error {
		attempts++
		if attempts == 1 {
			return telegramtest.RateLimited(1)
		}
		return nil
	})
	if err := fx.sender.React(t.Context(), workspace, "dest-1", "9", "👀"); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 || len(fx.slept) != 1 || fx.slept[0] != time.Second {
		t.Fatalf("attempts = %d, sleeps = %v", attempts, fx.slept)
	}
}

func TestReactDoesNotRetryPermanentRejection(t *testing.T) {
	fx := newFixture(t)
	want := &telegramapi.APIError{Code: 400, Description: "REACTION_INVALID"}
	fx.bots.OnReaction(func(context.Context, telegramapi.ReactionParams) error { return want })
	if err := fx.sender.React(t.Context(), workspace, "dest-1", "9", "👀"); !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
	if len(fx.bots.Reactions()) != 1 {
		t.Fatal("a permanent reaction rejection must not be retried")
	}
}

func TestReactRespectsOutboundDisablement(t *testing.T) {
	fx := newFixture(t)
	dest := fx.destination(t)
	dest.OutboundEnabled = false
	if _, err := fx.repo.UpdateDestination(t.Context(), workspace, dest, dest.GetRevision()); err != nil {
		t.Fatal(err)
	}
	if err := fx.sender.React(t.Context(), workspace, "dest-1", "9", "👀"); !errors.Is(err, ErrDestinationUnavailable) {
		t.Fatalf("err = %v, want destination unavailable", err)
	}
	if len(fx.bots.Reactions()) != 0 {
		t.Fatal("disabled destination must not receive a reaction")
	}
}
