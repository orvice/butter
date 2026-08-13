package telegramapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// File is Telegram's metadata for an uploaded file.
type File struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
}

// FileClient downloads media a Telegram user uploaded.
//
// It is separate from Client because only the worker that is about to invoke
// an Agent needs it: media is deliberately not fetched during webhook
// acknowledgement, which keeps the callback fast and keeps binary data out of
// Redis entirely.
type FileClient interface {
	GetFile(ctx context.Context, fileID string) (File, error)
	// DownloadFile fetches the bytes at a File's path, refusing anything
	// larger than limit. It returns the observed Content-Type so the caller
	// can validate the MIME before handing it to an Agent.
	DownloadFile(ctx context.Context, file File, limit int64) (data []byte, contentType string, err error)
}

func (c *HTTPClient) GetFile(ctx context.Context, fileID string) (File, error) {
	var result File
	if err := c.call(ctx, "getFile", map[string]any{"file_id": fileID}, &result); err != nil {
		return File{}, err
	}
	return result, nil
}

func (c *HTTPClient) DownloadFile(ctx context.Context, file File, limit int64) ([]byte, string, error) {
	if file.FilePath == "" {
		return nil, "", fmt.Errorf("telegram file %q has no path", file.FileID)
	}
	url := c.baseURL + "/file/bot" + c.token + "/" + file.FilePath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("build file download request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		// The URL carries the Bot Token, so never wrap the transport error.
		return nil, "", fmt.Errorf("telegram file download failed: %s",
			redactToken(err.Error(), c.token))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("telegram file download returned status %d", resp.StatusCode)
	}

	// Read one byte past the limit so an oversized file is detected rather
	// than silently truncated into a corrupt image.
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", fmt.Errorf("read telegram file: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, "", fmt.Errorf("telegram file exceeds the %d byte limit", limit)
	}
	return data, strings.TrimSpace(resp.Header.Get("Content-Type")), nil
}
