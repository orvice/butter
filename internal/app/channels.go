package app

import (
	"context"
	"fmt"
	"strings"

	"butterfly.orx.me/core/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/adk/session"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/auth/provider"
	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/mcpoauth"
	"go.orx.me/apps/butter/internal/repo/agentfile"
	agentfilememory "go.orx.me/apps/butter/internal/repo/agentfile/memory"
	agentfilemongo "go.orx.me/apps/butter/internal/repo/agentfile/mongo"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	apitokenmemory "go.orx.me/apps/butter/internal/repo/apitoken/memory"
	apitokenmongo "go.orx.me/apps/butter/internal/repo/apitoken/mongo"
	"go.orx.me/apps/butter/internal/repo/auth"
	authmongo "go.orx.me/apps/butter/internal/repo/auth/mongo"
	authredis "go.orx.me/apps/butter/internal/repo/auth/redis"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/forum"
	forummemory "go.orx.me/apps/butter/internal/repo/forum/memory"
	forummongo "go.orx.me/apps/butter/internal/repo/forum/mongo"
	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	invocationmongo "go.orx.me/apps/butter/internal/repo/invocation/mongo"
	mcpoauthrepo "go.orx.me/apps/butter/internal/repo/mcpoauth"
	mcpoauthmemory "go.orx.me/apps/butter/internal/repo/mcpoauth/memory"
	mcpoauthmongo "go.orx.me/apps/butter/internal/repo/mcpoauth/mongo"
	"go.orx.me/apps/butter/internal/repo/oauthstate"
	oauthstatememory "go.orx.me/apps/butter/internal/repo/oauthstate/memory"
	oauthstatemongo "go.orx.me/apps/butter/internal/repo/oauthstate/mongo"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	workspacemongo "go.orx.me/apps/butter/internal/repo/workspace/mongo"
	internalautomation "go.orx.me/apps/butter/internal/runtime/automation"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	mongomemory "go.orx.me/apps/butter/internal/runtime/memory/mongo"
	"go.orx.me/apps/butter/internal/runtime/runner"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
)

// BootstrapResult holds the services created during bootstrap.
type BootstrapResult struct {
	RunnerSvc           *runner.Service
	SessionSvc          session.Service
	CronScheduler       *internalcron.Scheduler
	CronRepo            internalcron.ExecutionRepo
	CronJobRepo         internalcron.JobRepo
	AutomationEngine    *internalautomation.Engine
	AutomationScheduler *internalautomation.Scheduler
	AutomationDefRepo   internalautomation.DefinitionRepo
	AutomationRunRepo   internalautomation.RunRepo
	AutomationStepRepo  internalautomation.StepRunRepo
	ChannelMgr          *channel.Manager
	MongoDB             *mongo.Database
	Redis               *redis.Client
	AuthRepo            auth.Repository
	OAuthStateRepo      oauthstate.Repository
	OAuthProviders      *provider.Registry
	APITokenRepo        apitoken.Repository
	InvocationRepo      invocation.Repository
	ForumRepo           forum.Repository
	WorkspaceRepo       workspacerepo.Repository
	MCPOAuthRepo        mcpoauthrepo.Repository
	MCPOAuthSvc         *mcpoauth.Service
	MCPAuthResolver     *mcpoauth.Resolver
	AgentFileRepo       agentfile.Repository
	AgentFileMaxBytes   int64
	LangfuseHost        string
	SessionCounter      func(ctx context.Context) (int64, error)
}

