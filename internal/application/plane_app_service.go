package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"butterfly.orx.me/core/log"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	"go.orx.me/apps/butter/internal/repo/planesetting"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// PlaneCallbackPath is the public OAuth callback shared by every
	// Plane App; the single-use state identifies the App.
	PlaneCallbackPath = "/api/plane/oauth/callback"
	// PlaneWebhookPathPrefix prefixes the per-App public webhook route.
	PlaneWebhookPathPrefix = "/api/plane/webhook/"

	planeMaxDisplayNameRunes = 100
	planeMaxClientIDLength   = 200
	planeMaxAllowedUsers     = 200
	planeMaxRunSeconds       = 24 * 60 * 60
)

// PlaneAppServiceServer implements agentsv1connect.PlaneAppServiceHandler
// (ADR-0015). A Plane App is one Plane OAuth app routed to one Agent: the
// app user is that Agent's identity in Plane. Its secrets are write-only,
// and its callback and webhook URLs are derived, never stored.
type PlaneAppServiceServer struct {
	repo             planerepo.Repository
	workspaceRepo    workspacerepo.Repository
	agentRepo        configrepo.AgentRepository
	keyring          *secretbox.Keyring
	settings         planesetting.Repository
	clock            func() time.Time
	dashboardBaseURL func() string
	// queue reports whether the durable receive queue is usable: inbound
	// depends on it, so enabling without it would accept deliveries the
	// fleet cannot durably hold.
	queue QueueProbe
}

func NewPlaneAppServiceServer(repo planerepo.Repository) *PlaneAppServiceServer {
	return &PlaneAppServiceServer{repo: repo}
}

func (s *PlaneAppServiceServer) SetRepo(repo planerepo.Repository) { s.repo = repo }

func (s *PlaneAppServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

// SetAgentRepo wires the Agent repository the routed-Agent check reads.
func (s *PlaneAppServiceServer) SetAgentRepo(repo configrepo.AgentRepository) { s.agentRepo = repo }

func (s *PlaneAppServiceServer) SetKeyring(keyring *secretbox.Keyring) { s.keyring = keyring }

// SetQueueProbe wires the durable-queue readiness check.
func (s *PlaneAppServiceServer) SetQueueProbe(probe QueueProbe) { s.queue = probe }

// requireDurableQueue refuses inbound without durable Redis Streams.
func (s *PlaneAppServiceServer) requireDurableQueue(ctx context.Context) error {
	if s.queue == nil || !s.queue.Available() {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("enabling inbound requires Redis, configured as a durable queue"))
	}
	if err := s.queue.CheckReady(ctx); err != nil {
		return connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("the Plane receive queue is not durable: %w", err))
	}
	return nil
}

// SetSettingsRepo wires the platform Plane settings the URLs derive from.
func (s *PlaneAppServiceServer) SetSettingsRepo(repo planesetting.Repository) { s.settings = repo }

func (s *PlaneAppServiceServer) requireReady() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("plane repository not configured"))
	}
	return nil
}

func (s *PlaneAppServiceServer) requireKeyring() error {
	if s.keyring == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("credential encryption is not configured"))
	}
	return nil
}

func mapPlaneRepoErr(err error) *connect.Error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, planerepo.ErrNotFound):
		return connectx.NotFound(err.Error())
	case errors.Is(err, planerepo.ErrClientIDExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, planerepo.ErrRevisionConflict):
		return connect.NewError(connect.CodeAborted, err)
	default:
		return connectx.InternalWith(err)
	}
}

// --- Derived URLs ----------------------------------------------------------

// planeBaseURL returns the configured public base URL, or "" when unset.
func (s *PlaneAppServiceServer) planeBaseURL(ctx context.Context) string {
	if s.settings == nil {
		return ""
	}
	settings, err := s.settings.Get(ctx)
	if err != nil {
		log.FromContext(ctx).Warn("could not read plane settings", "err", err)
		return ""
	}
	return strings.TrimRight(settings.GetPublicBaseUrl(), "/")
}

// PlaneCallbackURL derives the OAuth callback URL from the base URL.
func PlaneCallbackURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	return baseURL + PlaneCallbackPath
}

// PlaneWebhookURL derives an App's webhook URL from the base URL.
func PlaneWebhookURL(baseURL, appID string) string {
	if baseURL == "" || appID == "" {
		return ""
	}
	return baseURL + PlaneWebhookPathPrefix + appID
}

func stampPlaneURLs(app *agentsv1.PlaneApp, baseURL string) *agentsv1.PlaneApp {
	if app == nil {
		return nil
	}
	app.CallbackUrl = PlaneCallbackURL(baseURL)
	app.WebhookUrl = PlaneWebhookURL(baseURL, app.GetId())
	return app
}

// --- Validation ------------------------------------------------------------

