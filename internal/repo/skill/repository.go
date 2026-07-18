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
	Delete(ctx context.Context, workspaceID, name string) error
}
