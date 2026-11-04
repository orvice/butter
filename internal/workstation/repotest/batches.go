package repotest

import (
	"context"
	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"strconv"
	"testing"
	"time"
)

type pagedProvider struct {
	Provider
	calls map[int]int
	fail  bool
}

func (p *pagedProvider) ScanBatch(_ context.Context, _ string, _ ws.Scope, now time.Time, cursor string) (ws.ScanBatch, error) {
	page := 0
	if cursor != "" {
		page, _ = strconv.Atoi(cursor)
	}
	p.calls[page]++
	if page == 2 && p.fail {
		p.fail = false
		return ws.ScanBatch{}, &ws.Failure{Category: "DEGRADED"}
	}
	if page == 12 {
		return ws.ScanBatch{Complete: true, Coverage: "COMPLETE"}, nil
	}
	return ws.ScanBatch{Items: []*v1.WorkStationItem{Item(strconv.Itoa(page+1), "resumable", now)}, Next: strconv.Itoa(page + 1), Coverage: "COMPLETE"}, nil
}

func RunBatches(t *testing.T, repo ws.Repository, advance func(time.Duration)) {
	ctx := context.Background()
	p := &pagedProvider{calls: map[int]int{}, fail: true}
	s := NewService(repo, &Provider{})
	s.Provider = p
	r := selectScope(t, s, create(t, s, "A", "large"))
	must(t, s.RunOnce(ctx))
	current, err := repo.Get(ctx, "A", r.ID)
	must(t, err)
	if current.Scopes[0].Pending == nil || current.Scopes[0].Pending.Count != 2 || len(list(t, s, "A", ws.Query{})) != 0 {
		t.Fatal("failed second page published or lost registered prefix", current)
	}
	advance(6 * time.Minute)
	must(t, repo.Collect(ctx))
	// A new service instance only needs the shared repository and credential
	// seam; no process-local page cache participates in resuming the traversal.
	restarted := *s
	must(t, restarted.RunOnce(ctx))
	current, err = repo.Get(ctx, "A", r.ID)
	must(t, err)
	if current.Scopes[0].Pending == nil || current.Scopes[0].Pending.Count != 12 || current.Scopes[0].Health != "LAGGING" || len(list(t, s, "A", ws.Query{})) != 0 {
		t.Fatal("large scan falsely complete", current)
	}
	advance(6 * time.Minute)
	must(t, repo.Collect(ctx))
	must(t, restarted.RunOnce(ctx))
	items := list(t, s, "A", ws.Query{Limit: 100})
	current, err = repo.Get(ctx, "A", r.ID)
	must(t, err)
	if len(items) != 12 || current.Scopes[0].Pending != nil || current.Scopes[0].Checkpoint.IsZero() {
		t.Fatal("restart/GC lost prefix or completion checkpoint", len(items), current)
	}
	if p.calls[0] != 1 || p.calls[1] != 1 || p.calls[2] != 2 {
		t.Fatal("replayed completed prefix", p.calls)
	}
}
