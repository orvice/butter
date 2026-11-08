package plane

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

type blockingDeliveryTarget struct {
	mu    sync.Mutex
	calls int
	start chan struct{}
	block chan struct{}
}

type cancelDeliveryTarget struct {
	cancel context.CancelFunc
}

func (t cancelDeliveryTarget) PublishReply(ctx context.Context, _ *Event, _ string) error {
	t.cancel()
	return fmt.Errorf("transport: %w", ctx.Err())
}

type contextCheckingDeliveryStore struct{ *MemoryDeliveryStore }

func (s contextCheckingDeliveryStore) PutOwned(ctx context.Context, row *DeliveryAttempt, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.MemoryDeliveryStore.PutOwned(ctx, row, token)
}

func TestCanceledDeliveryPersistsUnknownOnDetachedContext(t *testing.T) {
	store := contextCheckingDeliveryStore{NewMemoryDeliveryStore()}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	p := NewDurablePublisher(cancelDeliveryTarget{cancel}, store)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a"}
	if err := p.PublishReply(ctx, e, "reply"); !errors.Is(err, context.Canceled) {
		t.Fatalf("publish error=%v", err)
	}
	row, err := store.Get(t.Context(), DeliveryKey(e))
	if err != nil || row.Status != DeliveryUnknown || row.ClaimToken != "" {
		t.Fatalf("canceled delivery record=%+v err=%v", row, err)
	}
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

type frozenDeliveryTarget struct {
	calls   []string
	prompts []string
}

func (t *frozenDeliveryTarget) PublishReply(_ context.Context, e *Event, reply string) error {
	t.calls = append(t.calls, reply)
	t.prompts = append(t.prompts, e.Prompt)
	if len(t.calls) == 1 {
		return retryErr{}
	}
	return nil
}

func TestRetrySendsPersistedReplyInsteadOfNewCallerInput(t *testing.T) {
	store := NewMemoryDeliveryStore()
	target := &frozenDeliveryTarget{}
	p := NewDurablePublisher(target, store)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", Prompt: "original"}
	_ = p.PublishReply(t.Context(), e, "frozen reply")
	e.Prompt = "changed"
	if err := p.PublishReply(t.Context(), e, "different reply"); err != nil {
		t.Fatal(err)
	}
	if target.calls[1] != "frozen reply" {
		t.Fatalf("retry sent %q, want persisted reply", target.calls[1])
	}
	if target.prompts[1] != "original" {
		t.Fatalf("retry used mutated event %q", target.prompts[1])
	}
}

func TestConcurrentOwnerConfirmationCannotOverwriteInFlightClaim(t *testing.T) {
	store := NewMemoryDeliveryStore()
	target := &blockingDeliveryTarget{start: make(chan struct{}), block: make(chan struct{})}
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", BotUserID: "owner"}
	seed := &DeliveryAttempt{Key: DeliveryKey(e), Event: e, Reply: "frozen", Status: DeliveryUnknown, Attempts: 1}
	if err := store.Put(t.Context(), seed); err != nil {
		t.Fatal(err)
	}
	p1 := NewDurablePublisher(target, store)
	p2 := NewDurablePublisher(target, store)
	first := make(chan error, 1)
	go func() { first <- p1.ConfirmAndRetry(t.Context(), seed.Key, "owner") }()
	<-target.start
	t.Cleanup(func() {
		close(target.block)
		if err := <-first; err != nil {
			t.Errorf("first confirmation completion: %v", err)
		}
	})
	if err := p2.ConfirmAndRetry(t.Context(), seed.Key, "owner"); !errors.Is(err, ErrDeliveryInProgress) {
		t.Fatalf("second confirmation=%v, want busy", err)
	}
	row, err := store.Get(t.Context(), seed.Key)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != DeliveryPending || row.ClaimToken == "" {
		t.Errorf("second confirmation corrupted active claim: status=%s token=%q", row.Status, row.ClaimToken)
	}
}
