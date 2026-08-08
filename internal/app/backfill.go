package app

import (
	"context"

	"butterfly.orx.me/core/log"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	internalautomation "go.orx.me/apps/butter/internal/runtime/automation"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
)

// backfillConsumerAgentIDs is a one-time startup reconciliation for the Agent
// ID cutover (issue #213). Channels, cron jobs, and automation invoke-agent
// steps written before the migration reference their agent only by the legacy
// agent_name. Now that every interface resolves agents strictly by agent_id,
// such records would stop resolving at runtime, so this fills in agent_id from
// the agent each name currently points at.
//
// It is idempotent: records that already carry an agent_id are skipped, and a
// name that no longer resolves to an assigned agent_id is left untouched (a
// warning is logged) rather than failing startup.
func backfillConsumerAgentIDs(
	ctx context.Context,
	agentRepo configrepo.AgentRepository,
	channelRepo configrepo.ChannelRepository,
	cronJobRepo internalcron.JobRepo,
	autoDefRepo internalautomation.DefinitionRepo,
) {
	// resolve maps (workspace, legacy name) -> assigned agent_id, memoized so a
	// workspace's agents are each looked up at most once.
	cache := map[string]string{}
	resolve := func(workspaceID, name string) string {
		if name == "" {
			return ""
		}
		key := workspaceID + "\x00" + name
		if id, ok := cache[key]; ok {
			return id
		}
		id := ""
		if a, err := agentRepo.GetAgent(ctx, workspaceID, name); err == nil {
			id = a.GetAgentId()
		}
		cache[key] = id
		return id
	}

	backfillChannels(ctx, channelRepo, resolve)
	backfillCronJobs(ctx, cronJobRepo, resolve)
	backfillAutomations(ctx, autoDefRepo, resolve)
}

type refResolver func(workspaceID, name string) string

func backfillChannels(ctx context.Context, repo configrepo.ChannelRepository, resolve refResolver) {
	logger := log.FromContext(ctx)
	channels, err := repo.ListChannelsAcrossWorkspaces(ctx)
	if err != nil {
		logger.Warn("agent_id backfill: list channels failed", "err", err)
		return
	}
	for _, ch := range channels {
		if ch.GetAgentId() != "" || ch.GetAgentName() == "" {
			continue
		}
		id := resolve(ch.GetWorkspaceId(), ch.GetAgentName())
		if id == "" {
			logger.Warn("agent_id backfill: channel agent unresolved",
				"channel", ch.GetName(), "workspace_id", ch.GetWorkspaceId(), "agent_name", ch.GetAgentName())
			continue
		}
		ch.AgentId = id
		if _, err := repo.UpdateChannel(ctx, ch.GetWorkspaceId(), ch); err != nil {
			logger.Warn("agent_id backfill: channel update failed",
				"channel", ch.GetName(), "workspace_id", ch.GetWorkspaceId(), "err", err)
			continue
		}
		logger.Info("agent_id backfill: channel", "channel", ch.GetName(), "workspace_id", ch.GetWorkspaceId(), "agent_id", id)
	}
}

func backfillCronJobs(ctx context.Context, repo internalcron.JobRepo, resolve refResolver) {
	logger := log.FromContext(ctx)
	jobs, err := repo.ListAll(ctx)
	if err != nil {
		logger.Warn("agent_id backfill: list cron jobs failed", "err", err)
		return
	}
	for _, job := range jobs {
		if job.GetAgentId() != "" || job.GetAgentName() == "" {
			continue
		}
		id := resolve(job.GetWorkspaceId(), job.GetAgentName())
		if id == "" {
			logger.Warn("agent_id backfill: cron job agent unresolved",
				"job", job.GetName(), "workspace_id", job.GetWorkspaceId(), "agent_name", job.GetAgentName())
			continue
		}
		job.AgentId = id
		if err := repo.Update(ctx, job); err != nil {
			logger.Warn("agent_id backfill: cron job update failed",
				"job", job.GetName(), "workspace_id", job.GetWorkspaceId(), "err", err)
			continue
		}
		logger.Info("agent_id backfill: cron job", "job", job.GetName(), "workspace_id", job.GetWorkspaceId(), "agent_id", id)
	}
}

func backfillAutomations(ctx context.Context, repo internalautomation.DefinitionRepo, resolve refResolver) {
	logger := log.FromContext(ctx)
	automations, err := repo.ListAll(ctx)
	if err != nil {
		logger.Warn("agent_id backfill: list automations failed", "err", err)
		return
	}
	for _, a := range automations {
		changed := false
		for _, step := range a.GetSteps() {
			inv := step.GetInvokeAgent()
			if inv == nil || inv.GetAgentId() != "" || inv.GetAgentName() == "" {
				continue
			}
			id := resolve(a.GetWorkspaceId(), inv.GetAgentName())
			if id == "" {
				logger.Warn("agent_id backfill: automation agent unresolved",
					"automation", a.GetName(), "workspace_id", a.GetWorkspaceId(), "step", step.GetName(), "agent_name", inv.GetAgentName())
				continue
			}
			inv.AgentId = id
			changed = true
		}
		if !changed {
			continue
		}
		if err := repo.Update(ctx, a); err != nil {
			logger.Warn("agent_id backfill: automation update failed",
				"automation", a.GetName(), "workspace_id", a.GetWorkspaceId(), "err", err)
			continue
		}
		logger.Info("agent_id backfill: automation", "automation", a.GetName(), "workspace_id", a.GetWorkspaceId())
	}
}
