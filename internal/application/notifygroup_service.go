package application

import (
	"context"
	"fmt"
	"strings"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type NotifyGroupServiceServer struct {
	repo configrepo.NotifyGroupRepository
}

func NewNotifyGroupServiceServer(repo configrepo.NotifyGroupRepository) *NotifyGroupServiceServer {
	return &NotifyGroupServiceServer{repo: repo}
}

func validateNotifyGroup(group *agentsv1.NotifyGroup) error {
	if group == nil {
		return connectx.RequiredArgument("notify_group")
	}
	if strings.TrimSpace(group.GetName()) == "" {
		return connectx.RequiredArgument("notify_group.name")
	}
	for i, target := range group.GetTargets() {
		field := fmt.Sprintf("notify_group.targets[%d]", i)
		if target == nil {
			return connectx.RequiredArgument(field)
		}
		switch target.GetType() {
		case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM:
			if target.GetTelegram() == nil {
				return connectx.RequiredArgument(field + ".telegram")
			}
			if strings.TrimSpace(target.GetTelegram().GetBotToken()) == "" {
				return connectx.RequiredArgument(field + ".telegram.bot_token")
			}
			if strings.TrimSpace(target.GetTelegram().GetChatId()) == "" {
				return connectx.RequiredArgument(field + ".telegram.chat_id")
			}
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
