package discord

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"butterfly.orx.me/core/log"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/genai"
)

const maxImageSize = 20 * 1024 * 1024 // 20MB

// hasImageAttachments returns true if the message contains any image attachments.
func hasImageAttachments(m *discordgo.MessageCreate) bool {
	for _, a := range m.Attachments {
		if isImageAttachment(a) {
			return true
		}
	}
	return false
}

// isImageAttachment returns true if the attachment has an image content type.
func isImageAttachment(a *discordgo.MessageAttachment) bool {
	ct := strings.ToLower(a.ContentType)
	return strings.HasPrefix(ct, "image/")
}

// downloadAttachment downloads a Discord image attachment and returns it as a genai.Part.
func downloadAttachment(ctx context.Context, a *discordgo.MessageAttachment) (*genai.Part, error) {
	logger := log.FromContext(ctx)

	if a.Size > maxImageSize {
		return nil, fmt.Errorf("attachment %q exceeds size limit: %d bytes (max %d)", a.Filename, a.Size, maxImageSize)
	}

	resp, err := http.Get(a.URL)
	if err != nil {
		return nil, fmt.Errorf("downloading attachment %q: %w", a.Filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading attachment %q: HTTP %d", a.Filename, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading attachment %q data: %w", a.Filename, err)
	}

	if len(data) > maxImageSize {
		return nil, fmt.Errorf("attachment %q exceeds size limit after download: %d bytes", a.Filename, len(data))
	}

	mimeType := a.ContentType
	if mimeType == "" {
		mimeType = "image/png"
	}

	logger.Debug("downloaded discord attachment",
		"filename", a.Filename,
		"size", len(data),
		"mime_type", mimeType,
	)

	return genai.NewPartFromBytes(data, mimeType), nil
}

// buildMessageParts constructs genai.Part slice from a Discord message.
// It includes msg.Content as text and downloads any image attachments as blob parts.
func buildMessageParts(ctx context.Context, m *discordgo.MessageCreate) []*genai.Part {
	logger := log.FromContext(ctx)
	var parts []*genai.Part

	if m.Content != "" {
		parts = append(parts, genai.NewPartFromText(m.Content))
	}

	for _, a := range m.Attachments {
		if !isImageAttachment(a) {
			continue
		}

		part, err := downloadAttachment(ctx, a)
		if err != nil {
			logger.Error("failed to download discord image attachment",
				"filename", a.Filename,
				"err", err,
			)
			continue // skip failed attachments
		}
		parts = append(parts, part)
	}

	return parts
}
