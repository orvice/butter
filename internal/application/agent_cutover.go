package application

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"connectrpc.com/connect"

	internalagent "go.orx.me/apps/butter/internal/agent"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	forumrepo "go.orx.me/apps/butter/internal/repo/forum"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	internalautomation "go.orx.me/apps/butter/internal/runtime/automation"
	internalcron "go.orx.me/apps/butter/internal/runtime/cron"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// AgentCutoverSources aggregates the read-only data the final cutover
// verifier (issue #241) inspects. Agents is required; every other source is
// optional — a nil source simply skips its checks, so partial wirings (tests,
// slim deployments) still verify what they can see.
type AgentCutoverSources struct {
	Agents      configrepo.AgentRepository
	Channels    configrepo.ChannelRepository
	CronJobs    internalcron.JobRepo
	Automations internalautomation.DefinitionRepo
	Forum       forumrepo.Repository
	Workspaces  workspacerepo.Repository
	// ReservedName reports whether a runtime name is reserved by a built-in
	// agent (the runner refuses to register such agents). Nil skips the check.
	ReservedName func(name string) bool
}

// SetCutoverSources wires the consumer-record sources for VerifyAgentIDCutover.
func (s *AgentServiceServer) SetCutoverSources(src AgentCutoverSources) {
	s.cutoverSources = &src
}

// VerifyAgentIDCutover runs the read-only final cutover verifier across every
// workspace. Global admins only: findings span workspaces the caller may not
// be a member of.
func (s *AgentServiceServer) VerifyAgentIDCutover(ctx context.Context, _ *connect.Request[agentsv1.VerifyAgentIDCutoverRequest]) (*connect.Response[agentsv1.VerifyAgentIDCutoverResponse], error) {
	if err := requireGlobalAdmin(ctx); err != nil {
		return nil, err
	}
	src := AgentCutoverSources{Agents: s.repo}
	if s.cutoverSources != nil {
		src = *s.cutoverSources
	}
	if s.runnerSvc != nil && src.ReservedName == nil {
		src.ReservedName = s.runnerSvc.IsReservedAgentName
	}
	findings, err := RunAgentIDCutoverVerifier(ctx, src)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.VerifyAgentIDCutoverResponse{
		Passed:   len(findings) == 0,
		Findings: findings,
	}), nil
}

// RunAgentIDCutoverVerifier executes every cutover check and returns the
// findings sorted by (workspace, check, entity). It never mutates anything.
func RunAgentIDCutoverVerifier(ctx context.Context, src AgentCutoverSources) ([]*agentsv1.AgentIDCutoverFinding, error) {
	if src.Agents == nil {
		return nil, errors.New("agent cutover verifier: agent repository is required")
	}
	agents, err := src.Agents.ListAgentsAcrossWorkspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("list agents across workspaces: %w", err)
	}

	findings := agentCutoverFindings(agents, src.ReservedName)

	if src.Channels != nil {
		channels, err := src.Channels.ListChannelsAcrossWorkspaces(ctx)
		if err != nil {
			return nil, fmt.Errorf("list channels across workspaces: %w", err)
		}
		for _, ch := range channels {
			if ch.GetAgentName() != "" && ch.GetAgentId() == "" {
				findings = append(findings, finding(ch.GetWorkspaceId(), "channel_agent_id_missing", "channel", ch.GetName(),
					fmt.Sprintf("channel %q references agent by legacy name %q without an agent_id", ch.GetName(), ch.GetAgentName())))
			}
		}
	}

	if src.CronJobs != nil {
		jobs, err := src.CronJobs.ListAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("list cron jobs: %w", err)
		}
		for _, job := range jobs {
			if job.GetAgentId() == "" {
				findings = append(findings, finding(job.GetWorkspaceId(), "cron_agent_id_missing", "cron_job", job.GetName(),
					fmt.Sprintf("cron job %q (agent_name %q) carries no agent_id; the scheduler will reject it", job.GetName(), job.GetAgentName())))
			}
		}
	}

	if src.Automations != nil {
		automations, err := src.Automations.ListAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("list automations: %w", err)
		}
		for _, a := range automations {
			for _, step := range a.GetSteps() {
				inv := step.GetInvokeAgent()
				if inv == nil || inv.GetAgentId() != "" {
					continue
				}
				findings = append(findings, finding(a.GetWorkspaceId(), "automation_agent_id_missing", "automation", a.GetName(),
					fmt.Sprintf("automation %q step %q (agent_name %q) carries no agent_id; the engine will reject it", a.GetName(), step.GetName(), inv.GetAgentName())))
			}
		}
	}

	if src.Forum != nil && src.Workspaces != nil {
		threadFindings, err := forumCutoverFindings(ctx, src.Forum, src.Workspaces)
		if err != nil {
			return nil, err
		}
		findings = append(findings, threadFindings...)
	}

	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.GetWorkspaceId() != b.GetWorkspaceId() {
			return a.GetWorkspaceId() < b.GetWorkspaceId()
		}
		if a.GetCheck() != b.GetCheck() {
			return a.GetCheck() < b.GetCheck()
		}
		return a.GetEntity() < b.GetEntity()
	})
	return findings, nil
}

