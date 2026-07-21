package skill

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// Repository stores workspace-scoped Skills: parsed frontmatter metadata plus
// the SKILL.md document, which backends persist separately (ADR 0004).
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	List(ctx context.Context, workspaceID string) ([]*agentsv1.Skill, error)
	Get(ctx context.Context, workspaceID, name string) (*agentsv1.Skill, error)
	GetSkillMD(ctx context.Context, workspaceID, name string) (string, error)
	Create(ctx context.Context, workspaceID string, skill *agentsv1.Skill, skillMD string) (*agentsv1.Skill, error)
	Update(ctx context.Context, workspaceID string, skill *agentsv1.Skill, skillMD string) (*agentsv1.Skill, error)
	// Delete removes the skill and cascades to all of its resources.
	Delete(ctx context.Context, workspaceID, name string) error

	// Resource files under a skill's references/, assets/, scripts/
	// directories (issue #154). Paths are pre-validated by callers via
	// CleanResourcePath; metadata is served without touching content
	// storage, content is binary-safe. All methods return ErrNotFound when
	// the skill (or, where addressed, the resource) does not exist.
	ListResources(ctx context.Context, workspaceID, skillName string) ([]*agentsv1.SkillResource, error)
	GetResource(ctx context.Context, workspaceID, skillName, path string) (*agentsv1.SkillResource, []byte, error)
	// PutResource creates or overwrites in place, stamping size and timestamps.
	PutResource(ctx context.Context, workspaceID, skillName string, resource *agentsv1.SkillResource, content []byte) (*agentsv1.SkillResource, error)
	DeleteResource(ctx context.Context, workspaceID, skillName, path string) error
}
