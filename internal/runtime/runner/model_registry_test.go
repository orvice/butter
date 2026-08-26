package runner

import (
	"context"
	"strings"
	"testing"

	"github.com/achetronic/adk-utils-go/plugin/contextguard"
	adkrunner "google.golang.org/adk/v2/runner"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeModelRegistry struct {
	contextWindows map[string]int
	outputTokens   map[string]int
}

func (r fakeModelRegistry) ContextWindow(modelID string) int {
	if value, ok := r.contextWindows[modelID]; ok {
		return value
	}
	return 128_000
}

func (r fakeModelRegistry) DefaultMaxTokens(modelID string) int {
	if value, ok := r.outputTokens[modelID]; ok {
		return value
	}
	return 4_096
}

func (r fakeModelRegistry) resolveContextWindow(modelID string) (int, contextWindowSource) {
	if value, ok := r.contextWindows[modelID]; ok {
		return value, contextWindowSourceEmbedded
	}
	return r.ContextWindow(modelID), contextWindowSourceFallback
}

func modelProvider(name string, models ...*agentsv1.ModelConfig) agentsv1.ModelProvider {
	return agentsv1.ModelProvider{Name: name, Models: models}
}

func TestConfiguredModelRegistryContract(t *testing.T) {
	fallback := fakeModelRegistry{
		contextWindows: map[string]int{
			"embedded": 200_000,
			"zero":     300_000,
		},
		outputTokens: map[string]int{
			"configured": 8_192,
		},
	}
	providers := []agentsv1.ModelProvider{
		modelProvider("first",
			&agentsv1.ModelConfig{Name: "configured", ContextWindowTokens: 64_000},
			&agentsv1.ModelConfig{Name: "zero"},
		),
		modelProvider("second",
			&agentsv1.ModelConfig{Name: "configured"},
			&agentsv1.ModelConfig{Name: "configured", ContextWindowTokens: 64_000},
		),
	}

	registry, err := newConfiguredModelRegistry(providers, fallback)
	if err != nil {
		t.Fatalf("newConfiguredModelRegistry: %v", err)
	}
	var contract contextguard.ModelRegistry = registry

	tests := []struct {
		name    string
		modelID string
		want    int
	}{
		{name: "configured capacity wins", modelID: "configured", want: 64_000},
		{name: "zero delegates", modelID: "zero", want: 300_000},
		{name: "embedded delegates", modelID: "embedded", want: 200_000},
		{name: "unknown delegates fallback", modelID: "unknown", want: 128_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contract.ContextWindow(tt.modelID); got != tt.want {
				t.Fatalf("ContextWindow(%q) = %d, want %d", tt.modelID, got, tt.want)
			}
		})
	}

	if got := contract.DefaultMaxTokens("configured"); got != 8_192 {
		t.Fatalf("DefaultMaxTokens(configured) = %d, want delegated 8192", got)
	}
	if got := contract.DefaultMaxTokens("unknown"); got != 4_096 {
		t.Fatalf("DefaultMaxTokens(unknown) = %d, want delegated 4096", got)
	}
}

func TestConfiguredModelRegistryResolvesEffectiveContextWindow(t *testing.T) {
	fallback := fakeModelRegistry{contextWindows: map[string]int{"embedded": 128_000}}
	registry, err := newConfiguredModelRegistry([]agentsv1.ModelProvider{
		modelProvider("configured", &agentsv1.ModelConfig{Name: "configured", ContextWindowTokens: 64_000}),
	}, fallback)
	if err != nil {
		t.Fatalf("newConfiguredModelRegistry: %v", err)
	}

	tests := []struct {
		name          string
		modelID       string
		agentOverride int
		want          contextWindowResolution
	}{
		{
			name:          "agent override wins and retains model metadata",
			modelID:       "configured",
			agentOverride: 32_000,
			want: contextWindowResolution{
				SelectedModelID:         "configured",
				MetadataSource:          contextWindowSourceAgent,
				ConfiguredAgentOverride: 32_000,
				ConfiguredModelCapacity: 64_000,
				EffectiveContextWindow:  32_000,
			},
		},
		{
			name:    "configured model wins",
			modelID: "configured",
			want: contextWindowResolution{
				SelectedModelID:         "configured",
				MetadataSource:          contextWindowSourceModel,
				ConfiguredModelCapacity: 64_000,
				EffectiveContextWindow:  64_000,
			},
		},
		{
			name:    "embedded metadata is attributed exactly",
			modelID: "embedded",
			want: contextWindowResolution{
				SelectedModelID:        "embedded",
				MetadataSource:         contextWindowSourceEmbedded,
				EffectiveContextWindow: 128_000,
			},
		},
		{
			name:    "unknown model uses fallback",
			modelID: "unknown",
			want: contextWindowResolution{
				SelectedModelID:        "unknown",
				MetadataSource:         contextWindowSourceFallback,
				EffectiveContextWindow: 128_000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := registry.Resolve(tt.modelID, tt.agentOverride); got != tt.want {
				t.Fatalf("Resolve(%q, %d) = %+v, want %+v", tt.modelID, tt.agentOverride, got, tt.want)
			}
		})
	}
}

