package telegram

// Photo tests (issue #264/#270): the download happens on the worker, size and
// MIME are validated before an Agent sees anything, and a failure produces a
// clear reply instead of a confident answer about an image that never
// arrived.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"go.orx.me/apps/butter/internal/telegramapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeFileClient serves photo bytes without Telegram.
type fakeFileClient struct {
	file        telegramapi.File
	data        []byte
	contentType string
	getErr      error
	downloadErr error
	downloads   int
}

func (c *fakeFileClient) GetFile(context.Context, string) (telegramapi.File, error) {
	if c.getErr != nil {
		return telegramapi.File{}, c.getErr
	}
	return c.file, nil
}

func (c *fakeFileClient) DownloadFile(_ context.Context, _ telegramapi.File, limit int64) ([]byte, string, error) {
	c.downloads++
	if c.downloadErr != nil {
		return nil, "", c.downloadErr
	}
	if int64(len(c.data)) > limit {
		return nil, "", fmt.Errorf("telegram file exceeds the %d byte limit", limit)
	}
	return c.data, c.contentType, nil
}

// pngBytes is a minimal valid PNG header, enough for content sniffing.
var pngBytes = append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 64)...)

func TestLargestPhotoPicksTheHighestResolution(t *testing.T) {
	photos := []telegramapi.PhotoSize{
		{FileID: "small", Width: 90, Height: 90},
		{FileID: "large", Width: 1280, Height: 960},
		{FileID: "medium", Width: 320, Height: 240},
	}

	largest, ok := LargestPhoto(photos)
	if !ok || largest.FileID != "large" {
		t.Fatalf("largest = %+v", largest)
	}
	if _, ok := LargestPhoto(nil); ok {
		t.Error("an empty photo list must report no photo")
	}
}

func TestDownloadPhotoProducesAnImagePart(t *testing.T) {
	client := &fakeFileClient{
		file: telegramapi.File{FileID: "f", FilePath: "photos/file_1.jpg", FileSize: 1024},
		data: []byte("jpeg-bytes"),
	}

	part, err := DownloadPhoto(t.Context(), client, telegramapi.PhotoSize{FileID: "f", FileSize: 1024})
	if err != nil {
		t.Fatalf("DownloadPhoto: %v", err)
	}
	if part.InlineData == nil || part.InlineData.MIMEType != "image/jpeg" {
		t.Fatalf("part = %+v", part)
	}
}

// Telegram reports the size in the update, so an oversized photo is refused
// without spending the download at all.
func TestOversizedPhotoIsRefusedBeforeDownloading(t *testing.T) {
	client := &fakeFileClient{file: telegramapi.File{FilePath: "photos/f.jpg"}}

	_, err := DownloadPhoto(t.Context(), client,
		telegramapi.PhotoSize{FileID: "f", FileSize: MaxImageBytes + 1})
	if err == nil {
		t.Fatal("expected the oversized photo to be refused")
	}
	if client.downloads != 0 {
		t.Error("an oversized photo was downloaded anyway")
	}
}

func TestUnsupportedMIMEIsRefused(t *testing.T) {
	client := &fakeFileClient{
		file:        telegramapi.File{FileID: "f", FilePath: "documents/file_1.pdf", FileSize: 10},
		data:        []byte("%PDF-1.4"),
		contentType: "application/pdf",
	}

	_, err := DownloadPhoto(t.Context(), client, telegramapi.PhotoSize{FileID: "f", FileSize: 10})
	if !errors.Is(err, ErrUnsupportedImage) {
		t.Fatalf("err = %v, want ErrUnsupportedImage", err)
	}
}

// Telegram's own file extension is authoritative; a CDN's generic
// Content-Type must not win over it.
func TestMIMEPrefersTheTelegramFileExtension(t *testing.T) {
	client := &fakeFileClient{
		file:        telegramapi.File{FileID: "f", FilePath: "photos/file_1.png", FileSize: 10},
		data:        pngBytes,
		contentType: "application/octet-stream",
	}

	part, err := DownloadPhoto(t.Context(), client, telegramapi.PhotoSize{FileID: "f", FileSize: 10})
	if err != nil {
		t.Fatalf("DownloadPhoto: %v", err)
	}
	if part.InlineData.MIMEType != "image/png" {
		t.Errorf("mime = %q", part.InlineData.MIMEType)
	}
}

