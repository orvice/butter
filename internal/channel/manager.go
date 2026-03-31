package channel

import (
	"context"
	"fmt"
	"sync"

	"butterfly.orx.me/core/log"
	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/channel/discord"
	"go.orx.me/apps/butter/internal/channel/telegram"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ChannelPoller is the common interface for platform-specific pollers.
type ChannelPoller interface {
	Start(ctx context.Context)
}

// Manager manages all channel pollers.
type Manager struct {
	pollers []ChannelPoller
}

// NewManager creates pollers for all enabled channels.
func NewManager(
	ctx context.Context,
	cfg *config.AppConfig,
	runnerSvc *runner.Service,
	rdb *redis.Client,
	modelNames []string,
) (*Manager, error) {
	logger := log.FromContext(ctx)
	agentNames := runnerSvc.AgentNames()
	var pollers []ChannelPoller

	logger.Info("initializing channel manager", "total_channels", len(cfg.Channels), "available_agents", agentNames)

	for i := range cfg.Channels {
		ch := &cfg.Channels[i]
		if !ch.GetEnabled() {
			logger.Info("skipping disabled channel", "channel", ch.GetName())
			continue
		}

		switch ch.GetPlatform() {
		case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM:
			if ch.GetTelegram().GetBotToken() == "" {
				logger.Warn("skipping telegram channel with empty bot token", "channel", ch.GetName())
				continue
			}
			tgSelector := telegram.NewAgentSelector(rdb)
			tgModelSelector := telegram.NewModelSelector(rdb)
			tgDebugToggle := telegram.NewDebugToggle(rdb)

			logger.Info("creating telegram poller",
				"channel", ch.GetName(),
				"default_agent", ch.GetAgentName(),
			)
			p, err := telegram.NewPoller(ch, runnerSvc, tgSelector, tgModelSelector, tgDebugToggle, agentNames, modelNames)
			if err != nil {
				return nil, fmt.Errorf("creating telegram poller for channel %q: %w", ch.GetName(), err)
			}
			pollers = append(pollers, p)

		case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD:
			if ch.GetDiscord().GetBotToken() == "" {
				logger.Warn("skipping discord channel with empty bot token", "channel", ch.GetName())
				continue
			}
			dcSelector := discord.NewAgentSelector(rdb)
			dcModelSelector := discord.NewModelSelector(rdb)
			dcDebugToggle := discord.NewDebugToggle(rdb)

			logger.Info("creating discord poller",
				"channel", ch.GetName(),
				"default_agent", ch.GetAgentName(),
			)
			p, err := discord.NewPoller(ch, runnerSvc, dcSelector, dcModelSelector, dcDebugToggle, agentNames, modelNames)
			if err != nil {
				return nil, fmt.Errorf("creating discord poller for channel %q: %w", ch.GetName(), err)
			}
			pollers = append(pollers, p)

		default:
			logger.Debug("skipping channel with unsupported platform", "channel", ch.GetName(), "platform", ch.GetPlatform().String())
		}
	}

	logger.Info("channel manager initialized", "active_pollers", len(pollers))
	return &Manager{pollers: pollers}, nil
}

// Start launches all pollers in goroutines. Blocks until ctx is cancelled.
func (m *Manager) Start(ctx context.Context) {
	logger := log.FromContext(ctx)

	if len(m.pollers) == 0 {
		logger.Info("no channels configured, channel manager idle")
		return
	}

	var wg sync.WaitGroup
	for _, p := range m.pollers {
		wg.Add(1)
		go func(p ChannelPoller) {
			defer wg.Done()
			p.Start(ctx)
		}(p)
	}

	logger.Info("all channel pollers started", "count", len(m.pollers))
	wg.Wait()
	logger.Info("all channel pollers stopped")
}
