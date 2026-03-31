package telegram

import "testing"

func TestDetectMIMEType(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		contentType string
		want        string
	}{
		{
			name:     "jpeg extension",
			filePath: "photos/file_123.jpg",
			want:     "image/jpeg",
		},
		{
			name:     "png extension",
			filePath: "photos/file_123.png",
			want:     "image/png",
		},
		{
			name:     "gif extension",
			filePath: "photos/file_123.gif",
			want:     "image/gif",
		},
		{
			name:     "webp extension",
			filePath: "photos/file_123.webp",
			want:     "image/webp",
		},
		{
			name:        "no extension with content-type",
			filePath:    "photos/file_123",
			contentType: "image/png",
			want:        "image/png",
		},
		{
			name:     "no extension no content-type defaults to jpeg",
			filePath: "photos/file_123",
			want:     "image/jpeg",
		},
		{
			name:        "non-image content-type ignored without image prefix",
			filePath:    "photos/file_123",
			contentType: "application/octet-stream",
			want:        "image/jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectMIMEType(tt.filePath, tt.contentType)
			if got != tt.want {
				t.Errorf("detectMIMEType(%q, %q) = %q, want %q", tt.filePath, tt.contentType, got, tt.want)
			}
		})
	}
}
