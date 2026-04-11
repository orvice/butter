package bootstrap

import (
	"context"

	"butterfly.orx.me/core/log"
	"github.com/achetronic/adk-utils-go/plugin/langfuse"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	adkrunner "google.golang.org/adk/runner"
	"google.golang.org/adk/session"

	internalagent "go.orx.me/apps/butter/internal/agent"
	systemagent "go.orx.me/apps/butter/internal/agent/system"
	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	internalcron "go.orx.me/apps/butter/internal/cron"
	mongomemory "go.orx.me/apps/butter/internal/memory/mongo"
	"go.orx.me/apps/butter/internal/repo/configstore"
	"go.orx.me/apps/butter/internal/runner"
	mongosession "go.orx.me/apps/butter/internal/session/mongo"
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
	mongoURI := cfg.MongoURI
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	logger.Info("connecting to mongodb", "uri", mongoURI)

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Error("failed to connect to mongodb", "err", err)
		return nil, err
	}

	dbName := cfg.MongoDB
	if dbName == "" {
		dbName = "butter"
	}
	logger.Info("mongodb connected", "database", dbName)

	db := mongoClient.Database(dbName)

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
	redisAddr := cfg.RedisAddr
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	logger.Info("connecting to redis", "addr", redisAddr)

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.RedisPassword,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("redis ping failed, agent selection may not persist", "err", err)
	} else {
		logger.Info("redis connected")
	}

	// Setup Langfuse plugin if configured.
	var pluginConfig adkrunner.PluginConfig
	if cfg.Langfuse.IsEnabled() {
		logger.Info("setting up langfuse plugin")
		pc, shutdown, err := langfuse.Setup(&cfg.Langfuse)
		if err != nil {
			logger.Error("failed to setup langfuse", "err", err)
			return nil, err
		}
		pluginConfig = pc
		go func() {
			<-ctx.Done()
			_ = shutdown(context.Background())
		}()
		logger.Info("langfuse plugin enabled")
	}

	// Build runner service.
	logger.Info("building runner service", "agent_count", len(cfg.Agents))
	runnerSvc, err := runner.NewService(ctx, cfg.Agents, cfg.ModelProviders, cfg.MCPServerConfigs, cfg.RemoteAgents, sessionSvc, memorySvc, pluginConfig)
	if err != nil {
		logger.Error("failed to build runner service", "err", err)
		return nil, err
	}

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

	// Initialize cron scheduler (jobs are loaded from MongoDB).
	cronExecRepo := internalcron.NewMongoExecutionRepo(db)
	cronJobRepo := internalcron.NewMongoJobRepo(db)
	cronScheduler, err := internalcron.NewScheduler(ctx, runnerSvc, cronJobRepo, cronExecRepo)
	if err != nil {
		logger.Error("failed to create cron scheduler", "err", err)
		return nil, err
	}
	cronScheduler.Start()
	logger.Info("cron scheduler started")

	go func() {
		<-ctx.Done()
		stopCtx := cronScheduler.Stop()
		<-stopCtx.Done()
		logger.Info("cron scheduler stopped")
	}()

	// Register built-in system agent.
	if runnerSvc.HasAgent(systemagent.AgentName) {
		logger.Warn("user-configured agent conflicts with built-in system agent, skipping user agent", "name", systemagent.AgentName)
	}
	if cfg.SystemAgentModel != "" {
		sysAgent, err := systemagent.NewAgent(ctx, cfgStore, cronScheduler, cronExecRepo, cfg.SystemAgentModel, cfg.ModelProviders)
		if err != nil {
			logger.Error("failed to create system agent", "err", err)
		} else {
			runnerSvc.RegisterAgent(systemagent.AgentName, sysAgent)
			logger.Info("system agent registered", "model", cfg.SystemAgentModel)
		}
	} else {
		logger.Info("system agent disabled (no system_agent_model configured)")
	}

	return &BootstrapResult{
		RunnerSvc:     runnerSvc,
		SessionSvc:    sessionSvc,
		CronScheduler: cronScheduler,
		CronRepo:      cronExecRepo,
	}, nil
}