// normalizePlaneApp validates the operator-supplied fields and returns the
// canonical form stored.
func normalizePlaneApp(app *agentsv1.PlaneApp) (*agentsv1.PlaneApp, error) {
	out := &agentsv1.PlaneApp{
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
	if utf8.RuneCountInString(out.DisplayName) > planeMaxDisplayNameRunes {
		return nil, connectx.InvalidArgument("app.display_name", fmt.Sprintf("must be at most %d characters", planeMaxDisplayNameRunes))
	}
	if out.ClientId == "" {
		return nil, connectx.RequiredArgument("app.client_id")
	}
	if len(out.ClientId) > planeMaxClientIDLength || strings.ContainsAny(out.ClientId, " \t\r\n") {
		return nil, connectx.InvalidArgument("app.client_id", "must be the Plane OAuth app's client ID")
	}
	if out.AgentId == "" {
		return nil, connectx.RequiredArgument("app.agent_id")
	}
	users, err := normalizePlaneUserIDs(app.GetAllowedUserIds())
	if err != nil {
		return nil, err
	}
	out.AllowedUserIds = users
	if app.MaxRunSeconds != nil {
		seconds := app.GetMaxRunSeconds()
		if seconds < 0 || seconds > planeMaxRunSeconds {
			return nil, connectx.InvalidArgument("app.max_run_seconds",
				fmt.Sprintf("must be between 0 (unlimited) and %d", planeMaxRunSeconds))
		}
		out.MaxRunSeconds = proto.Int32(seconds)
	}
	return out, nil
}

// normalizePlaneUserIDs canonicalizes and de-duplicates the allowlist,
// preserving order. Plane user IDs are UUIDs; anything else is a typo that
// would silently never match.
func normalizePlaneUserIDs(raw []string) ([]string, error) {
	out := make([]string, 0, len(raw))
	for _, entry := range raw {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		id, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, connectx.InvalidArgument("app.allowed_user_ids",
				fmt.Sprintf("%q is not a Plane user ID", trimmed))
		}
		if canonical := id.String(); !slices.Contains(out, canonical) {
			out = append(out, canonical)
		}
	}
	if len(out) > planeMaxAllowedUsers {
		return nil, connectx.InvalidArgument("app.allowed_user_ids",
			fmt.Sprintf("must list at most %d users", planeMaxAllowedUsers))
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

// encryptPlaneSecret turns a write-only secret into a storable credential.
// An empty value yields an unset credential (used to clear).
func (s *PlaneAppServiceServer) encryptPlaneSecret(ctx context.Context, field, value string) (planerepo.Credential, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return planerepo.Credential{}, nil
	}
	if err := s.requireKeyring(); err != nil {
		return planerepo.Credential{}, err
	}
	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(value))
	if err != nil {
		return planerepo.Credential{}, connectx.InternalWith(fmt.Errorf("encrypt plane %s: %w", field, err))
	}
	return planerepo.Credential{Ciphertext: ciphertext, KeyID: keyID}, nil
}

// requireInboundCredentials refuses inbound without both secrets: a webhook
// cannot be verified without the signing secret, and an installation cannot
// be completed or refreshed without the client secret.
func requirePlaneInboundCredentials(clientSecretSet, webhookSecretSet bool) error {
	if clientSecretSet && webhookSecretSet {
		return nil
	}
	return connect.NewError(connect.CodeFailedPrecondition,
		errors.New("enabling inbound requires both the client secret and the webhook signing secret"))
}

// --- Reads -----------------------------------------------------------------

