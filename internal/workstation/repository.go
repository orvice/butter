// Package workstation owns the shared projection, visibility and publication
// rules. Provider adapters have only read capabilities and never access storage.
package workstation

import (
	"context"
	"errors"
	"time"

	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrNotFound       = errors.New("work station resource not found")
	ErrConflict       = errors.New("configuration changed; reload and try again")
	ErrInvalid        = errors.New("invalid work station request")
	ErrUnsupported    = errors.New("provider profile is not supported")
	ErrPaused         = errors.New("connection is paused; an owner must test or replace its credential")
	ErrFenced         = errors.New("sync attempt lost its claim")
	ErrCleanupPending = errors.New("snapshot published; cleanup pending")
)

// Record is internal: neither it nor Credential can be serialized as an RPC.
// Config and all scope pointers share a document, so revocation and publication
// use the same CAS boundary. Item bodies are separate, immutable observations.
type Record struct {
	Source      SourceIdentity `bson:"source"`
	WorkspaceID string         `bson:"workspace_id"`
	ID          string         `bson:"id"`
	Name        string         `bson:"name"`
	HostID      string         `bson:"host_id"`
	PrincipalID string         `bson:"principal_id"`
	Revision    int64          `bson:"revision"`
	Version     int64          `bson:"version"`
	Epoch       string         `bson:"epoch"`
	Credential  Credential     `bson:"credential"`
	Paused      bool           `bson:"paused"`
	RetryAfter  time.Time      `bson:"retry_after"`
	ManualAfter time.Time      `bson:"manual_after"`
	Scopes      []Scope        `bson:"scopes"`
}
type Credential struct {
	Ciphertext string `bson:"ciphertext"`
	KeyID      string `bson:"key_id"`
}
type Scope struct {
	ID          string    `bson:"id"`
	Name        string    `bson:"name"`
	Generation  string    `bson:"generation"`
	Snapshot    string    `bson:"snapshot"`
	Health      string    `bson:"health"`
	Coverage    string    `bson:"coverage"`
	Unresolved  uint32    `bson:"unresolved"`
	LastSuccess time.Time `bson:"last_success"`
	Due         time.Time `bson:"due"`
	Claim       string    `bson:"claim"`
	LeaseUntil  time.Time `bson:"lease_until"`
	Hidden      bool      `bson:"hidden"`
	Scanned     uint32    `bson:"scanned_item_count"`
	SyncPhase   string    `bson:"sync_phase"`
}
type Fence struct {
	Token string
	Until time.Time
}

// Observation is source-qualified. Never copy body fields from a revoked source.
type Observation struct {
	WorkspaceID  string
	ConnectionID string
	ScopeID      string
	Snapshot     string
	Item         *v1.WorkStationItem
}
type Query struct {
	AllowedConnections                                         []string
	ID, Kind, ConnectionID, ScopeID, Status, Assignee, Keyword string
	Limit                                                      int
	AfterID                                                    string
	AfterTime                                                  time.Time
	Now                                                        time.Time
}

// Repository's CAS checks version and, for publication/error writes, the lease
// against database time at the write itself. Now supplies database time for
// scheduling; it cannot substitute for the CAS lease check.
type Repository interface {
	EnsureIndexes(context.Context) error
	Now(context.Context) (time.Time, error)
	List(context.Context, string) ([]Record, error)
	Due(context.Context) ([]Record, error)
	Get(context.Context, string, string) (Record, error)
	Create(context.Context, Record) error
	CAS(context.Context, Record, *Record, *Fence) (bool, error)
	Stage(context.Context, []Observation) error
	// PruneSnapshot removes one permanently superseded token, never "all except
	// current": another publisher may already be staging or publishing a successor.
	PruneSnapshot(context.Context, string, string, string, string) error
	Query(context.Context, string, Query) ([]*v1.WorkStationItem, error)
	HostInUse(context.Context, string) (bool, error)
	// Collect drains expired orphan batches until empty or context cancellation.
	// References (including hidden/paused ones) and recent staging are retained.
	Collect(context.Context) error
}

type Identity struct {
	PrincipalID string
	Source      SourceIdentity
}
type Page struct {
	Scopes []Scope
	Next   string
}
type Scan struct {
	Items      []*v1.WorkStationItem
	Coverage   string
	Unresolved uint32
}
type Provider interface {
	Probe(context.Context, string) (Identity, error)
	DiscoverScopes(context.Context, string, string) (Page, error)
	ResolveScope(context.Context, string, string) (Scope, error)
	Scan(context.Context, string, Scope, time.Time) (Scan, error)
}

// Failure categories contain no upstream body, URL, header or credential.
type Failure struct {
	Category   string
	RetryAfter time.Duration
}

func (f *Failure) Error() string { return f.Category }
func Classify(err error) *Failure {
	var f *Failure
	if errors.As(err, &f) {
		return f
	}
	if errors.Is(err, ErrPaused) {
		return &Failure{Category: "CONFIG_ERROR"}
	}
	return &Failure{Category: "DEGRADED"}
}
