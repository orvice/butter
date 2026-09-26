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
	return TurnInput{Session: sess}.messages(invocationID)
}

// TurnInput is one finished turn as the runner saw it.
type TurnInput struct {
	Session session.Session
	// FromEvent is the index of the first session event the turn appended.
	// A workflow resume reuses the paused invocation's ID, so the
	// invocation alone would resend the exchange from before the pause.
	FromEvent int
	// UserText is the user's input as sent, with media as a placeholder.
	// When set it replaces the session's user events: a resume rewraps the
	// reply as a FunctionResponse, which carries no text.
	UserText string
}

func (in TurnInput) messages(invocationID string) []mem0.Message {
	if in.Session == nil || invocationID == "" {
		return nil
	}
	var out []mem0.Message
	if text := strings.TrimSpace(in.UserText); text != "" {
		out = append(out, mem0.Message{Role: "user", Content: truncateRunes(text, maxMessageRunes)})
	}
	events := in.Session.Events()
	for i := max(in.FromEvent, 0); i < events.Len(); i++ {
		ev := events.At(i)
		if ev == nil || ev.InvocationID != invocationID || ev.Partial {
			continue
		}
		role := "assistant"
		if ev.Author == "user" {
			if in.UserText != "" {
				continue
			}
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

// PartsText renders user input parts the way capture sends them: text
// joined by newlines, inline media as a placeholder.
func PartsText(parts []*genai.Part) string {
	return eventText(&genai.Content{Parts: parts}, true)
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
