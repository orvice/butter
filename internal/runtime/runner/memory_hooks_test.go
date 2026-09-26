package runner

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/plugin"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	memoryconfigmem "go.orx.me/apps/butter/internal/repo/memoryconfig/memory"
	"go.orx.me/apps/butter/internal/runtime/mem0memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	"go.orx.me/apps/butter/internal/runtime/memoryhook"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// memoryTestMem0 is a fake mem0 OSS server: searches under Workspace Memory
// return one memory, and adds are recorded.
type memoryTestMem0 struct {
	srv *httptest.Server

	mu       sync.Mutex
	searches []map[string]any
	adds     []map[string]any
}

func newMemoryTestMem0(t *testing.T) *memoryTestMem0 {
	f := &memoryTestMem0{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		defer f.mu.Unlock()
		switch r.URL.Path {
		case "/search":
			f.searches = append(f.searches, body)
			filters, _ := body["filters"].(map[string]any)
			if filters["user_id"] == "ws:ws1" {
				_, _ = w.Write([]byte(`{"results":[{"id":"m1","memory":"The team deploys with pnpm","score":0.9,"created_at":"2026-09-01T10:00:00+00:00"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"results":[]}`))
		case "/memories":
			f.adds = append(f.adds, body)
			_, _ = w.Write([]byte(`{"results":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *memoryTestMem0) recorded() (searches, adds []map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]any(nil), f.searches...), append([]map[string]any(nil), f.adds...)
}

// recordingModel answers every call with a fixed reply and records the
// system instruction each call carried.
type recordingModel struct {
	reply string

	mu      sync.Mutex
	systems []string
}

func (m *recordingModel) Name() string { return "recording-model" }

func (m *recordingModel) GenerateContent(_ context.Context, req *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	var system strings.Builder
	if req.Config != nil && req.Config.SystemInstruction != nil {
		for _, p := range req.Config.SystemInstruction.Parts {
			system.WriteString(p.Text)
		}
	}
	m.mu.Lock()
	m.systems = append(m.systems, system.String())
	m.mu.Unlock()
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(&model.LLMResponse{
			Content:      genai.NewContentFromText(m.reply, genai.RoleModel),
			FinishReason: genai.FinishReasonStop,
		}, nil)
	}
}

func (m *recordingModel) recordedSystems() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.systems...)
}

type memoryFixture struct {
	svc      *Service
	hooks    *memoryhook.Hooks
	mem0     *memoryTestMem0
	model    *recordingModel
	sessions session.Service
}

// newMemoryFixture registers a Sequential root ("mem-root") with two LLM
// sub-agents that share one recording model, and wires memory hooks to a
// fake mem0 server with an enabled config for ws1.
func newMemoryFixture(t *testing.T, mc *agentsv1.MemoryConfig) *memoryFixture {
	t.Helper()
	fake := newMemoryTestMem0(t)
	repo := memoryconfigmem.New()
	if _, err := repo.Put(t.Context(), "ws1", &agentsv1.WorkspaceMemoryConfig{BaseUrl: fake.srv.URL, Enabled: true}, nil); err != nil {
		t.Fatalf("Put config: %v", err)
	}
	memSvc := mem0memory.New(memoryconn.NewResolver(repo, nil))
	memSvc.HTTPClient = fake.srv.Client()
	sessions := session.InMemoryService()

	svc, err := NewService(t.Context(), nil, nil, nil, nil, nil, sessions, memSvc, nil,
		adkrunner.PluginConfig{Plugins: []*plugin.Plugin{memoryhook.InjectionPlugin()}})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	hooks := memoryhook.New(memSvc, sessions)
	svc.SetMemoryHooks(hooks)

	rec := &recordingModel{reply: "Use pnpm."}
	var subs []agent.Agent
	for _, name := range []string{"planner", "writer"} {
		sub, err := llmagent.New(llmagent.Config{Name: name, Model: rec, Instruction: "You are " + name + "."})
		if err != nil {
			t.Fatalf("llmagent.New: %v", err)
		}
		subs = append(subs, sub)
	}
	root, err := sequentialagent.New(sequentialagent.Config{AgentConfig: agent.Config{Name: "mem-root", SubAgents: subs}})
	if err != nil {
		t.Fatalf("sequentialagent.New: %v", err)
	}
	svc.RegisterAgent("mem-root", root)
	svc.mu.Lock()
	svc.agentsProto["mem-root"] = &agentsv1.Agent{
		Name:        "mem-root",
		AgentId:     "mem-root",
		WorkspaceId: "ws1",
		Type:        agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
		Config:      &agentsv1.AgentConfig{Memory: mc},
	}
	svc.mu.Unlock()
	return &memoryFixture{svc: svc, hooks: hooks, mem0: fake, model: rec, sessions: sessions}
}

func memoryCtxInfo(sessionID string) *agentsv1.ContextInfo {
	return &agentsv1.ContextInfo{
		Uuid: sessionID, SessionId: sessionID, UserId: "u1", ChannelName: "web-chat", WorkspaceId: "ws1",
	}
}

func TestMemoryRecallIsInjectedIntoEveryModelCallAndNeverPersisted(t *testing.T) {
	fx := newMemoryFixture(t, &agentsv1.MemoryConfig{Enabled: true})

	turn, err := fx.svc.RunTurn(t.Context(), "mem-root", []*genai.Part{{Text: "Which package manager should the build use?"}}, "", memoryCtxInfo("s1"), nil, nil)
	if err != nil {
		t.Fatalf("RunTurn: %v", err)
	}
	fx.hooks.Wait()

	if turn.InvocationID == "" {
		t.Fatal("TurnResult.InvocationID not set")
	}

	// Recall ran once per turn: one search per scope, not one per model call.
	searches, adds := fx.mem0.recorded()
	if len(searches) != 2 {
		t.Fatalf("searches = %d, want 2 (workspace + agent, once per turn)", len(searches))
	}
	if searches[0]["query"] != "Which package manager should the build use?" {
		t.Fatalf("recall query = %v", searches[0]["query"])
	}

	// Both LLM sub-agents of the composite root saw the block.
	systems := fx.model.recordedSystems()
	if len(systems) != 2 {
		t.Fatalf("model calls = %d, want 2", len(systems))
	}
	for i, sys := range systems {
		if !strings.Contains(sys, "<memories>") || !strings.Contains(sys, "[2026-09-01] The team deploys with pnpm") {
			t.Fatalf("model call %d system instruction lacks the memory block:\n%s", i, sys)
		}
		if strings.Count(sys, "<memories>") != 1 {
			t.Fatalf("model call %d carries the block more than once:\n%s", i, sys)
		}
	}

	// The block never reaches session history.
	resp, err := fx.sessions.Get(t.Context(), &session.GetRequest{AppName: "web-chat", UserID: "u1", SessionID: "s1"})
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	for ev := range resp.Session.Events().All() {
		if ev.Content == nil {
			continue
		}
		for _, p := range ev.Content.Parts {
			if strings.Contains(p.Text, "<memories>") || strings.Contains(p.Text, "deploys with pnpm") {
				t.Fatalf("recalled memory persisted in session event by %s: %q", ev.Author, p.Text)
			}
		}
	}

	// Capture sent this turn to Workspace Memory in the background.
	if len(adds) != 1 {
		t.Fatalf("adds = %d, want 1", len(adds))
	}
	add := adds[0]
	if add["user_id"] != "ws:ws1" || add["infer"] != true {
		t.Fatalf("capture body = %v", add)
	}
	if _, ok := add["agent_id"]; ok {
		t.Fatalf("capture wrote Agent Memory: %v", add)
	}
	msgs, _ := add["messages"].([]any)
	if len(msgs) != 3 {
		t.Fatalf("captured messages = %v, want the user text and both sub-agents' replies", add["messages"])
	}
	first, _ := msgs[0].(map[string]any)
	if first["role"] != "user" || first["content"] != "Which package manager should the build use?" {
		t.Fatalf("first captured message = %v", first)
	}
	meta, _ := add["metadata"].(map[string]any)
	if meta["butter_invocation_id"] != turn.InvocationID || meta["butter_agent_id"] != "mem-root" || meta["butter_channel"] != "web-chat" {
		t.Fatalf("capture metadata = %v", meta)
	}
}

func TestMemoryCaptureSendsOnlyTheCurrentTurn(t *testing.T) {
	fx := newMemoryFixture(t, &agentsv1.MemoryConfig{Enabled: true})
	for _, text := range []string{"first question about deploys", "second question about builds"} {
		if _, err := fx.svc.RunTurn(t.Context(), "mem-root", []*genai.Part{{Text: text}}, "", memoryCtxInfo("s2"), nil, nil); err != nil {
			t.Fatalf("RunTurn: %v", err)
		}
		fx.hooks.Wait()
	}
	_, adds := fx.mem0.recorded()
	if len(adds) != 2 {
		t.Fatalf("adds = %d, want one per turn", len(adds))
	}
	second := adds[1]["messages"].([]any)
	for _, m := range second {
		if strings.Contains(m.(map[string]any)["content"].(string), "first question") {
			t.Fatalf("second capture resent the first turn: %v", second)
		}
	}
}

func TestMemoryShortReplyRecallsWithThePreviousUserMessage(t *testing.T) {
	fx := newMemoryFixture(t, &agentsv1.MemoryConfig{Enabled: true, DisableAutoCapture: true})
	for _, text := range []string{"How do we deploy the web app?", "ok, go on"} {
		if _, err := fx.svc.RunTurn(t.Context(), "mem-root", []*genai.Part{{Text: text}}, "", memoryCtxInfo("s3"), nil, nil); err != nil {
			t.Fatalf("RunTurn: %v", err)
		}
	}
	searches, adds := fx.mem0.recorded()
	if len(adds) != 0 {
		t.Fatalf("disable_auto_capture still captured: %v", adds)
	}
	last := searches[len(searches)-1]["query"]
	if last != "How do we deploy the web app?\nok, go on" {
		t.Fatalf("short-reply recall query = %q", last)
	}
}

func TestMemoryDisabledOrRecallOffInjectsNothing(t *testing.T) {
	for name, mc := range map[string]*agentsv1.MemoryConfig{
		"no memory config": nil,
		"memory disabled":  {Enabled: false},
		"auto recall off":  {Enabled: true, DisableAutoRecall: true, DisableAutoCapture: true},
	} {
		t.Run(name, func(t *testing.T) {
			fx := newMemoryFixture(t, mc)
			if _, err := fx.svc.RunTurn(t.Context(), "mem-root", []*genai.Part{{Text: "Which package manager?"}}, "", memoryCtxInfo("s4"), nil, nil); err != nil {
				t.Fatalf("RunTurn: %v", err)
			}
			fx.hooks.Wait()
			if searches, adds := fx.mem0.recorded(); len(searches)+len(adds) != 0 {
				t.Fatalf("mem0 was called: searches=%v adds=%v", searches, adds)
			}
			for _, sys := range fx.model.recordedSystems() {
				if strings.Contains(sys, "<memories>") {
					t.Fatalf("block injected: %s", sys)
				}
			}
		})
	}
}

func TestMemoryRecallFailureDegrades(t *testing.T) {
	fx := newMemoryFixture(t, &agentsv1.MemoryConfig{Enabled: true, DisableAutoCapture: true})
	fx.mem0.srv.Close() // mem0 is down

	out, err := fx.svc.Run(t.Context(), "mem-root", []*genai.Part{{Text: "Which package manager?"}}, "", memoryCtxInfo("s5"), nil, nil)
	if err != nil {
		t.Fatalf("Run failed because mem0 is down: %v", err)
	}
	if !strings.Contains(out, "Use pnpm.") {
		t.Fatalf("output = %q", out)
	}
	for _, sys := range fx.model.recordedSystems() {
		if strings.Contains(sys, "<memories>") {
			t.Fatalf("block injected despite failed recall: %s", sys)
		}
	}
}

// An LLM root runs through ADK's node runtime rather than the composite
// path; injection must reach it too.
func TestMemoryRecallReachesAnLLMRoot(t *testing.T) {
	fx := newMemoryFixture(t, &agentsv1.MemoryConfig{Enabled: true, DisableAutoCapture: true})
	rec := &recordingModel{reply: "Use pnpm."}
	root, err := llmagent.New(llmagent.Config{Name: "mem-llm", Model: rec, Instruction: "You help."})
	if err != nil {
		t.Fatalf("llmagent.New: %v", err)
	}
	fx.svc.RegisterAgent("mem-llm", root)
	fx.svc.mu.Lock()
	fx.svc.agentsProto["mem-llm"] = &agentsv1.Agent{
		Name: "mem-llm", AgentId: "mem-llm", WorkspaceId: "ws1",
		Type:   agentsv1.AgentType_AGENT_TYPE_LLM,
		Config: &agentsv1.AgentConfig{Memory: &agentsv1.MemoryConfig{Enabled: true, DisableAutoCapture: true}},
	}
	fx.svc.mu.Unlock()

	if _, err := fx.svc.RunTurn(t.Context(), "mem-llm", []*genai.Part{{Text: "Which package manager should we use?"}}, "", memoryCtxInfo("s6"), nil, nil); err != nil {
		t.Fatalf("RunTurn: %v", err)
	}
	systems := rec.recordedSystems()
	if len(systems) != 1 || !strings.Contains(systems[0], "The team deploys with pnpm") || !strings.Contains(systems[0], "You help.") {
		t.Fatalf("LLM root system instructions = %q", systems)
	}
}
