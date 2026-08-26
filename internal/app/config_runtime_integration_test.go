package app

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	mongomemory "go.orx.me/apps/butter/internal/runtime/memory/mongo"
	"go.orx.me/apps/butter/internal/runtime/runner"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestConfigRuntimeReloadRunnerUpdatesWarmedEffectiveContextWindows(t *testing.T) {
	const marker = "[System: The conversation was compacted because it exceeded the context window."
	backend := openaifake.New(t)
	ctx := workspace.WithID(context.Background(), "ws-test")
	store := NewConfigStore()

	agentConfig := &agentsv1.Agent{
		Name:        "config-runtime-guarded",
		AgentId:     "config-runtime-guarded",
		Type:        agentsv1.AgentType_AGENT_TYPE_LLM,
		WorkspaceId: "ws-test",
		Config: &agentsv1.AgentConfig{
			Model: "default-alias",
			ContextGuard: &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
			},
		},
	}
	if _, err := store.CreateAgent(ctx, "ws-test", agentConfig); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	provider := &agentsv1.ModelProvider{
		Name:    "fake",
		Type:    "openai",
		BaseUrl: backend.URL(),
		Models: []*agentsv1.ModelConfig{
			{Name: "actual-default", Alias: "default-alias", ContextWindowTokens: 128_000},
			{Name: "actual-override-v1", Alias: "override-alias", ContextWindowTokens: 128_000},
		},
	}
	if _, err := store.CreateModelProvider(ctx, "ws-test", provider); err != nil {
		t.Fatalf("CreateModelProvider: %v", err)
	}

	cfg := &config.AppConfig{}
	if err := store.SyncToConfig(ctx, cfg); err != nil {
		t.Fatalf("SyncToConfig: %v", err)
	}
	sessions := session.InMemoryService()
	runnerSvc, err := runner.NewService(ctx, cfg.Agents, cfg.ModelProviders, nil, nil, nil, sessions, nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	runtime := NewConfigRuntime(store, cfg)
	runtime.SetRunnerService(runnerSvc)

	input := strings.Repeat("ConfigRuntime must reload this effective context window. ", 20)
	defaultCtx := configRuntimeContextInfo("config-runtime-default")
	overrideCtx := configRuntimeContextInfo("config-runtime-override")
	if _, err := runnerSvc.Run(ctx, agentConfig.GetName(), []*genai.Part{{Text: input}}, "", defaultCtx, nil, nil); err != nil {
		t.Fatalf("warm default Run: %v", err)
	}
	if _, err := runnerSvc.Run(ctx, agentConfig.GetName(), []*genai.Part{{Text: input}}, "override-alias", overrideCtx, nil, nil); err != nil {
		t.Fatalf("warm override Run: %v", err)
	}
	if got := backend.CallCount("actual-default"); got != 1 {
		t.Fatalf("default calls before reload = %d, want 1", got)
	}
	if got := backend.CallCount("actual-override-v1"); got != 1 {
		t.Fatalf("override calls before reload = %d, want 1", got)
	}

	defaultBefore, err := runnerSvc.GetSession(ctx, defaultCtx.GetChannelName(), defaultCtx.GetSessionId(), defaultCtx.GetUserId())
	if err != nil {
		t.Fatalf("GetSession default before reload: %v", err)
	}
	overrideBefore, err := runnerSvc.GetSession(ctx, overrideCtx.GetChannelName(), overrideCtx.GetSessionId(), overrideCtx.GetUserId())
	if err != nil {
		t.Fatalf("GetSession override before reload: %v", err)
	}
	defaultEvents := defaultBefore.Events().Len()
	overrideEvents := overrideBefore.Events().Len()

	updated := &agentsv1.ModelProvider{
		Name:    provider.GetName(),
		Type:    provider.GetType(),
		BaseUrl: provider.GetBaseUrl(),
		Models: []*agentsv1.ModelConfig{
			{Name: "actual-default", Alias: "default-alias", ContextWindowTokens: 128},
			{Name: "actual-override-v2", Alias: "override-alias", ContextWindowTokens: 128},
		},
	}
	if _, err := store.UpdateModelProvider(ctx, "ws-test", updated); err != nil {
		t.Fatalf("UpdateModelProvider: %v", err)
	}
	if err := runtime.ReloadRunner(ctx); err != nil {
		t.Fatalf("ConfigRuntime.ReloadRunner: %v", err)
	}
	if got := cfg.ModelProviders[0].GetModels()[0].GetContextWindowTokens(); got != 128 {
		t.Fatalf("AppConfig capacity after reload = %d, want 128", got)
	}

	defaultAfterReload, err := runnerSvc.GetSession(ctx, defaultCtx.GetChannelName(), defaultCtx.GetSessionId(), defaultCtx.GetUserId())
	if err != nil {
		t.Fatalf("GetSession default after reload: %v", err)
	}
	overrideAfterReload, err := runnerSvc.GetSession(ctx, overrideCtx.GetChannelName(), overrideCtx.GetSessionId(), overrideCtx.GetUserId())
	if err != nil {
		t.Fatalf("GetSession override after reload: %v", err)
	}
	if got := defaultAfterReload.Events().Len(); got != defaultEvents {
		t.Fatalf("default session events after ConfigRuntime reload = %d, want %d", got, defaultEvents)
	}
	if got := overrideAfterReload.Events().Len(); got != overrideEvents {
		t.Fatalf("override session events after ConfigRuntime reload = %d, want %d", got, overrideEvents)
	}

	if _, err := runnerSvc.Run(ctx, agentConfig.GetName(), []*genai.Part{{Text: input}}, "", defaultCtx, nil, nil); err != nil {
		t.Fatalf("first default Run after reload: %v", err)
	}
	if got := backend.LastInput("actual-default"); !strings.Contains(got, marker) {
		t.Fatalf("first default request after ConfigRuntime reload did not compact: %q", got)
	}
	if _, err := runnerSvc.Run(ctx, agentConfig.GetName(), []*genai.Part{{Text: input}}, "override-alias", overrideCtx, nil, nil); err != nil {
		t.Fatalf("first override Run after reload: %v", err)
	}
	if got := backend.LastInput("actual-override-v2"); !strings.Contains(got, marker) {
		t.Fatalf("first override request after ConfigRuntime reload did not use the rebuilt Agent and compact: %q", got)
	}
	if got := backend.CallCount("actual-override-v1"); got != 1 {
		t.Fatalf("stale override model calls after ConfigRuntime reload = %d, want the warmed v1 Agent to remain unused", got)
	}
}

