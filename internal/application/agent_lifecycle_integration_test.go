package application

import (
	"errors"
	"testing"

	"connectrpc.com/connect"

	agentopmemory "go.orx.me/apps/butter/internal/repo/agentop/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// newLifecycleAgentService wires an AgentServiceServer to the real repo-binding
// content seam (fx.svc) so the lifecycle Sagas exercise the actual
// commit/sync/publish pipeline through the fake Git provider.
func newLifecycleAgentService(t *testing.T) (*AgentServiceServer, *bindingFixture, *fakeConfigRuntime) {
	t.Helper()
	fx, rt := newContentEditFixture(t)
	fx.fake.materialize = true
	svc := NewAgentServiceServer(fx.agentRepo)
	svc.SetWorkspaceRepo(fx.wsRepo)
	svc.SetRuntime(rt)
	svc.SetOperationRepo(agentopmemory.New())
	svc.SetContentCoordinator(fx.svc)
	return svc, fx, rt
}

func TestLifecycleCreateBoundAgentActivates(t *testing.T) {
	svc, fx, _ := newLifecycleAgentService(t)

	commitsBefore := len(fx.fake.commits)

	resp, err := svc.CreateAgent(ownerCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "saga-agent", AgentId: "saga-agent",
			Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		},
		InitialContent: &agentsv1.AgentContentInput{Prompt: "You are helpful."},
	}))
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	if resp.Msg.GetAgent().GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
		t.Fatalf("want ACTIVE, got %v", resp.Msg.GetAgent().GetLifecycleStatus())
	}
	op := resp.Msg.GetOperation()
	if op == nil || op.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("want SUCCEEDED operation, got %v", op)
	}
	if len(fx.fake.commits) <= commitsBefore {
		t.Fatalf("create saga did not commit initial content (commits %d→%d)", commitsBefore, len(fx.fake.commits))
	}

	// The operation is retrievable.
	got, err := svc.GetAgentOperation(ownerCtx(), connect.NewRequest(&agentsv1.GetAgentOperationRequest{
		OperationId: op.GetId(),
	}))
	if err != nil {
		t.Fatalf("GetAgentOperation: %v", err)
	}
	if got.Msg.GetOperation().GetId() != op.GetId() {
		t.Fatalf("operation id mismatch")
	}
}

func TestLifecycleCreateBoundRequiresOwnerOrAdmin(t *testing.T) {
	svc, _, _ := newLifecycleAgentService(t)

	_, err := svc.CreateAgent(memberCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "member-agent", AgentId: "member-agent",
			Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		},
		InitialContent: &agentsv1.AgentContentInput{Prompt: "hi"},
	}))
	wantCode(t, err, connect.CodePermissionDenied)
}

func TestLifecycleCreateBoundLLMRequiresPrompt(t *testing.T) {
	svc, _, _ := newLifecycleAgentService(t)

	_, err := svc.CreateAgent(ownerCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "no-prompt", AgentId: "no-prompt",
			Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		},
	}))
	wantCode(t, err, connect.CodeInvalidArgument)
}

func TestLifecycleTombstoneThenRestore(t *testing.T) {
	svc, _, _ := newLifecycleAgentService(t)
	ctx := ownerCtx()

	if _, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "rt-agent", AgentId: "rt-agent",
			Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		},
		InitialContent: &agentsv1.AgentContentInput{Prompt: "hi"},
	})); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Delete → tombstone.
	if _, err := svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{AgentId: "rt-agent"})); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "rt-agent"}))
	if got.Msg.GetAgent().GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
		t.Fatalf("want DELETED, got %v", got.Msg.GetAgent().GetLifecycleStatus())
	}

	// Re-create with the same ID is steered to restore.
	_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "rt-agent", AgentId: "rt-agent",
			Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		},
		InitialContent: &agentsv1.AgentContentInput{Prompt: "hi"},
	}))
	wantCode(t, err, connect.CodeFailedPrecondition)

	// Restore → ACTIVE.
	resp, err := svc.RestoreAgent(ctx, connect.NewRequest(&agentsv1.RestoreAgentRequest{AgentId: "rt-agent"}))
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if resp.Msg.GetAgent().GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
		t.Fatalf("want ACTIVE after restore, got %v", resp.Msg.GetAgent().GetLifecycleStatus())
	}
	if resp.Msg.GetAgent().GetDeletedAt() != nil {
		t.Fatalf("deleted_at should be cleared after restore")
	}
}

