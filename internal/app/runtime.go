package app

import (
	"context"

	"butterfly.orx.me/core/log"
	"github.com/achetronic/adk-utils-go/plugin/langfuse"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	adkrunner "google.golang.org/adk/runner"

	"go.orx.me/apps/butter/internal/config"
)

// connectMongo establishes a connection to MongoDB and returns the database handle.
func connectMongo(ctx context.Context, cfg *config.AppConfig) (*mongo.Database, error) {
	logger := log.FromContext(ctx)

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

	return mongoClient.Database(dbName), nil
}

// connectRedis establishes a connection to Redis.
func connectRedis(ctx context.Context, cfg *config.AppConfig) *redis.Client {
	logger := log.FromContext(ctx)

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
		logger.Warn("redis ping failed, auth sessions and agent selection may fail", "err", err)
	} else {
		logger.Info("redis connected")
	}

	return rdb
}

// setupLangfuse initializes the Langfuse plugin if configured.
func setupLangfuse(ctx context.Context, cfg *config.AppConfig) (adkrunner.PluginConfig, error) {
	logger := log.FromContext(ctx)

	var pluginConfig adkrunner.PluginConfig
	if cfg.Langfuse.IsEnabled() {
		logger.Info("setting up langfuse plugin")
		pc, shutdown, err := langfuse.Setup(&cfg.Langfuse)
		if err != nil {
			logger.Error("failed to setup langfuse", "err", err)
			return pluginConfig, err
		}
		pluginConfig = pc
		go func() {
			<-ctx.Done()
			_ = shutdown(context.Background())
		}()
		logger.Info("langfuse plugin enabled")
	}
	return pluginConfig, nil
}
