package memorytool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/mem0"
	memoryconfigmem "go.orx.me/apps/butter/internal/repo/memoryconfig/memory"
	"go.orx.me/apps/butter/internal/runtime/mem0memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	"go.orx.me/apps/butter/internal/runtime/memoryhook"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeMem0 records requests; searches return `results`, adds return one
// stored memory, and a non-zero status forces an error.
type fakeMem0 struct {
	srv *httptest.Server

	mu       sync.Mutex
	searches []map[string]any
	adds     []map[string]any
	results  string
	status   int
}

func newFakeMem0(t *testing.T) *fakeMem0 {
	f := &fakeMem0{results: `{"results":[]}`}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.status != 0 {
			http.Error(w, "boom", f.status)
			return
		}
		switch r.URL.Path {
		case "/search":
			f.searches = append(f.searches, body)
			_, _ = w.Write([]byte(f.results))
		case "/memories":
			f.adds = append(f.adds, body)
			_, _ = w.Write([]byte(`{"results":[{"id":"n1","memory":"Deploys use pnpm","event":"ADD"}]}`))
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeMem0) recorded() (searches, adds []map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]any(nil), f.searches...), append([]map[string]any(nil), f.adds...)
}

// toolCtx is the slice of agent.Context the handlers use: context values
// (for the memory turn) and the invocation ID.
type toolCtx struct {
	agent.Context
	ctx context.Context
}

func (c toolCtx) Deadline() (time.Time, bool) { return c.ctx.Deadline() }
func (c toolCtx) Value(key any) any           { return c.ctx.Value(key) }
func (c toolCtx) Done() <-chan struct{}       { return c.ctx.Done() }
func (c toolCtx) Err() error                  { return c.ctx.Err() }
func (c toolCtx) InvocationID() string        { return "inv-tool" }
func (c toolCtx) UserContent() *genai.Content {
	return nil
}

type readonlyCtx struct {
	agent.ReadonlyContext
	ctx context.Context
}

func (c readonlyCtx) Value(key any) any { return c.ctx.Value(key) }

type fixture struct {
	mem0 *fakeMem0
	svc  *mem0memory.Service
	ts   *Toolset
}

func newFixture(t *testing.T, enabled bool) *fixture {
	t.Helper()
	f := newFakeMem0(t)
	repo := memoryconfigmem.New()
	if _, err := repo.Put(t.Context(), "ws1", &agentsv1.WorkspaceMemoryConfig{BaseUrl: f.srv.URL, Enabled: enabled}, nil); err != nil {
		t.Fatal(err)
	}
	svc := mem0memory.New(memoryconn.NewResolver(repo, nil))
	svc.HTTPClient = f.srv.Client()
	ts, err := NewToolset(svc, "helper")
	if err != nil {
		t.Fatal(err)
	}
	return &fixture{mem0: f, svc: svc, ts: ts.(*Toolset)}
}

// turnCtx runs Hooks.Begin for a root agent and returns the context a tool
// call would carry.
func (fx *fixture) turnCtx(t *testing.T, rootID string, mc *agentsv1.MemoryConfig) context.Context {
	t.Helper()
	mc.DisableAutoRecall = true // keep recall traffic out of the assertions
	root := &agentsv1.Agent{AgentId: rootID, WorkspaceId: "ws1", Config: &agentsv1.AgentConfig{Memory: mc}}
	ctx, _ := memoryhook.New(fx.svc, nil).Begin(t.Context(), root, &agentsv1.ContextInfo{SessionId: "s1", UserId: "u1", ChannelName: "web-chat"}, nil, nil)
	return ctx
}

func toolNames(t *testing.T, fx *fixture, ctx context.Context) []string {
	t.Helper()
	tools, err := fx.ts.Tools(readonlyCtx{ctx: ctx})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tl := range tools {
		names = append(names, tl.Name())
	}
	return names
}

func TestToolsAreOfferedOnlyToTheRootWithEnabledMemory(t *testing.T) {
	fx := newFixture(t, true)
	tools := &agentsv1.MemoryConfig{Enabled: true, EnableTools: true}

	if got := toolNames(t, fx, fx.turnCtx(t, "helper", proto.Clone(tools).(*agentsv1.MemoryConfig))); strings.Join(got, ",") != "search_memory,add_memory" {
		t.Fatalf("root tools = %v", got)
	}
	if got := toolNames(t, fx, t.Context()); len(got) != 0 {
		t.Fatalf("tools outside a memory turn = %v", got)
	}
	// This agent running as another root's sub-agent.
	if got := toolNames(t, fx, fx.turnCtx(t, "other-root", proto.Clone(tools).(*agentsv1.MemoryConfig))); len(got) != 0 {
		t.Fatalf("tools for a sub-agent = %v", got)
	}
	// The root did not enable tools.
	if got := toolNames(t, fx, fx.turnCtx(t, "helper", &agentsv1.MemoryConfig{Enabled: true})); len(got) != 0 {
		t.Fatalf("tools without enable_tools = %v", got)
	}
	// The workspace has no enabled memory config.
	disabled := newFixture(t, false)
	if got := toolNames(t, disabled, disabled.turnCtx(t, "helper", proto.Clone(tools).(*agentsv1.MemoryConfig))); len(got) != 0 {
		t.Fatalf("tools with a disabled workspace config = %v", got)
	}
}

