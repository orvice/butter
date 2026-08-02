package application

import (
	"context"
	"errors"
	"iter"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// --- Fake TitleModelResolver ---

type fakeTitleResolver struct {
	agents map[string]fakeAgentInfo
}

type fakeAgentInfo struct {
	model       string
	workspaceID string
	agentType   agentsv1.AgentType
}

func (f *fakeTitleResolver) GetAgentMeta(name string) (string, string, agentsv1.AgentType, bool) {
	a, ok := f.agents[name]
	if !ok {
		return "", "", agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED, false
	}
	return a.model, a.workspaceID, a.agentType, true
}

// --- Fake WorkspaceModelProviderLister ---

type fakeProviderLister struct {
	providers map[string][]*agentsv1.ModelProvider
	err       error
}

func (f *fakeProviderLister) ListModelProviders(_ context.Context, workspaceID string) ([]*agentsv1.ModelProvider, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.providers[workspaceID], nil
}

// providerWithAlias builds a single-model provider registering alias→name.
func providerWithAlias(providerName, alias, modelName string) *agentsv1.ModelProvider {
	return &agentsv1.ModelProvider{
		Name: providerName,
		Type: "gemini",
		Models: []*agentsv1.ModelConfig{
			{Alias: alias, Name: modelName},
		},
	}
}

// --- Fake model.LLM ---

type fakeLLM struct {
	response *model.LLMResponse
	err      error
	called   bool
	name     string
}

func (f *fakeLLM) Name() string { return f.name }

func (f *fakeLLM) GenerateContent(_ context.Context, _ *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	f.called = true
	return func(yield func(*model.LLMResponse, error) bool) {
		if f.err != nil {
			yield(nil, f.err)
			return
		}
		yield(f.response, nil)
	}
}

// blockingLLM blocks until the request context is cancelled, then yields the
// context error. Used to exercise the bounded-timeout path.
type blockingLLM struct{}

func (b *blockingLLM) Name() string { return "blocking" }

func (b *blockingLLM) GenerateContent(ctx context.Context, _ *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		<-ctx.Done()
		yield(nil, ctx.Err())
	}
}

// fakeResolve returns a resolveModel seam handing out llm, and records the
// model refs and provider lists it was called with.
type fakeResolve struct {
	llm       model.LLM
	err       error
	refs      []string
	providers [][]agentsv1.ModelProvider
}

func (f *fakeResolve) fn(_ context.Context, modelRef string, providers []agentsv1.ModelProvider) (model.LLM, error) {
	f.refs = append(f.refs, modelRef)
	f.providers = append(f.providers, providers)
	if f.err != nil {
		return nil, f.err
	}
	return f.llm, nil
}

func textResponse(text string) *model.LLMResponse {
	return &model.LLMResponse{
		Content: &genai.Content{
			Parts: []*genai.Part{{Text: text}},
		},
	}
}

// --- deriveAgentNameFromEvents tests ---

func TestDeriveAgentName_FindsFirstNonUser(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("my-agent", textPart("Hi there")),
	}
	got := deriveAgentNameFromEvents(events)
	if got != "my-agent" {
		t.Fatalf("expected 'my-agent', got %q", got)
	}
}

func TestDeriveAgentName_SkipsToolEvents(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("tool-agent", funcCallPart("search")),
		makeEvent("real-agent", textPart("Response")),
	}
	got := deriveAgentNameFromEvents(events)
	if got != "real-agent" {
		t.Fatalf("expected 'real-agent', got %q", got)
	}
}

func TestDeriveAgentName_EmptyWhenNoAgent(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
	}
	got := deriveAgentNameFromEvents(events)
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestDeriveAgentName_EmptyEvents(t *testing.T) {
	got := deriveAgentNameFromEvents(nil)
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

// --- buildTitleInput tests ---

func TestBuildTitleInput_UserAndAssistant(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello world")),
		makeEvent("agent", textPart("Hi there")),
	}
	input := buildTitleInput(events)
	if input == "" {
		t.Fatal("expected non-empty input")
	}
	if !strings.Contains(input, "User: Hello world") {
		t.Fatalf("expected user text in input, got %q", input)
	}
	if !strings.Contains(input, "Assistant: Hi there") {
		t.Fatalf("expected assistant text in input, got %q", input)
	}
}

func TestBuildTitleInput_UserOnly(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello world")),
	}
	input := buildTitleInput(events)
	if !strings.Contains(input, "User: Hello world") {
		t.Fatalf("expected user text in input, got %q", input)
	}
}

