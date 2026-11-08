package memory

import (
	"context"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"testing"
)

func TestStoreScopesAndDeduplicates(t *testing.T) {
	s := New()
	r := &agentsv1.PlaneProcessingRecord{WorkspaceId: "ws", AppId: "app", InstallationId: "inst", RunId: "run", ActivityId: "act", EventAction: "prompted"}
	created, e := s.Create(context.Background(), r)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Create(context.Background(), r); e != planeprocessing.ErrDuplicate {
		t.Fatalf("duplicate error=%v", e)
	}
	if _, e = s.Get(context.Background(), "other", created.Id); e != planeprocessing.ErrNotFound {
		t.Fatalf("cross workspace error=%v", e)
	}
	rows, e := s.List(context.Background(), planeprocessing.Filter{WorkspaceID: "ws", AppID: "app", Limit: 10})
	if e != nil || len(rows) != 1 {
		t.Fatalf("rows=%d err=%v", len(rows), e)
	}
}

func TestStoreUpdatesStatusByBusinessKey(t *testing.T) {
	s := New()
	r := &agentsv1.PlaneProcessingRecord{WorkspaceId: "ws", AppId: "app", InstallationId: "inst", RunId: "run", ActivityId: "act", EventAction: "prompted"}
	if _, err := s.Create(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateStatus(context.Background(), "ws", "app", "inst", "run", "act", "prompted", agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED, "box unavailable"); err != nil {
		t.Fatal(err)
	}
	rows, err := s.List(context.Background(), planeprocessing.Filter{WorkspaceID: "ws"})
	if err != nil || len(rows) != 1 || rows[0].GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED || rows[0].GetFailureReason() != "box unavailable" {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}
