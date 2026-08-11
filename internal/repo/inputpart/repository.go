package inputpart

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ErrNotFound is returned when the requested input parts do not exist.
var ErrNotFound = errors.New("input parts not found")

// Record is one persisted piece of multimodal input, stored separately
// from the Invocation document so that large inline-data payloads do not
// exceed MongoDB's per-document limit.
type Record struct {
	InvocationID string
	Index        int
	Part         *agentsv1.InputPart
}

// Repository persists ordered multimodal Input Parts for async invocations.
// Parts are stored as individual records so a valid request (up to 20 MiB
// combined) never exceeds MongoDB's 16 MiB document limit. Implementations
// must be safe for concurrent use.
type Repository interface {
	// SaveAll persists all parts for an invocation atomically. If records for
	// the same invocation_id already exist the call is idempotent (no
	// duplicates created). Callers must validate parts before saving.
	SaveAll(ctx context.Context, invocationID string, parts []*agentsv1.InputPart) error

	// Load returns the ordered parts for one invocation, or ErrNotFound if
	// no records exist.
	Load(ctx context.Context, invocationID string) ([]*agentsv1.InputPart, error)

	// Delete removes all stored parts for an invocation. It is a no-op when
	// no records exist.
	Delete(ctx context.Context, invocationID string) error
}
