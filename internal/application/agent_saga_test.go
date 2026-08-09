package application

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

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
}

func (f *fakeContent) HasBinding(context.Context, string) (bool, error) { return f.hasBinding, nil }

func (f *fakeContent) CommitContent(_ context.Context, _ string, _ []*agentsv1.ContentFileAction, _, _ string) (string, []string, error) {
	f.commitCalls++
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
		return nil
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
	patch.Description = "patched"

	// Correct expected version succeeds.
	agent, op, vErrs, err := fx.coord.RunUpdateConfiguration(ctx, wsA, prev, patch, nil, prev.GetVersion(), "", "")
	if err != nil || len(vErrs) > 0 {
		t.Fatalf("update: err=%v vErrs=%v", err, vErrs)
	}
	if op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("want SUCCEEDED, got %v", op.GetStatus())
	}
	if agent.GetDescription() != "patched" {
		t.Fatalf("patch not applied: %q", agent.GetDescription())
	}
	if agent.GetVersion() != prev.GetVersion()+1 {
		t.Fatalf("version not bumped: %d", agent.GetVersion())
	}

	// Stale expected version aborts.
	stale := mustGet(t, fx, "a")
	stale.Description = "again"
	_, op2, _, err := fx.coord.RunUpdateConfiguration(ctx, wsA, stale, stale, nil, 0, "", "")
	if connect.CodeOf(err) != connect.CodeAborted {
		t.Fatalf("want Aborted on stale version, got %v", err)
	}
	if op2.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED {
		t.Fatalf("want op FAILED on conflict, got %v", op2.GetStatus())
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
