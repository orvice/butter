package app

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/mcpoauth"
	workspacemcp "go.orx.me/apps/butter/internal/mcpserver"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/internal/service"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// Handlers holds all HTTP and ConnectRPC handlers that need post-bootstrap wiring.
type Handlers struct {
	a2aHandler             *httpHandler.A2AHandler
	forumSvcServer         *application.ForumServiceServer
	agentSvcServer         *application.AgentServiceServer
	agentFileSvcServer     *application.AgentFileServiceServer
	mcpSvcServer           *application.MCPServerServiceServer
	modelProviderSvcServer *application.ModelProviderServiceServer
	notifyGroupSvcServer   *application.NotifyGroupServiceServer
	remoteSvcServer        *application.RemoteAgentServiceServer
	automationSvcServer    *application.AutomationServiceServer
	sessionSvcServer       *application.SessionServiceServer
	cronSvcServer          *application.CronJobServiceServer
	channelSvcServer       *application.ChannelServiceServer
	dashboardSvcServer     *application.DashboardServiceServer
	daemonSvcServer        *application.DaemonServiceServer
	daemonConnectorHandler *daemon.GRPCHandler
	apiTokenSvcServer      *application.APITokenServiceServer
	authSvcServer          *application.AuthServiceServer
	workspaceSvcServer     *application.WorkspaceServiceServer
	workspaceMCPSvc        *workspacemcp.Service
	authRepo               atomic.Value // auth.Repository
	apiTokenRepo           atomic.Value // apitoken.Repository
	forumRepo              atomic.Value // forum.Repository
	workspaceRepo          atomic.Value // workspace.Repository
	configStore            *ConfigStore
	configRuntime          *ConfigRuntime
	agentRepo              configrepo.AgentRepository
	mcpServerRepo          configrepo.MCPServerRepository
	modelProviderRepo      configrepo.ModelProviderRepository
	notifyGroupRepo        configrepo.NotifyGroupRepository
	remoteAgentRepo        configrepo.RemoteAgentRepository
	channelRepo            configrepo.ChannelRepository
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

func (h *Handlers) authRepoFromHolder() auth.Repository {
	if h == nil {
		return nil
	}
	v := h.authRepo.Load()
	if v == nil {
		return nil
	}
	repo, _ := v.(auth.Repository)
	return repo
}

func (h *Handlers) workspaceRepoFromHolder() workspace.Repository {
	if h == nil {
		return nil
	}
	v := h.workspaceRepo.Load()
	if v == nil {
		return nil
	}
	repo, _ := v.(workspace.Repository)
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
		if h.forumSvcServer != nil {
			h.forumSvcServer.SetRunnerService(result.RunnerSvc)
		}
	}
	if result.InvocationRepo != nil {
		h.agentSvcServer.SetInvocationRepo(result.InvocationRepo)
		if h.dashboardSvcServer != nil {
			h.dashboardSvcServer.SetInvocationRepo(result.InvocationRepo)
		}
	}
	if result.MCPOAuthSvc != nil {
		h.mcpSvcServer.SetOAuthService(result.MCPOAuthSvc)
	}
	if result.AgentFileRepo != nil && h.agentFileSvcServer != nil {
		h.agentFileSvcServer.SetRepo(result.AgentFileRepo)
		h.agentFileSvcServer.SetMaxFileBytes(result.AgentFileMaxBytes)
	}
	if result.AgentFileRepo != nil && h.workspaceMCPSvc != nil {
		h.workspaceMCPSvc.SetAgentFileRepo(result.AgentFileRepo)
	}
	if result.MCPAuthResolver != nil {
		h.mcpSvcServer.SetMCPHTTPClientFactory(result.MCPAuthResolver)
	}
	if result.ForumRepo != nil {
		h.forumRepo.Store(result.ForumRepo)
		if h.forumSvcServer != nil {
			h.forumSvcServer.SetRepo(result.ForumRepo)
		}
	}
	if result.LangfuseHost != "" {
		h.sessionSvcServer.SetLangfuseHost(result.LangfuseHost)
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
	if h.automationSvcServer != nil {
		if result.AutomationDefRepo != nil || result.AutomationRunRepo != nil || result.AutomationStepRepo != nil {
			h.automationSvcServer.SetRepos(result.AutomationDefRepo, result.AutomationRunRepo, result.AutomationStepRepo)
		}
		if result.AutomationEngine != nil {
			h.automationSvcServer.SetEngine(result.AutomationEngine)
		}
		if result.RunnerSvc != nil {
			h.automationSvcServer.SetAgentValidator(result.RunnerSvc)
		}
		if result.AutomationScheduler != nil {
			h.automationSvcServer.SetScheduler(result.AutomationScheduler)
		}
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
		h.modelProviderSvcServer.SetRuntime(h.configRuntime)
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
		if h.daemonSvcServer != nil {
			h.daemonSvcServer.SetAPITokenRepo(result.APITokenRepo)
		}
		if h.daemonConnectorHandler != nil {
			h.daemonConnectorHandler.SetAPITokenRepo(result.APITokenRepo)
		}
	}
	if result.AuthRepo != nil {
		h.authRepo.Store(result.AuthRepo)
		if h.authSvcServer != nil {
			h.authSvcServer.SetRepo(result.AuthRepo)
		}
	}
	if h.authSvcServer != nil {
		if result.OAuthStateRepo != nil {
			h.authSvcServer.SetOAuthStateRepo(result.OAuthStateRepo)
		}
		if result.OAuthProviders != nil {
			h.authSvcServer.SetProviderRegistry(result.OAuthProviders)
		}
	}
	if result.WorkspaceRepo != nil {
		h.workspaceRepo.Store(result.WorkspaceRepo)
		if h.workspaceSvcServer != nil {
			h.workspaceSvcServer.SetRepo(result.WorkspaceRepo)
		}
		if h.workspaceMCPSvc != nil {
			h.workspaceMCPSvc.SetWorkspaceRepo(result.WorkspaceRepo)
		}
		if h.authSvcServer != nil {
			h.authSvcServer.SetWorkspaceRepo(result.WorkspaceRepo)
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
		if result.CronRepo != nil {
			h.dashboardSvcServer.SetCronExecutionRepo(result.CronRepo)
		}
		if result.SessionCounter != nil {
			h.dashboardSvcServer.SetSessionCounter(result.SessionCounter)
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

// NotifyGroupRepo returns the notify group repository.
func (h *Handlers) NotifyGroupRepo() configrepo.NotifyGroupRepository {
	return h.notifyGroupRepo
}

// ConfigStore returns the shared workspace-scoped configuration repository.
func (h *Handlers) ConfigStore() *ConfigStore {
	return h.configStore
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
	// Lazy provider: SetupRoutes runs before core.New loads YAML into cfg,
	// so we read cfg.Static on every request instead of snapshotting now.
	uploadSvc := service.NewUploadServiceLazy(func() config.StaticConfig { return cfg.Static })
	uploadHandler := httpHandler.NewUploadHandler(uploadSvc)

	agentSvcServer := application.NewAgentServiceServer(configStore)
	agentFileSvcServer := application.NewAgentFileServiceServer(nil)
	mcpSvcServer := application.NewMCPServerServiceServer(configStore)
	modelProviderSvcServer := application.NewModelProviderServiceServer(configStore)
	notifyGroupSvcServer := application.NewNotifyGroupServiceServer(configStore)
	remoteSvcServer := application.NewRemoteAgentServiceServer(configStore)
	remoteSvcServer.SetDaemonRegistry(daemonRegistry)
	automationSvcServer := application.NewAutomationServiceServer()
	// Every Connect handler shares the same option set so the wire format
	// matches the pre-migration Twirp behavior (snake_case JSON). Without
	// this, dashboard callers that still read response fields like
	// connected_daemons / base_url / space_id without camelCase fallbacks
	// silently see undefined.
	connectOpts := connectx.HandlerOptions()
	forumSvcServer := application.NewForumServiceServer(nil)
	forumConnectPath, forumConnectHandler := agentsv1connect.NewForumServiceHandler(forumSvcServer, connectOpts...)
	agentConnectPath, agentConnectHandler := agentsv1connect.NewAgentServiceHandler(agentSvcServer, connectOpts...)
	agentFileConnectPath, agentFileConnectHandler := agentsv1connect.NewAgentFileServiceHandler(agentFileSvcServer, connectOpts...)
	mcpConnectPath, mcpConnectHandler := agentsv1connect.NewMCPServerServiceHandler(mcpSvcServer, connectOpts...)
	modelProviderConnectPath, modelProviderConnectHandler := agentsv1connect.NewModelProviderServiceHandler(modelProviderSvcServer, connectOpts...)
	notifyGroupConnectPath, notifyGroupConnectHandler := agentsv1connect.NewNotifyGroupServiceHandler(notifyGroupSvcServer, connectOpts...)
	remoteConnectPath, remoteConnectHandler := agentsv1connect.NewRemoteAgentServiceHandler(remoteSvcServer, connectOpts...)
	automationConnectPath, automationConnectHandler := agentsv1connect.NewAutomationServiceHandler(automationSvcServer, connectOpts...)
	channelSvcServer := application.NewChannelServiceServer(configStore)
	channelConnectPath, channelConnectHandler := agentsv1connect.NewChannelServiceHandler(channelSvcServer, connectOpts...)
	sessionSvcServer := application.NewSessionServiceServer()
	sessionConnectPath, sessionConnectHandler := agentsv1connect.NewSessionServiceHandler(sessionSvcServer, connectOpts...)
	cronSvcServer := application.NewCronJobServiceServer()
	cronConnectPath, cronConnectHandler := agentsv1connect.NewCronJobServiceHandler(cronSvcServer, connectOpts...)
	dashboardSvcServer := application.NewDashboardServiceServer(configStore, daemonRegistry)
	dashboardConnectPath, dashboardConnectHandler := agentsv1connect.NewDashboardServiceHandler(dashboardSvcServer, connectOpts...)
	daemonSvcServer := application.NewDaemonServiceServer(configStore, daemonRegistry)
	daemonConnectPath, daemonConnectHandler := agentsv1connect.NewDaemonServiceHandler(daemonSvcServer, connectOpts...)
	daemonConnectorHandler := daemon.NewGRPCHandler(daemonRegistry, nil, configStore)
	daemonConnectorConnectPath, daemonConnectorConnectHandler := agentsv1connect.NewDaemonConnectorServiceHandler(daemonConnectorHandler, connectOpts...)
	apiTokenSvcServer := application.NewAPITokenServiceServer(nil)
	apiTokenConnectPath, apiTokenConnectHandler := agentsv1connect.NewAPITokenServiceHandler(apiTokenSvcServer, connectOpts...)
	globalMCPSvcServer := application.NewGlobalMCPServerServiceServer(configStore, mcpSvcServer)
	globalMCPConnectPath, globalMCPConnectHandler := agentsv1connect.NewGlobalMCPServerServiceHandler(globalMCPSvcServer, connectOpts...)
	authSvcServer := application.NewAuthServiceServer(nil, cfg.Auth.EffectiveSessionTTL())
	authConnectPath, authConnectHandler := agentsv1connect.NewAuthServiceHandler(authSvcServer, connectOpts...)
	workspaceSvcServer := application.NewWorkspaceServiceServer(nil)
	workspaceConnectPath, workspaceConnectHandler := agentsv1connect.NewWorkspaceServiceHandler(workspaceSvcServer, connectOpts...)
	workspaceMCPSvc := workspacemcp.NewService(configStore)

	handlers := &Handlers{
		a2aHandler:             a2aHandler,
		forumSvcServer:         forumSvcServer,
		agentSvcServer:         agentSvcServer,
		agentFileSvcServer:     agentFileSvcServer,
		mcpSvcServer:           mcpSvcServer,
		modelProviderSvcServer: modelProviderSvcServer,
		notifyGroupSvcServer:   notifyGroupSvcServer,
		remoteSvcServer:        remoteSvcServer,
		automationSvcServer:    automationSvcServer,
		sessionSvcServer:       sessionSvcServer,
		cronSvcServer:          cronSvcServer,
		channelSvcServer:       channelSvcServer,
		dashboardSvcServer:     dashboardSvcServer,
		daemonSvcServer:        daemonSvcServer,
		daemonConnectorHandler: daemonConnectorHandler,
		apiTokenSvcServer:      apiTokenSvcServer,
		authSvcServer:          authSvcServer,
		workspaceSvcServer:     workspaceSvcServer,
		workspaceMCPSvc:        workspaceMCPSvc,
		configStore:            configStore,
		configRuntime:          configRuntime,
		agentRepo:              configStore,
		mcpServerRepo:          configStore,
		modelProviderRepo:      configStore,
		notifyGroupRepo:        configStore,
		remoteAgentRepo:        configStore,
		channelRepo:            configStore,
	}

	router := func(r *gin.Engine) {
		r.Use(httpHandler.AuthMiddleware(cfg, handlers.authRepoFromHolder, handlers.apiTokenRepoFromHolder, handlers.workspaceRepoFromHolder))
		healthHandler.Register(r)
		statusHandler.Register(r)
		a2aHandler.Register(r)
		uploadHandler.Register(r)
		httpHandler.RegisterWorkspaceMCP(r, workspaceMCPSvc.Handler(), handlers.workspaceRepoFromHolder)
		r.GET(mcpoauth.CallbackPath, func(c *gin.Context) {
			status := "error"
			serverID := ""
			target := oauthCallbackFallback(cfg)
			if oauthErr := strings.TrimSpace(c.Query("error")); oauthErr != "" {
				c.Redirect(http.StatusFound, appendOAuthCallbackParams(target, status, serverID))
				return
			}
			returnURL, oauthStatus, err := mcpSvcServer.CompleteMCPServerOAuthCallback(c.Request.Context(), c.Query("state"), c.Query("code"))
			if returnURL != "" {
				target = returnURL
			}
			if err == nil {
				status = "success"
				if oauthStatus != nil {
					serverID = oauthStatus.GetServerId()
				}
			}
			c.Redirect(http.StatusFound, appendOAuthCallbackParams(target, status, serverID))
		})

		// Connect handlers mount at /agents.v1.XxxService/ by default; we
		// strip the /api prefix before forwarding so the public URLs
		// (/api/agents.v1.XxxService/Method) stay stable across the
		// migration.
		r.Any("/api"+authConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", authConnectHandler)))
		r.Any("/api"+workspaceConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", workspaceConnectHandler)))
		r.Any("/api"+apiTokenConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", apiTokenConnectHandler)))
		r.Any("/api"+agentConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", agentConnectHandler)))
		r.Any("/api"+mcpConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", mcpConnectHandler)))
		r.Any("/api"+modelProviderConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", modelProviderConnectHandler)))
		r.Any("/api"+notifyGroupConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", notifyGroupConnectHandler)))
		r.Any("/api"+remoteConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", remoteConnectHandler)))
		r.Any("/api"+automationConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", automationConnectHandler)))
		r.Any("/api"+channelConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", channelConnectHandler)))
		r.Any("/api"+forumConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", forumConnectHandler)))
		r.Any("/api"+agentFileConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", agentFileConnectHandler)))
		r.Any("/api"+sessionConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", sessionConnectHandler)))
		r.Any("/api"+cronConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", cronConnectHandler)))
		r.Any("/api"+dashboardConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", dashboardConnectHandler)))
		r.Any("/api"+daemonConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", daemonConnectHandler)))
		r.Any("/api"+daemonConnectorConnectPath+"*path", gin.WrapH(enableFullDuplex(http.StripPrefix("/api", daemonConnectorConnectHandler))))
		r.Any("/api"+globalMCPConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", globalMCPConnectHandler)))
	}

	return router, handlers
}

func enableFullDuplex(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = http.NewResponseController(w).EnableFullDuplex()
		handler.ServeHTTP(w, r)
	})
}

func oauthCallbackFallback(cfg *config.AppConfig) string {
	base := strings.TrimSpace(cfg.MCPOAuth.DashboardBaseURL)
	if base == "" {
		return "/mcp-servers"
	}
	return strings.TrimRight(base, "/") + "/mcp-servers"
}

func appendOAuthCallbackParams(raw, status, serverID string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	q.Set("mcp_oauth", status)
	if serverID != "" {
		q.Set("server_id", serverID)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
