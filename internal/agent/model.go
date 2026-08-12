package agent

import (
	"context"
	"fmt"

	adkopenai "github.com/achetronic/adk-utils-go/genai/openai"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ResolveModel looks up the model ref (alias or name) in the provider list and returns the appropriate model.LLM.
// Resolution order: alias match → name match → Gemini fallback.
func ResolveModel(ctx context.Context, modelRef string, providers []agentsv1.ModelProvider) (model.LLM, error) {
	return ResolveModelPtr(ctx, modelRef, providerPtrs(providers))
}

// ResolveModelPtr is ResolveModel over an already-pointer provider list, for
// callers (e.g. repo-backed listings) that hold *agentsv1.ModelProvider —
// converting those to values would copy proto lock state (govet copylocks).
func ResolveModelPtr(ctx context.Context, modelRef string, providers []*agentsv1.ModelProvider) (model.LLM, error) {
	modelName, provider := resolveModelConfig(modelRef, providers)
	if provider != nil {
		return createModelFromProvider(ctx, modelName, provider)
	}

	// Fallback: use Gemini with the original ref as model name.
	return gemini.NewModel(ctx, modelRef, &genai.ClientConfig{})
}

// ResolveModelAlias resolves a model reference (alias or name) to the actual provider model name.
// Returns the resolved model name and true if found, or the original ref and false if not found.
func ResolveModelAlias(modelRef string, providers []agentsv1.ModelProvider) (modelName string, found bool) {
	return ResolveModelAliasPtr(modelRef, providerPtrs(providers))
}

// ResolveModelAliasPtr is ResolveModelAlias over an already-pointer provider list.
func ResolveModelAliasPtr(modelRef string, providers []*agentsv1.ModelProvider) (modelName string, found bool) {
	name, provider := resolveModelConfig(modelRef, providers)
	if provider != nil {
		return name, true
	}
	return modelRef, false
}

// providerPtrs views a value slice as pointers without copying the messages.
func providerPtrs(providers []agentsv1.ModelProvider) []*agentsv1.ModelProvider {
	out := make([]*agentsv1.ModelProvider, len(providers))
	for i := range providers {
		out[i] = &providers[i]
	}
	return out
}

// resolveModelConfig finds the actual model name and provider for a given ref (alias or name).
// Searches by alias first, then by name.
func resolveModelConfig(modelRef string, providers []*agentsv1.ModelProvider) (modelName string, provider *agentsv1.ModelProvider) {
	// First pass: search by alias.
	for _, p := range providers {
		for _, m := range p.GetModels() {
			if m.GetAlias() != "" && m.GetAlias() == modelRef {
				return m.GetName(), p
			}
		}
	}

	// Second pass: search by name.
	for _, p := range providers {
		for _, m := range p.GetModels() {
			if m.GetName() == modelRef {
				return m.GetName(), p
			}
		}
	}

	return modelRef, nil
}

// ValidateModelAliases checks that no two models across all providers share the same alias.
func ValidateModelAliases(providers []agentsv1.ModelProvider) error {
	seen := make(map[string]string) // alias → "providerName/modelName"
	for i := range providers {
		p := &providers[i]
		for _, m := range p.GetModels() {
			alias := m.GetAlias()
			if alias == "" {
				continue
			}
			key := p.GetName() + "/" + m.GetName()
			if prev, ok := seen[alias]; ok {
				return fmt.Errorf("duplicate model alias %q: used by %s and %s", alias, prev, key)
			}
			seen[alias] = key
		}
	}
	return nil
}

// AllModelAliases returns all model aliases (or names if no alias) across all providers.
// Each entry is {alias, name}. Useful for listing available models.
func AllModelAliases(providers []agentsv1.ModelProvider) []ModelInfo {
	var result []ModelInfo
	for i := range providers {
		p := &providers[i]
		for _, m := range p.GetModels() {
			alias := m.GetAlias()
			if alias == "" {
				alias = m.GetName()
			}
			result = append(result, ModelInfo{
				Alias: alias,
				Name:  m.GetName(),
			})
		}
	}
	return result
}

// ModelInfo holds a model's alias and actual provider name.
type ModelInfo struct {
	Alias string
	Name  string
}

func createModelFromProvider(ctx context.Context, modelName string, p *agentsv1.ModelProvider) (model.LLM, error) {
	switch p.GetType() {
	case "gemini":
		cfg := &genai.ClientConfig{}
		if key := p.GetApiKey(); key != "" {
			cfg.APIKey = key
		}
		if base := p.GetBaseUrl(); base != "" {
			cfg.HTTPOptions.BaseURL = base
		}
		return gemini.NewModel(ctx, modelName, cfg)
	case "openai":
		// ADK v2.1.0's native openaimodel is not production-safe for Butter yet:
		// it breaks multi-turn assistant history and does not map genai media
		// parts to the Responses API's supported input_image/input_file types.
		// See docs/research/adk-go-v2.1-openai.md before changing this adapter.
		return adkopenai.New(adkopenai.Config{
			APIKey:    p.GetApiKey(),
			BaseURL:   p.GetBaseUrl(),
			ModelName: modelName,
		}), nil
	default:
		return nil, fmt.Errorf("unknown provider type %q for model %q", p.GetType(), modelName)
	}
}
