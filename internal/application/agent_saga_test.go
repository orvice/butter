package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/agentcontent"
	agentoprepo "go.orx.me/apps/butter/internal/repo/agentop"
	agentopmemory "go.orx.me/apps/butter/internal/repo/agentop/memory"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeContent is an injectable agentContentCoordinator for saga tests.
type fakeContent struct {
	hasBinding  bool
	commitErr   error
	commitVErrs []string
	commitSHA   string
	commitCalls int
	syncErr     error
	syncCalls   int
	commitStart chan struct{}
	commitDone  chan struct{}
	startOnce   sync.Once
}

func (f *fakeContent) HasBinding(context.Context, string) (bool, error) { return f.hasBinding, nil }

func (f *fakeContent) IsContentGitOwned(context.Context, string) (bool, error) {
	return f.hasBinding, nil
}

func (f *fakeContent) GetActiveSnapshot(context.Context, string) (*agentcontent.Snapshot, error) {
	return nil, nil
}

func (f *fakeContent) CommitContent(_ context.Context, _ string, _ []*agentsv1.ContentFileAction, _, _ string) (string, []string, error) {
	f.commitCalls++
	if f.commitStart != nil {
		f.startOnce.Do(func() { close(f.commitStart) })
		<-f.commitDone
	}
	if f.commitErr != nil {
		return "", nil, f.commitErr
	}
	if len(f.commitVErrs) > 0 {
		return "", f.commitVErrs, nil
	}
	sha := f.commitSHA
	if sha == "" {
		sha = "sha-1"
	}
	return sha, nil, nil
}

func (f *fakeContent) SyncAndPublish(context.Context, string) error {
	f.syncCalls++
	return f.syncErr
}

type sagaFixture struct {
	agents      *configmemory.Store
	ops         agentoprepo.Repository
	content     *fakeContent
	coord       *agentOperationCoordinator
	reloadCalls int
	reloadErr   error
}

type failingOperationRepo struct {
	agentoprepo.Repository
	failWhen func(*agentsv1.AgentOperation) bool
	err      error
}

type renewTrackingOperationRepo struct {
	agentoprepo.Repository
	renewed chan struct{}
	once    sync.Once
}

func (r *renewTrackingOperationRepo) RenewLease(ctx context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error {
	if err := r.Repository.RenewLease(ctx, workspaceID, id, leaseToken, leaseExpiresAt); err != nil {
		return err
	}
	r.once.Do(func() { close(r.renewed) })
	return nil
}

func (r *failingOperationRepo) SaveClaimed(ctx context.Context, workspaceID, leaseToken string, op *agentsv1.AgentOperation) error {
	if r.failWhen != nil && r.failWhen(op) {
		return r.err
	}
	return r.Repository.SaveClaimed(ctx, workspaceID, leaseToken, op)
}

func newSagaFixture(t *testing.T) *sagaFixture {
	t.Helper()
	fx := &sagaFixture{
		agents:  configmemory.New(),
		ops:     agentopmemory.New(),
		content: &fakeContent{hasBinding: true},
	}
	fx.coord = newAgentOperationCoordinator(fx.agents, fx.ops, fx.content, func(context.Context) error {
		fx.reloadCalls++
		return fx.reloadErr
	})
	return fx
}

func llmAgent(id string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:    id,
		AgentId: id,
		Type:    agentsv1.AgentType_AGENT_TYPE_LLM,
	}
}

func promptInput(prompt string) *agentsv1.AgentContentInput {
	return &agentsv1.AgentContentInput{Prompt: prompt}
}

const wsA = "ws-a"

func TestSagaCreate_HappyPath(t *testing.T) {
	fx := newSagaFixture(t)
	ctx := context.Background()

	agent, op, vErrs, err := fx.coord.RunCreate(ctx, wsA, llmAgent("a"), promptInput("do things"), "")
	if err != nil || len(vErrs) > 0 {
		t.Fatalf("RunCreate: err=%v vErrs=%v", err, vErrs)
	}
	if agent.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
		t.Fatalf("want ACTIVE, got %v", agent.GetLifecycleStatus())
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("want op SUCCEEDED, got %v", op.GetStatus())
	}
	if op.GetCommittedSha() == "" {
		t.Fatalf("expected committed_sha to be recorded")
	}
	if fx.content.commitCalls != 1 || fx.content.syncCalls != 1 {
		t.Fatalf("commit=%d sync=%d", fx.content.commitCalls, fx.content.syncCalls)
	}
}

