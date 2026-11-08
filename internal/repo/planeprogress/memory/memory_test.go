package planprogressmemory

import (
	"context"
	mongoprogress "go.orx.me/apps/butter/internal/repo/planeprogress/mongo"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
	"testing"
)

func TestQueryOwnershipLimitAndCAS(t *testing.T) {
	s := New()
	e := &plane.Event{WorkspaceID: "w", AppID: "a", InstallationID: "i", AgentRunID: "r", ActivityID: "x"}
	if err := s.EmitProgress(context.Background(), e, plane.ProgressEvent{Key: "k", Sequence: 2, State: plane.RunStateRunning, WorkflowID: "wf", DeliveryID: "d", Error: "secret=hidden"}); err != nil {
		t.Fatal(err)
	}
	s.EmitProgress(context.Background(), e, plane.ProgressEvent{Key: "k", Sequence: 1, State: plane.RunStateFailed})
	rows, _ := s.List(context.Background(), mongoprogress.Query{WorkspaceID: "w", AppID: "a", InstallationID: "i", RunID: "r", ActivityID: "x", Limit: 1})
	if len(rows) != 1 || rows[0].Sequence != 2 || rows[0].WorkflowID != "wf" || rows[0].DeliveryID != "d" {
		t.Fatalf("rows=%+v", rows)
	}
	if len(rows[0].Error) > 256 {
		t.Fatal("unbounded error")
	}
	other, _ := s.List(context.Background(), mongoprogress.Query{WorkspaceID: "other"})
	if len(other) != 0 {
		t.Fatal("cross workspace result")
	}
}
