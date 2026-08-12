package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/telegramqueue"
	"go.orx.me/apps/butter/internal/telegramsend"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// SessionGuard serializes work inside one derived session.
type SessionGuard interface {
	// Acquire takes the session lease, returning false when another worker
	// holds it. The caller leaves the event unacknowledged so it is retried
	// rather than interleaved.
	Acquire(ctx context.Context, sessionID string) (release func(), ok bool, err error)
}

// preAgentAttempts bounds retries of transient failures that happen *before*
// Agent work starts. Only that window is safely retryable, so the budget
// deliberately does not extend past it.
const preAgentAttempts = 3

// preAgentBackoff is the base delay between pre-Agent retries.
const preAgentBackoff = 500 * time.Millisecond

// recordProgress advances the processing record, tolerating write failures:
// losing the bookkeeping is strictly better than failing work that succeeded.
func (o *Orchestrator) recordProgress(ctx context.Context, record *agentsv1.TelegramProcessingRecord, status agentsv1.TelegramProcessingStatus) {
	if o.processing == nil || record == nil {
		return
	}
	record.Status = status
	if _, err := o.processing.Update(ctx, record); err != nil {
		log.FromContext(ctx).Warn("could not record telegram processing state",
			"record_id", record.GetId(), "status", status.String(), "err", err)
	}
}

// recordFailure marks a record failed, choosing between the safely-retryable
// and the uncertain terminal state.
//
// The distinction is the whole point of the state machine: before Agent work
// starts, a retry is free; once it may have run tools, a retry could repeat
// side effects, so the record is dead-lettered for an operator instead.
func (o *Orchestrator) recordFailure(ctx context.Context, record *agentsv1.TelegramProcessingRecord, err error, agentMayHaveRun bool) {
	if o.processing == nil || record == nil {
		return
	}
	record.Error = sanitizeProcessingError(err)
	if agentMayHaveRun {
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN
		record.DeadLettered = true
	} else {
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED
	}
	if _, updateErr := o.processing.Update(ctx, record); updateErr != nil {
		log.FromContext(ctx).Warn("could not record telegram processing failure",
			"record_id", record.GetId(), "err", updateErr)
	}
}

// persistOutput stores the complete Agent response and its delivery plan
// before delivery begins.
//
// This is what makes a delivery failure recoverable without another Agent
// run: the text already exists, so a retry is a send, not a re-computation.
func (o *Orchestrator) persistOutput(ctx context.Context, record *agentsv1.TelegramProcessingRecord, delivery *telegramsend.Delivery) error {
	if o.processing == nil || record == nil {
		return nil
	}
	record.Output = joinSegments(delivery)
	record.Segments = SegmentsToProto(delivery)
	record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER
	if _, err := o.processing.Update(ctx, record); err != nil {
		return fmt.Errorf("persist telegram agent output: %w", err)
	}
	return nil
}

// syncDelivery writes back per-segment progress so a resend knows exactly
// what still needs to go out.
func (o *Orchestrator) syncDelivery(ctx context.Context, record *agentsv1.TelegramProcessingRecord, delivery *telegramsend.Delivery, deliverErr error) {
	if o.processing == nil || record == nil {
		return
	}
	record.Segments = SegmentsToProto(delivery)
	switch {
	case deliverErr == nil:
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_SUCCEEDED
		record.Error = ""
	default:
		// Delivery failed but the output survives, so this is retryable
		// without touching the Agent.
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED
		record.Error = sanitizeProcessingError(deliverErr)
	}
	if _, err := o.processing.Update(ctx, record); err != nil {
		log.FromContext(ctx).Warn("could not record telegram delivery state",
			"record_id", record.GetId(), "err", err)
	}
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
// The `claimed` result is false for an update that already succeeded, which
// is how a duplicate Telegram delivery is acknowledged without a second Agent
// run or a second message.
func (o *Orchestrator) claimRecord(ctx context.Context, event *telegramqueue.Event) (*agentsv1.TelegramProcessingRecord, bool, error) {
	if o.processing == nil {
		return nil, true, nil
	}
	record, claimed, err := o.processing.Claim(ctx, newProcessingRecord(event))
	if err != nil {
		return nil, false, fmt.Errorf("claim telegram processing record: %w", err)
	}
	return record, claimed, nil
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
