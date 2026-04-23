package app

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/repo"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/service"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Handlers holds all HTTP/Twirp handlers that need post-bootstrap wiring.
type Handlers struct {
	a2aHandler       *httpHandler.A2AHandler
	agentSvcServer   *application.AgentServiceServer
	mcpSvcServer     *application.MCPServerServiceServer
	remoteSvcServer  *application.RemoteAgentServiceServer
	sessionSvcServer *application.SessionServiceServer
	cronSvcServer    *application.CronJobServiceServer
	channelSvcServer *application.ChannelServiceServer
	configStore      *ConfigStore
	configRuntime    *ConfigRuntime
	agentRepo        configrepo.AgentRepository
	mcpServerRepo    configrepo.MCPServerRepository
	remoteAgentRepo  configrepo.RemoteAgentRepository
	channelRepo      configrepo.ChannelRepository
}

// Wire connects the bootstrap result to the handlers.
func (h *Handlers) Wire(result *BootstrapResult) {
	if result == nil {
		return
	}
	if result.RunnerSvc != nil {
		h.a2aHandler.SetRunnerService(result.RunnerSvc)
		h.sessionSvcServer.SetRunnerService(result.RunnerSvc)
	}
	if result.SessionSvc != nil {
		h.sessionSvcServer.SetSessionService(result.SessionSvc)
	}
	if result.CronScheduler != nil {
		h.cronSvcServer.SetScheduler(result.CronScheduler)
	}
	if result.CronRepo != nil {
		h.cronSvcServer.SetExecutionRepo(result.CronRepo)
	}
	if h.configRuntime != nil {
		if result.RunnerSvc != nil {
			h.configRuntime.SetRunnerService(result.RunnerSvc)
		}
		if result.ChannelMgr != nil {
			h.configRuntime.SetChannelManager(result.ChannelMgr)
		}
		h.agentSvcServer.SetRuntime(h.configRuntime)
		h.mcpSvcServer.SetRuntime(h.configRuntime)
		h.remoteSvcServer.SetRuntime(h.configRuntime)
		h.channelSvcServer.SetRuntime(h.configRuntime)
	}
}

// AgentRepo returns the agent repository.
func (h *Handlers) AgentRepo() configrepo.AgentRepository {
	return h.agentRepo
}

// ChannelRepo returns the channel repository.
func (h *Handlers) ChannelRepo() configrepo.ChannelRepository {
	return h.channelRepo
}

// SeedConfig initializes and seeds the config repositories from AppConfig.
func (h *Handlers) SeedConfig(ctx context.Context, cfg *config.AppConfig) error {
	if h.configStore == nil {
		return nil
	}
	return h.configStore.InitFromConfig(ctx, cfg)
}

// SetupRoutes creates all handlers and returns a Gin router function plus
// the Handlers struct for post-bootstrap wiring.
func SetupRoutes(cfg *config.AppConfig) (func(r *gin.Engine), *Handlers) {
	healthRepo := repo.NewHealthRepository()
	healthService := service.NewHealthService(healthRepo, cfg)
	healthHandler := httpHandler.NewHealthHandler(healthService)
	a2aHandler := httpHandler.NewA2AHandler(cfg)

	configStore := NewConfigStore()
	configRuntime := NewConfigRuntime(configStore, cfg)

	pathPrefix := twirp.WithServerPathPrefix("/api")
	agentSvcServer := application.NewAgentServiceServer(configStore)
	mcpSvcServer := application.NewMCPServerServiceServer(configStore)
	remoteSvcServer := application.NewRemoteAgentServiceServer(configStore)
	agentTwirp := agentsv1.NewAgentServiceServer(agentSvcServer, pathPrefix)
	mcpTwirp := agentsv1.NewMCPServerServiceServer(mcpSvcServer, pathPrefix)
	remoteTwirp := agentsv1.NewRemoteAgentServiceServer(remoteSvcServer, pathPrefix)
	channelSvcServer := application.NewChannelServiceServer(configStore)
	channelTwirp := agentsv1.NewChannelServiceServer(channelSvcServer, pathPrefix)
	sessionSvcServer := application.NewSessionServiceServer()
	sessionTwirp := agentsv1.NewSessionServiceServer(sessionSvcServer, pathPrefix)
	cronSvcServer := application.NewCronJobServiceServer()
	cronTwirp := agentsv1.NewCronJobServiceServer(cronSvcServer, pathPrefix)

	router := func(r *gin.Engine) {
		r.Use(httpHandler.APITokenAuthMiddleware(cfg))
		healthHandler.Register(r)
		a2aHandler.Register(r)

		// Mount Twirp handlers under /api prefix
		r.Any(agentTwirp.PathPrefix()+"*path", gin.WrapH(agentTwirp))
		r.Any(mcpTwirp.PathPrefix()+"*path", gin.WrapH(mcpTwirp))
		r.Any(remoteTwirp.PathPrefix()+"*path", gin.WrapH(remoteTwirp))
		r.Any(channelTwirp.PathPrefix()+"*path", gin.WrapH(channelTwirp))
		r.Any(sessionTwirp.PathPrefix()+"*path", gin.WrapH(sessionTwirp))
		r.Any(cronTwirp.PathPrefix()+"*path", gin.WrapH(cronTwirp))
	}

	handlers := &Handlers{
		a2aHandler:       a2aHandler,
		agentSvcServer:   agentSvcServer,
		mcpSvcServer:     mcpSvcServer,
		remoteSvcServer:  remoteSvcServer,
		sessionSvcServer: sessionSvcServer,
		cronSvcServer:    cronSvcServer,
		channelSvcServer: channelSvcServer,
		configStore:      configStore,
		configRuntime:    configRuntime,
		agentRepo:        configStore,
		mcpServerRepo:    configStore,
		remoteAgentRepo:  configStore,
		channelRepo:      configStore,
	}

	return router, handlers
}
