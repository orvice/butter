package plane

import (
	"context"
	"errors"
	"testing"
)

type unknownPublisher struct{}
type retryDeliveryError struct{}

func (retryDeliveryError) Error() string   { return "temporary" }
func (retryDeliveryError) Retryable() bool { return true }
func (unknownPublisher) PublishReply(context.Context, *Event, string) error {
	return retryDeliveryError{}
}
func TestDeliveryUnknownRequiresOwnerConfirmationAndCapsAutomaticAttempts(t *testing.T) {
	s := NewMemoryDeliveryStore()
	p := NewDurablePublisher(unknownPublisher{}, s)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", BotUserID: "owner"}
	for i := 0; i < 3; i++ {
		if err := p.PublishReply(context.Background(), e, "reply"); err == nil {
			t.Fatal("expected failure")
		}
	}
	row, _ := s.Get(context.Background(), DeliveryKey(e))
	if row.Status != DeliveryUnknown || row.Attempts != 3 {
		t.Fatalf("row=%+v", row)
	}
	if err := p.ConfirmAndRetry(context.Background(), row.Key, "other"); !errors.Is(err, ErrDeliveryNotOwner) {
		t.Fatalf("err=%v", err)
	}
}
