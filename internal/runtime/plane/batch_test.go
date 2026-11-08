package plane

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmem "go.orx.me/apps/butter/internal/repo/invocation/memory"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	processingmem "go.orx.me/apps/butter/internal/repo/planeprocessing/memory"
	"go.orx.me/apps/butter/internal/runtime/memoryhook"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/userinput"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"
)

type batchCall struct {
	input []string
	info  *agentsv1.ContextInfo
}

type batchRunner struct {
	mu      sync.Mutex
	calls   []batchCall
	started chan struct{}
	release chan struct{}
}

type batchPublisher struct {
	mu     sync.Mutex
	events []Event
}

func (p *batchPublisher) PublishReply(_ context.Context, e *Event, _ string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	copy := *e
	copy.Members = nil
	p.events = append(p.events, copy)
	return nil
}

func (r *batchRunner) Run(ctx context.Context, _ string, parts []*genai.Part, _ string, info *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	call := batchCall{info: proto.Clone(info).(*agentsv1.ContextInfo)}
	for _, part := range parts {
		call.input = append(call.input, part.Text)
	}
	r.mu.Lock()
	r.calls = append(r.calls, call)
	first := len(r.calls) == 1
	r.mu.Unlock()
	if first && r.started != nil {
		close(r.started)
		select {
		case <-r.release:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return strings.Repeat("完整答复🙂", 200), nil
}

func (r *batchRunner) snapshot() []batchCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]batchCall(nil), r.calls...)
}

func batchProcessor(q *StreamQueue, r Runner, records *processingmem.Store, owner string) *Processor {
	p := NewProcessorWithDurableDelivery(r, &batchPublisher{}, NewMemoryDeliveryStore())
	p.batches = &redisBatchSource{client: q.client}
	p.control = NewRedisRunControl(q.client)
	p.controlPoll = 10 * time.Millisecond
	p.SetClaimRepository(records, owner)
	p.processing = records
	p.publisher.(*DurablePublisher).observe = p.deliveryProjection
	return p
}

func acceptBatch(t *testing.T, q *StreamQueue, base *Event, texts ...string) []*Event {
	t.Helper()
	events := make([]*Event, len(texts))
	for i, text := range texts {
		e := *base
		e.AgentID = "agent"
		e.ActivityID = base.ActivityID + string(rune('A'+i))
		e.Prompt = text
		e.Members = nil
		if _, err := q.Accept(t.Context(), &e); err != nil {
			t.Fatal(err)
		}
		events[i] = &e
	}
	return events
}

func processingRows(t *testing.T, records *processingmem.Store, e *Event) []*agentsv1.PlaneProcessingRecord {
	t.Helper()
	rows, err := records.List(t.Context(), planeprocessing.Filter{WorkspaceID: e.WorkspaceID})
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestRedisBatchPreservesPromptsActorsAndSharedInvocation(t *testing.T) {
	q, base := planeQueueFixture(t)
	events := []*Event{}
	for i, text := range []string{"第一条\n原始内容🙂", "second prompt", "third prompt"} {
		e := *base
		e.AgentID = "agent"
		e.ActivityID += string(rune('A' + i))
		e.Prompt = text
		if i == 1 {
			e.ActorUserID = "other-user"
		}
		if _, err := q.Accept(t.Context(), &e); err != nil {
			t.Fatal(err)
		}
		events = append(events, &e)
	}
	r := &batchRunner{}
	records := processingmem.New()
	p := batchProcessor(q, r, records, "pod-a")
	w := NewWorker(q, p, &planeTestResolver{app: &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}}, "pod-a")
	entries, err := q.Read(t.Context(), "pod-a", 3, time.Millisecond)
	if err != nil || len(entries) != 3 {
		t.Fatalf("entries=%d err=%v", len(entries), err)
	}
	for _, entry := range entries {
		if err := w.handle(t.Context(), entry); err != nil {
			t.Fatal(err)
		}
	}
	calls := r.snapshot()
	if len(calls) != 1 {
		t.Fatalf("executed %d times for one batch", len(calls))
	}
	call := calls[0]
	if len(call.input) != 6 || call.input[1] != events[0].Prompt || call.input[3] != events[1].Prompt || call.input[5] != events[2].Prompt {
		t.Fatalf("prompt boundaries changed: %v", call.input)
	}
	if call.info.GetMetadata()[memoryhook.DisableCaptureMetadataKey] != "true" || call.info.GetSessionId() != planeSession(base, "agent") {
		t.Fatalf("identity=%v", call.info)
	}
	rows := processingRows(t, records, base)
	if len(rows) != 3 {
		t.Fatalf("rows=%v", rows)
	}
	for _, row := range rows {
		if row.GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED ||
			row.GetInvocationId() != call.info.GetUuid() || len(row.GetBatchActivityIds()) != 3 ||
			(row.GetActorUserId() != "user" && row.GetActorUserId() != "other-user") || row.GetMemoryCaptureDisabledReason() == "" || row.GetAgentId() != "agent" || row.GetExecutionFence() != 1 {
			t.Fatalf("projection=%v", row)
		}
	}
	delivery, err := p.publisher.(*DurablePublisher).store.Get(t.Context(), DeliveryKey(events[0]))
	if err != nil || delivery.Status != DeliveryDelivered || len(delivery.Reply) <= 500 {
		t.Fatalf("complete output not persisted: row=%v err=%v", delivery, err)
	}
}

