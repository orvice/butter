package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/repo/auth"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/repo/linearstate"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// linearInstallStateTTL bounds how long an install flow may take in the
// browser.
const linearInstallStateTTL = 10 * time.Minute

// linearRevokeTimeout bounds the best-effort revoke on removal.
const linearRevokeTimeout = 10 * time.Second

// Install callback outcomes, reported to the dashboard as `reason`.
const (
	linearInstallDenied         = "denied"
	linearInstallStateInvalid   = "state_invalid"
	linearInstallAppMissing     = "app_missing"
	linearInstallExchangeFailed = "exchange_failed"
	linearInstallIdentityFailed = "identity_failed"
	linearInstallStorageFailed  = "storage_failed"
)

// SetInstallStateRepo wires the single-use install state store.
func (s *LinearAppServiceServer) SetInstallStateRepo(repo linearstate.Repository) { s.states = repo }

// SetLinearClient overrides the Linear API client. Used by tests.
func (s *LinearAppServiceServer) SetLinearClient(client *linearapi.Client) {
	if client != nil {
		s.linear = client
	}
}

// SetClock overrides the service clock. Used by tests.
func (s *LinearAppServiceServer) SetClock(now func() time.Time) {
	if now != nil {
		s.clock = now
	}
}

// SetDashboardBaseURL wires the dashboard's origin, which is both an allowed
// return origin and the base of the fallback redirect.
func (s *LinearAppServiceServer) SetDashboardBaseURL(provider func() string) {
	s.dashboardBaseURL = provider
}

func (s *LinearAppServiceServer) now() time.Time {
	if s.clock != nil {
		return s.clock()
	}
	return time.Now()
}

func (s *LinearAppServiceServer) linearClient() *linearapi.Client {
	if s.linear == nil {
		s.linear = linearapi.New(nil, linearapi.DefaultEndpoints)
	}
	return s.linear
}

func (s *LinearAppServiceServer) dashboardBase() string {
	if s.dashboardBaseURL == nil {
		return ""
	}
	return strings.TrimRight(strings.TrimSpace(s.dashboardBaseURL()), "/")
}

// decryptLinear decrypts one stored credential.
func (s *LinearAppServiceServer) decryptLinear(ctx context.Context, cred linearrepo.Credential) (string, error) {
	if err := s.requireKeyring(); err != nil {
		return "", err
	}
	plain, err := s.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (s *LinearAppServiceServer) BeginLinearInstall(ctx context.Context, req *connect.Request[agentsv1.BeginLinearInstallRequest]) (*connect.Response[agentsv1.BeginLinearInstallResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "linear"); err != nil {
		return nil, err
	}
	if s.states == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("linear install state store not configured"))
	}
	app, err := s.repo.GetApp(ctx, workspaceID, req.Msg.GetAppId())
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	baseURL := s.linearBaseURL(ctx)
	if baseURL == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("a platform admin must configure the Linear public base URL first"))
	}
	if !app.GetClientSecretSet() {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("set the Linear App's client secret before installing it"))
	}
	userID := ""
	if user, ok := auth.UserFromContext(ctx); ok {
		userID = user.GetId()
	}
	state, err := randomLinearState()
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	now := s.now()
	redirectURI := LinearCallbackURL(baseURL)
	if err := s.states.Create(ctx, &linearstate.Entry{
		State:       state,
		WorkspaceID: workspaceID,
		AppID:       app.GetId(),
		UserID:      userID,
		RedirectURI: redirectURI,
		ReturnURL:   s.sanitizeInstallReturnURL(req.Msg.GetReturnUrl(), baseURL, app.GetId()),
		CreatedAt:   now,
		ExpiresAt:   now.Add(linearInstallStateTTL),
	}); err != nil {
		return nil, connectx.InternalWith(err)
	}
	log.FromContext(ctx).Info("linear install started",
		"workspace_id", workspaceID, "app_id", app.GetId(), "user_id", userID)
	return connect.NewResponse(&agentsv1.BeginLinearInstallResponse{
		AuthorizeUrl: s.linearClient().AuthorizeURL(app.GetClientId(), redirectURI, state),
	}), nil
}

