package application

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"

	"connectrpc.com/connect"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/linearsetting"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// LinearAdminServiceServer implements agentsv1connect.LinearAdminServiceHandler
// (ADR-0015). The public base URL is platform-level: restricting it to global
// admins is what stops a workspace owner from pointing another tenant's
// Linear webhooks and OAuth callbacks somewhere else.
type LinearAdminServiceServer struct {
	repo linearsetting.Repository
}

func NewLinearAdminServiceServer(repo linearsetting.Repository) *LinearAdminServiceServer {
	return &LinearAdminServiceServer{repo: repo}
}

func (s *LinearAdminServiceServer) SetRepo(repo linearsetting.Repository) { s.repo = repo }

func (s *LinearAdminServiceServer) requireAdmin(ctx context.Context) error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("linear settings repository not configured"))
	}
	if !auth.IsAdmin(ctx) {
		return connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	return nil
}

func (s *LinearAdminServiceServer) GetLinearSettings(ctx context.Context, _ *connect.Request[agentsv1.GetLinearSettingsRequest]) (*connect.Response[agentsv1.GetLinearSettingsResponse], error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.GetLinearSettingsResponse{Settings: settings}), nil
}

func (s *LinearAdminServiceServer) UpdateLinearSettings(ctx context.Context, req *connect.Request[agentsv1.UpdateLinearSettingsRequest]) (*connect.Response[agentsv1.UpdateLinearSettingsResponse], error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(req.Msg.GetSettings().GetPublicBaseUrl()), "/")
	if baseURL != "" {
		if err := validateLinearBaseURL(baseURL); err != nil {
			return nil, err
		}
	}
	stored, err := s.repo.Put(ctx, &agentsv1.LinearSettings{PublicBaseUrl: baseURL})
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	log.FromContext(ctx).Info("linear platform settings updated",
		"audit", "linear_settings_update", "public_base_url", baseURL)
	return connect.NewResponse(&agentsv1.UpdateLinearSettingsResponse{Settings: stored}), nil
}

// validateLinearBaseURL accepts an absolute https URL without a path, or a
// plain-http loopback URL for local development.
func validateLinearBaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return connectx.InvalidArgument("settings.public_base_url",
			"must be an absolute URL, e.g. https://butter.example.com")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return connectx.InvalidArgument("settings.public_base_url",
			"must not include a path; callback and webhook paths are derived")
	}
	switch parsed.Scheme {
	case "https":
		return nil
	case "http":
		if isLoopbackHost(parsed.Hostname()) {
			return nil
		}
	}
	return connectx.InvalidArgument("settings.public_base_url",
		"must use https (plain http is only accepted for localhost)")
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
