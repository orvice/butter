package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"

	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/telegramapi"
)

// MaxImageBytes is Telegram's own upload limit for photos. Enforcing it here
// too means an oversized file is refused before an Agent is invoked rather
// than after.
const MaxImageBytes int64 = 20 * 1024 * 1024

// supportedImageMIMEs is the set an Agent can actually consume. Anything else
// is refused up front: handing an unusable blob to a model produces a
// confident answer about nothing.
var supportedImageMIMEs = []string{
	"image/jpeg", "image/png", "image/webp", "image/gif", "image/heic", "image/heif",
}

// ErrUnsupportedImage means the download succeeded but the content is not an
// image an Agent can use.
var ErrUnsupportedImage = errors.New("unsupported image type")

// LargestPhoto picks the highest-resolution rendition Telegram offers.
// Telegram orders `photo` from smallest to largest.
func LargestPhoto(photos []telegramapi.PhotoSize) (telegramapi.PhotoSize, bool) {
	if len(photos) == 0 {
		return telegramapi.PhotoSize{}, false
	}
	largest := photos[0]
	for _, photo := range photos[1:] {
		if photo.Width*photo.Height >= largest.Width*largest.Height {
			largest = photo
		}
	}
	return largest, true
}

// DownloadPhoto fetches one photo and returns it as an Agent input part.
//
// This runs on the worker immediately before invocation, never during webhook
// acknowledgement: keeping the download here is what keeps the callback fast
// and keeps binary data out of Redis, which holds only the `file_id`.
func DownloadPhoto(ctx context.Context, client telegramapi.FileClient, photo telegramapi.PhotoSize) (*genai.Part, error) {
	// Telegram reports the size in the update, so an oversized photo can be
	// refused without spending the download at all.
	if photo.FileSize > MaxImageBytes {
		return nil, fmt.Errorf("photo is %d bytes, over the %d byte limit", photo.FileSize, MaxImageBytes)
	}

	file, err := client.GetFile(ctx, photo.FileID)
	if err != nil {
		return nil, fmt.Errorf("read telegram file metadata: %w", err)
	}
	if file.FileSize > MaxImageBytes {
		return nil, fmt.Errorf("photo is %d bytes, over the %d byte limit", file.FileSize, MaxImageBytes)
	}

	data, contentType, err := client.DownloadFile(ctx, file, MaxImageBytes)
	if err != nil {
		return nil, err
	}
	mime := detectImageMIME(file.FilePath, contentType, data)
	if !slices.Contains(supportedImageMIMEs, mime) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedImage, mime)
	}
	return genai.NewPartFromBytes(data, mime), nil
}

// detectImageMIME resolves the content type, preferring the file extension
// Telegram assigned, then the response header, then content sniffing.
//
// The extension comes first because Telegram's own path is authoritative for
// what it stored, while a CDN's Content-Type is frequently a generic
// application/octet-stream.
func detectImageMIME(filePath, contentType string, data []byte) string {
	switch strings.ToLower(path.Ext(filePath)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".heic":
		return "image/heic"
	case ".heif":
		return "image/heif"
	}
	if base, _, _ := strings.Cut(contentType, ";"); strings.HasPrefix(base, "image/") {
		return strings.TrimSpace(base)
	}
	if sniffed, _, _ := strings.Cut(http.DetectContentType(data), ";"); strings.HasPrefix(sniffed, "image/") {
		return strings.TrimSpace(sniffed)
	}
	return strings.TrimSpace(contentType)
}

// BuildInputParts assembles the ordered multimodal input for a turn.
//
// The caption leads because it is what the user wrote about the image; an
// Agent reading parts in order sees the instruction before the subject.
func BuildInputParts(caption string, image *genai.Part) []*genai.Part {
	var parts []*genai.Part
	if trimmed := strings.TrimSpace(caption); trimmed != "" {
		parts = append(parts, &genai.Part{Text: trimmed})
	}
	if image != nil {
		parts = append(parts, image)
	}
	return parts
}
