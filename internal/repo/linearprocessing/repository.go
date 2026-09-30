// Package linearprocessing persists the auditable state of every accepted
// Linear delivery (ADR-0015 §8).
//
// The record answers one question honestly: may the Agent have run?
// Everything before Agent work starts is safely retryable; once the Agent
// may have produced side effects it is not. Recording that transition is
// what lets a crashed worker's work be reclaimed without silently repeating
// a tool call (ADR-0009).
package linearprocessing

import (
	"context"
	"errors"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrNotFound means no record with that ID exists.
	ErrNotFound = errors.New("not found")
	// ErrInProgress means another worker or operator holds the live
	// processing lease.
	ErrInProgress = errors.New("linear processing record is already claimed")
	// ErrLeaseLost means a conditional update came from a former owner.
	ErrLeaseLost = errors.New("linear processing lease lost")
)

// RetentionPeriod is how long records and their persisted replies survive.
const RetentionPeriod = 30 * 24 * time.Hour

// Filter narrows a listing.
type Filter struct {
	WorkspaceID    string
	AppID          string
	AgentSessionID string
	Status         agentsv1.LinearProcessingStatus
	Limit          int
}

// ClaimAction tells a worker what can safely happen after claiming a
// delivery. It is derived from persisted state, so recovery never guesses
// whether Agent work already ran.
type ClaimAction int

const (
	// ClaimRunAgent starts or retries work that has not crossed the Agent
	// side-effect boundary.
	ClaimRunAgent ClaimAction = iota
	// ClaimResumeDelivery posts a reply that was already persisted.
	ClaimResumeDelivery
	// ClaimReportInterrupted tells the session that a turn which may have
	// run was cut short. The claim already marked the record
	// FAILED_UNCERTAIN.
	ClaimReportInterrupted
	// ClaimAcknowledge completes the queue delivery without doing more.
	ClaimAcknowledge
)

// RecoveryAction derives the only safe action from a persisted record.
func RecoveryAction(record *agentsv1.LinearProcessingRecord) ClaimAction {
	switch record.GetStatus() {
	case agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_RECEIVED:
		return ClaimRunAgent
	case agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING:
		return ClaimReportInterrupted
	case agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_READY_TO_DELIVER:
		return ClaimResumeDelivery
	case agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED:
		if record.GetOutput() != "" {
			return ClaimResumeDelivery
		}
		return ClaimRunAgent
	default:
		return ClaimAcknowledge
	}
}

// MarkInterruptedUncertain records a reclaimed side-effect boundary. It
// returns true when the record changed.
func MarkInterruptedUncertain(record *agentsv1.LinearProcessingRecord) bool {
	if record.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING {
		return false
	}
	record.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN
	record.DeadLettered = true
	record.Error = "the worker stopped while the Agent was running"
	return true
}

// Repository persists processing records.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Claim creates or re-claims the record of one accepted delivery, keyed
	// by (app ID, delivery ID), and derives the only safe action for its
	// persisted state. It returns ErrInProgress while another owner's lease
	// is live.
	Claim(ctx context.Context, record *agentsv1.LinearProcessingRecord, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.LinearProcessingRecord, ClaimAction, error)

	// ClaimForResend exclusively leases an existing record for an operator
	// resend. It shares the lease with automatic recovery.
	ClaimForResend(ctx context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.LinearProcessingRecord, error)

	// UpdateClaimed replaces mutable state only while leaseToken owns the
	// record, so a stale owner cannot overwrite a new recovery run.
	UpdateClaimed(ctx context.Context, record *agentsv1.LinearProcessingRecord, leaseToken string) (*agentsv1.LinearProcessingRecord, error)
	// Update replaces mutable state without a lease. Used for records the
	// session lease already serializes, such as queued follow-ups.
	Update(ctx context.Context, record *agentsv1.LinearProcessingRecord) (*agentsv1.LinearProcessingRecord, error)

	RenewClaim(ctx context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error
	ReleaseClaim(ctx context.Context, workspaceID, id, leaseToken string) error

	Get(ctx context.Context, workspaceID, id string) (*agentsv1.LinearProcessingRecord, error)
	// List returns records newest first.
	List(ctx context.Context, filter Filter) ([]*agentsv1.LinearProcessingRecord, error)
}
