package channel

import (
	"context"
	"fmt"
	"net/http"
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

type WebhookPoller interface {
	WebhookHandler() http.Handler
}

type buildResult struct {
	pollers  []ChannelPoller
	webhooks map[string]http.Handler
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
	webhooks  map[string]http.Handler
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

func (m *Manager) buildPollers(ctx context.Context) (*buildResult, error) {
	logger := log.FromContext(ctx)
	channels, err := m.repo.ListChannelsAcrossWorkspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}

	var pollers []ChannelPoller
	webhooks := make(map[string]http.Handler)

	logger.Info("initializing channel manager", "total_channels", len(channels))

	for _, ch := range channels {
		if !ch.GetEnabled() {
			logger.Info("skipping disabled channel", "channel", ch.GetName())
			continue
		}

		// Each channel only sees agents from its own workspace; passing the
		// global list here would let workspace A's bot offer (and route to)
		// workspace B's agents.
		agentNames := m.runnerSvc.AgentNamesForWorkspace(ch.GetWorkspaceId())

		switch ch.GetPlatform() {
		case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM:
			if ch.GetTelegram().GetBotToken() == "" {
				logger.Warn("skipping telegram channel with empty bot token", "channel", ch.GetName())
				continue
			}

			logger.Info("creating telegram poller",
				"channel", ch.GetName(),
				"workspace", ch.GetWorkspaceId(),
				"default_agent", ch.GetAgentName(),
				"available_agents", agentNames,
			)
			p, err := m.telegramFactory(ch, m.runnerSvc, m.rdb, agentNames, m.modelNames)
			if err != nil {
				return nil, fmt.Errorf("creating telegram poller for channel %q: %w", ch.GetName(), err)
			}
			pollers = append(pollers, p)
			if ch.GetTelegram().GetWebhookUrl() != "" {
				wh, ok := p.(WebhookPoller)
				if !ok {
					return nil, fmt.Errorf("telegram poller for channel %q does not expose webhook handler", ch.GetName())
				}
				webhooks[ch.GetName()] = wh.WebhookHandler()
			}

		case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD:
			if ch.GetDiscord().GetBotToken() == "" {
				logger.Warn("skipping discord channel with empty bot token", "channel", ch.GetName())
				continue
			}

			logger.Info("creating discord poller",
				"channel", ch.GetName(),
				"workspace", ch.GetWorkspaceId(),
				"default_agent", ch.GetAgentName(),
				"available_agents", agentNames,
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
	return &buildResult{pollers: pollers, webhooks: webhooks}, nil
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
	built, err := m.buildPollers(ctx)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		m.webhooks = built.webhooks
		return nil
	}

	if m.runCancel != nil {
		m.runCancel()
		m.runWG.Wait()
	}

	m.webhooks = built.webhooks
	m.startPollersLocked(built.pollers)
	log.FromContext(ctx).Info("channel manager reloaded", "active_pollers", len(built.pollers))
	return nil
}

func (m *Manager) start(ctx context.Context) error {
	built, err := m.buildPollers(ctx)
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
	m.webhooks = built.webhooks
	m.startPollersLocked(built.pollers)

	logger := log.FromContext(ctx)
	if len(built.pollers) == 0 {
		logger.Info("no channels configured, channel manager idle")
	} else {
		logger.Info("all channel pollers started", "count", len(built.pollers))
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

// TelegramWebhookHandler returns the current HTTP handler for a webhook-backed
// Telegram channel.
func (m *Manager) TelegramWebhookHandler(channelName string) (http.Handler, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.webhooks[channelName]
	return h, ok
}

// RuntimeState describes the running state of a configured channel.
type RuntimeState int

const (
	RuntimeStateUnknown RuntimeState = iota
	RuntimeStateLive
	RuntimeStateDisabled
	RuntimeStateUnsupported
	RuntimeStateNotFound
)

// ChannelStatus returns the runtime state and a human-readable detail for the
// channel with the given name. Phase 1 only distinguishes live vs disabled vs
// not-configured; per-poller heartbeats arrive in a later phase.
func (m *Manager) ChannelStatus(ctx context.Context, name string) (RuntimeState, string, error) {
	channels, err := m.repo.ListChannelsAcrossWorkspaces(ctx)
	if err != nil {
		return RuntimeStateUnknown, "", err
	}
	for _, ch := range channels {
		if ch.GetName() != name {
			continue
		}
		if !ch.GetEnabled() {
			return RuntimeStateDisabled, "channel disabled in config", nil
		}
		switch ch.GetPlatform() {
		case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
			agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD:
			m.mu.Lock()
			started := m.started
			m.mu.Unlock()
			if !started {
				return RuntimeStateDisabled, "channel manager not started", nil
			}
			return RuntimeStateLive, "", nil
		default:
			return RuntimeStateUnsupported, "platform not supported by manager", nil
		}
	}
	return RuntimeStateNotFound, "channel not found", nil
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
