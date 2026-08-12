package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type NotifyGroupServiceServer struct {
	repo configrepo.NotifyGroupRepository
	// telegramRepo resolves Telegram Destination references so a group can
	// never be saved pointing at an address that does not exist, belongs to
	// another workspace, or cannot send.
	telegramRepo telegramrepo.Repository
}

func NewNotifyGroupServiceServer(repo configrepo.NotifyGroupRepository) *NotifyGroupServiceServer {
	return &NotifyGroupServiceServer{repo: repo}
}

// SetTelegramRepo wires Destination resolution after bootstrap.
func (s *NotifyGroupServiceServer) SetTelegramRepo(repo telegramrepo.Repository) {
	s.telegramRepo = repo
}

func validateNotifyGroup(group *agentsv1.NotifyGroup) error {
	if group == nil {
		return connectx.RequiredArgument("notify_group")
	}
	if strings.TrimSpace(group.GetName()) == "" {
		return connectx.RequiredArgument("notify_group.name")
	}
	// Two targets pointing at the same Telegram Destination would deliver the
	// same alert twice to the same chat. The same Destination in *different*
	// groups is fine and intentional.
	seenDestinations := make(map[string]int)
	for i, target := range group.GetTargets() {
		field := fmt.Sprintf("notify_group.targets[%d]", i)
		if target == nil {
			return connectx.RequiredArgument(field)
		}
		switch target.GetType() {
		case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM:
			telegram := target.GetTelegram()
			if telegram == nil {
				return connectx.RequiredArgument(field + ".telegram")
			}
			// The legacy raw-address fields are refused rather than ignored:
			// silently dropping a Bot Token an operator just pasted would
			// leave them believing it is in use (issue #264).
			if err := rejectLegacyTelegramTargetFields(field, telegram); err != nil {
				return err
			}
			destinationID := strings.TrimSpace(telegram.GetDestinationId())
			if destinationID == "" {
				return connectx.RequiredArgument(field + ".telegram.destination_id")
			}
			if prev, dup := seenDestinations[destinationID]; dup {
				return connectx.InvalidArgument(field+".telegram.destination_id",
					fmt.Sprintf("duplicates target %d in the same group", prev))
			}
			seenDestinations[destinationID] = i
		case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK:
			if target.GetLark() == nil {
				return connectx.RequiredArgument(field + ".lark")
			}
			if strings.TrimSpace(target.GetLark().GetWebhookUrl()) == "" {
				return connectx.RequiredArgument(field + ".lark.webhook_url")
			}
		case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK:
			if target.GetDiscord() == nil {
				return connectx.RequiredArgument(field + ".discord")
			}
			if strings.TrimSpace(target.GetDiscord().GetWebhookUrl()) == "" {
				return connectx.RequiredArgument(field + ".discord.webhook_url")
			}
		default:
			return connectx.InvalidArgument(field+".type", fmt.Sprintf("unsupported notify target type %s", target.GetType()))
		}
	}
	return nil
}

// rejectLegacyTelegramTargetFields refuses the pre-#264 raw address fields.
func rejectLegacyTelegramTargetFields(field string, telegram *agentsv1.TelegramNotifyTarget) error {
	//nolint:staticcheck // deliberately reading deprecated fields to refuse them
	switch {
	case strings.TrimSpace(telegram.GetBotToken()) != "":
		return connectx.InvalidArgument(field+".telegram.bot_token",
			"is no longer accepted; reference a telegram destination instead")
	case strings.TrimSpace(telegram.GetChatId()) != "":
		return connectx.InvalidArgument(field+".telegram.chat_id",
			"is no longer accepted; reference a telegram destination instead")
	case strings.TrimSpace(telegram.GetParseMode()) != "":
		return connectx.InvalidArgument(field+".telegram.parse_mode",
			"is no longer accepted; markdown handling is centralized")
	case telegram.GetMessageThreadId() != 0:
		return connectx.InvalidArgument(field+".telegram.message_thread_id",
			"is no longer accepted; the destination owns the forum topic")
	}
	return nil
}

