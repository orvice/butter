package a2ui

import (
	"strings"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func generation(g agentsv1.ResultCardGeneration) *agentsv1.ResultCardConfig {
	return &agentsv1.ResultCardConfig{Generation: g}
}

// The Card Policy only narrows down the agent tree: a run's root allows
// cards, DISABLED turns them off, and nothing below a disabled agent turns
// them back on.
func TestCardPolicyNarrow(t *testing.T) {
	disabled := CardPolicy{}.Narrow(generation(agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED))
	cases := []struct {
		name   string
		parent CardPolicy
		config *agentsv1.ResultCardConfig
		want   bool
	}{
		{name: "a run's root", parent: CardPolicy{}, want: true},
		{name: "root, empty config", parent: CardPolicy{}, config: &agentsv1.ResultCardConfig{}, want: true},
		{name: "root, unspecified", parent: CardPolicy{}, config: generation(agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_UNSPECIFIED), want: true},
		{name: "root, disabled", parent: CardPolicy{}, config: generation(agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED), want: false},
		{name: "below a disabled agent", parent: disabled, want: false},
		{name: "below a disabled agent, unspecified", parent: disabled, config: generation(agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_UNSPECIFIED), want: false},
		{name: "below a disabled agent, disabled", parent: disabled, config: generation(agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.parent.Narrow(tc.config).Allowed(); got != tc.want {
				t.Fatalf("Allowed() = %v, want %v", got, tc.want)
			}
		})
	}
	if !(CardPolicy{}).Allowed() {
		t.Fatal("the zero value, a run's root, does not allow cards")
	}
}

func presentation(pr agentsv1.ResultCardPresentation) *agentsv1.ResultCardConfig {
	return &agentsv1.ResultCardConfig{Presentation: pr}
}

// Presentation is inherited down the tree, the nearest explicit value
// winning; a run's root uses AUTO.
func TestCardPolicyNarrowPresentation(t *testing.T) {
	preferred := CardPolicy{}.Narrow(presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED))
	auto := CardPolicy{}.Narrow(presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_AUTO))
	cases := []struct {
		name   string
		parent CardPolicy
		config *agentsv1.ResultCardConfig
		want   bool
	}{
		{name: "a run's root", parent: CardPolicy{}, want: false},
		{name: "root, preferred", parent: CardPolicy{}, config: presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED), want: true},
		{name: "root, auto", parent: CardPolicy{}, config: presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_AUTO), want: false},
		{name: "below preferred, unset", parent: preferred, want: true},
		{name: "below preferred, unspecified", parent: preferred, config: presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_UNSPECIFIED), want: true},
		{name: "below preferred, auto", parent: preferred, config: presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_AUTO), want: false},
		{name: "below auto, preferred", parent: auto, config: presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED), want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.parent.Narrow(tc.config).Preferred(); got != tc.want {
				t.Fatalf("Preferred() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValidateCardPolicy(t *testing.T) {
	cases := []struct {
		name    string
		config  *agentsv1.ResultCardConfig
		wantErr string
	}{
		{name: "no policy"},
		{name: "empty policy", config: &agentsv1.ResultCardConfig{}},
		{name: "disabled", config: generation(agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED)},
		{name: "unknown generation", config: generation(agentsv1.ResultCardGeneration(7)), wantErr: "config.result_cards.generation"},
		{name: "auto", config: presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_AUTO)},
		{name: "preferred", config: presentation(agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED)},
		{name: "unknown presentation", config: presentation(agentsv1.ResultCardPresentation(9)), wantErr: "config.result_cards.presentation"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCardPolicy(tc.config)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want mention of %q", err, tc.wantErr)
			}
		})
	}
}
