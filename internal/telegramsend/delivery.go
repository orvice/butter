package telegramsend

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/telegramapi"
)

// SegmentStatus is the delivery state of one response segment.
type SegmentStatus string

const (
	// SegmentPending has not been attempted yet.
	SegmentPending SegmentStatus = "pending"
	// SegmentSent reached Telegram.
	SegmentSent SegmentStatus = "sent"
	// SegmentFailed was attempted and rejected.
	SegmentFailed SegmentStatus = "failed"
)

// Segment is one piece of a response plus what happened to it.
type Segment struct {
	// Index is the segment's position in the response, zero-based.
	Index int
	// Text is the segment body.
	Text string
	// Status reports whether it has been delivered.
	Status SegmentStatus
	// MessageID is Telegram's message ID once sent.
	MessageID string
	// Error is a sanitized failure summary.
	Error string
	// PlainTextFallback records that MarkdownV2 was rejected for this
	// segment specifically.
	PlainTextFallback bool
}

// Delivery is the full response together with per-segment progress.
//
// It is a value the caller persists *before* delivery starts. That is what
// makes a failed segment recoverable: a retry resends only what is still
// pending or failed, and never re-runs the Agent to reproduce text that was
// already produced.
type Delivery struct {
	Segments []Segment
	// PlaceholderMessageID is a "processing" message to edit into the first
	// segment rather than leaving it stranded above the answer.
	PlaceholderMessageID string
	// ReplyToMessageID quotes the inbound message on the first segment only.
	ReplyToMessageID string
}

// NewDelivery splits text into a deliverable, ordered plan.
func NewDelivery(text, placeholderMessageID, replyToMessageID string) *Delivery {
	parts := SplitMessage(text)
	segments := make([]Segment, 0, len(parts))
	for i, part := range parts {
		segments = append(segments, Segment{Index: i, Text: part, Status: SegmentPending})
	}
	return &Delivery{
		Segments:             segments,
		PlaceholderMessageID: placeholderMessageID,
		ReplyToMessageID:     replyToMessageID,
	}
}

// Pending reports whether anything is still undelivered.
func (d *Delivery) Pending() bool {
	for _, segment := range d.Segments {
		if segment.Status != SegmentSent {
			return true
		}
	}
	return false
}

// MessageIDs lists the Telegram message IDs produced so far, in order.
func (d *Delivery) MessageIDs() []string {
	var ids []string
	for _, segment := range d.Segments {
		if segment.MessageID != "" {
			ids = append(ids, segment.MessageID)
		}
	}
	return ids
}

// DeliverSegments sends every segment that is not yet delivered, in order.
//
// It mutates the Delivery in place so the caller can persist progress after
// the call whether it succeeded or not. A failure stops the run — later
// segments would arrive out of order otherwise — but leaves them pending, so
// a retry continues from where it stopped instead of duplicating what already
// landed.
func (s *Sender) DeliverSegments(ctx context.Context, workspaceID, destinationID string, delivery *Delivery) error {
	resolved, err := s.Resolve(ctx, workspaceID, destinationID)
	if err != nil {
		return err
	}
	logger := log.FromContext(ctx)

	for i := range delivery.Segments {
		segment := &delivery.Segments[i]
		if segment.Status == SegmentSent {
			continue
		}

		// The processing placeholder becomes the first segment rather than
		// lingering above the answer. Only the first segment can be an edit;
		// the rest are new messages so ordering is Telegram's to keep.
		if i == 0 && delivery.PlaceholderMessageID != "" {
			if err := s.editSegment(ctx, resolved, delivery, segment); err != nil {
				s.recordOutcome(ctx, workspaceID, resolved.Destination, err)
				return err
			}
			continue
		}
		if err := s.sendSegment(ctx, resolved, delivery, segment, i == 0); err != nil {
			s.recordOutcome(ctx, workspaceID, resolved.Destination, err)
			return err
		}
	}

	s.recordOutcome(ctx, workspaceID, resolved.Destination, nil)
	logger.Debug("telegram response delivered",
		"destination_id", destinationID, "segments", len(delivery.Segments))
	return nil
}

