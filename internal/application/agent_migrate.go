package application

import (
	"context"
	"fmt"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func migrateAgentsV2(ctx context.Context, s *AgentServiceServer, req *connect.Request[agentsv1.MigrateAgentsV2Request]) (*connect.Response[agentsv1.MigrateAgentsV2Response], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireOwnerOrAdmin(ctx, wsID); err != nil {
		return nil, err
	}

	mode := req.Msg.GetMode()
	switch mode {
	case agentsv1.MigrateMode_MIGRATE_MODE_DRY_RUN:
		return migrateDryRun(ctx, s, wsID)
	case agentsv1.MigrateMode_MIGRATE_MODE_APPLY:
		return migrateApply(ctx, s, wsID)
	case agentsv1.MigrateMode_MIGRATE_MODE_VERIFY:
		return migrateVerify(ctx, s, wsID)
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("mode must be DRY_RUN, APPLY, or VERIFY"))
	}
}

func migrateDryRun(ctx context.Context, s *AgentServiceServer, wsID string) (*connect.Response[agentsv1.MigrateAgentsV2Response], error) {
	agents, err := s.repo.ListAgents(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}

	var results []*agentsv1.MigrateAgentResult
	var migrated, skipped, errCount int32

	for _, a := range agents {
		subs := a.GetSubAgents()
		if len(subs) == 0 {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:    a.GetName(),
				AgentId: a.GetAgentId(),
				Action:  "already_independent",
				Detail:  "no embedded sub_agents",
			})
			skipped++
			continue
		}

		if a.GetAgentId() == "" {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:   a.GetName(),
				Action: "missing_id",
				Detail: "root agent has no agent_id; assign one via AssignAgentID first",
			})
			errCount++
			continue
		}

		allChildrenReady := true
		for _, sub := range subs {
			if sub.GetAgentId() == "" {
				results = append(results, &agentsv1.MigrateAgentResult{
					Name:   sub.GetName(),
					Action: "missing_id",
					Detail: fmt.Sprintf("sub-agent of %q has no agent_id", a.GetName()),
				})
				errCount++
				allChildrenReady = false
			}
		}
		if allChildrenReady {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:    a.GetName(),
				AgentId: a.GetAgentId(),
				Action:  "expandable",
				Detail:  fmt.Sprintf("%d sub-agents will be expanded to independent agents", len(subs)),
			})
			migrated++
		}
	}

	return connect.NewResponse(&agentsv1.MigrateAgentsV2Response{
		Mode:     agentsv1.MigrateMode_MIGRATE_MODE_DRY_RUN,
		Results:  results,
		Total:    int32(len(agents)),
		Migrated: migrated,
		Skipped:  skipped,
		Errors:   errCount,
	}), nil
}

func migrateApply(ctx context.Context, s *AgentServiceServer, wsID string) (*connect.Response[agentsv1.MigrateAgentsV2Response], error) {
	logger := log.FromContext(ctx)
	agents, err := s.repo.ListAgents(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}

	existingByName := make(map[string]*agentsv1.Agent, len(agents))
	for _, a := range agents {
		existingByName[a.GetName()] = a
	}

	var results []*agentsv1.MigrateAgentResult
	var migrated, skipped, errCount int32

	for _, a := range agents {
		subs := a.GetSubAgents()
		if len(subs) == 0 {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:    a.GetName(),
				AgentId: a.GetAgentId(),
				Action:  "already_independent",
			})
			skipped++
			continue
		}

		if a.GetAgentId() == "" {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:   a.GetName(),
				Action: "missing_id",
				Detail: "root agent has no agent_id; skipping",
			})
			errCount++
			continue
		}

		childIDs := make([]string, 0, len(subs))
		allReady := true

		for _, sub := range subs {
			if sub.GetAgentId() == "" {
				results = append(results, &agentsv1.MigrateAgentResult{
					Name:   sub.GetName(),
					Action: "migration_required",
					Detail: fmt.Sprintf("sub-agent of %q missing agent_id", a.GetName()),
				})
				errCount++
				allReady = false
				continue
			}

			childIDs = append(childIDs, sub.GetAgentId())

			if _, exists := existingByName[sub.GetName()]; exists {
				results = append(results, &agentsv1.MigrateAgentResult{
					Name:    sub.GetName(),
					AgentId: sub.GetAgentId(),
					Action:  "skipped",
					Detail:  "already exists as independent agent",
				})
				continue
			}

			independent := proto.Clone(sub).(*agentsv1.Agent)
			independent.SubAgents = nil
			independent.DisplayName = sub.GetName()
			independent.LegacyName = sub.GetName()
			independent.WorkspaceId = wsID
			independent.LifecycleStatus = agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE

			if _, err := s.repo.CreateAgent(ctx, wsID, independent); err != nil {
				results = append(results, &agentsv1.MigrateAgentResult{
					Name:    sub.GetName(),
					AgentId: sub.GetAgentId(),
					Action:  "error",
					Detail:  fmt.Sprintf("failed to create independent agent: %v", err),
				})
				errCount++
				continue
			}
			existingByName[sub.GetName()] = independent
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:    sub.GetName(),
				AgentId: sub.GetAgentId(),
				Action:  "expanded",
				Detail:  fmt.Sprintf("expanded from parent %q", a.GetName()),
			})
			logger.Info("expanded sub-agent to independent", "parent", a.GetName(), "child", sub.GetName(), "child_agent_id", sub.GetAgentId())
		}

		if !allReady {
			continue
		}

		updated := proto.Clone(a).(*agentsv1.Agent)
		updated.SubAgents = nil
		updated.ChildAgentIds = childIDs
		updated.LifecycleStatus = agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE
		updated.LegacyName = a.GetName()
		if updated.DisplayName == "" {
			updated.DisplayName = a.GetName()
		}

		if wf := updated.GetConfig().GetWorkflow(); wf != nil {
			convertWorkflowNodeRefs(wf, subs)
		}

		if _, err := s.repo.UpdateAgent(ctx, wsID, updated); err != nil {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:    a.GetName(),
				AgentId: a.GetAgentId(),
				Action:  "error",
				Detail:  fmt.Sprintf("failed to update parent: %v", err),
			})
			errCount++
			continue
		}

		results = append(results, &agentsv1.MigrateAgentResult{
			Name:    a.GetName(),
			AgentId: a.GetAgentId(),
			Action:  "expanded",
			Detail:  fmt.Sprintf("cleared sub_agents, set child_agent_ids (%d children)", len(childIDs)),
		})
		migrated++
		logger.Info("parent agent migrated", "agent", a.GetName(), "children", len(childIDs))
	}

	if err := s.reloadRuntime(ctx); err != nil {
		logger.Error("failed to reload runtime after migration", "err", err)
	}

	return connect.NewResponse(&agentsv1.MigrateAgentsV2Response{
		Mode:     agentsv1.MigrateMode_MIGRATE_MODE_APPLY,
		Results:  results,
		Total:    int32(len(agents)),
		Migrated: migrated,
		Skipped:  skipped,
		Errors:   errCount,
	}), nil
}

