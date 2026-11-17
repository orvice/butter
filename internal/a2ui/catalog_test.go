package a2ui

import (
	"fmt"
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

func TestAnalyticsChartValidatesBoundedReadOnlyPoints(t *testing.T) {
	result, err := Apply(nil, Batch{Fallback: "Requests chart", Messages: []map[string]any{{
		"updateComponents": map[string]any{"components": []any{
			map[string]any{"id": "root", "component": "AnalyticsChart", "data": []any{
				map[string]any{"label": "Mon", "value": float64(12)},
			}, "title": "Requests"},
		}},
	}}}, Anchor{}, DefaultLimits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(result.Card.Components) != 1 || result.Card.Components[0]["component"] != "AnalyticsChart" {
		t.Fatalf("card components = %#v", result.Card.Components)
	}

	tooMany := make([]any, maxAnalyticsPoints+1)
	for i := range tooMany {
		tooMany[i] = map[string]any{"label": fmt.Sprintf("P%d", i), "value": float64(i)}
	}
	for name, data := range map[string]any{
		"too many points":      tooMany,
		"extra point property": []any{map[string]any{"label": "Mon", "value": float64(1), "url": "https://example.com"}},
		"url label":            []any{map[string]any{"label": "https://example.com", "value": float64(1)}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Apply(nil, Batch{Fallback: "Requests chart", Messages: []map[string]any{{
				"updateComponents": map[string]any{"components": []any{
					map[string]any{"id": "root", "component": "AnalyticsChart", "data": data},
				}},
			}}}, Anchor{}, DefaultLimits)
			if err == nil {
				t.Fatal("Apply() succeeded for invalid chart data")
			}
		})
	}
}

func TestAnalyticsChartAcceptsDataModelBinding(t *testing.T) {
	_, err := Apply(nil, Batch{Fallback: "Requests chart", Messages: []map[string]any{{
		"updateComponents": map[string]any{"components": []any{
			map[string]any{"id": "root", "component": "AnalyticsChart", "data": map[string]any{"path": "/metrics"}},
		}},
	}}}, Anchor{}, DefaultLimits)
	if err != nil {
		t.Fatalf("Apply() rejected data binding: %v", err)
	}
}
