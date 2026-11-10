package runner

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/adk/v2/agent"
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
