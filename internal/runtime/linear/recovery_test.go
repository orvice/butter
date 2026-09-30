package linear

// Retry-boundary tests (ADR-0009 applied to Linear, #365): a delivery is
// processed at least once, and an Agent that may have run is never run
// again automatically.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/linearapi/lineartest"
	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	linearprocessingmemory "go.orx.me/apps/butter/internal/repo/linearprocessing/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// flakyTokens fails the first n token lookups with a transient error.
type flakyTokens struct {
	TokenProvider
	mu       sync.Mutex
	failures int
	calls    int
}

func (f *flakyTokens) AccessToken(ctx context.Context, workspaceID, installationID string) (string, error) {
	f.mu.Lock()
	f.calls++
	fail := f.calls <= f.failures
	f.mu.Unlock()
	if fail {
		return "", errors.New("mongo blip")
	}
	return f.TokenProvider.AccessToken(ctx, workspaceID, installationID)
}

func newRecoveryFixture(t *testing.T) (*orchestratorFixture, *linearprocessingmemory.Store) {
	t.Helper()
	fx := newOrchestratorFixture(t, nil)
	records := linearprocessingmemory.New()
	fx.orch.SetProcessingRepo(records)
	fx.orch.preAgentBackoff = time.Millisecond
	return fx, records
}

func onlyRecord(t *testing.T, records *linearprocessingmemory.Store) *agentsv1.LinearProcessingRecord {
	t.Helper()
	list, err := records.List(t.Context(), linearprocessing.Filter{WorkspaceID: "ws-a"})
	if err != nil || len(list) != 1 {
		t.Fatalf("records = %d, %v; want exactly one", len(list), err)
	}
	return list[0]
}

func TestASuccessfulTurnIsRecordedAsSucceeded(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.handle(t, fx.event(ActionCreated))
	record := onlyRecord(t, records)
	if record.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED || !record.GetDelivered() {
		t.Fatalf("record = %+v; want SUCCEEDED and delivered", record)
	}
	if record.GetOutput() != "Fixed the login bug." || record.GetOutputType() != linearapi.ActivityResponse {
		t.Fatalf("persisted reply = %q (%s)", record.GetOutput(), record.GetOutputType())
	}
	if record.GetAgentSessionId() != "session-1" || record.GetPromptingUserId() != linearUserOne || record.GetIssueIdentifier() != "ENG-1" {
		t.Fatalf("record identity = %+v", record)
	}
}

func TestACompletedRedeliveryIsAcknowledgedWithoutRerunning(t *testing.T) {
	fx, _ := newRecoveryFixture(t)
	ev := fx.event(ActionCreated)
	fx.handle(t, ev)
	fx.handle(t, ev)
	if n := len(fx.runner.turns()); n != 1 {
		t.Fatalf("turns = %d, want 1", n)
	}
	if n := len(fx.linear.Activities()); n != 2 {
		t.Fatalf("activities = %d, want the first delivery's two only", n)
	}
}

func TestAPreAgentFailureRetries(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	flaky := &flakyTokens{TokenProvider: fx.orch.tokens, failures: 2}
	fx.orch.tokens = flaky
	fx.handle(t, fx.event(ActionCreated))
	if flaky.calls != 3 || len(fx.runner.turns()) != 1 {
		t.Fatalf("token calls = %d, turns = %d; want 3 attempts then one turn", flaky.calls, len(fx.runner.turns()))
	}
	if onlyRecord(t, records).GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED {
		t.Fatal("the retried delivery did not succeed")
	}

	// A pre-Agent failure that outlasts the retries is left for the queue
	// to redeliver, with nothing run.
	always := &flakyTokens{TokenProvider: fx.orch.tokens, failures: 100}
	fx.orch.tokens = always
	if err := fx.orch.Handle(t.Context(), fx.prompted("again")); err == nil {
		t.Fatal("Handle succeeded while every pre-Agent attempt failed")
	}
	if len(fx.runner.turns()) != 1 {
		t.Fatal("the Agent ran after pre-Agent failures")
	}
}

