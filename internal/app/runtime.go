package app

import (
	"context"

	"butterfly.orx.me/core/log"
	"butterfly.orx.me/core/store/s3"
	"github.com/achetronic/adk-utils-go/plugin/langfuse"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/adk/v2/artifact"
	adkrunner "google.golang.org/adk/v2/runner"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/agentfile"
	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	"go.orx.me/apps/butter/pkg/adkutils"
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

// registeredS3 resolves a store key against the butterfly `store.s3`
// registry. ok is false when the key is empty or no client is registered
// under it — callers then apply their own fallback.
func registeredS3(storeKey string) (client *awss3.Client, bucket string, ok bool) {
	if storeKey == "" {
		return nil, "", false
	}
	client = s3.GetClient(storeKey)
	bucket = s3.GetBucket(storeKey)
	if client == nil || bucket == "" {
		return nil, "", false
	}
	return client, bucket, true
}

// setupArtifactService builds the ADK artifact.Service from cfg.Artifact.
// Returns nil when the bucket is not configured or the referenced S3 client
// is not registered — ADK then runs without artifact persistence.
func setupArtifactService(ctx context.Context, cfg *config.AppConfig) artifact.Service {
	logger := log.FromContext(ctx)
	if !cfg.Artifact.Enabled() {
		logger.Info("artifact service disabled (artifact.s3_bucket not set)")
		return nil
	}
	client, bucket, ok := registeredS3(cfg.Artifact.S3Bucket)
	if !ok {
		logger.Warn("artifact service disabled: s3 client not registered",
			"store_key", cfg.Artifact.S3Bucket,
		)
		return nil
	}
	var opts []adkutils.Option
	if cfg.Artifact.KeyPrefix != "" {
		opts = append(opts, adkutils.WithKeyPrefix(cfg.Artifact.KeyPrefix))
	}
	logger.Info("artifact service enabled",
		"store_key", cfg.Artifact.S3Bucket,
		"bucket", bucket,
		"key_prefix", cfg.Artifact.KeyPrefix,
	)
	return adkutils.NewS3ArtifactService(bucket, client, opts...)
}

func setupAgentFileContentStore(ctx context.Context, cfg *config.AppConfig) agentfile.ContentStore {
	logger := log.FromContext(ctx)
	if cfg.AgentFiles.S3Bucket == "" {
		logger.Info("agent files content store using memory (agent_files.s3_bucket not set)")
		return agentfile.NewMemoryContentStore()
	}
	client, bucket, ok := registeredS3(cfg.AgentFiles.S3Bucket)
	if !ok {
		logger.Warn("agent files content store falling back to memory: s3 client not registered",
			"store_key", cfg.AgentFiles.S3Bucket,
		)
		return agentfile.NewMemoryContentStore()
	}
	logger.Info("agent files content store enabled",
		"store_key", cfg.AgentFiles.S3Bucket,
		"bucket", bucket,
		"key_prefix", cfg.AgentFiles.KeyPrefix,
	)
	return agentfile.NewS3ContentStore(bucket, client, cfg.AgentFiles.KeyPrefix)
}

// setupSkillContentStore builds the skill.ContentStore from cfg.Skills.
// Falls back to the in-memory store (with a warning) when no bucket is
// configured or the referenced S3 client is not registered, so local
// development needs zero infrastructure (issue #153).
func setupSkillContentStore(ctx context.Context, cfg *config.AppConfig) skillrepo.ContentStore {
	logger := log.FromContext(ctx)
	if cfg.Skills.S3Bucket == "" {
		logger.Warn("skills content store using memory (skills.s3_bucket not set); skill bodies will not survive restarts")
		return skillrepo.NewMemoryContentStore()
	}
	client, bucket, ok := registeredS3(cfg.Skills.S3Bucket)
	if !ok {
		logger.Warn("skills content store falling back to memory: s3 client not registered",
			"store_key", cfg.Skills.S3Bucket,
		)
		return skillrepo.NewMemoryContentStore()
	}
	logger.Info("skills content store enabled",
		"store_key", cfg.Skills.S3Bucket,
		"bucket", bucket,
		"key_prefix", cfg.Skills.KeyPrefix,
	)
	return skillrepo.NewS3ContentStore(bucket, client, cfg.Skills.KeyPrefix)
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
