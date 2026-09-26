package mem0memory

import (
	"strings"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/mem0"
)

// maxMessageRunes caps one captured message. Extraction is an LLM call on
// the mem0 side; an oversized paste would cost far more than the facts it
// yields.
const maxMessageRunes = 16000

// imagePlaceholder stands in for inline media, which mem0 cannot extract.
const imagePlaceholder = "[image]"

// TurnMessages returns what Memory Capture sends for one invocation: the
// user's text and the agents' final response text, in order. Tool calls,
// tool results, thoughts, and partial (streaming) events are excluded, and
// inline media becomes a placeholder.
func TurnMessages(sess session.Session, invocationID string) []mem0.Message {
	if sess == nil || invocationID == "" {
		return nil
	}
	var out []mem0.Message
	for ev := range sess.Events().All() {
		if ev == nil || ev.InvocationID != invocationID || ev.Partial {
			continue
		}
		role := "assistant"
		if ev.Author == "user" {
			role = "user"
		} else if !ev.IsFinalResponse() {
			continue
		}
		text := eventText(ev.Content, role == "user")
		if text == "" {
			continue
		}
		out = append(out, mem0.Message{Role: role, Content: truncateRunes(text, maxMessageRunes)})
	}
	return out
}

// eventText joins an event's non-thought text parts. Inline media counts
// only for user input: a model never "said" an image it returned.
func eventText(content *genai.Content, withMedia bool) string {
	if content == nil {
		return ""
	}
	var parts []string
	for _, p := range content.Parts {
		switch {
		case p == nil || p.Thought:
		case p.Text != "":
			if t := strings.TrimSpace(p.Text); t != "" {
				parts = append(parts, t)
			}
		case withMedia && (p.InlineData != nil || p.FileData != nil):
			parts = append(parts, imagePlaceholder)
		}
	}
	return strings.Join(parts, "\n")
}

func truncateRunes(s string, limit int) string {
	if limit <= 0 {
		return s
	}
	runes := 0
	for i := range s {
		if runes == limit {
			return s[:i] + "…"
		}
		runes++
	}
	return s
}
