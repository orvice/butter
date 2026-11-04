package repotest

import (
	"context"
	"errors"
	ws "go.orx.me/apps/butter/internal/workstation"
	"testing"
	"time"
)

func RunFoundation(t *testing.T, newRepo func(*testing.T) ws.Repository) {
	t.Run("facets_follow_source_revocation_and_query_identity", func(t *testing.T) {
		repo := newRepo(t)
		s := NewService(repo, &Provider{})
		ctx := context.Background()
		r1 := selectScope(t, s, create(t, s, "A", "older"))
		r2 := selectScope(t, s, create(t, s, "A", "newer"))
		now, _ := repo.Now(ctx)
		old := Item("1", "safe", now.Add(-time.Hour))
		old.Assignees = []string{"safe-a", "safe-b"}
		old.ProjectId = "project-safe"
		old.ProjectName = "Safe"
		publish(t, s, claim(t, s, r1), old)
		newer := Item("1", "private", now)
		newer.Assignees = []string{"private"}
		newer.ProjectId = "project-private"
		newer.ProjectName = "Private"
		publish(t, s, claim(t, s, r2), newer)
		options, _, err := s.FilterOptions(ctx, "A", ws.Query{}, "assignee", "", "", 50)
		must(t, err)
		if len(options) != 1 || options[0].Id != "private" {
			t.Fatalf("source choice: %v", options)
		}
		_, err = s.Select(ctx, "A", r2.ID, r2.Revision, nil)
		must(t, err)
		options, next, err := s.FilterOptions(ctx, "A", ws.Query{}, "assignee", "", "", 1)
		must(t, err)
		if len(options) != 1 || options[0].Id != "safe-a" || next == "" {
			t.Fatalf("fallback/page: %v %q", options, next)
		}
		second, _, err := s.FilterOptions(ctx, "A", ws.Query{}, "assignee", "", next, 1)
		must(t, err)
		if len(second) != 1 || second[0].Id != "safe-b" {
			t.Fatal(second)
		}
		for _, q := range []ws.Query{{Provider: "gitlab"}, {ConnectionID: r2.ID}} {
			got, _, err := s.FilterOptions(ctx, "A", q, "assignee", "", "", 50)
			must(t, err)
			if len(got) != 0 {
				t.Fatal("hidden facet", got)
			}
		}
		_, _, err = s.FilterOptions(ctx, "B", ws.Query{}, "assignee", "", next, 1)
		if !errors.Is(err, ws.ErrInvalid) {
			t.Fatal("crossworkspace cursor", err)
		}
		_, _, err = s.FilterOptions(ctx, "A", ws.Query{}, "assignee", "", next, 2)
		if !errors.Is(err, ws.ErrInvalid) {
			t.Fatal("pagesize cursor", err)
		}
		options, _, err = s.FilterOptions(ctx, "A", ws.Query{}, "project", "Private", "", 50)
		must(t, err)
		if len(options) != 0 {
			t.Fatal("revoked metadata leaked", options)
		}
		if len(list(t, s, "A", ws.Query{ProjectID: "project-private"})) != 0 || len(list(t, s, "A", ws.Query{ProjectID: "project-safe", Provider: "github"})) != 1 {
			t.Fatal("provider/project filter")
		}
	})
	t.Run("host_registry_revision_and_reference_guard", func(t *testing.T) {
		repo := newRepo(t)
		ctx := context.Background()
		cloud, err := repo.GetHost(ctx, ws.PlaneCloudHost)
		must(t, err)
		if cloud.Profile != "plane-v2" {
			t.Fatal(cloud)
		}
		h := ws.Host{ID: "private-plane", Name: "Approved", APIBaseURL: "https://plane.example", WebBaseURL: "https://plane.example", Profile: "plane-v2", Revision: 1}
		must(t, repo.SaveHost(ctx, h, 0))
		if err := repo.SaveHost(ctx, h, 0); !errors.Is(err, ws.ErrConflict) {
			t.Fatal("duplicate", err)
		}
		h.Revision = 2
		must(t, repo.SaveHost(ctx, h, 1))
		if err := repo.SaveHost(ctx, h, 1); !errors.Is(err, ws.ErrConflict) {
			t.Fatal("stale revision", err)
		}
		must(t, repo.Create(ctx, ws.Record{WorkspaceID: "A", ID: "ref", HostID: h.ID, Version: 1}))
		if err := repo.DeleteHost(ctx, h.ID, 2); !errors.Is(err, ws.ErrConflict) {
			t.Fatal("deleted referenced host", err)
		}
		if err := repo.DeleteHost(ctx, ws.PlaneCloudHost, 1); !errors.Is(err, ws.ErrInvalid) {
			t.Fatal("deleted built-in", err)
		}
	})
	t.Run("budgets_shared_across_workspace_and_credentials", func(t *testing.T) {
		repo := newRepo(t)
		ctx := context.Background()
		a := ws.Record{Source: ws.GitHubSource(), PrincipalID: "same", WorkspaceID: "A", ID: "a"}
		b := a
		b.WorkspaceID = "B"
		b.ID = "b"
		b.Credential = ws.Credential{Ciphertext: "different"}
		ia, aa := ws.BudgetKeys(a)
		ib, ab := ws.BudgetKeys(b)
		if ia != ib || aa != ab {
			t.Fatal("connection multiplied budget")
		}
		wait, err := repo.ReserveBudget(ctx, aa, time.Minute)
		must(t, err)
		if wait != 0 {
			t.Fatal(wait)
		}
		wait, err = repo.ReserveBudget(ctx, ab, time.Minute)
		must(t, err)
		if wait <= 0 {
			t.Fatal("second key bypassed budget")
		}
		must(t, repo.DeferBudget(ctx, aa, 5*time.Minute))
		wait, err = repo.ReserveBudget(ctx, ab, time.Millisecond)
		must(t, err)
		if wait < 4*time.Minute {
			t.Fatal("retry-after bypass", wait)
		}
		b.PrincipalID = "other"
		_, other := ws.BudgetKeys(b)
		wait, err = repo.ReserveBudget(ctx, other, time.Millisecond)
		must(t, err)
		if wait != 0 {
			t.Fatal("unrelated account blocked")
		}
	})
}
