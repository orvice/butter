package telegram

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/adk/session"
)

const maxArgLen = 200

// FormatDebugEvent converts an ADK event into a Markdown-formatted debug string
// with emojis for readability in Telegram.
// Returns empty string if the event has no debug-relevant content.
func FormatDebugEvent(evt *session.Event) string {
	var parts []string

	// Check for agent transfer.
	if evt.Actions.TransferToAgent != "" {
		from := evt.Author
		if from == "" {
			from = "unknown"
		}
		parts = append(parts, fmt.Sprintf("🔀 *Transfer*: `%s` ➡️ `%s`", from, evt.Actions.TransferToAgent))
	}

	// Check for function calls in content parts.
	if evt.Content != nil {
		for _, part := range evt.Content.Parts {
			if part.FunctionCall != nil {
				fc := part.FunctionCall
				args := formatArgs(fc.Args)
				if args != "" {
					parts = append(parts, fmt.Sprintf("🔧 *Tool*: `%s`\n```json\n%s\n```", fc.Name, args))
				} else {
					parts = append(parts, fmt.Sprintf("🔧 *Tool*: `%s()`", fc.Name))
				}
			}
		}
	}

	return strings.Join(parts, "\n")
}

// FormatCompactionEvent returns a Markdown-formatted debug string for context
// compaction events.
func FormatCompactionEvent(agentName string) string {
	return fmt.Sprintf("📦 *Context compacted* — agent: `%s`", agentName)
}

func formatArgs(args map[string]any) string {
	if len(args) == 0 {
		return ""
	}
	b, err := json.MarshalIndent(args, "", "  ")
	if err != nil {
		return "..."
	}
	s := string(b)
	return truncate(s, maxArgLen)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
