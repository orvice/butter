package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"github.com/twitchtv/twirp"

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
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Handlers holds all HTTP/Twirp handlers that need post-bootstrap wiring.
type Handlers struct {
	a2aHandler             *httpHandler.A2AHandler
	chatStreamHandler      *httpHandler.ChatStreamHandler
	forumSvcServer         *application.ForumServiceServer
	agentSvcServer         *application.AgentServiceServer
	agentFileSvcServer     *application.AgentFileServiceServer
	mcpSvcServer           *application.MCPServerServiceServer
	modelProviderSvcServer *application.ModelProviderServiceServer
	notifyGroupSvcServer   *application.NotifyGroupServiceServer
	remoteSvcServer        *application.RemoteAgentServiceServer
	sessionSvcServer       *application.SessionServiceServer
	cronSvcServer          *application.CronJobServiceServer
	channelSvcServer       *application.ChannelServiceServer
	dashboardSvcServer     *application.DashboardServiceServer
	daemonSvcServer        *application.DaemonServiceServer
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
	globalMCPServerRepo    configrepo.GlobalMCPServerRepository
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
		h.chatStreamHandler.SetRunnerService(result.RunnerSvc)
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
	chatStreamHandler := httpHandler.NewChatStreamHandler()
	// Lazy provider: SetupRoutes runs before core.New loads YAML into cfg,
	// so we read cfg.Static on every request instead of snapshotting now.
	uploadSvc := service.NewUploadServiceLazy(func() config.StaticConfig { return cfg.Static })
	uploadHandler := httpHandler.NewUploadHandler(uploadSvc)

	pathPrefix := twirp.WithServerPathPrefix("/api")
	agentSvcServer := application.NewAgentServiceServer(configStore)
	agentFileSvcServer := application.NewAgentFileServiceServer(nil)
	mcpSvcServer := application.NewMCPServerServiceServer(configStore)
	modelProviderSvcServer := application.NewModelProviderServiceServer(configStore)
	notifyGroupSvcServer := application.NewNotifyGroupServiceServer(configStore)
	remoteSvcServer := application.NewRemoteAgentServiceServer(configStore)
	remoteSvcServer.SetDaemonRegistry(daemonRegistry)
	forumSvcServer := application.NewForumServiceServer(nil)
	forumTwirp := agentsv1.NewForumServiceServer(forumSvcServer, pathPrefix)
	agentConnectPath, agentConnectHandler := agentsv1connect.NewAgentServiceHandler(application.NewAgentServiceConnectAdapter(agentSvcServer))
	agentFileTwirp := agentsv1.NewAgentFileServiceServer(agentFileSvcServer, pathPrefix)
	mcpConnectPath, mcpConnectHandler := agentsv1connect.NewMCPServerServiceHandler(application.NewMCPServerServiceConnectAdapter(mcpSvcServer))
	modelProviderConnectPath, modelProviderConnectHandler := agentsv1connect.NewModelProviderServiceHandler(application.NewModelProviderServiceConnectAdapter(modelProviderSvcServer))
	notifyGroupConnectPath, notifyGroupConnectHandler := agentsv1connect.NewNotifyGroupServiceHandler(application.NewNotifyGroupServiceConnectAdapter(notifyGroupSvcServer))
	remoteConnectPath, remoteConnectHandler := agentsv1connect.NewRemoteAgentServiceHandler(application.NewRemoteAgentServiceConnectAdapter(remoteSvcServer))
	channelSvcServer := application.NewChannelServiceServer(configStore)
	channelConnectPath, channelConnectHandler := agentsv1connect.NewChannelServiceHandler(application.NewChannelServiceConnectAdapter(channelSvcServer))
	sessionSvcServer := application.NewSessionServiceServer()
	sessionTwirp := agentsv1.NewSessionServiceServer(sessionSvcServer, pathPrefix)
	cronSvcServer := application.NewCronJobServiceServer()
	cronTwirp := agentsv1.NewCronJobServiceServer(cronSvcServer, pathPrefix)
	dashboardSvcServer := application.NewDashboardServiceServer(configStore, daemonRegistry)
	dashboardTwirp := agentsv1.NewDashboardServiceServer(dashboardSvcServer, pathPrefix)
	daemonSvcServer := application.NewDaemonServiceServer(daemonRegistry)
	daemonTwirp := agentsv1.NewDaemonServiceServer(daemonSvcServer, pathPrefix)
	apiTokenSvcServer := application.NewAPITokenServiceServer(nil)
	apiTokenConnectPath, apiTokenConnectHandler := agentsv1connect.NewAPITokenServiceHandler(application.NewAPITokenServiceConnectAdapter(apiTokenSvcServer))
	authSvcServer := application.NewAuthServiceServer(nil, cfg.Auth.EffectiveSessionTTL())
	authConnectPath, authConnectHandler := agentsv1connect.NewAuthServiceHandler(application.NewAuthServiceConnectAdapter(authSvcServer))
	workspaceSvcServer := application.NewWorkspaceServiceServer(nil)
	workspaceConnectPath, workspaceConnectHandler := agentsv1connect.NewWorkspaceServiceHandler(application.NewWorkspaceServiceConnectAdapter(workspaceSvcServer))
	workspaceMCPSvc := workspacemcp.NewService(configStore)

	handlers := &Handlers{
		a2aHandler:             a2aHandler,
		chatStreamHandler:      chatStreamHandler,
		forumSvcServer:         forumSvcServer,
		agentSvcServer:         agentSvcServer,
		agentFileSvcServer:     agentFileSvcServer,
		mcpSvcServer:           mcpSvcServer,
		modelProviderSvcServer: modelProviderSvcServer,
		notifyGroupSvcServer:   notifyGroupSvcServer,
		remoteSvcServer:        remoteSvcServer,
		sessionSvcServer:       sessionSvcServer,
		cronSvcServer:          cronSvcServer,
		channelSvcServer:       channelSvcServer,
		dashboardSvcServer:     dashboardSvcServer,
		daemonSvcServer:        daemonSvcServer,
		apiTokenSvcServer:      apiTokenSvcServer,
		authSvcServer:          authSvcServer,
		workspaceSvcServer:     workspaceSvcServer,
		workspaceMCPSvc:        workspaceMCPSvc,
		configStore:            configStore,
		configRuntime:          configRuntime,
		agentRepo:              configStore,
		globalMCPServerRepo:    configStore,
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
		chatStreamHandler.Register(r)
		uploadHandler.Register(r)
		httpHandler.RegisterWorkspaceMCP(r, workspaceMCPSvc.Handler(), handlers.workspaceRepoFromHolder)
		registerGlobalMCPServerRoutes(r, handlers, mcpSvcServer)
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

		// Mount Twirp handlers under /api prefix
		r.Any(forumTwirp.PathPrefix()+"*path", gin.WrapH(forumTwirp))
		r.Any(agentFileTwirp.PathPrefix()+"*path", gin.WrapH(agentFileTwirp))
		r.Any(sessionTwirp.PathPrefix()+"*path", gin.WrapH(sessionTwirp))
		r.Any(cronTwirp.PathPrefix()+"*path", gin.WrapH(cronTwirp))
		r.Any(dashboardTwirp.PathPrefix()+"*path", gin.WrapH(dashboardTwirp))
		r.Any(daemonTwirp.PathPrefix()+"*path", gin.WrapH(daemonTwirp))
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
		r.Any("/api"+channelConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", channelConnectHandler)))
	}

	return router, handlers
}

func oauthCallbackFallback(cfg *config.AppConfig) string {
	base := strings.TrimSpace(cfg.MCPOAuth.DashboardBaseURL)
	if base == "" {
		return "/mcp-servers"
	}
	return strings.TrimRight(base, "/") + "/mcp-servers"
}

type installGlobalMCPServerRequest struct {
	WorkspaceID string `json:"workspace_id,omitempty"`
}

func registerGlobalMCPServerRoutes(r *gin.Engine, handlers *Handlers, mcpSvc *application.MCPServerServiceServer) {
	r.GET("/api/global-mcp-servers", func(c *gin.Context) {
		servers, err := handlers.globalMCPServerRepo.ListGlobalMCPServers(c.Request.Context())
		if err != nil {
			writeError(c, err)
			return
		}
		writeMCPServerList(c, servers, !auth.IsAdmin(c.Request.Context()))
	})

	r.POST("/api/admin/global-mcp-servers", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		server, ok := readMCPServer(c)
		if !ok {
			return
		}
		created, err := handlers.globalMCPServerRepo.CreateGlobalMCPServer(c.Request.Context(), server)
		if err != nil {
			writeError(c, err)
			return
		}
		writeMCPServer(c, http.StatusCreated, created, false)
	})

	r.PUT("/api/admin/global-mcp-servers/:id", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		server, ok := readMCPServer(c)
		if !ok {
			return
		}
		if server.GetId() == "" {
			server.Id = c.Param("id")
		}
		if server.GetId() != c.Param("id") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id mismatch"})
			return
		}
		updated, err := handlers.globalMCPServerRepo.UpdateGlobalMCPServer(c.Request.Context(), server)
		if err != nil {
			writeError(c, err)
			return
		}
		writeMCPServer(c, http.StatusOK, updated, false)
	})

	r.DELETE("/api/admin/global-mcp-servers/:id", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		if err := handlers.globalMCPServerRepo.DeleteGlobalMCPServer(c.Request.Context(), c.Param("id")); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.POST("/api/global-mcp-servers/:id/install", func(c *gin.Context) {
		workspaceID, ok := installWorkspaceID(c)
		if !ok {
			return
		}
		var req installGlobalMCPServerRequest
		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
		requestedWorkspaceID := strings.TrimSpace(req.WorkspaceID)
		if requestedWorkspaceID != "" {
			if !auth.IsAdmin(c.Request.Context()) && requestedWorkspaceID != workspaceID {
				c.JSON(http.StatusForbidden, gin.H{"error": "admin role required for cross-workspace install"})
				return
			}
			if requestedWorkspaceID != workspaceID {
				auditAdminCrossWorkspaceInstall(c, workspaceID, requestedWorkspaceID, c.Param("id"))
			}
			workspaceID = requestedWorkspaceID
		}
		preset, err := handlers.globalMCPServerRepo.GetGlobalMCPServer(c.Request.Context(), c.Param("id"))
		if err != nil {
			writeError(c, err)
			return
		}
		server := proto.Clone(preset).(*agentsv1.MCPServer)
		server.WorkspaceId = ""
		application.MarkInstalledGlobalMCPPreset(server, preset.GetId())
		ctx := wsctx.WithID(c.Request.Context(), workspaceID)
		created, err := mcpSvc.CreateMCPServer(ctx, &agentsv1.CreateMCPServerRequest{McpServer: server})
		if err != nil {
			writeError(c, err)
			return
		}
		writeMCPServer(c, http.StatusCreated, created.GetMcpServer(), true)
	})
}

