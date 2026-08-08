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

// normalizeAgentRefWithRepo validates a record's agent_id against the
// workspace's stored agents at write time and returns the agent's current
// runtime name so legacy readers stay consistent. agent_id is the sole agent
// reference on every interface (issue #213 contract step). field names the
// proto message being validated (e.g. "channel") for error messages. Unlike
// the runner-based resolution, the repo lookup also sees MIGRATION_REQUIRED
// agents, so records can be pointed at them before migration completes.
func normalizeAgentRefWithRepo(ctx context.Context, repo configrepo.AgentRepository, wsID, agentID, field string) (string, string, error) {
	if agentID == "" {
		return "", "", connectx.RequiredArgument(field + ".agent_id")
	}
	a, err := repo.GetAgentByID(ctx, wsID, agentID)
	if err != nil {
		if errors.Is(err, configrepo.ErrNotFound) {
			return "", "", connectx.InvalidArgument(field+".agent_id", fmt.Sprintf("agent %q does not exist in workspace %q", agentID, wsID))
		}
		return "", "", toConnectError(err)
	}
	return agentID, a.GetName(), nil
}
