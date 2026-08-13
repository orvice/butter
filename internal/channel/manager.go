// Package channel holds what remains of the legacy generic AgentChannel
// runtime after the Telegram cutover (issue #264/#273).
//
// Nothing here starts a transport any more. Telegram runs on Telegram
// Channels and Destinations (`internal/runtime/telegram`), and Discord is
// unsupported in this release pending a redesign. The Manager survives only
// to make that visible: on start and on every reload it reports the legacy
// records still in the database, so an operator can see what will not run
// rather than discovering it when a bot goes quiet.
//
// Legacy records are deliberately not migrated. A generic AgentChannel has no
// exact address — it has a chat allowlist — so there is no honest way to
// derive the Destination it should become, and guessing would send an agent's
// replies somewhere nobody chose.
package channel

import (
	"context"
	"sync"

	"butterfly.orx.me/core/log"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Manager reports legacy AgentChannel records. It starts nothing.
type Manager struct {
	repo configrepo.ChannelRepository

	mu      sync.Mutex
	started bool
}

// NewManager creates the legacy channel reporter.
func NewManager(ctx context.Context, repo configrepo.ChannelRepository) (*Manager, error) {
	m := &Manager{repo: repo}
	m.report(ctx)
	return m, nil
}

// Start blocks until ctx is cancelled. It launches no transports.
func (m *Manager) Start(ctx context.Context) {
	m.mu.Lock()
	m.started = true
	m.mu.Unlock()
	<-ctx.Done()
}

// Reload re-reports legacy records after a configuration change.
func (m *Manager) Reload(ctx context.Context) error {
	m.report(ctx)
	return nil
}

// report logs every enabled legacy record as unsupported.
//
// It logs rather than failing startup on purpose: a deployment upgrading with
// legacy rows still in its database must come up healthy, serve the Dashboard,
// and let an operator migrate deliberately — refusing to boot would take the
// whole service down over configuration nobody is using yet.
func (m *Manager) report(ctx context.Context) {
	logger := log.FromContext(ctx)
	channels, err := m.repo.ListChannelsAcrossWorkspaces(ctx)
	if err != nil {
		logger.Warn("could not list legacy agent channels", "err", err)
		return
	}

	unsupported := 0
	for _, ch := range channels {
		if !ch.GetEnabled() {
			continue
		}
		unsupported++
		logger.Warn("legacy agent channel is not started; it is unsupported in this release",
			"channel", ch.GetName(),
			"workspace_id", ch.GetWorkspaceId(),
			"platform", ch.GetPlatform().String(),
			"migrate_to", migrationHint(ch.GetPlatform()),
		)
	}
	if unsupported > 0 {
		logger.Warn("legacy agent channels remain in the database and will not run",
			"count", unsupported,
			"action", "recreate them as telegram channels and destinations in the dashboard")
	}
}

// migrationHint tells an operator where the capability went.
func migrationHint(platform agentsv1.AgentChannelPlatform) string {
	switch platform {
	case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM:
		return "telegram channel + destination"
	case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD:
		return "unsupported in this release"
	default:
		return "unsupported"
	}
}
