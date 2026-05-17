package app

import (
	"context"
	"strings"

	"butterfly.orx.me/core/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/adk/session"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	apitokenmemory "go.orx.me/apps/butter/internal/repo/apitoken/memory"
	apitokenmongo "go.orx.me/apps/butter/internal/repo/apitoken/mongo"
	"go.orx.me/apps/butter/internal/repo/auth"
	authmongo "go.orx.me/apps/butter/internal/repo/auth/mongo"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	invocationmongo "go.orx.me/apps/butter/internal/repo/invocation/mongo"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	workspacemongo "go.orx.me/apps/butter/internal/repo/workspace/mongo"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	mongomemory "go.orx.me/apps/butter/internal/runtime/memory/mongo"
	"go.orx.me/apps/butter/internal/runtime/runner"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
)

// BootstrapResult holds the services created during bootstrap.
type BootstrapResult struct {
	RunnerSvc      *runner.Service
	SessionSvc     session.Service
	CronScheduler  *internalcron.Scheduler
	CronRepo       internalcron.ExecutionRepo
	CronJobRepo    internalcron.JobRepo
	ChannelMgr     *channel.Manager
	MongoDB        *mongo.Database
	Redis          *redis.Client
	AuthRepo       auth.Repository
	APITokenRepo   apitoken.Repository
	InvocationRepo invocation.Repository
	WorkspaceRepo  workspacerepo.Repository
	LangfuseHost   string
	SessionCounter func(ctx context.Context) (int64, error)
}

// StartChannels initializes MongoDB, Redis, runner service, channel manager,
// and cron scheduler. It returns the bootstrap result.
// agentRepo is the shared agent repository used by the system agent for agent queries.
func StartChannels(ctx context.Context, cfg *config.AppConfig, agentRepo configrepo.AgentRepository, channelRepo configrepo.ChannelRepository, daemonRegistry *daemon.Registry) (*BootstrapResult, error) {
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
		authRepo  auth.Repository
		tokenRepo apitoken.Repository
		invRepo   invocation.Repository
		wsRepo    workspacerepo.Repository
	)
	authRepo = authmongo.New(db)
	if err := application.BootstrapInitialAdmin(ctx, authRepo, cfg.Auth); err != nil {
		logger.Error("failed to initialize auth", "err", err)
		return nil, err
	}
	if strings.ToLower(strings.TrimSpace(cfg.StorageBackend)) == "mongo" {
		tokenRepo = apitokenmongo.New(db)
		invRepo = invocationmongo.New(db)
		wsRepo = workspacemongo.New(db)
	} else {
		tokenRepo = apitokenmemory.New()
		invRepo = invocationmemory.New()
		wsRepo = workspacememory.New()
	}
	if err := wsRepo.EnsureIndexes(ctx); err != nil {
		logger.Error("failed to create workspace indexes", "err", err)
		return nil, err
	}
	if err := application.BootstrapDefaultWorkspace(ctx, wsRepo, authRepo); err != nil {
		logger.Error("failed to bootstrap default workspace", "err", err)
		return nil, err
	}

	// Setup Langfuse plugin if configured.
	pluginConfig, err := setupLangfuse(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Build runner service.
	logger.Info("building runner service", "agent_count", len(cfg.Agents))
	runnerSvc, err := runner.NewService(ctx, cfg.Agents, cfg.ModelProviders, cfg.MCPServerConfigs, cfg.RemoteAgents, daemonRegistry, sessionSvc, memorySvc, pluginConfig)
	if err == nil {
		runnerSvc.SetInvocationRecorder(invRepo)
	}
	if err != nil {
		logger.Error("failed to build runner service", "err", err)
		return nil, err
	}

	// Initialize cron scheduler.
	cronScheduler, cronExecRepo, cronJobRepo, err := startCron(ctx, db, runnerSvc)
	if err != nil {
		return nil, err
	}

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
		RunnerSvc:      runnerSvc,
		SessionSvc:     sessionSvc,
		CronScheduler:  cronScheduler,
		CronRepo:       cronExecRepo,
		CronJobRepo:    cronJobRepo,
		ChannelMgr:     mgr,
		MongoDB:        db,
		Redis:          rdb,
		AuthRepo:       authRepo,
		APITokenRepo:   tokenRepo,
		InvocationRepo: invRepo,
		WorkspaceRepo:  wsRepo,
		LangfuseHost:   cfg.Langfuse.Host,
		SessionCounter: sessionSvc.CountSessions,
	}, nil
}
