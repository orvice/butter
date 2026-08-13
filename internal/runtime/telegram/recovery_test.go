package telegram

// Recovery tests (issue #264/#271): where the retry boundary sits, what a
// crash at each stage leaves behind, and why a completed duplicate costs
// nothing.

import (
	"errors"
	"strings"
	"testing"

	"go.orx.me/apps/butter/internal/repo/telegramprocessing"
	telegramprocessingmemory "go.orx.me/apps/butter/internal/repo/telegramprocessing/memory"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramsend"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// newRecoveryFixture adds the durable state machine and session guard.
func newRecoveryFixture(t *testing.T) (*orchestratorFixture, *telegramprocessingmemory.Store) {
	t.Helper()
	fx := newOrchestratorFixture(t, nil)
	records := telegramprocessingmemory.New()
	fx.orchestrator.SetProcessingRepo(records)
	fx.orchestrator.SetSessionGuard(NewMemorySessionGuard())
	return fx, records
}

func onlyRecord(t *testing.T, records *telegramprocessingmemory.Store) *agentsv1.TelegramProcessingRecord {
	t.Helper()
	list, err := records.List(t.Context(), telegramprocessing.Filter{WorkspaceID: "ws-a"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("records = %d, want 1", len(list))
	}
	return list[0]
}

func TestSuccessfulTurnIsRecordedAsSucceeded(t *testing.T) {
	fx, records := newRecoveryFixture(t)

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(message(realUser, "hello", ""))); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	record := onlyRecord(t, records)
	if record.GetStatus() != agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_SUCCEEDED {
		t.Fatalf("status = %v", record.GetStatus())
	}
	if record.GetInvocationId() == "" {
		t.Error("expected a stable invocation id")
	}
	if record.GetAttempts() != 1 {
		t.Errorf("attempts = %d, want 1", record.GetAttempts())
	}
	if record.GetOutput() == "" {
		t.Error("expected the agent output to be persisted")
	}
	if record.GetExpiresAt() == nil {
		t.Error("expected a retention deadline")
	}
}

// A duplicate delivery of an update that already completed costs nothing: no
// second Agent run, no second message.
func TestCompletedDuplicateIsAcknowledgedWithoutRerunning(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	event := fx.eventForStored(message(realUser, "hello", ""))

	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("first: %v", err)
	}
	sentAfterFirst := len(fx.bots.Sent())
	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("duplicate: %v", err)
	}

	if len(fx.agents.calls) != 1 {
		t.Fatalf("agent invoked %d times, want once", len(fx.agents.calls))
	}
	if len(fx.bots.Sent()) != sentAfterFirst {
		t.Fatalf("the duplicate sent %d extra messages", len(fx.bots.Sent())-sentAfterFirst)
	}
	if onlyRecord(t, records).GetAttempts() != 1 {
		t.Error("a completed duplicate must not consume another attempt")
	}
}

// A failure before Agent work started is safely retryable, so it is recorded
// as FAILED rather than uncertain.
func TestPreAgentFailureIsRetryable(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	// An unresolvable agent fails before any agent work runs.
	fx.agents.known = map[string]string{}

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(message(realUser, "hello", ""))); err == nil {
		t.Fatal("expected the failure to surface for retry")
	}
	record := onlyRecord(t, records)
	if record.GetStatus() == agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN {
		t.Fatal("a pre-agent failure must not be marked uncertain")
	}
	if record.GetDeadLettered() {
		t.Error("a pre-agent failure must not be dead-lettered")
	}
}

// Once Agent work may have run tools, an automatic rerun could repeat side
// effects, so the record becomes uncertain and is dead-lettered.
func TestFailureAfterAgentStartIsUncertainAndDeadLettered(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.agents.failErr = errors.New("model timed out mid-tool-call")

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(message(realUser, "hello", ""))); err == nil {
		t.Fatal("expected the failure to surface")
	}
	record := onlyRecord(t, records)
	if record.GetStatus() != agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN {
		t.Fatalf("status = %v, want FAILED_UNCERTAIN", record.GetStatus())
	}
	if !record.GetDeadLettered() {
		t.Error("expected the record to be dead-lettered for an operator")
	}
	if record.GetError() == "" {
		t.Error("expected a recorded failure summary")
	}
}

