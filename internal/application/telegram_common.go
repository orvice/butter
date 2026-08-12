package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// requireTelegramManageRole grants workspace "owner"/"admin" members and
// global admins (the bypass is audited). Members lacking the role receive
// PermissionDenied; non-members NotFound, which is what the middleware would
// already have returned for a workspace they cannot see.
func requireTelegramManageRole(ctx context.Context, wsRepo workspacerepo.Repository, workspaceID, action string) error {
	if auth.IsAdmin(ctx) {
		if user, ok := auth.UserFromContext(ctx); ok {
			log.FromContext(ctx).Info("global admin managing telegram configuration",
				"audit", "admin_telegram_access", "action", action,
				"workspace_id", workspaceID, "user_id", user.GetId())
		}
		return nil
	}
	if wsRepo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace repository not configured"))
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	member, err := wsRepo.GetMember(ctx, workspaceID, user.GetId())
	if err != nil {
		if errors.Is(err, workspacerepo.ErrNotFound) {
			return connectx.NotFound("workspace not found")
		}
		return connectx.InternalWith(err)
	}
	if slices.Contains([]string{"owner", "admin"}, member.GetRole()) {
		return nil
	}
	return connect.NewError(connect.CodePermissionDenied, errors.New("workspace owner or admin role required"))
}

// telegramWorkspace reads the workspace the request is scoped to.
func telegramWorkspace(ctx context.Context) (string, error) {
	id, err := workspace.MustFromContext(ctx)
	if err != nil {
		return "", connect.NewError(connect.CodeFailedPrecondition, err)
	}
	return id, nil
}

// mapTelegramRepoErr translates repository invariants into Connect codes.
// Each repository error names a distinct operator-visible problem, so they
// map to distinct codes rather than a generic Internal.
func mapTelegramRepoErr(err error) *connect.Error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, telegramrepo.ErrNotFound):
		return connectx.NotFound(err.Error())
	case errors.Is(err, telegramrepo.ErrKeyExists),
		errors.Is(err, telegramrepo.ErrBotExists),
		errors.Is(err, telegramrepo.ErrAddressExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, telegramrepo.ErrRevisionConflict):
		// Aborted is the code Connect clients retry after re-reading, which
		// is exactly the right response to a stale revision.
		return connect.NewError(connect.CodeAborted, err)
	case errors.Is(err, telegramrepo.ErrChannelInUse):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, telegramrepo.ErrNoCredential):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	default:
		return connectx.InternalWith(err)
	}
}

// mapTelegramAPIErr translates a Telegram Bot API failure. A rejected token
// is a caller problem, not an internal one, so it never becomes a 500.
func mapTelegramAPIErr(field string, err error) *connect.Error {
	if errors.Is(err, telegramapi.ErrUnauthorized) {
		return connectx.InvalidArgument(field, "was rejected by Telegram")
	}
	return connect.NewError(connect.CodeFailedPrecondition,
		fmt.Errorf("telegram rejected the request: %w", err))
}

// --- Telegram identifier validation ---------------------------------------

// normalizeTelegramID validates a Telegram int64 identifier supplied as a
// decimal string and returns its canonical form. Canonicalizing on write is
// what makes the (channel, chat, thread) uniqueness index trustworthy: "007"
// and "7" must not be able to register as two Destinations.
func normalizeTelegramID(field, raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", connectx.RequiredArgument(field)
	}
	id, err := telegramapi.ParseID(trimmed)
	if err != nil {
		return "", connectx.InvalidArgument(field, "must be a decimal Telegram identifier")
	}
	return telegramapi.FormatID(id), nil
}

// normalizeThreadID canonicalizes the optional Forum Topic ID. Empty means
// "not a Topic address" and stays empty — it is deliberately distinct from
// any real thread ID, so it is never coerced to "0".
func normalizeThreadID(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	id, err := telegramapi.ParseID(trimmed)
	if err != nil {
		return "", connectx.InvalidArgument("message_thread_id", "must be a decimal Telegram identifier")
	}
	if id <= 0 {
		return "", connectx.InvalidArgument("message_thread_id", "must be positive when present")
	}
	return telegramapi.FormatID(id), nil
}

// normalizeUserIDs canonicalizes and de-duplicates a Telegram user ID list,
// preserving the caller's order so the Dashboard round-trips unchanged.
func normalizeUserIDs(field string, raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, entry := range raw {
		if strings.TrimSpace(entry) == "" {
			continue
		}
		id, err := normalizeTelegramID(field, entry)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

// --- Cross-resource reference validation -----------------------------------

// resolveActiveAgent returns the ACTIVE agent with the given ID in the
// workspace. Destinations hold strong references: a topic that points at a
// tombstoned agent is a silently broken route, so the reference is rejected
// at write time rather than discovered at the first message.
func resolveActiveAgent(ctx context.Context, repo configrepo.AgentRepository, workspaceID, field, agentID string) (*agentsv1.Agent, error) {
	if repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("agent repository not configured"))
	}
	agent, err := repo.GetAgentByID(ctx, workspaceID, agentID)
	if err != nil {
		if errors.Is(err, configrepo.ErrNotFound) {
			return nil, connectx.InvalidArgument(field, fmt.Sprintf("references unknown agent %q", agentID))
		}
		return nil, connectx.InternalWith(err)
	}
	if !isActiveAgent(agent) {
		return nil, connectx.InvalidArgument(field,
			fmt.Sprintf("references agent %q which is not active", agentID))
	}
	return agent, nil
}

// isActiveAgent treats UNSPECIFIED as active so agents written before the
// lifecycle field existed remain usable.
func isActiveAgent(agent *agentsv1.Agent) bool {
	switch agent.GetLifecycleStatus() {
	case agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
		agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_UNSPECIFIED:
		return true
	default:
		return false
	}
}

// workspaceModelAliases collects every model alias configured in the
// workspace. An explicit Destination model must name one of these.
func workspaceModelAliases(ctx context.Context, repo configrepo.ModelProviderRepository, workspaceID string) (map[string]struct{}, error) {
	if repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("model provider repository not configured"))
	}
	providers, err := repo.ListModelProviders(ctx, workspaceID)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	aliases := make(map[string]struct{})
	for _, provider := range providers {
		for _, model := range provider.GetModels() {
			if alias := strings.TrimSpace(model.GetAlias()); alias != "" {
				aliases[alias] = struct{}{}
			}
			if name := strings.TrimSpace(model.GetName()); name != "" {
				aliases[name] = struct{}{}
			}
		}
	}
	return aliases, nil
}
