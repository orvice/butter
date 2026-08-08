// Package githost stores the platform-level allowlist of Git hosts that
// workspaces may bind repositories from (issue #214). Hosts are configured
// by global admins only, so workspace input can never introduce an arbitrary
// API base URL.
package githost

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// Repository stores GitHost entries. Hosts are platform scoped: no
// workspaceID parameter. IDs are caller-assigned UUIDs.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	List(ctx context.Context) ([]*agentsv1.GitHost, error)
	Get(ctx context.Context, id string) (*agentsv1.GitHost, error)
	Create(ctx context.Context, host *agentsv1.GitHost) (*agentsv1.GitHost, error)
	Update(ctx context.Context, host *agentsv1.GitHost) (*agentsv1.GitHost, error)
	Delete(ctx context.Context, id string) error
}
