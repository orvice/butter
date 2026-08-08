package agent

import (
	"fmt"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ValidateAgentRelationships checks that the agent's child_agent_ids and
// workflow agent_id references are valid within the workspace agent pool.
//
// Checks performed:
//   - Every child_agent_id exists in the pool and belongs to the same workspace
//   - No cycles through child_agent_ids
//   - Each child is claimed by at most one parent (single-parent forest)
//   - Workflow AGENT nodes reference agents that exist in the pool
//   - No self-references
func ValidateAgentRelationships(agent *agentsv1.Agent, pool []*agentsv1.Agent) error {
	if agent == nil {
		return nil
	}
	childIDs := agent.GetChildAgentIds()
	if len(childIDs) == 0 && !hasWorkflowAgentIDRefs(agent) {
		return nil
	}

	byID := buildPoolIndex(pool)
	wsID := agent.GetWorkspaceId()

	for _, cid := range childIDs {
		if cid == agent.GetAgentId() {
			return fmt.Errorf("agent %q: child_agent_ids contains self-reference", agent.GetAgentId())
		}
		child, ok := byID[cid]
		if !ok {
			return fmt.Errorf("agent %q: child_agent_id %q not found in workspace", agent.GetAgentId(), cid)
		}
		if child.GetWorkspaceId() != wsID {
			return fmt.Errorf("agent %q: child_agent_id %q belongs to workspace %q, not %q",
				agent.GetAgentId(), cid, child.GetWorkspaceId(), wsID)
		}
	}

	if len(childIDs) > 0 {
		if err := detectCycle(agent.GetAgentId(), byID); err != nil {
			return err
		}
	}

	if err := validateSingleParent(agent, pool); err != nil {
		return err
	}

	if err := validateWorkflowNodeRefs(agent, byID); err != nil {
		return err
	}

	return nil
}

// ValidateNoOrphanedReferences checks that no other agent in the pool
// references the given agentID in its child_agent_ids or workflow nodes.
// Used by DeleteAgent to reject deletion of referenced agents.
func ValidateNoOrphanedReferences(agentID string, pool []*agentsv1.Agent) error {
	for _, a := range pool {
		if a.GetAgentId() == agentID {
			continue
		}
		for _, cid := range a.GetChildAgentIds() {
			if cid == agentID {
				return fmt.Errorf("agent %q is referenced by agent %q in child_agent_ids", agentID, a.GetAgentId())
			}
		}
		if wf := a.GetConfig().GetWorkflow(); wf != nil {
			for _, node := range wf.GetNodes() {
				if node.GetKind() == agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT && node.GetAgentId() == agentID {
					return fmt.Errorf("agent %q is referenced by workflow node %q in agent %q", agentID, node.GetName(), a.GetAgentId())
				}
			}
		}
	}
	return nil
}

func buildPoolIndex(pool []*agentsv1.Agent) map[string]*agentsv1.Agent {
	m := make(map[string]*agentsv1.Agent, len(pool))
	for _, a := range pool {
		if id := a.GetAgentId(); id != "" {
			m[id] = a
		}
	}
	return m
}

func hasWorkflowAgentIDRefs(agent *agentsv1.Agent) bool {
	wf := agent.GetConfig().GetWorkflow()
	if wf == nil {
		return false
	}
	for _, node := range wf.GetNodes() {
		if node.GetKind() == agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT && node.GetAgentId() != "" {
			return true
		}
	}
	return false
}

// detectCycle performs DFS from startID following child_agent_ids to detect cycles.
func detectCycle(startID string, byID map[string]*agentsv1.Agent) error {
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	var dfs func(id string) error
	dfs = func(id string) error {
		if stack[id] {
			return fmt.Errorf("cycle detected involving agent %q", id)
		}
		if visited[id] {
			return nil
		}
		visited[id] = true
		stack[id] = true

		if a, ok := byID[id]; ok {
			for _, cid := range a.GetChildAgentIds() {
				if err := dfs(cid); err != nil {
					return err
				}
			}
		}

		stack[id] = false
		return nil
	}

	return dfs(startID)
}

// validateSingleParent ensures that the children claimed by this agent are
// not already claimed as children by another agent in the pool.
func validateSingleParent(agent *agentsv1.Agent, pool []*agentsv1.Agent) error {
	myID := agent.GetAgentId()
	myChildren := make(map[string]struct{}, len(agent.GetChildAgentIds()))
	for _, cid := range agent.GetChildAgentIds() {
		myChildren[cid] = struct{}{}
	}
	if len(myChildren) == 0 {
		return nil
	}

	for _, other := range pool {
		if other.GetAgentId() == myID || other.GetAgentId() == "" {
			continue
		}
		for _, cid := range other.GetChildAgentIds() {
			if _, claimed := myChildren[cid]; claimed {
				return fmt.Errorf("agent %q: child %q is already claimed by agent %q (single-parent constraint)",
					myID, cid, other.GetAgentId())
			}
		}
	}
	return nil
}

func validateWorkflowNodeRefs(agent *agentsv1.Agent, byID map[string]*agentsv1.Agent) error {
	wf := agent.GetConfig().GetWorkflow()
	if wf == nil {
		return nil
	}
	wsID := agent.GetWorkspaceId()
	for _, node := range wf.GetNodes() {
		if node.GetKind() != agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT {
			continue
		}
		aid := node.GetAgentId()
		if aid == "" {
			continue
		}
		child, ok := byID[aid]
		if !ok {
			return fmt.Errorf("agent %q: workflow node %q references agent_id %q not found in workspace",
				agent.GetAgentId(), node.GetName(), aid)
		}
		if child.GetWorkspaceId() != wsID {
			return fmt.Errorf("agent %q: workflow node %q references agent_id %q from workspace %q, not %q",
				agent.GetAgentId(), node.GetName(), aid, child.GetWorkspaceId(), wsID)
		}
	}
	return nil
}
