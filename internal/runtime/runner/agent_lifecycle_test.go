package runner

import (
	"context"
	"errors"
	"iter"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// lifecycleProviders keeps agent construction offline: the OpenAI client is
// created, never called.
func lifecycleProviders() []agentsv1.ModelProvider {
	return []agentsv1.ModelProvider{
		{
			Name:   "openai",
			Type:   "openai",
			Models: []*agentsv1.ModelConfig{{Name: "m1"}},
		},
	}
}

// lifecycleMCPServer is an MCP server the ADK toolset only connects to when a
// turn lists tools, so building the service stays offline.
func lifecycleMCPServer() agentsv1.MCPServer {
	return agentsv1.MCPServer{
		Id:        "srv",
		Name:      "srv",
		Url:       "http://127.0.0.1:1/mcp",
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
	}
}

func lifecycleAgent(name string) agentsv1.Agent {
	return agentsv1.Agent{
		AgentId: name,
		Name:    name,
		Type:    agentsv1.AgentType_AGENT_TYPE_LLM,
		Config: &agentsv1.AgentConfig{
			Model:        "m1",
			McpServerIds: []string{"srv"},
		},
	}
}

func newLifecycleService(t *testing.T) *Service {
	t.Helper()
	svc, err := NewService(context.Background(),
		[]agentsv1.Agent{lifecycleAgent("mcp-agent")},
		lifecycleProviders(),
		[]agentsv1.MCPServer{lifecycleMCPServer()},
		nil, nil,
		session.InMemoryService(), nil, nil,
		adkrunner.PluginConfig{},
	)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	return svc
}

func resourcesOf(t *testing.T, svc *Service, name string) *internalagent.Resources {
	t.Helper()
	svc.mu.Lock()
	defer svc.mu.Unlock()
	res := svc.agentResources[name]
	if res == nil {
		t.Fatalf("agent %q owns no resources", name)
	}
	return res
}

func TestNewServiceOwnsAgentMCPResources(t *testing.T) {
	svc := newLifecycleService(t)
	res := resourcesOf(t, svc, "mcp-agent")

	release, ok := res.Acquire()
	if !ok {
		t.Fatal("fresh agent resources refused a run")
	}
	release()
}

func TestServiceCloseRetiresAgentMCPResources(t *testing.T) {
	svc := newLifecycleService(t)
	res := resourcesOf(t, svc, "mcp-agent")

	if err := svc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, ok := res.Acquire(); ok {
		t.Fatal("a closed service still handed out live agent resources")
	}
	svc.mu.Lock()
	_, remains := svc.agentResources["mcp-agent"]
	svc.mu.Unlock()
	if remains {
		t.Fatal("Close kept the agent's resource entry")
	}
}

func TestServiceCloseRetiresCachedOverrideResources(t *testing.T) {
	svc := newLifecycleService(t)
	override := &internalagent.Resources{}
	svc.mu.Lock()
	svc.runtimeCache.overriddenResources["mcp-agent:other-model"] = override
	svc.mu.Unlock()

	if err := svc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, ok := override.Acquire(); ok {
		t.Fatal("Close kept a cached model-overridden agent's resources open")
	}
}

func TestReloadReplacesAgentMCPResources(t *testing.T) {
	svc := newLifecycleService(t)
	old := resourcesOf(t, svc, "mcp-agent")

	// A cached model-overridden agent of the old generation must go too.
	override := &internalagent.Resources{}
	svc.mu.Lock()
	svc.runtimeCache.overriddenResources["mcp-agent:other-model"] = override
	svc.mu.Unlock()

	if err := svc.ReloadProtoAgents(context.Background(),
		[]agentsv1.Agent{lifecycleAgent("mcp-agent")},
		lifecycleProviders(),
		[]agentsv1.MCPServer{lifecycleMCPServer()},
		nil,
	); err != nil {
		t.Fatalf("ReloadProtoAgents: %v", err)
	}

	fresh := resourcesOf(t, svc, "mcp-agent")
	if fresh == old {
		t.Fatal("reload reused the replaced agent's resources")
	}
	if _, ok := old.Acquire(); ok {
		t.Fatal("reload kept the replaced agent's resources open")
	}
	if _, ok := override.Acquire(); ok {
		t.Fatal("reload kept a cached model-overridden agent's resources open")
	}
	if _, ok := fresh.Acquire(); !ok {
		t.Fatal("reload published already-retired resources")
	}
}

func TestReloadClosesResourcesOfRemovedAgents(t *testing.T) {
	svc := newLifecycleService(t)
	old := resourcesOf(t, svc, "mcp-agent")

	if err := svc.ReloadProtoAgents(context.Background(), nil, lifecycleProviders(), nil, nil); err != nil {
		t.Fatalf("ReloadProtoAgents: %v", err)
	}
	if _, ok := old.Acquire(); ok {
		t.Fatal("reload kept the resources of an agent it dropped")
	}
}

func TestRegisterAgentClosesReplacedResources(t *testing.T) {
	svc := newLifecycleService(t)
	old := resourcesOf(t, svc, "mcp-agent")

	svc.RegisterAgent("mcp-agent", nil)
	if _, ok := old.Acquire(); ok {
		t.Fatal("RegisterAgent kept the replaced agent's resources open")
	}
}

func TestRegisterAgentWithBuilderClosesReplacedResources(t *testing.T) {
	svc := newLifecycleService(t)
	old := resourcesOf(t, svc, "mcp-agent")

	svc.RegisterAgentWithBuilder("mcp-agent", nil, nil)
	if _, ok := old.Acquire(); ok {
		t.Fatal("RegisterAgentWithBuilder kept the replaced agent's resources open")
	}
}

func TestBuildOverriddenAgentSharesAndRetiresResources(t *testing.T) {
	svc := newLifecycleService(t)
	svc.mu.Lock()
	generation := svc.runtimeCache.generation
	svc.mu.Unlock()

	built, res, current, err := svc.buildOverriddenAgent(context.Background(), "mcp-agent", "m1", generation)
	if err != nil {
		t.Fatalf("buildOverriddenAgent: %v", err)
	}
	if !current {
		t.Fatal("buildOverriddenAgent reported a reload that never happened")
	}
	if built == nil || res == nil {
		t.Fatalf("buildOverriddenAgent returned agent %v, resources %v", built, res)
	}

	cached, cachedRes, current, err := svc.buildOverriddenAgent(context.Background(), "mcp-agent", "m1", generation)
	if err != nil || !current {
		t.Fatalf("cached buildOverriddenAgent = (%v, %v), want current", current, err)
	}
	if cachedRes != res {
		t.Fatal("the cached model-overridden agent came back with different resources")
	}
	if cached == nil {
		t.Fatal("the cached model-overridden agent is missing")
	}

	// The rebuild for a reload owns its own connections; the old ones retire.
	if err := svc.ReloadProtoAgents(context.Background(),
		[]agentsv1.Agent{lifecycleAgent("mcp-agent")},
		lifecycleProviders(),
		[]agentsv1.MCPServer{lifecycleMCPServer()},
		nil,
	); err != nil {
		t.Fatalf("ReloadProtoAgents: %v", err)
	}
	if _, ok := res.Acquire(); ok {
		t.Fatal("reload kept a model-overridden agent's resources open")
	}
}

// A builder agent — the built-in system agent, or anything registered through
// RegisterAgent / RegisterAgentWithBuilder — has no proto and therefore no
// Resources entry. Its nil owner cannot fail Acquire after Close, so shutdown
// must be refused by the registry lock itself, before any turn starts.
func TestRunAfterCloseRejectsNewTurns(t *testing.T) {
	svc, err := NewService(context.Background(),
		[]agentsv1.Agent{lifecycleAgent("mcp-agent")},
		lifecycleProviders(),
		[]agentsv1.MCPServer{lifecycleMCPServer()},
		nil, nil,
		session.InMemoryService(), nil, nil,
		adkrunner.PluginConfig{},
	)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	// This model answers any turn that reaches it, so a turn served after
	// shutdown produces output instead of the closed error and fails the test.
	base, err := newGenerationTestAgent("served-after-close")
	if err != nil {
		t.Fatalf("newGenerationTestAgent: %v", err)
	}
	svc.RegisterAgentWithBuilder("dynamic-agent", base, func(context.Context, string) (agent.Agent, error) {
		return newGenerationTestAgent("served-after-close")
	})

	svc.mu.Lock()
	builderResources := svc.agentResources["dynamic-agent"]
	svc.mu.Unlock()
	if builderResources != nil {
		t.Fatal("builder agent owns resources; this test no longer covers the nil-owner path")
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	for _, name := range []string{"mcp-agent", "dynamic-agent"} {
		ctxInfo := &agentsv1.ContextInfo{
			Uuid:        "closed-" + name,
			SessionId:   "closed-" + name,
			UserId:      "u1",
			ChannelName: "test-app",
		}
		if _, err := svc.Run(context.Background(), name, []*genai.Part{{Text: "hello"}}, "", ctxInfo, nil, nil); !errors.Is(err, errRunnerServiceClosed) {
			t.Fatalf("Run(%q) after Close = %v, want %v", name, err, errRunnerServiceClosed)
		}
		// A repeat turn is refused too: a caller retry loop must not spin or
		// revive the service.
		if _, err := svc.RunTurn(context.Background(), name, []*genai.Part{{Text: "again"}}, "", ctxInfo, nil, nil); !errors.Is(err, errRunnerServiceClosed) {
			t.Fatalf("RunTurn(%q) after Close = %v, want %v", name, err, errRunnerServiceClosed)
		}
	}
}

// blockedTestModel answers a turn only once the test releases it, so a turn can
// be held inside its invocation while shutdown runs.
type blockedTestModel struct {
	started  chan struct{}
	release  chan struct{}
	once     sync.Once
	reply    string
	finished atomic.Bool
}

func (m *blockedTestModel) Name() string { return "blocked-model" }

func (m *blockedTestModel) GenerateContent(_ context.Context, _ *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		m.once.Do(func() { close(m.started) })
		<-m.release
		m.finished.Store(true)
		yield(&model.LLMResponse{
			Content:      genai.NewContentFromText(m.reply, genai.RoleModel),
			FinishReason: genai.FinishReasonStop,
		}, nil)
	}
}

// A turn whose builder lookup is in flight when Close lands must be refused
// when it comes back, not admitted against a registry that no longer exists.
func TestRunBlockedInBuilderLookupIsRefusedWhenCloseLands(t *testing.T) {
	svc, err := NewService(context.Background(), nil, lifecycleProviders(), nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	base, err := newGenerationTestAgent("served-after-close")
	if err != nil {
		t.Fatalf("newGenerationTestAgent: %v", err)
	}
	buildStarted := make(chan struct{})
	releaseBuild := make(chan struct{})
	var once sync.Once
	svc.RegisterAgentWithBuilder("dynamic-agent", base, func(context.Context, string) (agent.Agent, error) {
		once.Do(func() { close(buildStarted) })
		<-releaseBuild
		return newGenerationTestAgent("served-after-close")
	})

	done := make(chan error, 1)
	go func() {
		ctxInfo := &agentsv1.ContextInfo{Uuid: "close-race", SessionId: "close-race", UserId: "u1", ChannelName: "test-app"}
		_, runErr := svc.Run(context.Background(), "dynamic-agent", []*genai.Part{{Text: "hello"}}, "override-model", ctxInfo, nil, nil)
		done <- runErr
	}()

	select {
	case <-buildStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("builder lookup did not start")
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	close(releaseBuild)

	select {
	case runErr := <-done:
		if !errors.Is(runErr, errRunnerServiceClosed) {
			t.Fatalf("turn whose builder lookup overlapped Close = %v, want %v", runErr, errRunnerServiceClosed)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("turn did not finish after Close")
	}
}

// A turn admitted before Close keeps its lease and finishes normally: shutdown
// retires the lease but does not yank it away from work already in progress.
func TestRunAdmittedBeforeCloseCompletes(t *testing.T) {
	svc, err := NewService(context.Background(), nil, lifecycleProviders(), nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	generated := &blockedTestModel{
		started: make(chan struct{}),
		release: make(chan struct{}),
		reply:   "finished after close",
	}
	ag, err := llmagent.New(llmagent.Config{Name: "builder-agent", Model: generated})
	if err != nil {
		t.Fatalf("llmagent.New: %v", err)
	}
	svc.RegisterAgentWithBuilder("builder-agent", ag, nil)

	type runResult struct {
		output string
		err    error
	}
	done := make(chan runResult, 1)
	go func() {
		ctxInfo := &agentsv1.ContextInfo{Uuid: "admitted", SessionId: "admitted", UserId: "u1", ChannelName: "test-app"}
		output, runErr := svc.Run(context.Background(), "builder-agent", []*genai.Part{{Text: "hello"}}, "", ctxInfo, nil, nil)
		done <- runResult{output: output, err: runErr}
	}()

	select {
	case <-generated.started:
	case <-time.After(5 * time.Second):
		t.Fatal("turn never reached the model")
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("Close during a running turn: %v", err)
	}
	close(generated.release)

	select {
	case result := <-done:
		if result.err != nil {
			t.Fatalf("admitted turn failed after Close: %v", result.err)
		}
		if result.output != "finished after close" {
			t.Fatalf("admitted turn output = %q, want %q", result.output, "finished after close")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("admitted turn did not finish after Close")
	}
}

// The exact window the loop-top check cannot cover: Close lands after a turn
// read the registry but before it is admitted, with no generation change to
// invalidate the lookup. Admission is refused on s.closed alone.
func TestAdmissionRefusesTurnClosedWithoutGenerationChange(t *testing.T) {
	svc := newLifecycleService(t)

	svc.mu.Lock()
	generation := svc.runtimeCache.generation
	svc.closed = true
	svc.mu.Unlock()

	if svc.runtimeGenerationCurrent(generation) != true {
		t.Fatalf("test setup no longer models a Close without a generation change")
	}
	if release, ok := svc.admitTurn(); ok {
		release()
		t.Fatal("a turn was admitted after the service closed")
	}
}

// The pre-admission window, end to end: a turn has finished every unlocked
// lookup and generation check, and Close lands before it is admitted. Close
// retires the admission lease under s.mu, so the turn must be refused there —
// for a builder agent the agent-owned resources are nil and cannot refuse it,
// and its generation snapshot predates the shutdown, so nothing else can.
//
// The shutdown transition is applied directly instead of through Close() so the
// interleaving is deterministic: a real Close also resets the runtime
// generation, which would hide the admission decision behind a stale-snapshot
// retry. The state modelled is the one Close publishes under s.mu.
func TestRunInFlightBeforeAdmissionIsRefusedWhenCloseLands(t *testing.T) {
	svc, err := NewService(context.Background(), nil, lifecycleProviders(), nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	base, err := newGenerationTestAgent("served-after-close")
	if err != nil {
		t.Fatalf("newGenerationTestAgent: %v", err)
	}
	buildStarted := make(chan struct{})
	releaseBuild := make(chan struct{})
	var once sync.Once
	svc.RegisterAgentWithBuilder("dynamic-agent", base, func(context.Context, string) (agent.Agent, error) {
		once.Do(func() { close(buildStarted) })
		<-releaseBuild
		return newGenerationTestAgent("served-after-close")
	})

	type runResult struct {
		output string
		err    error
	}
	done := make(chan runResult, 1)
	go func() {
		ctxInfo := &agentsv1.ContextInfo{Uuid: "pre-admission", SessionId: "pre-admission", UserId: "u1", ChannelName: "test-app"}
		output, runErr := svc.Run(context.Background(), "dynamic-agent", []*genai.Part{{Text: "hello"}}, "override-model", ctxInfo, nil, nil)
		done <- runResult{output: output, err: runErr}
	}()

	select {
	case <-buildStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("builder lookup did not start")
	}

	// Close's side of the lock: no new turn may be admitted from here on.
	svc.mu.Lock()
	svc.closed = true
	_ = svc.admissionLocked().Close()
	svc.mu.Unlock()

	close(releaseBuild)

	select {
	case result := <-done:
		if !errors.Is(result.err, errRunnerServiceClosed) {
			t.Fatalf("turn admitted across Close = (%q, %v), want %v", result.output, result.err, errRunnerServiceClosed)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("turn did not finish after Close")
	}
}
