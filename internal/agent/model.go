package agent

import (
	"context"
	"fmt"

	adkopenai "github.com/achetronic/adk-utils-go/genai/openai"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ResolveModel looks up the model ref (alias or name) in the provider list and returns the appropriate model.LLM.
// Resolution order: alias match → name match → Gemini fallback.
func ResolveModel(ctx context.Context, modelRef string, providers []agentsv1.ModelProvider) (model.LLM, error) {
	return resolveModel(ctx, modelRef, providers)
}

func resolveModel(ctx context.Context, modelRef string, providers []agentsv1.ModelProvider) (model.LLM, error) {
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
	name, provider := resolveModelConfig(modelRef, providers)
	if provider != nil {
		return name, true
	}
	return modelRef, false
}

// resolveModelConfig finds the actual model name and provider for a given ref (alias or name).
// Searches by alias first, then by name.
func resolveModelConfig(modelRef string, providers []agentsv1.ModelProvider) (modelName string, provider *agentsv1.ModelProvider) {
	// First pass: search by alias.
	for i := range providers {
		p := &providers[i]
		for _, m := range p.GetModels() {
			if m.GetAlias() != "" && m.GetAlias() == modelRef {
				return m.GetName(), p
			}
		}
	}

	// Second pass: search by name.
	for i := range providers {
		p := &providers[i]
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
		return gemini.NewModel(ctx, modelName, &genai.ClientConfig{})
	case "openai":
		return adkopenai.New(adkopenai.Config{
			APIKey:    p.GetApiKey(),
			BaseURL:   p.GetBaseUrl(),
			ModelName: modelName,
		}), nil
	default:
		return nil, fmt.Errorf("unknown provider type %q for model %q", p.GetType(), modelName)
	}
}
