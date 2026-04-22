package app

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/twitchtv/twirp"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go.orx.me/apps/butter/internal/config"
	configmongo "go.orx.me/apps/butter/internal/repo/config/mongo"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type runtimeTracker struct {
	runnerReloads  int
	channelReloads int
}

func (r *runtimeTracker) ReloadRunner(context.Context) error {
	r.runnerReloads++
	return nil
}

func (r *runtimeTracker) ReloadChannels(context.Context) error {
	r.channelReloads++
	return nil
}

type twirpIntegrationFixture struct {
	ctx     context.Context
	cfg     *config.AppConfig
	server  *httptest.Server
	db      *mongo.Database
	tracker *runtimeTracker
}

func newTwirpIntegrationFixture(t *testing.T) *twirpIntegrationFixture {
	t.Helper()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		t.Skip("MONGO_URI is required for integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	cfg := &config.AppConfig{
		StorageBackend: "mongo",
		MongoURI:       mongoURI,
		MongoDB:        "butter_twirp_" + uuid.NewString(),
	}

	routerFn, handlers := SetupRoutes(cfg)
	if err := handlers.SeedConfig(ctx, cfg); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	tracker := &runtimeTracker{}
	handlers.agentSvcServer.SetRuntime(tracker)
	handlers.mcpSvcServer.SetRuntime(tracker)
	handlers.remoteSvcServer.SetRuntime(tracker)
	handlers.channelSvcServer.SetRuntime(tracker)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	routerFn(engine)
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)

	db, err := connectMongo(ctx, cfg)
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
	})

	return &twirpIntegrationFixture{
		ctx:     ctx,
		cfg:     cfg,
		server:  server,
		db:      db,
		tracker: tracker,
	}
}

func TestChannelServiceTwirpIntegration(t *testing.T) {
	fx := newTwirpIntegrationFixture(t)

	client := agentsv1.NewChannelServiceProtobufClient(fx.server.URL, fx.server.Client(), twirp.WithClientPathPrefix("/api"))

	createResp, err := client.CreateChannel(fx.ctx, &agentsv1.CreateChannelRequest{
		Channel: &agentsv1.AgentChannel{
			Name:      "telegram-main",
			AgentName: "agent-alpha",
			Platform:  agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
			Enabled:   true,
			Telegram:  &agentsv1.TelegramChannelConfig{BotToken: "123456:integration-token"},
		},
	})
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if createResp.GetChannel().GetName() != "telegram-main" {
		t.Fatalf("expected created channel telegram-main, got %q", createResp.GetChannel().GetName())
	}
	if fx.tracker.channelReloads != 1 {
		t.Fatalf("expected 1 channel reload, got %d", fx.tracker.channelReloads)
	}

	updateResp, err := client.UpdateChannel(fx.ctx, &agentsv1.UpdateChannelRequest{
		Channel: &agentsv1.AgentChannel{
			Name:      "telegram-main",
			AgentName: "agent-beta",
			Platform:  agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
			Enabled:   true,
			Telegram:  &agentsv1.TelegramChannelConfig{BotToken: "123456:integration-token"},
		},
	})
	if err != nil {
		t.Fatalf("update channel: %v", err)
	}
	if updateResp.GetChannel().GetAgentName() != "agent-beta" {
		t.Fatalf("expected updated agent-beta, got %q", updateResp.GetChannel().GetAgentName())
	}
	if fx.tracker.channelReloads != 2 {
		t.Fatalf("expected 2 channel reloads, got %d", fx.tracker.channelReloads)
	}

	getResp, err := client.GetChannel(fx.ctx, &agentsv1.GetChannelRequest{Name: "telegram-main"})
	if err != nil {
		t.Fatalf("get channel: %v", err)
	}
	if getResp.GetChannel().GetAgentName() != "agent-beta" {
		t.Fatalf("expected persisted agent-beta, got %q", getResp.GetChannel().GetAgentName())
	}

	repo := configmongo.New(fx.db)
	channels, err := repo.ListChannels(fx.ctx)
	if err != nil {
		t.Fatalf("list persisted channels: %v", err)
	}
	if len(channels) != 1 || channels[0].GetAgentName() != "agent-beta" {
		t.Fatalf("expected persisted updated channel, got %+v", channels)
	}

	if _, err := client.DeleteChannel(fx.ctx, &agentsv1.DeleteChannelRequest{Name: "telegram-main"}); err != nil {
		t.Fatalf("delete channel: %v", err)
	}
	if fx.tracker.channelReloads != 3 {
		t.Fatalf("expected 3 channel reloads, got %d", fx.tracker.channelReloads)
	}

	_, err = client.GetChannel(fx.ctx, &agentsv1.GetChannelRequest{Name: "telegram-main"})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.NotFound {
		t.Fatalf("expected NotFound after delete, got %v", err)
	}

	channels, err = repo.ListChannels(fx.ctx)
	if err != nil {
		t.Fatalf("list channels after delete: %v", err)
	}
	if len(channels) != 0 {
		t.Fatalf("expected 0 persisted channels after delete, got %d", len(channels))
	}
}

