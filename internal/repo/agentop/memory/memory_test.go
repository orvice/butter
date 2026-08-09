package memory

import (
	"strconv"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	agentoprepo "go.orx.me/apps/butter/internal/repo/agentop"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func op(id, ws string, status agentsv1.AgentOperationStatus, createdUnix int64) *agentsv1.AgentOperation {
	return &agentsv1.AgentOperation{
		Id:          id,
		WorkspaceId: ws,
		AgentId:     "agent-" + id,
		Type:        agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_CREATE,
		Status:      status,
		CreatedAt:   timestamppb.New(time.Unix(createdUnix, 0)),
	}
}

func itoa(i int64) string { return strconv.FormatInt(i, 10) }

func TestMemoryRepo_SaveGet(t *testing.T) {
	r := New()
	ctx := t.Context()
	want := op("1", "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING, 100)
	if err := r.Create(ctx, "ws-a", want); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := r.Get(ctx, "ws-a", "1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.GetId() != "1" || got.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING {
		t.Fatalf("unexpected op: %+v", got)
	}

	// Cross-workspace isolation.
	if _, err := r.Get(ctx, "ws-b", "1"); err != agentoprepo.ErrNotFound {
		t.Fatalf("cross-workspace Get: want ErrNotFound, got %v", err)
	}
	if _, err := r.Get(ctx, "ws-a", "missing"); err != agentoprepo.ErrNotFound {
		t.Fatalf("missing Get: want ErrNotFound, got %v", err)
	}
}

func TestMemoryRepo_AllowsSameIDAcrossWorkspaces(t *testing.T) {
	r := New()
	ctx := t.Context()

	for _, ws := range []string{"ws-a", "ws-b"} {
		if err := r.Create(ctx, ws, op("shared-id", ws, agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_PENDING, 100)); err != nil {
			t.Fatalf("Create(%s): %v", ws, err)
		}
	}

	for _, ws := range []string{"ws-a", "ws-b"} {
		got, err := r.Get(ctx, ws, "shared-id")
		if err != nil {
			t.Fatalf("Get(%s): %v", ws, err)
		}
		if got.GetWorkspaceId() != ws {
			t.Fatalf("Get(%s) returned workspace %q", ws, got.GetWorkspaceId())
		}
	}
}

func TestMemoryRepo_SaveIsSnapshot(t *testing.T) {
	r := New()
	ctx := t.Context()
	in := op("1", "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_PENDING, 100)
	if err := r.Create(ctx, "ws-a", in); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Mutating the input after Save must not affect stored state.
	in.Status = agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED
	got, _ := r.Get(ctx, "ws-a", "1")
	if got.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_PENDING {
		t.Fatalf("stored op was mutated through the input pointer: %v", got.GetStatus())
	}
}

func TestMemoryRepo_ListFilterAndOrder(t *testing.T) {
	r := New()
	ctx := t.Context()
	_ = r.Create(ctx, "ws-a", op("1", "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED, 100))
	_ = r.Create(ctx, "ws-a", op("2", "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED, 200))
	_ = r.Create(ctx, "ws-a", op("3", "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED, 300))
	_ = r.Create(ctx, "ws-b", op("4", "ws-b", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED, 400))

	// All in ws-a, newest first.
	all, next, err := r.List(ctx, "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_UNSPECIFIED, 20, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if next != "" {
		t.Fatalf("unexpected next token %q", next)
	}
	if len(all) != 3 || all[0].GetId() != "3" || all[2].GetId() != "1" {
		t.Fatalf("unexpected order/count: %v", ids(all))
	}

	// Status filter.
	failed, _, _ := r.List(ctx, "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED, 20, "")
	if len(failed) != 1 || failed[0].GetId() != "2" {
		t.Fatalf("status filter: %v", ids(failed))
	}
}

func TestMemoryRepo_ListPagination(t *testing.T) {
	r := New()
	ctx := t.Context()
	for i := int64(1); i <= 5; i++ {
		_ = r.Create(ctx, "ws-a", op(itoa(i), "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED, i))
	}
	page1, next, _ := r.List(ctx, "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_UNSPECIFIED, 2, "")
	if len(page1) != 2 || next == "" {
		t.Fatalf("page1 len=%d next=%q", len(page1), next)
	}
	page2, next2, _ := r.List(ctx, "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_UNSPECIFIED, 2, next)
	if len(page2) != 2 || next2 == "" {
		t.Fatalf("page2 len=%d next=%q", len(page2), next2)
	}
	page3, next3, _ := r.List(ctx, "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_UNSPECIFIED, 2, next2)
	if len(page3) != 1 || next3 != "" {
		t.Fatalf("page3 len=%d next=%q", len(page3), next3)
	}
}

func TestMemoryRepo_ClaimPreventsConcurrentExecutor(t *testing.T) {
	r := New()
	ctx := t.Context()
	pending := op("claim", "ws-a", agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_PENDING, 100)
	pending.UpdatedAt = timestamppb.New(time.Unix(100, 0))
	if err := r.Create(ctx, "ws-a", pending); err != nil {
		t.Fatalf("Create: %v", err)
	}

	claimedAt := time.Unix(200, 0)
	claimed, err := r.Claim(ctx, "ws-a", "claim", "lease-a", claimedAt, claimedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if claimed.GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING || claimed.GetAttemptCount() != 1 {
		t.Fatalf("unexpected claimed operation: %v", claimed)
	}
	if err := r.RenewLease(ctx, "ws-a", "claim", "lease-a", claimedAt.Add(2*time.Minute)); err != nil {
		t.Fatalf("RenewLease: %v", err)
	}
	if _, err := r.Claim(ctx, "ws-a", "claim", "lease-b", claimedAt.Add(90*time.Second), claimedAt.Add(150*time.Second)); err != agentoprepo.ErrInProgress {
		t.Fatalf("concurrent Claim: want ErrInProgress, got %v", err)
	}
	claimed, err = r.Claim(ctx, "ws-a", "claim", "lease-b", claimedAt.Add(121*time.Second), claimedAt.Add(181*time.Second))
	if err != nil {
		t.Fatalf("Claim after lease expiry: %v", err)
	}

	claimed.Status = agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED
	if err := r.SaveClaimed(ctx, "ws-a", "lease-a", claimed); err != agentoprepo.ErrLeaseLost {
		t.Fatalf("stale SaveClaimed: want ErrLeaseLost, got %v", err)
	}
	if err := r.SaveClaimed(ctx, "ws-a", "lease-b", claimed); err != nil {
		t.Fatalf("SaveClaimed: %v", err)
	}
}

func ids(ops []*agentsv1.AgentOperation) []string {
	out := make([]string, len(ops))
	for i, o := range ops {
		out[i] = o.GetId()
	}
	return out
}
