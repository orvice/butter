package app

import (
	"context"
	"errors"
	"fmt"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/agentcontent"
	"go.orx.me/apps/butter/internal/channel"
	"go.orx.me/apps/butter/internal/config"
	agentcontentrepo "go.orx.me/apps/butter/internal/repo/agentcontent"
	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type ConfigRuntime struct {
	store      configSyncer
	cfg        *config.AppConfig
	runnerSvc  protoAgentReloader
	channelMgr channelReloader

	bindingRepo repobindingrepo.Repository
	contentRepo agentcontentrepo.Repository
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

// SetAgentContentRepos wires the repositories used to overlay Git-managed
// Agent Content onto DB-owned agent protos during runner reload.
func (r *ConfigRuntime) SetAgentContentRepos(bindingRepo repobindingrepo.Repository, contentRepo agentcontentrepo.Repository) {
	r.bindingRepo = bindingRepo
	r.contentRepo = contentRepo
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
	if err := r.applyActiveContent(ctx); err != nil {
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

// applyActiveContent overlays Git-managed Agent Content onto the DB-loaded
// agent protos. For each workspace with an active binding + content snapshot,
// the description/instruction/global_instruction fields are replaced.
func (r *ConfigRuntime) applyActiveContent(ctx context.Context) error {
	if r.bindingRepo == nil || r.contentRepo == nil || r.cfg == nil {
		return nil
	}
	return applyActiveContent(ctx, r.cfg.Agents, r.bindingRepo, r.contentRepo)
}

func applyActiveContent(ctx context.Context, agents []agentsv1.Agent, bindingRepo repobindingrepo.Repository, contentRepo agentcontentrepo.Repository) error {
	if bindingRepo == nil || contentRepo == nil {
		return nil
	}
	logger := log.FromContext(ctx)

	workspaces := make(map[string]struct{})
	for i := range agents {
		if ws := agents[i].GetWorkspaceId(); ws != "" {
			workspaces[ws] = struct{}{}
		}
	}

	var overlayErrs []error
	for ws := range workspaces {
		binding, err := bindingRepo.Get(ctx, ws)
		if err != nil {
			if errors.Is(err, repobindingrepo.ErrNotFound) {
				continue
			}
			overlayErrs = append(overlayErrs, fmt.Errorf("get repository binding for workspace %q: %w", ws, err))
			continue
		}
		if binding.GetActiveCommitSha() == "" {
			continue
		}
		snapshot, err := contentRepo.GetSnapshot(ctx, ws, binding.GetActiveCommitSha())
		if err != nil {
			overlayErrs = append(overlayErrs, fmt.Errorf("get active Agent Content snapshot for workspace %q at %q: %w", ws, binding.GetActiveCommitSha(), err))
			continue
		}
		if len(snapshot.Entries) == 0 {
			continue
		}
		agentcontent.ApplyToProto(agents, ws, snapshot.Entries)
		logger.Info("applied active agent content overlay",
			"workspace_id", ws,
			"commit_sha", snapshot.CommitSHA,
			"agents_overlaid", len(snapshot.Entries))
	}
	return errors.Join(overlayErrs...)
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
