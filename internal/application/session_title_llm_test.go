package application

import (
	"context"
	"errors"
	"iter"
	"strings"
	"testing"

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

func (f *fakeTitleResolver) GetAgentModel(name string) string {
	if a, ok := f.agents[name]; ok {
		return a.model
	}
	return ""
}

func (f *fakeTitleResolver) GetAgentWorkspaceID(name string) string {
	if a, ok := f.agents[name]; ok {
		return a.workspaceID
	}
	return ""
}

func (f *fakeTitleResolver) GetAgentType(name string) agentsv1.AgentType {
	if a, ok := f.agents[name]; ok {
		return a.agentType
	}
	return agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED
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

// --- extractModelResponseText tests ---

func TestExtractModelResponseText_ValidResponse(t *testing.T) {
	resp := &model.LLMResponse{
		Content: &genai.Content{
			Parts: []*genai.Part{
				{Text: "My Title"},
			},
		},
	}
	got := extractModelResponseText(resp)
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
	resp := &model.LLMResponse{
		Content: &genai.Content{
			Parts: []*genai.Part{
				{Text: "  My Title  \n"},
			},
		},
	}
	got := extractModelResponseText(resp)
	if got != "My Title" {
		t.Fatalf("expected trimmed 'My Title', got %q", got)
	}
}

// --- generateLLMTitle integration tests ---

func TestGenerateLLMTitle_NoResolver(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent", textPart("Hi")),
	}
	_, ok := generateLLMTitle(context.Background(), events, "", nil, nil, "s1")
	if ok {
		t.Fatal("expected fallback when no resolver")
	}
}

func TestGenerateLLMTitle_NoAgentInEvents(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{}}
	_, ok := generateLLMTitle(context.Background(), events, "", resolver, nil, "s1")
	if ok {
		t.Fatal("expected fallback when no agent in events")
	}
}

func TestGenerateLLMTitle_NoWorkspace(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent", textPart("Hi")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "flash", workspaceID: "", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
	_, ok := generateLLMTitle(context.Background(), events, "", resolver, nil, "s1")
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
	_, ok := generateLLMTitle(context.Background(), events, "", resolver, nil, "s1")
	if ok {
		t.Fatal("expected fallback for non-LLM agent")
	}
}

func TestGenerateLLMTitle_NoModel(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent", textPart("Hi")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
	_, ok := generateLLMTitle(context.Background(), events, "", resolver, nil, "s1")
	if ok {
		t.Fatal("expected fallback when no model")
	}
}

func TestGenerateLLMTitle_ProviderListError(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent", textPart("Hi")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "flash", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
	lister := &fakeProviderLister{err: errors.New("db error")}
	_, ok := generateLLMTitle(context.Background(), events, "", resolver, lister, "s1")
	if ok {
		t.Fatal("expected fallback on provider list error")
	}
}

func TestGenerateLLMTitle_EmptyInput(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", imagePart()),
		makeEvent("agent", funcCallPart("analyze")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "flash", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}
	_, ok := generateLLMTitle(context.Background(), events, "", resolver, nil, "s1")
	if ok {
		t.Fatal("expected fallback on empty input")
	}
}

// --- DuplicateAliases cross-workspace isolation test ---

func TestGenerateLLMTitle_WorkspaceIsolation(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent-a", textPart("Hi from A")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent-a": {model: "flash", workspaceID: "ws-a", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
		"agent-b": {model: "flash", workspaceID: "ws-b", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}

	// Track which workspace's providers were queried.
	lister := &trackingProviderLister{
		inner: &fakeProviderLister{
			providers: map[string][]*agentsv1.ModelProvider{
				"ws-a": {},
				"ws-b": {},
			},
		},
	}

	// Falls back because no provider matches, but we assert it queried ws-a.
	_, ok := generateLLMTitle(context.Background(), events, "", resolver, lister, "s1")
	if ok {
		t.Fatal("expected fallback (no real provider)")
	}
	if lister.queriedWorkspace != "ws-a" {
		t.Fatalf("expected workspace 'ws-a' to be queried, got %q", lister.queriedWorkspace)
	}
}

func TestGenerateLLMTitle_PrefersDedicatedModel(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent", textPart("Hi")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "expensive-model", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}

	lister := &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
		"ws1": {},
	}}

	// With chat_title_model="cheap-model" set, the dedicated model should
	// be tried. Since no provider matches either, falls back.
	_, ok := generateLLMTitle(context.Background(), events, "cheap-model", resolver, lister, "s1")
	if ok {
		t.Fatal("expected fallback when no provider matches")
	}
}

func TestGenerateLLMTitle_AgentModelFallback_WhenDedicatedFails(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent", textPart("Hi")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "agent-model", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_LLM},
	}}

	lister := &fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
		"ws1": {},
	}}

	// With chat_title_model="bad-model", dedicated fails, then the agent's
	// "agent-model" should be tried. Both fail here (no matching provider),
	// but the retry logic is exercised.
	_, ok := generateLLMTitle(context.Background(), events, "bad-model", resolver, lister, "s1")
	if ok {
		t.Fatal("expected fallback when no provider matches")
	}
}

func TestGenerateLLMTitle_UnspecifiedAgentType(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello")),
		makeEvent("agent", textPart("Hi")),
	}
	resolver := &fakeTitleResolver{agents: map[string]fakeAgentInfo{
		"agent": {model: "flash", workspaceID: "ws1", agentType: agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED},
	}}
	_, ok := generateLLMTitle(context.Background(), events, "", resolver, nil, "s1")
	if ok {
		t.Fatal("expected fallback for AGENT_TYPE_UNSPECIFIED")
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
	inner             WorkspaceModelProviderLister
	queriedWorkspace  string
}

func (t *trackingProviderLister) ListModelProviders(ctx context.Context, workspaceID string) ([]*agentsv1.ModelProvider, error) {
	t.queriedWorkspace = workspaceID
	return t.inner.ListModelProviders(ctx, workspaceID)
}
