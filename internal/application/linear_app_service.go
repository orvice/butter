package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/repo/linearsetting"
	"go.orx.me/apps/butter/internal/repo/linearstate"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// LinearCallbackPath is the public OAuth callback shared by every
	// Linear App; the single-use state identifies the App.
	LinearCallbackPath = "/api/linear/oauth/callback"
	// LinearWebhookPathPrefix prefixes the per-App public webhook route.
	LinearWebhookPathPrefix = "/api/linear/webhook/"

	linearMaxDisplayNameRunes = 100
	linearMaxClientIDLength   = 200
	linearMaxAllowedUsers     = 200
	linearMaxRunSeconds       = 24 * 60 * 60
)

// LinearAppServiceServer implements agentsv1connect.LinearAppServiceHandler
// (ADR-0015). A Linear App is one Linear OAuth app routed to one Agent: the
// app user is that Agent's identity in Linear. Its secrets are write-only,
// and its callback and webhook URLs are derived, never stored.
type LinearAppServiceServer struct {
	repo          linearrepo.Repository
	workspaceRepo workspacerepo.Repository
	agentRepo     configrepo.AgentRepository
	keyring       *secretbox.Keyring
	settings      linearsetting.Repository
	// states holds single-use install flow states; linear calls Linear.
	states           linearstate.Repository
	linear           *linearapi.Client
	clock            func() time.Time
	dashboardBaseURL func() string
}

func NewLinearAppServiceServer(repo linearrepo.Repository) *LinearAppServiceServer {
	return &LinearAppServiceServer{repo: repo}
}

func (s *LinearAppServiceServer) SetRepo(repo linearrepo.Repository) { s.repo = repo }

func (s *LinearAppServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

// SetAgentRepo wires the Agent repository the routed-Agent check reads.
func (s *LinearAppServiceServer) SetAgentRepo(repo configrepo.AgentRepository) { s.agentRepo = repo }

func (s *LinearAppServiceServer) SetKeyring(keyring *secretbox.Keyring) { s.keyring = keyring }

// SetSettingsRepo wires the platform Linear settings the URLs derive from.
func (s *LinearAppServiceServer) SetSettingsRepo(repo linearsetting.Repository) { s.settings = repo }

func (s *LinearAppServiceServer) requireReady() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("linear repository not configured"))
	}
	return nil
}

func (s *LinearAppServiceServer) requireKeyring() error {
	if s.keyring == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("credential encryption is not configured"))
	}
	return nil
}

func mapLinearRepoErr(err error) *connect.Error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, linearrepo.ErrNotFound):
		return connectx.NotFound(err.Error())
	case errors.Is(err, linearrepo.ErrClientIDExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, linearrepo.ErrRevisionConflict):
		return connect.NewError(connect.CodeAborted, err)
	default:
		return connectx.InternalWith(err)
	}
}

// --- Derived URLs ----------------------------------------------------------

// linearBaseURL returns the configured public base URL, or "" when unset.
func (s *LinearAppServiceServer) linearBaseURL(ctx context.Context) string {
	if s.settings == nil {
		return ""
	}
	settings, err := s.settings.Get(ctx)
	if err != nil {
		log.FromContext(ctx).Warn("could not read linear settings", "err", err)
		return ""
	}
	return strings.TrimRight(settings.GetPublicBaseUrl(), "/")
}

// LinearCallbackURL derives the OAuth callback URL from the base URL.
func LinearCallbackURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	return baseURL + LinearCallbackPath
}

// LinearWebhookURL derives an App's webhook URL from the base URL.
func LinearWebhookURL(baseURL, appID string) string {
	if baseURL == "" || appID == "" {
		return ""
	}
	return baseURL + LinearWebhookPathPrefix + appID
}

func stampLinearURLs(app *agentsv1.LinearApp, baseURL string) *agentsv1.LinearApp {
	if app == nil {
		return nil
	}
	app.CallbackUrl = LinearCallbackURL(baseURL)
	app.WebhookUrl = LinearWebhookURL(baseURL, app.GetId())
	return app
}

