package application

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"

	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/transport/connectx"
)

// LinearReferenceGuard blocks removing an Agent a Linear App routes to. It is
// a strong reference for the same reason as Telegram's: an App whose Agent
// vanished is an issue delegation that silently never answers.
type LinearReferenceGuard struct {
	repo linearrepo.Repository
}

func NewLinearReferenceGuard(repo linearrepo.Repository) *LinearReferenceGuard {
	return &LinearReferenceGuard{repo: repo}
}

// CheckAgentRemovable returns a FailedPrecondition error naming every Linear
// App that routes to the Agent.
func (g *LinearReferenceGuard) CheckAgentRemovable(ctx context.Context, workspaceID, agentID string) error {
	if g == nil || g.repo == nil || strings.TrimSpace(agentID) == "" {
		return nil
	}
	apps, err := g.repo.ListApps(ctx, workspaceID)
	if err != nil {
		return connectx.InternalWith(err)
	}
	var refs []string
	for _, app := range apps {
		if app.GetAgentId() == agentID {
			refs = append(refs, app.GetId())
		}
	}
	if len(refs) == 0 {
		return nil
	}
	slices.Sort(refs)
	return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
		"agent %q is routed to by linear apps: %s", agentID, strings.Join(refs, ", ")))
}
