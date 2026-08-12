package telegram

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/channel/pipeline"
)

func TestFormatDebugSummary_AlwaysShowsZeroCounts(t *testing.T) {
	got := formatDebugSummary(pipeline.DebugSummary{})
	for _, want := range []string{"Tools:** 0", "Transfers:** 0", "Compactions:** 0"} {
		if !strings.Contains(got, want) {
			t.Errorf("summary missing %q:\n%s", want, got)
		}
	}
}

func TestFormatDebugSummary_SortsAndCollapsesToolNames(t *testing.T) {
	counts := make(map[string]int)
	total := 0
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("tool-%02d", i)
		counts[name] = 10 - i
		total += 10 - i
	}

	got := formatDebugSummary(pipeline.DebugSummary{
		ToolCalls:   total,
		ToolCounts:  counts,
		Transfers:   2,
		Compactions: 1,
	})

	if strings.Index(got, "`tool-00` 10") > strings.Index(got, "`tool-01` 9") {
		t.Errorf("tool details not sorted by count:\n%s", got)
	}
	if strings.Contains(got, "`tool-08`") || strings.Contains(got, "`tool-09`") {
		t.Errorf("summary displayed more than eight tool names:\n%s", got)
	}
	if !strings.Contains(got, "Other 2 tools: 3") {
		t.Errorf("summary missing collapsed tools:\n%s", got)
	}
}

func TestFormatProcessingAndFinalMessage_LatestOnlyWhileProcessing(t *testing.T) {
	evt := session.NewEvent(t.Context(), "inv-1")
	evt.Content = &genai.Content{Parts: []*genai.Part{
		{FunctionCall: &genai.FunctionCall{Name: "search", Args: map[string]any{"query": "butter"}}},
	}}
	summary := pipeline.DebugSummary{
		ToolCalls:   1,
		ToolCounts:  map[string]int{"search": 1},
		LatestEvent: evt,
	}

	processing := formatProcessingMessage("assistant", "12:34:56", "chat:1", &summary)
	if !strings.Contains(processing, "Latest") || !strings.Contains(processing, "butter") {
		t.Errorf("processing message missing latest debug:\n%s", processing)
	}

	final := formatFinalMessage("assistant", "Done", "12:34:57", "chat:1", &summary)
	if strings.Contains(final, "Latest") || strings.Contains(final, "butter") {
		t.Errorf("final message retained latest debug:\n%s", final)
	}
	for _, want := range []string{"Done", "Tools:** 1", "`search` 1"} {
		if !strings.Contains(final, want) {
			t.Errorf("final message missing %q:\n%s", want, final)
		}
	}
	if strings.Count(final, "─────────") != 1 {
		t.Errorf("final message should use one footer separator:\n%s", final)
	}
}

func TestFormatProcessingMessage_DebugOffKeepsLegacyLayout(t *testing.T) {
	got := formatProcessingMessage("assistant", "12:34:56", "chat:1", nil)
	if !strings.Contains(got, "🕐 `12:34:56`\n💬 `chat:1`") {
		t.Errorf("debug-off processing layout changed:\n%s", got)
	}
}
