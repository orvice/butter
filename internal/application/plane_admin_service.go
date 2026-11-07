package application

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/planeapi"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/planesetting"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// PlaneAdminServiceServer implements agentsv1connect.PlaneAdminServiceHandler
// (spec #487, ticket #497).
//
// The public base URL is platform-level: restricting it to global admins is
// what stops a workspace owner from pointing another tenant's Plane callbacks
// somewhere else. The provider profile is read-only here — it comes from
// startup configuration, so no RPC can change the origins the deployment
// talks to.
type PlaneAdminServiceServer struct {
	repo    planesetting.Repository
	profile func() (planeapi.Profile, error)
}

func NewPlaneAdminServiceServer(repo planesetting.Repository) *PlaneAdminServiceServer {
	return &PlaneAdminServiceServer{repo: repo}
}

func (s *PlaneAdminServiceServer) SetRepo(repo planesetting.Repository) { s.repo = repo }

// SetProfileProvider wires the fixed provider profile built from startup
// configuration.
func (s *PlaneAdminServiceServer) SetProfileProvider(provider func() (planeapi.Profile, error)) {
	s.profile = provider
}

func (s *PlaneAdminServiceServer) requireAdmin(ctx context.Context) error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("plane settings repository not configured"))
	}
	if !auth.IsAdmin(ctx) {
		return connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	return nil
}

// profileProto returns the fixed profile, or a FailedPrecondition error when
// it is unconfigured or invalid.
func (s *PlaneAdminServiceServer) profileProto() (*agentsv1.PlaneProviderProfile, error) {
	if s.profile == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("plane provider profile is not configured"))
	}
	profile, err := s.profile()
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("the plane provider profile is misconfigured"))
	}
	return planeProfileProto(profile), nil
}

func (s *PlaneAdminServiceServer) GetPlaneSettings(ctx context.Context, _ *connect.Request[agentsv1.GetPlaneSettingsRequest]) (*connect.Response[agentsv1.GetPlaneSettingsResponse], error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	profile, err := s.profileProto()
	if err != nil {
		return nil, err
	}
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.GetPlaneSettingsResponse{Settings: settings, Provider: profile}), nil
}

func (s *PlaneAdminServiceServer) UpdatePlaneSettings(ctx context.Context, req *connect.Request[agentsv1.UpdatePlaneSettingsRequest]) (*connect.Response[agentsv1.UpdatePlaneSettingsResponse], error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	profile, err := s.profileProto()
	if err != nil {
		return nil, err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(req.Msg.GetSettings().GetPublicBaseUrl()), "/")
	if baseURL != "" {
		if err := planeapi.ValidateOrigin(baseURL); err != nil {
			return nil, connectx.InvalidArgument("settings.public_base_url", err.Error())
		}
	}
	stored, err := s.repo.Put(ctx, &agentsv1.PlaneSettings{PublicBaseUrl: baseURL})
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	log.FromContext(ctx).Info("plane platform settings updated",
		"audit", "plane_settings_update", "public_base_url", baseURL)
	return connect.NewResponse(&agentsv1.UpdatePlaneSettingsResponse{Settings: stored, Provider: profile}), nil
}
