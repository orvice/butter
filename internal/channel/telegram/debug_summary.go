package telegram

import (
	"fmt"
	"sort"
	"strings"

	"go.orx.me/apps/butter/internal/channel/pipeline"
)

const maxToolSummaryNames = 8

type toolCount struct {
	name  string
	count int
}

func formatDebugSummary(summary pipeline.DebugSummary) string {
	tools := make([]toolCount, 0, len(summary.ToolCounts))
	for name, count := range summary.ToolCounts {
		tools = append(tools, toolCount{name: name, count: count})
	}
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].count != tools[j].count {
			return tools[i].count > tools[j].count
		}
		return tools[i].name < tools[j].name
	})

	lines := []string{fmt.Sprintf("🔧 **Tools:** %d", summary.ToolCalls)}
	visible := min(len(tools), maxToolSummaryNames)
	if visible > 0 {
		details := make([]string, 0, visible)
		for _, tool := range tools[:visible] {
			details = append(details, fmt.Sprintf("`%s` %d", tool.name, tool.count))
		}
		lines = append(lines, strings.Join(details, " · "))
	}
	if len(tools) > visible {
		otherCalls := 0
		for _, tool := range tools[visible:] {
			otherCalls += tool.count
		}
		lines = append(lines, fmt.Sprintf("Other %d tools: %d", len(tools)-visible, otherCalls))
	}
	lines = append(lines, fmt.Sprintf("🔀 **Transfers:** %d · 📦 **Compactions:** %d", summary.Transfers, summary.Compactions))
	return strings.Join(lines, "\n")
}

func formatLatestDebug(summary pipeline.DebugSummary) string {
	switch {
	case summary.LatestEvent != nil:
		return FormatDebugEvent(summary.LatestEvent)
	case summary.LatestCompaction != "":
		return FormatCompactionEvent(summary.LatestCompaction)
	default:
		return ""
	}
}
