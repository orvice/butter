package memory

import (
	"context"
	repo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"testing"
)

func TestGetScopedIsolation(t *testing.T) {
	s := New()
	ctx := context.Background()
	_, e := s.Upsert(ctx, "w", "a", &agentsv1.PlaneInstallation{Id: "row", AppId: "a", AppInstallationId: "inst"}, repo.Tokens{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.GetScoped(ctx, "w", "a", "inst"); e != nil {
		t.Fatal(e)
	}
	for _, x := range [][3]string{{"other", "a", "inst"}, {"w", "other", "inst"}, {"w", "a", "other"}, {"", "a", "inst"}} {
		if _, e = s.GetScoped(ctx, x[0], x[1], x[2]); e != repo.ErrNotFound {
			t.Fatalf("scope=%v err=%v", x, e)
		}
	}
}
