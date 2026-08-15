package automation

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestEngineRunNowAsyncAcceptsThenCompletes(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	block := make(chan struct{})
	runnerSvc := &engineRunner{outputs: []string{"done"}, block: block}
	engine := NewEngine(defRepo, runRepo, stepRepo, EngineOptions{Runner: runnerSvc})

	automation := &agentsv1.Automation{
		Name:        "slow",
		Enabled:     true,
		WorkspaceId: "ws1",
		Trigger:     &agentsv1.AutomationTrigger{Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL},
		Steps: []*agentsv1.AutomationStep{
			{Name: "invoke", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id", Input: "go"}},
		},
	}
	if err := defRepo.Create(ctx, automation); err != nil {
		t.Fatalf("create automation: %v", err)
	}

	// Cancel the caller's context immediately after acceptance: the background
	// run must not be affected, because it executes on the engine context.
	callerCtx, cancelCaller := context.WithCancel(ctx)
	accepted, err := engine.RunNowAsync(callerCtx, "ws1", "slow", `{}`)
	if err != nil {
		t.Fatalf("RunNowAsync: %v", err)
	}
	cancelCaller()
	if accepted.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING {
		t.Fatalf("accepted status = %s, want RUNNING", accepted.GetStatus())
	}
	stored, err := runRepo.Get(ctx, "ws1", accepted.GetId())
	if err != nil {
		t.Fatalf("accepted run not persisted: %v", err)
	}
	if stored.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING {
		t.Fatalf("persisted status = %s, want RUNNING while the agent turn blocks", stored.GetStatus())
	}

	close(block)
	waitForRunStatus(t, runRepo, "ws1", accepted.GetId(), agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED)
}

func waitForRunStatus(t *testing.T, runRepo RunRepo, workspaceID, id string, want agentsv1.AutomationRunStatus) *agentsv1.AutomationRun {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		run, err := runRepo.Get(context.Background(), workspaceID, id)
		if err != nil {
			t.Fatalf("Get run: %v", err)
		}
		if run.GetStatus() == want {
			return run
		}
		if time.Now().After(deadline) {
			t.Fatalf("run status = %s (err %q), want %s", run.GetStatus(), run.GetError(), want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// fakeRunGuard simulates the cross-instance lock: `held` keys are owned by
// another instance, and Acquire records every attempt.
type fakeRunGuard struct {
	mu       sync.Mutex
	held     map[string]bool
	err      error
	acquires int
}

func (g *fakeRunGuard) Acquire(ctx context.Context, key string) (context.Context, func(), bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.acquires++
	if g.err != nil {
		return ctx, func() {}, false, g.err
	}
	if g.held[key] {
		return ctx, func() {}, false, nil
	}
	return ctx, func() {}, true, nil
}

func TestEngineRunGuardSkipsWhenHeldElsewhere(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	runnerSvc := &engineRunner{}
	guard := &fakeRunGuard{held: map[string]bool{RunLeaseKeyPrefix + automationID("ws1", "guarded"): true}}
	engine := NewEngine(defRepo, runRepo, stepRepo, EngineOptions{Runner: runnerSvc, RunGuard: guard})

	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "guarded",
		Enabled:     true,
		WorkspaceId: "ws1",
		Policy:      &agentsv1.AutomationPolicy{Concurrency: agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_SKIP},
		Steps: []*agentsv1.AutomationStep{
			{Name: "invoke", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}},
		},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SKIPPED {
		t.Fatalf("status = %s, want SKIPPED", run.GetStatus())
	}
	if !strings.Contains(run.GetError(), "another instance") {
		t.Fatalf("skip reason = %q, want it to blame the other instance", run.GetError())
	}
	if runnerSvc.calls != 0 {
		t.Fatalf("runner calls = %d, want 0", runnerSvc.calls)
	}
}

func TestEngineRunGuardQueueWaitsForLock(t *testing.T) {
	old := runGuardRetryInterval
	runGuardRetryInterval = 5 * time.Millisecond
	defer func() { runGuardRetryInterval = old }()

	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	runnerSvc := &engineRunner{outputs: []string{"done"}}
	key := RunLeaseKeyPrefix + automationID("ws1", "queued")
	guard := &fakeRunGuard{held: map[string]bool{key: true}}
	engine := NewEngine(defRepo, runRepo, stepRepo, EngineOptions{Runner: runnerSvc, RunGuard: guard})

	// Free the lock shortly after the run starts polling for it.
	go func() {
		time.Sleep(25 * time.Millisecond)
		guard.mu.Lock()
		delete(guard.held, key)
		guard.mu.Unlock()
	}()

	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "queued",
		Enabled:     true,
		WorkspaceId: "ws1",
		Policy:      &agentsv1.AutomationPolicy{Concurrency: agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_QUEUE},
		Steps: []*agentsv1.AutomationStep{
			{Name: "invoke", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}},
		},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED {
		t.Fatalf("status = %s, want SUCCEEDED after the lock freed; err=%s", run.GetStatus(), run.GetError())
	}
	guard.mu.Lock()
	attempts := guard.acquires
	guard.mu.Unlock()
	if attempts < 2 {
		t.Fatalf("guard acquires = %d, want >= 2 (initial miss, later win)", attempts)
	}
}

func TestEngineRunGuardErrorFailsClosed(t *testing.T) {
	ctx := context.Background()
	runnerSvc := &engineRunner{}
	engine := NewEngine(NewMemoryDefinitionRepo(), NewMemoryRunRepo(), NewMemoryStepRunRepo(), EngineOptions{
		Runner:   runnerSvc,
		RunGuard: &fakeRunGuard{err: errors.New("redis down")},
	})

	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "unlockable",
		Enabled:     true,
		WorkspaceId: "ws1",
		Policy:      &agentsv1.AutomationPolicy{Concurrency: agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_SKIP},
		Steps: []*agentsv1.AutomationStep{
			{Name: "invoke", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}},
		},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED {
		t.Fatalf("status = %s, want FAILED (fail closed, never double-run)", run.GetStatus())
	}
	if runnerSvc.calls != 0 {
		t.Fatalf("runner calls = %d, want 0", runnerSvc.calls)
	}
}

func TestEngineReconcileStaleRuns(t *testing.T) {
	ctx := context.Background()
	runRepo := NewMemoryRunRepo()
	engine := NewEngine(NewMemoryDefinitionRepo(), runRepo, NewMemoryStepRunRepo(), EngineOptions{})

	stale := &agentsv1.AutomationRun{
		Id:             "stale-run",
		AutomationName: "old",
		Status:         agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING,
		StartedAt:      timestamppb.New(time.Now().UTC().Add(-StaleRunAge - time.Hour)),
		WorkspaceId:    "ws1",
	}
	fresh := &agentsv1.AutomationRun{
		Id:             "fresh-run",
		AutomationName: "new",
		Status:         agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING,
		StartedAt:      timestamppb.New(time.Now().UTC()),
		WorkspaceId:    "ws1",
	}
	waiting := &agentsv1.AutomationRun{
		Id:             "waiting-run",
		AutomationName: "paused",
		Status:         agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT,
		StartedAt:      timestamppb.New(time.Now().UTC().Add(-StaleRunAge - time.Hour)),
		WorkspaceId:    "ws1",
	}
	for _, run := range []*agentsv1.AutomationRun{stale, fresh, waiting} {
		if err := runRepo.Save(ctx, run); err != nil {
			t.Fatalf("Save(%s): %v", run.GetId(), err)
		}
	}

	count, err := engine.ReconcileStaleRuns(ctx)
	if err != nil {
		t.Fatalf("ReconcileStaleRuns: %v", err)
	}
	if count != 1 {
		t.Fatalf("reconciled = %d, want 1", count)
	}
	got, _ := runRepo.Get(ctx, "ws1", "stale-run")
	if got.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED {
		t.Fatalf("stale run status = %s, want FAILED", got.GetStatus())
	}
	if got.GetFinishedAt() == nil {
		t.Fatal("stale run finished_at not set")
	}
	if got, _ := runRepo.Get(ctx, "ws1", "fresh-run"); got.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING {
		t.Fatalf("fresh run status = %s, want RUNNING untouched", got.GetStatus())
	}
	// WAITING_INPUT runs wait on a human, not a process: never reconciled.
	if got, _ := runRepo.Get(ctx, "ws1", "waiting-run"); got.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT {
		t.Fatalf("waiting run status = %s, want WAITING_INPUT untouched", got.GetStatus())
	}
}

// fakeSchedulerLeader flips leadership on demand and records releases.
type fakeSchedulerLeader struct {
	mu       sync.Mutex
	held     bool
	released bool
}

func (l *fakeSchedulerLeader) Acquire(context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.held, nil
}

func (l *fakeSchedulerLeader) Release(context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.released = true
	return nil
}

func TestSchedulerLeaderGatesExecution(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	engine := NewEngine(defRepo, NewMemoryRunRepo(), NewMemoryStepRunRepo(), EngineOptions{})

	follower, err := NewScheduler(ctx, defRepo, engine)
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	leaderLease := &fakeSchedulerLeader{held: false}
	follower.SetLeader(leaderLease)
	follower.Start()
	if follower.HoldsLease() {
		t.Fatal("non-leader instance reports HoldsLease = true")
	}
	stopCtx := follower.Stop()
	<-stopCtx.Done()
	if !leaderLease.released {
		t.Fatal("Stop did not release the leader lease")
	}

	// Without a leader configured, a single instance always executes.
	solo, err := NewScheduler(ctx, defRepo, engine)
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	solo.Start()
	defer solo.Stop()
	if !solo.HoldsLease() {
		t.Fatal("leaderless scheduler must always execute its fires")
	}
}
