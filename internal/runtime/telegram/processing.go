package telegram

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"

	"go.orx.me/apps/butter/internal/repo/telegramprocessing"
	"go.orx.me/apps/butter/internal/telegramqueue"
	"go.orx.me/apps/butter/internal/telegramsend"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// SessionGuard serializes work inside one derived session.
type SessionGuard interface {
	// Acquire takes the session lease, returning false when another worker
	// holds it. The caller leaves the event unacknowledged so it is retried
	// rather than interleaved. leaseCtx is cancelled if renewal fails or the
	// lease is lost, stopping the turn before it writes history as a stale
	// owner.
	Acquire(ctx context.Context, sessionID string) (leaseCtx context.Context, release func(), ok bool, err error)
}

// preAgentAttempts bounds retries of transient failures that happen *before*
// Agent work starts. Only that window is safely retryable, so the budget
// deliberately does not extend past it.
const preAgentAttempts = 3

// preAgentBackoff is the base delay between pre-Agent retries.
const preAgentBackoff = 500 * time.Millisecond

// processingLeaseTTL prevents a concurrent queue delivery or operator resend
// from entering the same processing record while one owner is active.
const processingLeaseTTL = 5 * time.Minute

// recordProgress advances the durable state machine. The caller must stop if
// this write fails: crossing a side-effect boundary without recording it can
// make a later retry repeat Agent or Telegram work.
func (o *Orchestrator) recordProgress(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, status agentsv1.TelegramProcessingStatus) error {
	if o.processing == nil || record == nil {
		return nil
	}
	record.Status = status
	if _, err := o.updateProcessing(ctx, record, leaseToken); err != nil {
		return fmt.Errorf("record telegram processing state %s: %w", status.String(), err)
	}
	return nil
}

// recordFailure marks a record failed, choosing between the safely-retryable
// and the uncertain terminal state.
//
// The distinction is the whole point of the state machine: before Agent work
// starts, a retry is free; once it may have run tools, a retry could repeat
// side effects, so the record is dead-lettered for an operator instead.
func (o *Orchestrator) recordFailure(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, err error, agentMayHaveRun bool) error {
	if o.processing == nil || record == nil {
		return nil
	}
	record.Error = sanitizeProcessingError(err)
	if agentMayHaveRun {
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN
		record.DeadLettered = true
	} else {
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED
	}
	if _, updateErr := o.updateProcessing(ctx, record, leaseToken); updateErr != nil {
		return fmt.Errorf("record telegram processing failure: %w", updateErr)
	}
	return nil
}

// persistOutput stores the complete Agent response and its delivery plan
// before delivery begins.
//
// This is what makes a delivery failure recoverable without another Agent
// run: the text already exists, so a retry is a send, not a re-computation.
func (o *Orchestrator) persistOutput(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, delivery *telegramsend.Delivery) error {
	if o.processing == nil || record == nil {
		return nil
	}
	record.Output = joinSegments(delivery)
	record.Segments = SegmentsToProto(delivery)
	record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER
	if _, err := o.updateProcessing(ctx, record, leaseToken); err != nil {
		return fmt.Errorf("persist telegram agent output: %w", err)
	}
	return nil
}

// syncDelivery writes back per-segment progress so a resend knows exactly
// what still needs to go out.
func (o *Orchestrator) syncDelivery(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, delivery *telegramsend.Delivery, deliverErr error) error {
	if o.processing == nil || record == nil {
		return nil
	}
	record.Segments = SegmentsToProto(delivery)
	switch {
	case deliverErr == nil:
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_SUCCEEDED
		record.Error = ""
		record.DeadLettered = false
	case deliveryUncertain(delivery):
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN
		record.Error = sanitizeProcessingError(deliverErr)
		record.DeadLettered = true
	default:
		// Delivery failed but the output survives, so this is retryable
		// without touching the Agent.
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED
		record.Error = sanitizeProcessingError(deliverErr)
	}
	if _, err := o.updateProcessing(ctx, record, leaseToken); err != nil {
		return fmt.Errorf("record telegram delivery state: %w", err)
	}
	return nil
}

