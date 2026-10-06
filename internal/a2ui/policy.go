package a2ui

import (
	"fmt"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// CardPolicy is the Card Policy in force at one agent's place in a run's
// agent tree (ADR-0014, the Card Policy amendment): whether its model may
// render Result Cards. It only narrows: an agent's policy is its parent's,
// narrowed by the agent's own config. The zero value is a run's root, where
// cards are allowed.
type CardPolicy struct {
	disabled bool
}

// Narrow returns the policy of an agent whose own config is cfg, placed
// under p. A nil cfg inherits p.
func (p CardPolicy) Narrow(cfg *agentsv1.ResultCardConfig) CardPolicy {
	if cfg.GetGeneration() == agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED {
		p.disabled = true
	}
	return p
}

// Allowed reports whether the agent's model may render Result Cards.
func (p CardPolicy) Allowed() bool { return !p.disabled }

// ValidateCardPolicy checks an agent's config.result_cards on write. Type
// restrictions live with the agent: PI and CURSOR agents reject the field,
// and composite agents accept it to narrow their subtree.
func ValidateCardPolicy(cfg *agentsv1.ResultCardConfig) error {
	switch g := cfg.GetGeneration(); g {
	case agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_UNSPECIFIED,
		agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED:
		return nil
	default:
		return fmt.Errorf("config.result_cards.generation %v is not a known value: leave it unset to inherit, or use RESULT_CARD_GENERATION_DISABLED", g)
	}
}
