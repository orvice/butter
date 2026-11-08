package plane

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExpiredPendingDeliveryBecomesUnknownWithoutResend(t *testing.T) {
	store := NewMemoryDeliveryStore()
	target := &deliveryTarget{}
	p := NewDurablePublisher(target, store)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", BotUserID: "owner"}
	key := DeliveryKey(e)
	old := time.Now().UTC().Add(-time.Minute)
	if err := store.Put(context.Background(), &DeliveryAttempt{
		Key: key, Event: e, Reply: "reply", Status: DeliveryPending,
		Attempts: 1, LeaseUntil: &old,
	}); err != nil {
		t.Fatal(err)
	}
	if err := p.PublishReply(context.Background(), e, "reply"); !errors.Is(err, ErrPermanentDelivery) {
		t.Fatalf("publish error = %v, want ErrPermanentDelivery", err)
	}
	if target.calls != 0 {
		t.Fatalf("target calls = %d, want 0", target.calls)
	}
	row, err := store.Get(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != DeliveryUnknown {
		t.Fatalf("status = %q, want unknown", row.Status)
	}
}
