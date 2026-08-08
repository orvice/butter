package application

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// equalSubAgents reports whether two embedded sub-agent trees are identical —
// used by UpdateAgent to allow legacy agents to round-trip unchanged while
// rejecting embedded-tree mutations (the write path was removed in V2).
func equalSubAgents(a, b []*agentsv1.Agent) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !proto.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

// normalizeAgentRefWithRepo reconciles a record's (agent_id, legacy
// agent_name) pair against the workspace's stored agents at write time:
//
//   - agent_id set: the agent must exist in the workspace; the returned name
//     is the agent's current runtime name so legacy readers stay consistent.
//     A set-but-unknown agent_id never falls back to the name.
//   - only agent_name set: the agent must exist; its agent_id (when already
//     assigned) is returned so runtime consumers can resolve by ID.
//
// field names the proto field being validated (e.g. "channel") for error
// messages. Unlike the runner-based resolution, the repo lookup also sees
// MIGRATION_REQUIRED agents, so records can be re-pointed before migration.
func normalizeAgentRefWithRepo(ctx context.Context, repo configrepo.AgentRepository, wsID, agentID, agentName, field string) (string, string, error) {
	switch {
	case agentID != "":
		a, err := repo.GetAgentByID(ctx, wsID, agentID)
		if err != nil {
			if errors.Is(err, configrepo.ErrNotFound) {
				return "", "", connectx.InvalidArgument(field+".agent_id", fmt.Sprintf("agent %q does not exist in workspace %q", agentID, wsID))
			}
			return "", "", toConnectError(err)
		}
		return agentID, a.GetName(), nil
	case agentName != "":
		a, err := repo.GetAgent(ctx, wsID, agentName)
		if err != nil {
			if errors.Is(err, configrepo.ErrNotFound) {
				return "", "", connectx.InvalidArgument(field+".agent_name", fmt.Sprintf("agent %q does not exist in workspace %q", agentName, wsID))
			}
			return "", "", toConnectError(err)
		}
		return a.GetAgentId(), agentName, nil
	default:
		return "", "", connectx.RequiredArgument(field + ".agent_id")
	}
}
