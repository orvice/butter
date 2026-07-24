package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func cmdConfig() Config {
	return Config{
		ChannelName:  "tg",
		WorkspaceID:  "ws1",
		DefaultAgent: "assistant",
		AgentNames:   []string{"assistant", "researcher"},
		ModelNames:   []string{"gpt", "claude"},
	}
}

func cmd(text string) IncomingMessage {
	m := baseMsg()
	m.Text = text
	m.BuildParts = textParts(text)
	return m
}

func TestHandle_AgentList(t *testing.T) {
	h, _, agentSel, _, _, tr := newHarness(cmdConfig())
	_ = agentSel.Set(context.Background(), "tg", "chat:1", "researcher")

	h.Handle(context.Background(), cmd("/agent list"))

	if len(tr.agentLists) != 1 {
		t.Fatalf("expected 1 agent list, got %d", len(tr.agentLists))
	}
	choices := tr.agentLists[0]
	if len(choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(choices))
	}
	var active string
	for _, c := range choices {
		if c.Active {
			active = c.Name
		}
	}
	if active != "researcher" {
		t.Errorf("active agent = %q, want researcher", active)
	}
}

func TestHandle_AgentSwitch_Known(t *testing.T) {
	h, r, agentSel, _, _, tr := newHarness(cmdConfig())
	r.knownAgents["researcher"] = true

	h.Handle(context.Background(), cmd("/agent researcher"))

	if got, _ := agentSel.Get(context.Background(), "tg", "chat:1"); got != "researcher" {
		t.Errorf("selection = %q, want researcher", got)
	}
	if len(tr.replies) != 1 || !strings.Contains(tr.replies[0], "researcher") {
		t.Errorf("replies = %v, want a switch confirmation", tr.replies)
	}
}

func TestHandle_AgentSwitch_Unknown(t *testing.T) {
	h, _, agentSel, _, _, tr := newHarness(cmdConfig())

	h.Handle(context.Background(), cmd("/agent nope"))

	if len(agentSel.setKeys) != 0 {
		t.Errorf("expected no selection set, got %v", agentSel.setKeys)
	}
	if len(tr.replies) != 1 || !strings.Contains(strings.ToLower(tr.replies[0]), "unknown") {
		t.Errorf("replies = %v, want an unknown-agent message", tr.replies)
	}
}

func TestHandle_ModelList(t *testing.T) {
	h, _, _, modelSel, _, tr := newHarness(cmdConfig())
	_ = modelSel.Set(context.Background(), "tg", "chat:1", "claude")

	h.Handle(context.Background(), cmd("/model"))

	if len(tr.modelLists) != 1 {
		t.Fatalf("expected 1 model list, got %d", len(tr.modelLists))
	}
	var active string
	for _, c := range tr.modelLists[0] {
		if c.Active {
			active = c.Alias
		}
	}
	if active != "claude" {
		t.Errorf("active model = %q, want claude", active)
	}
}

func TestHandle_Debug_Toggles(t *testing.T) {
	h, _, _, _, debug, tr := newHarness(cmdConfig())
	debug.toggleState = true

	h.Handle(context.Background(), cmd("/debug"))

	if debug.toggleCalls != 1 {
		t.Errorf("toggle calls = %d, want 1", debug.toggleCalls)
	}
	if len(tr.debugStatus) != 1 || tr.debugStatus[0] != true {
		t.Errorf("debugStatus = %v, want [true]", tr.debugStatus)
	}
}

func TestHandle_Debug_ToggleError(t *testing.T) {
	h, _, _, _, debug, tr := newHarness(cmdConfig())
	debug.toggleErr = errors.New("redis down")

	h.Handle(context.Background(), cmd("/debug"))

	if len(tr.debugStatus) != 0 {
		t.Errorf("expected no debug status on error, got %v", tr.debugStatus)
	}
	if len(tr.replies) != 1 {
		t.Errorf("expected 1 error reply, got %v", tr.replies)
	}
}

func TestHandle_Status(t *testing.T) {
	h, _, _, _, _, tr := newHarness(cmdConfig())

	h.Handle(context.Background(), cmd("/status"))

	if len(tr.statusViews) != 1 {
		t.Fatalf("expected 1 status view, got %d", len(tr.statusViews))
	}
	if tr.statusViews[0].ActiveAgent != "assistant" {
		t.Errorf("status active agent = %q, want assistant", tr.statusViews[0].ActiveAgent)
	}
	if tr.statusViews[0].SessionID != "chat:1" {
		t.Errorf("status session id = %q, want chat:1", tr.statusViews[0].SessionID)
	}
}

func TestHandle_Clear(t *testing.T) {
	h, r, _, _, _, tr := newHarness(cmdConfig())

	h.Handle(context.Background(), cmd("/clear"))

	if r.clearedCalls != 1 {
		t.Errorf("clear calls = %d, want 1", r.clearedCalls)
	}
	if len(tr.replies) != 1 || !strings.Contains(tr.replies[0], "cleared") {
		t.Errorf("replies = %v, want a cleared confirmation", tr.replies)
	}
}

func TestHandle_Command_DoesNotInvokeRunner(t *testing.T) {
	h, r, _, _, _, _ := newHarness(cmdConfig())
	r.knownAgents["researcher"] = true

	h.Handle(context.Background(), cmd("/agent researcher"))

	if len(r.runCalls) != 0 {
		t.Errorf("expected commands not to invoke runner, got %d calls", len(r.runCalls))
	}
}

func TestParseAgentCommand(t *testing.T) {
	tests := []struct {
		text    string
		wantSub string
		wantArg string
	}{
		{"/agent", "list", ""},
		{"/agent list", "list", ""},
		{"/agent foo", "switch", "foo"},
		{"  /agent  bar ", "switch", "bar"},
		{"/model", "", ""},
	}
	for _, tt := range tests {
		sub, arg := parseAgentCommand(tt.text)
		if sub != tt.wantSub || arg != tt.wantArg {
			t.Errorf("parseAgentCommand(%q) = (%q,%q), want (%q,%q)", tt.text, sub, arg, tt.wantSub, tt.wantArg)
		}
	}
}