func requireAdmin(c *gin.Context) bool {
	if auth.IsAdmin(c.Request.Context()) {
		return true
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
	return false
}

// auditAdminCrossWorkspaceInstall emits a structured audit log entry when an
// admin installs a global MCP preset into a workspace whose context they did
// not explicitly enter. Cross-workspace installs are intentionally allowed
// for ops/automation, but they need a paper trail because a compromised
// admin can use this path to plant SSRF-capable MCP servers in any tenant.
func auditAdminCrossWorkspaceInstall(c *gin.Context, contextWorkspaceID, targetWorkspaceID, presetID string) {
	ctx := c.Request.Context()
	logger := log.FromContext(ctx).With("audit", "admin_cross_workspace_install")
	fields := []any{
		"preset_id", presetID,
		"context_workspace_id", contextWorkspaceID,
		"target_workspace_id", targetWorkspaceID,
		"remote_addr", c.ClientIP(),
	}
	if user, ok := auth.UserFromContext(ctx); ok {
		fields = append(fields, "user_id", user.GetId(), "user_role", user.GetRole())
	}
	logger.Warn("admin installed global MCP preset into another workspace", fields...)
}

func installWorkspaceID(c *gin.Context) (string, bool) {
	workspaceID, ok := wsctx.FromContext(c.Request.Context())
	if ok {
		return workspaceID, true
	}
	if auth.IsAdmin(c.Request.Context()) {
		return "", true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "workspace required"})
	return "", false
}

