package application

import (
	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"go.orx.me/apps/butter/internal/planeapi"
	authrepo "go.orx.me/apps/butter/internal/repo/auth"
	planeinstallation "go.orx.me/apps/butter/internal/repo/planeinstallation"
	planestate "go.orx.me/apps/butter/internal/repo/planestate"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"net/url"
	"strings"
	"time"
)

const planeInstallTTL = 10 * time.Minute

func (s *PlaneAppServiceServer) SetPlaneInstallStateRepo(r planestate.Repository) {
	s.installStates = r
}
func (s *PlaneAppServiceServer) SetPlaneInstallationRepo(r planeinstallation.Repository) {
	s.installRepo = r
}
func (s *PlaneAppServiceServer) SetPlaneHTTPClient(c *planeapi.Client)    { s.planeHTTPClient = c }
func (s *PlaneAppServiceServer) SetPlaneClock(now func() time.Time)       { s.clock = now }
func (s *PlaneAppServiceServer) SetPlaneDashboardBaseURL(f func() string) { s.dashboardBaseURL = f }
func (s *PlaneAppServiceServer) planeNow() time.Time {
	if s.clock != nil {
		return s.clock()
	}
	return time.Now()
}
func randomPlaneState() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func (s *PlaneAppServiceServer) BeginPlaneInstall(ctx context.Context, req *connect.Request[agentsv1.BeginPlaneInstallRequest]) (*connect.Response[agentsv1.BeginPlaneInstallResponse], error) {
	if e := s.requireReady(); e != nil {
		return nil, e
	}
	if s.installStates == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane install state store not configured"))
	}
	if s.installRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane installation storage is not configured"))
	}
	workspaceID, e := requireWorkspace(ctx)
	if e != nil {
		return nil, e
	}
	user, ok := authrepo.UserFromContext(ctx)
	if !ok || user.GetId() == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authenticated initiating user is required"))
	}
	if e = requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "plane"); e != nil {
		return nil, e
	}
	app, e := s.repo.GetApp(ctx, workspaceID, req.Msg.GetAppId())
	if e != nil {
		return nil, mapPlaneRepoErr(e)
	}
	creds, e := s.repo.GetAppCredentials(ctx, workspaceID, app.GetId())
	if e != nil || !creds.ClientSecret.Set() {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("set the Plane App client secret before installing it"))
	}
	profile, e := s.requireProfile()
	if e != nil {
		return nil, e
	}
	if app.GetProviderProfile() != profile.Name {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("Plane App provider profile does not match this deployment"))
	}
	if s.settingsRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("configure the Plane public base URL before installing"))
	}
	settings, e := s.settingsRepo.Get(ctx)
	if e != nil {
		return nil, connectx.InternalWith(e)
	}
	publicBase := strings.TrimRight(strings.TrimSpace(settings.GetPublicBaseUrl()), "/")
	if e = planeapi.ValidateOrigin(publicBase); e != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("configure the Plane public base URL before installing"))
	}
	state, e := randomPlaneState()
	if e != nil {
		return nil, connectx.InternalWith(e)
	}
	base := publicBase
	if s.dashboardBaseURL != nil {
		if dashboard := strings.TrimRight(strings.TrimSpace(s.dashboardBaseURL()), "/"); dashboard != "" {
			base = dashboard
		}
	}
	redirect := publicBase + "/api/plane/oauth/callback"
	ret := sanitizePlaneReturn(req.Msg.GetReturnUrl(), base, app.GetId())
	now := s.planeNow()
	if e = s.installStates.Create(ctx, &planestate.Entry{State: state, WorkspaceID: workspaceID, AppID: app.GetId(), ClientID: app.GetClientId(), UserID: user.GetId(), RedirectURI: redirect, ReturnURL: ret, CreatedAt: now, ExpiresAt: now.Add(planeInstallTTL)}); e != nil {
		return nil, connectx.InternalWith(e)
	}
	log.FromContext(ctx).Info("plane install started", "workspace_id", workspaceID, "app_id", app.GetId())
	return connect.NewResponse(&agentsv1.BeginPlaneInstallResponse{AuthorizeUrl: planeapi.NewClient(nil, profile).AuthorizeURL(app.GetClientId(), redirect, state)}), nil
}
func sanitizePlaneReturn(raw, base, app string) string {
	fallback := base + "/plane-apps/" + app + "/edit"
	u, e := url.Parse(strings.TrimSpace(raw))
	if e != nil || u.String() == "" || strings.Contains(raw, "\\") {
		return fallback
	}
	if u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//") {
		return strings.TrimRight(base, "/") + u.String()
	}
	b, _ := url.Parse(base)
	if b != nil && u.Scheme == b.Scheme && u.Host == b.Host {
		return u.String()
	}
	return fallback
}
