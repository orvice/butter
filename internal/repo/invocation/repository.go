package invocation

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ErrNotFound is returned by Get when an invocation does not exist.
var ErrNotFound = errors.New("invocation not found")

// ListFilter narrows results returned by List.
type ListFilter struct {
	WorkspaceID string
	AgentName   string
	SessionID   string
}

// Repository persists invocation records produced by runner.Service.
//
// Implementations must accept Upsert semantics in Save: the runner first
// records the invocation as RUNNING, then updates it with the terminal status
// after the call completes.
type Repository interface {
	Save(ctx context.Context, inv *agentsv1.Invocation) error
	List(ctx context.Context, filter ListFilter, pageSize int32, pageToken string) ([]*agentsv1.Invocation, string, int32, error)
	Get(ctx context.Context, id string) (*agentsv1.Invocation, error)
	// ListRecent returns the most recent invocations across all agents, used
	// to drive the dashboard activity feed.
	ListRecent(ctx context.Context, limit int32, pageToken string) ([]*agentsv1.Invocation, string, error)
}
