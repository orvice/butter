package app

import (
	"github.com/gin-gonic/gin"
	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/repo"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/service"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Handlers holds all HTTP/Twirp handlers that need post-bootstrap wiring.
type Handlers struct {
	a2aHandler       *httpHandler.A2AHandler
	sessionSvcServer *application.SessionServiceServer
	cronSvcServer    *application.CronJobServiceServer
	configStore      *memory.Store
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
}

// AgentRepo returns the agent repository.
func (h *Handlers) AgentRepo() configrepo.AgentRepository {
	return h.agentRepo
}

// SeedConfig seeds the config repositories from YAML config.
func (h *Handlers) SeedConfig(cfg *config.AppConfig) {
	// For the memory backend, use the direct Seed method.
	if h.configStore != nil {
		h.configStore.Seed(nil, cfg.Agents, cfg.MCPServerConfigs, cfg.RemoteAgents, cfg.Channels)
	}
}

// SetupRoutes creates all handlers and returns a Gin router function plus
// the Handlers struct for post-bootstrap wiring.
func SetupRoutes(cfg *config.AppConfig) (func(r *gin.Engine), *Handlers) {
	healthRepo := repo.NewHealthRepository()
	healthService := service.NewHealthService(healthRepo, cfg)
	healthHandler := httpHandler.NewHealthHandler(healthService)
	a2aHandler := httpHandler.NewA2AHandler(cfg)

	memStore := memory.New()

	pathPrefix := twirp.WithServerPathPrefix("/api")
	agentTwirp := agentsv1.NewAgentServiceServer(application.NewAgentServiceServer(memStore), pathPrefix)
	mcpTwirp := agentsv1.NewMCPServerServiceServer(application.NewMCPServerServiceServer(memStore), pathPrefix)
	remoteTwirp := agentsv1.NewRemoteAgentServiceServer(application.NewRemoteAgentServiceServer(memStore), pathPrefix)
	channelTwirp := agentsv1.NewChannelServiceServer(application.NewChannelServiceServer(memStore), pathPrefix)
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
		sessionSvcServer: sessionSvcServer,
		cronSvcServer:    cronSvcServer,
		configStore:      memStore,
		agentRepo:        memStore,
		mcpServerRepo:    memStore,
		remoteAgentRepo:  memStore,
		channelRepo:      memStore,
	}

	return router, handlers
}
