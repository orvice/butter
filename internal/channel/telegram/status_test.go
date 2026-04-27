package telegram

import (
	"errors"
	"strings"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/runtime/runner"
)

func TestFormatStatusMessage(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	msg := formatStatusMessage(
		&runner.AgentStatus{
			Name:        "planner",
			Description: "Main coordinator",
			MCPServers:  []string{"mcp-alpha", "mcp-beta"},
			SubAgents: []*runner.AgentStatus{{
				Name:       "worker",
				MCPServers: []string{"mcp-gamma"},
			}},
		},
		"planner",
		"`fast` -> `gemini-2.5-pro`",
		"telegram-chat-1",
		&sessionStatus{
			eventCount: 42,
			lastUpdate: now.Add(-5 * time.Minute),
		},
		nil,
		now,
	)

	wants := []string{
		"**Status**",
		"🤖 **Agent**",
		"- Name: `planner`",
		"- Description: Main coordinator",
		"- MCP servers: `mcp-alpha`, `mcp-beta`",
		"- Sub-agents:",
		"  - `worker`",
		"  - MCP servers: `mcp-gamma`",
		"🧠 **Model**",
		"- Active: `fast` -> `gemini-2.5-pro`",
		"💬 **Session**",
		"- ID: `telegram-chat-1`",
		"- Events: `42`",
		"- Last update: `2026-04-28T11:55:00Z`",
		"- Age: `5m0s`",
	}
	for _, want := range wants {
		if !strings.Contains(msg, want) {
			t.Fatalf("formatted status missing %q in:\n%s", want, msg)
		}
	}
}

func TestFormatStatusMessageSessionError(t *testing.T) {
	msg := formatStatusMessage(nil, "fallback-agent", "", "session-2", nil, errors.New("session not found"), time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC))

	wants := []string{
		"- Name: `fallback-agent`",
		"- Detail: unavailable",
		"- Active: unavailable",
		"- ID: `session-2`",
		"- Warning: session not found",
	}
	for _, want := range wants {
		if !strings.Contains(msg, want) {
			t.Fatalf("formatted status missing %q in:\n%s", want, msg)
		}
	}
}
