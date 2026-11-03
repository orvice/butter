package workstation_test

import (
	"context"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestReconcilerDurableRefreshCooldownAndRestart(t *testing.T) {
	ctx := context.Background()
	clock := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	repo := memory.New()
	repo.Clock = func() time.Time { return clock }
	p := &repotest.Provider{Items: []*v1.WorkStationItem{repotest.Item("1", "old snapshot", clock)}}
	s := repotest.NewService(repo, p)
	r, err := s.Create(ctx, "A", "main", "host", "key")
	if err != nil {
		t.Fatal(err)
	}
	r, err = s.Select(ctx, "A", r.ID, r.Revision, []string{"42"})
	if err != nil {
		t.Fatal(err)
	}
	// Another service instance consumes the saved initial intent.
	restarted := *s
	if err := restarted.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	rows, _, err := s.List(ctx, "A", ws.Query{}, "")
	if err != nil || len(rows) != 1 {
		t.Fatal("initial intent lost", err)
	}
	status, _, err := s.Refresh(ctx, "A", r.ID)
	if err != nil || status != "QUEUED" {
		t.Fatal("manual intent", status, err)
	}
	status, after, err := s.Refresh(ctx, "A", r.ID)
	if err != nil || status != "DEFERRED" || !after.Equal(clock.Add(time.Minute)) {
		t.Fatal("cooldown", status, after, err)
	}
	p.Failure = &ws.Failure{Category: "DEGRADED"}
	if err := restarted.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	rows, _, err = s.List(ctx, "A", ws.Query{}, "")
	if err != nil || len(rows) != 1 || rows[0].Title != "old snapshot" {
		t.Fatal("network failure replaced committed data")
	}
	clock = clock.Add(2 * time.Minute)
	status, _, err = s.Refresh(ctx, "A", r.ID)
	if err != nil || status != "QUEUED" {
		t.Fatal(status, err)
	}
	p.Failure = &ws.Failure{Category: "RATE_LIMITED", RetryAfter: 10 * time.Minute}
	if err := restarted.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	clock = clock.Add(2 * time.Minute)
	status, after, err = s.Refresh(ctx, "A", r.ID)
	if err != nil || status != "DEFERRED" || !after.After(clock.Add(7*time.Minute)) {
		t.Fatal("Retry-After not durable", status, after, err)
	}
}

type isolatedProvider struct{ repotest.Provider }

func (p *isolatedProvider) Scan(_ context.Context, token string, _ ws.Scope, now time.Time) (ws.Scan, error) {
	if token == "bad" {
		return ws.Scan{}, &ws.Failure{Category: "AUTH_INVALID"}
	}
	return ws.Scan{Coverage: "COMPLETE", Items: []*v1.WorkStationItem{repotest.Item("1", "healthy connection", now)}}, nil
}
func TestReconcilerFailureIsolatedAcrossConnections(t *testing.T) {
	ctx := context.Background()
	repo := memory.New()
	s := repotest.NewService(repo, &repotest.Provider{})
	s.Provider = &isolatedProvider{}
	for _, key := range []string{"bad", "good"} {
		r, err := s.Create(ctx, "A", key, "host", key)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.Select(ctx, "A", r.ID, r.Revision, []string{"42"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	rows, _, err := s.List(ctx, "A", ws.Query{}, "")
	if err != nil || len(rows) != 1 || rows[0].Title != "healthy connection" || len(rows[0].Sources) != 1 {
		t.Fatal("one failure blocked or polluted another", rows, err)
	}
}
func TestRetentionExactBoundary(t *testing.T) {
	now := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	item := repotest.Item("1", "closed", now)
	item.Terminal = true
	item.EndedAt = item.SourceUpdatedAt
	item.EndedAt.Seconds = now.Add(-90 * 24 * time.Hour).Unix()
	if !ws.Retained(item, now) {
		t.Fatal("90-day boundary excluded")
	}
	if ws.Retained(item, now.Add(time.Nanosecond)) {
		t.Fatal("query-time expiry not applied")
	}
}
