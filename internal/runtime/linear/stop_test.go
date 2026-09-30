package linear

// Stop tests (ADR-0015 §7, #367): Linear's stop button halts the Agent on
// whichever Pod runs it, discards queued messages, and is confirmed once
// nothing more can be posted.

import (
	"strings"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/runtime/linearconn"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func (fx *orchestratorFixture) stopEvent() *Event {
	ev := fx.event(ActionPrompted)
	ev.Stop = true
	ev.DeliveryID = "d-stop"
	return ev
}

// otherPod is a second orchestrator sharing this one's coordination and
// records, as a second Pod sharing Redis and Mongo would.
func (fx *orchestratorFixture) otherPod(t *testing.T) *Orchestrator {
	t.Helper()
	other := NewOrchestrator(fx.repo, fx.runner, fx.orch.tokens, fx.linear.Client())
	other.SetSessionCoordinator(fx.coord)
	other.SetProcessingRepo(fx.orch.processing)
	other.preAgentBackoff = time.Millisecond
	return other
}

func TestAStopHaltsTheTurnOnAnotherPodAndIsConfirmed(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	_, done := startBlockedTurn(t, fx)

	if err := fx.otherPod(t).Handle(t.Context(), fx.stopEvent()); err != nil {
		t.Fatalf("Handle(stop) on the other pod: %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("the stopped turn's Handle = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the running turn was not stopped")
	}
	acts := fx.linear.Activities()
	last := acts[len(acts)-1]
	if last.Type != linearapi.ActivityResponse || !strings.Contains(last.Body, "Stopped") {
		t.Fatalf("last activity = %+v; want the stop confirmation", last)
	}
	for _, a := range acts {
		if a.Type == linearapi.ActivityError {
			t.Fatalf("a stopped turn posted an error: %+v", a)
		}
	}
	if got := recordFor(t, records, "d-created"); got.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_CANCELLED || got.GetDeadLettered() {
		t.Fatalf("stopped turn record = %+v; want CANCELLED, not dead-lettered", got)
	}
	if got := recordFor(t, records, "d-stop").GetStatus(); got != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED {
		t.Fatalf("stop record = %v, want SUCCEEDED", got)
	}

	// The conversation continues after a stop.
	fx.runner.block = nil
	fx.handle(t, fx.prompted("try again"))
	turns := fx.runner.turns()
	if last := turns[len(turns)-1]; !strings.HasSuffix(last.text, "try again") || last.sessionID != turns[0].sessionID {
		t.Fatalf("turn after stop = %+v; want the same session", last)
	}
}

func TestAStopWithNothingRunningSaysSo(t *testing.T) {
	fx, _ := newRecoveryFixture(t)
	fx.handle(t, fx.stopEvent())
	acts := fx.linear.Activities()
	if len(acts) != 1 || acts[0].Type != linearapi.ActivityResponse || !strings.Contains(acts[0].Body, "nothing to stop") {
		t.Fatalf("activities = %+v; want one nothing-to-stop response", acts)
	}
	if len(fx.runner.turns()) != 0 {
		t.Fatal("a stop ran the Agent")
	}
}

func TestAStopDiscardsQueuedMessages(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	release, done := startBlockedTurn(t, fx)
	defer release()
	for _, text := range []string{"one", "two"} {
		if err := fx.orch.Handle(t.Context(), fx.prompted(text)); err != nil {
			t.Fatalf("Handle(%s): %v", text, err)
		}
	}
	if err := fx.otherPod(t).Handle(t.Context(), fx.stopEvent()); err != nil {
		t.Fatalf("Handle(stop): %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("first Handle: %v", err)
	}
	if n := len(fx.runner.turns()); n != 1 {
		t.Fatalf("turns = %d; the discarded messages must never run", n)
	}
	for _, text := range []string{"one", "two"} {
		if got := recordFor(t, records, "d-"+text).GetStatus(); got != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_CANCELLED {
			t.Fatalf("%s status = %v, want CANCELLED", text, got)
		}
	}
}

func TestAStopFromAnUnlistedUserIsRefused(t *testing.T) {
	fx, _ := newRecoveryFixture(t)
	app := fx.app
	app.AllowedUserIds = []string{linearUserOne}
	if _, err := fx.repo.UpdateApp(t.Context(), "ws-a", app, app.GetRevision()); err != nil {
		t.Fatalf("UpdateApp: %v", err)
	}
	_, done := startBlockedTurn(t, fx)
	stop := fx.stopEvent()
	stop.PromptingUserID = linearUserTwo
	if err := fx.otherPod(t).Handle(t.Context(), stop); err != nil {
		t.Fatalf("Handle(stop): %v", err)
	}
	select {
	case <-done:
		t.Fatal("an unlisted user stopped the turn")
	case <-time.After(100 * time.Millisecond):
	}
	close(fx.runner.block)
	<-done
}

// The stop reaches a Pi Agent's box through the bridge's abort path.
func TestAStopAbortsAPiAgentOnItsBox(t *testing.T) {
	fx, _ := newRecoveryFixture(t)
	piClient := &linearPiClient{hold: make(chan struct{}), submitted: make(chan struct{}, 1)}
	agents := piRunner(t, piClient)
	fx.orch.runner = agents
	done := make(chan error, 1)
	go func() { done <- fx.orch.Handle(t.Context(), fx.event(ActionCreated)) }()
	select {
	case <-piClient.submitted:
	case <-time.After(3 * time.Second):
		t.Fatal("the pi turn never started")
	}

	other := NewOrchestrator(fx.repo, agents, linearconn.NewTokenSource(fx.repo, nil), fx.linear.Client())
	other.tokens = fx.orch.tokens
	other.SetSessionCoordinator(fx.coord)
	if err := other.Handle(t.Context(), fx.stopEvent()); err != nil {
		t.Fatalf("Handle(stop): %v", err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the pi turn was not stopped")
	}
	if _, _, aborts := piClient.snapshot(); len(aborts) == 0 {
		t.Fatal("the stop never reached the box's abort")
	}
}