func finding(ws, check, kind, entity, detail string) *agentsv1.AgentIDCutoverFinding {
	return &agentsv1.AgentIDCutoverFinding{
		WorkspaceId: ws,
		Check:       check,
		EntityKind:  kind,
		Entity:      entity,
		Detail:      detail,
	}
}

// agentCutoverFindings runs the agent-scoped checks over the flat
// cross-workspace agent list.
func agentCutoverFindings(agents []*agentsv1.Agent, reserved func(string) bool) []*agentsv1.AgentIDCutoverFinding {
	var findings []*agentsv1.AgentIDCutoverFinding

	// Index agent IDs per workspace for duplicate and resolution checks.
	idsByWorkspace := make(map[string]map[string]int)
	for _, a := range agents {
		ws, id := a.GetWorkspaceId(), a.GetAgentId()
		if id == "" {
			continue
		}
		if idsByWorkspace[ws] == nil {
			idsByWorkspace[ws] = make(map[string]int)
		}
		idsByWorkspace[ws][id]++
	}

	// entityRef prefers the agent_id and falls back to the runtime name so a
	// finding always points at something.
	entityRef := func(a *agentsv1.Agent) string {
		if a.GetAgentId() != "" {
			return a.GetAgentId()
		}
		return a.GetName()
	}

	for _, a := range agents {
		ws := a.GetWorkspaceId()

		switch {
		case a.GetAgentId() == "":
			findings = append(findings, finding(ws, "agent_id_missing", "agent", a.GetName(),
				fmt.Sprintf("agent %q has no agent_id", a.GetName())))
		default:
			if err := internalagent.ValidateAgentID(a.GetAgentId()); err != nil {
				findings = append(findings, finding(ws, "agent_id_invalid", "agent", entityRef(a),
					fmt.Sprintf("agent %q has an invalid agent_id %q: %v", a.GetName(), a.GetAgentId(), err)))
			}
			if idsByWorkspace[ws][a.GetAgentId()] > 1 {
				findings = append(findings, finding(ws, "agent_id_duplicate", "agent", a.GetAgentId(),
					fmt.Sprintf("agent_id %q is used by %d agents in workspace %q", a.GetAgentId(), idsByWorkspace[ws][a.GetAgentId()], ws)))
			}
		}

		if len(a.GetSubAgents()) > 0 {
			findings = append(findings, finding(ws, "embedded_sub_agents", "agent", entityRef(a),
				fmt.Sprintf("agent %q still embeds %d legacy sub_agents; children must be independent agents referenced via child_agent_ids", a.GetName(), len(a.GetSubAgents()))))
		}

		if a.GetLifecycleStatus() == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED {
			findings = append(findings, finding(ws, "lifecycle_migration_required", "agent", entityRef(a),
				fmt.Sprintf("agent %q is still MIGRATION_REQUIRED", a.GetName())))
		}

		for _, childID := range a.GetChildAgentIds() {
			if idsByWorkspace[ws][childID] == 0 {
				findings = append(findings, finding(ws, "child_agent_id_unresolved", "agent", entityRef(a),
					fmt.Sprintf("agent %q references child_agent_id %q which does not resolve in workspace %q", a.GetName(), childID, ws)))
			}
		}

		for _, node := range a.GetConfig().GetWorkflow().GetNodes() {
			if node.GetKind() != agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT {
				continue
			}
			if node.GetAgentId() == "" {
				findings = append(findings, finding(ws, "workflow_legacy_agent_ref", "agent", entityRef(a),
					fmt.Sprintf("agent %q workflow node %q references its target by legacy name %q instead of agent_id", a.GetName(), node.GetName(), node.GetAgent())))
				continue
			}
			if idsByWorkspace[ws][node.GetAgentId()] == 0 {
				findings = append(findings, finding(ws, "workflow_agent_id_unresolved", "agent", entityRef(a),
					fmt.Sprintf("agent %q workflow node %q references agent_id %q which does not resolve in workspace %q", a.GetName(), node.GetName(), node.GetAgentId(), ws)))
			}
		}
	}

	// Runtime-name constraint: the runner registers runnable agents under
	// their bare name and requires it to be unique across workspaces.
	runnable := func(a *agentsv1.Agent) bool {
		st := a.GetLifecycleStatus()
		return st == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE ||
			st == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_UNSPECIFIED
	}
	byName := make(map[string][]*agentsv1.Agent)
	for _, a := range agents {
		if runnable(a) {
			byName[a.GetName()] = append(byName[a.GetName()], a)
		}
	}
	for name, group := range byName {
		if len(group) > 1 {
			workspaces := make([]string, 0, len(group))
			for _, a := range group {
				workspaces = append(workspaces, a.GetWorkspaceId())
			}
			sort.Strings(workspaces)
			for _, a := range group {
				findings = append(findings, finding(a.GetWorkspaceId(), "runtime_name_conflict", "agent", entityRef(a),
					fmt.Sprintf("runtime name %q is used by %d runnable agents (workspaces %v); the runner requires globally unique names", name, len(group), workspaces)))
			}
		}
		if reserved != nil && reserved(name) {
			for _, a := range group {
				findings = append(findings, finding(a.GetWorkspaceId(), "runtime_name_conflict", "agent", entityRef(a),
					fmt.Sprintf("runtime name %q is reserved by a built-in agent; the runner will not register it", name)))
			}
		}
	}

	return findings
}

// forumCutoverFindings pages every workspace's threads looking for records
// that reference agents only by legacy names.
func forumCutoverFindings(ctx context.Context, forum forumrepo.Repository, workspaces workspacerepo.Repository) ([]*agentsv1.AgentIDCutoverFinding, error) {
	wss, err := workspaces.ListWorkspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	var findings []*agentsv1.AgentIDCutoverFinding
	for _, ws := range wss {
		token := ""
		for {
			threads, next, _, err := forum.ListThreads(ctx, forumrepo.ThreadListFilter{WorkspaceID: ws.GetId()}, 200, token)
			if err != nil {
				return nil, fmt.Errorf("list forum threads (workspace %q): %w", ws.GetId(), err)
			}
			for _, t := range threads {
				if len(t.GetAgentNames()) > 0 && len(t.GetAgentIds()) == 0 {
					findings = append(findings, finding(ws.GetId(), "forum_agent_id_missing", "forum_thread", t.GetId(),
						fmt.Sprintf("forum thread %q references agents by legacy names %v without agent_ids", t.GetTitle(), t.GetAgentNames())))
				}
			}
			if next == "" {
				break
			}
			token = next
		}
	}
	return findings, nil
}
