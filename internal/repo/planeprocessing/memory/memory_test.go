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
