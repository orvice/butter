package interrupt

import (
	"testing"

	"google.golang.org/adk/v2/session"
)

func TestPending_NilAndEmpty(t *testing.T) {
	if got := Pending(nil); got != nil {
		t.Errorf("Pending(nil) = %+v, want nil", got)
	}
	if got := Pending(&fakeSession{}); got != nil {
		t.Errorf("Pending(empty) = %+v, want nil", got)
	}
}

func TestPending_SingleUnanswered(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{
		askEvent("i1", "Approve this draft?"),
	}}
	got := Pending(sess)
	if len(got) != 1 {
		t.Fatalf("len(Pending) = %d, want 1", len(got))
	}
	if got[0].InterruptID != "i1" {
		t.Errorf("InterruptID = %q, want %q", got[0].InterruptID, "i1")
	}
	if got[0].Question != "Approve this draft?" {
		t.Errorf("Question = %q, want %q", got[0].Question, "Approve this draft?")
	}
}

// TestPending_AnsweredFilteredOut: a call with a matching response is no
// longer pending.
func TestPending_AnsweredFilteredOut(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{
		askEvent("i1", "Q1?"),
		answerEvent("i1"),
	}}
	if got := Pending(sess); got != nil {
		t.Errorf("Pending = %+v, want nil (i1 answered)", got)
	}
}

// TestPending_FIFOMultiplePending: two parallel branches each pause; the
// unanswered interrupts come back oldest-first in event order.
func TestPending_FIFOMultiplePending(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{
		askEvent("i1", "Question A?"),
		askEvent("i2", "Question B?"),
	}}
	got := Pending(sess)
	if len(got) != 2 {
		t.Fatalf("len(Pending) = %d, want 2", len(got))
	}
	if got[0].InterruptID != "i1" || got[1].InterruptID != "i2" {
		t.Errorf("order = [%q, %q], want [i1, i2] (FIFO)", got[0].InterruptID, got[1].InterruptID)
	}
}

// TestPending_FIFOWithOneAnswered: oldest answered, the remaining pending one
// is still returned.
func TestPending_FIFOWithOneAnswered(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{
		askEvent("i1", "Question A?"),
		askEvent("i2", "Question B?"),
		answerEvent("i1"),
	}}
	got := Pending(sess)
	if len(got) != 1 {
		t.Fatalf("len(Pending) = %d, want 1", len(got))
	}
	if got[0].InterruptID != "i2" {
		t.Errorf("InterruptID = %q, want i2", got[0].InterruptID)
	}
}