func TestSagaCreate_ContentValidationLeavesErrored(t *testing.T) {
	fx := newSagaFixture(t)
	fx.content.commitVErrs = []string{`agent "a" (agents/a/prompt.md): LLM agent requires a non-empty prompt.md`}
	ctx := context.Background()

	_, op, vErrs, err := fx.coord.RunCreate(ctx, wsA, llmAgent("a"), promptInput("x"), "")
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if len(vErrs) == 0 {
		t.Fatalf("expected validation errors")
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED {
		t.Fatalf("want op FAILED, got %v", op.GetStatus())
	}
	// The DB row is retained (not runnable) so retry can recover it.
	stored, err := fx.agents.GetAgentByID(ctx, wsA, "a")
	if err != nil {
		t.Fatalf("agent should remain in DB after a failed create: %v", err)
	}
	if stored.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ERROR {
		t.Fatalf("want ERROR, got %v", stored.GetLifecycleStatus())
	}
}

func TestSagaCreate_SyncFailureThenRetryResumes(t *testing.T) {
	fx := newSagaFixture(t)
	fx.content.syncErr = errors.New("git unavailable")
	ctx := context.Background()

	_, op, _, err := fx.coord.RunCreate(ctx, wsA, llmAgent("a"), promptInput("do things"), "")
	if err == nil {
		t.Fatalf("expected sync failure")
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED {
		t.Fatalf("want FAILED, got %v", op.GetStatus())
	}
	if fx.content.commitCalls != 1 {
		t.Fatalf("commit should have run once, got %d", fx.content.commitCalls)
	}

	// Fix the fault and retry: commit must NOT run again (idempotent resume),
	// sync + activate must complete.
	fx.content.syncErr = nil
	agent, op2, vErrs, err := fx.coord.Retry(ctx, wsA, op.GetId())
	if err != nil || len(vErrs) > 0 {
		t.Fatalf("Retry: err=%v vErrs=%v", err, vErrs)
	}
	if fx.content.commitCalls != 1 {
		t.Fatalf("commit re-ran on retry (not idempotent): %d", fx.content.commitCalls)
	}
	if agent.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
		t.Fatalf("want ACTIVE after retry, got %v", agent.GetLifecycleStatus())
	}
	if op2.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("want SUCCEEDED after retry, got %v", op2.GetStatus())
	}
	if op2.GetAttemptCount() != 2 {
		t.Fatalf("want attempt_count 2, got %d", op2.GetAttemptCount())
	}
}

func TestSagaCreate_RejectsOperationIDReusedForDifferentRequest(t *testing.T) {
	fx := newSagaFixture(t)
	ctx := context.Background()

	if _, _, _, err := fx.coord.RunCreate(ctx, wsA, llmAgent("a"), promptInput("first"), "same-op"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, _, _, err := fx.coord.RunCreate(ctx, wsA, llmAgent("a"), promptInput("different"), "same-op")
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want InvalidArgument for mismatched replay, got %v", err)
	}
}

func TestSagaCreate_ConcurrentReplayDoesNotRepeatCommit(t *testing.T) {
	fx := newSagaFixture(t)
	fx.content.commitStart = make(chan struct{})
	fx.content.commitDone = make(chan struct{})
	ctx := context.Background()

	type result struct {
		op  *agentsv1.AgentOperation
		err error
	}
	first := make(chan result, 1)
	go func() {
		_, op, _, err := fx.coord.RunCreate(ctx, wsA, llmAgent("concurrent"), promptInput("hi"), "concurrent-op")
		first <- result{op: op, err: err}
	}()
	<-fx.content.commitStart

	second := make(chan result, 1)
	go func() {
		_, op, _, err := fx.coord.RunCreate(ctx, wsA, llmAgent("concurrent"), promptInput("hi"), "concurrent-op")
		second <- result{op: op, err: err}
	}()
	select {
	case got := <-second:
		if got.err != nil {
			t.Fatalf("concurrent replay: %v", got.err)
		}
		if got.op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING {
			t.Fatalf("replay status = %v, want RUNNING", got.op.GetStatus())
		}
	case <-time.After(time.Second):
		t.Fatal("concurrent replay blocked behind the active external step")
	}
	if fx.content.commitCalls != 1 {
		t.Fatalf("commit calls = %d, want 1", fx.content.commitCalls)
	}

	close(fx.content.commitDone)
	if got := <-first; got.err != nil {
		t.Fatalf("first create: %v", got.err)
	}
}

