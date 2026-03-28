package channel

import (
	"context"
	"fmt"
	"sync"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/channel/telegram"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Manager manages all channel pollers.
type Manager struct {
	pollers []*telegram.Poller
}

// NewManager creates pollers for all enabled Telegram channels.
func NewManager(
	ctx context.Context,
	cfg *config.AppConfig,
	runnerSvc *runner.Service,
	selector *telegram.AgentSelector,
) (*Manager, error) {
	logger := log.FromContext(ctx)
	agentNames := runnerSvc.AgentNames()
	var pollers []*telegram.Poller

	logger.Info("initializing channel manager", "total_channels", len(cfg.Channels), "available_agents", agentNames)

	for i := range cfg.Channels {
		ch := &cfg.Channels[i]
		if ch.GetPlatform() != agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM {
			logger.Debug("skipping non-telegram channel", "channel", ch.GetName(), "platform", ch.GetPlatform().String())
			continue
		}
		if !ch.GetEnabled() {
			logger.Info("skipping disabled channel", "channel", ch.GetName())
			continue
		}
		if ch.GetTelegram().GetBotToken() == "" {
			logger.Warn("skipping channel with empty bot token", "channel", ch.GetName())
			continue
		}

		logger.Info("creating poller for channel",
			"channel", ch.GetName(),
			"default_agent", ch.GetAgentName(),
			"triggers", len(ch.GetTriggers()),
			"session_scope", ch.GetSession().GetScope().String(),
		)

		p, err := telegram.NewPoller(ch, runnerSvc, selector, agentNames)
		if err != nil {
			return nil, fmt.Errorf("creating poller for channel %q: %w", ch.GetName(), err)
		}
		pollers = append(pollers, p)
	}

	logger.Info("channel manager initialized", "active_pollers", len(pollers))
	return &Manager{pollers: pollers}, nil
}

// Start launches all pollers in goroutines. Blocks until ctx is cancelled.
func (m *Manager) Start(ctx context.Context) {
	logger := log.FromContext(ctx)

	if len(m.pollers) == 0 {
		logger.Info("no telegram channels configured, channel manager idle")
		return
	}

	var wg sync.WaitGroup
	for _, p := range m.pollers {
		wg.Add(1)
		go func(p *telegram.Poller) {
			defer wg.Done()
			p.Start(ctx)
		}(p)
	}

	logger.Info("all telegram pollers started", "count", len(m.pollers))
	wg.Wait()
	logger.Info("all telegram pollers stopped")
}
