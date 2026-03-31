package telegram

import (
	"strings"
	"testing"

	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func TestFormatDebugEvent_ToolCall(t *testing.T) {
	evt := session.NewEvent("inv-1")
	evt.Content = &genai.Content{
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{
				Name: "search",
				Args: map[string]any{"query": "hello world"},
			}},
		},
	}

	got := FormatDebugEvent(evt)
	if !strings.Contains(got, "🔧") {
		t.Errorf("expected tool emoji, got: %s", got)
	}
	if !strings.Contains(got, "`search`") {
		t.Errorf("expected tool name in backticks, got: %s", got)
	}
	if !strings.Contains(got, "hello world") {
		t.Errorf("expected args in output, got: %s", got)
	}
}

func TestFormatDebugEvent_ToolCallNoArgs(t *testing.T) {
	evt := session.NewEvent("inv-1")
	evt.Content = &genai.Content{
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{
				Name: "ping",
				Args: nil,
			}},
		},
	}

	got := FormatDebugEvent(evt)
	if !strings.Contains(got, "`ping()`") {
		t.Errorf("expected tool name with empty parens, got: %s", got)
	}
}

func TestFormatDebugEvent_Transfer(t *testing.T) {
	evt := session.NewEvent("inv-1")
	evt.Author = "router"
	evt.Actions.TransferToAgent = "specialist"

	got := FormatDebugEvent(evt)
	if !strings.Contains(got, "🔀") {
		t.Errorf("expected transfer emoji, got: %s", got)
	}
	if !strings.Contains(got, "`router`") {
		t.Errorf("expected source agent in backticks, got: %s", got)
	}
	if !strings.Contains(got, "`specialist`") {
		t.Errorf("expected target agent in backticks, got: %s", got)
	}
	if !strings.Contains(got, "➡️") {
		t.Errorf("expected arrow emoji, got: %s", got)
	}
}

func TestFormatDebugEvent_Empty(t *testing.T) {
	evt := session.NewEvent("inv-1")
	got := FormatDebugEvent(evt)
	if got != "" {
		t.Errorf("expected empty string for event with no debug content, got: %s", got)
	}
}

func TestFormatDebugEvent_TransferAndToolCall(t *testing.T) {
	evt := session.NewEvent("inv-1")
	evt.Author = "agent-a"
	evt.Actions.TransferToAgent = "agent-b"
	evt.Content = &genai.Content{
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{
				Name: "lookup",
				Args: map[string]any{"id": 42},
			}},
		},
	}

	got := FormatDebugEvent(evt)
	if !strings.Contains(got, "`agent-a`") {
		t.Errorf("expected source agent, got: %s", got)
	}
	if !strings.Contains(got, "`agent-b`") {
		t.Errorf("expected target agent, got: %s", got)
	}
	if !strings.Contains(got, "`lookup`") {
		t.Errorf("expected tool call, got: %s", got)
	}
}

func TestFormatCompactionEvent(t *testing.T) {
	got := FormatCompactionEvent("my-agent")
	if !strings.Contains(got, "📦") {
		t.Errorf("expected compaction emoji, got: %s", got)
	}
	if !strings.Contains(got, "`my-agent`") {
		t.Errorf("expected agent name in backticks, got: %s", got)
	}
}

func TestTruncate(t *testing.T) {
	short := "hello"
	if got := truncate(short, 200); got != short {
		t.Errorf("short string should not be truncated, got: %s", got)
	}

	long := strings.Repeat("a", 300)
	got := truncate(long, 200)
	if len(got) != 203 { // 200 + "..."
		t.Errorf("expected truncated length 203, got: %d", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ... suffix, got: %s", got[len(got)-5:])
	}
}

func TestFormatArgs_Empty(t *testing.T) {
	got := formatArgs(nil)
	if got != "" {
		t.Errorf("expected empty for nil args, got: %s", got)
	}
}