// CompleteLinearInstall finishes an install flow for the public OAuth
// callback and returns where to send the browser. It never fails: every
// outcome is a redirect back to the dashboard carrying linear_install and,
// on failure, a machine-readable reason. Nothing is written unless the whole
// exchange succeeds.
func (s *LinearAppServiceServer) CompleteLinearInstall(ctx context.Context, state, code, linearError string) string {
	logger := log.FromContext(ctx)
	if s.states == nil || s.repo == nil {
		return s.installRedirect("", "error", linearInstallStateInvalid, "")
	}
	entry, err := s.states.Consume(ctx, state, s.now())
	if err != nil {
		if !errors.Is(err, linearstate.ErrNotFound) {
			logger.Error("could not consume linear install state", "err", err)
		}
		return s.installRedirect("", "error", linearInstallStateInvalid, "")
	}
	fail := func(reason string, err error) string {
		logger.Warn("linear install failed", "workspace_id", entry.WorkspaceID, "app_id", entry.AppID,
			"reason", reason, "err", err)
		return s.installRedirect(entry.ReturnURL, "error", reason, "")
	}
	if strings.TrimSpace(linearError) != "" {
		return fail(linearInstallDenied, errors.New(linearError))
	}
	if strings.TrimSpace(code) == "" {
		return fail(linearInstallExchangeFailed, errors.New("callback carried no code"))
	}

	app, err := s.repo.GetApp(ctx, entry.WorkspaceID, entry.AppID)
	if err != nil {
		return fail(linearInstallAppMissing, err)
	}
	creds, err := s.repo.GetAppCredentials(ctx, entry.WorkspaceID, entry.AppID)
	if err != nil || !creds.ClientSecret.Set() {
		return fail(linearInstallExchangeFailed, fmt.Errorf("client secret unavailable: %v", err))
	}
	clientSecret, err := s.decryptLinear(ctx, creds.ClientSecret)
	if err != nil {
		return fail(linearInstallExchangeFailed, err)
	}
	token, err := s.linearClient().ExchangeCode(ctx, app.GetClientId(), clientSecret, entry.RedirectURI, code)
	if err != nil {
		return fail(linearInstallExchangeFailed, err)
	}
	identity, err := s.linearClient().Identity(ctx, token.AccessToken)
	if err != nil {
		return fail(linearInstallIdentityFailed, err)
	}
	access, err := s.encryptLinearSecret(ctx, "access token", token.AccessToken)
	if err != nil {
		return fail(linearInstallStorageFailed, err)
	}
	refresh, err := s.encryptLinearSecret(ctx, "refresh token", token.RefreshToken)
	if err != nil {
		return fail(linearInstallStorageFailed, err)
	}
	stored, err := s.repo.UpsertInstallation(ctx, entry.WorkspaceID, &agentsv1.LinearInstallation{
		Id:               uuid.NewString(),
		AppId:            app.GetId(),
		OrganizationId:   identity.OrganizationID,
		OrganizationName: identity.OrganizationName,
		AppUserId:        identity.AppUserID,
		Scopes:           token.Scopes,
	}, linearrepo.InstallationTokens{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    token.ExpiresAt(s.now()),
	})
	if err != nil {
		return fail(linearInstallStorageFailed, err)
	}
	logger.Info("linear app installed",
		"workspace_id", entry.WorkspaceID, "app_id", app.GetId(), "installation_id", stored.GetId(),
		"organization_id", stored.GetOrganizationId(), "app_user_id", stored.GetAppUserId(),
		"user_id", entry.UserID)
	return s.installRedirect(entry.ReturnURL, "success", "", stored.GetId())
}

