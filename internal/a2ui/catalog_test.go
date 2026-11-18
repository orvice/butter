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

func TestCodeBlockValidatesLanguageAndPlainText(t *testing.T) {
	result, err := Apply(nil, Batch{Fallback: "Example code", Messages: []map[string]any{{
		"updateComponents": map[string]any{"components": []any{
			map[string]any{
				"id": "root", "component": "CodeBlock", "language": "typescript",
				"title": "Example", "code": "const answer: number = 42",
			},
		}},
	}}}, Anchor{}, DefaultLimits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(result.Card.Components) != 1 || result.Card.Components[0]["component"] != "CodeBlock" {
		t.Fatalf("card components = %#v", result.Card.Components)
	}

	for name, component := range map[string]map[string]any{
		"unsupported language": {"id": "root", "component": "CodeBlock", "code": "echo hi", "language": "shell"},
		"oversized code":       {"id": "root", "component": "CodeBlock", "code": strings.Repeat("x", maxCodeRunes+1)},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Apply(nil, Batch{Fallback: "Example code", Messages: []map[string]any{{
				"updateComponents": map[string]any{"components": []any{component}},
			}}}, Anchor{}, DefaultLimits)
			if err == nil {
				t.Fatal("Apply() succeeded for invalid code block")
			}
		})
	}
	if _, err := Apply(nil, Batch{Fallback: "Example code", Messages: []map[string]any{{
		"updateComponents": map[string]any{"components": []any{map[string]any{
			"id": "root", "component": "CodeBlock", "code": map[string]any{"path": "/source"},
		}}},
	}}}, Anchor{}, DefaultLimits); err != nil {
		t.Fatalf("Apply() rejected code data binding: %v", err)
	}
	if _, err := Apply(nil, Batch{Fallback: "Example code", Messages: []map[string]any{{
		"updateComponents": map[string]any{"components": []any{map[string]any{
			"id": "root", "component": "CodeBlock", "code": "<script>alert(1)</script>",
		}}},
	}}}, Anchor{}, DefaultLimits); err != nil {
		t.Fatalf("Apply() rejected escaped HTML code text: %v", err)
	}
}

func TestHaloProgressValidatesBoundedNumbers(t *testing.T) {
	apply := func(component map[string]any) error {
		_, err := Apply(nil, Batch{Fallback: "Progress", Messages: []map[string]any{{
			"updateComponents": map[string]any{"components": []any{component}},
		}}}, Anchor{}, DefaultLimits)
		return err
	}
	for name, component := range map[string]map[string]any{
		"literal":      {"id": "root", "component": "HaloProgress", "value": 42.0, "max": 100.0, "label": "Deploy", "tone": "success"},
		"value only":   {"id": "root", "component": "HaloProgress", "value": 3.0},
		"data binding": {"id": "root", "component": "HaloProgress", "value": map[string]any{"path": "/done"}, "max": map[string]any{"path": "/total"}},
	} {
		t.Run("accepts "+name, func(t *testing.T) {
			if err := apply(component); err != nil {
				t.Fatalf("Apply() error = %v", err)
			}
		})
	}
	for name, component := range map[string]map[string]any{
		"missing value":    {"id": "root", "component": "HaloProgress", "max": 100.0},
		"string value":     {"id": "root", "component": "HaloProgress", "value": "50"},
		"huge value":       {"id": "root", "component": "HaloProgress", "value": 1e30},
		"bad tone":         {"id": "root", "component": "HaloProgress", "value": 1.0, "tone": "rainbow"},
		"extra property":   {"id": "root", "component": "HaloProgress", "value": 1.0, "color": "red"},
		"url label":        {"id": "root", "component": "HaloProgress", "value": 1.0, "label": "see https://example.com"},
		"bad binding":      {"id": "root", "component": "HaloProgress", "value": map[string]any{"path": "/a", "x": 1}},
		"function binding": {"id": "root", "component": "HaloProgress", "value": map[string]any{"call": "now"}},
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			if err := apply(component); err == nil {
				t.Fatal("Apply() succeeded for invalid progress")
			}
		})
	}
}
