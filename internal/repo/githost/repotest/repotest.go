// Package repotest is a conformance suite for githost.Repository
// implementations: the memory and mongo backends must behave identically
// through this seam (issue #214).
package repotest

import (
	"context"
	"errors"
	"testing"

	githostrepo "go.orx.me/apps/butter/internal/repo/githost"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory returns a fresh, empty repository for one subtest.
type Factory func(t *testing.T) githostrepo.Repository

func newHost(id, name string) *agentsv1.GitHost {
	return &agentsv1.GitHost{
		Id:         id,
		Name:       name,
		Kind:       agentsv1.GitHostKind_GIT_HOST_KIND_GITHUB,
		ApiBaseUrl: "https://api.github.com",
		WebBaseUrl: "https://github.com",
	}
}

func create(t *testing.T, repo githostrepo.Repository, id, name string) *agentsv1.GitHost {
	t.Helper()
	created, err := repo.Create(context.Background(), newHost(id, name))
	if err != nil {
		t.Fatalf("Create %s: %v", id, err)
	}
	return created
}

// Run executes the conformance suite against the given repository factory.
func Run(t *testing.T, factory Factory) {
	ctx := context.Background()

	t.Run("CreateThenGetRoundTrip", func(t *testing.T) {
		repo := factory(t)
		created := create(t, repo, "gh-1", "GitHub.com")
		if created.GetCreatedAt() == nil || created.GetUpdatedAt() == nil {
			t.Fatalf("expected timestamps set, got %v / %v", created.GetCreatedAt(), created.GetUpdatedAt())
		}
		got, err := repo.Get(ctx, "gh-1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.GetName() != "GitHub.com" || got.GetKind() != agentsv1.GitHostKind_GIT_HOST_KIND_GITHUB {
			t.Fatalf("host did not round-trip: %v", got)
		}
		if got.GetApiBaseUrl() != "https://api.github.com" || got.GetWebBaseUrl() != "https://github.com" {
			t.Fatalf("urls did not round-trip: %v", got)
		}
	})

	t.Run("GetMissingIsNotFound", func(t *testing.T) {
		repo := factory(t)
		if _, err := repo.Get(ctx, "nope"); !errors.Is(err, githostrepo.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("DuplicateCreateFails", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "gh-1", "GitHub.com")
		if _, err := repo.Create(ctx, newHost("gh-1", "Other")); !errors.Is(err, githostrepo.ErrAlreadyExists) {
			t.Fatalf("err = %v, want ErrAlreadyExists", err)
		}
	})

	t.Run("ListSortedByName", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "gl-1", "Zeta GitLab")
		create(t, repo, "gh-1", "Alpha GitHub")
		hosts, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(hosts) != 2 || hosts[0].GetName() != "Alpha GitHub" || hosts[1].GetName() != "Zeta GitLab" {
			t.Fatalf("unexpected list: %v", hosts)
		}
	})

	t.Run("UpdateRoundTripAndPreservesCreatedAt", func(t *testing.T) {
		repo := factory(t)
		created := create(t, repo, "gh-1", "GitHub.com")
		mod := newHost("gh-1", "GitHub Enterprise")
		mod.ApiBaseUrl = "https://ghe.example.com/api/v3"
		updated, err := repo.Update(ctx, mod)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.GetName() != "GitHub Enterprise" || updated.GetApiBaseUrl() != "https://ghe.example.com/api/v3" {
			t.Fatalf("update did not apply: %v", updated)
		}
		if !updated.GetCreatedAt().AsTime().Equal(created.GetCreatedAt().AsTime()) {
			t.Fatalf("created_at changed: %v -> %v", created.GetCreatedAt(), updated.GetCreatedAt())
		}
	})

	t.Run("UpdateMissingIsNotFound", func(t *testing.T) {
		repo := factory(t)
		if _, err := repo.Update(ctx, newHost("nope", "X")); !errors.Is(err, githostrepo.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("DeleteThenGetIsNotFound", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "gh-1", "GitHub.com")
		if err := repo.Delete(ctx, "gh-1"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.Get(ctx, "gh-1"); !errors.Is(err, githostrepo.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if err := repo.Delete(ctx, "gh-1"); !errors.Is(err, githostrepo.ErrNotFound) {
			t.Fatalf("second delete err = %v, want ErrNotFound", err)
		}
	})

	t.Run("MutatingResultDoesNotAffectStore", func(t *testing.T) {
		repo := factory(t)
		created := create(t, repo, "gh-1", "GitHub.com")
		created.Name = "mutated"
		got, err := repo.Get(ctx, "gh-1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.GetName() != "GitHub.com" {
			t.Fatalf("store was mutated through returned pointer: %v", got)
		}
	})
}
