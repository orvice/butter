package telegram

import "testing"

func TestMarkdownToTelegramHTML(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain text",
			in:   "Hello world",
			want: "Hello world",
		},
		{
			name: "bold",
			in:   "This is **bold** text",
			want: "This is <b>bold</b> text",
		},
		{
			name: "italic",
			in:   "This is *italic* text",
			want: "This is <i>italic</i> text",
		},
		{
			name: "inline code",
			in:   "Use `fmt.Println` here",
			want: "Use <code>fmt.Println</code> here",
		},
		{
			name: "strikethrough",
			in:   "This is ~~deleted~~ text",
			want: "This is <s>deleted</s> text",
		},
		{
			name: "link",
			in:   "Visit [Google](https://google.com) now",
			want: `Visit <a href="https://google.com">Google</a> now`,
		},
		{
			name: "fenced code block",
			in:   "Example:\n```go\nfmt.Println(\"hello\")\n```\nDone.",
			want: "Example:\n<pre><code class=\"language-go\">fmt.Println(\"hello\")</code></pre>\nDone.",
		},
		{
			name: "fenced code block no lang",
			in:   "```\nsome code\n```",
			want: "<pre><code>some code</code></pre>",
		},
		{
			name: "html escaping",
			in:   "Use <div> & \"quotes\"",
			want: "Use &lt;div&gt; &amp; \"quotes\"",
		},
		{
			name: "mixed formatting",
			in:   "**bold** and *italic* and `code`",
			want: "<b>bold</b> and <i>italic</i> and <code>code</code>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := markdownToTelegramHTML(tt.in)
			if got != tt.want {
				t.Errorf("markdownToTelegramHTML(%q)\n got: %q\nwant: %q", tt.in, got, tt.want)
			}
		})
	}
}