func TestBatchInputProvenanceAndByteBudget(t *testing.T) {
	first := &Event{ActorUserID: "alice", ActivityID: "a", Prompt: "你好🙂", AcceptedOrder: 1}
	second := &Event{ActorUserID: "bob", ActivityID: "b", Prompt: "回答\n两行", AcceptedOrder: 2}
	parts, _, metadata, err := batchInput([]*Event{first, second})
	if err != nil || parts[1].Text != first.Prompt || parts[3].Text != second.Prompt ||
		metadata[memoryhook.DisableCaptureMetadataKey] != "true" || metadata[memoryhook.PrincipalMetadataKey] != "" ||
		!strings.Contains(metadata["plane_prompts"], `"actor":"bob"`) || metadata["memory_capture_disabled_reason"] == "" {
		t.Fatalf("parts=%v metadata=%v err=%v", parts, metadata, err)
	}
	second.ActorUserID = ""
	_, _, metadata, err = batchInput([]*Event{second})
	if err != nil || metadata[memoryhook.DisableCaptureMetadataKey] != "true" {
		t.Fatalf("unknown actor may capture: %v %v", metadata, err)
	}
	second.Prompt = strings.Repeat("x", userinput.MaxTextBytes+1)
	if _, _, _, err := batchInput([]*Event{second}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("oversized input accepted: %v", err)
	}
}

func TestRedisBatchBudgetSplitsWithoutTruncatingUnicode(t *testing.T) {
	q, base := planeQueueFixture(t)
	text := strings.Repeat("中🙂", 90000)
	events := acceptBatch(t, q, base, text, text+"\nnext")
	r := &batchRunner{}
	records := processingmem.New()
	p := batchProcessor(q, r, records, "pod")
	for _, event := range events {
		if _, err := p.Process(t.Context(), &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}, event); err != nil {
			t.Fatal(err)
		}
	}
	calls := r.snapshot()
	if len(calls) != 2 || calls[0].input[0] != events[0].Prompt || calls[1].input[0] != events[1].Prompt {
		t.Fatal("budget splitting lost, duplicated or truncated an input")
	}
	if calls[0].info.GetSessionId() != calls[1].info.GetSessionId() {
		t.Fatal("successive batches lost conversation identity")
	}
	other := *base
	other.InstallationID += "-other"
	if planeSession(base, "agent") == planeSession(&other, "agent") || planeSession(base, "agent") == planeSession(base, "other-agent") {
		t.Fatal("session scope collision")
	}
}

