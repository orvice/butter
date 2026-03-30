package bootstrap

import (
	"context"

	"butterfly.orx.me/core/log"
	"github.com/achetronic/adk-utils-go/plugin/langfuse"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	adkrunner "google.golang.org/adk/runner"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/channel/telegram"
	"go.orx.me/apps/butter/internal/config"
	mongomemory "go.orx.me/apps/butter/internal/memory/mongo"
	"go.orx.me/apps/butter/internal/runner"
	mongosession "go.orx.me/apps/butter/internal/session/mongo"
)

// StartChannels initializes MongoDB, Redis, runner service, and channel manager,
// then starts polling in a background goroutine.
// It returns the runner service so callers (e.g. A2A handler) can use it.
func StartChannels(ctx context.Context, cfg *config.AppConfig) (*runner.Service, error) {
	logger := log.FromContext(ctx)

	if len(cfg.Channels) == 0 {
		logger.Info("no channels configured, skipping channel manager")
		return nil, nil
	}

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

	selector := telegram.NewAgentSelector(rdb)
	modelSelector := telegram.NewModelSelector(rdb)
	debugToggle := telegram.NewDebugToggle(rdb)

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

	// Collect model aliases for the /model command.
	modelInfos := internalagent.AllModelAliases(cfg.ModelProviders)
	modelNames := make([]string, len(modelInfos))
	for i, m := range modelInfos {
		modelNames[i] = m.Alias
	}

	// Build channel manager.
	mgr, err := channel.NewManager(ctx, cfg, runnerSvc, selector, modelSelector, debugToggle, modelNames)
	if err != nil {
		logger.Error("failed to create channel manager", "err", err)
		return nil, err
	}

	// Start channels in background.
	logger.Info("starting channel manager in background")
	go mgr.Start(ctx)

	return runnerSvc, nil
}
