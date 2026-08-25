package application

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ButterBoxReferenceGuard blocks deleting a ButterBox that is still referenced
// by PI-type agents (ADR-0011). Follows the same pattern as
// TelegramReferenceGuard.
type ButterBoxReferenceGuard struct {
	agentRepo configrepo.AgentRepository
}

// NewButterBoxReferenceGuard creates a guard that scans agents for ButterBox
// references.
func NewButterBoxReferenceGuard(agentRepo configrepo.AgentRepository) *ButterBoxReferenceGuard {
	return &ButterBoxReferenceGuard{agentRepo: agentRepo}
}

// CheckRemovable returns a FailedPrecondition error listing every PI agent
// that still references the given ButterBox ID.
func (g *ButterBoxReferenceGuard) CheckRemovable(ctx context.Context, workspaceID, butterboxID string) error {
	if g == nil || g.agentRepo == nil || strings.TrimSpace(butterboxID) == "" {
		return nil
	}
	agents, err := g.agentRepo.ListAgents(ctx, workspaceID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("check butterbox references: %w", err))
	}
	var refs []string
	for _, a := range agents {
		if a.GetType() != agentsv1.AgentType_AGENT_TYPE_PI {
			continue
		}
		if a.GetConfig().GetPi().GetButterboxId() == butterboxID {
			refs = append(refs, a.GetAgentId())
		}
	}
	if len(refs) == 0 {
		return nil
	}
	return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
		"butterbox %q is referenced by pi agents: %s", butterboxID, strings.Join(refs, ", ")))
}
