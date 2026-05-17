package app

import (
	"context"

	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type ConfigRuntime struct {
	store      configSyncer
	cfg        *config.AppConfig
	runnerSvc  protoAgentReloader
	channelMgr channelReloader
}

func NewConfigRuntime(store *ConfigStore, cfg *config.AppConfig) *ConfigRuntime {
	return &ConfigRuntime{
		store: store,
		cfg:   cfg,
	}
}

type configSyncer interface {
	SyncToConfig(ctx context.Context, cfg *config.AppConfig) error
}

type protoAgentReloader interface {
	ReloadProtoAgents(ctx context.Context, agents []agentsv1.Agent, providers []agentsv1.ModelProvider, mcpRegistry []agentsv1.MCPServer, remoteAgentRegistry []agentsv1.RemoteAgent) error
}

type channelReloader interface {
	Reload(ctx context.Context) error
}

func (r *ConfigRuntime) SetRunnerService(runnerSvc *runner.Service) {
	r.runnerSvc = runnerSvc
}

func (r *ConfigRuntime) SetChannelManager(channelMgr *channel.Manager) {
	r.channelMgr = channelMgr
}

func (r *ConfigRuntime) Sync(ctx context.Context) error {
	if r.store == nil || r.cfg == nil {
		return nil
	}
	return r.store.SyncToConfig(ctx, r.cfg)
}

func (r *ConfigRuntime) ReloadRunner(ctx context.Context) error {
	if err := r.Sync(ctx); err != nil {
		return err
	}
	if r.runnerSvc == nil {
		return nil
	}
	if err := r.runnerSvc.ReloadProtoAgents(ctx, r.cfg.Agents, r.cfg.ModelProviders, r.cfg.MCPServerConfigs, r.cfg.RemoteAgents); err != nil {
		return err
	}
	if r.channelMgr == nil {
		return nil
	}
	return r.channelMgr.Reload(ctx)
}

func (r *ConfigRuntime) ReloadChannels(ctx context.Context) error {
	if err := r.Sync(ctx); err != nil {
		return err
	}
	if r.channelMgr == nil {
		return nil
	}
	return r.channelMgr.Reload(ctx)
}
