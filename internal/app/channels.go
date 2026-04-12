package app

import (
	"context"

	"butterfly.orx.me/core/log"
	"google.golang.org/adk/session"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	mongomemory "go.orx.me/apps/butter/internal/runtime/memory/mongo"
	"go.orx.me/apps/butter/internal/runtime/runner"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
	configstore "go.orx.me/apps/butter/internal/store/config"
)

// BootstrapResult holds the services created during bootstrap.
type BootstrapResult struct {
	RunnerSvc     *runner.Service
	SessionSvc    session.Service
	CronScheduler *internalcron.Scheduler
	CronRepo      internalcron.ExecutionRepo
}

// StartChannels initializes MongoDB, Redis, runner service, channel manager,
// and cron scheduler. It returns the bootstrap result.
// cfgStore is the shared config store used by the system agent for agent queries.
func StartChannels(ctx context.Context, cfg *config.AppConfig, cfgStore *configstore.Store) (*BootstrapResult, error) {
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

	// Setup Langfuse plugin if configured.
	pluginConfig, err := setupLangfuse(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Build runner service.
	logger.Info("building runner service", "agent_count", len(cfg.Agents))
	runnerSvc, err := runner.NewService(ctx, cfg.Agents, cfg.ModelProviders, cfg.MCPServerConfigs, cfg.RemoteAgents, sessionSvc, memorySvc, pluginConfig)
	if err != nil {
		logger.Error("failed to build runner service", "err", err)
		return nil, err
	}

	// Initialize cron scheduler.
	cronScheduler, cronExecRepo, err := startCron(ctx, db, runnerSvc)
	if err != nil {
		return nil, err
	}

	// Register built-in system agent before channel manager so it appears
	// in the agent list exposed to Telegram/Discord.
	registerSystemAgent(ctx, cfg, runnerSvc, cfgStore, cronScheduler, cronExecRepo)

	// Start channels if configured.
	if len(cfg.Channels) > 0 {
		modelInfos := internalagent.AllModelAliases(cfg.ModelProviders)
		modelNames := make([]string, len(modelInfos))
		for i, m := range modelInfos {
			modelNames[i] = m.Alias
		}

		mgr, err := channel.NewManager(ctx, cfg, runnerSvc, rdb, modelNames)
		if err != nil {
			logger.Error("failed to create channel manager", "err", err)
			return nil, err
		}

		logger.Info("starting channel manager in background")
		go mgr.Start(ctx)
	} else {
		logger.Info("no channels configured, skipping channel manager")
	}

	return &BootstrapResult{
		RunnerSvc:     runnerSvc,
		SessionSvc:    sessionSvc,
		CronScheduler: cronScheduler,
		CronRepo:      cronExecRepo,
	}, nil
}