func TestRedisBatchCrashBoundaryAndRecovery(t *testing.T) {
	q, base := planeQueueFixture(t)
	events := acceptBatch(t, q, base, "first", "second")
	source := &redisBatchSource{client: q.client}
	e := events[0]
	if err := source.prepare(t.Context(), e); err != nil {
		t.Fatal(err)
	}
	control := NewRedisRunControl(q.client)
	if err := control.Begin(t.Context(), e, "reserved-holder"); err != nil {
		t.Fatal(err)
	}
	if q.client.Exists(t.Context(), startedKey(e)).Val() != 0 {
		t.Fatal("reservation crossed the execution boundary")
	}
	// Simulate lease expiry before Start: a new holder can retry this batch.
	q.client.Del(t.Context(), controlKey(e))
	if err := control.Begin(t.Context(), e, "executed-holder"); err != nil {
		t.Fatal(err)
	}
	if err := control.StartBatch(t.Context(), e, "executed-holder"); err != nil {
		t.Fatal(err)
	}
	var consumed *consumedTurn
	if err := source.prepare(t.Context(), events[1]); !errors.Is(err, ErrRunBusy) {
		t.Fatalf("live execution was consumed: %v", err)
	}
	// Now the holder dies after execution started. Every member must recover
	// the same invocation without calling the model or blocking a successor.
	q.client.Del(t.Context(), controlKey(e))
	r := &batchRunner{}
	records := processingmem.New()
	p := batchProcessor(q, r, records, "recovering-pod")
	invocations := invocationmem.New()
	p.invocations = invocations
	if err := invocations.Save(t.Context(), &agentsv1.Invocation{Id: "executed-holder", WorkspaceId: base.WorkspaceID, AppName: "plane", Status: agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING}); err != nil {
		t.Fatal(err)
	}
	w := NewWorker(q, p, &planeTestResolver{app: &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}}, "recovering-pod")
	entries, _ := q.Read(t.Context(), "recovering-pod", 2, time.Millisecond)
	for _, entry := range entries {
		if err := w.handle(t.Context(), entry); err != nil {
			t.Fatal(err)
		}
	}
	if len(r.snapshot()) != 0 {
		t.Fatal("crashed batch executed again")
	}
	for _, row := range processingRows(t, records, base) {
		if row.GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED_UNCERTAIN || row.GetInvocationId() != "executed-holder" {
			t.Fatalf("recovery row=%v", row)
		}
	}
	inv, _ := invocations.Get(t.Context(), base.WorkspaceID, "executed-holder")
	if inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("orphan Invocation still active: %v", inv)
	}
	next := *base
	next.AgentID, next.ActivityID, next.Prompt = "agent", "successor", "continue"
	if _, err := q.Accept(t.Context(), &next); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Process(t.Context(), &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}, &next); err != nil {
		t.Fatal(err)
	}
	if _, err := control.Finish(t.Context(), e, "executed-holder", RunStateSucceeded); !errors.Is(err, ErrRunLeaseLost) {
		t.Fatalf("old holder regained authority: %v", err)
	}
	if err := source.prepare(t.Context(), events[1]); !errors.As(err, &consumed) || consumed.invocation != "executed-holder" {
		t.Fatalf("replaced holder lost its durable result: %v", err)
	}
}

func TestRedisRejectedInputDoesNotBlockSuccessor(t *testing.T) {
	q, base := planeQueueFixture(t)
	events := acceptBatch(t, q, base, strings.Repeat("x", userinput.MaxTextBytes+1), "valid")
	records := processingmem.New()
	r := &batchRunner{}
	p := batchProcessor(q, r, records, "pod")
	w := NewWorker(q, p, &planeTestResolver{app: &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}}, "pod")
	entries, _ := q.Read(t.Context(), "pod", 2, time.Millisecond)
	for _, entry := range entries {
		if err := w.handle(t.Context(), entry); err != nil {
			t.Fatal(err)
		}
	}
	if len(r.snapshot()) != 1 || r.snapshot()[0].input[0] != events[1].Prompt {
		t.Fatal("rejected prompt blocked or contaminated the next turn")
	}
}

func waitBatch(t *testing.T, condition func() bool, message string) {
	t.Helper()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if condition() {
			return
		}
		select {
		case <-deadline.C:
			t.Fatal(message)
		case <-tick.C:
		}
	}
}

