package app

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	adkrunner "google.golang.org/adk/runner"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	configmongo "go.orx.me/apps/butter/internal/repo/config/mongo"
	mongomemory "go.orx.me/apps/butter/internal/runtime/memory/mongo"
	"go.orx.me/apps/butter/internal/runtime/runner"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestMongoBackedConfigRuntimeIntegration(t *testing.T) {
	mongoURI := os.Getenv("MONGO_URI")
	redisAddr := os.Getenv("REDIS_ADDR")
	if mongoURI == "" || redisAddr == "" {
		t.Skip("MONGO_URI and REDIS_ADDR are required for integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	runnerSvc, err := runner.NewService(ctx, cfg.Agents, cfg.ModelProviders, cfg.MCPServerConfigs, cfg.RemoteAgents, sessionSvc, memorySvc, adkrunner.PluginConfig{})
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

	channelMgr, err := channel.NewManager(ctx, configStore, runnerSvc, rdb, nil)
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
	channelSvc := application.NewChannelServiceServer(configStore)
	channelSvc.SetRuntime(configRuntime)

	_, err = mcpSvc.CreateMCPServer(ctx, &agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{
			Id:        "mcp-1",
			Name:      "primary-mcp",
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
			Url:       "http://127.0.0.1:8099/mcp",
		},
	})
	if err != nil {
		t.Fatalf("create mcp server: %v", err)
	}
	if len(cfg.MCPServerConfigs) != 1 {
		t.Fatalf("expected 1 mcp config, got %d", len(cfg.MCPServerConfigs))
	}

	_, err = remoteSvc.CreateRemoteAgent(ctx, &agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{
			Id:       "remote-1",
			Name:     "remote-agent",
			Url:      "http://127.0.0.1:8081/a2a/remote-agent/.well-known/agent.json",
			Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A,
		},
	})
	if err != nil {
		t.Fatalf("create remote agent: %v", err)
	}
	if len(cfg.RemoteAgents) != 1 {
		t.Fatalf("expected 1 remote agent config, got %d", len(cfg.RemoteAgents))
	}

	_, err = agentSvc.CreateAgent(ctx, &agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "workflow-agent",
			Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
			Config: &agentsv1.AgentConfig{
				McpServerIds:   []string{"mcp-1"},
				RemoteAgentIds: []string{"remote-1"},
			},
		},
	})
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

	_, err = channelSvc.CreateChannel(ctx, &agentsv1.CreateChannelRequest{
		Channel: &agentsv1.AgentChannel{
			Name:      "telegram-main",
			AgentName: "workflow-agent",
			Platform:  agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
			Enabled:   true,
			Telegram:  &agentsv1.TelegramChannelConfig{BotToken: "123456:integration-token"},
		},
	})
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if len(cfg.Channels) != 1 {
		t.Fatalf("expected 1 channel config, got %d", len(cfg.Channels))
	}

	channelStore := configmongo.New(db)
	channels, err := channelStore.ListChannels(ctx)
	if err != nil {
		t.Fatalf("list persisted channels: %v", err)
	}
	if len(channels) != 1 || channels[0].GetName() != "telegram-main" {
		t.Fatalf("expected persisted channel telegram-main, got %+v", channels)
	}

	_, err = agentSvc.DeleteAgent(ctx, &agentsv1.DeleteAgentRequest{Name: "workflow-agent"})
	if err != nil {
		t.Fatalf("delete agent: %v", err)
	}
	if runnerSvc.HasAgent("workflow-agent") {
		t.Fatal("expected workflow-agent to be removed from runner service after reload")
	}
}
