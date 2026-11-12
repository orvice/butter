package a2ui

import (
	"strings"
	"testing"
)

func TestFluidAIWorkloadsIsAReadOnlyCatalogComponent(t *testing.T) {
	result, err := Apply(nil, Batch{
		Fallback: "AI workload flow",
		Messages: []map[string]any{{
			"updateComponents": map[string]any{
				"components": []any{
					map[string]any{
						"id":        "root",
						"component": "FluidAIWorkloads",
					},
				},
			},
		}},
	}, Anchor{}, DefaultLimits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(result.Card.Components) != 1 || result.Card.Components[0]["component"] != "FluidAIWorkloads" {
		t.Fatalf("card components = %#v", result.Card.Components)
	}
}

func TestFluidAIWorkloadsDoesNotExposePresentationProps(t *testing.T) {
	_, err := Apply(nil, Batch{
		Fallback: "AI workload flow",
		Messages: []map[string]any{{
			"updateComponents": map[string]any{
				"components": []any{
					map[string]any{
						"id":        "root",
						"component": "FluidAIWorkloads",
						"width":     448,
					},
				},
			},
		}},
	}, Anchor{}, DefaultLimits)
	if err == nil || !strings.Contains(err.Error(), "FluidAIWorkloads has no property \"width\"") {
		t.Fatalf("Apply() error = %v, want width rejection", err)
	}
}
