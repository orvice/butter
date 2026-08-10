package app

import (
	"context"
	"testing"
	"time"

	repobindingmemory "go.orx.me/apps/butter/internal/repo/repobinding/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type recordingSyncTrigger struct {
	called chan string
}

func (t *recordingSyncTrigger) TriggerSyncAndPublish(_ context.Context, workspaceID string) error {
	t.called <- workspaceID
	return nil
}

func TestReconcilerRunsImmediatelyOnStart(t *testing.T) {
	ctx := context.Background()
	repo := repobindingmemory.New()
	if _, err := repo.Put(ctx, "ws-a", &agentsv1.WorkspaceRepoBinding{}); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetCredential(ctx, "ws-a", "encrypted"); err != nil {
		t.Fatal(err)
	}
	trigger := &recordingSyncTrigger{called: make(chan string, 1)}
	reconciler := NewReconciler(repo, trigger, time.Hour)
	reconciler.Start(ctx)
	t.Cleanup(reconciler.Stop)

	select {
	case workspaceID := <-trigger.called:
		if workspaceID != "ws-a" {
			t.Fatalf("workspace = %q, want ws-a", workspaceID)
		}
	case <-time.After(time.Second):
		t.Fatal("reconciler did not run immediately")
	}
}
