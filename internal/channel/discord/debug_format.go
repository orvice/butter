package discord

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/adk/session"
)

const maxArgLen = 200

// FormatDebugEvent converts an ADK event into a human-readable debug string.
func FormatDebugEvent(evt *session.Event) string {
	var parts []string

	if evt.Actions.TransferToAgent != "" {
		from := evt.Author
		if from == "" {
			from = "unknown"
		}
		parts = append(parts, fmt.Sprintf("[DEBUG] Transfer: %s -> %s", from, evt.Actions.TransferToAgent))
	}

	if evt.Content != nil {
		for _, part := range evt.Content.Parts {
			if part.FunctionCall != nil {
				fc := part.FunctionCall
				args := formatArgs(fc.Args)
				parts = append(parts, fmt.Sprintf("[DEBUG] Tool: %s(%s)", fc.Name, args))
			}
		}
	}

	return strings.Join(parts, "\n")
}

func formatArgs(args map[string]any) string {
	if len(args) == 0 {
		return ""
	}
	b, err := json.Marshal(args)
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
