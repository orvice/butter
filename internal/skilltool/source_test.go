package skilltool

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"sync"
	"testing"

	adkskill "google.golang.org/adk/v2/tool/skilltoolset/skill"

	skillmemory "go.orx.me/apps/butter/internal/repo/skill/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// listedNames returns the sorted skill names ListFrontmatters exposes.
func listedNames(t *testing.T, ctx context.Context, src *Source) []string {
	t.Helper()
	fms, err := src.ListFrontmatters(ctx)
	if err != nil {
		t.Fatalf("ListFrontmatters: %v", err)
	}
	names := make([]string, 0, len(fms))
	for _, fm := range fms {
		names = append(names, fm.Name)
	}
	slices.Sort(names)
	return names
}

// writeSkill stores a skill with a minimal spec-valid SKILL.md document and
// returns any error, so it is safe to call from a non-test goroutine (where
// t.Fatalf is illegal).
func writeSkill(repo *skillmemory.Store, workspaceID, name, description, body string) error {
	md := "---\nname: " + name + "\ndescription: " + description + "\n---\n" + body
	_, err := repo.Create(context.Background(), workspaceID, &agentsv1.Skill{
		Name:        name,
		Description: description,
	}, md)
	return err
}

// seedSkill is the test-goroutine wrapper around writeSkill.
func seedSkill(t *testing.T, repo *skillmemory.Store, workspaceID, name, description, body string) {
	t.Helper()
	if err := writeSkill(repo, workspaceID, name, description, body); err != nil {
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

// seedResource stores a resource file on an existing skill.
func seedResource(t *testing.T, repo *skillmemory.Store, workspaceID, skillName, path string, content []byte) {
	t.Helper()
	_, err := repo.PutResource(context.Background(), workspaceID, skillName, &agentsv1.SkillResource{Path: path}, content)
	if err != nil {
		t.Fatalf("seed resource %s/%s: %v", skillName, path, err)
	}
}

func TestLoadResourceReturnsContent(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")
	binary := []byte{0x89, 'P', 'N', 'G', 0x00, 0x1f, 0xff}
	seedResource(t, repo, "ws-1", "alpha", "assets/logo.png", binary)

	src := NewSource(repo, "ws-1", []string{"alpha"})

	rc, err := src.LoadResource(ctx, "alpha", "assets/logo.png")
	if err != nil {
		t.Fatalf("LoadResource: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if !bytes.Equal(got, binary) {
		t.Errorf("resource content = %v, want %v", got, binary)
	}
}

func TestLoadResourceSentinelsAndPathSafety(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")
	seedSkill(t, repo, "ws-1", "hidden", "Not attached to the agent", "Secret.\n")
	seedResource(t, repo, "ws-1", "alpha", "references/notes.md", []byte("notes"))

	src := NewSource(repo, "ws-1", []string{"alpha", "dangling"})

	// Missing resource on a visible skill: ADK's resource sentinel.
	if _, err := src.LoadResource(ctx, "alpha", "references/absent.md"); !errors.Is(err, adkskill.ErrResourceNotFound) {
		t.Errorf("LoadResource(absent) err = %v, want ErrResourceNotFound", err)
	}
	// Invisible skills: skill sentinel, same as before the resources slice.
	if _, err := src.LoadResource(ctx, "hidden", "references/notes.md"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadResource(hidden) err = %v, want ErrSkillNotFound", err)
	}
	if _, err := src.LoadResource(ctx, "dangling", "references/notes.md"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadResource(dangling) err = %v, want ErrSkillNotFound", err)
	}
	// Traversal and out-of-spec paths: rejected before touching storage.
	for _, p := range []string{"../secrets.txt", "references/../../etc/passwd", "docs/readme.md", "references/dir/../notes.md"} {
		if _, err := src.LoadResource(ctx, "alpha", p); !errors.Is(err, adkskill.ErrInvalidResourcePath) {
			t.Errorf("LoadResource(%q) err = %v, want ErrInvalidResourcePath", p, err)
		}
	}
}

func TestListResourcesFiltersBySubpath(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")
	seedResource(t, repo, "ws-1", "alpha", "references/api.md", []byte("api"))
	seedResource(t, repo, "ws-1", "alpha", "references/deep/notes.md", []byte("notes"))
	seedResource(t, repo, "ws-1", "alpha", "scripts/run.sh", []byte("#!/bin/sh"))

	src := NewSource(repo, "ws-1", []string{"alpha"})

	// Root listing ("" and ".") returns every resource, skill-root relative.
	for _, root := range []string{"", "."} {
		paths, err := src.ListResources(ctx, "alpha", root)
		if err != nil {
			t.Fatalf("ListResources(%q): %v", root, err)
		}
		want := []string{"references/api.md", "references/deep/notes.md", "scripts/run.sh"}
		if !slices.Equal(paths, want) {
			t.Errorf("ListResources(%q) = %v, want %v", root, paths, want)
		}
	}

	// Subpath narrows to one directory.
	paths, err := src.ListResources(ctx, "alpha", "references")
	if err != nil {
		t.Fatalf("ListResources(references): %v", err)
	}
	if want := []string{"references/api.md", "references/deep/notes.md"}; !slices.Equal(paths, want) {
		t.Errorf("ListResources(references) = %v, want %v", paths, want)
	}

	// Subpath with no matches mirrors FileSystemSource's missing-directory error.
	if _, err := src.ListResources(ctx, "alpha", "assets"); !errors.Is(err, adkskill.ErrResourceNotFound) {
		t.Errorf("ListResources(assets) err = %v, want ErrResourceNotFound", err)
	}
	// Out-of-spec or traversing subpaths are invalid.
	for _, p := range []string{"docs", "../alpha/references", "references/.."} {
		if _, err := src.ListResources(ctx, "alpha", p); !errors.Is(err, adkskill.ErrInvalidResourcePath) {
			t.Errorf("ListResources(%q) err = %v, want ErrInvalidResourcePath", p, err)
		}
	}
}

func TestListResourcesSentinels(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")
	seedSkill(t, repo, "ws-1", "hidden", "Not attached to the agent", "Secret.\n")

	src := NewSource(repo, "ws-1", []string{"alpha", "dangling"})

	// A visible skill with no resources lists none at the root.
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
}

func TestListFrontmattersReflectsDeletion(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkill(t, repo, "ws-1", "alpha", "Alpha skill", "Use alpha.\n")
	seedSkill(t, repo, "ws-1", "beta", "Beta skill", "Use beta.\n")

	src := NewSource(repo, "ws-1", []string{"alpha", "beta"})

	if names := listedNames(t, ctx, src); !slices.Equal(names, []string{"alpha", "beta"}) {
		t.Fatalf("initial listing = %v, want [alpha beta]", names)
	}

	// Delete one referenced skill: the agent keeps running, the remaining
	// skill is still listed, and the deleted one is silently omitted.
	if err := repo.Delete(ctx, "ws-1", "beta"); err != nil {
		t.Fatalf("delete beta: %v", err)
	}
	if names := listedNames(t, ctx, src); !slices.Equal(names, []string{"alpha"}) {
		t.Errorf("listing after delete = %v, want [alpha]", names)
	}
	if _, err := src.LoadInstructions(ctx, "beta"); !errors.Is(err, adkskill.ErrSkillNotFound) {
		t.Errorf("LoadInstructions(beta) after delete err = %v, want ErrSkillNotFound", err)
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
			if err := writeSkill(repo, "ws-1", "beta", "Beta skill", "Use beta.\n"); err != nil {
				t.Errorf("write beta: %v", err)
				return
			}
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
