package application

// Lifecycle tests (issue #264/#269): an Agent or Model a Telegram Destination
// routes to cannot be removed out from under it, and the error says which
// Destinations to repair.

import (
	"strings"
	"testing"

	"connectrpc.com/connect"

	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func guardWithDestination(t *testing.T, config *agentsv1.TelegramDestinationConfig) *TelegramReferenceGuard {
	t.Helper()
	repo := telegrammemory.New()
	if _, err := repo.CreateDestination(t.Context(), "ws-a", &agentsv1.TelegramDestination{
		Id: "dest-1", Key: "incidents", ChannelId: "ch-1", ChatId: "-100",
		InboundEnabled: true, OutboundEnabled: true, Config: config,
	}); err != nil {
		t.Fatalf("seed destination: %v", err)
	}
	return NewTelegramReferenceGuard(repo)
}

func TestAgentIsProtectedInEveryReferencePosition(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *agentsv1.TelegramDestinationConfig
	}{
		{"default", &agentsv1.TelegramDestinationConfig{AgentId: "support"}},
		{
			"selectable",
			&agentsv1.TelegramDestinationConfig{
				AgentId:            "other",
				SelectableAgentIds: []string{"other", "support"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			guard := guardWithDestination(t, tc.config)

			err := guard.CheckAgentRemovable(t.Context(), "ws-a", "support")
			if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
				t.Fatalf("code = %v, want FailedPrecondition", code)
			}
			if !strings.Contains(err.Error(), "dest-1") {
				t.Errorf("error should name the destination, got %q", err.Error())
			}
		})
	}
}

func TestUnreferencedAgentIsRemovable(t *testing.T) {
	guard := guardWithDestination(t, &agentsv1.TelegramDestinationConfig{AgentId: "support"})

	if err := guard.CheckAgentRemovable(t.Context(), "ws-a", "unused"); err != nil {
		t.Fatalf("an unreferenced agent was blocked: %v", err)
	}
	// A Destination in another workspace cannot block this one.
	if err := guard.CheckAgentRemovable(t.Context(), "ws-b", "support"); err != nil {
		t.Fatalf("a cross-workspace reference blocked the delete: %v", err)
	}
}

func TestModelIsProtectedInEveryReferencePosition(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *agentsv1.TelegramDestinationConfig
	}{
		{"explicit default", &agentsv1.TelegramDestinationConfig{AgentId: "a", Model: "pro"}},
		{
			"selectable",
			&agentsv1.TelegramDestinationConfig{
				AgentId: "a", Model: "fast", SelectableModels: []string{"fast", "pro"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			guard := guardWithDestination(t, tc.config)

			err := guard.CheckModelsRemovable(t.Context(), "ws-a", []string{"pro"})
			if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
				t.Fatalf("code = %v, want FailedPrecondition", code)
			}
			if !strings.Contains(err.Error(), "dest-1") || !strings.Contains(err.Error(), "pro") {
				t.Errorf("error should name the model and destination, got %q", err.Error())
			}
		})
	}
}

func TestUnreferencedModelIsRemovable(t *testing.T) {
	guard := guardWithDestination(t, &agentsv1.TelegramDestinationConfig{AgentId: "a", Model: "fast"})

	if err := guard.CheckModelsRemovable(t.Context(), "ws-a", []string{"unused"}); err != nil {
		t.Fatalf("an unreferenced model was blocked: %v", err)
	}
}

// Editing a provider to drop a model breaks a reference just as surely as
// deleting the provider would, so the same aliases must be reported.
func TestRemovedAliasesReportsWhatAnUpdateWouldDrop(t *testing.T) {
	prev := &agentsv1.ModelProvider{Models: []*agentsv1.ModelConfig{
		{Name: "gpt-5", Alias: "fast"},
		{Name: "gpt-5-pro", Alias: "pro"},
	}}
	next := &agentsv1.ModelProvider{Models: []*agentsv1.ModelConfig{
		{Name: "gpt-5", Alias: "fast"},
	}}

	removed := RemovedAliases(prev, next)
	if !strings.Contains(strings.Join(removed, ","), "pro") {
		t.Fatalf("removed = %v, want the dropped alias reported", removed)
	}
	if strings.Contains(strings.Join(removed, ","), "fast") {
		t.Errorf("removed = %v, a retained alias must not be reported", removed)
	}
}

// A nil guard must never block: services wired without Telegram keep working.
func TestNilGuardIsPermissive(t *testing.T) {
	var guard *TelegramReferenceGuard
	if err := guard.CheckAgentRemovable(t.Context(), "ws-a", "support"); err != nil {
		t.Fatalf("nil guard blocked an agent delete: %v", err)
	}
	if err := guard.CheckModelsRemovable(t.Context(), "ws-a", []string{"pro"}); err != nil {
		t.Fatalf("nil guard blocked a model delete: %v", err)
	}
}