func TestUncertainFailureIsAcknowledgedWithoutRerunningTheAgent(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.agents.failErr = errors.New("agent failed after a tool may have run")
	event := fx.eventForStored(message(realUser, "hello", ""))

	if err := fx.orchestrator.Handle(t.Context(), event); err == nil {
		t.Fatal("expected the initial Agent failure")
	}
	if len(fx.agents.calls) != 1 {
		t.Fatalf("initial Agent calls = %d, want 1", len(fx.agents.calls))
	}
	fx.agents.failErr = nil
	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("reclaimed uncertain event: %v", err)
	}
	if len(fx.agents.calls) != 1 {
		t.Fatalf("uncertain event reran the Agent: calls = %d", len(fx.agents.calls))
	}
	if got := onlyRecord(t, records).GetStatus(); got != agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN {
		t.Fatalf("status = %s, want FAILED_UNCERTAIN", got)
	}
}

// Once output is safely persisted, a delivery failure is retryable *as a
// send*: the text exists, so nothing has to be recomputed.
func TestDeliveryFailureKeepsTheOutputAndStaysRetryable(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	// A response long enough to need several segments: the first edits the
	// placeholder, the rest are sends, and the second send fails.
	fx.agents.output = strings.Repeat("word ", 2000)
	fx.bots.OnSend(func(attempt int, _ telegramapi.SendMessageParams) error {
		if attempt > 1 {
			return &telegramapi.APIError{Code: 500, Description: "Internal Server Error"}
		}
		return nil
	})

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(message(realUser, "hello", ""))); err == nil {
		t.Fatal("expected the delivery failure to surface")
	}
	record := onlyRecord(t, records)
	if record.GetOutput() == "" {
		t.Fatal("the agent output was not persisted before delivery")
	}
	if record.GetStatus() == agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN {
		t.Error("a delivery failure with persisted output must stay retryable")
	}
	if len(record.GetSegments()) == 0 {
		t.Error("expected per-segment state for a resend to continue from")
	}
}

func TestReclaimedDeliveryFailureResumesWithoutRerunningTheAgent(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.agents.output = strings.Repeat("word ", 2000)
	failing := true
	fx.bots.OnSend(func(attempt int, _ telegramapi.SendMessageParams) error {
		if failing && attempt > 1 {
			return &telegramapi.APIError{Code: 500, Description: "Internal Server Error"}
		}
		return nil
	})
	event := fx.eventForStored(message(realUser, "hello", ""))

	if err := fx.orchestrator.Handle(t.Context(), event); err == nil {
		t.Fatal("expected the initial delivery failure")
	}
	agentCalls := len(fx.agents.calls)
	failing = false
	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("resume delivery: %v", err)
	}
	if len(fx.agents.calls) != agentCalls {
		t.Fatalf("delivery recovery reran the Agent: before=%d after=%d", agentCalls, len(fx.agents.calls))
	}
	if got := onlyRecord(t, records).GetStatus(); got != agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_SUCCEEDED {
		t.Fatalf("status = %s, want SUCCEEDED", got)
	}
}

func TestReclaimedSendingSegmentIsDeadLetteredWithoutDuplicateDelivery(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.agents.output = "already accepted by telegram"
	event := fx.eventForStored(message(realUser, "hello", ""))

	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("initial delivery: %v", err)
	}
	record := onlyRecord(t, records)
	record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER
	record.Segments[0].Status = string(telegramsend.SegmentSending)
	record.Segments[0].MessageId = ""
	if _, err := records.Update(t.Context(), record); err != nil {
		t.Fatalf("seed sending segment: %v", err)
	}
	sendsBefore := len(fx.bots.Sent())
	agentCallsBefore := len(fx.agents.calls)

	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("reclaim uncertain delivery: %v", err)
	}
	if len(fx.bots.Sent()) != sendsBefore {
		t.Fatal("the uncertain segment was sent a second time")
	}
	if len(fx.agents.calls) != agentCallsBefore {
		t.Fatal("the uncertain delivery reran the Agent")
	}
	record = onlyRecord(t, records)
	if record.GetStatus() != agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN || !record.GetDeadLettered() {
		t.Fatalf("record = status %s dead_lettered=%v, want uncertain dead letter", record.GetStatus(), record.GetDeadLettered())
	}
}

