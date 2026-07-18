package skilltool

import (
	"context"
	"errors"
	"sync"
	"testing"

	adkskill "google.golang.org/adk/v2/tool/skilltoolset/skill"

	skillmemory "go.orx.me/apps/butter/internal/repo/skill/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// seedSkill stores a skill with a minimal spec-valid SKILL.md document.
func seedSkill(t *testing.T, repo *skillmemory.Store, workspaceID, name, description, body string) {
	t.Helper()
	md := "---\nname: " + name + "\ndescription: " + description + "\n---\n" + body
	_, err := repo.Create(context.Background(), workspaceID, &agentsv1.Skill{
		Name:        name,
		Description: description,
	}, md)
	if err != nil {
		t.Fatalf("seed skill %q: %v", name, err)
	}
}

func TestLoadFrontmatter(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")
	seedSkill(t, repo, "ws-1", "hidden", "Not attached to the agent", "Secret.\n")

	src := NewSource(repo, "ws-1", []string{"alpha", "dangling"})

	fm, err := src.LoadFrontmatter(ctx, "alpha")
	if err != nil {
		t.Fatalf("LoadFrontmatter(alpha): %v", err)
	}
	if fm.Name != "alpha" || fm.Description != "Alpha skill" {
		t.Errorf("frontmatter = %+v, want name=alpha description=%q", fm, "Alpha skill")
	}

	// Outside the allowlist: invisible even though it exists in the workspace.
	if _, err := src.LoadFrontmatter(ctx, "hidden"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadFrontmatter(hidden) err = %v, want ErrSkillNotFound", err)
	}
	// Allowlisted but deleted from the repository: same sentinel.
	if _, err := src.LoadFrontmatter(ctx, "dangling"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadFrontmatter(dangling) err = %v, want ErrSkillNotFound", err)
	}
}

func TestLoadInstructionsReturnsBodyWithoutFrontmatter(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "# Alpha\n\nUse alpha wisely.\n")
	seedSkill(t, repo, "ws-1", "hidden", "Not attached to the agent", "Secret.\n")

	src := NewSource(repo, "ws-1", []string{"alpha", "dangling"})

	body, err := src.LoadInstructions(ctx, "alpha")
	if err != nil {
		t.Fatalf("LoadInstructions(alpha): %v", err)
	}
	if body != "# Alpha\n\nUse alpha wisely.\n" {
		t.Errorf("instructions = %q, want the markdown body without frontmatter", body)
	}

	if _, err := src.LoadInstructions(ctx, "hidden"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadInstructions(hidden) err = %v, want ErrSkillNotFound", err)
	}
	if _, err := src.LoadInstructions(ctx, "dangling"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadInstructions(dangling) err = %v, want ErrSkillNotFound", err)
	}
}

func TestLoadInstructionsSeesRepositoryEdits(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Old body.\n")

	src := NewSource(repo, "ws-1", []string{"alpha"})
	if body, err := src.LoadInstructions(ctx, "alpha"); err != nil || body != "Old body.\n" {
		t.Fatalf("LoadInstructions before edit = %q, %v", body, err)
	}

	md := "---\nname: alpha\ndescription: Alpha skill v2\n---\nNew body.\n"
	if _, err := repo.Update(ctx, "ws-1", &agentsv1.Skill{Name: "alpha", Description: "Alpha skill v2"}, md); err != nil {
		t.Fatalf("update skill: %v", err)
	}

	// Same Source instance, no rebuild: the edit is visible on the next call.
	if body, err := src.LoadInstructions(ctx, "alpha"); err != nil || body != "New body.\n" {
		t.Errorf("LoadInstructions after edit = %q, %v, want %q", body, err, "New body.\n")
	}
	fm, err := src.LoadFrontmatter(ctx, "alpha")
	if err != nil || fm.Description != "Alpha skill v2" {
		t.Errorf("LoadFrontmatter after edit = %+v, %v, want description %q", fm, err, "Alpha skill v2")
	}
}

func TestResourceMethodsWithoutResourcesSlice(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")
	seedSkill(t, repo, "ws-1", "hidden", "Not attached to the agent", "Secret.\n")

	src := NewSource(repo, "ws-1", []string{"alpha", "dangling"})

	// Skills have no resources in v1: an existing skill lists none.
	paths, err := src.ListResources(ctx, "alpha", "")
	if err != nil {
		t.Fatalf("ListResources(alpha): %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("ListResources(alpha) = %v, want empty", paths)
	}
	if _, err := src.ListResources(ctx, "hidden", ""); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("ListResources(hidden) err = %v, want ErrSkillNotFound", err)
	}
	if _, err := src.ListResources(ctx, "dangling", ""); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("ListResources(dangling) err = %v, want ErrSkillNotFound", err)
	}

	if _, err := src.LoadResource(ctx, "alpha", "references/notes.md"); !errors.Is(err, adkskill.ErrResourceNotFound) {
		t.Errorf("LoadResource(alpha) err = %v, want ErrResourceNotFound", err)
	}
	if _, err := src.LoadResource(ctx, "hidden", "references/notes.md"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadResource(hidden) err = %v, want ErrSkillNotFound", err)
	}
	if _, err := src.LoadResource(ctx, "dangling", "references/notes.md"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadResource(dangling) err = %v, want ErrSkillNotFound", err)
	}
}

func TestSourceIsSafeForConcurrentUse(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")

	src := NewSource(repo, "ws-1", []string{"alpha", "beta"})

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 50 {
				if _, err := src.ListFrontmatters(ctx); err != nil {
					t.Errorf("ListFrontmatters: %v", err)
					return
				}
				if _, err := src.LoadInstructions(ctx, "alpha"); err != nil {
					t.Errorf("LoadInstructions: %v", err)
					return
				}
				_, _ = src.LoadFrontmatter(ctx, "beta") // dangling, may race with the writer below
			}
		})
	}
	// Concurrent writer: skills are created and deleted while readers run.
	wg.Go(func() {
		for range 50 {
			seedSkill(t, repo, "ws-1", "beta", "Beta skill", "Use beta.\n")
			if err := repo.Delete(ctx, "ws-1", "beta"); err != nil {
				t.Errorf("delete beta: %v", err)
				return
			}
		}
	})
	wg.Wait()
}

func TestListFrontmattersFiltersToAllowlist(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")
	seedSkill(t, repo, "ws-1", "beta", "Beta skill", "Use beta.\n")
	seedSkill(t, repo, "ws-1", "hidden", "Not attached to the agent", "Secret.\n")

	src := NewSource(repo, "ws-1", []string{"alpha", "beta", "dangling"})

	fms, err := src.ListFrontmatters(ctx)
	if err != nil {
		t.Fatalf("ListFrontmatters: %v", err)
	}
	got := make(map[string]string, len(fms))
	for _, fm := range fms {
		got[fm.Name] = fm.Description
	}
	want := map[string]string{"alpha": "Alpha skill", "beta": "Beta skill"}
	if len(got) != len(want) {
		t.Fatalf("frontmatters = %v, want exactly %v (hidden filtered, dangling omitted)", got, want)
	}
	for name, desc := range want {
		if got[name] != desc {
			t.Errorf("frontmatter %q description = %q, want %q", name, got[name], desc)
		}
	}
}
