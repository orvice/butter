package linear

// Follow-up tests (ADR-0015 §6, #366): a message sent while the Agent works
// is acknowledged as queued at once, runs right after the current turn, and
// is never lost and never run concurrently with it.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// startBlockedTurn handles a created event in the background with the
// runner held, and waits until the Agent is running.
func startBlockedTurn(t *testing.T, fx *orchestratorFixture) (release func(), done <-chan error) {
	t.Helper()
	fx.runner.block = make(chan struct{})
	fx.runner.started = make(chan struct{}, 8)
	errs := make(chan error, 1)
	go func() { errs <- fx.orch.Handle(t.Context(), fx.event(ActionCreated)) }()
	select {
	case <-fx.runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("the first turn never started")
	}
	var closed bool
	return func() {
		if !closed {
			closed = true
			close(fx.runner.block)
		}
	}, errs
}

func recordFor(t *testing.T, records linearprocessing.Repository, deliveryID string) *agentsv1.LinearProcessingRecord {
	t.Helper()
	list, err := records.List(t.Context(), linearprocessing.Filter{WorkspaceID: "ws-a"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range list {
		if r.GetDeliveryId() == deliveryID {
			return r
		}
	}
	t.Fatalf("no record for delivery %s", deliveryID)
	return nil
}

func TestAMessageForABusySessionIsQueuedAtOnceAndRunsNext(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	release, done := startBlockedTurn(t, fx)

	follow := fx.prompted("are you done?")
	if err := fx.orch.Handle(t.Context(), follow); err != nil {
		t.Fatalf("Handle(follow-up) = %v, want it acknowledged at once", err)
	}
	acts := fx.linear.Activities()
	if last := acts[len(acts)-1]; last.Type != linearapi.ActivityThought || !strings.Contains(last.Body, "Queued") {
		t.Fatalf("last activity = %+v; want the queued thought", last)
	}
	if got := recordFor(t, records, follow.DeliveryID).GetStatus(); got != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_QUEUED {
		t.Fatalf("follow-up status = %v, want QUEUED", got)
	}

	release()
	if err := <-done; err != nil {
		t.Fatalf("first Handle: %v", err)
	}
	turns := fx.runner.turns()
	if len(turns) != 2 || turns[1].text != "are you done?" || turns[1].sessionID != turns[0].sessionID {
		t.Fatalf("turns = %+v; want the follow-up run next in the same session", turns)
	}
	if got := recordFor(t, records, follow.DeliveryID).GetStatus(); got != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED {
		t.Fatalf("follow-up status after its turn = %v, want SUCCEEDED", got)
	}
}

func TestQueuedMessagesRunAsOneTurnInOrder(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	release, done := startBlockedTurn(t, fx)
	var followUps []*Event
	for _, text := range []string{"first", "second", "third"} {
		ev := fx.prompted(text)
		followUps = append(followUps, ev)
		if err := fx.orch.Handle(t.Context(), ev); err != nil {
			t.Fatalf("Handle(%s): %v", text, err)
		}
	}
	release()
	if err := <-done; err != nil {
		t.Fatalf("first Handle: %v", err)
	}
	turns := fx.runner.turns()
	if len(turns) != 2 {
		t.Fatalf("turns = %d, want the blocked turn and one joined follow-up turn", len(turns))
	}
	if turns[1].text != "first\n\nsecond\n\nthird" {
		t.Fatalf("joined input = %q", turns[1].text)
	}
	invocation := ""
	for _, ev := range followUps {
		r := recordFor(t, records, ev.DeliveryID)
		if r.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED {
			t.Fatalf("%s status = %v, want SUCCEEDED", ev.PromptText, r.GetStatus())
		}
		if invocation == "" {
			invocation = r.GetInvocationId()
		} else if r.GetInvocationId() != invocation {
			t.Fatalf("follow-ups ran under different invocations: %q vs %q", r.GetInvocationId(), invocation)
		}
	}
	// One reply answers the joined turn.
	responses := 0
	for _, a := range fx.linear.Activities() {
		if a.Type == linearapi.ActivityResponse {
			responses++
		}
	}
	if responses != 2 {
		t.Fatalf("responses = %d, want one per turn", responses)
	}
}

func TestAfterAHolderCrashTheInterruptedTurnIsSettledAndItsFollowUpsRun(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	created := fx.event(ActionCreated)
	sessionID := SessionID("app-1", created.AgentSessionID, "support")

	// A holder on another Pod took the session and was running the created
	// turn when it died, with a follow-up queued behind it.
	hold, _, err := fx.coord.EnqueueOrAcquire(t.Context(), sessionID, FollowUp{})
	if err != nil || hold == nil {
		t.Fatalf("seed hold = %v, %v", hold, err)
	}
	past := time.Now().Add(-10 * time.Minute)
	dead, _, err := records.Claim(t.Context(), newRecord(created), "dead-pod", past, past.Add(time.Minute))
	if err != nil {
		t.Fatalf("seed record: %v", err)
	}
	dead.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING
	if _, err := records.UpdateClaimed(t.Context(), dead, "dead-pod"); err != nil {
		t.Fatalf("seed status: %v", err)
	}
	follow := fx.prompted("still there?")
	if err := fx.orch.Handle(t.Context(), follow); err != nil {
		t.Fatalf("Handle(follow-up): %v", err)
	}
	hold.Abandon() // the dead Pod's lease expires; the list survives

	// The queue redelivers the dead Pod's event.
	if err := fx.orch.Handle(t.Context(), created); err != nil {
		t.Fatalf("Handle(redelivery): %v", err)
	}
	if got := recordFor(t, records, created.DeliveryID).GetStatus(); got != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN {
		t.Fatalf("interrupted record = %v, want FAILED_UNCERTAIN", got)
	}
	turns := fx.runner.turns()
	if len(turns) != 1 || !strings.HasSuffix(turns[0].text, "still there?") {
		t.Fatalf("turns = %+v; want only the follow-up, never the interrupted turn", turns)
	}
	if got := recordFor(t, records, follow.DeliveryID).GetStatus(); got != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED {
		t.Fatalf("follow-up status = %v, want SUCCEEDED", got)
	}
}

func TestLosingTheSessionLeaseCancelsTheTurn(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	_, done := startBlockedTurn(t, fx)
	fx.coord.Steal(SessionID("app-1", "session-1", "support"))
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Handle = %v, want the turn cancelled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the turn kept running after the lease was lost")
	}
	// The fenced-out turn leaves its record for the reclaim to settle.
	if got := recordFor(t, records, "d-created").GetStatus(); got != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING {
		t.Fatalf("status = %v, want PROCESSING", got)
	}
}
