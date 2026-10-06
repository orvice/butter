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