func readMCPServer(c *gin.Context) (*agentsv1.MCPServer, bool) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, false
	}
	server := &agentsv1.MCPServer{}
	if err := protojson.Unmarshal(body, server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, false
	}
	server.WorkspaceId = ""
	return server, true
}

func writeMCPServerList(c *gin.Context, servers []*agentsv1.MCPServer, redact bool) {
	out := make([]json.RawMessage, 0, len(servers))
	for _, server := range servers {
		raw, err := marshalMCPServer(mcpServerForResponse(server, redact))
		if err != nil {
			writeError(c, err)
			return
		}
		out = append(out, raw)
	}
	c.JSON(http.StatusOK, gin.H{"mcp_servers": out})
}

func writeMCPServer(c *gin.Context, status int, server *agentsv1.MCPServer, redact bool) {
	raw, err := marshalMCPServer(mcpServerForResponse(server, redact))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(status, gin.H{"mcp_server": raw})
}

func marshalMCPServer(server *agentsv1.MCPServer) (json.RawMessage, error) {
	b, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(server)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func mcpServerForResponse(server *agentsv1.MCPServer, redact bool) *agentsv1.MCPServer {
	if server == nil {
		return nil
	}
	clone := proto.Clone(server).(*agentsv1.MCPServer)
	if redact {
		redactMCPServerSecret(clone)
	}
	return clone
}

func redactMCPServerSecret(server *agentsv1.MCPServer) {
	oauth := server.GetAuth().GetOauth2()
	if oauth != nil {
		oauth.ClientSecret = ""
	}
}

func writeError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if twerr, ok := err.(twirp.Error); ok {
		switch twerr.Code() {
		case twirp.InvalidArgument, twirp.Malformed:
			c.JSON(http.StatusBadRequest, gin.H{"error": twerr.Msg()})
		case twirp.PermissionDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": twerr.Msg()})
		case twirp.NotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": twerr.Msg()})
		case twirp.AlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": twerr.Msg()})
		case twirp.FailedPrecondition:
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": twerr.Msg()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": twerr.Msg()})
		}
		return
	}
	// Once a service is migrated to a Connect adapter the upstream calls
	// surfaced via this REST helper start returning *connect.Error instead
	// of twirp.Error; map the codes back to HTTP so REST callers see no
	// change.
	var cerr *connect.Error
	if errors.As(err, &cerr) {
		c.JSON(connectCodeToHTTPStatus(cerr.Code()), gin.H{"error": cerr.Message()})
		return
	}
	if errors.Is(err, configrepo.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, configrepo.ErrAlreadyExists) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func connectCodeToHTTPStatus(code connect.Code) int {
	switch code {
	case connect.CodeInvalidArgument:
		return http.StatusBadRequest
	case connect.CodeUnauthenticated:
		return http.StatusUnauthorized
	case connect.CodePermissionDenied:
		return http.StatusForbidden
	case connect.CodeNotFound:
		return http.StatusNotFound
	case connect.CodeAlreadyExists:
		return http.StatusConflict
	case connect.CodeFailedPrecondition:
		return http.StatusPreconditionFailed
	case connect.CodeUnimplemented:
		return http.StatusNotImplemented
	case connect.CodeUnavailable:
		return http.StatusServiceUnavailable
	case connect.CodeDeadlineExceeded:
		return http.StatusGatewayTimeout
	case connect.CodeResourceExhausted:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
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