func configRuntimeContextInfo(sessionID string) *agentsv1.ContextInfo {
	return &agentsv1.ContextInfo{
		Uuid:        sessionID,
		SessionId:   sessionID,
		UserId:      "u1",
		ChannelName: "config-runtime-app",
		WorkspaceId: "ws-test",
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
	}
}

func TestMongoBackedConfigRuntimeIntegration(t *testing.T) {
	mongoURI := os.Getenv("MONGO_URI")
	redisAddr := os.Getenv("REDIS_ADDR")
	if mongoURI == "" || redisAddr == "" {
		t.Skip("MONGO_URI and REDIS_ADDR are required for integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = workspace.WithID(ctx, "ws-test")

	cfg := &config.AppConfig{
		StorageBackend: "mongo",
		MongoURI:       mongoURI,
		MongoDB:        "butter_integration_" + uuid.NewString(),
		RedisAddr:      redisAddr,
	}

	configStore := NewConfigStore()
	if err := configStore.InitFromConfig(ctx, cfg); err != nil {
		t.Fatalf("init config store: %v", err)
	}

	db, err := connectMongo(ctx, cfg)
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
	})

	sessionSvc, err := mongosession.New(ctx, db)
	if err != nil {
		t.Fatalf("new session service: %v", err)
	}
	memorySvc, err := mongomemory.New(ctx, db)
	if err != nil {
		t.Fatalf("new memory service: %v", err)
	}

	runnerSvc, err := runner.NewService(ctx, cfg.Agents, cfg.ModelProviders, cfg.MCPServerConfigs, cfg.RemoteAgents, nil, sessionSvc, memorySvc, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("new runner service: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	t.Cleanup(func() {
		_ = rdb.FlushDB(context.Background()).Err()
		_ = rdb.Close()
	})

	channelMgr, err := channel.NewManager(ctx, configStore)
	if err != nil {
		t.Fatalf("new channel manager: %v", err)
	}

	configRuntime := NewConfigRuntime(configStore, cfg)
	configRuntime.SetRunnerService(runnerSvc)
	configRuntime.SetChannelManager(channelMgr)

	mcpSvc := application.NewMCPServerServiceServer(configStore)
	mcpSvc.SetRuntime(configRuntime)
	remoteSvc := application.NewRemoteAgentServiceServer(configStore)
	remoteSvc.SetRuntime(configRuntime)
	agentSvc := application.NewAgentServiceServer(configStore)
	agentSvc.SetRuntime(configRuntime)

	_, err = mcpSvc.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{
			Id:        "mcp-1",
			Name:      "primary-mcp",
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
			Url:       "http://127.0.0.1:8099/mcp",
		},
	}))
	if err != nil {
		t.Fatalf("create mcp server: %v", err)
	}
	if len(cfg.MCPServerConfigs) != 1 {
		t.Fatalf("expected 1 mcp config, got %d", len(cfg.MCPServerConfigs))
	}

	_, err = remoteSvc.CreateRemoteAgent(ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{
			Id:       "remote-1",
			Name:     "remote-agent",
			Url:      "http://127.0.0.1:8081/a2a/remote-agent/.well-known/agent.json",
			Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A,
		},
	}))
	if err != nil {
		t.Fatalf("create remote agent: %v", err)
	}
	if len(cfg.RemoteAgents) != 1 {
		t.Fatalf("expected 1 remote agent config, got %d", len(cfg.RemoteAgents))
	}

	_, err = agentSvc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name:    "workflow-agent",
			AgentId: "workflow-agent",
			Type:    agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
			Config: &agentsv1.AgentConfig{
				McpServerIds:   []string{"mcp-1"},
				RemoteAgentIds: []string{"remote-1"},
			},
		},
	}))
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if !runnerSvc.HasAgent("workflow-agent") {
		t.Fatal("expected runner service to include workflow-agent after reload")
	}

	status := runnerSvc.GetAgentStatus("workflow-agent")
	if status == nil {
		t.Fatal("expected workflow-agent status")
	}
	if len(status.MCPServers) != 1 || status.MCPServers[0] != "primary-mcp" {
		t.Fatalf("expected resolved MCP server in agent status, got %+v", status.MCPServers)
	}

	_, err = agentSvc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{AgentId: "workflow-agent"}))
	if err != nil {
		t.Fatalf("delete agent: %v", err)
	}
	if runnerSvc.HasAgent("workflow-agent") {
		t.Fatal("expected workflow-agent to be removed from runner service after reload")
	}
}
