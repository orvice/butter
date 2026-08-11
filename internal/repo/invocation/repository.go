package invocation

import (
	"context"
	"errors"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ErrNotFound is returned by Get when an invocation does not exist.
var ErrNotFound = errors.New("invocation not found")

// ListFilter narrows results returned by List. AgentID filters by the
// immutable agent_id; AgentName is the legacy filter kept so historical
// records written before the Agent ID migration stay reachable.
type ListFilter struct {
	WorkspaceID string
	AgentID     string
	AgentName   string
	SessionID   string
}

// StatusSummary describes the runtime-relevant view of one agent's
// invocations: the most recent record and the number currently RUNNING.
type StatusSummary struct {
	Latest  *agentsv1.Invocation
	Running int32
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
	// FindByRequestID returns the invocation associated with the given
	// workspace+request_id pair, or ErrNotFound when none exists.
	FindByRequestID(ctx context.Context, workspaceID, requestID string) (*agentsv1.Invocation, error)
	// ListRecent returns the most recent invocations across all agents, used
	// to drive the dashboard activity feed.
	ListRecent(ctx context.Context, limit int32, pageToken string) ([]*agentsv1.Invocation, string, error)
	// StatusSummaries returns, for each named agent in the workspace, its most
	// recent invocation and the count of currently RUNNING invocations.
	// Agents with no invocations are absent from the returned map.
	StatusSummaries(ctx context.Context, workspaceID string, agentNames []string) (map[string]StatusSummary, error)
	// CountByTimeRange returns, across all workspaces, the number of
	// invocations whose started_at falls within the half-open window
	// [start, end), together with the subset that ended in
	// INVOCATION_STATUS_FAILED. Drives the dashboard Activity metric cards.
	CountByTimeRange(ctx context.Context, start, end time.Time) (total int64, failed int64, err error)
	// MarkStaleRunning transitions all QUEUED/RUNNING invocations in the
	// collection to FAILED with the given reason. Used at startup to
	// reconcile invocations orphaned by a previous process exit.
	MarkStaleRunning(ctx context.Context, reason string) (int64, error)
	// FindActiveBySession returns the QUEUED or RUNNING invocation for the
	// given session, or ErrNotFound when there is no active invocation.
	FindActiveBySession(ctx context.Context, workspaceID, sessionID string) (*agentsv1.Invocation, error)
}