// The caption leads so an Agent reading parts in order sees the instruction
// before the subject.
func TestInputPartsOrderCaptionBeforeImage(t *testing.T) {
	client := &fakeFileClient{
		file: telegramapi.File{FileID: "f", FilePath: "photos/file_1.jpg", FileSize: 10},
		data: []byte("jpeg"),
	}
	image, err := DownloadPhoto(t.Context(), client, telegramapi.PhotoSize{FileID: "f", FileSize: 10})
	if err != nil {
		t.Fatalf("DownloadPhoto: %v", err)
	}

	parts := BuildInputParts("what is this?", image)
	if len(parts) != 2 {
		t.Fatalf("parts = %d, want caption and image", len(parts))
	}
	if parts[0].Text != "what is this?" {
		t.Errorf("first part = %+v, want the caption", parts[0])
	}
	if parts[1].InlineData == nil {
		t.Error("second part is not the image")
	}

	// A caption-only message is text; an image with no caption is just the
	// image.
	if got := BuildInputParts("", image); len(got) != 1 || got[0].InlineData == nil {
		t.Errorf("image-only parts = %#v", got)
	}
	if got := BuildInputParts("just text", nil); len(got) != 1 || got[0].Text != "just text" {
		t.Errorf("caption-only parts = %#v", got)
	}
}

// --- Orchestration ---------------------------------------------------------

// photoUpdate builds an update carrying a photo and optional caption.
func photoUpdate(caption string) string {
	return fmt.Sprintf(`{"update_id":1,"message":{"message_id":9,"message_thread_id":42,"is_topic_message":true,"from":%s,"chat":{"id":-100,"type":"supergroup"},"caption":%q,"photo":[{"file_id":"small","width":90,"height":90,"file_size":100},{"file_id":"large","width":1280,"height":960,"file_size":2048}]}}`,
		realUser, caption)
}

// The queued event carries only Telegram metadata; the download happens here.
func TestPhotoIsDownloadedAtInvocationTime(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	client := &fakeFileClient{
		file: telegramapi.File{FileID: "large", FilePath: "photos/file_1.jpg", FileSize: 2048},
		data: []byte("jpeg-bytes"),
	}
	fx.orchestrator.SetFileClientFactory(
		func(context.Context, string, string) (telegramapi.FileClient, error) { return client, nil })

	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(photoUpdate("what is this?"))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if client.downloads != 1 {
		t.Fatalf("downloads = %d, want exactly one at invocation time", client.downloads)
	}
	if len(fx.agents.calls) != 1 {
		t.Fatalf("agent invoked %d times", len(fx.agents.calls))
	}
	if fx.agents.calls[0].text != "what is this?" {
		t.Errorf("caption = %q", fx.agents.calls[0].text)
	}
}

// A photo that cannot be fetched must not reach the Agent as a caption alone:
// it would answer confidently about an image it never saw.
func TestImageFailureStopsBeforeTheAgent(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	client := &fakeFileClient{
		file:        telegramapi.File{FileID: "large", FilePath: "photos/file_1.jpg", FileSize: 2048},
		downloadErr: errors.New("connection reset"),
	}
	fx.orchestrator.SetFileClientFactory(
		func(context.Context, string, string) (telegramapi.FileClient, error) { return client, nil })

	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(photoUpdate("what is this?"))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.agents.calls) != 0 {
		t.Fatal("the agent was invoked despite the image failing to download")
	}
	sent := fx.bots.Sent()
	if len(sent) != 1 {
		t.Fatalf("sent %d messages, want a failure notice", len(sent))
	}
	if sent[0].Params.MessageThreadID != "42" {
		t.Errorf("the failure notice left the topic: %q", sent[0].Params.MessageThreadID)
	}
	if !strings.Contains(strings.ToLower(sent[0].Params.Text), "image") {
		t.Errorf("failure notice = %q, want it to name the image", sent[0].Params.Text)
	}
}

// Media groups are deliberately not aggregated: each update is its own
// interaction. Documented as a follow-up in the PRD.
func TestMediaGroupMembersAreSeparateInteractions(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	client := &fakeFileClient{
		file: telegramapi.File{FileID: "large", FilePath: "photos/file_1.jpg", FileSize: 2048},
		data: []byte("jpeg-bytes"),
	}
	fx.orchestrator.SetFileClientFactory(
		func(context.Context, string, string) (telegramapi.FileClient, error) { return client, nil })

	for i := range 2 {
		event := fx.eventForStored(photoUpdate("album"))
		event.UpdateID = int64(i + 1)
		if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
			t.Fatalf("Handle %d: %v", i, err)
		}
	}
	if len(fx.agents.calls) != 2 {
		t.Fatalf("agent invoked %d times, want one per update", len(fx.agents.calls))
	}
}

// A photo destination still respects the destination's trigger policy.
func TestPhotoRespectsTriggerMode(t *testing.T) {
	fx := newOrchestratorFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		c.TriggerMode = agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_COMMAND
	})
	fx.orchestrator.SetFileClientFactory(
		func(context.Context, string, string) (telegramapi.FileClient, error) {
			t.Fatal("a non-triggered photo must not be downloaded")
			return nil, nil
		})

	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(photoUpdate("hi"))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.agents.calls) != 0 {
		t.Fatal("a non-triggered photo invoked the agent")
	}
}