func (o *Orchestrator) persistDeliveryProgress(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, delivery *telegramsend.Delivery) error {
	if o.processing == nil || record == nil {
		return nil
	}
	record.Segments = SegmentsToProto(delivery)
	if _, err := o.updateProcessing(ctx, record, leaseToken); err != nil {
		return fmt.Errorf("persist telegram delivery progress: %w", err)
	}
	return nil
}

func (o *Orchestrator) updateProcessing(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string) (*agentsv1.TelegramProcessingRecord, error) {
	if leaseToken == "" {
		return o.processing.Update(ctx, record)
	}
	return o.processing.UpdateClaimed(ctx, record, leaseToken)
}

func deliveryUncertain(delivery *telegramsend.Delivery) bool {
	if delivery == nil {
		return false
	}
	for _, segment := range delivery.Segments {
		if segment.Status == telegramsend.SegmentSending {
			return true
		}
	}
	return false
}

// SegmentsToProto exports per-segment delivery state for persistence and for
// the resend path in the application layer.
func SegmentsToProto(delivery *telegramsend.Delivery) []*agentsv1.TelegramDeliverySegment {
	if delivery == nil {
		return nil
	}
	out := make([]*agentsv1.TelegramDeliverySegment, 0, len(delivery.Segments))
	for _, segment := range delivery.Segments {
		out = append(out, &agentsv1.TelegramDeliverySegment{
			Index:     int32(segment.Index),
			Text:      segment.Text,
			Status:    string(segment.Status),
			MessageId: segment.MessageID,
			Error:     segment.Error,
		})
	}
	return out
}

// DeliveryFromRecord rebuilds the delivery plan from a persisted record, so a
// resend continues from unsent and failed segments instead of restarting.
func DeliveryFromRecord(record *agentsv1.TelegramProcessingRecord) *telegramsend.Delivery {
	segments := make([]telegramsend.Segment, 0, len(record.GetSegments()))
	for _, stored := range record.GetSegments() {
		segments = append(segments, telegramsend.Segment{
			Index:     int(stored.GetIndex()),
			Text:      stored.GetText(),
			Status:    telegramsend.SegmentStatus(stored.GetStatus()),
			MessageID: stored.GetMessageId(),
			Error:     stored.GetError(),
		})
	}
	return &telegramsend.Delivery{Segments: segments}
}

func joinSegments(delivery *telegramsend.Delivery) string {
	if delivery == nil {
		return ""
	}
	var out string
	for i, segment := range delivery.Segments {
		if i > 0 {
			out += "\n"
		}
		out += segment.Text
	}
	return out
}

// sanitizeProcessingError keeps the operator-visible reason and drops
// anything that could carry credential material or a full update body.
func sanitizeProcessingError(err error) string {
	if err == nil {
		return ""
	}
	const maxErrorRunes = 500
	message := err.Error()
	runes := []rune(message)
	if len(runes) > maxErrorRunes {
		return string(runes[:maxErrorRunes]) + "…"
	}
	return message
}

// newProcessingRecord builds the record for a freshly claimed event.
func newProcessingRecord(event *telegramqueue.Event) *agentsv1.TelegramProcessingRecord {
	return &agentsv1.TelegramProcessingRecord{
		WorkspaceId:         event.WorkspaceID,
		ChannelId:           event.ChannelID,
		DestinationId:       event.DestinationID,
		UpdateId:            event.UpdateID,
		DestinationRevision: event.DestinationRevision,
		Status:              agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_RECEIVED,
	}
}

