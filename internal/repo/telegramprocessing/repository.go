// Package telegramprocessing persists the auditable state of every accepted
// Telegram update (issue #264).
//
// The record exists to answer one question honestly: is it safe to run this
// again? Everything up to the moment Agent work starts is safely retryable;
// once the Agent may have produced side effects it is not. Recording the
// transition is what lets a crashed worker's work be reclaimed without
// silently repeating a tool call.
package telegramprocessing

import (
	"context"
	"errors"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrNotFound means no record with that ID (or address) exists.
	ErrNotFound = errors.New("not found")
)

// RetentionPeriod is how long records, output, and delivery segments survive.
// Long enough to investigate an incident, short enough that Telegram content
// is not retained indefinitely.
const RetentionPeriod = 30 * 24 * time.Hour

// Filter narrows a listing.
type Filter struct {
	WorkspaceID   string
	ChannelID     string
	DestinationID string
	Status        agentsv1.TelegramProcessingStatus
	Limit         int
}

// Repository persists processing records.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Claim creates or re-claims the record for one accepted update and
	// increments its attempt count. `claimed` is false when the update has
	// already reached a terminal success — the caller then acknowledges
	// without invoking anything, which is how a duplicate delivery avoids a
	// second Agent run.
	Claim(ctx context.Context, record *agentsv1.TelegramProcessingRecord) (stored *agentsv1.TelegramProcessingRecord, claimed bool, err error)

	// Update replaces the mutable state of an existing record.
	Update(ctx context.Context, record *agentsv1.TelegramProcessingRecord) (*agentsv1.TelegramProcessingRecord, error)

	Get(ctx context.Context, workspaceID, id string) (*agentsv1.TelegramProcessingRecord, error)
	List(ctx context.Context, filter Filter) ([]*agentsv1.TelegramProcessingRecord, error)
}
