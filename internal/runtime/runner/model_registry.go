package runner

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/achetronic/adk-utils-go/plugin/contextguard"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// configuredModelRegistry overlays operator-supplied context capacities on a
// fallback registry. Model IDs are the provider-facing names, not aliases.
type configuredModelRegistry struct {
	fallback       contextguard.ModelRegistry
	contextWindows map[string]int
}

var _ contextguard.ModelRegistry = (*configuredModelRegistry)(nil)

func newConfiguredModelRegistry(providers []agentsv1.ModelProvider, fallback contextguard.ModelRegistry) (contextguard.ModelRegistry, error) {
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

func (r *configuredModelRegistry) ContextWindow(modelID string) int {
	if capacity, ok := r.contextWindows[modelID]; ok {
		return capacity
	}
	return r.fallback.ContextWindow(modelID)
}

func (r *configuredModelRegistry) DefaultMaxTokens(modelID string) int {
	return r.fallback.DefaultMaxTokens(modelID)
}
