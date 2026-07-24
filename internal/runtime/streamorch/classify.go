package streamorch

import "google.golang.org/adk/v2/session"

// hasOnlyTextParts reports whether every part of evt's content is plain text
// (or a thought), with no function call/response, code execution, or binary
// data. Such events are fully represented by their TextDelta frames alone.
func hasOnlyTextParts(evt *session.Event) bool {
	if evt == nil || evt.Content == nil || len(evt.Content.Parts) == 0 {
		return false
	}
	for _, part := range evt.Content.Parts {
		if part == nil {
			continue
		}
		if part.Text == "" && !part.Thought {
			return false
		}
		if part.FunctionCall != nil || part.FunctionResponse != nil ||
			part.CodeExecutionResult != nil || part.ExecutableCode != nil ||
			part.InlineData != nil || part.FileData != nil {
			return false
		}
	}
	return true
}

// textParts returns the non-thought text chunks of a partial event, in
// order. Only partial events stream text deltas; the final event's text is
// carried by the run's response instead.
func textParts(evt *session.Event) []string {
	if evt == nil || !evt.Partial || evt.Content == nil {
		return nil
	}
	out := make([]string, 0, len(evt.Content.Parts))
	for _, part := range evt.Content.Parts {
		if part == nil || part.Text == "" || part.Thought {
			continue
		}
		out = append(out, part.Text)
	}
	return out
}