func TestRedisTwoWorkersDrainQueuedBatchWithoutBlockingOtherRuns(t *testing.T) {
	q, base := planeQueueFixture(t)
	acceptBatch(t, q, base, "first")
	r := &batchRunner{started: make(chan struct{}), release: make(chan struct{})}
	var once sync.Once
	release := func() { once.Do(func() { close(r.release) }) }
	defer release()
	records := processingmem.New()
	store, target := NewMemoryDeliveryStore(), &batchPublisher{}
	app := &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}
	workers := []*Worker{}
	for _, owner := range []string{"pod-a", "pod-b"} {
		p := batchProcessor(q, r, records, owner)
		p.publisher = NewDurablePublisher(target, store)
		p.publisher.(*DurablePublisher).observe = p.deliveryProjection
		w := NewWorker(q, p, &planeTestResolver{app: app}, owner)
		w.block, w.heartbeat = 10*time.Millisecond, 20*time.Millisecond
		workers = append(workers, w)
	}
	if err := workers[0].Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer workers[0].Stop()
	awaitPlane(t, r.started, "first batch invocation")
	next := *base
	next.ActivityID = "followup"
	queued := acceptBatch(t, q, &next, "second", "third")
	independent := *base
	independent.AgentRunID, independent.ActivityID = "independent-run", "independent-prompt"
	acceptBatch(t, q, &independent, "parallel work")
	if err := workers[1].Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer workers[1].Stop()
	waitBatch(t, func() bool {
		rows := processingRows(t, records, base)
		queuedCount, parallelDone := 0, false
		for _, row := range rows {
			if row.GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_QUEUED && row.GetQueueConfirmationStatus() == "delivered" {
				queuedCount++
			}
			if row.GetRunId() == independent.AgentRunID && row.GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED {
				parallelDone = true
			}
		}
		return queuedCount == 2 && parallelDone
	}, "queued prompts or independent Run were starved")
	release()
	waitBatch(t, func() bool {
		rows := processingRows(t, records, base)
		if len(rows) != 4 {
			return false
		}
		for _, row := range rows {
			if row.GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED {
				return false
			}
		}
		return true
	}, "own pending messages were not retried after holder release")
	workers[0].Stop()
	workers[1].Stop()
	calls := r.snapshot()
	if len(calls) != 3 {
		t.Fatalf("expected first + parallel + joined followups, got %d", len(calls))
	}
	joined := 0
	for _, call := range calls {
		if len(call.input) == 4 && call.input[1] == queued[0].Prompt && call.input[3] == queued[1].Prompt {
			joined++
		}
	}
	if joined != 1 {
		t.Fatal("followups did not execute in a single FIFO batch")
	}
	target.mu.Lock()
	defer target.mu.Unlock()
	thoughts, responses := 0, 0
	for _, event := range target.events {
		if event.ActivityType == ActivityThought {
			thoughts++
		} else {
			responses++
		}
	}
	if thoughts != 2 || responses != 3 {
		t.Fatalf("queue confirmations masqueraded as results or duplicated: thoughts=%d responses=%d", thoughts, responses)
	}
}

type failTerminalProjection struct {
	*processingmem.Store
	fail bool
}

func (s *failTerminalProjection) Project(ctx context.Context, r *agentsv1.PlaneProcessingRecord) error {
	if s.fail && planeprocessing.Terminal(r.GetStatus()) {
		s.fail = false
		return errors.New("temporary projection failure")
	}
	return s.Store.Project(ctx, r)
}

func TestRedisSettledBatchRepairsProjectionWithoutRerunning(t *testing.T) {
	q, base := planeQueueFixture(t)
	acceptBatch(t, q, base, "first", "second")
	records := processingmem.New()
	r := &batchRunner{}
	p := batchProcessor(q, r, records, "pod")
	p.processing = &failTerminalProjection{Store: records, fail: true}
	w := NewWorker(q, p, &planeTestResolver{app: &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}}, "pod")
	entries, _ := q.Read(t.Context(), "pod", 2, time.Millisecond)
	for _, entry := range entries {
		if err := w.handle(t.Context(), entry); err != nil {
			t.Fatal(err)
		}
	}
	if len(r.snapshot()) != 1 {
		t.Fatal("projection repair reran the Agent")
	}
	for _, row := range processingRows(t, records, base) {
		if row.GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED || len(row.GetBatchActivityIds()) != 2 {
			t.Fatalf("settled batch projection not repaired: %v", row)
		}
	}
	// The reply was staged before the failed terminal write. A fresh
	// publisher can deliver it without access to the original invocation.
	durable := p.publisher.(*DurablePublisher)
	restarted := NewDurablePublisher(durable.target, durable.store)
	restarted.observe = p.deliveryProjection
	if err := restarted.Recover(t.Context()); err != nil {
		t.Fatal(err)
	}
	for _, row := range processingRows(t, records, base) {
		if row.GetReplyDeliveryStatus() != "delivered" {
			t.Fatalf("recovered delivery did not update every batch member: %v", row)
		}
	}
}

type workflowBatchRunner struct {
	*batchRunner
	registry *runner.Service
}

func (r *workflowBatchRunner) SupportsPromptBatching(workspace, agent string) bool {
	return r.registry.SupportsPromptBatching(workspace, agent)
}

