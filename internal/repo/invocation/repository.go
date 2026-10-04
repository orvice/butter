package invocation

import (
	"context"
	"errors"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ErrNotFound is returned by Get when an invocation does not exist.
var ErrNotFound = errors.New("invocation not found")

// SourceAGUIDetached is the source of the Invocation records the AG-UI path
// owns: those of Detached Runs (ADR-0016 decision 3). The runner records
// nothing for these runs, so other entry points tell their records apart by
// it.
const SourceAGUIDetached = "agui-detached"

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

// StaleSelection says which QUEUED and RUNNING records MarkStaleRunning
// fails: the ones whose process is gone.
type StaleSelection struct {
	// LostOwners are instance IDs whose liveness has lapsed. Every QUEUED or
	// RUNNING record stamped with one of them is failed.
	LostOwners []string
	// LegacyBefore covers records without an owner stamp, written before
	// records carried one: those that started before it are failed. A record
	// with no start time counts as older than any cutoff. The zero time
	// fails none of them.
	LegacyBefore time.Time
	// Reason is the error recorded on each failed record, given its owner
	// stamp (empty for a record without one).
	Reason func(owner string) string
}

// Repository persists invocation records produced by runner.Service.
//
// Implementations must accept Upsert semantics in Save: the runner first
// records the invocation as RUNNING, then updates it with the terminal status
// after the call completes.
//
// Every record has an owner stamp: the instance ID of the process that
// created it, which is the process that runs it (#390). A store stamps the
// owner it was opened with (the stores' WithOwner) when Save creates a
// record, and keeps the stamp through every later save, whichever process
// makes it. The stamp is not part of the Invocation message, so it never
// reaches an API response.
type Repository interface {
	Save(ctx context.Context, inv *agentsv1.Invocation) error
	List(ctx context.Context, filter ListFilter, pageSize int32, pageToken string) ([]*agentsv1.Invocation, string, int32, error)
	Get(ctx context.Context, workspaceID, id string) (*agentsv1.Invocation, error)
	// GetAcrossWorkspaces is the explicit global-admin/runtime lookup path.
	GetAcrossWorkspaces(ctx context.Context, id string) (*agentsv1.Invocation, error)
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
	// ActiveOwners returns the owner stamps carried by QUEUED and RUNNING
	// records, each once. Records without a stamp are left out.
	ActiveOwners(ctx context.Context) ([]string, error)
	// MarkStaleRunning fails the QUEUED and RUNNING records whose process is
	// gone, as sel selects them: records stamped with a lost owner, and
	// records without a stamp that started before the legacy cutoff. A record
	// stamped with any other owner is never touched, so a live process's
	// runs survive another process starting. A record saved again between
	// being selected and being failed is left for the next sweep. Returns how
	// many records it failed. StaleSweeper calls it at startup and
	// periodically.
	MarkStaleRunning(ctx context.Context, sel StaleSelection) (int64, error)
	// FindActiveBySession returns the QUEUED or RUNNING invocation for the
	// given session, or ErrNotFound when there is no active invocation.
	FindActiveBySession(ctx context.Context, workspaceID, sessionID string) (*agentsv1.Invocation, error)
	// FindLatestBySession returns the most recent invocation for the given
	// session regardless of status, or ErrNotFound when the session has
	// none. Drives the inline failed/stopped turn rendering after reload.
	FindLatestBySession(ctx context.Context, workspaceID, sessionID string) (*agentsv1.Invocation, error)
	// ListBySession returns all invocations for a session, ordered by
	// started_at descending. Returns an empty slice (not an error) when
	// the session has no invocations.
	ListBySession(ctx context.Context, workspaceID, sessionID string) ([]*agentsv1.Invocation, error)
	// RedactContent clears the content-bearing fields (input, output,
	// error) on a single invocation while preserving operational metadata
	// (agent identity, status, timestamps, latency). Used during session
	// deletion to remove user content from retained audit records.
	RedactContent(ctx context.Context, workspaceID, id string) error
}
