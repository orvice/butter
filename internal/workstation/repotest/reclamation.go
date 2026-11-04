package repotest

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RunReclamation inspects physical storage, not just the authorized projection:
// invisible superseded rows are exactly the regression this suite must catch.
func RunReclamation(t *testing.T, factory func(*testing.T) (ws.Repository, func() int64)) {
	t.Run("late_prune_and_cleanup_failure", func(t *testing.T) { repo, count := factory(t); RunDelayedPruning(t, repo, count) })
	t.Run("repeated_full_publish_bounds_physical_rows", func(t *testing.T) {
		repo, count := factory(t)
		ctx := context.Background()
		s := NewService(repo, &Provider{})
		r := selectScope(t, s, create(t, s, "A", "main"))
		other := selectScope(t, s, create(t, s, "A", "other"))
		foreign := selectScope(t, s, create(t, s, "B", "foreign"))
		now, err := repo.Now(ctx)
		must(t, err)
		publish(t, s, claim(t, s, other), Item("1", "other source", now))
		publish(t, s, claim(t, s, foreign), Item("1", "other workspace", now))
		// Another publisher has staged data; retiring A/main must leave it alone.
		dueNow(t, repo, other)
		pending := claim(t, s, other)
		staged := Item("2", "in-flight", now)
		staged.Id = ws.ItemID("A", ws.GitHubSource(), "42", "ISSUE", "2")
		staged.ObservedAt = timestamppb.New(now)
		must(t, repo.Stage(ctx, []ws.Observation{{WorkspaceID: "A", ConnectionID: other.ID, ScopeID: "42", Snapshot: pending.Token, Item: staged}}))
		for round := 0; round < 4; round++ {
			dueNow(t, repo, r)
			a := claim(t, s, r)
			items := make([]*v1.WorkStationItem, 1001)
			for i := range items {
				items[i] = Item(fmt.Sprint(i+1), fmt.Sprintf("round %d", round), now)
			}
			publish(t, s, a, items...)
			if got := count(); got != 1004 {
				t.Fatalf("round %d: physical rows=%d; want current 1001 + 2 unrelated + 1 staging", round, got)
			}
		}

	})
	t.Run("collect_drains_multiple_batches_and_preserves_references", func(t *testing.T) {
		repo, count := factory(t)
		ctx := context.Background()
		now, err := repo.Now(ctx)
		must(t, err)
		r := ws.Record{WorkspaceID: "A", ID: "paused", Version: 1, Revision: 1, Paused: true, Credential: ws.Credential{Ciphertext: "encrypted"}, Scopes: []ws.Scope{{ID: "42", Snapshot: "current", Hidden: true}}}
		must(t, repo.Create(ctx, r))
		obs := []ws.Observation{}
		for i := 0; i < 2501; i++ {
			item := Item(fmt.Sprint(i+1), "abandoned", now)
			item.Id = fmt.Sprint(i)
			item.ObservedAt = timestamppb.New(now.Add(-25 * time.Hour))
			obs = append(obs, ws.Observation{WorkspaceID: "A", ConnectionID: r.ID, ScopeID: "42", Snapshot: "abandoned", Item: item})
		}
		current := Item("9999", "current but hidden", now)
		current.Id = "current"
		current.ObservedAt = timestamppb.New(now.Add(-25 * time.Hour))
		pending := Item("9998", "recent staging", now)
		pending.Id = "pending"
		pending.ObservedAt = timestamppb.New(now)
		obs = append(obs, ws.Observation{WorkspaceID: "A", ConnectionID: r.ID, ScopeID: "42", Snapshot: "current", Item: current}, ws.Observation{WorkspaceID: "A", ConnectionID: r.ID, ScopeID: "42", Snapshot: "pending", Item: pending})
		must(t, repo.Stage(ctx, obs))
		must(t, repo.Collect(ctx))
		if got := count(); got != 2 {
			t.Fatalf("after one GC pass physical rows=%d, want current reference + recent staging", got)
		}
	})
}
func dueNow(t *testing.T, repo ws.Repository, r ws.Record) {
	t.Helper()
	ctx := context.Background()
	old, err := repo.Get(ctx, r.WorkspaceID, r.ID)
	must(t, err)
	next := old
	next.Scopes = append([]ws.Scope(nil), old.Scopes...)
	now, err := repo.Now(ctx)
	must(t, err)
	next.Scopes[0].Due = now
	next.Version++
	ok, err := repo.CAS(ctx, old, &next, nil)
	must(t, err)
	if !ok {
		t.Fatal("fixture CAS conflict")
	}
}

type pruneRequest struct{ workspace, connection, scope, snapshot string }
type deferredPruner struct {
	ws.Repository
	requests []pruneRequest
	fail     bool
}

func (r *deferredPruner) PruneSnapshot(_ context.Context, w, c, s, snapshot string) error {
	r.requests = append(r.requests, pruneRequest{w, c, s, snapshot})
	if r.fail {
		return fmt.Errorf("synthetic cleanup failure")
	}
	return nil
}
func RunDelayedPruning(t *testing.T, repo ws.Repository, count func() int64) {
	ctx := context.Background()
	wrapped := &deferredPruner{Repository: repo}
	p := &Provider{}
	s := NewService(wrapped, p)
	r := selectScope(t, s, create(t, s, "A", "main"))
	now, err := repo.Now(ctx)
	must(t, err)
	for round := 0; round < 3; round++ {
		dueNow(t, repo, r)
		publish(t, s, claim(t, s, r), Item("1", fmt.Sprintf("round %d", round), now))
	}
	if len(wrapped.requests) != 2 {
		t.Fatal("expected both predecessor retirements")
	}
	// Execute callbacks backwards, after both successors are already committed.
	for i := len(wrapped.requests) - 1; i >= 0; i-- {
		v := wrapped.requests[i]
		must(t, repo.PruneSnapshot(ctx, v.workspace, v.connection, v.scope, v.snapshot))
	}
	if got := count(); got != 1 {
		t.Fatalf("late pruning retained predecessors or deleted successor: %d", got)
	}
	got := list(t, s, "A", ws.Query{})
	if len(got) != 1 || got[0].Title != "round 2" {
		t.Fatal("late pruner deleted newer publisher")
	}
	// Cleanup failure occurs after commit; it must surface without downgrading
	// the successfully published snapshot to a failed scan.
	wrapped.fail = true
	p.Items = []*v1.WorkStationItem{Item("1", "published despite cleanup failure", now)}
	dueNow(t, repo, r)
	err = s.RunOnce(ctx)
	if !errors.Is(err, ws.ErrCleanupPending) {
		t.Fatal("cleanup failure not reported", err)
	}
	stored, err := repo.Get(ctx, "A", r.ID)
	must(t, err)
	if stored.Scopes[0].Health != "HEALTHY" {
		t.Fatal("successful publication was marked failed")
	}
	got = list(t, s, "A", ws.Query{})
	if len(got) != 1 || got[0].Title != "published despite cleanup failure" {
		t.Fatal("cleanup failure rolled back publication")
	}
}
