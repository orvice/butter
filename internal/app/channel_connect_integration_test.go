package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go.orx.me/apps/butter/internal/config"
	configmongo "go.orx.me/apps/butter/internal/repo/config/mongo"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
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

type connectIntegrationFixture struct {
	ctx     context.Context
	cfg     *config.AppConfig
	server  *httptest.Server
	db      *mongo.Database
	tracker *runtimeTracker
}

func (fx *connectIntegrationFixture) baseURL() string { return fx.server.URL + "/api" }

func (fx *connectIntegrationFixture) httpClient() *http.Client { return fx.server.Client() }

// workspaceHeaderInterceptor injects the X-Workspace-ID header onto every
// outbound Connect request so the test fixture's workspace context survives
// the round-trip through the gin AuthMiddleware.
func workspaceHeaderInterceptor(workspaceID string) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set(workspace.HeaderName, workspaceID)
			return next(ctx, req)
		}
	})
}

func (fx *connectIntegrationFixture) clientOptions() []connect.ClientOption {
	return []connect.ClientOption{connect.WithInterceptors(workspaceHeaderInterceptor("ws-test"))}
}

func newConnectIntegrationFixture(t *testing.T) *connectIntegrationFixture {
	t.Helper()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		t.Skip("MONGO_URI is required for integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	ctx = workspace.WithID(ctx, "ws-test")

	cfg := &config.AppConfig{
		StorageBackend: "mongo",
		MongoURI:       mongoURI,
		MongoDB:        "butter_connect_" + uuid.NewString(),
		Auth:           config.AuthConfig{AllowUnauthenticated: true},
	}

	routerFn, handlers := SetupRoutes(cfg, daemon.NewRegistry())
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

	return &connectIntegrationFixture{
		ctx:     ctx,
		cfg:     cfg,
		server:  server,
		db:      db,
		tracker: tracker,
	}
}

func TestChannelService_ConnectIntegration(t *testing.T) {
	fx := newConnectIntegrationFixture(t)

	client := agentsv1connect.NewChannelServiceClient(fx.httpClient(), fx.baseURL(), fx.clientOptions()...)

	createResp, err := client.CreateChannel(fx.ctx, connect.NewRequest(&agentsv1.CreateChannelRequest{
		Channel: &agentsv1.AgentChannel{
			Name:      "telegram-main",
			AgentName: "agent-alpha",
			Platform:  agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
			Enabled:   true,
			Telegram:  &agentsv1.TelegramChannelConfig{BotToken: "123456:integration-token"},
		},
	}))
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if createResp.Msg.GetChannel().GetName() != "telegram-main" {
		t.Fatalf("expected created channel telegram-main, got %q", createResp.Msg.GetChannel().GetName())
	}
	if fx.tracker.channelReloads != 1 {
		t.Fatalf("expected 1 channel reload, got %d", fx.tracker.channelReloads)
	}

	updateResp, err := client.UpdateChannel(fx.ctx, connect.NewRequest(&agentsv1.UpdateChannelRequest{
		Channel: &agentsv1.AgentChannel{
			Name:      "telegram-main",
			AgentName: "agent-beta",
			Platform:  agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
			Enabled:   true,
			Telegram:  &agentsv1.TelegramChannelConfig{BotToken: "123456:integration-token"},
		},
	}))
	if err != nil {
		t.Fatalf("update channel: %v", err)
	}
	if updateResp.Msg.GetChannel().GetAgentName() != "agent-beta" {
		t.Fatalf("expected updated agent-beta, got %q", updateResp.Msg.GetChannel().GetAgentName())
	}
	if fx.tracker.channelReloads != 2 {
		t.Fatalf("expected 2 channel reloads, got %d", fx.tracker.channelReloads)
	}

	getResp, err := client.GetChannel(fx.ctx, connect.NewRequest(&agentsv1.GetChannelRequest{Name: "telegram-main"}))
	if err != nil {
		t.Fatalf("get channel: %v", err)
	}
	if getResp.Msg.GetChannel().GetAgentName() != "agent-beta" {
		t.Fatalf("expected persisted agent-beta, got %q", getResp.Msg.GetChannel().GetAgentName())
	}

	repo := configmongo.New(fx.db)
	channels, err := repo.ListChannels(fx.ctx, "ws-test")
	if err != nil {
		t.Fatalf("list persisted channels: %v", err)
	}
	if len(channels) != 1 || channels[0].GetAgentName() != "agent-beta" {
		t.Fatalf("expected persisted updated channel, got %+v", channels)
	}

	if _, err := client.DeleteChannel(fx.ctx, connect.NewRequest(&agentsv1.DeleteChannelRequest{Name: "telegram-main"})); err != nil {
		t.Fatalf("delete channel: %v", err)
	}
	if fx.tracker.channelReloads != 3 {
		t.Fatalf("expected 3 channel reloads, got %d", fx.tracker.channelReloads)
	}

	_, err = client.GetChannel(fx.ctx, connect.NewRequest(&agentsv1.GetChannelRequest{Name: "telegram-main"}))
	if cerr, ok := err.(*connect.Error); !ok || cerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound after delete, got %v", err)
	}

	channels, err = repo.ListChannels(fx.ctx, "ws-test")
	if err != nil {
		t.Fatalf("list channels after delete: %v", err)
	}
	if len(channels) != 0 {
		t.Fatalf("expected 0 persisted channels after delete, got %d", len(channels))
	}
}

