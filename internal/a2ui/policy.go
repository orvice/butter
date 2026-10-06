package a2ui

import (
	"fmt"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// CardPolicy is the Card Policy in force at one agent's place in a run's
// agent tree (ADR-0014, the Card Policy amendment): whether its model may
// render Result Cards, and how readily it should. An agent's policy is its
// parent's, narrowed by the agent's own config. The zero value is a run's
// root: cards allowed, presentation AUTO.
type CardPolicy struct {
	disabled  bool
	preferred bool
}

// Narrow returns the policy of an agent whose own config is cfg, placed
// under p. Generation only narrows; presentation takes the nearest explicit
// value. A nil cfg inherits p.
func (p CardPolicy) Narrow(cfg *agentsv1.ResultCardConfig) CardPolicy {
	if cfg.GetGeneration() == agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED {
		p.disabled = true
	}
	switch cfg.GetPresentation() {
	case agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_AUTO:
		p.preferred = false
	case agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED:
		p.preferred = true
	}
	return p
}

// Allowed reports whether the agent's model may render Result Cards.
func (p CardPolicy) Allowed() bool { return !p.disabled }

// Preferred reports whether the agent's model is asked to show structured
// results in a card as well as in text. It is a hint, never a requirement,
// and means nothing where cards are not allowed.
func (p CardPolicy) Preferred() bool { return p.preferred }

// ValidateCardPolicy checks an agent's config.result_cards on write. Type
// restrictions live with the agent: PI and CURSOR agents reject the field,
// and composite agents accept it to narrow their subtree.
func ValidateCardPolicy(cfg *agentsv1.ResultCardConfig) error {
	switch g := cfg.GetGeneration(); g {
	case agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_UNSPECIFIED,
		agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED:
	default:
		return fmt.Errorf("config.result_cards.generation %v is not a known value: leave it unset to inherit, or use RESULT_CARD_GENERATION_DISABLED", g)
	}
	switch pr := cfg.GetPresentation(); pr {
	case agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_UNSPECIFIED,
		agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_AUTO,
		agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED:
	default:
		return fmt.Errorf("config.result_cards.presentation %v is not a known value: leave it unset to inherit, or use RESULT_CARD_PRESENTATION_AUTO or RESULT_CARD_PRESENTATION_PREFERRED", pr)
	}
	return nil
}
