// Package skilltool adapts the workspace skill repository to ADK's
// skill.Source interface so agents can list and load skills at runtime
// (ADR 0004). A Source is bound at construction to one workspace and one
// agent's skill allowlist; every call queries the repository live, so skill
// edits take effect on the next LLM turn without rebuilding the agent.
package skilltool

import (
	"context"
	"errors"
	"fmt"
	"io"

	adkskill "google.golang.org/adk/v2/tool/skilltoolset/skill"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Source implements adkskill.Source over a skill repository. It holds no
// mutable state, so it is safe for concurrent use as long as the repository
// is.
type Source struct {
	repo        skillrepo.Repository
	workspaceID string
	allow       map[string]struct{}
}

// NewSource binds a repository to a workspace and an agent's skill-name
// allowlist. Names outside the allowlist are invisible; allowlisted names
// missing from the repository are silently omitted (mirroring mcp_server_ids
// leniency).
func NewSource(repo skillrepo.Repository, workspaceID string, skillNames []string) *Source {
	allow := make(map[string]struct{}, len(skillNames))
	for _, name := range skillNames {
		allow[name] = struct{}{}
	}
	return &Source{repo: repo, workspaceID: workspaceID, allow: allow}
}

func (s *Source) ListFrontmatters(ctx context.Context) ([]*adkskill.Frontmatter, error) {
	skills, err := s.repo.List(ctx, s.workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	out := make([]*adkskill.Frontmatter, 0, len(skills))
	for _, sk := range skills {
		if _, ok := s.allow[sk.GetName()]; !ok {
			continue
		}
		out = append(out, frontmatterFromProto(sk))
	}
	return out, nil
}

// notFound wraps ADK's sentinel with the skill name for log context.
func notFound(name string) error {
	return fmt.Errorf("skill %q: %w", name, adkskill.ErrSkillNotFound)
}

func (s *Source) LoadFrontmatter(ctx context.Context, name string) (*adkskill.Frontmatter, error) {
	sk, err := s.getAllowed(ctx, name)
	if err != nil {
		return nil, err
	}
	return frontmatterFromProto(sk), nil
}

// getAllowed returns the skill if it is allowlisted and present, otherwise the
// ErrSkillNotFound sentinel. Allowlist misses and repository misses are
// deliberately indistinguishable to callers: a skill outside the allowlist is
// as invisible as one that never existed.
func (s *Source) getAllowed(ctx context.Context, name string) (*agentsv1.Skill, error) {
	if _, ok := s.allow[name]; !ok {
		return nil, notFound(name)
	}
	sk, err := s.repo.Get(ctx, s.workspaceID, name)
	if err != nil {
		if errors.Is(err, skillrepo.ErrNotFound) {
			return nil, notFound(name)
		}
		return nil, fmt.Errorf("get skill %q: %w", name, err)
	}
	return sk, nil
}

func (s *Source) LoadInstructions(ctx context.Context, name string) (string, error) {
	if _, ok := s.allow[name]; !ok {
		return "", notFound(name)
	}
	md, err := s.repo.GetSkillMD(ctx, s.workspaceID, name)
	if err != nil {
		if errors.Is(err, skillrepo.ErrNotFound) {
			return "", notFound(name)
		}
		return "", fmt.Errorf("get skill %q document: %w", name, err)
	}
	// The repository stores the full SKILL.md (validated at write time);
	// callers of LoadInstructions expect only the markdown body.
	_, body, err := adkskill.ParseBytes([]byte(md))
	if err != nil {
		return "", fmt.Errorf("skill %q: %w: %v", name, adkskill.ErrInvalidFrontmatter, err)
	}
	return body, nil
}

// The skill repository does not yet expose resource files (SKILL.md bodies
// only), so these methods resolve the skill for correct sentinel semantics,
// then report no resources — the scope cut called out in issue #152 ("resource
// methods may return not-found until the resources slice lands").

func (s *Source) ListResources(ctx context.Context, name, subpath string) ([]string, error) {
	if err := s.ensureVisible(ctx, name); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Source) LoadResource(ctx context.Context, name, resourcePath string) (io.ReadCloser, error) {
	if err := s.ensureVisible(ctx, name); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("skill %q resource %q: %w", name, resourcePath, adkskill.ErrResourceNotFound)
}

// ensureVisible checks that a skill is allowlisted and present in the
// repository, returning ErrSkillNotFound otherwise.
func (s *Source) ensureVisible(ctx context.Context, name string) error {
	_, err := s.getAllowed(ctx, name)
	return err
}

func frontmatterFromProto(sk *agentsv1.Skill) *adkskill.Frontmatter {
	return &adkskill.Frontmatter{
		Name:          sk.GetName(),
		Description:   sk.GetDescription(),
		License:       sk.GetLicense(),
		Compatibility: sk.GetCompatibility(),
		Metadata:      sk.GetMetadata(),
		AllowedTools:  sk.GetAllowedTools(),
	}
}

var _ adkskill.Source = (*Source)(nil)