func (s *Sender) sendSegment(ctx context.Context, resolved *Resolved, delivery *Delivery, segment *Segment, first bool) error {
	params := telegramapi.SendMessageParams{
		ChatID: resolved.Destination.GetChatId(),
		// Every segment carries the topic. A later segment that lost it
		// would strand half the answer in the group's general chat.
		MessageThreadID: resolved.Destination.GetMessageThreadId(),
		Text:            ToTelegramMarkdownV2(segment.Text),
		ParseMode:       telegramapi.ParseModeMarkdownV2,
	}
	if first {
		params.ReplyToMessageID = delivery.ReplyToMessageID
	}

	sent, err := s.sendWithRetry(ctx, resolved.Client, params)
	if err != nil && isFormattingRejection(err) {
		// Fall back per segment: one segment's markup problem must not drop
		// the segments after it.
		params.Text = segment.Text
		params.ParseMode = telegramapi.ParseModeNone
		sent, err = s.sendWithRetry(ctx, resolved.Client, params)
		segment.PlainTextFallback = err == nil
	}
	if err != nil {
		segment.Status = SegmentFailed
		segment.Error = sanitizeError(err)
		return fmt.Errorf("deliver segment %d: %w", segment.Index, err)
	}
	segment.Status = SegmentSent
	segment.MessageID = telegramapi.FormatMessageID(sent.ID)
	segment.Error = ""
	return nil
}

func (s *Sender) editSegment(ctx context.Context, resolved *Resolved, delivery *Delivery, segment *Segment) error {
	params := telegramapi.EditMessageParams{
		ChatID:    resolved.Destination.GetChatId(),
		MessageID: delivery.PlaceholderMessageID,
		Text:      ToTelegramMarkdownV2(segment.Text),
		ParseMode: telegramapi.ParseModeMarkdownV2,
	}
	_, err := resolved.Client.EditMessageText(ctx, params)
	if err != nil && isFormattingRejection(err) {
		params.Text = segment.Text
		params.ParseMode = telegramapi.ParseModeNone
		_, err = resolved.Client.EditMessageText(ctx, params)
		segment.PlainTextFallback = err == nil
	}
	if err != nil {
		// The placeholder is a convenience, not the answer. If it cannot be
		// edited — it may have been deleted — send the segment instead so the
		// user still gets the response.
		log.FromContext(ctx).Warn("could not edit the telegram placeholder; sending instead",
			"destination_id", resolved.Destination.GetId(), "err", err)
		delivery.PlaceholderMessageID = ""
		return s.sendSegment(ctx, resolved, delivery, segment, true)
	}
	segment.Status = SegmentSent
	segment.MessageID = delivery.PlaceholderMessageID
	segment.Error = ""
	return nil
}

// SendProcessing posts a placeholder that a later DeliverSegments call edits
// into the first segment. An empty return means no placeholder was created,
// which the delivery path handles by sending normally.
func (s *Sender) SendProcessing(ctx context.Context, workspaceID, destinationID, text, replyToMessageID string) (string, error) {
	resolved, err := s.Resolve(ctx, workspaceID, destinationID)
	if err != nil {
		return "", err
	}
	sent, err := s.sendWithRetry(ctx, resolved.Client, telegramapi.SendMessageParams{
		ChatID:           resolved.Destination.GetChatId(),
		MessageThreadID:  resolved.Destination.GetMessageThreadId(),
		Text:             text,
		ParseMode:        telegramapi.ParseModeNone,
		ReplyToMessageID: replyToMessageID,
	})
	if err != nil {
		return "", fmt.Errorf("send telegram processing placeholder: %w", err)
	}
	return telegramapi.FormatMessageID(sent.ID), nil
}

// ErrNoSegments reports that there was nothing to deliver.
var ErrNoSegments = errors.New("response produced no segments")

// DeliverText is the convenience path for a response with no placeholder: it
// splits, delivers, and returns the resulting message IDs.
func (s *Sender) DeliverText(ctx context.Context, workspaceID, destinationID, text, replyToMessageID string) ([]string, error) {
	if strings.TrimSpace(text) == "" {
		return nil, ErrNoSegments
	}
	delivery := NewDelivery(text, "", replyToMessageID)
	if err := s.DeliverSegments(ctx, workspaceID, destinationID, delivery); err != nil {
		return delivery.MessageIDs(), err
	}
	return delivery.MessageIDs(), nil
}
