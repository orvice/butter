package telegramsend

// Segmentation and delivery tests (issue #264/#270): UTF-8 safety, natural
// boundaries, per-segment Markdown fallback, placeholder editing, ordered
// delivery, and resumable partial failure.

import (
	"strings"
	"testing"
	"unicode/utf8"

	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramapi/telegramtest"
)

func TestShortTextIsOneSegment(t *testing.T) {
	segments := SplitMessage("a short reply")
	if len(segments) != 1 || segments[0] != "a short reply" {
		t.Fatalf("segments = %#v", segments)
	}
}

// Telegram counts UTF-16 code units, where an emoji consumes two units. Every
// cut must also preserve valid UTF-8.
func TestSplittingRespectsTelegramUTF16Limit(t *testing.T) {
	text := strings.Repeat("🙂", MaxSegmentUTF16Units)

	segments := SplitMessage(text)
	if len(segments) < 2 {
		t.Fatalf("expected multiple segments, got %d", len(segments))
	}
	for i, segment := range segments {
		if !utf8.ValidString(segment) {
			t.Fatalf("segment %d is not valid UTF-8", i)
		}
		if units := telegramapi.UTF16Len(segment); units > MaxSegmentUTF16Units {
			t.Fatalf("segment %d has %d UTF-16 units, over the limit", i, units)
		}
	}
	if joined := strings.Join(segments, ""); joined != text {
		t.Error("segments did not reassemble into the original text")
	}
}

// A reply cut between paragraphs reads as a continuation; one cut mid-word
// reads as corruption.
func TestSplittingPrefersParagraphBoundaries(t *testing.T) {
	paragraph := strings.Repeat("word ", 400) // ~2000 runes
	text := paragraph + "\n\n" + paragraph + "\n\n" + paragraph

	segments := SplitMessage(text)
	if len(segments) < 2 {
		t.Fatalf("expected multiple segments, got %d", len(segments))
	}
	for i, segment := range segments[:len(segments)-1] {
		if strings.HasSuffix(segment, "wor") || strings.HasSuffix(segment, "wo") {
			t.Errorf("segment %d ends mid-word: %q", i, segment[len(segment)-10:])
		}
	}
}

// One unbroken token longer than a whole message still has to be delivered.
func TestUnbreakableTokenIsHardCut(t *testing.T) {
	text := strings.Repeat("x", MaxSegmentUTF16Units+500)

	segments := SplitMessage(text)
	if len(segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(segments))
	}
	if strings.Join(segments, "") != text {
		t.Error("a hard cut lost content")
	}
}

func TestEmptyTextProducesNoSegments(t *testing.T) {
	if got := SplitMessage("   \n\n"); len(got) != 0 {
		t.Fatalf("segments = %#v, want none", got)
	}
}

// --- Delivery ---------------------------------------------------------------

func TestSegmentsAreDeliveredInOrderAndAllCarryTheTopic(t *testing.T) {
	fx := newFixture(t)
	delivery := NewDelivery(strings.Repeat("word ", 2000), "", "9")

	if err := fx.sender.DeliverSegments(t.Context(), workspace, "dest-1", delivery); err != nil {
		t.Fatalf("DeliverSegments: %v", err)
	}
	sent := fx.bots.Sent()
	if len(sent) < 2 {
		t.Fatalf("sent %d messages, want a segmented response", len(sent))
	}
	for i, message := range sent {
		if message.Params.MessageThreadID != "42" {
			t.Errorf("segment %d left the topic: %q", i, message.Params.MessageThreadID)
		}
	}
	// Only the first segment quotes the inbound message.
	if sent[0].Params.ReplyToMessageID != "9" {
		t.Errorf("first segment reply_to = %q", sent[0].Params.ReplyToMessageID)
	}
	for i, message := range sent[1:] {
		if message.Params.ReplyToMessageID != "" {
			t.Errorf("segment %d quoted the inbound message", i+1)
		}
	}
	if delivery.Pending() {
		t.Error("delivery still reports pending segments after success")
	}
	if len(delivery.MessageIDs()) != len(sent) {
		t.Errorf("recorded %d message IDs for %d sends", len(delivery.MessageIDs()), len(sent))
	}
}