func TestSearchUsesTheTurnScopeAndConfig(t *testing.T) {
	fx := newFixture(t, true)
	fx.mem0.results = `{"results":[{"id":"m1","memory":"The team deploys with pnpm","score":0.9,"created_at":"2026-09-01T10:00:00+00:00"}]}`
	ctx := toolCtx{ctx: fx.turnCtx(t, "helper", &agentsv1.MemoryConfig{Enabled: true, EnableTools: true, TopK: proto.Int32(7), Threshold: proto.Float32(0.5)})}

	got, err := fx.ts.search(ctx, searchArgs{Query: "package manager"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if got.Scope != "workspace" || len(got.Memories) != 1 || got.Memories[0].Memory != "The team deploys with pnpm" || got.Memories[0].Date != "2026-09-01" {
		t.Fatalf("result = %+v", got)
	}

	if _, err := fx.ts.search(ctx, searchArgs{Query: "style", Scope: "agent", TopK: 500}); err != nil {
		t.Fatalf("agent search: %v", err)
	}
	searches, _ := fx.mem0.recorded()
	if len(searches) != 2 {
		t.Fatalf("searches = %d", len(searches))
	}
	first, second := searches[0], searches[1]
	if first["filters"].(map[string]any)["user_id"] != "ws:ws1" || first["top_k"] != float64(7) || first["threshold"] != 0.5 {
		t.Fatalf("workspace search = %v", first)
	}
	if second["filters"].(map[string]any)["agent_id"] != "ws:ws1:agent:helper" || second["top_k"] != float64(maxTopK) {
		t.Fatalf("agent search = %v", second)
	}
}

func TestAddWritesTheChosenScopeWithProvenance(t *testing.T) {
	fx := newFixture(t, true)
	ctx := toolCtx{ctx: fx.turnCtx(t, "helper", &agentsv1.MemoryConfig{Enabled: true, EnableTools: true, AllowAgentScopeWrite: true})}

	got, err := fx.ts.add(ctx, addArgs{Content: "The team deploys with pnpm"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if got.Scope != "workspace" || len(got.Stored) != 1 || got.Stored[0] != "Deploys use pnpm" {
		t.Fatalf("result = %+v", got)
	}
	if _, err := fx.ts.add(ctx, addArgs{Content: "Prefer terse answers", Scope: "agent"}); err != nil {
		t.Fatalf("agent add: %v", err)
	}
	_, adds := fx.mem0.recorded()
	if adds[0]["user_id"] != "ws:ws1" || adds[0]["infer"] != true {
		t.Fatalf("workspace add = %v", adds[0])
	}
	meta := adds[0]["metadata"].(map[string]any)
	if meta["butter_invocation_id"] != "inv-tool" || meta["butter_agent_id"] != "helper" || meta["butter_session_id"] != "s1" {
		t.Fatalf("metadata = %v", meta)
	}
	if adds[1]["agent_id"] != "ws:ws1:agent:helper" {
		t.Fatalf("agent add = %v", adds[1])
	}
}

func TestAddRejections(t *testing.T) {
	fx := newFixture(t, true)
	ctx := toolCtx{ctx: fx.turnCtx(t, "helper", &agentsv1.MemoryConfig{Enabled: true, EnableTools: true})}

	for name, args := range map[string]addArgs{
		"agent scope not allowed": {Content: "x", Scope: "agent"},
		"unknown scope":           {Content: "x", Scope: "global"},
		"empty content":           {Content: "  "},
		"too long":                {Content: strings.Repeat("a", maxContentRunes+1)},
	} {
		if _, err := fx.ts.add(ctx, args); err == nil {
			t.Fatalf("%s: add succeeded", name)
		}
	}
	if _, adds := fx.mem0.recorded(); len(adds) != 0 {
		t.Fatalf("a rejected add reached mem0: %v", adds)
	}

	// Outside a turn this toolset may serve.
	if _, err := fx.ts.add(toolCtx{ctx: t.Context()}, addArgs{Content: "x"}); err == nil || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("add outside a memory turn = %v", err)
	}
}

func TestFailuresAreReadable(t *testing.T) {
	fx := newFixture(t, true)
	ctx := toolCtx{ctx: fx.turnCtx(t, "helper", &agentsv1.MemoryConfig{Enabled: true, EnableTools: true})}
	fx.mem0.status = http.StatusBadGateway

	_, err := fx.ts.search(ctx, searchArgs{Query: "x"})
	if err == nil || !strings.Contains(err.Error(), "memory search failed") {
		t.Fatalf("search error = %v", err)
	}
	if got := readable("save", memoryconn.ErrNotConfigured); got.Error() != "workspace memory is not configured" {
		t.Fatalf("not-configured error = %v", got)
	}
}

func TestFormatSearchCapsText(t *testing.T) {
	found := []mem0.Memory{
		{Memory: strings.Repeat("a", 30)},
		{Memory: strings.Repeat("b", 30)},
		{Memory: strings.Repeat("c", 30)},
	}
	got := formatSearch(mem0memory.TargetAgent, found, 70)
	if len(got.Memories) != 2 || !got.Truncated || got.Scope != "agent" {
		t.Fatalf("formatSearch = %+v", got)
	}
	if empty := formatSearch(mem0memory.TargetWorkspace, nil, 70); empty.Memories == nil || len(empty.Memories) != 0 {
		t.Fatalf("empty result should be an empty list: %+v", empty)
	}
}
