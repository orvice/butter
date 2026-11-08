package application

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"

	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	"go.orx.me/apps/butter/internal/transport/connectx"
)

// PlaneReferenceGuard blocks removing an Agent a Plane App routes to. It is a
// strong reference for the same reason as Telegram's and Linear's: a Plane
// App whose Agent vanished would be a registration that silently never
// answers.
type PlaneReferenceGuard struct {
	repo planerepo.Repository
}

func NewPlaneReferenceGuard(repo planerepo.Repository) *PlaneReferenceGuard {
	return &PlaneReferenceGuard{repo: repo}
}

// CheckAgentRemovable returns a FailedPrecondition error naming every Plane
// App that routes to the Agent.
func (g *PlaneReferenceGuard) CheckAgentRemovable(ctx context.Context, workspaceID, agentID string) error {
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
			refs = append(refs, fmt.Sprintf("plane app %s (%s)", app.GetId(), app.GetDisplayName()))
		}
	}
	if len(refs) == 0 {
		return nil
	}
	slices.Sort(refs)
	return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
		"agent %q is routed to by %s", agentID, strings.Join(refs, ", ")))
}
