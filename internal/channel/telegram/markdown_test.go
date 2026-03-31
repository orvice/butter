package telegram

import "testing"

func TestMarkdownToTelegramMarkdownV2(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{
			name: "plain text",
			in:   "Hello world",
		},
		{
			name: "bold",
			in:   "This is **bold** text",
		},
		{
			name: "italic",
			in:   "This is *italic* text",
		},
		{
			name: "inline code",
			in:   "Use `fmt.Println` here",
		},
		{
			name: "strikethrough",
			in:   "This is ~~deleted~~ text",
		},
		{
			name: "link",
			in:   "Visit [Google](https://google.com) now",
		},
		{
			name: "fenced code block",
			in:   "Example:\n```go\nfmt.Println(\"hello\")\n```\nDone.",
		},
		{
			name: "fenced code block no lang",
			in:   "```\nsome code\n```",
		},
		{
			name: "mixed formatting",
			in:   "**bold** and *italic* and `code`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := markdownToTelegramMarkdownV2(tt.in)
			if got == "" {
				t.Errorf("markdownToTelegramMarkdownV2(%q) returned empty string", tt.in)
			}
			// Ensure conversion doesn't return the exact input (i.e. some transformation occurred)
			// for inputs that contain Markdown syntax.
			if tt.name != "plain text" && got == tt.in {
				t.Errorf("markdownToTelegramMarkdownV2(%q) returned unchanged input, expected MarkdownV2 conversion", tt.in)
			}
		})
	}
}
