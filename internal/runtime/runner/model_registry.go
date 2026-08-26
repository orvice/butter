package runner

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"charm.land/catwalk/pkg/embedded"
	"github.com/achetronic/adk-utils-go/plugin/contextguard"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type contextWindowSource string

const (
	contextWindowSourceAgent    contextWindowSource = "agent"
	contextWindowSourceModel    contextWindowSource = "model"
	contextWindowSourceEmbedded contextWindowSource = "embedded"
	contextWindowSourceFallback contextWindowSource = "fallback"
)

// contextWindowResolution is the complete, source-aware Effective Context
// Window decision for one callback-time model selection.
type contextWindowResolution struct {
	SelectedModelID         string
	MetadataSource          contextWindowSource
	ConfiguredAgentOverride int
	ConfiguredModelCapacity int
	EffectiveContextWindow  int
}

// sourceAwareContextWindowRegistry is implemented by fallbacks that can
// distinguish real metadata from their unknown-model default.
type sourceAwareContextWindowRegistry interface {
	contextguard.ModelRegistry
	resolveContextWindow(modelID string) (int, contextWindowSource)
}

// embeddedModelRegistry preserves Crush's metadata behavior while retaining
// enough catalog knowledge to attribute an equal-valued 128k result exactly.
type embeddedModelRegistry struct {
	contextguard.ModelRegistry
	contextWindows map[string]int
}

func newEmbeddedModelRegistry() *embeddedModelRegistry {
	windows := make(map[string]int)
	for _, provider := range embedded.GetAll() {
		for _, model := range provider.Models {
			// Crush uses the same last-ID-wins traversal.
			windows[model.ID] = int(model.ContextWindow)
		}
	}
	return &embeddedModelRegistry{
		ModelRegistry:  contextguard.NewCrushRegistry(),
		contextWindows: windows,
	}
}

func (r *embeddedModelRegistry) resolveContextWindow(modelID string) (int, contextWindowSource) {
	if capacity := r.contextWindows[modelID]; capacity > 0 {
		return r.ContextWindow(modelID), contextWindowSourceEmbedded
	}
	return r.ContextWindow(modelID), contextWindowSourceFallback
}

// configuredModelRegistry overlays operator-supplied context capacities on a
// fallback registry. Model IDs are the provider-facing names, not aliases.
type configuredModelRegistry struct {
	fallback       contextguard.ModelRegistry
	contextWindows map[string]int
}

var _ contextguard.ModelRegistry = (*configuredModelRegistry)(nil)

func newConfiguredModelRegistry(providers []agentsv1.ModelProvider, fallback contextguard.ModelRegistry) (*configuredModelRegistry, error) {
	if fallback == nil {
		return nil, fmt.Errorf("configured model registry requires a fallback registry")
	}

	capacities := make(map[string]map[uint32]struct{})
	for i := range providers {
		for _, model := range providers[i].GetModels() {
			capacity := model.GetContextWindowTokens()
			if capacity == 0 {
				continue
			}
			modelID := model.GetName()
			if capacities[modelID] == nil {
				capacities[modelID] = make(map[uint32]struct{})
			}
			capacities[modelID][capacity] = struct{}{}
		}
	}

	modelIDs := make([]string, 0, len(capacities))
	for modelID := range capacities {
		modelIDs = append(modelIDs, modelID)
	}
	slices.Sort(modelIDs)

	windows := make(map[string]int, len(capacities))
	for _, modelID := range modelIDs {
		values := make([]uint32, 0, len(capacities[modelID]))
		for value := range capacities[modelID] {
			values = append(values, value)
		}
		slices.Sort(values)
		if len(values) > 1 {
			formatted := make([]string, len(values))
			for i, value := range values {
				formatted[i] = strconv.FormatUint(uint64(value), 10)
			}
			return nil, fmt.Errorf("model %q has conflicting context_window_tokens values: %s", modelID, strings.Join(formatted, ", "))
		}
		windows[modelID] = int(values[0])
	}

	return &configuredModelRegistry{
		fallback:       fallback,
		contextWindows: windows,
	}, nil
}

// Resolve applies Effective Context Window precedence for the actual selected
// model ID. Agent overrides are supplied by ContextGuard policy, never aliases.
func (r *configuredModelRegistry) Resolve(modelID string, agentOverride int) contextWindowResolution {
	resolution := contextWindowResolution{
		SelectedModelID:         modelID,
		ConfiguredAgentOverride: agentOverride,
		ConfiguredModelCapacity: r.contextWindows[modelID],
	}

	if agentOverride > 0 {
		resolution.MetadataSource = contextWindowSourceAgent
		resolution.EffectiveContextWindow = agentOverride
		return resolution
	}
	if resolution.ConfiguredModelCapacity > 0 {
		resolution.MetadataSource = contextWindowSourceModel
		resolution.EffectiveContextWindow = resolution.ConfiguredModelCapacity
		return resolution
	}
	if fallback, ok := r.fallback.(sourceAwareContextWindowRegistry); ok {
		resolution.EffectiveContextWindow, resolution.MetadataSource = fallback.resolveContextWindow(modelID)
		return resolution
	}
	resolution.MetadataSource = contextWindowSourceFallback
	resolution.EffectiveContextWindow = r.fallback.ContextWindow(modelID)
	return resolution
}

func (r *configuredModelRegistry) ContextWindow(modelID string) int {
	return r.Resolve(modelID, 0).EffectiveContextWindow
}

func (r *configuredModelRegistry) DefaultMaxTokens(modelID string) int {
	return r.fallback.DefaultMaxTokens(modelID)
}
