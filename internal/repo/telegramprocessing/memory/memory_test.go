package memory

import (
	"errors"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/repo/telegramprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func testRecord(updateID int64) *agentsv1.TelegramProcessingRecord {
	return &agentsv1.TelegramProcessingRecord{
		WorkspaceId: "ws-a",
		ChannelId:   "channel-a",
		UpdateId:    updateID,
		Status:      agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_RECEIVED,
	}
}

func TestActiveClaimRejectsConcurrentOwner(t *testing.T) {
	store := New()
	now := time.Now().UTC()
	record, action, err := store.Claim(t.Context(), testRecord(1), "lease-a", now, now.Add(time.Minute))
	if err != nil || action != telegramprocessing.ClaimRunAgent {
		t.Fatalf("first claim: action=%v err=%v", action, err)
	}

	_, _, err = store.Claim(t.Context(), testRecord(1), "lease-b", now.Add(time.Second), now.Add(time.Minute))
	if !errors.Is(err, telegramprocessing.ErrInProgress) {
		t.Fatalf("second claim err = %v, want ErrInProgress", err)
	}
	if _, err := store.UpdateClaimed(t.Context(), record, "lease-b"); !errors.Is(err, telegramprocessing.ErrLeaseLost) {
		t.Fatalf("stale update err = %v, want ErrLeaseLost", err)
	}
}

func TestExpiredProcessingClaimBecomesUncertain(t *testing.T) {
	store := New()
	now := time.Now().UTC()
	record, _, err := store.Claim(t.Context(), testRecord(2), "lease-a", now, now.Add(time.Second))
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_PROCESSING
	if _, err := store.UpdateClaimed(t.Context(), record, "lease-a"); err != nil {
		t.Fatalf("mark processing: %v", err)
	}

	reclaimed, action, err := store.Claim(t.Context(), testRecord(2), "lease-b", now.Add(2*time.Second), now.Add(time.Minute))
	if err != nil {
		t.Fatalf("reclaim: %v", err)
	}
	if action != telegramprocessing.ClaimAcknowledge {
		t.Fatalf("action = %v, want acknowledge", action)
	}
	if reclaimed.GetStatus() != agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN || !reclaimed.GetDeadLettered() {
		t.Fatalf("record = status %s dead_lettered=%v", reclaimed.GetStatus(), reclaimed.GetDeadLettered())
	}
}

func TestQueueAndOperatorDeliveryClaimsAreExclusive(t *testing.T) {
	store := New()
	now := time.Now().UTC()
	record, _, err := store.Claim(t.Context(), testRecord(3), "lease-agent", now, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	record.Output = "complete output"
	record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER
	record.Segments = []*agentsv1.TelegramDeliverySegment{{Text: "complete output", Status: "pending"}}
	if _, err := store.UpdateClaimed(t.Context(), record, "lease-agent"); err != nil {
		t.Fatalf("persist output: %v", err)
	}
	if err := store.ReleaseClaim(t.Context(), "ws-a", record.GetId(), "lease-agent"); err != nil {
		t.Fatalf("release agent claim: %v", err)
	}

	if _, err := store.ClaimDelivery(t.Context(), "ws-a", record.GetId(), "lease-operator", now.Add(time.Second), now.Add(time.Minute)); err != nil {
		t.Fatalf("operator claim: %v", err)
	}
	_, _, err = store.Claim(t.Context(), testRecord(3), "lease-queue", now.Add(2*time.Second), now.Add(time.Minute))
	if !errors.Is(err, telegramprocessing.ErrInProgress) {
		t.Fatalf("queue claim err = %v, want ErrInProgress", err)
	}
}
