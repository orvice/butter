package mongo

import (
	"testing"

	"google.golang.org/genai"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name    string
		content *genai.Content
		want    string
	}{
		{
			name:    "single text part",
			content: genai.NewContentFromText("hello world", "model"),
			want:    "hello world",
		},
		{
			name: "multiple text parts",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "hello"},
					{Text: "world"},
				},
			},
			want: "hello world",
		},
		{
			name: "empty parts skipped",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "hello"},
					{Text: ""},
					{Text: "world"},
				},
			},
			want: "hello world",
		},
		{
			name: "no text parts",
			content: &genai.Content{
				Parts: []*genai.Part{{}},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractText(tt.content)
			if got != tt.want {
				t.Errorf("extractText() = %q, want %q", got, tt.want)
			}
		})
	}
}