// claimRecord creates or re-claims the record for an event.
//
// The returned action distinguishes fresh Agent work from resumable delivery
// and terminal acknowledgement.
func (o *Orchestrator) claimRecord(ctx context.Context, event *telegramqueue.Event) (*agentsv1.TelegramProcessingRecord, telegramprocessing.ClaimAction, string, error) {
	if o.processing == nil {
		return nil, telegramprocessing.ClaimRunAgent, "", nil
	}
	now := time.Now().UTC()
	leaseToken := uuid.NewString()
	ttl := o.processingLeaseTTL
	if ttl <= 0 {
		ttl = processingLeaseTTL
	}
	record, action, err := o.processing.Claim(ctx, newProcessingRecord(event), leaseToken, now, now.Add(ttl))
	if err != nil {
		return nil, telegramprocessing.ClaimAcknowledge, "", fmt.Errorf("claim telegram processing record: %w", err)
	}
	if action == telegramprocessing.ClaimAcknowledge {
		leaseToken = ""
	}
	return record, action, leaseToken, nil
}

func (o *Orchestrator) withProcessingHeartbeat(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string) (context.Context, func()) {
	if o.processing == nil || record == nil || leaseToken == "" {
		return ctx, func() {}
	}
	ttl := o.processingLeaseTTL
	if ttl <= 0 {
		ttl = processingLeaseTTL
	}
	leaseCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(ttl / 3)
		defer ticker.Stop()
		for {
			select {
			case <-leaseCtx.Done():
				return
			case <-ticker.C:
				expiresAt := time.Now().UTC().Add(ttl)
				if err := o.processing.RenewClaim(leaseCtx, record.GetWorkspaceId(), record.GetId(), leaseToken, expiresAt); err != nil {
					log.FromContext(ctx).Error("telegram processing claim lost", "record_id", record.GetId(), "err", err)
					cancel()
					return
				}
			}
		}
	}()
	var once sync.Once
	return leaseCtx, func() {
		once.Do(func() {
			cancel()
			<-done
		})
	}
}

func (o *Orchestrator) releaseProcessingClaim(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string) {
	if o.processing == nil || record == nil || leaseToken == "" {
		return
	}
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := o.processing.ReleaseClaim(releaseCtx, record.GetWorkspaceId(), record.GetId(), leaseToken); err != nil && !errors.Is(err, telegramprocessing.ErrLeaseLost) {
		log.FromContext(ctx).Warn("could not release telegram processing claim", "record_id", record.GetId(), "err", err)
	}
}

func (o *Orchestrator) resumeDelivery(ctx context.Context, event *telegramqueue.Event, record *agentsv1.TelegramProcessingRecord, leaseToken string) error {
	delivery := DeliveryFromRecord(record)
	if !delivery.Pending() {
		return o.syncDelivery(ctx, record, leaseToken, delivery, nil)
	}
	deliverErr := o.sender.DeliverSegments(ctx, event.WorkspaceID, record.GetDestinationId(), delivery,
		func(current *telegramsend.Delivery) error {
			return o.persistDeliveryProgress(ctx, record, leaseToken, current)
		})
	if syncErr := o.syncDelivery(ctx, record, leaseToken, delivery, deliverErr); syncErr != nil {
		return syncErr
	}
	if deliverErr != nil {
		return fmt.Errorf("resume telegram response delivery: %w", deliverErr)
	}
	return nil
}

// retryPreAgent runs a transient pre-Agent step with bounded backoff.
//
// Only this window retries automatically. Everything after it either succeeds
// or becomes a recorded state an operator can act on.
func (o *Orchestrator) retryPreAgent(ctx context.Context, what string, step func() error) error {
	var lastErr error
	for attempt := 1; attempt <= preAgentAttempts; attempt++ {
		if err := step(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt == preAgentAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * preAgentBackoff):
		}
	}
	return fmt.Errorf("%s failed after %d attempts: %w", what, preAgentAttempts, lastErr)
}

// ErrSessionBusy means another worker holds this session's lease.
var ErrSessionBusy = errors.New("another worker is handling this session")
