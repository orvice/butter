package agent

import (
	"fmt"

	"go.orx.me/apps/butter/internal/a2ui"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ValidateResultCardConfig checks an agent's Card Policy (ADR-0014). Type
// restrictions live elsewhere: PI and CURSOR agents reject the field through
// rejectBoxOwnedFields, and composite agents accept it to narrow their
// subtree.
func ValidateResultCardConfig(pb *agentsv1.Agent) error {
	if err := a2ui.ValidateCardPolicy(pb.GetConfig().GetResultCards()); err != nil {
		return fmt.Errorf("agent %q: %w", pb.GetName(), err)
	}
	return nil
}
