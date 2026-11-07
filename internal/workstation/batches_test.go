package workstation_test

import (
	"context"
	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

type overlappingTerminalProvider struct{ repotest.Provider }

func (*overlappingTerminalProvider) ScanBatch(_ context.Context, _ string, _ ws.Scope, now time.Time, cursor string) (ws.ScanBatch, error) {
	if cursor == "" {
		return ws.ScanBatch{Coverage: "COMPLETE", Next: "terminal", Items: []*v1.WorkStationItem{repotest.Item("1", "old open", now.Add(-time.Hour)), repotest.Item("2", "old open", now.Add(-time.Hour))}}, nil
	}
	unknown := repotest.Item("1", "now terminal but no date", now)
	unknown.Terminal = true
	unknown.Status = "DONE"
	expired := repotest.Item("2", "edited after old closure", now)
	expired.Terminal = true
	expired.Status = "DONE"
	expired.EndedAt = timestamppb.New(now.Add(-91 * 24 * time.Hour))
	return ws.ScanBatch{Complete: true, Coverage: "PARTIAL", Unresolved: 1, Items: []*v1.WorkStationItem{unknown, expired}}, nil
}
func TestLatestTerminalMarkersCannotResurrectOlderStagedOpenRows(t *testing.T) {
	ctx := context.Background()
	repo := memory.New()
	s := repotest.NewService(repo, &repotest.Provider{})
	s.Provider = &overlappingTerminalProvider{}
	r, err := s.Create(ctx, "A", "overlap", "host", "synthetic")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Select(ctx, "A", r.ID, r.Revision, []string{"42"}); err != nil {
		t.Fatal(err)
	}
	if err := s.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	items, _, err := s.List(ctx, "A", ws.Query{Status: "all"}, "")
	if err != nil || len(items) != 0 {
		t.Fatal("resurrected older open observation", items, err)
	}
	current, err := repo.Get(ctx, "A", r.ID)
	if err != nil || current.Scopes[0].Coverage != "PARTIAL" || current.Scopes[0].Unresolved != 1 {
		t.Fatal("unresolved marker not counted exactly once", current, err)
	}
}