func TestConfiguredModelRegistryRejectsConflictsDeterministically(t *testing.T) {
	orders := [][]agentsv1.ModelProvider{
		{
			modelProvider("z", &agentsv1.ModelConfig{Name: "shared", ContextWindowTokens: 96_000}),
			modelProvider("a", &agentsv1.ModelConfig{Name: "shared", ContextWindowTokens: 32_000}),
			modelProvider("zero", &agentsv1.ModelConfig{Name: "shared"}),
		},
		{
			modelProvider("zero", &agentsv1.ModelConfig{Name: "shared"}),
			modelProvider("a", &agentsv1.ModelConfig{Name: "shared", ContextWindowTokens: 32_000}),
			modelProvider("z", &agentsv1.ModelConfig{Name: "shared", ContextWindowTokens: 96_000}),
		},
	}

	const want = `model "shared" has conflicting context_window_tokens values: 32000, 96000`
	for i, providers := range orders {
		_, err := newConfiguredModelRegistry(providers, fakeModelRegistry{})
		if err == nil || err.Error() != want {
			t.Fatalf("order %d error = %v, want %q", i, err, want)
		}
	}
}

func TestConfiguredModelRegistryWrapsCrushFallback(t *testing.T) {
	registry, err := newConfiguredModelRegistry(
		[]agentsv1.ModelProvider{modelProvider("custom",
			&agentsv1.ModelConfig{Name: "custom-model", ContextWindowTokens: 24_000},
		)},
		newEmbeddedModelRegistry(),
	)
	if err != nil {
		t.Fatalf("newConfiguredModelRegistry: %v", err)
	}
	if got := registry.ContextWindow("custom-model"); got != 24_000 {
		t.Fatalf("configured context window = %d, want 24000", got)
	}
	if got := registry.ContextWindow("unknown-model"); got != 128_000 {
		t.Fatalf("unknown context window = %d, want Crush fallback 128000", got)
	}

	known := registry.Resolve("gpt-4o", 0)
	if known.EffectiveContextWindow != 128_000 || known.MetadataSource != contextWindowSourceEmbedded {
		t.Fatalf("known 128k resolution = %+v, want embedded source with 128000 tokens", known)
	}
	unknown := registry.Resolve("unknown-model", 0)
	if unknown.EffectiveContextWindow != 128_000 || unknown.MetadataSource != contextWindowSourceFallback {
		t.Fatalf("unknown 128k resolution = %+v, want fallback source with 128000 tokens", unknown)
	}
}

func TestNewServiceRejectsConflictingModelCapacitiesWithoutContextGuardAgents(t *testing.T) {
	providers := []agentsv1.ModelProvider{
		modelProvider("one", &agentsv1.ModelConfig{Name: "shared", ContextWindowTokens: 32_000}),
		modelProvider("two", &agentsv1.ModelConfig{Name: "shared", ContextWindowTokens: 64_000}),
	}

	_, err := NewService(context.Background(), nil, providers, nil, nil, nil, nil, nil, nil, adkrunner.PluginConfig{})
	if err == nil {
		t.Fatal("NewService succeeded with conflicting model capacities")
	}
	if message := err.Error(); !strings.Contains(message, "shared") || !strings.Contains(message, "32000, 64000") {
		t.Fatalf("NewService error = %q, want model ID and sorted conflicting values", message)
	}
}
