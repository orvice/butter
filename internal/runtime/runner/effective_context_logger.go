package runner

import (
	"butterfly.orx.me/core/log"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/plugin"
	adkrunner "google.golang.org/adk/v2/runner"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type contextGuardPolicy struct {
	strategy      agentsv1.ContextGuardStrategy
	agentOverride int
}

// newEffectiveContextWindowLoggerPlugin logs the callback-time resolution
// immediately before ContextGuard evaluates the same request.
func newEffectiveContextWindowLoggerPlugin(registry *configuredModelRegistry, policies map[string]contextGuardPolicy) (adkrunner.PluginConfig, error) {
	p, err := plugin.New(plugin.Config{
		Name: "effective_context_window_logger",
		BeforeModelCallback: llmagent.BeforeModelCallback(func(ctx agent.Context, req *model.LLMRequest) (*model.LLMResponse, error) {
			policy, ok := policies[ctx.AgentName()]
			if !ok || req == nil {
				return nil, nil
			}

			resolution := registry.Resolve(req.Model, policy.agentOverride)
			log.FromContext(ctx).Info("effective context window resolved",
				"agent", ctx.AgentName(),
				"strategy", policy.strategy.String(),
				"selected_model_id", resolution.SelectedModelID,
				"metadata_source", string(resolution.MetadataSource),
				"configured_agent_override", resolution.ConfiguredAgentOverride,
				"configured_model_capacity", resolution.ConfiguredModelCapacity,
				"effective_context_window", resolution.EffectiveContextWindow,
			)
			return nil, nil
		}),
	})
	if err != nil {
		return adkrunner.PluginConfig{}, err
	}
	return adkrunner.PluginConfig{Plugins: []*plugin.Plugin{p}}, nil
}