func TestConfigServices_ConnectIntegration(t *testing.T) {
	fx := newConnectIntegrationFixture(t)
	opts := fx.clientOptions()

	mcpClient := agentsv1connect.NewMCPServerServiceClient(fx.httpClient(), fx.baseURL(), opts...)
	remoteClient := agentsv1connect.NewRemoteAgentServiceClient(fx.httpClient(), fx.baseURL(), opts...)
	agentClient := agentsv1connect.NewAgentServiceClient(fx.httpClient(), fx.baseURL(), opts...)

	if _, err := mcpClient.CreateMCPServer(fx.ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{
			Id:        "mcp-1",
			Name:      "primary-mcp",
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
			Url:       "http://127.0.0.1:8099/mcp",
		},
	})); err != nil {
		t.Fatalf("create mcp server: %v", err)
	}
	if fx.tracker.runnerReloads != 1 {
		t.Fatalf("expected 1 runner reload after mcp create, got %d", fx.tracker.runnerReloads)
	}

	if _, err := remoteClient.CreateRemoteAgent(fx.ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{
			Id:       "remote-1",
			Name:     "remote-agent",
			Url:      "http://127.0.0.1:8081/a2a/remote-agent/.well-known/agent.json",
			Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A,
		},
	})); err != nil {
		t.Fatalf("create remote agent: %v", err)
	}
	if fx.tracker.runnerReloads != 2 {
		t.Fatalf("expected 2 runner reloads after remote create, got %d", fx.tracker.runnerReloads)
	}

	createResp, err := agentClient.CreateAgent(fx.ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: "workflow-agent",
			Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
			Config: &agentsv1.AgentConfig{
				McpServerIds:   []string{"mcp-1"},
				RemoteAgentIds: []string{"remote-1"},
			},
		},
	}))
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if createResp.Msg.GetAgent().GetName() != "workflow-agent" {
		t.Fatalf("expected created agent workflow-agent, got %q", createResp.Msg.GetAgent().GetName())
	}
	if fx.tracker.runnerReloads != 3 {
		t.Fatalf("expected 3 runner reloads after agent create, got %d", fx.tracker.runnerReloads)
	}

	getResp, err := agentClient.GetAgent(fx.ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "workflow-agent"}))
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if getResp.Msg.GetAgent().GetConfig().GetMcpServerIds()[0] != "mcp-1" {
		t.Fatalf("expected persisted mcp reference mcp-1, got %+v", getResp.Msg.GetAgent().GetConfig().GetMcpServerIds())
	}

	repo := configmongo.New(fx.db)
	agents, err := repo.ListAgents(fx.ctx, "ws-test")
	if err != nil {
		t.Fatalf("list persisted agents: %v", err)
	}
	if len(agents) != 1 || agents[0].GetName() != "workflow-agent" {
		t.Fatalf("expected persisted workflow-agent, got %+v", agents)
	}

	if _, err := agentClient.DeleteAgent(fx.ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{Name: "workflow-agent"})); err != nil {
		t.Fatalf("delete agent: %v", err)
	}
	if fx.tracker.runnerReloads != 4 {
		t.Fatalf("expected 4 runner reloads after agent delete, got %d", fx.tracker.runnerReloads)
	}

	_, err = agentClient.GetAgent(fx.ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "workflow-agent"}))
	if cerr, ok := err.(*connect.Error); !ok || cerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound after agent delete, got %v", err)
	}
}