func TestBuildTitleInput_EmptyEvents(t *testing.T) {
	input := buildTitleInput(nil)
	if input != "" {
		t.Fatalf("expected empty input, got %q", input)
	}
}

func TestBuildTitleInput_SkipsToolEvents(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Question")),
		makeEvent("agent", funcCallPart("search")),
		makeEvent("agent", funcRespPart("search")),
		makeEvent("agent", textPart("Answer")),
	}
	input := buildTitleInput(events)
	if !strings.Contains(input, "User: Question") {
		t.Fatalf("expected user text, got %q", input)
	}
	if !strings.Contains(input, "Assistant: Answer") {
		t.Fatalf("expected assistant text, got %q", input)
	}
}

func TestBuildTitleInput_SkipsNonFinalAssistantText(t *testing.T) {
	// Assistant text emitted alongside a function call is not a final
	// response; the later text-only event is.
	mixed := makeEvent("agent", textPart("Let me look that up"))
	mixed.Content.Parts = append(mixed.Content.Parts, &genai.Part{
		FunctionCall: &genai.FunctionCall{Name: "search"},
	})
	events := []*session.Event{
		makeEvent("user", textPart("Question")),
		mixed,
		makeEvent("agent", funcRespPart("search")),
		makeEvent("agent", textPart("Final answer")),
	}
	input := buildTitleInput(events)
	if strings.Contains(input, "Let me look that up") {
		t.Fatalf("expected intermediate assistant text to be skipped, got %q", input)
	}
	if !strings.Contains(input, "Assistant: Final answer") {
		t.Fatalf("expected final assistant text, got %q", input)
	}
}

func TestBuildTitleInput_BoundsCombinedInput(t *testing.T) {
	longUser := strings.Repeat("甲", 3000)
	longAssistant := strings.Repeat("b", 3000)
	events := []*session.Event{
		makeEvent("user", textPart(longUser)),
		makeEvent("agent", textPart(longAssistant)),
	}
	input := buildTitleInput(events)
	// Total budget plus the "\nUser: " / "\nAssistant: " scaffolding.
	scaffolding := len([]rune("\nUser: \nAssistant: "))
	if got := len([]rune(input)); got > titleGenMaxInputCodePoints+scaffolding {
		t.Fatalf("combined input exceeds budget: %d code points", got)
	}
	if !strings.Contains(input, "Assistant: ") || !strings.Contains(input, "bbb") {
		t.Fatalf("expected assistant excerpt to be present, got tail %q", input[len(input)-40:])
	}
}

// --- extractModelResponseText tests ---

func TestExtractModelResponseText_ValidResponse(t *testing.T) {
	got := extractModelResponseText(textResponse("My Title"))
	if got != "My Title" {
		t.Fatalf("expected 'My Title', got %q", got)
	}
}

func TestExtractModelResponseText_NilResponse(t *testing.T) {
	got := extractModelResponseText(nil)
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestExtractModelResponseText_EmptyContent(t *testing.T) {
	resp := &model.LLMResponse{Content: nil}
	got := extractModelResponseText(resp)
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestExtractModelResponseText_TrimsWhitespace(t *testing.T) {
	got := extractModelResponseText(textResponse("  My Title  \n"))
	if got != "My Title" {
		t.Fatalf("expected trimmed 'My Title', got %q", got)
	}
}

// --- titleGenerator.generate tests ---

func defaultTitleEvents() []*session.Event {
	return []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent", textPart("Hi")),
	}
}

func llmAgentResolver(workspaceID, modelRef string) *fakeTitleResolver {
	return &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: modelRef, workspaceID: workspaceID, agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
}

func TestGenerateLLMTitle_NoResolver(t *testing.T) {
	g := titleGenerator{}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if ok {
		t.Fatal("expected fallback when no resolver")
	}
}

func TestGenerateLLMTitle_NoAgentInEvents(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
	}
	g := titleGenerator{resolver: &fakeTitleResolver{agents: map[string]fakeAgentInfo{}}}
	_, ok := g.generate(context.Background(), events, "s1")
	if ok {
		t.Fatal("expected fallback when no agent in events")
	}
}

func TestGenerateLLMTitle_NoWorkspace(t *testing.T) {
	g := titleGenerator{resolver: llmAgentResolver("", "flash")}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if ok {
		t.Fatal("expected fallback when no workspace")
	}
}

