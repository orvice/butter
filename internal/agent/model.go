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

// resolveModel looks up the model name in the provider list and returns the appropriate model.LLM.
// Falls back to Gemini if no provider matches.
func resolveModel(ctx context.Context, modelName string, providers []agentsv1.ModelProvider) (model.LLM, error) {
	for i := range providers {
		p := &providers[i]
		for _, m := range p.GetModels() {
			if m == modelName {
				return createModelFromProvider(ctx, modelName, p)
			}
		}
	}

	// Fallback: use Gemini.
	return gemini.NewModel(ctx, modelName, &genai.ClientConfig{})
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
