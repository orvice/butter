package application

import (
	"butterfly.orx.me/core/log"
	"context"
	"errors"
	"go.orx.me/apps/butter/internal/planeapi"
	authrepo "go.orx.me/apps/butter/internal/repo/auth"
	planeinstallation "go.orx.me/apps/butter/internal/repo/planeinstallation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"net/url"
	"strings"
)

func (s *PlaneAppServiceServer) CompletePlaneInstall(ctx context.Context, state, code, oauthErr string) string {
	entry, e := s.installStates.Consume(ctx, state, s.planeNow())
	if e != nil {
		return s.planeInstallRedirect("", "error", "state_invalid", "")
	}
	fail := func(reason string, err error) string {
		log.FromContext(ctx).Warn("plane install failed", "workspace_id", entry.WorkspaceID, "app_id", entry.AppID, "reason", reason, "err", err)
		return s.planeInstallRedirect(entry.ReturnURL, "error", reason, "")
	}
	checkCtx := authrepo.WithAuthenticated(ctx, &agentsv1.User{Id: entry.UserID}, nil)
	if e = requireWorkspaceManageRole(checkCtx, s.workspaceRepo, entry.WorkspaceID, "plane"); e != nil {
		return fail("permission_changed", e)
	}
	if strings.TrimSpace(oauthErr) != "" || strings.TrimSpace(code) == "" {
		return fail("denied", errors.New("authorization callback did not contain a code"))
	}
	if s.installRepo == nil || s.repo == nil {
		return fail("storage_failed", errors.New("plane installation storage is not configured"))
	}
	app, e := s.repo.GetApp(ctx, entry.WorkspaceID, entry.AppID)
	if e != nil {
		return fail("app_missing", e)
	}
	if app.GetClientId() != entry.ClientID {
		return fail("app_mismatch", errors.New("state client binding mismatch"))
	}
	creds, e := s.repo.GetAppCredentials(ctx, entry.WorkspaceID, app.GetId())
	if e != nil || !creds.ClientSecret.Set() {
		return fail("exchange_failed", errors.New("client secret unavailable"))
	}
	secret, e := s.keyring.Decrypt(ctx, creds.ClientSecret.Ciphertext, creds.ClientSecret.KeyID)
	if e != nil {
		return fail("exchange_failed", e)
	}
	profile, e := s.requireProfile()
	if e != nil {
		return fail("exchange_failed", e)
	}
	client := s.planeHTTPClient
	if client == nil {
		client = planeapi.NewClient(nil, profile)
	}
	token, e := client.ExchangeCode(ctx, app.GetClientId(), string(secret), code)
	if e != nil {
		return fail("exchange_failed", e)
	}
	identity, e := client.GetInstallation(ctx, token.AccessToken, token.InstallationID)
	if e != nil {
		return fail("identity_failed", e)
	}
	if identity.ID != token.InstallationID || !identity.Installed {
		return fail("identity_failed", errors.New("installation identity mismatch"))
	}
	stored, e := s.installRepo.Upsert(ctx, entry.WorkspaceID, app.GetId(), &agentsv1.PlaneInstallation{AppId: app.GetId(), PlaneWorkspaceId: identity.WorkspaceID, WorkspaceSlug: identity.WorkspaceSlug, AppInstallationId: identity.ID, AppBotUserId: identity.BotUserID, AppBotName: identity.BotName, GrantedScopes: identity.Scopes, Status: "installed", CredentialState: "valid"}, planeinstallation.Tokens{AccessToken: token.AccessToken, ExpiresAt: s.planeNow().Add(token.ExpiresIn)})
	if e != nil {
		return fail("storage_failed", e)
	}
	return s.planeInstallRedirect(entry.ReturnURL, "success", "", stored.GetId())
}
func (s *PlaneAppServiceServer) planeInstallRedirect(raw, outcome, reason, id string) string {
	if raw == "" {
		raw = s.dashboardBaseURLPath()
	}
	u, e := url.Parse(raw)
	if e != nil {
		u = &url.URL{Path: "/plane-apps"}
	}
	q := u.Query()
	q.Set("plane_install", outcome)
	if reason != "" {
		q.Set("reason", reason)
	}
	if id != "" {
		q.Set("installation_id", id)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
func (s *PlaneAppServiceServer) dashboardBaseURLPath() string {
	if s.dashboardBaseURL != nil {
		return strings.TrimRight(s.dashboardBaseURL(), "/") + "/plane-apps"
	}
	return "/plane-apps"
}
