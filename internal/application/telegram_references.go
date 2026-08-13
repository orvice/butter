package application

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// TelegramReferenceGuard blocks removing an Agent or Model that a Telegram
// Destination still routes to (issue #264/#269).
//
// These are strong references on purpose. A Destination whose Agent vanished
// is a topic that silently stops working: messages arrive, admission passes,
// and nothing answers. Blocking the delete and naming the Destinations turns
// that into a decision the operator makes deliberately.
type TelegramReferenceGuard struct {
	repo telegramrepo.Repository
}

func NewTelegramReferenceGuard(repo telegramrepo.Repository) *TelegramReferenceGuard {
	return &TelegramReferenceGuard{repo: repo}
}

// CheckAgentRemovable returns a FailedPrecondition error naming every
// Destination that uses the Agent as its default or as a selectable choice.
func (g *TelegramReferenceGuard) CheckAgentRemovable(ctx context.Context, workspaceID, agentID string) error {
	if g == nil || g.repo == nil || strings.TrimSpace(agentID) == "" {
		return nil
	}
	destinations, err := g.repo.ListDestinations(ctx, workspaceID, "")
	if err != nil {
		return connectx.InternalWith(err)
	}
	var refs []string
	for _, dest := range destinations {
		config := dest.GetConfig()
		if config.GetAgentId() == agentID || slices.Contains(config.GetSelectableAgentIds(), agentID) {
			refs = append(refs, dest.GetId())
		}
	}
	if len(refs) == 0 {
		return nil
	}
	return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
		"agent %q is referenced by telegram destinations: %s", agentID, strings.Join(refs, ", ")))
}

// CheckModelsRemovable returns a FailedPrecondition error naming every
// Destination that references one of the given model aliases as an explicit
// default or a selectable choice.
//
// It takes the aliases that are going away rather than a provider, so the
// same check covers deleting a provider and editing one to drop a model.
func (g *TelegramReferenceGuard) CheckModelsRemovable(ctx context.Context, workspaceID string, aliases []string) error {
	if g == nil || g.repo == nil || len(aliases) == 0 {
		return nil
	}
	destinations, err := g.repo.ListDestinations(ctx, workspaceID, "")
	if err != nil {
		return connectx.InternalWith(err)
	}
	blocked := make(map[string][]string)
	for _, dest := range destinations {
		config := dest.GetConfig()
		for _, alias := range aliases {
			if config.GetModel() == alias || slices.Contains(config.GetSelectableModels(), alias) {
				blocked[alias] = append(blocked[alias], dest.GetId())
			}
		}
	}
	if len(blocked) == 0 {
		return nil
	}
	parts := make([]string, 0, len(blocked))
	for alias, refs := range blocked {
		parts = append(parts, fmt.Sprintf("%s (%s)", alias, strings.Join(refs, ", ")))
	}
	slices.Sort(parts)
	return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
		"models are referenced by telegram destinations: %s", strings.Join(parts, "; ")))
}

// ModelAliasesOf lists every alias a provider exposes, including bare model
// names, which a Destination may equally reference.
func ModelAliasesOf(provider *agentsv1.ModelProvider) []string {
	var aliases []string
	for _, model := range provider.GetModels() {
		if alias := strings.TrimSpace(model.GetAlias()); alias != "" {
			aliases = append(aliases, alias)
		}
		if name := strings.TrimSpace(model.GetName()); name != "" {
			aliases = append(aliases, name)
		}
	}
	return aliases
}

// RemovedAliases returns the aliases present in prev but absent from next —
// exactly the references an update would break.
func RemovedAliases(prev, next *agentsv1.ModelProvider) []string {
	remaining := make(map[string]struct{})
	for _, alias := range ModelAliasesOf(next) {
		remaining[alias] = struct{}{}
	}
	var removed []string
	for _, alias := range ModelAliasesOf(prev) {
		if _, ok := remaining[alias]; !ok {
			removed = append(removed, alias)
		}
	}
	return removed
}
