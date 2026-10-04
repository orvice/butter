package app

import (
	"testing"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestStartProcessLivenessWithoutRedisPublishesThisProcessOnly(t *testing.T) {
	p := startProcessLiveness(t.Context(), nil)
	if p.shared || p.instanceID == "" {
		t.Fatalf("liveness = %+v, want an unshared registry with an instance ID", p)
	}
	alive, err := p.registry.Alive(t.Context(), []string{p.instanceID, "another-process"})
	if err != nil {
		t.Fatalf("Alive: %v", err)
	}
	if !alive[p.instanceID] || alive["another-process"] {
		t.Fatalf("Alive = %v, want this process alive and no other", alive)
	}
}

// Acceptance (#390): without Redis there is one process, and startup
// reconciliation behaves as before. Every QUEUED or RUNNING record left by an
// earlier process is failed, however young, and nothing is replayed: the
// sweep has no runner to replay with.
func TestWithoutRedisStartupFailsEveryRecordAnEarlierProcessLeft(t *testing.T) {
	p := startProcessLiveness(t.Context(), nil)
	store := invocationmemory.New()
	put := func(repo invocation.Repository, id string, status agentsv1.InvocationStatus) {
		t.Helper()
		if err := repo.Save(t.Context(), &agentsv1.Invocation{
			Id: id, WorkspaceId: "ws-1", Status: status, StartedAt: timestamppb.Now(),
		}); err != nil {
			t.Fatalf("Save(%s): %v", id, err)
		}
	}
	earlier := store.WithOwner("earlier-process")
	put(earlier, "stale-running", agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING)
	put(earlier, "stale-queued", agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED)
	put(store, "legacy-young", agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING)
	put(earlier, "ok", agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED)

	startInvocationSweep(t.Context(), store.WithOwner(p.instanceID), p, invocation.LegacyStaleAge)

	for id, want := range map[string]agentsv1.InvocationStatus{
		"stale-running": agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED,
		"stale-queued":  agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED,
		"legacy-young":  agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED,
		"ok":            agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
	} {
		inv, err := store.GetAcrossWorkspaces(t.Context(), id)
		if err != nil {
			t.Fatalf("Get(%s): %v", id, err)
		}
		if inv.GetStatus() != want {
			t.Fatalf("%s is %v, want %v", id, inv.GetStatus(), want)
		}
		if want == agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED && inv.GetFinishedAt() == nil {
			t.Fatalf("%s has no finished_at", id)
		}
	}
}
