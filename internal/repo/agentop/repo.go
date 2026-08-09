// Package agentop persists durable Agent lifecycle operation records (issue
// #218). Each record tracks a cross-system Saga (create, composite save,
// tombstone, restore, or content purge) with a stable operation ID, an explicit
// status, and per-step progress so partial failures are visible and operations
// can be retried idempotently.
package agentop

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrNotFound means no operation exists with the requested ID in the workspace.
	ErrNotFound = errors.New("agent operation not found")
	// ErrAlreadyExists means the operation ID is already bound in the workspace.
	ErrAlreadyExists = errors.New("agent operation already exists")
	// ErrInProgress means another executor currently owns the operation lease.
	ErrInProgress = errors.New("agent operation is already running")
	// ErrCompleted means the operation already completed successfully.
	ErrCompleted = errors.New("agent operation already completed")
	// ErrLeaseLost means a stale executor attempted to persist after its lease was replaced.
	ErrLeaseLost = errors.New("agent operation lease lost")
)

// DecodePageToken decodes an opaque offset token, returning 0 for an empty or
// malformed token. Shared by the memory and mongo implementations so both use
// the same wire format.
func DecodePageToken(token string) int {
	if token == "" {
		return 0
	}
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(string(raw))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// EncodePageToken encodes an offset into an opaque page token.
func EncodePageToken(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

// ClampPageSize applies the default (20) and maximum (200) page sizes.
func ClampPageSize(pageSize int32) int32 {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 200 {
		return 200
	}
	return pageSize
}

// Repository stores workspace-scoped Agent lifecycle operation records.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Create inserts a new operation. It returns ErrAlreadyExists when another
	// request already created the same workspace-scoped operation ID.
	Create(ctx context.Context, workspaceID string, op *agentsv1.AgentOperation) error

	// Claim atomically moves PENDING/FAILED (or expired RUNNING) to RUNNING,
	// increments attempt_count, and assigns a renewable lease to the caller.
	Claim(ctx context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.AgentOperation, error)

	// RenewLease extends an active executor's lease. A replaced lease returns
	// ErrLeaseLost so the old executor can stop before performing more work.
	RenewLease(ctx context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error

	// SaveClaimed persists progress only while leaseToken still owns the
	// operation. A replaced lease returns ErrLeaseLost instead of overwriting
	// the current executor's progress.
	SaveClaimed(ctx context.Context, workspaceID, leaseToken string, op *agentsv1.AgentOperation) error

	// Get returns the operation with the given ID in the workspace, or
	// ErrNotFound.
	Get(ctx context.Context, workspaceID, id string) (*agentsv1.AgentOperation, error)

	// List returns operations in the workspace, newest first, optionally
	// filtered by status (UNSPECIFIED returns all), with opaque-token
	// pagination.
	List(ctx context.Context, workspaceID string, status agentsv1.AgentOperationStatus, pageSize int32, pageToken string) ([]*agentsv1.AgentOperation, string, error)
}