func TestSagaExecute_HeartbeatsPreventStaleTakeover(t *testing.T) {
	fx := newSagaFixture(t)
	tracking := &renewTrackingOperationRepo{Repository: fx.ops, renewed: make(chan struct{})}
	fx.ops = tracking
	fx.coord.ops = tracking
	fx.coord.leaseDuration = time.Minute
	fx.coord.leaseHeartbeatInterval = 20 * time.Millisecond

	base := time.Unix(1_700_000_000, 0)
	var nowNanos atomic.Int64
	nowNanos.Store(base.UnixNano())
	fx.coord.now = func() time.Time { return time.Unix(0, nowNanos.Load()) }
	fx.content.commitStart = make(chan struct{})
	fx.content.commitDone = make(chan struct{})

	first := make(chan error, 1)
	go func() {
		_, _, _, err := fx.coord.RunCreate(context.Background(), wsA, llmAgent("heartbeat"), promptInput("hi"), "heartbeat-op")
		first <- err
	}()
	<-fx.content.commitStart

	// Move the logical clock beyond the original lease. The heartbeat must
	// renew against the new time before a retry attempts to claim the operation.
	nowNanos.Store(base.Add(2 * time.Minute).UnixNano())
	select {
	case <-tracking.renewed:
	case <-time.After(time.Second):
		t.Fatal("operation lease was not renewed")
	}

	_, op, _, err := fx.coord.Retry(context.Background(), wsA, "heartbeat-op")
	if err != nil {
		t.Fatalf("concurrent retry: %v", err)
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING {
		t.Fatalf("retry status = %v, want RUNNING", op.GetStatus())
	}
	if fx.content.commitCalls != 1 {
		t.Fatalf("commit calls = %d, want 1", fx.content.commitCalls)
	}

	close(fx.content.commitDone)
	if err := <-first; err != nil {
		t.Fatalf("first create: %v", err)
	}
}

func TestSagaExecute_SurfacesFailurePersistenceError(t *testing.T) {
	fx := newSagaFixture(t)
	store := fx.ops
	fx.ops = &failingOperationRepo{
		Repository: store,
		err:        errors.New("operation store unavailable"),
		failWhen: func(op *agentsv1.AgentOperation) bool {
			return op.GetStatus() == agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED
		},
	}
	fx.coord.ops = fx.ops
	fx.content.syncErr = errors.New("publish failed")

	_, _, _, err := fx.coord.RunCreate(context.Background(), wsA, llmAgent("a"), promptInput("do things"), "persist-failure")
	if err == nil || !strings.Contains(err.Error(), "operation store unavailable") {
		t.Fatalf("want persistence error to be surfaced, got %v", err)
	}
}

func TestSagaDelete_Tombstone(t *testing.T) {
	fx := newSagaFixture(t)
	ctx := context.Background()
	seedActive(t, fx, "a")
	commitsBefore := fx.content.commitCalls

	op, err := fx.coord.RunDelete(ctx, wsA, mustGet(t, fx, "a"), "")
	if err != nil {
		t.Fatalf("RunDelete: %v", err)
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("want SUCCEEDED, got %v", op.GetStatus())
	}
	stored := mustGet(t, fx, "a")
	if stored.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
		t.Fatalf("want DELETED, got %v", stored.GetLifecycleStatus())
	}
	if stored.GetDeletedAt() == nil {
		t.Fatalf("expected deleted_at to be set")
	}
	// Content untouched by the tombstone.
	if fx.content.commitCalls != commitsBefore {
		t.Fatalf("delete must not touch content, commitCalls went %d→%d", commitsBefore, fx.content.commitCalls)
	}
}

func TestSagaRestore_RoundTrip(t *testing.T) {
	fx := newSagaFixture(t)
	ctx := context.Background()
	seedActive(t, fx, "a")
	if _, err := fx.coord.RunDelete(ctx, wsA, mustGet(t, fx, "a"), ""); err != nil {
		t.Fatalf("delete: %v", err)
	}

	agent, op, vErrs, err := fx.coord.RunRestore(ctx, wsA, mustGet(t, fx, "a"), "")
	if err != nil || len(vErrs) > 0 {
		t.Fatalf("RunRestore: err=%v vErrs=%v", err, vErrs)
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("want SUCCEEDED, got %v", op.GetStatus())
	}
	if agent.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
		t.Fatalf("want ACTIVE, got %v", agent.GetLifecycleStatus())
	}
	if agent.GetDeletedAt() != nil {
		t.Fatalf("deleted_at should be cleared on restore")
	}
}