func TestGenerateLLMTitle_NonLLMAgent(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("workflow-agent", textPart("Hi")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"workflow-agent": {model: "flash", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_WORKFLOW},
	}}
	g := titleGenerator{resolver: resolver}
	_, ok := g.generate(context.Background(), events, "s1")
	if ok {
		t.Fatal("expected fallback for non-LLM agent")
	}
}

func TestGenerateLLMTitle_UnspecifiedAgentType(t *testing.T) {
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "flash", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED},
	}}
	g := titleGenerator{resolver: resolver}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if ok {
		t.Fatal("expected fallback for AGENT_TYPE_UNSPECIFIED")
	}
}

func TestGenerateLLMTitle_NoModel(t *testing.T) {
	g := titleGenerator{resolver: llmAgentResolver("ws1", "")}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if ok {
		t.Fatal("expected fallback when no model")
	}
}

func TestGenerateLLMTitle_ProviderListError(t *testing.T) {
	g := titleGenerator{
		resolver:       llmAgentResolver("ws1", "flash"),
		providerLister: &fakeProviderLister{err: errors.New("db error")},
	}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if ok {
		t.Fatal("expected fallback on provider list error")
	}
}

func TestGenerateLLMTitle_EmptyInput(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", imagePart()),
		makeEvent("agent", funcCallPart("analyze")),
	}
	g := titleGenerator{resolver: llmAgentResolver("ws1", "flash")}
	_, ok := g.generate(context.Background(), events, "s1")
	if ok {
		t.Fatal("expected fallback on empty input")
	}
}

func TestGenerateLLMTitle_UnresolvedModelNeverCallsProvider(t *testing.T) {
	// The agent's model alias is not registered in its workspace: generation
	// must fall back deterministically WITHOUT constructing any model —
	// never reaching for credentials outside the workspace's providers.
	resolve := &fakeResolve{llm: &fakeLLM{response: textResponse("t")}}
	g := titleGenerator{
		resolver:       llmAgentResolver("ws1", "flash"),
		providerLister: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{"ws1": {}}},
		resolveModel:   resolve.fn,
	}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if ok {
		t.Fatal("expected fallback when model does not resolve in workspace")
	}
	if len(resolve.refs) != 0 {
		t.Fatalf("expected no model construction, got calls for %v", resolve.refs)
	}
}

func TestGenerateLLMTitle_Success(t *testing.T) {
	llm := &fakeLLM{response: textResponse("  Greeting Chat  \n")}
	resolve := &fakeResolve{llm: llm}
	g := titleGenerator{
		resolver: llmAgentResolver("ws1", "flash"),
		providerLister: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws1": {providerWithAlias("p1", "flash", "gemini-2.0-flash")},
		}},
		resolveModel: resolve.fn,
	}
	title, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if !ok {
		t.Fatal("expected successful LLM title")
	}
	if !llm.called {
		t.Fatal("expected the LLM to be called")
	}
	if title != "Greeting Chat" {
		t.Fatalf("expected normalized title 'Greeting Chat', got %q", title)
	}
}

func TestGenerateLLMTitle_NormalizesLongMultilineOutput(t *testing.T) {
	raw := strings.Repeat("很", 20) + "\n" + strings.Repeat("长", 20)
	llm := &fakeLLM{response: textResponse(raw)}
	resolve := &fakeResolve{llm: llm}
	g := titleGenerator{
		resolver: llmAgentResolver("ws1", "flash"),
		providerLister: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws1": {providerWithAlias("p1", "flash", "gemini-2.0-flash")},
		}},
		resolveModel: resolve.fn,
	}
	title, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if !ok {
		t.Fatal("expected successful LLM title")
	}
	if strings.ContainsAny(title, "\n\r") {
		t.Fatalf("expected single-line title, got %q", title)
	}
	if got := len([]rune(title)); got > maxAutoTitleCodePoints {
		t.Fatalf("expected at most %d code points, got %d (%q)", maxAutoTitleCodePoints, got, title)
	}
}

func TestGenerateLLMTitle_ModelCallError(t *testing.T) {
	resolve := &fakeResolve{llm: &fakeLLM{err: errors.New("provider exploded")}}
	g := titleGenerator{
		resolver: llmAgentResolver("ws1", "flash"),
		providerLister: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws1": {providerWithAlias("p1", "flash", "gemini-2.0-flash")},
		}},
		resolveModel: resolve.fn,
	}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if ok {
		t.Fatal("expected fallback on model call error")
	}
}