// --- Validation ------------------------------------------------------------

// normalizeLinearApp validates the operator-supplied fields and returns the
// canonical form stored.
func normalizeLinearApp(app *agentsv1.LinearApp) (*agentsv1.LinearApp, error) {
	out := &agentsv1.LinearApp{
		Id:             app.GetId(),
		DisplayName:    strings.TrimSpace(app.GetDisplayName()),
		ClientId:       strings.TrimSpace(app.GetClientId()),
		AgentId:        strings.TrimSpace(app.GetAgentId()),
		InboundEnabled: app.GetInboundEnabled(),
		Revision:       app.GetRevision(),
	}
	if out.DisplayName == "" {
		return nil, connectx.RequiredArgument("app.display_name")
	}
	if utf8.RuneCountInString(out.DisplayName) > linearMaxDisplayNameRunes {
		return nil, connectx.InvalidArgument("app.display_name", fmt.Sprintf("must be at most %d characters", linearMaxDisplayNameRunes))
	}
	if out.ClientId == "" {
		return nil, connectx.RequiredArgument("app.client_id")
	}
	if len(out.ClientId) > linearMaxClientIDLength || strings.ContainsAny(out.ClientId, " \t\r\n") {
		return nil, connectx.InvalidArgument("app.client_id", "must be the Linear OAuth app's client ID")
	}
	if out.AgentId == "" {
		return nil, connectx.RequiredArgument("app.agent_id")
	}
	users, err := normalizeLinearUserIDs(app.GetAllowedUserIds())
	if err != nil {
		return nil, err
	}
	out.AllowedUserIds = users
	if app.MaxRunSeconds != nil {
		seconds := app.GetMaxRunSeconds()
		if seconds < 0 || seconds > linearMaxRunSeconds {
			return nil, connectx.InvalidArgument("app.max_run_seconds",
				fmt.Sprintf("must be between 0 (unlimited) and %d", linearMaxRunSeconds))
		}
		out.MaxRunSeconds = proto.Int32(seconds)
	}
	return out, nil
}

// normalizeLinearUserIDs canonicalizes and de-duplicates the allowlist,
// preserving order. Linear user IDs are UUIDs; anything else is a typo that
// would silently never match.
func normalizeLinearUserIDs(raw []string) ([]string, error) {
	out := make([]string, 0, len(raw))
	for _, entry := range raw {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		id, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, connectx.InvalidArgument("app.allowed_user_ids",
				fmt.Sprintf("%q is not a Linear user ID", trimmed))
		}
		if canonical := id.String(); !slices.Contains(out, canonical) {
			out = append(out, canonical)
		}
	}
	if len(out) > linearMaxAllowedUsers {
		return nil, connectx.InvalidArgument("app.allowed_user_ids",
			fmt.Sprintf("must list at most %d users", linearMaxAllowedUsers))
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

// encryptLinearSecret turns a write-only secret into a storable credential.
// An empty value yields an unset credential (used to clear).
func (s *LinearAppServiceServer) encryptLinearSecret(ctx context.Context, field, value string) (linearrepo.Credential, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return linearrepo.Credential{}, nil
	}
	if err := s.requireKeyring(); err != nil {
		return linearrepo.Credential{}, err
	}
	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(value))
	if err != nil {
		return linearrepo.Credential{}, connectx.InternalWith(fmt.Errorf("encrypt linear %s: %w", field, err))
	}
	return linearrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}, nil
}

// requireInboundCredentials refuses inbound without both secrets: a webhook
// cannot be verified without the signing secret, and an installation cannot
// be completed or refreshed without the client secret.
func requireInboundCredentials(clientSecretSet, webhookSecretSet bool) error {
	if clientSecretSet && webhookSecretSet {
		return nil
	}
	return connect.NewError(connect.CodeFailedPrecondition,
		errors.New("enabling inbound requires both the client secret and the webhook signing secret"))
}

