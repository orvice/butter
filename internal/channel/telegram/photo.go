package telegram

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"butterfly.orx.me/core/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"google.golang.org/genai"
)

const maxImageSize = 20 * 1024 * 1024 // 20MB

// downloadPhoto downloads the largest photo from a Telegram message and returns it as a genai.Part.
// Returns nil if the photo cannot be downloaded or exceeds the size limit.
func downloadPhoto(ctx context.Context, b *bot.Bot, photos []models.PhotoSize) (*genai.Part, error) {
	if len(photos) == 0 {
		return nil, fmt.Errorf("no photos provided")
	}

	// Pick the largest photo (last element in the slice).
	largest := photos[len(photos)-1]

	if largest.FileSize > maxImageSize {
		return nil, fmt.Errorf("photo exceeds size limit: %d bytes (max %d)", largest.FileSize, maxImageSize)
	}

	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: largest.FileID})
	if err != nil {
		return nil, fmt.Errorf("getting file info: %w", err)
	}

	downloadURL := b.FileDownloadLink(file)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("downloading photo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading photo: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading photo data: %w", err)
	}

	if len(data) > maxImageSize {
		return nil, fmt.Errorf("photo exceeds size limit after download: %d bytes", len(data))
	}

	mimeType := detectMIMEType(file.FilePath, resp.Header.Get("Content-Type"))
	return genai.NewPartFromBytes(data, mimeType), nil
}

// buildMessageParts constructs genai.Part slice from a Telegram message.
// For photo messages, it downloads the photo and uses msg.Caption as text.
// For text messages, it uses msg.Text.
func buildMessageParts(ctx context.Context, b *bot.Bot, msg *models.Message) ([]*genai.Part, error) {
	logger := log.FromContext(ctx)
	var parts []*genai.Part

	if len(msg.Photo) > 0 {
		// Photo message: use caption as text part.
		if msg.Caption != "" {
			parts = append(parts, genai.NewPartFromText(msg.Caption))
		}

		photoPart, err := downloadPhoto(ctx, b, msg.Photo)
		if err != nil {
			logger.Error("failed to download telegram photo", "err", err)
			return nil, fmt.Errorf("failed to process image: %w", err)
		}
		parts = append(parts, photoPart)
	} else {
		// Text-only message.
		if msg.Text != "" {
			parts = append(parts, genai.NewPartFromText(msg.Text))
		}
	}

	return parts, nil
}

// detectMIMEType determines the MIME type from file path extension or Content-Type header.
func detectMIMEType(filePath, contentType string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	}

	if contentType != "" && strings.HasPrefix(contentType, "image/") {
		return contentType
	}

	return "image/jpeg" // default for Telegram photos
}
