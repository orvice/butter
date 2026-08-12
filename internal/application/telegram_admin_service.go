package application

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"connectrpc.com/connect"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/telegramsetting"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// TelegramAdminServiceServer implements
// agentsv1connect.TelegramAdminServiceHandler (issue #264).
//
// The Webhook base URL is platform-level, not workspace-level: it names the
// public address of this deployment behind its load balancer. Restricting it
// to global admins is what stops a workspace owner from pointing another
// tenant's Telegram callbacks somewhere else.
type TelegramAdminServiceServer struct {
	repo telegramsetting.Repository
}

func NewTelegramAdminServiceServer(repo telegramsetting.Repository) *TelegramAdminServiceServer {
	return &TelegramAdminServiceServer{repo: repo}
}

func (s *TelegramAdminServiceServer) SetRepo(repo telegramsetting.Repository) { s.repo = repo }

func (s *TelegramAdminServiceServer) requireReady() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("telegram settings repository not configured"))
	}
	return nil
}

func (s *TelegramAdminServiceServer) GetTelegramSettings(ctx context.Context, _ *connect.Request[agentsv1.GetTelegramSettingsRequest]) (*connect.Response[agentsv1.GetTelegramSettingsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	if !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.GetTelegramSettingsResponse{Settings: settings}), nil
}

func (s *TelegramAdminServiceServer) UpdateTelegramSettings(ctx context.Context, req *connect.Request[agentsv1.UpdateTelegramSettingsRequest]) (*connect.Response[agentsv1.UpdateTelegramSettingsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	if !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}

	baseURL := strings.TrimRight(strings.TrimSpace(req.Msg.GetSettings().GetWebhookBaseUrl()), "/")
	if baseURL != "" {
		parsed, err := url.Parse(baseURL)
		if err != nil || parsed.Host == "" {
			return nil, connectx.InvalidArgument("settings.webhook_base_url",
				"must be an absolute URL, e.g. https://butter.example.com")
		}
		// Telegram refuses to deliver to plain HTTP, so accepting one here
		// would produce a registration that silently never works.
		if parsed.Scheme != "https" {
			return nil, connectx.InvalidArgument("settings.webhook_base_url",
				"must use https: Telegram only delivers webhooks over TLS")
		}
		if parsed.Path != "" && parsed.Path != "/" {
			return nil, connectx.InvalidArgument("settings.webhook_base_url",
				"must not include a path; callback paths are derived per channel")
		}
	}

	stored, err := s.repo.Put(ctx, &agentsv1.TelegramSettings{WebhookBaseUrl: baseURL})
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	log.FromContext(ctx).Info("telegram platform settings updated",
		"audit", "telegram_settings_update", "webhook_base_url", baseURL)
	return connect.NewResponse(&agentsv1.UpdateTelegramSettingsResponse{Settings: stored}), nil
}
