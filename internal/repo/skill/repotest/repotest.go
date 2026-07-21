// Package repotest is a conformance suite for skill.Repository
// implementations. The memory and mongo backends must behave identically
// through this seam so SkillService and the skill Source work against either
// unchanged (issue #153).
package repotest

import (
	"bytes"
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

	runResources(t, factory)
}

func newResource(path, contentType string) *agentsv1.SkillResource {
	return &agentsv1.SkillResource{Path: path, ContentType: contentType}
}

func putResource(t *testing.T, repo skillrepo.Repository, ws, skill, path string, content []byte) *agentsv1.SkillResource {
	t.Helper()
	res, err := repo.PutResource(context.Background(), ws, skill, newResource(path, "application/octet-stream"), content)
	if err != nil {
		t.Fatalf("PutResource %s/%s/%s: %v", ws, skill, path, err)
	}
	return res
}

// runResources covers the skill resource slice of the Repository seam
// (issue #154): binary-safe round-trips, metadata-listed paths, overwrite in
// place, and DeleteSkill cascading to resource content.
func runResources(t *testing.T, factory Factory) {
	binary := []byte{0x00, 0xff, 0x1f, 0x8b, 'P', 'N', 'G', 0x00, 0x7f}

	t.Run("ResourcePutThenGetRoundTripsBinary", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		created := putResource(t, repo, "ws-a", "pdf-report", "assets/logo.png", binary)
		if created.GetSizeBytes() != int64(len(binary)) {
			t.Fatalf("expected size %d, got %d", len(binary), created.GetSizeBytes())
		}
		if created.GetCreatedAt() == nil || created.GetUpdatedAt() == nil {
			t.Fatalf("expected timestamps set, got %v / %v", created.GetCreatedAt(), created.GetUpdatedAt())
		}

		meta, content, err := repo.GetResource(context.Background(), "ws-a", "pdf-report", "assets/logo.png")
		if err != nil {
			t.Fatalf("GetResource: %v", err)
		}
		if !bytes.Equal(content, binary) {
			t.Fatalf("content did not round-trip: got %v want %v", content, binary)
		}
		if meta.GetPath() != "assets/logo.png" || meta.GetContentType() != "application/octet-stream" {
			t.Fatalf("metadata did not round-trip: %v", meta)
		}
	})

	t.Run("ResourcePutOverwritesInPlace", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		first := putResource(t, repo, "ws-a", "pdf-report", "references/api.md", []byte("v1"))
		putResource(t, repo, "ws-a", "pdf-report", "references/api.md", []byte("v2 longer"))

		meta, content, err := repo.GetResource(context.Background(), "ws-a", "pdf-report", "references/api.md")
		if err != nil {
			t.Fatalf("GetResource: %v", err)
		}
		if string(content) != "v2 longer" {
			t.Fatalf("expected overwrite, got %q", content)
		}
		if meta.GetSizeBytes() != int64(len("v2 longer")) {
			t.Fatalf("expected size updated to %d, got %d", len("v2 longer"), meta.GetSizeBytes())
		}
		// created_at is stamped once and preserved across overwrites.
		if !meta.GetCreatedAt().AsTime().Equal(first.GetCreatedAt().AsTime()) {
			t.Fatalf("expected created_at preserved, got %v want %v", meta.GetCreatedAt(), first.GetCreatedAt())
		}

		resources, err := repo.ListResources(context.Background(), "ws-a", "pdf-report")
		if err != nil {
			t.Fatalf("ListResources: %v", err)
		}
		if len(resources) != 1 {
			t.Fatalf("expected 1 resource after overwrite, got %d", len(resources))
		}
	})

	t.Run("ResourcePutOnMissingSkillIsNotFound", func(t *testing.T) {
		repo := factory(t)
		_, err := repo.PutResource(context.Background(), "ws-a", "absent", newResource("assets/a.txt", ""), []byte("x"))
		if !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ResourceGetMissingIsNotFound", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		if _, _, err := repo.GetResource(context.Background(), "ws-a", "pdf-report", "assets/absent.txt"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ResourceListSortedAndScopedToSkill", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		create(t, repo, "ws-a", "other-skill")
		create(t, repo, "ws-b", "pdf-report")
		putResource(t, repo, "ws-a", "pdf-report", "scripts/run.sh", []byte("#!/bin/sh"))
		putResource(t, repo, "ws-a", "pdf-report", "assets/logo.png", binary)
		putResource(t, repo, "ws-a", "other-skill", "assets/other.png", binary)
		putResource(t, repo, "ws-b", "pdf-report", "assets/foreign.png", binary)

		resources, err := repo.ListResources(context.Background(), "ws-a", "pdf-report")
		if err != nil {
			t.Fatalf("ListResources: %v", err)
		}
		if len(resources) != 2 || resources[0].GetPath() != "assets/logo.png" || resources[1].GetPath() != "scripts/run.sh" {
			paths := make([]string, 0, len(resources))
			for _, r := range resources {
				paths = append(paths, r.GetPath())
			}
			t.Fatalf("expected [assets/logo.png scripts/run.sh], got %v", paths)
		}
	})

	t.Run("ResourceListOnMissingSkillIsNotFound", func(t *testing.T) {
		repo := factory(t)
		if _, err := repo.ListResources(context.Background(), "ws-a", "absent"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ResourceDeleteRemovesExactlyOne", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		putResource(t, repo, "ws-a", "pdf-report", "assets/keep.png", binary)
		putResource(t, repo, "ws-a", "pdf-report", "assets/drop.png", binary)

		if err := repo.DeleteResource(context.Background(), "ws-a", "pdf-report", "assets/drop.png"); err != nil {
			t.Fatalf("DeleteResource: %v", err)
		}
		if _, _, err := repo.GetResource(context.Background(), "ws-a", "pdf-report", "assets/drop.png"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected deleted resource gone, got %v", err)
		}
		if _, _, err := repo.GetResource(context.Background(), "ws-a", "pdf-report", "assets/keep.png"); err != nil {
			t.Fatalf("expected sibling resource kept, got %v", err)
		}
	})

	t.Run("ResourceDeleteMissingIsNotFound", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		if err := repo.DeleteResource(context.Background(), "ws-a", "pdf-report", "assets/absent.png"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("DeleteSkillCascadesToResources", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", "pdf-report")
		putResource(t, repo, "ws-a", "pdf-report", "assets/logo.png", binary)
		putResource(t, repo, "ws-a", "pdf-report", "references/api.md", []byte("api"))

		if err := repo.Delete(context.Background(), "ws-a", "pdf-report"); err != nil {
			t.Fatalf("Delete skill: %v", err)
		}

		// Names are reusable after delete; the recreated skill must not see
		// the old skill's resources.
		create(t, repo, "ws-a", "pdf-report")
		resources, err := repo.ListResources(context.Background(), "ws-a", "pdf-report")
		if err != nil {
			t.Fatalf("ListResources after recreate: %v", err)
		}
		if len(resources) != 0 {
			t.Fatalf("expected no resources after cascade delete, got %v", resources)
		}
		if _, _, err := repo.GetResource(context.Background(), "ws-a", "pdf-report", "assets/logo.png"); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected resource content gone after cascade, got %v", err)
		}
	})
}
