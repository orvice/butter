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

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ErrNotFound means no operation exists with the requested ID in the workspace.
var ErrNotFound = errors.New("agent operation not found")

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

	// Save upserts the operation by its ID within the given workspace.
	// The workspaceID must match op.WorkspaceId; implementations enforce
	// this so one workspace cannot overwrite another's records.
	Save(ctx context.Context, workspaceID string, op *agentsv1.AgentOperation) error

	// Get returns the operation with the given ID in the workspace, or
	// ErrNotFound.
	Get(ctx context.Context, workspaceID, id string) (*agentsv1.AgentOperation, error)

	// List returns operations in the workspace, newest first, optionally
	// filtered by status (UNSPECIFIED returns all), with opaque-token
	// pagination.
	List(ctx context.Context, workspaceID string, status agentsv1.AgentOperationStatus, pageSize int32, pageToken string) ([]*agentsv1.AgentOperation, string, error)

	// ListResumableAcrossWorkspaces returns every operation still in RUNNING or
	// FAILED across all workspaces, so a background driver can heal stuck
	// operations. Unused by the synchronous path; provided for a future
	// reconciler.
	ListResumableAcrossWorkspaces(ctx context.Context) ([]*agentsv1.AgentOperation, error)
}