// convertWorkflowNodeRefs converts workflow AGENT nodes from name-based
// `agent` references to `agent_id` references using the provided sub-agents.
func convertWorkflowNodeRefs(wf *agentsv1.WorkflowConfig, subs []*agentsv1.Agent) {
	nameToID := make(map[string]string, len(subs))
	for _, sub := range subs {
		if sub.GetAgentId() != "" {
			nameToID[sub.GetName()] = sub.GetAgentId()
		}
	}
	for _, node := range wf.GetNodes() {
		if node.GetKind() != agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT {
			continue
		}
		if node.GetAgentId() != "" {
			continue
		}
		if aid, ok := nameToID[node.GetAgent()]; ok {
			node.AgentId = aid
		}
	}
}

func migrateVerify(ctx context.Context, s *AgentServiceServer, wsID string) (*connect.Response[agentsv1.MigrateAgentsV2Response], error) {
	agents, err := s.repo.ListAgents(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}

	byID := make(map[string]*agentsv1.Agent, len(agents))
	for _, a := range agents {
		if id := a.GetAgentId(); id != "" {
			byID[id] = a
		}
	}

	var results []*agentsv1.MigrateAgentResult
	var errCount int32

	for _, a := range agents {
		if len(a.GetSubAgents()) > 0 {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:    a.GetName(),
				AgentId: a.GetAgentId(),
				Action:  "error",
				Detail:  "still has embedded sub_agents",
			})
			errCount++
			continue
		}
		if a.GetLifecycleStatus() == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:    a.GetName(),
				AgentId: a.GetAgentId(),
				Action:  "migration_required",
				Detail:  "lifecycle_status is MIGRATION_REQUIRED",
			})
			errCount++
			continue
		}

		for _, cid := range a.GetChildAgentIds() {
			if _, ok := byID[cid]; !ok {
				results = append(results, &agentsv1.MigrateAgentResult{
					Name:    a.GetName(),
					AgentId: a.GetAgentId(),
					Action:  "error",
					Detail:  fmt.Sprintf("child_agent_id %q not found", cid),
				})
				errCount++
			}
		}

		if wf := a.GetConfig().GetWorkflow(); wf != nil {
			for _, node := range wf.GetNodes() {
				if node.GetKind() != agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT {
					continue
				}
				if aid := node.GetAgentId(); aid != "" {
					if _, ok := byID[aid]; !ok {
						results = append(results, &agentsv1.MigrateAgentResult{
							Name:    a.GetName(),
							AgentId: a.GetAgentId(),
							Action:  "error",
							Detail:  fmt.Sprintf("workflow node %q references agent_id %q not found", node.GetName(), aid),
						})
						errCount++
					}
				}
			}
		}

		if errCount == 0 || len(results) == 0 || results[len(results)-1].GetName() != a.GetName() {
			results = append(results, &agentsv1.MigrateAgentResult{
				Name:    a.GetName(),
				AgentId: a.GetAgentId(),
				Action:  "ok",
			})
		}
	}

	return connect.NewResponse(&agentsv1.MigrateAgentsV2Response{
		Mode:     agentsv1.MigrateMode_MIGRATE_MODE_VERIFY,
		Results:  results,
		Total:    int32(len(agents)),
		Migrated: int32(len(agents)) - errCount,
		Errors:   errCount,
	}), nil
}
