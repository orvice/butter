package linear

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"go.orx.me/apps/butter/internal/redact"
)

const (
	// maxResponseRunes bounds a posted response or error body. It matches
	// butter-box#20 until Linear's real limit is confirmed.
	maxResponseRunes = 8000
	// maxParameterRunes bounds an action's one-line parameter.
	maxParameterRunes = 200
	// maxErrorRunes bounds an error's detail in a posted failure.
	maxErrorRunes = 500
)

// turnInput is the text a turn sends the Agent. The first turn of a session
// — or any turn that finds no history, after a clear or an Agent change —
// carries Linear's context; later turns carry only the messages, joined in
// the order they arrived.
func turnInput(event *Event, messages []string, hasHistory bool) string {
	var b strings.Builder
	if !hasHistory {
		if pc := strings.TrimSpace(event.PromptContext); pc != "" {
			b.WriteString("Linear context:\n")
			b.WriteString(pc)
		} else if label := event.Issue.Label(); label != "" {
			fmt.Fprintf(&b, "Linear issue: %s", label)
			if event.Issue.URL != "" {
				fmt.Fprintf(&b, " (%s)", event.Issue.URL)
			}
		}
	}
	var joined []string
	for _, msg := range messages {
		if msg = strings.TrimSpace(msg); msg != "" {
			joined = append(joined, msg)
		}
	}
	if len(joined) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n\nMessage from the Linear user:\n")
		}
		b.WriteString(strings.Join(joined, "\n\n"))
	}
	if b.Len() == 0 {
		return "Continue the Linear session and report your status."
	}
	return b.String()
}

// acknowledgement is the first thought of a new session.
func acknowledgement(issue Issue, agentName string) string {
	if label := issue.Label(); label != "" {
		return fmt.Sprintf("Picked up %s — working on it with %s.", label, agentName)
	}
	return fmt.Sprintf("Picked this up — working on it with %s.", agentName)
}

// truncateResponse bounds a response, pointing at the full reply in Butter
// when it had to cut.
func truncateResponse(text, fullReplyURL string) string {
	text = redact.Text(text)
	if utf8.RuneCountInString(text) <= maxResponseRunes {
		return text
	}
	note := "\n\n_…truncated._"
	if fullReplyURL != "" {
		note = "\n\n_…truncated. The full reply is in Butter: " + fullReplyURL + "_"
	}
	return truncateRunes(text, maxResponseRunes-utf8.RuneCountInString(note)) + note
}

// boundBody redacts and bounds any activity body.
func boundBody(body string) string {
	if body == "" {
		return ""
	}
	return truncateRunes(redact.Text(body), maxResponseRunes)
}

// boundParameter redacts an action parameter and keeps it to one line.
func boundParameter(parameter string) string {
	if parameter == "" {
		return ""
	}
	return truncateRunes(strings.Join(strings.Fields(redact.Text(parameter)), " "), maxParameterRunes)
}

func sanitizeError(err error) string {
	return truncateRunes(redact.Text(err.Error()), maxErrorRunes)
}

// truncateRunes cuts s to at most limit runes, marking the cut.
func truncateRunes(s string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit-1]) + "…"
}

// formatElapsed renders a duration for people: "45s", "3m 12s", "1h 5m".
func formatElapsed(d time.Duration) string {
	d = d.Round(time.Second)
	h, m, s := int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60
	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case m > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	default:
		return fmt.Sprintf("%ds", s)
	}
}
