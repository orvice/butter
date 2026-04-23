package channel

import (
	"context"
	"fmt"
	"sync"

	"butterfly.orx.me/core/log"
	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/channel/discord"
	"go.orx.me/apps/butter/internal/channel/telegram"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ChannelPoller is the common interface for platform-specific pollers.
type ChannelPoller interface {
	Start(ctx context.Context)
}

type pollerFactory func(
	channelCfg *agentsv1.AgentChannel,
	runnerSvc *runner.Service,
	rdb *redis.Client,
	agentNames []string,
	modelNames []string,
) (ChannelPoller, error)

// Manager manages all channel pollers.
type Manager struct {
	repo            configrepo.ChannelRepository
	runnerSvc       *runner.Service
	rdb             *redis.Client
	modelNames      []string
	telegramFactory pollerFactory
	discordFactory  pollerFactory

	mu        sync.Mutex
	parentCtx context.Context
	runCancel context.CancelFunc
	runWG     sync.WaitGroup
	started   bool
}

// NewManager creates a reloadable channel manager backed by the config repository.
func NewManager(
	ctx context.Context,
	repo configrepo.ChannelRepository,
	runnerSvc *runner.Service,
	rdb *redis.Client,
	modelNames []string,
) (*Manager, error) {
	m := &Manager{
		repo:       repo,
		runnerSvc:  runnerSvc,
		rdb:        rdb,
		modelNames: modelNames,
		telegramFactory: func(channelCfg *agentsv1.AgentChannel, runnerSvc *runner.Service, rdb *redis.Client, agentNames []string, modelNames []string) (ChannelPoller, error) {
			return telegram.NewPoller(channelCfg, runnerSvc, telegram.NewAgentSelector(rdb), telegram.NewModelSelector(rdb), telegram.NewDebugToggle(rdb), agentNames, modelNames)
		},
		discordFactory: func(channelCfg *agentsv1.AgentChannel, runnerSvc *runner.Service, rdb *redis.Client, agentNames []string, modelNames []string) (ChannelPoller, error) {
			return discord.NewPoller(channelCfg, runnerSvc, discord.NewAgentSelector(rdb), discord.NewModelSelector(rdb), discord.NewDebugToggle(rdb), agentNames, modelNames)
		},
	}

	if _, err := m.buildPollers(ctx); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) buildPollers(ctx context.Context) ([]ChannelPoller, error) {
	logger := log.FromContext(ctx)
	channels, err := m.repo.ListChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}

	agentNames := m.runnerSvc.AgentNames()
	var pollers []ChannelPoller

	logger.Info("initializing channel manager", "total_channels", len(channels), "available_agents", agentNames)

	for _, ch := range channels {
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

			logger.Info("creating telegram poller",
				"channel", ch.GetName(),
				"default_agent", ch.GetAgentName(),
			)
			p, err := m.telegramFactory(ch, m.runnerSvc, m.rdb, agentNames, m.modelNames)
			if err != nil {
				return nil, fmt.Errorf("creating telegram poller for channel %q: %w", ch.GetName(), err)
			}
			pollers = append(pollers, p)

		case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD:
			if ch.GetDiscord().GetBotToken() == "" {
				logger.Warn("skipping discord channel with empty bot token", "channel", ch.GetName())
				continue
			}

			logger.Info("creating discord poller",
				"channel", ch.GetName(),
				"default_agent", ch.GetAgentName(),
			)
			p, err := m.discordFactory(ch, m.runnerSvc, m.rdb, agentNames, m.modelNames)
			if err != nil {
				return nil, fmt.Errorf("creating discord poller for channel %q: %w", ch.GetName(), err)
			}
			pollers = append(pollers, p)

		default:
			logger.Debug("skipping channel with unsupported platform", "channel", ch.GetName(), "platform", ch.GetPlatform().String())
		}
	}

	logger.Info("channel manager initialized", "active_pollers", len(pollers))
	return pollers, nil
}

// Start launches all pollers in goroutines. Blocks until ctx is cancelled.
func (m *Manager) Start(ctx context.Context) {
	if err := m.start(ctx); err != nil {
		log.FromContext(ctx).Error("failed to start channel manager", "err", err)
		return
	}

	<-ctx.Done()
	m.stop()
}

// Reload refreshes the running pollers from the current config repository state.
func (m *Manager) Reload(ctx context.Context) error {
	pollers, err := m.buildPollers(ctx)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	if m.runCancel != nil {
		m.runCancel()
		m.runWG.Wait()
	}

	m.startPollersLocked(pollers)
	log.FromContext(ctx).Info("channel manager reloaded", "active_pollers", len(pollers))
	return nil
}

func (m *Manager) start(ctx context.Context) error {
	pollers, err := m.buildPollers(ctx)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}

	m.parentCtx = ctx
	m.started = true
	m.startPollersLocked(pollers)

	logger := log.FromContext(ctx)
	if len(pollers) == 0 {
		logger.Info("no channels configured, channel manager idle")
	} else {
		logger.Info("all channel pollers started", "count", len(pollers))
	}
	return nil
}

func (m *Manager) startPollersLocked(pollers []ChannelPoller) {
	runCtx, cancel := context.WithCancel(m.parentCtx)
	m.runCancel = cancel
	for _, poller := range pollers {
		m.runWG.Add(1)
		go func(p ChannelPoller) {
			defer m.runWG.Done()
			p.Start(runCtx)
		}(poller)
	}
}

func (m *Manager) stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return
	}

	if m.runCancel != nil {
		m.runCancel()
		m.runWG.Wait()
	}
	m.started = false
	m.runCancel = nil
	log.FromContext(m.parentCtx).Info("all channel pollers stopped")
}