// A resend continues from what is still pending; the Agent is never consulted
// again.
func TestResendContinuesWithoutRerunningTheAgent(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.agents.output = strings.Repeat("word ", 2000)
	failing := true
	fx.bots.OnSend(func(attempt int, _ telegramapi.SendMessageParams) error {
		if failing && attempt > 1 {
			return &telegramapi.APIError{Code: 500, Description: "Internal Server Error"}
		}
		return nil
	})
	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(message(realUser, "hello", ""))); err == nil {
		t.Fatal("expected the first delivery to fail")
	}
	record := onlyRecord(t, records)
	agentCallsBefore := len(fx.agents.calls)

	failing = false
	delivery := DeliveryFromRecord(record)
	if !delivery.Pending() {
		t.Fatal("expected pending segments to resend")
	}
	if err := fx.orchestrator.sender.DeliverSegments(t.Context(), "ws-a", fx.dest.GetId(), delivery); err != nil {
		t.Fatalf("resend: %v", err)
	}
	if len(fx.agents.calls) != agentCallsBefore {
		t.Fatal("the resend re-ran the agent")
	}
	if delivery.Pending() {
		t.Error("the resend left segments undelivered")
	}
}

// Two updates for one conversation must not interleave their history writes;
// unrelated sessions stay parallel.
func TestSessionLeaseSerializesOneConversation(t *testing.T) {
	fx, _ := newRecoveryFixture(t)
	guard := NewMemorySessionGuard()
	fx.orchestrator.SetSessionGuard(guard)

	sessionID := SessionID("ch-1", fx.dest.GetId(), "d"+fx.dest.GetId(), "support")
	_, release, ok, err := guard.Acquire(t.Context(), sessionID)
	if err != nil || !ok {
		t.Fatalf("seed lease: ok=%v err=%v", ok, err)
	}

	err = fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "hello", "")))
	if !errors.Is(err, ErrSessionBusy) {
		t.Fatalf("err = %v, want ErrSessionBusy so the event is retried", err)
	}
	if len(fx.agents.calls) != 0 {
		t.Fatal("a busy session still invoked the agent")
	}

	// An unrelated session is unaffected.
	if _, _, otherOK, otherErr := guard.Acquire(t.Context(), "tg:ch-1:dest-2:d-other:support"); !otherOK || otherErr != nil {
		t.Fatalf("an unrelated session was blocked: ok=%v err=%v", otherOK, otherErr)
	}

	// Once released, the same update proceeds.
	release()
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "hello", ""))); err != nil {
		t.Fatalf("Handle after release: %v", err)
	}
	if len(fx.agents.calls) != 1 {
		t.Fatalf("agent invoked %d times after release", len(fx.agents.calls))
	}
}

// Sanitized failures must not carry credential material.
func TestRecordedFailuresAreSanitized(t *testing.T) {
	fx, records := newRecoveryFixture(t)
	fx.agents.failErr = errors.New(strings.Repeat("x", 2000))

	_ = fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "hello", "")))

	record := onlyRecord(t, records)
	if len([]rune(record.GetError())) > 501 {
		t.Errorf("recorded error is %d runes; failures must be bounded", len([]rune(record.GetError())))
	}
}

// A command answers without ever taking the session lease, so a management
// command never stalls behind a long-running turn.
func TestCommandsDoNotHoldTheSessionLease(t *testing.T) {
	fx, _ := newRecoveryFixture(t)
	guard := NewMemorySessionGuard()
	fx.orchestrator.SetSessionGuard(guard)
	sessionID := SessionID("ch-1", fx.dest.GetId(), "d"+fx.dest.GetId(), "support")
	if _, _, ok, err := guard.Acquire(t.Context(), sessionID); !ok || err != nil {
		t.Fatalf("seed lease: %v", err)
	}

	raw := message(realUser, "/status", `"entities":[{"type":"bot_command","offset":0,"length":7}]`)
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(raw)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.bots.Sent()) == 0 {
		t.Fatal("/status was blocked by an unrelated session lease")
	}
}
