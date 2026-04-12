package app

import (
	"context"

	"butterfly.orx.me/core/log"

	internalagent "go.orx.me/apps/butter/internal/agent"
	systemagent "go.orx.me/apps/butter/internal/agent/system"
	"go.orx.me/apps/butter/internal/config"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/runtime/runner"
	configstore "go.orx.me/apps/butter/internal/store/config"
)

// registerSystemAgent registers the built-in system agent.
// If system_agent_model is configured, it uses that model directly.
// Otherwise, it uses the first available model from providers as default,
// and the agent will inherit the chat's model override at runtime.
func registerSystemAgent(ctx context.Context, cfg *config.AppConfig, runnerSvc *runner.Service, cfgStore *configstore.Store, cronScheduler *internalcron.Scheduler, cronExecRepo internalcron.ExecutionRepo) {
	logger := log.FromContext(ctx)

	if runnerSvc.HasAgent(systemagent.AgentName) {
		logger.Warn("user-configured agent conflicts with built-in system agent, skipping registration", "name", systemagent.AgentName)
		return
	}

	// Determine the default model for the system agent.
	model := cfg.SystemAgentModel
	if model == "" {
		// Use the first available model from providers.
		models := internalagent.AllModelAliases(cfg.ModelProviders)
		if len(models) == 0 {
			logger.Info("system agent disabled (no models available)")
			return
		}
		model = models[0].Alias
		logger.Info("system agent using default model from providers", "model", model)
	}

	sysAgent, err := systemagent.NewAgent(ctx, cfgStore, cronScheduler, cronExecRepo, model, cfg.ModelProviders)
	if err != nil {
		logger.Error("failed to create system agent", "err", err)
		return
	}

	builder := systemagent.NewBuilderFunc(cfgStore, cronScheduler, cronExecRepo, cfg.ModelProviders)
	runnerSvc.RegisterAgentWithBuilder(systemagent.AgentName, sysAgent, builder)
	logger.Info("system agent registered", "default_model", model)
}