// installRedirect appends the outcome to the return URL, falling back to
// the Linear Apps page when there is none.
func (s *LinearAppServiceServer) installRedirect(returnURL, outcome, reason, installationID string) string {
	if returnURL == "" {
		returnURL = s.dashboardBase() + "/linear-apps"
	}
	u, err := url.Parse(returnURL)
	if err != nil {
		u = &url.URL{Path: "/linear-apps"}
	}
	q := u.Query()
	q.Set("linear_install", outcome)
	if reason != "" {
		q.Set("reason", reason)
	}
	if installationID != "" {
		q.Set("installation_id", installationID)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// sanitizeInstallReturnURL accepts a relative path, or an absolute URL on
// the Linear base URL's or the dashboard's origin. Anything else — an open
// redirect in waiting — falls back to the App's page.
func (s *LinearAppServiceServer) sanitizeInstallReturnURL(raw, baseURL, appID string) string {
	fallback := s.dashboardBase() + "/linear-apps/" + appID + "/edit"
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fallback
	}
	if u.Scheme == "" && u.Host == "" {
		if strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//") && !strings.HasPrefix(u.Path, "/\\") {
			return raw
		}
		return fallback
	}
	for _, origin := range []string{baseURL, s.dashboardBase()} {
		if sameOrigin(u, origin) {
			return raw
		}
	}
	return fallback
}

func sameOrigin(u *url.URL, origin string) bool {
	if origin == "" || u.Scheme == "" || u.Host == "" {
		return false
	}
	o, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Scheme, o.Scheme) && strings.EqualFold(u.Host, o.Host)
}

func randomLinearState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate linear install state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *LinearAppServiceServer) ListLinearInstallations(ctx context.Context, req *connect.Request[agentsv1.ListLinearInstallationsRequest]) (*connect.Response[agentsv1.ListLinearInstallationsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetApp(ctx, workspaceID, req.Msg.GetAppId()); err != nil {
		return nil, mapLinearRepoErr(err)
	}
	installs, err := s.repo.ListInstallations(ctx, workspaceID, req.Msg.GetAppId())
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.ListLinearInstallationsResponse{Installations: installs}), nil
}

func (s *LinearAppServiceServer) DeleteLinearInstallation(ctx context.Context, req *connect.Request[agentsv1.DeleteLinearInstallationRequest]) (*connect.Response[agentsv1.DeleteLinearInstallationResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "linear"); err != nil {
		return nil, err
	}
	inst, err := s.repo.GetInstallation(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	if inst.GetAppId() != req.Msg.GetAppId() {
		return nil, connectx.NotFound(fmt.Sprintf("linear installation %q not found", req.Msg.GetId()))
	}
	s.revokeInstallationBestEffort(ctx, workspaceID, inst)
	if err := s.repo.DeleteInstallation(ctx, workspaceID, inst.GetId()); err != nil {
		return nil, mapLinearRepoErr(err)
	}
	log.FromContext(ctx).Info("linear installation removed",
		"workspace_id", workspaceID, "app_id", inst.GetAppId(), "installation_id", inst.GetId(),
		"organization_id", inst.GetOrganizationId())
	return connect.NewResponse(&agentsv1.DeleteLinearInstallationResponse{}), nil
}

// revokeInstallationBestEffort revokes the installation's token at Linear.
// Removal proceeds whatever the outcome: the operator asked for the
// installation to be gone from Butter, and a token Butter no longer stores
// is one it can no longer use.
func (s *LinearAppServiceServer) revokeInstallationBestEffort(ctx context.Context, workspaceID string, inst *agentsv1.LinearInstallation) {
	logger := log.FromContext(ctx)
	tokens, err := s.repo.GetInstallationTokens(ctx, workspaceID, inst.GetId())
	if err != nil || !tokens.AccessToken.Set() {
		return
	}
	access, err := s.decryptLinear(ctx, tokens.AccessToken)
	if err != nil {
		logger.Warn("could not decrypt linear token to revoke it", "installation_id", inst.GetId(), "err", err)
		return
	}
	revokeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), linearRevokeTimeout)
	defer cancel()
	if err := s.linearClient().Revoke(revokeCtx, access); err != nil {
		logger.Warn("could not revoke linear token; removing the installation anyway",
			"installation_id", inst.GetId(), "err", err)
	}
}
