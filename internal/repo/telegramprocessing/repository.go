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
	// ErrInProgress means another worker or operator owns the live processing
	// lease. The caller must not perform work or acknowledge queued work.
	ErrInProgress = errors.New("telegram processing record is already claimed")
	// ErrLeaseLost means a conditional update came from a former owner.
	ErrLeaseLost = errors.New("telegram processing lease lost")
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

// ClaimAction tells a worker what can safely happen after claiming an
// accepted update. The action is derived from persisted state so recovery
// never has to guess whether Agent work already ran.
type ClaimAction int

const (
	// ClaimRunAgent starts or retries work that has not crossed the Agent
	// side-effect boundary.
	ClaimRunAgent ClaimAction = iota
	// ClaimResumeDelivery sends output that was already persisted without
	// invoking the Agent again.
	ClaimResumeDelivery
	// ClaimAcknowledge completes the queue delivery without doing more work.
	ClaimAcknowledge
)

// RecoveryAction derives the only safe action from a persisted record.
func RecoveryAction(record *agentsv1.TelegramProcessingRecord) ClaimAction {
	if HasSendingSegment(record) {
		return ClaimAcknowledge
	}
	switch record.GetStatus() {
	case agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER:
		return ClaimResumeDelivery
	case agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED:
		if record.GetOutput() != "" && len(record.GetSegments()) > 0 {
			return ClaimResumeDelivery
		}
		return ClaimRunAgent
	case agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_RECEIVED:
		return ClaimRunAgent
	default:
		return ClaimAcknowledge
	}
}

// MarkInterruptedUncertain records a reclaimed side-effect boundary. It
// returns true when the record changed.
func MarkInterruptedUncertain(record *agentsv1.TelegramProcessingRecord) bool {
	switch {
	case HasSendingSegment(record):
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN
		record.DeadLettered = true
		record.Error = "worker stopped while a Telegram segment delivery was in progress"
		return true
	case record.GetStatus() == agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_PROCESSING:
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN
		record.DeadLettered = true
		record.Error = "worker stopped while Agent work was in progress"
		return true
	default:
		return false
	}
}

// HasSendingSegment reports whether Telegram may have accepted a segment
// whose result was not persisted.
func HasSendingSegment(record *agentsv1.TelegramProcessingRecord) bool {
	for _, segment := range record.GetSegments() {
		if segment.GetStatus() == "sending" {
			return true
		}
	}
	return false
}

// Repository persists processing records.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Claim creates or re-claims the record for one accepted update and
	// derives the only recovery action that is safe for its persisted state.
	Claim(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, claimedAt, leaseExpiresAt time.Time) (stored *agentsv1.TelegramProcessingRecord, action ClaimAction, err error)

	// ClaimDelivery exclusively leases an existing record for an operator
	// resend. It shares the same lease as automatic queue recovery.
	ClaimDelivery(ctx context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.TelegramProcessingRecord, error)

	// UpdateClaimed replaces mutable state only while leaseToken still owns the
	// record. This prevents a stale sender from overwriting a new recovery run.
	UpdateClaimed(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string) (*agentsv1.TelegramProcessingRecord, error)

	// RenewClaim extends a live claim while work is still making progress.
	RenewClaim(ctx context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error

	// ReleaseClaim gives up a live lease after the work path returns.
	ReleaseClaim(ctx context.Context, workspaceID, id, leaseToken string) error

	// Update replaces the mutable state of an existing record.
	Update(ctx context.Context, record *agentsv1.TelegramProcessingRecord) (*agentsv1.TelegramProcessingRecord, error)

	Get(ctx context.Context, workspaceID, id string) (*agentsv1.TelegramProcessingRecord, error)
	List(ctx context.Context, filter Filter) ([]*agentsv1.TelegramProcessingRecord, error)
}