func (s *PlaneAppServiceServer) ListPlaneApps(ctx context.Context, _ *connect.Request[agentsv1.ListPlaneAppsRequest]) (*connect.Response[agentsv1.ListPlaneAppsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	apps, err := s.repo.ListApps(ctx, workspaceID)
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	baseURL := s.planeBaseURL(ctx)
	for _, app := range apps {
		stampPlaneURLs(app, baseURL)
	}
	return connect.NewResponse(&agentsv1.ListPlaneAppsResponse{Apps: apps}), nil
}

func (s *PlaneAppServiceServer) GetPlaneApp(ctx context.Context, req *connect.Request[agentsv1.GetPlaneAppRequest]) (*connect.Response[agentsv1.GetPlaneAppResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	app, err := s.repo.GetApp(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.GetPlaneAppResponse{App: stampPlaneURLs(app, s.planeBaseURL(ctx))}), nil
}

// --- Writes ----------------------------------------------------------------

func (s *PlaneAppServiceServer) CreatePlaneApp(ctx context.Context, req *connect.Request[agentsv1.CreatePlaneAppRequest]) (*connect.Response[agentsv1.CreatePlaneAppResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "plane"); err != nil {
		return nil, err
	}
	app, err := normalizePlaneApp(req.Msg.GetApp())
	if err != nil {
		return nil, err
	}
	if _, err := resolveActiveAgent(ctx, s.agentRepo, workspaceID, "app.agent_id", app.GetAgentId()); err != nil {
		return nil, err
	}
	clientSecret, err := s.encryptPlaneSecret(ctx, "client secret", req.Msg.GetClientSecret())
	if err != nil {
		return nil, err
	}
	webhookSecret, err := s.encryptPlaneSecret(ctx, "webhook secret", req.Msg.GetWebhookSecret())
	if err != nil {
		return nil, err
	}
	if app.GetInboundEnabled() {
		if err := requireInboundCredentials(clientSecret.Set(), webhookSecret.Set()); err != nil {
			return nil, err
		}
		if err := s.requireDurableQueue(ctx); err != nil {
			return nil, err
		}
	}
	app.Id = uuid.NewString()
	created, err := s.repo.CreateApp(ctx, workspaceID, app, planerepo.AppCredentials{
		ClientSecret: clientSecret, WebhookSecret: webhookSecret,
	})
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	log.FromContext(ctx).Info("plane app created",
		"workspace_id", workspaceID, "app_id", created.GetId(), "agent_id", created.GetAgentId())
	return connect.NewResponse(&agentsv1.CreatePlaneAppResponse{App: stampPlaneURLs(created, s.planeBaseURL(ctx))}), nil
}

func (s *PlaneAppServiceServer) UpdatePlaneApp(ctx context.Context, req *connect.Request[agentsv1.UpdatePlaneAppRequest]) (*connect.Response[agentsv1.UpdatePlaneAppResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "plane"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Msg.GetApp().GetId()) == "" {
		return nil, connectx.RequiredArgument("app.id")
	}
	prev, err := s.repo.GetApp(ctx, workspaceID, req.Msg.GetApp().GetId())
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	// The client ID is immutable; validating the stored one keeps the
	// caller from having to resend it.
	input := proto.Clone(req.Msg.GetApp()).(*agentsv1.PlaneApp)
	input.ClientId = prev.GetClientId()
	app, err := normalizePlaneApp(input)
	if err != nil {
		return nil, err
	}
	if _, err := resolveActiveAgent(ctx, s.agentRepo, workspaceID, "app.agent_id", app.GetAgentId()); err != nil {
		return nil, err
	}
	if app.GetInboundEnabled() {
		if err := requirePlaneInboundCredentials(prev.GetClientSecretSet(), prev.GetWebhookSecretSet()); err != nil {
			return nil, err
		}
		if err := s.requireDurableQueue(ctx); err != nil {
			return nil, err
		}
	}
	updated, err := s.repo.UpdateApp(ctx, workspaceID, app, req.Msg.GetApp().GetRevision())
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.UpdatePlaneAppResponse{App: stampPlaneURLs(updated, s.planeBaseURL(ctx))}), nil
}

func (s *PlaneAppServiceServer) PutPlaneAppCredentials(ctx context.Context, req *connect.Request[agentsv1.PutPlaneAppCredentialsRequest]) (*connect.Response[agentsv1.PutPlaneAppCredentialsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "plane"); err != nil {
		return nil, err
	}
	prev, err := s.repo.GetApp(ctx, workspaceID, req.Msg.GetAppId())
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}

	var change planerepo.CredentialChange
	clears := false
	if req.Msg.ClientSecret != nil {
		cred, err := s.encryptPlaneSecret(ctx, "client secret", req.Msg.GetClientSecret())
		if err != nil {
			return nil, err
		}
		change.ClientSecret = &cred
		clears = clears || !cred.Set()
	}
	if req.Msg.WebhookSecret != nil {
		cred, err := s.encryptPlaneSecret(ctx, "webhook secret", req.Msg.GetWebhookSecret())
		if err != nil {
			return nil, err
		}
		change.WebhookSecret = &cred
		clears = clears || !cred.Set()
	}
	if clears && prev.GetInboundEnabled() {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("disable inbound before clearing a Plane App secret"))
	}
	updated, err := s.repo.SetAppCredentials(ctx, workspaceID, prev.GetId(), change)
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	log.FromContext(ctx).Info("plane app credentials updated",
		"workspace_id", workspaceID, "app_id", prev.GetId(),
		"client_secret_set", updated.GetClientSecretSet(), "webhook_secret_set", updated.GetWebhookSecretSet())
	return connect.NewResponse(&agentsv1.PutPlaneAppCredentialsResponse{App: stampPlaneURLs(updated, s.planeBaseURL(ctx))}), nil
}

func (s *PlaneAppServiceServer) DeletePlaneApp(ctx context.Context, req *connect.Request[agentsv1.DeletePlaneAppRequest]) (*connect.Response[agentsv1.DeletePlaneAppResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "plane"); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteApp(ctx, workspaceID, req.Msg.GetId()); err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	log.FromContext(ctx).Info("plane app deleted", "workspace_id", workspaceID, "app_id", req.Msg.GetId())
	return connect.NewResponse(&agentsv1.DeletePlaneAppResponse{}), nil
}