// StartChannels initializes MongoDB, Redis, runner service, channel manager,
// and cron scheduler. It returns the bootstrap result.
// agentRepo is the shared agent repository used by the system agent for agent queries.
func StartChannels(ctx context.Context, cfg *config.AppConfig, agentRepo configrepo.AgentRepository, channelRepo configrepo.ChannelRepository, notifyGroupRepo configrepo.NotifyGroupRepository, daemonRegistry *daemon.Registry) (*BootstrapResult, error) {
	logger := log.FromContext(ctx)

	// Connect to MongoDB.
	db, err := connectMongo(ctx, cfg)
	if err != nil {
		return nil, err
	}

	sessionSvc, err := mongosession.New(ctx, db)
	if err != nil {
		logger.Error("failed to create mongo session service", "err", err)
		return nil, err
	}

	memorySvc, err := mongomemory.New(ctx, db)
	if err != nil {
		logger.Error("failed to create mongo memory service", "err", err)
		return nil, err
	}

	// Connect to Redis.
	rdb := connectRedis(ctx, cfg)

	// Pick auth, API token + invocation repository backends.
	var (
		authRepo       auth.Repository
		tokenRepo      apitoken.Repository
		invRepo        invocation.Repository
		forumRepo      forum.Repository
		wsRepo         workspacerepo.Repository
		oauthRepo      mcpoauthrepo.Repository
		fileRepo       agentfile.Repository
		oauthStateRepo oauthstate.Repository
	)
	authUserRepo := authmongo.New(db)
	logger.Info("initializing auth bootstrap")
	if err := application.BootstrapInitialAdmin(ctx, authUserRepo, cfg.Auth); err != nil {
		logger.Error("failed to initialize auth", "err", err)
		return nil, err
	}
	logger.Info("auth bootstrap completed")
	authRepo = authredis.New(authUserRepo, rdb)
	switch backend := strings.ToLower(strings.TrimSpace(cfg.StorageBackend)); backend {
	case "", "mongo":
		tokenRepo = apitokenmongo.New(db)
		invMongo := invocationmongo.New(db)
		if err := invMongo.EnsureIndexes(ctx); err != nil {
			logger.Error("failed to create invocation indexes", "err", err)
			return nil, err
		}
		invRepo = invMongo
		forumRepo = forummongo.New(db)
		wsRepo = workspacemongo.New(db)
		oauthRepo = mcpoauthmongo.New(db)
		fileRepo = agentfilemongo.New(db, setupAgentFileContentStore(ctx, cfg))
		oauthStateRepo = oauthstatemongo.New(db)
	case "memory":
		tokenRepo = apitokenmemory.New()
		invRepo = invocationmemory.New()
		forumRepo = forummemory.New()
		wsRepo = workspacememory.New()
		oauthRepo = mcpoauthmemory.New()
		fileRepo = agentfilememory.New()
		oauthStateRepo = oauthstatememory.New()
	default:
		return nil, fmt.Errorf("unsupported storage backend %q", cfg.StorageBackend)
	}
	if err := wsRepo.EnsureIndexes(ctx); err != nil {
		logger.Error("failed to create workspace indexes", "err", err)
		return nil, err
	}
	if forumRepo != nil {
		if err := forumRepo.EnsureIndexes(ctx); err != nil {
			logger.Error("failed to create forum indexes", "err", err)
			return nil, err
		}
	}
	if err := application.BootstrapDefaultWorkspace(ctx, wsRepo, authRepo); err != nil {
		logger.Error("failed to bootstrap default workspace", "err", err)
		return nil, err
	}
	if err := oauthRepo.EnsureIndexes(ctx); err != nil {
		logger.Error("failed to create mcp oauth indexes", "err", err)
		return nil, err
	}
	if err := fileRepo.EnsureIndexes(ctx); err != nil {
		logger.Error("failed to create agent file indexes", "err", err)
		return nil, err
	}
	if err := oauthStateRepo.EnsureIndexes(ctx); err != nil {
		logger.Error("failed to create oauth state indexes", "err", err)
		return nil, err
	}
	oauthProviders := provider.BuildRegistry(cfg.Auth)
	if oauthProviders != nil {
		logger.Info("oauth login providers configured", "count", len(oauthProviders.List()))
	}
	oauthConfigProvider := func() mcpoauth.Config {
		return mcpoauth.Config{
			CallbackBaseURL:   cfg.MCPOAuth.CallbackBaseURL,
			DashboardBaseURL:  cfg.MCPOAuth.DashboardBaseURL,
			EncryptionKey:     cfg.MCPOAuth.EncryptionKey,
			AllowInsecureHTTP: cfg.MCPOAuth.AllowInsecureHTTP,
		}
	}
	oauthSvc := mcpoauth.NewService(oauthRepo, mcpoauth.NewMemoryFlowStore(), oauthConfigProvider)
	mcpAuthResolver := mcpoauth.NewResolver(oauthRepo, oauthConfigProvider)

	// Setup Langfuse plugin if configured.
	pluginConfig, err := setupLangfuse(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Setup S3-backed artifact service if configured. nil disables artifacts.
	artifactSvc := setupArtifactService(ctx, cfg)

	// Build runner service.
	logger.Info("building runner service", "agent_count", len(cfg.Agents))
	runnerSvc, err := runner.NewServiceWithMCPHTTPClientFactory(ctx, cfg.Agents, cfg.ModelProviders, cfg.MCPServerConfigs, cfg.RemoteAgents, daemonRegistry, sessionSvc, memorySvc, artifactSvc, fileRepo, cfg.AgentFiles.EffectiveMaxFileBytes(), pluginConfig, mcpAuthResolver)
	if err == nil {
		runnerSvc.SetInvocationRecorder(invRepo)
	}
	if err != nil {
		logger.Error("failed to build runner service", "err", err)
		return nil, err
	}

	// Initialize cron scheduler.
	cronScheduler, cronExecRepo, cronJobRepo, err := startCron(ctx, db, runnerSvc, notifyGroupRepo, channelRepo)
	if err != nil {
		return nil, err
	}

	automationDefRepo := internalautomation.NewMongoDefinitionRepo(db)
	automationRunRepo := internalautomation.NewMongoRunRepo(db)
	automationStepRepo := internalautomation.NewMongoStepRunRepo(db)
	for name, repo := range map[string]interface{ EnsureIndexes(context.Context) error }{
		"automation definitions": automationDefRepo,
		"automation runs":        automationRunRepo,
		"automation step runs":   automationStepRepo,
	} {
		if err := repo.EnsureIndexes(ctx); err != nil {
			logger.Error("failed to create automation indexes", "repo", name, "err", err)
			return nil, err
		}
	}
	automationEngine := internalautomation.NewEngine(automationDefRepo, automationRunRepo, automationStepRepo, internalautomation.EngineOptions{
		Runner:          runnerSvc,
		NotifyGroupRepo: notifyGroupRepo,
		ForumRepo:       forumRepo,
	})
	automationScheduler, err := internalautomation.NewScheduler(ctx, automationDefRepo, automationEngine)
	if err != nil {
		logger.Error("failed to create automation scheduler", "err", err)
		return nil, err
	}
	automationScheduler.Start()
	logger.Info("automation scheduler started")

	go func() {
		<-ctx.Done()
		stopCtx := automationScheduler.Stop()
		<-stopCtx.Done()
		logger.Info("automation scheduler stopped")
	}()

	// Register built-in system agent before channel manager so it appears
	// in the agent list exposed to Telegram/Discord.
	registerSystemAgent(ctx, cfg, runnerSvc, agentRepo, cronScheduler, cronExecRepo)

	modelInfos := internalagent.AllModelAliases(cfg.ModelProviders)
	modelNames := make([]string, len(modelInfos))
	for i, m := range modelInfos {
		modelNames[i] = m.Alias
	}

	mgr, err := channel.NewManager(ctx, channelRepo, runnerSvc, rdb, modelNames)
	if err != nil {
		logger.Error("failed to create channel manager", "err", err)
		return nil, err
	}

	logger.Info("starting channel manager in background")
	go mgr.Start(ctx)

	return &BootstrapResult{
		RunnerSvc:           runnerSvc,
		SessionSvc:          sessionSvc,
		CronScheduler:       cronScheduler,
		CronRepo:            cronExecRepo,
		CronJobRepo:         cronJobRepo,
		AutomationEngine:    automationEngine,
		AutomationScheduler: automationScheduler,
		AutomationDefRepo:   automationDefRepo,
		AutomationRunRepo:   automationRunRepo,
		AutomationStepRepo:  automationStepRepo,
		ChannelMgr:          mgr,
		MongoDB:             db,
		Redis:               rdb,
		AuthRepo:            authRepo,
		OAuthStateRepo:      oauthStateRepo,
		OAuthProviders:      oauthProviders,
		APITokenRepo:        tokenRepo,
		InvocationRepo:      invRepo,
		ForumRepo:           forumRepo,
		WorkspaceRepo:       wsRepo,
		MCPOAuthRepo:        oauthRepo,
		MCPOAuthSvc:         oauthSvc,
		MCPAuthResolver:     mcpAuthResolver,
		AgentFileRepo:       fileRepo,
		AgentFileMaxBytes:   cfg.AgentFiles.EffectiveMaxFileBytes(),
		LangfuseHost:        cfg.Langfuse.Host,
		SessionCounter:      sessionSvc.CountSessions,
	}, nil
}
