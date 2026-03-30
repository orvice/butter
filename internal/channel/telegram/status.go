package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"go.orx.me/apps/butter/internal/runner"
)

func (p *Poller) handleStatusCommand(ctx context.Context, b *bot.Bot, msg *models.Message) {
	sessionID := p.deriveSessionID(msg)
	userID := fmt.Sprintf("%d", userIDFromMsg(msg))
	activeAgent := p.getActiveAgent(ctx, sessionID)

	var sb strings.Builder

	// Agent status.
	agentStatus := p.runner.GetAgentStatus(activeAgent)
	if agentStatus != nil {
		sb.WriteString(formatAgentStatus(agentStatus, 0))
	} else {
		sb.WriteString(fmt.Sprintf("Agent: %s (no detail available)\n", activeAgent))
	}

	// Session status.
	sb.WriteString("\n")
	sess, err := p.runner.GetSession(ctx, p.channelName, sessionID, userID)
	if err != nil {
		sb.WriteString(fmt.Sprintf("Session: %s\n  (not found or error: %v)\n", sessionID, err))
	} else {
		sb.WriteString(fmt.Sprintf("Session: %s\n", sessionID))
		sb.WriteString(fmt.Sprintf("  Events: %d\n", sess.Events().Len()))
		lastUpdate := sess.LastUpdateTime()
		if !lastUpdate.IsZero() {
			sb.WriteString(fmt.Sprintf("  Last update: %s (%s ago)\n",
				lastUpdate.Format(time.RFC3339),
				time.Since(lastUpdate).Truncate(time.Second),
			))
		}
	}

	p.sendReply(ctx, b, msg, sb.String())
}

// formatAgentStatus recursively formats an agent status tree with indentation.
func formatAgentStatus(st *runner.AgentStatus, depth int) string {
	indent := strings.Repeat("  ", depth)
	var sb strings.Builder

	if depth == 0 {
		sb.WriteString(fmt.Sprintf("Agent: %s\n", st.Name))
	} else {
		sb.WriteString(fmt.Sprintf("%s- %s\n", indent, st.Name))
	}

	if st.Description != "" && depth == 0 {
		sb.WriteString(fmt.Sprintf("  Description: %s\n", st.Description))
	}

	if len(st.MCPServers) > 0 {
		sb.WriteString(fmt.Sprintf("%s  MCP servers: %s\n", indent, strings.Join(st.MCPServers, ", ")))
	}

	if len(st.SubAgents) > 0 {
		sb.WriteString(fmt.Sprintf("%s  Sub-agents:\n", indent))
		for _, sub := range st.SubAgents {
			sb.WriteString(formatAgentStatus(sub, depth+2))
		}
	}

	return sb.String()
}