func TestSagaUpdateConfiguration_VersionMatchAndMismatch(t *testing.T) {
	fx := newSagaFixture(t)
	ctx := context.Background()
	seedActive(t, fx, "a") // version 1 after seeding via create saga

	prev := mustGet(t, fx, "a")
	patch := proto.Clone(prev).(*agentsv1.Agent)
	patch.DisplayName = "patched-display"

	// Correct expected version succeeds.
	agent, op, vErrs, err := fx.coord.RunUpdateConfiguration(ctx, wsA, prev, patch, nil, prev.GetVersion(), "", "")
	if err != nil || len(vErrs) > 0 {
		t.Fatalf("update: err=%v vErrs=%v", err, vErrs)
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("want SUCCEEDED, got %v", op.GetStatus())
	}
	if agent.GetDisplayName() != "patched-display" {
		t.Fatalf("patch not applied: %q", agent.GetDisplayName())
	}
	if agent.GetVersion() != prev.GetVersion()+1 {
		t.Fatalf("version not bumped: %d", agent.GetVersion())
	}

	// Stale expected version aborts.
	stale := mustGet(t, fx, "a")
	stale.DisplayName = "again"
	_, op2, _, err := fx.coord.RunUpdateConfiguration(ctx, wsA, stale, stale, nil, 0, "", "")
	if connect.CodeOf(err) != connect.CodeAborted {
		t.Fatalf("want Aborted on stale version, got %v", err)
	}
	if op2.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED {
		t.Fatalf("want op FAILED on conflict, got %v", op2.GetStatus())
	}
}

func TestSagaUpdateConfiguration_CannotResurrectDeletedAgent(t *testing.T) {
	fx := newSagaFixture(t)
	ctx := context.Background()
	seedActive(t, fx, "a")
	stale := mustGet(t, fx, "a")
	patch := proto.Clone(stale).(*agentsv1.Agent)
	patch.DisplayName = "stale update"

	if _, err := fx.coord.RunDelete(ctx, wsA, mustGet(t, fx, "a"), "delete-op"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, op, _, err := fx.coord.RunUpdateConfiguration(ctx, wsA, stale, patch, nil, stale.GetVersion(), "", "stale-update-op")
	if connect.CodeOf(err) != connect.CodeAborted {
		t.Fatalf("want Aborted for stale update after delete, got %v", err)
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED {
		t.Fatalf("want FAILED operation, got %v", op.GetStatus())
	}
	stored := mustGet(t, fx, "a")
	if stored.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
		t.Fatalf("stale update resurrected agent: %v", stored.GetLifecycleStatus())
	}
}

func TestSagaUpdateConfiguration_RetryAfterReloadFailure(t *testing.T) {
	fx := newSagaFixture(t)
	ctx := context.Background()
	seedActive(t, fx, "reload-retry")
	prev := mustGet(t, fx, "reload-retry")
	patch := proto.Clone(prev).(*agentsv1.Agent)
	patch.DisplayName = "updated"

	fx.reloadErr = errors.New("runtime unavailable")
	_, op, _, err := fx.coord.RunUpdateConfiguration(ctx, wsA, prev, patch, nil, prev.GetVersion(), "", "reload-retry-op")
	if err == nil {
		t.Fatal("expected reload failure")
	}
	stored := mustGet(t, fx, "reload-retry")
	if stored.GetDisplayName() != "updated" {
		t.Fatalf("DB patch should remain durable after reload failure, got %q", stored.GetDisplayName())
	}

	fx.reloadErr = nil
	got, retried, vErrs, err := fx.coord.Retry(ctx, wsA, op.GetId())
	if err != nil || len(vErrs) > 0 {
		t.Fatalf("Retry: err=%v validation=%v", err, vErrs)
	}
	if retried.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("operation status = %v, want SUCCEEDED", retried.GetStatus())
	}
	if got.GetDisplayName() != "updated" {
		t.Fatalf("retried agent display name = %q", got.GetDisplayName())
	}
}

// seedActive provisions an agent to ACTIVE via the create saga.
func seedActive(t *testing.T, fx *sagaFixture, id string) {
	t.Helper()
	if _, _, _, err := fx.coord.RunCreate(context.Background(), wsA, llmAgent(id), promptInput("do things"), ""); err != nil {
		t.Fatalf("seed create: %v", err)
	}
}

func mustGet(t *testing.T, fx *sagaFixture, id string) *agentsv1.Agent {
	t.Helper()
	a, err := fx.agents.GetAgentByID(context.Background(), wsA, id)
	if err != nil {
		t.Fatalf("GetAgentByID %q: %v", id, err)
	}
	return a
}
