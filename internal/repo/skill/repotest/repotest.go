// Package repotest is a conformance suite for skill.Repository
// implementations. The memory and mongo backends must behave identically
// through this seam so SkillService and the skill Source work against either
// unchanged (issue #153).
package repotest

import (
	"context"
	"errors"
	"testing"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory returns a fresh, empty repository for one subtest.
type Factory func(t *testing.T) skillrepo.Repository

func skillMD(name string) string {
	return "---\nname: " + name + "\ndescription: test skill\n---\n# " + name + "\n"
}

func newSkill(name string) *agentsv1.Skill {
	return &agentsv1.Skill{
		Name:         name,
		Description:  "test skill",
		License:      "MIT",
		Metadata:     map[string]string{"author": "butter"},
		AllowedTools: []string{"agent_files_read_file"},
		SizeBytes:    int64(len(skillMD(name))),
	}
}

func create(t *testing.T, repo skillrepo.Repository, ws, name string) *agentsv1.Skill {
	t.Helper()
	created, err := repo.Create(context.Background(), ws, newSkill(name), skillMD(name))
	if err != nil {
		t.Fatalf("Create %s/%s: %v", ws, name, err)
	}
	return created
}

// Run executes the conformance suite against the given repository factory.
func Run(t *testing.T, factory Factory) {
	t.Run("CreateThenGetRoundTrip", func(t *testing.T) {
		repo := factory(t)
		created := create(t, repo, "ws-a", "pdf-report")
		if created.GetWorkspaceId() != "ws-a" {
			t.Fatalf("expected workspace stamped, got %q", created.GetWorkspaceId())
		}
		if created.GetCreatedAt() == nil || created.GetUpdatedAt() == nil {
			t.Fatalf("expected timestamps set, got %v / %v", created.GetCreatedAt(), created.GetUpdatedAt())
		}

		got, err := repo.Get(context.Background(), "ws-a", "pdf-report")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.GetDescription() != "test skill" || got.GetLicense() != "MIT" {
			t.Fatalf("metadata did not round-trip: %v", got)
		}
		if got.GetMetadata()["author"] != "butter" {
			t.Fatalf("metadata map did not round-trip: %v", got.GetMetadata())
		}
		if len(got.GetAllowedTools()) != 1 || got.GetAllowedTools()[0] != "agent_files_read_file" {
			t.Fatalf("allowed tools did not round-trip: %v", got.GetAllowedTools())
		}
	})

	t.Run("GetSkillMDRoundTrip", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		md, err := repo.GetSkillMD(context.Background(), "ws-a", "pdf-report")
		if err != nil {
			t.Fatalf("GetSkillMD: %v", err)
		}
		if md != skillMD("pdf-report") {
			t.Fatalf("SKILL.md did not round-trip:\n%s", md)
		}
	})

	t.Run("GetMissingIsNotFound", func(t *testing.T) {
		repo := factory(t)
		if _, err := repo.Get(context.Background(), "ws-a", "absent"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("Get: expected ErrNotFound, got %v", err)
		}
		if _, err := repo.GetSkillMD(context.Background(), "ws-a", "absent"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("GetSkillMD: expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ListScopedToWorkspaceAndSorted", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "zeta")
		create(t, repo, "ws-a", "alpha")
		create(t, repo, "ws-b", "other")

		skills, err := repo.List(context.Background(), "ws-a")
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(skills) != 2 || skills[0].GetName() != "alpha" || skills[1].GetName() != "zeta" {
			names := make([]string, 0, len(skills))
			for _, sk := range skills {
				names = append(names, sk.GetName())
			}
			t.Fatalf("expected [alpha zeta], got %v", names)
		}
	})

	t.Run("CreateDuplicateSameWorkspaceFails", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		_, err := repo.Create(context.Background(), "ws-a", newSkill("pdf-report"), skillMD("pdf-report"))
		if !errors.Is(err, skillrepo.ErrAlreadyExists) {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("SameNameAcrossWorkspacesCoexists", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		create(t, repo, "ws-b", "pdf-report")

		for _, ws := range []string{"ws-a", "ws-b"} {
			got, err := repo.Get(context.Background(), ws, "pdf-report")
			if err != nil {
				t.Fatalf("Get %s: %v", ws, err)
			}
			if got.GetWorkspaceId() != ws {
				t.Fatalf("expected workspace %q, got %q", ws, got.GetWorkspaceId())
			}
		}
	})

	t.Run("UpdatePersistsAndKeepsCreatedAt", func(t *testing.T) {
		repo := factory(t)
		created := create(t, repo, "ws-a", "pdf-report")

		updated := newSkill("pdf-report")
		updated.Description = "updated description"
		newMD := "---\nname: pdf-report\ndescription: updated description\n---\nnew body\n"
		got, err := repo.Update(context.Background(), "ws-a", updated, newMD)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got.GetDescription() != "updated description" {
			t.Fatalf("update did not persist description: %v", got)
		}
		if !got.GetCreatedAt().AsTime().Equal(created.GetCreatedAt().AsTime()) {
			t.Fatalf("expected created_at preserved, got %v want %v", got.GetCreatedAt(), created.GetCreatedAt())
		}

		md, err := repo.GetSkillMD(context.Background(), "ws-a", "pdf-report")
		if err != nil {
			t.Fatalf("GetSkillMD: %v", err)
		}
		if md != newMD {
			t.Fatalf("updated SKILL.md did not persist:\n%s", md)
		}
	})

	t.Run("UpdateMissingIsNotFound", func(t *testing.T) {
		repo := factory(t)
		_, err := repo.Update(context.Background(), "ws-a", newSkill("absent"), skillMD("absent"))
		if !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("DeleteRemovesSkill", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		if err := repo.Delete(context.Background(), "ws-a", "pdf-report"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.Get(context.Background(), "ws-a", "pdf-report"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("Get after delete: expected ErrNotFound, got %v", err)
		}
		if _, err := repo.GetSkillMD(context.Background(), "ws-a", "pdf-report"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("GetSkillMD after delete: expected ErrNotFound, got %v", err)
		}
		// Names are the sole identifier (ADR 0004): a deleted name is reusable.
		create(t, repo, "ws-a", "pdf-report")
	})

	t.Run("DeleteMissingIsNotFound", func(t *testing.T) {
		repo := factory(t)
		if err := repo.Delete(context.Background(), "ws-a", "absent"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
