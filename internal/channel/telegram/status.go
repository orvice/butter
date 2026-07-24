package telegram

import (
	"fmt"
	"strings"
	"time"

	"go.orx.me/apps/butter/internal/runtime/runner"
)

type sessionStatus struct {
	eventCount int
	lastUpdate time.Time
}

func formatStatusMessage(agentStatus *runner.AgentStatus, activeAgent, modelText, sessionID string, sess *sessionStatus, sessErr error, now time.Time) string {
	var sb strings.Builder
	sb.WriteString("**Status**\n\n")
	sb.WriteString(formatAgentSection(agentStatus, activeAgent))
	sb.WriteString("\n")
	sb.WriteString(formatModelSection(modelText))
	sb.WriteString("\n")
	sb.WriteString(formatSessionSection(sessionID, sess, sessErr, now))
	return sb.String()
}

func formatAgentSection(st *runner.AgentStatus, activeAgent string) string {
	var sb strings.Builder
	sb.WriteString("🤖 **Agent**\n")
	if st == nil {
		sb.WriteString(fmt.Sprintf("- Name: `%s`\n", activeAgent))
		sb.WriteString("- ℹ️ Detail: unavailable\n")
		return sb.String()
	}

	sb.WriteString(formatAgentStatus(st, 0))
	return sb.String()
}

func formatModelSection(modelText string) string {
	var sb strings.Builder
	sb.WriteString("🧠 **Model**\n")
	if modelText == "" {
		sb.WriteString("- ⚪ Active: unavailable\n")
		return sb.String()
	}
	sb.WriteString(fmt.Sprintf("- 🧠 Active: %s\n", modelText))
	return sb.String()
}

func formatSessionSection(sessionID string, sess *sessionStatus, sessErr error, now time.Time) string {
	var sb strings.Builder
	sb.WriteString("💬 **Session**\n")
	sb.WriteString(fmt.Sprintf("- 🆔 ID: `%s`\n", sessionID))
	if sessErr != nil {
		sb.WriteString(fmt.Sprintf("- ⚠️ Warning: %v\n", sessErr))
		return sb.String()
	}
	if sess == nil {
		sb.WriteString("- ℹ️ Detail: unavailable\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("- 📝 Events: `%d`\n", sess.eventCount))
	if !sess.lastUpdate.IsZero() {
		sb.WriteString(fmt.Sprintf("- 🕐 Last update: `%s`\n", sess.lastUpdate.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- ⏳ Age: `%s`\n", now.Sub(sess.lastUpdate).Truncate(time.Second)))
	}
	return sb.String()
}

// formatAgentStatus recursively formats an agent status tree with indentation.
func formatAgentStatus(st *runner.AgentStatus, depth int) string {
	indent := strings.Repeat("  ", depth)
	var sb strings.Builder

	if depth == 0 {
		sb.WriteString(fmt.Sprintf("- Name: `%s`\n", st.Name))
	} else {
		sb.WriteString(fmt.Sprintf("%s- 🔹 `%s`\n", indent, st.Name))
	}

	if st.Description != "" && depth == 0 {
		sb.WriteString(fmt.Sprintf("- 📄 Description: %s\n", st.Description))
	}

	if len(st.MCPServers) > 0 {
		sb.WriteString(fmt.Sprintf("%s- 🔌 MCP servers: `%s`\n", indent, strings.Join(st.MCPServers, "`, `")))
	}

	if len(st.SubAgents) > 0 {
		sb.WriteString(fmt.Sprintf("%s- 👥 Sub-agents:\n", indent))
		for _, sub := range st.SubAgents {
			sb.WriteString(formatAgentStatus(sub, depth+1))
		}
	}

	return sb.String()
}
