package repotest

import (
	"context"
	"testing"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ModelProviderFactory returns a fresh, empty repository for one subtest.
type ModelProviderFactory func(t *testing.T) configrepo.ModelProviderRepository

// RunModelProviders exercises ModelConfig persistence through the shared
// repository interface so memory and Mongo implementations keep the same wire
// behavior.
func RunModelProviders(t *testing.T, factory ModelProviderFactory) {
	const workspaceID = "ws-models"
	ctx := context.Background()

	t.Run("context_capacity_round_trip", func(t *testing.T) {
		repo := factory(t)
		provider := &agentsv1.ModelProvider{
			Name: "custom",
			Type: "openai",
			Models: []*agentsv1.ModelConfig{
				{Name: "configured", Alias: "primary", ContextWindowTokens: 96_000},
				{Name: "embedded"},
			},
		}

		created, err := repo.CreateModelProvider(ctx, workspaceID, provider)
		if err != nil {
			t.Fatalf("CreateModelProvider: %v", err)
		}
		assertModelCapacities(t, created, 96_000, 0)

		got, err := repo.GetModelProvider(ctx, workspaceID, provider.GetName())
		if err != nil {
			t.Fatalf("GetModelProvider: %v", err)
		}
		assertModelCapacities(t, got, 96_000, 0)

		listed, err := repo.ListModelProviders(ctx, workspaceID)
		if err != nil {
			t.Fatalf("ListModelProviders: %v", err)
		}
		if len(listed) != 1 {
			t.Fatalf("ListModelProviders returned %d providers, want 1", len(listed))
		}
		assertModelCapacities(t, listed[0], 96_000, 0)

		across, err := repo.ListModelProvidersAcrossWorkspaces(ctx)
		if err != nil {
			t.Fatalf("ListModelProvidersAcrossWorkspaces: %v", err)
		}
		if len(across) != 1 {
			t.Fatalf("ListModelProvidersAcrossWorkspaces returned %d providers, want 1", len(across))
		}
		assertModelCapacities(t, across[0], 96_000, 0)

		got.Models[0].ContextWindowTokens = 128_000
		updated, err := repo.UpdateModelProvider(ctx, workspaceID, got)
		if err != nil {
			t.Fatalf("UpdateModelProvider: %v", err)
		}
		assertModelCapacities(t, updated, 128_000, 0)

		reloaded, err := repo.GetModelProvider(ctx, workspaceID, provider.GetName())
		if err != nil {
			t.Fatalf("GetModelProvider after update: %v", err)
		}
		assertModelCapacities(t, reloaded, 128_000, 0)
	})
}

func assertModelCapacities(t *testing.T, provider *agentsv1.ModelProvider, first, second uint32) {
	t.Helper()
	if len(provider.GetModels()) != 2 {
		t.Fatalf("provider has %d models, want 2", len(provider.GetModels()))
	}
	if got := provider.GetModels()[0].GetContextWindowTokens(); got != first {
		t.Fatalf("first context_window_tokens = %d, want %d", got, first)
	}
	if got := provider.GetModels()[1].GetContextWindowTokens(); got != second {
		t.Fatalf("second context_window_tokens = %d, want %d", got, second)
	}
}
