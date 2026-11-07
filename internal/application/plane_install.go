package application

import (
	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"go.orx.me/apps/butter/internal/planeapi"
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
	workspaceID, e := requireWorkspace(ctx)
	if e != nil {
		return nil, e
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
	state, e := randomPlaneState()
	if e != nil {
		return nil, connectx.InternalWith(e)
	}
	base := ""
	if s.dashboardBaseURL != nil {
		base = strings.TrimRight(strings.TrimSpace(s.dashboardBaseURL()), "/")
	}
	redirect := base + "/api/plane/oauth/callback"
	ret := sanitizePlaneReturn(req.Msg.GetReturnUrl(), base, app.GetId())
	now := s.planeNow()
	if e = s.installStates.Create(ctx, &planestate.Entry{State: state, WorkspaceID: workspaceID, AppID: app.GetId(), ClientID: app.GetClientId(), UserID: "", RedirectURI: redirect, ReturnURL: ret, CreatedAt: now, ExpiresAt: now.Add(planeInstallTTL)}); e != nil {
		return nil, connectx.InternalWith(e)
	}
	log.FromContext(ctx).Info("plane install started", "workspace_id", workspaceID, "app_id", app.GetId())
	return connect.NewResponse(&agentsv1.BeginPlaneInstallResponse{AuthorizeUrl: planeapi.NewClient(nil, profile).AuthorizeURL(app.GetClientId(), redirect, state)}), nil
}
func sanitizePlaneReturn(raw, base, app string) string {
	fallback := base + "/plane-apps/" + app
	u, e := url.Parse(strings.TrimSpace(raw))
	if e != nil || u.String() == "" {
		return fallback
	}
	if u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//") {
		return u.String()
	}
	b, _ := url.Parse(base)
	if b != nil && u.Scheme == b.Scheme && u.Host == b.Host {
		return u.String()
	}
	return fallback
}