func TestAFailureAfterTheAgentStartedIsDeadLetteredAndNeverRerun(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.runner.err = errors.New("tool call blew up after writing to the database; token=sk-ant-api03-abcdefghijklmnopqrstuvwxyz0123")
	ev := fx.event(ActionCreated)
	fx.handle(t, ev)
	record := onlyRecord(t, records)
	if record.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN || !record.GetDeadLettered() {
		t.Fatalf("record = %+v; want FAILED_UNCERTAIN and dead-lettered", record)
	}
	if strings.Contains(record.GetError(), "abcdefghijklmnop") {
		t.Fatalf("recorded error leaked a credential: %q", record.GetError())
	}

	fx.runner.err = nil
	fx.handle(t, ev)
	if n := len(fx.runner.turns()); n != 1 {
		t.Fatalf("turns after redelivery = %d, want the Agent run only once", n)
	}
}

func TestAPostFailureKeepsTheReplyAndARedeliveryPostsItWithoutRerunning(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.linear.OnActivity(func(a lineartest.Activity) *lineartest.Failure {
		if a.Type == linearapi.ActivityResponse {
			return &lineartest.Failure{Status: 502, Message: "linear is down"}
		}
		return nil
	})
	ev := fx.event(ActionCreated)
	fx.handle(t, ev)
	record := onlyRecord(t, records)
	if record.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED || record.GetOutput() == "" || record.GetDelivered() {
		t.Fatalf("record = %+v; want FAILED with the reply kept", record)
	}

	fx.linear.OnActivity(nil)
	fx.handle(t, ev)
	if n := len(fx.runner.turns()); n != 1 {
		t.Fatalf("turns = %d, want 1: the redelivery must post, not rerun", n)
	}
	acts := fx.linear.Activities()
	if last := acts[len(acts)-1]; last.Type != linearapi.ActivityResponse || last.Body != "Fixed the login bug." {
		t.Fatalf("last activity = %+v; want the persisted reply", last)
	}
	if onlyRecord(t, records).GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED {
		t.Fatal("the redelivered reply was not recorded as succeeded")
	}
}

func TestAReclaimedTurnThatWasRunningIsReportedAsCutShort(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	ev := fx.event(ActionCreated)
	// A worker crashed mid-turn: its record is PROCESSING with an expired
	// lease.
	past := time.Now().Add(-10 * time.Minute)
	record, _, err := records.Claim(t.Context(), newRecord(ev), "dead-worker", past, past.Add(time.Minute))
	if err != nil {
		t.Fatalf("seed claim: %v", err)
	}
	record.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING
	if _, err := records.UpdateClaimed(t.Context(), record, "dead-worker"); err != nil {
		t.Fatalf("seed update: %v", err)
	}

	fx.handle(t, ev)
	if len(fx.runner.turns()) != 0 {
		t.Fatal("the interrupted turn was rerun")
	}
	acts := fx.linear.Activities()
	if len(acts) != 1 || acts[0].Type != linearapi.ActivityError || !strings.Contains(acts[0].Body, "cut short") {
		t.Fatalf("activities = %+v; want one cut-short error", acts)
	}
	stored := onlyRecord(t, records)
	if stored.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN || !stored.GetDeadLettered() {
		t.Fatalf("record = %+v; want FAILED_UNCERTAIN", stored)
	}
}

func TestResendPostsThePersistedReplyWithoutInvokingTheAgent(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.linear.OnActivity(func(a lineartest.Activity) *lineartest.Failure {
		if a.Type == linearapi.ActivityResponse {
			return &lineartest.Failure{Status: 502}
		}
		return nil
	})
	fx.handle(t, fx.event(ActionCreated))
	failed := onlyRecord(t, records)
	fx.linear.OnActivity(nil)

	resent, err := fx.orch.Resend(t.Context(), "ws-a", failed.GetId())
	if err != nil {
		t.Fatalf("Resend: %v", err)
	}
	if resent.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED || !resent.GetDelivered() {
		t.Fatalf("resent = %+v; want SUCCEEDED", resent)
	}
	if len(fx.runner.turns()) != 1 {
		t.Fatal("Resend invoked the Agent")
	}
	if _, err := fx.orch.Resend(t.Context(), "ws-a", failed.GetId()); !errors.Is(err, ErrNotResendable) {
		t.Fatalf("second Resend = %v, want ErrNotResendable", err)
	}
}

func TestAnswersWithoutAnAgentAreRecordedToo(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.runner.known = map[string]string{}
	fx.handle(t, fx.event(ActionCreated))
	record := onlyRecord(t, records)
	if record.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED || record.GetOutputType() != linearapi.ActivityError {
		t.Fatalf("record = %+v; want the error answer recorded as delivered", record)
	}
}
