package app

import (
	"context"

	"butterfly.orx.me/core/log"
	"go.mongodb.org/mongo-driver/v2/mongo"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/runtime/runner"
)

// startCron initializes the cron scheduler with MongoDB repos, starts it,
// and sets up graceful shutdown.
func startCron(ctx context.Context, db *mongo.Database, runnerSvc *runner.Service, groupRepo configrepo.NotifyGroupRepository) (*internalcron.Scheduler, internalcron.ExecutionRepo, internalcron.JobRepo, error) {
	logger := log.FromContext(ctx)

	cronExecRepo := internalcron.NewMongoExecutionRepo(db)
	cronJobRepo := internalcron.NewMongoJobRepo(db)
	cronScheduler, err := internalcron.NewScheduler(ctx, runnerSvc, cronJobRepo, cronExecRepo, groupRepo)
	if err != nil {
		logger.Error("failed to create cron scheduler", "err", err)
		return nil, nil, nil, err
	}
	cronScheduler.Start()
	logger.Info("cron scheduler started")

	go func() {
		<-ctx.Done()
		stopCtx := cronScheduler.Stop()
		<-stopCtx.Done()
		logger.Info("cron scheduler stopped")
	}()

	return cronScheduler, cronExecRepo, cronJobRepo, nil
}