func TestRedisWorkflowAnswersRemainSeparateExactTurns(t *testing.T) {
	q, base := planeQueueFixture(t)
	events := acceptBatch(t, q, base, `{"decision":"approve"}`, "answer to the next question\n🙂")
	workflow := &agentsv1.Agent{WorkspaceId: base.WorkspaceID, AgentId: "workflow", Name: "workflow", Type: agentsv1.AgentType_AGENT_TYPE_WORKFLOW}
	registry := runner.NewServiceForTestAgents(
		&agentsv1.Agent{WorkspaceId: base.WorkspaceID, AgentId: "agent", Name: "composite", Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL, ChildAgentIds: []string{"workflow"}},
		workflow,
		&agentsv1.Agent{WorkspaceId: base.WorkspaceID, AgentId: "llm", Name: "llm", Type: agentsv1.AgentType_AGENT_TYPE_LLM},
	)
	if !registry.SupportsPromptBatching(base.WorkspaceID, "llm") || registry.SupportsPromptBatching("other-workspace", "llm") {
		t.Fatal("batch capability escaped the workspace or rejected a plain LLM")
	}
	r := &workflowBatchRunner{batchRunner: &batchRunner{}, registry: registry}
	p := batchProcessor(q, r, processingmem.New(), "pod")
	for _, event := range events {
		if _, err := p.Process(t.Context(), &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}, event); err != nil {
			t.Fatal(err)
		}
	}
	calls := r.snapshot()
	if len(calls) != 2 {
		t.Fatalf("Workflow answers merged into %d turns", len(calls))
	}
	for i, call := range calls {
		if len(call.input) != 1 || call.input[0] != events[i].Prompt {
			t.Fatalf("Workflow FIFO reply was decorated or joined: %v", call.input)
		}
		if call.info.GetUserId() != planeServiceUser(base) || call.info.GetMetadata()[memoryhook.PrincipalMetadataKey] != "plane:user" {
			t.Fatalf("installation service identity or actor provenance lost: %v", call.info)
		}
	}
}

func TestRedisBatchRechecksEachFrozenAgentAndActor(t *testing.T) {
	for _, scenario := range []string{"disabled", "rebound", "missing-actor", "bot-actor", "revoked-actor"} {
		t.Run(scenario, func(t *testing.T) {
			q, base := planeQueueFixture(t)
			switch scenario {
			case "missing-actor":
				base.ActorUserID = ""
			case "bot-actor":
				base.ActorUserID, base.BotUserID = "bot", "bot"
			}
			event := acceptBatch(t, q, base, "do not run")[0]
			app := &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}
			switch scenario {
			case "disabled":
				app.InboundEnabled = false
			case "rebound":
				app.AgentId = "new-agent"
			case "revoked-actor":
				app.AllowedUserIds = []string{"other-user"}
			}
			r, records := &batchRunner{}, processingmem.New()
			p := batchProcessor(q, r, records, "pod")
			w := NewWorker(q, p, &planeTestResolver{app: app}, "pod")
			entries, _ := q.Read(t.Context(), "pod", 1, time.Millisecond)
			if err := w.handle(t.Context(), entries[0]); err != nil {
				t.Fatal(err)
			}
			if len(r.snapshot()) != 0 || q.client.Exists(t.Context(), startedKey(event)).Val() != 0 {
				t.Fatal("stale or unauthorized prompt crossed the execution boundary")
			}
			rows := processingRows(t, records, base)
			if len(rows) != 1 || rows[0].GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED {
				t.Fatalf("rejected event has no terminal record: %v", rows)
			}
		})
	}
}

