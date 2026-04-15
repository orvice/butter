package app

import (
	"context"

	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
)

type ConfigRuntime struct {
	store      *ConfigStore
	cfg        *config.AppConfig
	runnerSvc  *runner.Service
	channelMgr *channel.Manager
}

func NewConfigRuntime(store *ConfigStore, cfg *config.AppConfig) *ConfigRuntime {
	return &ConfigRuntime{
		store: store,
		cfg:   cfg,
	}
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
	return r.runnerSvc.ReloadProtoAgents(ctx, r.cfg.Agents, r.cfg.MCPServerConfigs, r.cfg.RemoteAgents)
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
