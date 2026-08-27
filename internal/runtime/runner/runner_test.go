package runner

import (
	"context"
	"iter"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
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

func TestReloadProtoAgentsSkipsNonRunnableLifecycle(t *testing.T) {
	providers := []agentsv1.ModelProvider{{
		Name:   "p",
		Type:   "openai",
		Models: []*agentsv1.ModelConfig{{Name: "m1"}},
	}}
	svc, err := NewService(context.Background(), nil, providers, nil, nil, nil, nil, nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	llm := func(name string, status agentsv1.AgentLifecycleStatus) agentsv1.Agent {
		return agentsv1.Agent{
			Name:            name,
			AgentId:         name,
			WorkspaceId:     "ws-a",
			Type:            agentsv1.AgentType_AGENT_TYPE_LLM,
			LifecycleStatus: status,
			Config:          &agentsv1.AgentConfig{Model: "m1", Instruction: "do things"},
		}
	}
	reload := []agentsv1.Agent{
		llm("active", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE),
		llm("legacy", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_UNSPECIFIED),
		llm("provisioning", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_PROVISIONING),
		llm("errored", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ERROR),
		llm("deleting", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETING),
		llm("deleted", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED),
		llm("migrating", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED),
	}
	if err := svc.ReloadProtoAgents(context.Background(), reload, providers, nil, nil); err != nil {
		t.Fatalf("ReloadProtoAgents: %v", err)
	}

	registered := func(name string) bool {
		svc.mu.Lock()
		defer svc.mu.Unlock()
		_, ok := svc.agents[name]
		return ok
	}
	for _, name := range []string{"active", "legacy"} {
		if !registered(name) {
			t.Fatalf("expected runnable agent %q to be registered", name)
		}
	}
	for _, name := range []string{"provisioning", "errored", "deleting", "deleted", "migrating"} {
		if registered(name) {
			t.Fatalf("non-runnable agent %q must not be registered", name)
		}
	}
}

func TestRunDiscardsOverriddenAgentBuiltAcrossReload(t *testing.T) {
	svc, err := NewService(context.Background(), nil, nil, nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	base, err := newGenerationTestAgent("base")
	if err != nil {
		t.Fatalf("new base agent: %v", err)
	}
	firstBuildStarted := make(chan struct{})
	releaseFirstBuild := make(chan struct{})
	var builds atomic.Int32
	svc.RegisterAgentWithBuilder("dynamic-agent", base, func(_ context.Context, _ string) (agent.Agent, error) {
		build := builds.Add(1)
		if build == 1 {
			close(firstBuildStarted)
			<-releaseFirstBuild
		}
		return newGenerationTestAgent("build-" + string(rune('0'+build)))
	})

	type runResult struct {
		output string
		err    error
	}
	result := make(chan runResult, 1)
	go func() {
		ctxInfo := &agentsv1.ContextInfo{Uuid: "generation-1", SessionId: "generation-1", UserId: "u1", ChannelName: "test-app"}
		output, runErr := svc.Run(context.Background(), "dynamic-agent", []*genai.Part{{Text: "hello"}}, "override", ctxInfo, nil, nil)
		result <- runResult{output: output, err: runErr}
	}()

	select {
	case <-firstBuildStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("first overridden-Agent build did not start")
	}
	if err := svc.ReloadProtoAgents(context.Background(), nil, nil, nil, nil); err != nil {
		t.Fatalf("ReloadProtoAgents: %v", err)
	}
	close(releaseFirstBuild)

	var first runResult
	select {
	case first = <-result:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not finish after reload")
	}
	if first.err != nil {
		t.Fatalf("Run: %v", first.err)
	}
	if first.output != "build-2" {
		t.Fatalf("overlapping Run output = %q, want post-reload build-2", first.output)
	}

	ctxInfo := &agentsv1.ContextInfo{Uuid: "generation-2", SessionId: "generation-2", UserId: "u1", ChannelName: "test-app"}
	second, err := svc.Run(context.Background(), "dynamic-agent", []*genai.Part{{Text: "hello again"}}, "override", ctxInfo, nil, nil)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if second != "build-2" {
		t.Fatalf("cached post-reload Run output = %q, want build-2", second)
	}
	if got := builds.Load(); got != 2 {
		t.Fatalf("overridden-Agent builds = %d, want stale build discarded and one post-reload build cached", got)
	}
}

type generationTestModel struct {
	reply string
}

func (m *generationTestModel) Name() string { return "generation-model" }

func (m *generationTestModel) GenerateContent(_ context.Context, _ *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(&model.LLMResponse{
			Content:      genai.NewContentFromText(m.reply, genai.RoleModel),
			FinishReason: genai.FinishReasonStop,
		}, nil)
	}
}

func newGenerationTestAgent(reply string) (agent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:  "dynamic-agent",
		Model: &generationTestModel{reply: reply},
	})
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