func TestLifecycleUpdateConfigurationVersionConflict(t *testing.T) {
	svc, _, _ := newLifecycleAgentService(t)
	ctx := ownerCtx()

	created, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "cfg-agent", AgentId: "cfg-agent",
			Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		},
		InitialContent: &agentsv1.AgentContentInput{Prompt: "hi"},
	}))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	version := created.Msg.GetAgent().GetVersion()

	patch := &agentsv1.Agent{
		Name: "cfg-agent", AgentId: "cfg-agent",
		Type:        agentsv1.AgentType_AGENT_TYPE_LLM,
		DisplayName: "Renamed",
	}
	// Correct version succeeds.
	upd, err := svc.UpdateAgentConfiguration(ctx, connect.NewRequest(&agentsv1.UpdateAgentConfigurationRequest{
		AgentPatch: patch, ExpectedAgentVersion: version,
	}))
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.Msg.GetAgent().GetDisplayName() != "Renamed" {
		t.Fatalf("patch not applied")
	}

	// Stale version aborts.
	_, err = svc.UpdateAgentConfiguration(ctx, connect.NewRequest(&agentsv1.UpdateAgentConfigurationRequest{
		AgentPatch: patch, ExpectedAgentVersion: version,
	}))
	wantCode(t, err, connect.CodeAborted)
}

func TestLifecycleCreateFailsThenRetryViaRPC(t *testing.T) {
	svc, fx, _ := newLifecycleAgentService(t)
	ctx := ownerCtx()

	// Break the post-commit sync so the create Saga fails at SYNC_PUBLISH.
	fx.fake.treeErr = errors.New("tree unavailable")

	_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "retry-agent", AgentId: "retry-agent",
			Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		},
		InitialContent: &agentsv1.AgentContentInput{Prompt: "hi"},
	}))
	if err == nil {
		t.Fatal("expected create to fail while sync is broken")
	}

	// The failed agent is retained in ERROR (retained, not runnable).
	got, gErr := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "retry-agent"}))
	if gErr != nil {
		t.Fatalf("failed-create agent should be retained: %v", gErr)
	}
	if got.Msg.GetAgent().GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ERROR {
		t.Fatalf("want ERROR after failed create, got %v", got.Msg.GetAgent().GetLifecycleStatus())
	}

	// Find the FAILED operation.
	list, err := svc.ListAgentOperations(ctx, connect.NewRequest(&agentsv1.ListAgentOperationsRequest{
		Status: agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED,
	}))
	if err != nil {
		t.Fatalf("ListAgentOperations: %v", err)
	}
	if len(list.Msg.GetOperations()) != 1 {
		t.Fatalf("want 1 failed operation, got %d", len(list.Msg.GetOperations()))
	}
	opID := list.Msg.GetOperations()[0].GetId()

	// Heal the fault and retry via the RPC → agent activates.
	fx.fake.treeErr = nil
	retry, err := svc.RetryAgentOperation(ctx, connect.NewRequest(&agentsv1.RetryAgentOperationRequest{
		OperationId: opID,
	}))
	if err != nil {
		t.Fatalf("RetryAgentOperation: %v", err)
	}
	if retry.Msg.GetAgent().GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE {
		t.Fatalf("want ACTIVE after retry, got %v", retry.Msg.GetAgent().GetLifecycleStatus())
	}
	if retry.Msg.GetOperation().GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("want operation SUCCEEDED after retry, got %v", retry.Msg.GetOperation().GetStatus())
	}
}