// resolveTelegramTargets checks that every Telegram target names a
// Destination in this workspace that is currently able to send. Resolving at
// write time is what turns "the alert silently went nowhere" into a
// validation error the operator sees while editing.
func (s *NotifyGroupServiceServer) resolveTelegramTargets(ctx context.Context, workspaceID string, group *agentsv1.NotifyGroup) error {
	if s.telegramRepo == nil {
		return nil
	}
	for i, target := range group.GetTargets() {
		if target.GetType() != agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM {
			continue
		}
		field := fmt.Sprintf("notify_group.targets[%d].telegram.destination_id", i)
		destinationID := strings.TrimSpace(target.GetTelegram().GetDestinationId())
		dest, err := s.telegramRepo.GetDestination(ctx, workspaceID, destinationID)
		if err != nil {
			if errors.Is(err, telegramrepo.ErrNotFound) {
				return connectx.InvalidArgument(field, "references an unknown telegram destination")
			}
			return connectx.InternalWith(err)
		}
		if !dest.GetOutboundEnabled() {
			return connectx.InvalidArgument(field,
				fmt.Sprintf("destination %q has outbound delivery disabled", dest.GetKey()))
		}
	}
	return nil
}

func (s *NotifyGroupServiceServer) ListNotifyGroups(ctx context.Context, _ *connect.Request[agentsv1.ListNotifyGroupsRequest]) (*connect.Response[agentsv1.ListNotifyGroupsResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	groups, err := s.repo.ListNotifyGroups(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.ListNotifyGroupsResponse{NotifyGroups: groups}), nil
}

func (s *NotifyGroupServiceServer) GetNotifyGroup(ctx context.Context, req *connect.Request[agentsv1.GetNotifyGroupRequest]) (*connect.Response[agentsv1.GetNotifyGroupResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	group, err := s.repo.GetNotifyGroup(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.GetNotifyGroupResponse{NotifyGroup: group}), nil
}

func (s *NotifyGroupServiceServer) CreateNotifyGroup(ctx context.Context, req *connect.Request[agentsv1.CreateNotifyGroupRequest]) (*connect.Response[agentsv1.CreateNotifyGroupResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateNotifyGroup(req.Msg.GetNotifyGroup()); err != nil {
		return nil, err
	}
	if err := s.resolveTelegramTargets(ctx, wsID, req.Msg.GetNotifyGroup()); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating notify group", "workspace_id", wsID, "name", req.Msg.GetNotifyGroup().GetName())
	group, err := s.repo.CreateNotifyGroup(ctx, wsID, req.Msg.GetNotifyGroup())
	if err != nil {
		logger.Error("create notify group failed", "workspace_id", wsID, "name", req.Msg.GetNotifyGroup().GetName(), "err", err)
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.CreateNotifyGroupResponse{NotifyGroup: group}), nil
}

func (s *NotifyGroupServiceServer) UpdateNotifyGroup(ctx context.Context, req *connect.Request[agentsv1.UpdateNotifyGroupRequest]) (*connect.Response[agentsv1.UpdateNotifyGroupResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateNotifyGroup(req.Msg.GetNotifyGroup()); err != nil {
		return nil, err
	}
	if err := s.resolveTelegramTargets(ctx, wsID, req.Msg.GetNotifyGroup()); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("updating notify group", "workspace_id", wsID, "name", req.Msg.GetNotifyGroup().GetName())
	group, err := s.repo.UpdateNotifyGroup(ctx, wsID, req.Msg.GetNotifyGroup())
	if err != nil {
		logger.Error("update notify group failed", "workspace_id", wsID, "name", req.Msg.GetNotifyGroup().GetName(), "err", err)
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.UpdateNotifyGroupResponse{NotifyGroup: group}), nil
}

func (s *NotifyGroupServiceServer) DeleteNotifyGroup(ctx context.Context, req *connect.Request[agentsv1.DeleteNotifyGroupRequest]) (*connect.Response[agentsv1.DeleteNotifyGroupResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("deleting notify group", "workspace_id", wsID, "name", req.Msg.GetName())
	if err := s.repo.DeleteNotifyGroup(ctx, wsID, req.Msg.GetName()); err != nil {
		logger.Error("delete notify group failed", "workspace_id", wsID, "name", req.Msg.GetName(), "err", err)
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.DeleteNotifyGroupResponse{}), nil
}