// The placeholder becomes the answer rather than lingering above it.
func TestPlaceholderIsEditedIntoTheFirstSegment(t *testing.T) {
	fx := newFixture(t)
	delivery := NewDelivery(strings.Repeat("word ", 2000), "555", "9")

	if err := fx.sender.DeliverSegments(t.Context(), workspace, "dest-1", delivery); err != nil {
		t.Fatalf("DeliverSegments: %v", err)
	}
	sent := fx.bots.Sent()
	if sent[0].Edit == nil {
		t.Fatal("the first segment was sent as a new message instead of editing the placeholder")
	}
	if sent[0].Edit.MessageID != "555" {
		t.Errorf("edited message %q, want the placeholder", sent[0].Edit.MessageID)
	}
	if delivery.Segments[0].MessageID != "555" {
		t.Errorf("first segment recorded message %q", delivery.Segments[0].MessageID)
	}
	for i, message := range sent[1:] {
		if message.Edit != nil {
			t.Errorf("segment %d was an edit; later segments must be new messages", i+1)
		}
	}
}

// One segment's markup problem must not drop the segments after it.
func TestMarkdownFallbackIsPerSegment(t *testing.T) {
	fx := newFixture(t)
	// Reject only the first attempt, which is the first segment's MarkdownV2.
	fx.bots.OnSend(func(attempt int, _ telegramapi.SendMessageParams) error {
		if attempt == 1 {
			return telegramtest.MarkdownRejection()
		}
		return nil
	})
	delivery := NewDelivery(strings.Repeat("word ", 2000), "", "")

	if err := fx.sender.DeliverSegments(t.Context(), workspace, "dest-1", delivery); err != nil {
		t.Fatalf("DeliverSegments: %v", err)
	}
	if !delivery.Segments[0].PlainTextFallback {
		t.Error("expected the first segment to record its fallback")
	}
	if len(delivery.Segments) > 1 && delivery.Segments[1].PlainTextFallback {
		t.Error("a later segment inherited the first segment's fallback")
	}
	if delivery.Pending() {
		t.Error("the fallback left segments undelivered")
	}
}

// A failed segment leaves the rest recoverable without re-running the Agent.
func TestPartialFailureLeavesLaterSegmentsResumable(t *testing.T) {
	fx := newFixture(t)
	failing := true
	fx.bots.OnSend(func(attempt int, _ telegramapi.SendMessageParams) error {
		if failing && attempt >= 2 {
			return &telegramapi.APIError{Code: 500, Description: "Internal Server Error"}
		}
		return nil
	})
	delivery := NewDelivery(strings.Repeat("word ", 2000), "", "")

	if err := fx.sender.DeliverSegments(t.Context(), workspace, "dest-1", delivery); err == nil {
		t.Fatal("expected the delivery to fail")
	}
	if delivery.Segments[0].Status != SegmentSent {
		t.Fatal("the segment that landed was not recorded as sent")
	}
	if !delivery.Pending() {
		t.Fatal("expected the remaining segments to stay pending")
	}
	delivered := len(delivery.MessageIDs())

	// A retry resends only what is still pending; nothing is duplicated and
	// the Agent is not consulted again.
	failing = false
	fx.bots.Reset()
	if err := fx.sender.DeliverSegments(t.Context(), workspace, "dest-1", delivery); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if delivery.Pending() {
		t.Error("the retry left segments undelivered")
	}
	if resent := len(fx.bots.Sent()); resent >= len(delivery.Segments)+delivered {
		t.Errorf("the retry resent %d messages; already-delivered segments were duplicated", resent)
	}
}

func TestDeliverTextRejectsAnEmptyResponse(t *testing.T) {
	fx := newFixture(t)

	if _, err := fx.sender.DeliverText(t.Context(), workspace, "dest-1", "   ", ""); err == nil {
		t.Fatal("expected an empty response to be rejected")
	}
	if len(fx.bots.Sent()) != 0 {
		t.Error("an empty response reached Telegram")
	}
}