// --- Reads -----------------------------------------------------------------

func (s *LinearAppServiceServer) ListLinearApps(ctx context.Context, _ *connect.Request[agentsv1.ListLinearAppsRequest]) (*connect.Response[agentsv1.ListLinearAppsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	apps, err := s.repo.ListApps(ctx, workspaceID)
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	baseURL := s.linearBaseURL(ctx)
	for _, app := range apps {
		stampLinearURLs(app, baseURL)
	}
	return connect.NewResponse(&agentsv1.ListLinearAppsResponse{Apps: apps}), nil
}

func (s *LinearAppServiceServer) GetLinearApp(ctx context.Context, req *connect.Request[agentsv1.GetLinearAppRequest]) (*connect.Response[agentsv1.GetLinearAppResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	app, err := s.repo.GetApp(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.GetLinearAppResponse{App: stampLinearURLs(app, s.linearBaseURL(ctx))}), nil
}

// --- Writes ----------------------------------------------------------------

func (s *LinearAppServiceServer) CreateLinearApp(ctx context.Context, req *connect.Request[agentsv1.CreateLinearAppRequest]) (*connect.Response[agentsv1.CreateLinearAppResponse], error) {
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
	app, err := normalizeLinearApp(req.Msg.GetApp())
	if err != nil {
		return nil, err
	}
	if _, err := resolveActiveAgent(ctx, s.agentRepo, workspaceID, "app.agent_id", app.GetAgentId()); err != nil {
		return nil, err
	}
	clientSecret, err := s.encryptLinearSecret(ctx, "client secret", req.Msg.GetClientSecret())
	if err != nil {
		return nil, err
	}
	webhookSecret, err := s.encryptLinearSecret(ctx, "webhook secret", req.Msg.GetWebhookSecret())
	if err != nil {
		return nil, err
	}
	if app.GetInboundEnabled() {
		if err := requireInboundCredentials(clientSecret.Set(), webhookSecret.Set()); err != nil {
			return nil, err
		}
	}
	app.Id = uuid.NewString()
	created, err := s.repo.CreateApp(ctx, workspaceID, app, linearrepo.AppCredentials{
		ClientSecret: clientSecret, WebhookSecret: webhookSecret,
	})
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	log.FromContext(ctx).Info("linear app created",
		"workspace_id", workspaceID, "app_id", created.GetId(), "agent_id", created.GetAgentId())
	return connect.NewResponse(&agentsv1.CreateLinearAppResponse{App: stampLinearURLs(created, s.linearBaseURL(ctx))}), nil
}

func (s *LinearAppServiceServer) UpdateLinearApp(ctx context.Context, req *connect.Request[agentsv1.UpdateLinearAppRequest]) (*connect.Response[agentsv1.UpdateLinearAppResponse], error) {
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
	if strings.TrimSpace(req.Msg.GetApp().GetId()) == "" {
		return nil, connectx.RequiredArgument("app.id")
	}
	prev, err := s.repo.GetApp(ctx, workspaceID, req.Msg.GetApp().GetId())
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	// The client ID is immutable; validating the stored one keeps the
	// caller from having to resend it.
	input := proto.Clone(req.Msg.GetApp()).(*agentsv1.LinearApp)
	input.ClientId = prev.GetClientId()
	app, err := normalizeLinearApp(input)
	if err != nil {
		return nil, err
	}
	if _, err := resolveActiveAgent(ctx, s.agentRepo, workspaceID, "app.agent_id", app.GetAgentId()); err != nil {
		return nil, err
	}
	if app.GetInboundEnabled() {
		if err := requireInboundCredentials(prev.GetClientSecretSet(), prev.GetWebhookSecretSet()); err != nil {
			return nil, err
		}
	}
	updated, err := s.repo.UpdateApp(ctx, workspaceID, app, req.Msg.GetApp().GetRevision())
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.UpdateLinearAppResponse{App: stampLinearURLs(updated, s.linearBaseURL(ctx))}), nil
}

func (s *LinearAppServiceServer) PutLinearAppCredentials(ctx context.Context, req *connect.Request[agentsv1.PutLinearAppCredentialsRequest]) (*connect.Response[agentsv1.PutLinearAppCredentialsResponse], error) {
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
	prev, err := s.repo.GetApp(ctx, workspaceID, req.Msg.GetAppId())
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}

	var change linearrepo.CredentialChange
	clears := false
	if req.Msg.ClientSecret != nil {
		cred, err := s.encryptLinearSecret(ctx, "client secret", req.Msg.GetClientSecret())
		if err != nil {
			return nil, err
		}
		change.ClientSecret = &cred
		clears = clears || !cred.Set()
	}
	if req.Msg.WebhookSecret != nil {
		cred, err := s.encryptLinearSecret(ctx, "webhook secret", req.Msg.GetWebhookSecret())
		if err != nil {
			return nil, err
		}
		change.WebhookSecret = &cred
		clears = clears || !cred.Set()
	}
	if clears && prev.GetInboundEnabled() {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("disable inbound before clearing a Linear App secret"))
	}
	updated, err := s.repo.SetAppCredentials(ctx, workspaceID, prev.GetId(), change)
	if err != nil {
		return nil, mapLinearRepoErr(err)
	}
	log.FromContext(ctx).Info("linear app credentials updated",
		"workspace_id", workspaceID, "app_id", prev.GetId(),
		"client_secret_set", updated.GetClientSecretSet(), "webhook_secret_set", updated.GetWebhookSecretSet())
	return connect.NewResponse(&agentsv1.PutLinearAppCredentialsResponse{App: stampLinearURLs(updated, s.linearBaseURL(ctx))}), nil
}