func TestGenerateLLMTitle_Timeout(t *testing.T) {
	resolve := &fakeResolve{llm: &blockingLLM{}}
	g := titleGenerator{
		resolver: llmAgentResolver("ws1", "flash"),
		providerLister: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws1": {providerWithAlias("p1", "flash", "gemini-2.0-flash")},
		}},
		resolveModel: resolve.fn,
		timeout:      10 * time.Millisecond,
	}
	done := make(chan struct{})
	var ok bool
	go func() {
		_, ok = g.generate(context.Background(), defaultTitleEvents(), "s1")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("generate did not honor the bounded timeout")
	}
	if ok {
		t.Fatal("expected fallback on timeout")
	}
}

func TestGenerateLLMTitle_EmptyModelOutput(t *testing.T) {
	resolve := &fakeResolve{llm: &fakeLLM{response: textResponse("   \n  ")}}
	g := titleGenerator{
		resolver: llmAgentResolver("ws1", "flash"),
		providerLister: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws1": {providerWithAlias("p1", "flash", "gemini-2.0-flash")},
		}},
		resolveModel: resolve.fn,
	}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if ok {
		t.Fatal("expected fallback on empty model output")
	}
}

// --- workspace isolation tests ---

func TestGenerateLLMTitle_DuplicateAliasAcrossWorkspaces(t *testing.T) {
	// The same alias "flash" exists in two workspaces backed by different
	// providers. Generation for agent-a (ws-a) must resolve against ws-a's
	// provider list only.
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent-a", textPart("Hi from A")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent-a": {model: "flash", workspaceID: "ws-a", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
		"agent-b": {model: "flash", workspaceID: "ws-b", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
	tracking := &trackingProviderLister{
		inner: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws-a": {providerWithAlias("provider-a", "flash", "model-a")},
			"ws-b": {providerWithAlias("provider-b", "flash", "model-b")},
		}},
	}
	llm := &fakeLLM{response: textResponse("Title A")}
	resolve := &fakeResolve{llm: llm}
	g := titleGenerator{
		resolver:       resolver,
		providerLister: tracking,
		resolveModel:   resolve.fn,
	}

	title, ok := g.generate(context.Background(), events, "s1")
	if !ok {
		t.Fatal("expected successful LLM title")
	}
	if title != "Title A" {
		t.Fatalf("expected 'Title A', got %q", title)
	}
	if tracking.queriedWorkspace != "ws-a" {
		t.Fatalf("expected workspace 'ws-a' to be queried, got %q", tracking.queriedWorkspace)
	}
	if len(resolve.providers) != 1 || len(resolve.providers[0]) != 1 {
		t.Fatalf("expected one resolution against one provider, got %v", resolve.providers)
	}
	if got := resolve.providers[0][0].GetName(); got != "provider-a" {
		t.Fatalf("expected resolution against provider-a's credentials, got %q", got)
	}
}

func TestGenerateLLMTitle_WorkspaceIsolation_EmptyWorkspaceFallsBack(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent-a", textPart("Hi from A")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent-a": {model: "flash", workspaceID: "ws-a", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
	// ws-a has no providers even though ws-b registers the alias: no call
	// may leave ws-a.
	tracking := &trackingProviderLister{
		inner: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws-a": {},
			"ws-b": {providerWithAlias("provider-b", "flash", "model-b")},
		}},
	}
	resolve := &fakeResolve{llm: &fakeLLM{response: textResponse("t")}}
	g := titleGenerator{resolver: resolver, providerLister: tracking, resolveModel: resolve.fn}

	_, ok := g.generate(context.Background(), events, "s1")
	if ok {
		t.Fatal("expected fallback when the agent's workspace has no matching provider")
	}
	if tracking.queriedWorkspace != "ws-a" {
		t.Fatalf("expected workspace 'ws-a' to be queried, got %q", tracking.queriedWorkspace)
	}
	if len(resolve.refs) != 0 {
		t.Fatalf("expected no model construction, got calls for %v", resolve.refs)
	}
}

// --- model preference tests ---

func TestGenerateLLMTitle_PrefersDedicatedModel(t *testing.T) {
	llm := &fakeLLM{response: textResponse("Cheap Title")}
	resolve := &fakeResolve{llm: llm}
	g := titleGenerator{
		resolver: llmAgentResolver("ws1", "expensive-model"),
		providerLister: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws1": {
				providerWithAlias("p1", "cheap-model", "gemini-2.0-flash"),
				providerWithAlias("p2", "expensive-model", "gemini-2.0-pro"),
			},
		}},
		chatTitleModel: "cheap-model",
		resolveModel:   resolve.fn,
	}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if !ok {
		t.Fatal("expected successful LLM title")
	}
	if len(resolve.refs) != 1 || resolve.refs[0] != "cheap-model" {
		t.Fatalf("expected dedicated model to be resolved first, got %v", resolve.refs)
	}
}

