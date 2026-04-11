package app

import (
	"context"

	"butterfly.orx.me/core/log"

	systemagent "go.orx.me/apps/butter/internal/agent/system"
	"go.orx.me/apps/butter/internal/config"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/runtime/runner"
	configstore "go.orx.me/apps/butter/internal/store/config"
)

// registerSystemAgent registers the built-in system agent if configured.
func registerSystemAgent(ctx context.Context, cfg *config.AppConfig, runnerSvc *runner.Service, cfgStore *configstore.Store, cronScheduler *internalcron.Scheduler, cronExecRepo internalcron.ExecutionRepo) {
	logger := log.FromContext(ctx)

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
}
