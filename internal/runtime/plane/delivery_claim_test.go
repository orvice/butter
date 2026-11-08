package plane

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type blockingDeliveryTarget struct {
	mu    sync.Mutex
	calls int
	start chan struct{}
	block chan struct{}
}

func (t *blockingDeliveryTarget) PublishReply(context.Context, *Event, string) error {
	t.mu.Lock()
	t.calls++
	t.mu.Unlock()
	close(t.start)
	<-t.block
	return nil
}

func TestDurablePublisherClaimsOneConcurrentDelivery(t *testing.T) {
	store := NewMemoryDeliveryStore()
	target := &blockingDeliveryTarget{start: make(chan struct{}), block: make(chan struct{})}
	p := NewDurablePublisher(target, store)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a"}

	first := make(chan error, 1)
	go func() { first <- p.PublishReply(context.Background(), e, "reply") }()
	<-target.start
	second := p.PublishReply(context.Background(), e, "reply")
	if !errors.Is(second, ErrDeliveryInProgress) {
		t.Fatalf("second delivery error = %v, want ErrDeliveryInProgress", second)
	}
	close(target.block)
	if err := <-first; err != nil {
		t.Fatalf("first delivery error = %v", err)
	}
	target.mu.Lock()
	calls := target.calls
	target.mu.Unlock()
	if calls != 1 {
		t.Fatalf("target calls = %d, want 1", calls)
	}
}

func TestDurablePublisherDoesNotAutomaticallyResendUnknown(t *testing.T) {
	store := NewMemoryDeliveryStore()
	p := NewDurablePublisher(unknownPublisher{}, store)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", BotUserID: "owner"}
	for i := 0; i < maxAutomaticDeliveryAttempts; i++ {
		_ = p.PublishReply(context.Background(), e, "reply")
	}
	if err := p.PublishReply(context.Background(), e, "reply"); !errors.Is(err, ErrPermanentDelivery) {
		t.Fatalf("unknown resend error = %v, want ErrPermanentDelivery", err)
	}
	row, err := store.Get(context.Background(), DeliveryKey(e))
	if err != nil {
		t.Fatal(err)
	}
	if row.Attempts != maxAutomaticDeliveryAttempts || row.Status != DeliveryUnknown {
		t.Fatalf("row after unknown resend = %+v", row)
	}
}