func TestGenerateLLMTitle_AgentModelFallback_WhenDedicatedNotInWorkspace(t *testing.T) {
	// chat_title_model doesn't resolve in ws1; the agent's own model does.
	llm := &fakeLLM{response: textResponse("Agent Model Title")}
	resolve := &fakeResolve{llm: llm}
	g := titleGenerator{
		resolver: llmAgentResolver("ws1", "agent-model"),
		providerLister: &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
			"ws1": {providerWithAlias("p1", "agent-model", "gemini-2.0-flash")},
		}},
		chatTitleModel: "bad-model",
		resolveModel:   resolve.fn,
	}
	_, ok := g.generate(context.Background(), defaultTitleEvents(), "s1")
	if !ok {
		t.Fatal("expected successful LLM title via agent model")
	}
	if len(resolve.refs) != 1 || resolve.refs[0] != "agent-model" {
		t.Fatalf("expected only the agent model to be resolved, got %v", resolve.refs)
	}
}

// --- GenerateSessionTitle handler-level tests for LLM ---

func newLLMTitleTestService(store *stubTitleStore, resolver TitleModelResolver, lister WorkspaceModelProviderLister, chatModel string, sessions ...session.Session) *SessionServiceServer {
	svc := NewSessionServiceServer()
	svc.titleStore = store
	svc.sessionSvc = &titleSeamSessionService{sessions: sessions}
	svc.titleResolver = resolver
	svc.titleProviderLister = lister
	svc.chatTitleModel = chatModel
	return svc
}

func TestGenerateSessionTitle_FallsThroughToDeterm_WhenNoResolver(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Hello"))},
	}
	store := &stubTitleStore{}
	svc := newLLMTitleTestService(store, nil, nil, "", sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Msg.GetGenerated() {
		t.Fatal("expected generated=true from deterministic fallback")
	}
	if resp.Msg.GetSession().GetTitle() != "Hello" {
		t.Fatalf("expected deterministic title 'Hello', got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestGenerateSessionTitle_FallsThroughToDeterm_WhenNoAgent(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Hello"))},
	}
	store := &stubTitleStore{}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{}}
	svc := newLLMTitleTestService(store, resolver, nil, "", sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Msg.GetGenerated() {
		t.Fatal("expected generated=true from deterministic fallback")
	}
	if resp.Msg.GetSession().GetTitle() != "Hello" {
		t.Fatalf("expected deterministic title 'Hello', got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestGenerateSessionTitle_ExistingTitleStillWins(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "Manual Title", state: &fakeState{data: map[string]any{}}},
		events: []*session.Event{
			makeEvent("user", textPart("Hello")),
			makeEvent("my-agent", textPart("Hi")),
		},
	}
	store := &stubTitleStore{}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"my-agent": {model: "flash", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
	svc := newLLMTitleTestService(store, resolver, nil, "flash", sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("expected generated=false when title already exists")
	}
	if resp.Msg.GetSession().GetTitle() != "Manual Title" {
		t.Fatalf("expected existing title, got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestGenerateSessionTitle_ConcurrentManualUpdateStillWins(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events: []*session.Event{
			makeEvent("user", textPart("Hello")),
			makeEvent("agent", textPart("Hi")),
		},
	}
	// CAS returns false (manual title won the race).
	store := &stubTitleStore{existingTitle: "Manual Override"}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "flash", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
	svc := newLLMTitleTestService(store, resolver, nil, "", sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("expected generated=false when manual update won")
	}
	if resp.Msg.GetSession().GetTitle() != "Manual Override" {
		t.Fatalf("expected manual title, got %q", resp.Msg.GetSession().GetTitle())
	}
}

// trackingProviderLister wraps a WorkspaceModelProviderLister and records
// the workspace that was queried.
type trackingProviderLister struct {
	inner            WorkspaceModelProviderLister
	queriedWorkspace string
}

func (t *trackingProviderLister) ListModelProviders(ctx context.Context, workspaceID string) ([]*agentsv1.ModelProvider, error) {
	t.queriedWorkspace = workspaceID
	return t.inner.ListModelProviders(ctx, workspaceID)
}
