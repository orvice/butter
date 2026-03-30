package telegram

import (
	"regexp"
	"strings"
)

// markdownToTelegramHTML converts standard Markdown (as produced by LLMs)
// into Telegram-compatible HTML. It handles fenced code blocks, inline code,
// bold, italic, strikethrough, and links.
//
// Telegram HTML reference: https://core.telegram.org/bots/api#html-style
func markdownToTelegramHTML(text string) string {
	// Split by fenced code blocks first to avoid processing inside them.
	parts := splitCodeBlocks(text)

	var b strings.Builder
	for _, p := range parts {
		if p.isCode {
			// Fenced code block → <pre><code>
			escaped := escapeHTML(p.content)
			if p.lang != "" {
				b.WriteString("<pre><code class=\"language-")
				b.WriteString(escapeHTML(p.lang))
				b.WriteString("\">")
			} else {
				b.WriteString("<pre><code>")
			}
			b.WriteString(escaped)
			b.WriteString("</code></pre>")
		} else {
			b.WriteString(convertInlineMarkdown(p.content))
		}
	}
	return b.String()
}

type codePart struct {
	content string
	lang    string
	isCode  bool
}

var fencedCodeRe = regexp.MustCompile("(?s)```(\\w*)\\n?(.*?)```")

// splitCodeBlocks splits text into code and non-code segments.
func splitCodeBlocks(text string) []codePart {
	var parts []codePart
	lastEnd := 0

	for _, loc := range fencedCodeRe.FindAllStringSubmatchIndex(text, -1) {
		// loc[0]:loc[1] = full match
		// loc[2]:loc[3] = lang
		// loc[4]:loc[5] = code content
		if loc[0] > lastEnd {
			parts = append(parts, codePart{content: text[lastEnd:loc[0]]})
		}
		lang := text[loc[2]:loc[3]]
		code := text[loc[4]:loc[5]]
		// Trim trailing newline inside the fence.
		code = strings.TrimSuffix(code, "\n")
		parts = append(parts, codePart{content: code, lang: lang, isCode: true})
		lastEnd = loc[1]
	}
	if lastEnd < len(text) {
		parts = append(parts, codePart{content: text[lastEnd:]})
	}
	return parts
}

// convertInlineMarkdown converts inline Markdown (outside of fenced code
// blocks) to Telegram HTML.
func convertInlineMarkdown(text string) string {
	// 1. Escape HTML entities first so inserted tags aren't double-escaped.
	text = escapeHTML(text)

	// 2. Inline code: `...` → <code>...</code>
	//    Process before bold/italic so backtick content is protected.
	text = replaceInlineCode(text)

	// 3. Bold: **...** → <b>...</b>
	text = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(text, "<b>$1</b>")

	// 4. Italic: *...* → <i>...</i>  (but not inside bold)
	text = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(text, "<i>$1</i>")

	// 5. Strikethrough: ~~...~~ → <s>...</s>
	text = regexp.MustCompile(`~~(.+?)~~`).ReplaceAllString(text, "<s>$1</s>")

	// 6. Links: [text](url) → <a href="url">text</a>
	text = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(text, `<a href="$2">$1</a>`)

	return text
}

var inlineCodeRe = regexp.MustCompile("`([^`]+)`")

// replaceInlineCode replaces `code` with <code>code</code>.
func replaceInlineCode(text string) string {
	return inlineCodeRe.ReplaceAllString(text, "<code>$1</code>")
}

// escapeHTML escapes characters that are special in Telegram HTML.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
