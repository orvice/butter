package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/repo/configstore"
	"go.orx.me/apps/butter/internal/service"
	"go.orx.me/apps/butter/internal/service/configapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Handlers holds all HTTP/Twirp handlers that need post-bootstrap wiring.
type Handlers struct {
	a2aHandler       *httpHandler.A2AHandler
	sessionSvcServer *configapi.SessionServiceServer
	cronSvcServer    *configapi.CronJobServiceServer
	cfgStore         *configstore.Store
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

// SeedConfig seeds the config store from YAML config.
func (h *Handlers) SeedConfig(cfg *config.AppConfig) {
	h.cfgStore.Seed(cfg.Agents, cfg.MCPServerConfigs, cfg.RemoteAgents)
}

// SetupRoutes creates all handlers and returns a Gin router function plus
// the Handlers struct for post-bootstrap wiring.
func SetupRoutes(cfg *config.AppConfig) (func(r *gin.Engine), *Handlers) {
	healthRepo := repo.NewHealthRepository()
	healthService := service.NewHealthService(healthRepo, cfg)
	healthHandler := httpHandler.NewHealthHandler(healthService)
	a2aHandler := httpHandler.NewA2AHandler(cfg)

	cfgStore := configstore.New()
	pathPrefix := twirp.WithServerPathPrefix("/api")
	agentTwirp := agentsv1.NewAgentServiceServer(configapi.NewAgentServiceServer(cfgStore), pathPrefix)
	mcpTwirp := agentsv1.NewMCPServerServiceServer(configapi.NewMCPServerServiceServer(cfgStore), pathPrefix)
	remoteTwirp := agentsv1.NewRemoteAgentServiceServer(configapi.NewRemoteAgentServiceServer(cfgStore), pathPrefix)
	sessionSvcServer := configapi.NewSessionServiceServer()
	sessionTwirp := agentsv1.NewSessionServiceServer(sessionSvcServer, pathPrefix)
	cronSvcServer := configapi.NewCronJobServiceServer()
	cronTwirp := agentsv1.NewCronJobServiceServer(cronSvcServer, pathPrefix)

	router := func(r *gin.Engine) {
		r.Use(httpHandler.APITokenAuthMiddleware(cfg))
		healthHandler.Register(r)
		a2aHandler.Register(r)

		// Mount Twirp handlers under /api prefix
		r.Any(agentTwirp.PathPrefix()+"*path", gin.WrapH(agentTwirp))
		r.Any(mcpTwirp.PathPrefix()+"*path", gin.WrapH(mcpTwirp))
		r.Any(remoteTwirp.PathPrefix()+"*path", gin.WrapH(remoteTwirp))
		r.Any(sessionTwirp.PathPrefix()+"*path", gin.WrapH(sessionTwirp))
		r.Any(cronTwirp.PathPrefix()+"*path", gin.WrapH(cronTwirp))
	}

	handlers := &Handlers{
		a2aHandler:       a2aHandler,
		sessionSvcServer: sessionSvcServer,
		cronSvcServer:    cronSvcServer,
		cfgStore:         cfgStore,
	}

	return router, handlers
}