func (s *LinearAppServiceServer) DeleteLinearApp(ctx context.Context, req *connect.Request[agentsv1.DeleteLinearAppRequest]) (*connect.Response[agentsv1.DeleteLinearAppResponse], error) {
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
	// The App's installations go with it; revoke their tokens first so
	// nothing Butter no longer tracks keeps working at Linear.
	if installs, err := s.repo.ListInstallations(ctx, workspaceID, req.Msg.GetId()); err == nil {
		for _, inst := range installs {
			s.revokeInstallationBestEffort(ctx, workspaceID, inst)
		}
	}
	if err := s.repo.DeleteApp(ctx, workspaceID, req.Msg.GetId()); err != nil {
		return nil, mapLinearRepoErr(err)
	}
	log.FromContext(ctx).Info("linear app deleted", "workspace_id", workspaceID, "app_id", req.Msg.GetId())
	return connect.NewResponse(&agentsv1.DeleteLinearAppResponse{}), nil
}

// --- Reference guard ---------------------------------------------------------

// LinearReferenceGuard blocks removing an Agent a Linear App routes to. It is
// a strong reference for the same reason as Telegram's: an App whose Agent
// vanished is an issue delegation that silently never answers.
type LinearReferenceGuard struct {
	repo linearrepo.Repository
}

func NewLinearReferenceGuard(repo linearrepo.Repository) *LinearReferenceGuard {
	return &LinearReferenceGuard{repo: repo}
}

// CheckAgentRemovable returns a FailedPrecondition error naming every Linear
// App that routes to the Agent.
func (g *LinearReferenceGuard) CheckAgentRemovable(ctx context.Context, workspaceID, agentID string) error {
	if g == nil || g.repo == nil || strings.TrimSpace(agentID) == "" {
		return nil
	}
	apps, err := g.repo.ListApps(ctx, workspaceID)
	if err != nil {
		return connectx.InternalWith(err)
	}
	var refs []string
	for _, app := range apps {
		if app.GetAgentId() == agentID {
			refs = append(refs, app.GetId())
		}
	}
	if len(refs) == 0 {
		return nil
	}
	slices.Sort(refs)
	return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
		"agent %q is routed to by linear apps: %s", agentID, strings.Join(refs, ", ")))
}

// --- Installations -----------------------------------------------------------

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
