package agent

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func makeProviders() []agentsv1.ModelProvider {
	return []agentsv1.ModelProvider{
		{
			Name: "gemini",
			Type: "gemini",
			Models: []*agentsv1.ModelConfig{
				{Name: "gemini-2.5-flash", Alias: "flash"},
				{Name: "gemini-2.5-pro", Alias: "pro"},
			},
		},
		{
			Name: "openai",
			Type: "openai",
			Models: []*agentsv1.ModelConfig{
				{Name: "gpt-4o", Alias: "4o"},
				{Name: "gpt-4o-mini"}, // no alias
			},
		},
	}
}

func TestResolveModelAlias_ByAlias(t *testing.T) {
	providers := makeProviders()

	name, found := ResolveModelAlias("flash", providers)
	if !found {
		t.Fatal("expected to find model by alias 'flash'")
	}
	if name != "gemini-2.5-flash" {
		t.Errorf("expected 'gemini-2.5-flash', got %q", name)
	}
}

func TestResolveModelAlias_ByName(t *testing.T) {
	providers := makeProviders()

	name, found := ResolveModelAlias("gpt-4o-mini", providers)
	if !found {
		t.Fatal("expected to find model by name 'gpt-4o-mini'")
	}
	if name != "gpt-4o-mini" {
		t.Errorf("expected 'gpt-4o-mini', got %q", name)
	}
}

func TestResolveModelAlias_NotFound(t *testing.T) {
	providers := makeProviders()

	name, found := ResolveModelAlias("nonexistent", providers)
	if found {
		t.Fatal("expected not to find 'nonexistent'")
	}
	if name != "nonexistent" {
		t.Errorf("expected original ref 'nonexistent', got %q", name)
	}
}

func TestResolveModelAlias_AliasOverName(t *testing.T) {
	// If an alias matches, it should take priority even if another model has
	// that string as its name.
	providers := []agentsv1.ModelProvider{
		{
			Name: "p1",
			Type: "gemini",
			Models: []*agentsv1.ModelConfig{
				{Name: "model-a", Alias: "special"},
			},
		},
		{
			Name: "p2",
			Type: "openai",
			Models: []*agentsv1.ModelConfig{
				{Name: "special"}, // name matches the alias of p1's model
			},
		},
	}

	name, found := ResolveModelAlias("special", providers)
	if !found {
		t.Fatal("expected to find model")
	}
	// Should resolve to p1's model (alias match), not p2's model (name match).
	if name != "model-a" {
		t.Errorf("expected 'model-a' (alias match), got %q", name)
	}
}

func TestValidateModelAliases_OK(t *testing.T) {
	providers := makeProviders()
	if err := ValidateModelAliases(providers); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateModelAliases_Duplicate(t *testing.T) {
	providers := []agentsv1.ModelProvider{
		{
			Name: "p1",
			Type: "gemini",
			Models: []*agentsv1.ModelConfig{
				{Name: "model-a", Alias: "pro"},
			},
		},
		{
			Name: "p2",
			Type: "openai",
			Models: []*agentsv1.ModelConfig{
				{Name: "model-b", Alias: "pro"},
			},
		},
	}

	err := ValidateModelAliases(providers)
	if err == nil {
		t.Fatal("expected error for duplicate alias 'pro'")
	}
}

func TestAllModelAliases(t *testing.T) {
	providers := makeProviders()
	infos := AllModelAliases(providers)

	expected := []ModelInfo{
		{Alias: "flash", Name: "gemini-2.5-flash"},
		{Alias: "pro", Name: "gemini-2.5-pro"},
		{Alias: "4o", Name: "gpt-4o"},
		{Alias: "gpt-4o-mini", Name: "gpt-4o-mini"}, // no alias, falls back to name
	}

	if len(infos) != len(expected) {
		t.Fatalf("expected %d model infos, got %d", len(expected), len(infos))
	}

	for i, info := range infos {
		if info.Alias != expected[i].Alias || info.Name != expected[i].Name {
			t.Errorf("index %d: expected {%s, %s}, got {%s, %s}",
				i, expected[i].Alias, expected[i].Name, info.Alias, info.Name)
		}
	}
}