func TestRedisStopCancelsEveryStartedBatchMemberAndPreservesNewPrompt(t *testing.T) {
	q, base := planeQueueFixture(t)
	events := acceptBatch(t, q, base, "first", "second")
	r := &controlledTestRunner{started: make(chan struct{}), canceled: make(chan struct{}), settle: make(chan struct{})}
	var once sync.Once
	settle := func() { once.Do(func() { close(r.settle) }) }
	defer settle()
	records := processingmem.New()
	holder := batchProcessor(q, r, records, "holder-pod")
	app := &agentsv1.PlaneApp{AgentId: "agent", InboundEnabled: true}
	done := make(chan error, 1)
	go func() {
		_, err := holder.Process(t.Context(), app, events[0])
		done <- err
	}()
	awaitPlane(t, r.started, "batch execution")
	discarded := *base
	discarded.ActivityID = "old-unread"
	old := acceptBatch(t, q, &discarded, "discard before Stop")[0]
	stop := *base
	stop.ActivityID, stop.Stop, stop.Signal, stop.Prompt = "stop", true, "stop", ""
	if _, err := q.Accept(t.Context(), &stop); err != nil {
		t.Fatal(err)
	}
	awaitPlane(t, r.canceled, "batch Stop cancellation")
	future := *base
	future.ActivityID = "new-after-stop"
	next := acceptBatch(t, q, &future, "new work")[0]
	successorRunner := &batchRunner{}
	successor := batchProcessor(q, successorRunner, records, "other-pod")
	if _, err := successor.Process(t.Context(), app, next); !errors.Is(err, planeprocessing.ErrClaimConflict) {
		t.Fatalf("successor ran before cancellation settled: %v", err)
	}
	settle()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Stop lost to a clean return: %v", err)
	}
	if _, err := successor.Process(t.Context(), app, next); err != nil {
		t.Fatal(err)
	}
	w := NewWorker(q, successor, &planeTestResolver{app: app}, "other-pod")
	entries, _ := q.Read(t.Context(), "other-pod", 4, time.Millisecond)
	for _, entry := range entries {
		if err := w.handle(t.Context(), entry); err != nil {
			t.Fatal(err)
		}
	}
	for _, row := range processingRows(t, records, base) {
		want := agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_CANCELLED
		if row.GetActivityId() == next.ActivityID {
			want = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED
		}
		if row.GetStatus() != want {
			t.Fatalf("Stop affected the wrong batch generation: %v", row)
		}
		if row.GetActivityId() == old.ActivityID && row.GetInvocationId() != "" {
			t.Fatalf("discarded pending input manufactured an Invocation: %v", row)
		}
	}
	if r.calls.Load() != 1 || len(successorRunner.snapshot()) != 1 {
		t.Fatal("stopped batch was repeated or a new prompt was discarded")
	}
}

func TestRedisBatchUsesRealRunnerInvocationAndContinuesSessionHistory(t *testing.T) {
	q, base := planeQueueFixture(t)
	recorder := invocationmem.New()
	p, e, app := invocationProcessor(t, func(agent.InvocationContext) (string, error) {
		return "real Runner result", nil
	}, recorder)
	app.InboundEnabled = true
	e.InstallationID, e.ActorUserID, e.BotUserID, e.AgentID = base.InstallationID, "actor", "bot", app.GetAgentId()
	t.Cleanup(func() {
		ctx := context.Background()
		keys, _ := q.client.Keys(ctx, controlKey(e)+"*").Result()
		if len(keys) > 0 {
			_ = q.client.Del(ctx, keys...).Err()
		}
	})
	records := processingmem.New()
	p.batches, p.control = &redisBatchSource{client: q.client}, NewRedisRunControl(q.client)
	p.processing = records
	p.SetClaimRepository(records, "pod")
	first, second := *e, *e
	first.ActivityID, first.Prompt = "first", "first input"
	second.ActivityID, second.Prompt = "second", "second input"
	for _, event := range []*Event{&first, &second} {
		if _, err := q.Accept(t.Context(), event); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := p.Process(t.Context(), app, &first); err != nil {
		t.Fatal(err)
	}
	invocations, _, _, err := recorder.List(t.Context(), invocation.ListFilter{WorkspaceID: e.WorkspaceID}, 10, "")
	if err != nil || len(invocations) != 1 {
		t.Fatalf("real Runner produced duplicate Invocations: %v %v", invocations, err)
	}
	for _, row := range processingRows(t, records, e) {
		if row.GetInvocationId() != invocations[0].GetId() {
			t.Fatalf("processing record points to a nonexistent Invocation: %v", row)
		}
	}
	svc := p.runner.(*runner.Service)
	sessionID, userID := planeSession(e, e.AgentID), planeServiceUser(e)
	before, err := svc.GetSession(t.Context(), "plane", sessionID, userID)
	if err != nil {
		t.Fatal(err)
	}
	beforeCount := before.Events().Len()
	next := *e
	next.ActivityID, next.Prompt = "next", "continue this Run"
	if _, err := q.Accept(t.Context(), &next); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Process(t.Context(), app, &next); err != nil {
		t.Fatal(err)
	}
	after, err := svc.GetSession(t.Context(), "plane", sessionID, userID)
	if err != nil || beforeCount == 0 || after.Events().Len() <= beforeCount {
		t.Fatalf("successive batch lost its session history: before=%d err=%v", beforeCount, err)
	}
}
