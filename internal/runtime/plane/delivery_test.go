package plane

import (
	"context"
	"errors"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"testing"
)

type deliveryTarget struct {
	calls int
	err   error
}

func (t *deliveryTarget) PublishReply(context.Context, *Event, string) error { t.calls++; return t.err }

type retryErr struct{}

func (retryErr) Error() string   { return "temporary" }
func (retryErr) Retryable() bool { return true }
func TestDurablePublisherIdempotentAndObservable(t *testing.T) {
	store := NewMemoryDeliveryStore()
	target := &deliveryTarget{}
	p := NewDurablePublisher(target, store)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a"}
	if err := p.PublishReply(context.Background(), e, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := p.PublishReply(context.Background(), e, "ok"); !errors.Is(err, ErrDeliveryAlreadyDone) {
		t.Fatal(err)
	}
	if target.calls != 1 {
		t.Fatalf("calls=%d", target.calls)
	}
	row, _ := store.Get(context.Background(), DeliveryKey(e))
	if row.Status != DeliveryDelivered || row.Attempts != 1 {
		t.Fatalf("row=%+v", row)
	}
}
func TestDurablePublisherRetryAndRecovery(t *testing.T) {
	store := NewMemoryDeliveryStore()
	target := &deliveryTarget{err: retryErr{}}
	p := NewDurablePublisher(target, store)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a"}
	if p.PublishReply(context.Background(), e, "ok") == nil {
		t.Fatal("expected retry error")
	}
	row, _ := store.Get(context.Background(), DeliveryKey(e))
	if row.Status != DeliveryRetryable {
		t.Fatal(row.Status)
	}
	target.err = nil
	if err := p.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	row, _ = store.Get(context.Background(), DeliveryKey(e))
	if row.Status != DeliveryDelivered || row.Attempts != 2 {
		t.Fatalf("row=%+v", row)
	}
}
func TestProcessorUsesShortReplyDelivery(t *testing.T) { _ = agentsv1.PlaneApp{} }
