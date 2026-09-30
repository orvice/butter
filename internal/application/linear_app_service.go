package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"butterfly.orx.me/core/log"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/repo/linearsetting"
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