func TestConfigServicesTwirpIntegration(t *testing.T) {
	fx := newTwirpIntegrationFixture(t)

	mcpClient := agentsv1.NewMCPServerServiceProtobufClient(fx.server.URL, fx.server.Client(), twirp.WithClientPathPrefix("/api"))
	remoteClient := agentsv1.NewRemoteAgentServiceProtobufClient(fx.server.URL, fx.server.Client(), twirp.WithClientPathPrefix("/api"))
	agentClient := agentsv1.NewAgentServiceProtobufClient(fx.server.URL, fx.server.Client(), twirp.WithClientPathPrefix("/api"))

	if _, err := mcpClient.CreateMCPServer(fx.ctx, &agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{
			Id:        "mcp-1",
			Name:      "primary-mcp",
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
			Url:       "http://127.0.0.1:8099/mcp",
		},
	}); err != nil {
		t.Fatalf("create mcp server: %v", err)
	}
	if fx.tracker.runnerReloads != 1 {
		t.Fatalf("expected 1 runner reload after mcp create, got %d", fx.tracker.runnerReloads)
	}

	if _, err := remoteClient.CreateRemoteAgent(fx.ctx, &agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{
			Id:       "remote-1",
			Name:     "remote-agent",
			Url:      "http://127.0.0.1:8081/a2a/remote-agent/.well-known/agent.json",
			Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A,
		},
	}); err != nil {
		t.Fatalf("create remote agent: %v", err)
	}
	if fx.tracker.runnerReloads != 2 {
		t.Fatalf("expected 2 runner reloads after remote create, got %d", fx.tracker.runnerReloads)
	}

	createResp, err := agentClient.CreateAgent(fx.ctx, &agentsv1.CreateAgentRequest{
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
	if createResp.GetAgent().GetName() != "workflow-agent" {
		t.Fatalf("expected created agent workflow-agent, got %q", createResp.GetAgent().GetName())
	}
	if fx.tracker.runnerReloads != 3 {
		t.Fatalf("expected 3 runner reloads after agent create, got %d", fx.tracker.runnerReloads)
	}

	getResp, err := agentClient.GetAgent(fx.ctx, &agentsv1.GetAgentRequest{Name: "workflow-agent"})
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if getResp.GetAgent().GetConfig().GetMcpServerIds()[0] != "mcp-1" {
		t.Fatalf("expected persisted mcp reference mcp-1, got %+v", getResp.GetAgent().GetConfig().GetMcpServerIds())
	}

	repo := configmongo.New(fx.db)
	agents, err := repo.ListAgents(fx.ctx)
	if err != nil {
		t.Fatalf("list persisted agents: %v", err)
	}
	if len(agents) != 1 || agents[0].GetName() != "workflow-agent" {
		t.Fatalf("expected persisted workflow-agent, got %+v", agents)
	}

	if _, err := agentClient.DeleteAgent(fx.ctx, &agentsv1.DeleteAgentRequest{Name: "workflow-agent"}); err != nil {
		t.Fatalf("delete agent: %v", err)
	}
	if fx.tracker.runnerReloads != 4 {
		t.Fatalf("expected 4 runner reloads after agent delete, got %d", fx.tracker.runnerReloads)
	}

	_, err = agentClient.GetAgent(fx.ctx, &agentsv1.GetAgentRequest{Name: "workflow-agent"})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.NotFound {
		t.Fatalf("expected NotFound after agent delete, got %v", err)
	}
}
