package repotest

import (
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
	ctx := t.Context()

	for _, api := range []string{"", "chat_completions", "responses"} {
		t.Run("api_round_trip/"+api, func(t *testing.T) {
			repo := factory(t)
			provider := &agentsv1.ModelProvider{Name: "protocol", Type: "openai", OpenaiApi: api}
			created, err := repo.CreateModelProvider(t.Context(), workspaceID, provider)
			if err != nil {
				t.Fatal(err)
			}
			got, err := repo.GetModelProvider(t.Context(), workspaceID, provider.Name)
			if err != nil {
				t.Fatal(err)
			}
			listed, err := repo.ListModelProviders(t.Context(), workspaceID)
			if err != nil || len(listed) != 1 {
				t.Fatalf("list = (%v, %v), want one provider", listed, err)
			}
			across, err := repo.ListModelProvidersAcrossWorkspaces(t.Context())
			if err != nil || len(across) != 1 {
				t.Fatalf("list across workspaces = (%v, %v), want one provider", across, err)
			}
			for _, p := range []*agentsv1.ModelProvider{created, got, listed[0], across[0]} {
				if p.GetOpenaiApi() != api {
					t.Errorf("openai_api = %q, want %q", p.GetOpenaiApi(), api)
				}
			}
			got.OpenaiApi = "responses"
			if api == "responses" {
				got.OpenaiApi = "chat_completions"
			}
			updated, err := repo.UpdateModelProvider(t.Context(), workspaceID, got)
			if err != nil {
				t.Fatal(err)
			}
			reloaded, err := repo.GetModelProvider(t.Context(), workspaceID, provider.Name)
			if err != nil || reloaded.GetOpenaiApi() != got.OpenaiApi || updated.GetOpenaiApi() != got.OpenaiApi {
				t.Fatalf("API update was not persisted: %v, %v", reloaded, err)
			}
		})
	}

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
