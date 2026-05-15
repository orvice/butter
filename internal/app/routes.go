package app

import (
	"context"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/internal/service"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Handlers holds all HTTP/Twirp handlers that need post-bootstrap wiring.
type Handlers struct {
	a2aHandler         *httpHandler.A2AHandler
	agentSvcServer     *application.AgentServiceServer
	mcpSvcServer       *application.MCPServerServiceServer
	remoteSvcServer    *application.RemoteAgentServiceServer
	sessionSvcServer   *application.SessionServiceServer
	cronSvcServer      *application.CronJobServiceServer
	channelSvcServer   *application.ChannelServiceServer
	dashboardSvcServer *application.DashboardServiceServer
	daemonSvcServer    *application.DaemonServiceServer
	apiTokenSvcServer  *application.APITokenServiceServer
	apiTokenRepo       atomic.Value // apitoken.Repository
	configStore        *ConfigStore
	configRuntime      *ConfigRuntime
	agentRepo          configrepo.AgentRepository
	mcpServerRepo      configrepo.MCPServerRepository
	remoteAgentRepo    configrepo.RemoteAgentRepository
	channelRepo        configrepo.ChannelRepository
}

// apiTokenRepoFromHolder returns the currently wired apitoken repository, if any.
func (h *Handlers) apiTokenRepoFromHolder() apitoken.Repository {
	if h == nil {
		return nil
	}
	v := h.apiTokenRepo.Load()
	if v == nil {
		return nil
	}
	repo, _ := v.(apitoken.Repository)
	return repo
}

// Wire connects the bootstrap result to the handlers.
func (h *Handlers) Wire(result *BootstrapResult) {
	if result == nil {
		return
	}
	if result.RunnerSvc != nil {
		h.a2aHandler.SetRunnerService(result.RunnerSvc)
		h.sessionSvcServer.SetRunnerService(result.RunnerSvc)
		h.agentSvcServer.SetRunnerService(result.RunnerSvc)
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
	if result.ChannelMgr != nil {
		h.channelSvcServer.SetChannelManager(result.ChannelMgr)
	}
	if result.APITokenRepo != nil {
		h.apiTokenRepo.Store(result.APITokenRepo)
		if h.apiTokenSvcServer != nil {
			h.apiTokenSvcServer.SetRepo(result.APITokenRepo)
		}
	}
	if h.dashboardSvcServer != nil {
		if result.MongoDB != nil {
			h.dashboardSvcServer.SetMongo(result.MongoDB)
		}
		if result.Redis != nil {
			h.dashboardSvcServer.SetRedis(result.Redis)
		}
		if result.CronJobRepo != nil {
			h.dashboardSvcServer.SetCronJobRepo(result.CronJobRepo)
		}
		if result.RunnerSvc != nil {
			runner := result.RunnerSvc
			h.dashboardSvcServer.SetRunnerReady(func() bool { return runner != nil })
		}
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
func SetupRoutes(cfg *config.AppConfig, daemonRegistry *daemon.Registry) (func(r *gin.Engine), *Handlers) {
	configStore := NewConfigStore()
	configRuntime := NewConfigRuntime(configStore, cfg)

	healthRepo := repo.NewHealthRepository()
	healthService := service.NewHealthService(healthRepo, cfg)
	healthHandler := httpHandler.NewHealthHandler(healthService)
	statusService := service.NewStatusService(cfg, configStore)
	statusHandler := httpHandler.NewStatusHandler(statusService)
	a2aHandler := httpHandler.NewA2AHandler(cfg)

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
	dashboardSvcServer := application.NewDashboardServiceServer(configStore, daemonRegistry)
	dashboardTwirp := agentsv1.NewDashboardServiceServer(dashboardSvcServer, pathPrefix)
	daemonSvcServer := application.NewDaemonServiceServer(daemonRegistry)
	daemonTwirp := agentsv1.NewDaemonServiceServer(daemonSvcServer, pathPrefix)
	apiTokenSvcServer := application.NewAPITokenServiceServer(nil)
	apiTokenTwirp := agentsv1.NewAPITokenServiceServer(apiTokenSvcServer, pathPrefix)

	handlers := &Handlers{
		a2aHandler:         a2aHandler,
		agentSvcServer:     agentSvcServer,
		mcpSvcServer:       mcpSvcServer,
		remoteSvcServer:    remoteSvcServer,
		sessionSvcServer:   sessionSvcServer,
		cronSvcServer:      cronSvcServer,
		channelSvcServer:   channelSvcServer,
		dashboardSvcServer: dashboardSvcServer,
		daemonSvcServer:    daemonSvcServer,
		apiTokenSvcServer:  apiTokenSvcServer,
		configStore:        configStore,
		configRuntime:      configRuntime,
		agentRepo:          configStore,
		mcpServerRepo:      configStore,
		remoteAgentRepo:    configStore,
		channelRepo:        configStore,
	}

	router := func(r *gin.Engine) {
		r.Use(httpHandler.APITokenAuthMiddleware(cfg, handlers.apiTokenRepoFromHolder))
		healthHandler.Register(r)
		statusHandler.Register(r)
		a2aHandler.Register(r)

		// Mount Twirp handlers under /api prefix
		r.Any(agentTwirp.PathPrefix()+"*path", gin.WrapH(agentTwirp))
		r.Any(mcpTwirp.PathPrefix()+"*path", gin.WrapH(mcpTwirp))
		r.Any(remoteTwirp.PathPrefix()+"*path", gin.WrapH(remoteTwirp))
		r.Any(channelTwirp.PathPrefix()+"*path", gin.WrapH(channelTwirp))
		r.Any(sessionTwirp.PathPrefix()+"*path", gin.WrapH(sessionTwirp))
		r.Any(cronTwirp.PathPrefix()+"*path", gin.WrapH(cronTwirp))
		r.Any(dashboardTwirp.PathPrefix()+"*path", gin.WrapH(dashboardTwirp))
		r.Any(daemonTwirp.PathPrefix()+"*path", gin.WrapH(daemonTwirp))
		r.Any(apiTokenTwirp.PathPrefix()+"*path", gin.WrapH(apiTokenTwirp))
	}

	return router, handlers
}
