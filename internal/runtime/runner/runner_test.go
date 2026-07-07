package runner

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"google.golang.org/adk/v2/agent"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestDeriveSessionID(t *testing.T) {
	tests := []struct {
		name   string
		scope  agentsv1.AgentSessionScope
		chatID int64
		userID int64
		want   string
	}{
		{
			name:   "user scope",
			scope:  agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_USER,
			chatID: 100, userID: 42,
			want: "user:42",
		},
		{
			name:   "chat scope",
			scope:  agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_CHAT,
			chatID: 100, userID: 42,
			want: "chat:100",
		},
		{
			name:   "unspecified defaults to chat",
			scope:  agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_UNSPECIFIED,
			chatID: 100, userID: 42,
			want: "chat:100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveSessionID(tt.scope, tt.chatID, tt.userID)
			if got != tt.want {
				t.Errorf("DeriveSessionID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateUTF8Boundary(t *testing.T) {
	tests := []struct {
		name string
		in   string
		max  int
	}{
		{name: "ascii", in: strings.Repeat("a", 10), max: 4},
		{name: "cjk cut mid-rune", in: strings.Repeat("中", 10), max: 4},
		{name: "emoji cut mid-rune", in: strings.Repeat("🙂", 10), max: 5},
		{name: "no truncation needed", in: "中文", max: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.in, tt.max)
			if !utf8.ValidString(got) {
				t.Fatalf("truncate(%q, %d) = %q is not valid UTF-8", tt.in, tt.max, got)
			}
			if len(tt.in) <= tt.max && got != tt.in {
				t.Fatalf("truncate(%q, %d) = %q, want unchanged input", tt.in, tt.max, got)
			}
		})
	}
}

func TestCancelInvocationWorkspaceScope(t *testing.T) {
	s := &Service{}
	cancelled := false
	s.registerCancel("inv-1", "ws-a", func() { cancelled = true })

	if s.CancelInvocation("inv-1", "ws-b") {
		t.Fatal("cross-workspace cancel should be rejected")
	}
	if cancelled {
		t.Fatal("cancel func must not run for a rejected request")
	}
	if !s.CancelInvocation("inv-1", "ws-a") {
		t.Fatal("same-workspace cancel should succeed")
	}
	if !cancelled {
		t.Fatal("cancel func should have run")
	}

	systemCancelled := false
	s.registerCancel("inv-2", "ws-a", func() { systemCancelled = true })
	if !s.CancelInvocation("inv-2", "") {
		t.Fatal("system path (empty workspace) should cancel any invocation")
	}
	if !systemCancelled {
		t.Fatal("cancel func should have run for system path")
	}

	if s.CancelInvocation("missing", "") {
		t.Fatal("unknown invocation id should return false")
	}
}

func TestNewServiceRejectsCrossWorkspaceDuplicateNames(t *testing.T) {
	providers := []agentsv1.ModelProvider{{
		Name:   "p",
		Type:   "openai",
		Models: []*agentsv1.ModelConfig{{Name: "m1"}},
	}}
	agents := []agentsv1.Agent{
		{Name: "dup", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "m1"}},
		{Name: "dup", WorkspaceId: "ws-b", Config: &agentsv1.AgentConfig{Model: "m1"}},
	}
	_, err := NewService(context.Background(), agents, providers, nil, nil, nil, nil, nil, nil, adkrunner.PluginConfig{})
	if err == nil || !strings.Contains(err.Error(), "unique across workspaces") {
		t.Fatalf("expected cross-workspace duplicate name error, got %v", err)
	}
}

func TestReloadProtoAgentsSkipsReservedBuilderNames(t *testing.T) {
	providers := []agentsv1.ModelProvider{{
		Name:   "p",
		Type:   "openai",
		Models: []*agentsv1.ModelConfig{{Name: "m1"}},
	}}
	svc, err := NewService(context.Background(), nil, providers, nil, nil, nil, nil, nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.RegisterAgentWithBuilder("system", nil, func(context.Context, string) (agent.Agent, error) {
		return nil, nil
	})
	if !svc.IsReservedAgentName("system") {
		t.Fatal("system should be a reserved builder name")
	}

	// A proto agent that collides with the builder name must be skipped so the
	// builder stays authoritative and no stale proto entry is registered.
	reload := []agentsv1.Agent{
		{Name: "system", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "m1"}},
	}
	if err := svc.ReloadProtoAgents(context.Background(), reload, providers, nil, nil); err != nil {
		t.Fatalf("ReloadProtoAgents: %v", err)
	}

	svc.mu.Lock()
	_, hasProto := svc.agentsProto["system"]
	_, hasAgent := svc.agents["system"]
	_, hasBuilder := svc.agentBuilders["system"]
	svc.mu.Unlock()

	if hasProto {
		t.Fatal("reserved name must not be registered as a proto agent")
	}
	if !hasAgent {
		t.Fatal("builder agent must remain registered after reload")
	}
	if !hasBuilder {
		t.Fatal("builder func must remain registered after reload")
	}
}

func TestSummarizeEvent(t *testing.T) {
	evt := session.NewEvent(t.Context(), "inv-1")
	evt.Content = &genai.Content{Parts: []*genai.Part{
		{Text: "hello"},
		{FunctionCall: &genai.FunctionCall{Name: "tool_a"}},
		{FunctionResponse: &genai.FunctionResponse{Name: "tool_a"}},
		{CodeExecutionResult: &genai.CodeExecutionResult{Outcome: genai.OutcomeOK}},
	}}
	evt.Actions.StateDelta["foo"] = "bar"
	evt.Actions.ArtifactDelta["report.txt"] = 1

	summary := summarizeEvent(evt)

	if summary.textParts != 1 {
		t.Fatalf("textParts = %d, want 1", summary.textParts)
	}
	if summary.functionCalls != 1 {
		t.Fatalf("functionCalls = %d, want 1", summary.functionCalls)
	}
	if summary.functionResponses != 1 {
		t.Fatalf("functionResponses = %d, want 1", summary.functionResponses)
	}
	if summary.codeExecutionResults != 1 {
		t.Fatalf("codeExecutionResults = %d, want 1", summary.codeExecutionResults)
	}
	if summary.stateDeltaKeys != 1 {
		t.Fatalf("stateDeltaKeys = %d, want 1", summary.stateDeltaKeys)
	}
	if summary.artifactDeltaKeys != 1 {
		t.Fatalf("artifactDeltaKeys = %d, want 1", summary.artifactDeltaKeys)
	}
}
