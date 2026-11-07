package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"butterfly.orx.me/core/log"
	planeapi "go.orx.me/apps/butter/internal/planeapi"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	planeinstallation "go.orx.me/apps/butter/internal/repo/planeinstallation"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	planeprocessingmemory "go.orx.me/apps/butter/internal/repo/planeprocessing/memory"
	planestate "go.orx.me/apps/butter/internal/repo/planestate"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	planeMaxDisplayNameRunes = 100
	planeMaxClientIDLength   = 200
	planeMaxAllowedUsers     = 200
	planeMaxRunSeconds       = 24 * 60 * 60
)

// PlaneAppServiceServer implements agentsv1connect.PlaneAppServiceHandler
// (spec #487, ticket #497).
//
// #497 stops at the App resource: registration, workspace isolation, the
// credential seam, revision fencing and reference protection. Installation
// (#498) and inbound delivery are not implemented here, so:
//
//   - inbound is always stored disabled and cannot be turned on;
//   - no callback or webhook route exists, so nothing can auto-create an App
//     on a redirect or refresh — an App exists only when a manager submits
//     CreatePlaneApp;
//   - the provider profile is server-assigned and immutable, and the fixed
//     origins live in deployment configuration, not workspace input.
type PlaneAppServiceServer struct {
	repo             planerepo.Repository
	workspaceRepo    workspacerepo.Repository
	agentRepo        configrepo.AgentRepository
	keyring          *secretbox.Keyring
	profile          func() (planeapi.Profile, error)
	installStates    planestate.Repository
	clock            func() time.Time
	dashboardBaseURL func() string
	installRepo      planeinstallation.Repository
	processingRepo   planeprocessing.Repository
	delivery         *plane.DeliveryManagement
	planeHTTPClient  *planeapi.Client
}

func NewPlaneAppServiceServer(repo planerepo.Repository) *PlaneAppServiceServer {
	return &PlaneAppServiceServer{repo: repo, processingRepo: planeprocessingmemory.New()}
}

func (s *PlaneAppServiceServer) SetRepo(repo planerepo.Repository) { s.repo = repo }

func (s *PlaneAppServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

// SetAgentRepo wires the Agent repository the routed-Agent check reads.
func (s *PlaneAppServiceServer) SetAgentRepo(repo configrepo.AgentRepository) { s.agentRepo = repo }

func (s *PlaneAppServiceServer) SetKeyring(keyring *secretbox.Keyring) { s.keyring = keyring }

func (s *PlaneAppServiceServer) SetPlaneProcessingRepo(repo planeprocessing.Repository) {
	s.processingRepo = repo
}

func (s *PlaneAppServiceServer) PlaneProcessingRepository() planeprocessing.Repository {
	return s.processingRepo
}

func (s *PlaneAppServiceServer) SetPlaneDeliveryManagement(m *plane.DeliveryManagement) {
	s.delivery = m
}

// SetProfileProvider wires the fixed provider profile built from startup
// configuration. It is a provider rather than a value because SetupRoutes
// runs before config is loaded.
func (s *PlaneAppServiceServer) SetProfileProvider(provider func() (planeapi.Profile, error)) {
	s.profile = provider
}

func (s *PlaneAppServiceServer) requireReady() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("plane repository not configured"))
	}
	if s.profile == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("plane provider profile is not configured"))
	}
	return nil
}

func (s *PlaneAppServiceServer) requireProfile() (planeapi.Profile, error) {
	if s.profile == nil {
		return planeapi.Profile{}, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane provider profile is not configured"))
	}
	profile, err := s.profile()
	if err != nil {
		return planeapi.Profile{}, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("the plane provider profile is misconfigured: %w", err))
	}
	return profile, nil
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

// --- Validation ------------------------------------------------------------

// normalizePlaneApp validates the operator-supplied fields and returns the
// canonical form stored. Server-owned fields — identity, provider profile,
// revision, timestamps, inbound and credential state — are rebuilt here so a
// caller can never smuggle them through.
func normalizePlaneApp(app *agentsv1.PlaneApp) (*agentsv1.PlaneApp, error) {
	out := &agentsv1.PlaneApp{
		Id:          strings.TrimSpace(app.GetId()),
		DisplayName: strings.TrimSpace(app.GetDisplayName()),
		ClientId:    strings.TrimSpace(app.GetClientId()),
		AgentId:     strings.TrimSpace(app.GetAgentId()),
	}
	if out.DisplayName == "" {
		return nil, connectx.RequiredArgument("app.display_name")
	}
	if utf8.RuneCountInString(out.DisplayName) > planeMaxDisplayNameRunes {
		return nil, connectx.InvalidArgument("app.display_name",
			fmt.Sprintf("must be at most %d characters", planeMaxDisplayNameRunes))
	}
	if out.ClientId == "" {
		return nil, connectx.RequiredArgument("app.client_id")
	}
	if len(out.ClientId) > planeMaxClientIDLength || strings.ContainsAny(out.ClientId, " \t\r\n") {
		return nil, connectx.InvalidArgument("app.client_id", "must be the Plane app's OAuth client ID")
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
		canonical := id.String()
		duplicate := false
		for _, existing := range out {
			if existing == canonical {
				duplicate = true
				break
			}
		}
		if !duplicate {
			out = append(out, canonical)
		}
	}
	if len(out) > planeMaxAllowedUsers {
		return nil, connectx.InvalidArgument("app.allowed_user_ids",
			fmt.Sprintf("must list at most %d users", planeMaxAllowedUsers))
	}
	return out, nil
}

// rejectInboundEnablement is the #497 half of the default-off requirement:
// the resource records the desired state for a later ticket, but until the
// execution path exists accepting inbound would silently drop Plane
// deliveries (or worse, run them through a half-built path). It fails loudly
// instead of storing a flag nothing honours.
func rejectInboundEnablement(app *agentsv1.PlaneApp) error {
	if !app.GetInboundEnabled() {
		return nil
	}
	return connect.NewError(connect.CodeFailedPrecondition, errors.New("plane inbound requires an installed, authorized App and verified provider profile"))
}

func requireInboundEnablement(clientSecretSet, webhookSecretSet, enabled bool, profile planeapi.Profile) error {
	if !enabled {
		return nil
	}
	if !clientSecretSet || !webhookSecretSet || !profile.Verified {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("plane inbound requires an installed, authorized App and verified provider profile"))
	}
	return nil
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
	return connect.NewResponse(&agentsv1.GetPlaneAppResponse{App: app}), nil
}

func (s *PlaneAppServiceServer) GetPlaneProviderProfile(ctx context.Context, _ *connect.Request[agentsv1.GetPlaneProviderProfileRequest]) (*connect.Response[agentsv1.GetPlaneProviderProfileResponse], error) {
	if _, err := requireWorkspace(ctx); err != nil {
		return nil, err
	}
	profile, err := s.requireProfile()
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.GetPlaneProviderProfileResponse{Profile: planeProfileProto(profile)}), nil
}

// planeProfileProto redacts the profile to what a workspace may see.
func planeProfileProto(profile planeapi.Profile) *agentsv1.PlaneProviderProfile {
	return &agentsv1.PlaneProviderProfile{
		Name:         profile.Name,
		ApiBaseUrl:   profile.APIBaseURL,
		OauthBaseUrl: profile.OAuthURL,
		WebBaseUrl:   profile.WebURL,
		Edition:      string(profile.Edition),
		Verified:     profile.Verified,
	}
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
	profile, err := s.requireProfile()
	if err != nil {
		return nil, err
	}
	app, err := normalizePlaneApp(req.Msg.GetApp())
	if err != nil {
		return nil, err
	}
	if err := rejectInboundEnablement(req.Msg.GetApp()); err != nil {
		return nil, err
	}
	// The provider profile is not caller input: a workspace registers
	// against the deployment's fixed profile or not at all.
	if requested := strings.TrimSpace(req.Msg.GetApp().GetProviderProfile()); requested != "" && requested != profile.Name {
		return nil, connectx.InvalidArgument("app.provider_profile",
			fmt.Sprintf("must be the fixed profile %q; a deployment registers Apps against one Plane provider", profile.Name))
	}
	app.ProviderProfile = profile.Name
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
	app.Id = uuid.NewString()
	created, err := s.repo.CreateApp(ctx, workspaceID, app, planerepo.AppCredentials{
		ClientSecret: clientSecret, WebhookSecret: webhookSecret,
	})
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	log.FromContext(ctx).Info("plane app created",
		"workspace_id", workspaceID, "app_id", created.GetId(), "agent_id", created.GetAgentId(),
		"provider_profile", created.GetProviderProfile(),
		"client_secret_set", created.GetClientSecretSet(), "webhook_secret_set", created.GetWebhookSecretSet())
	return connect.NewResponse(&agentsv1.CreatePlaneAppResponse{App: created}), nil
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
	input := req.Msg.GetApp()
	if strings.TrimSpace(input.GetId()) == "" {
		return nil, connectx.RequiredArgument("app.id")
	}
	if input.GetRevision() <= 0 {
		return nil, connectx.RequiredArgument("app.revision")
	}
	prev, err := s.repo.GetApp(ctx, workspaceID, input.GetId())
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	profile, err := s.requireProfile()
	if err != nil {
		return nil, err
	}
	if input.GetInboundEnabled() {
		if err := requireInboundEnablement(prev.GetClientSecretSet(), prev.GetWebhookSecretSet(), true, profile); err != nil {
			return nil, err
		}
	}

	// Identity and provider profile are immutable. Validating the stored
	// values keeps the caller from having to resend them, and a differing
	// profile is a different App, not an edit.
	if requested := strings.TrimSpace(input.GetProviderProfile()); requested != "" && requested != prev.GetProviderProfile() {
		return nil, connectx.InvalidArgument("app.provider_profile",
			"is immutable; register a new Plane App to use another provider profile")
	}
	if requested := strings.TrimSpace(input.GetClientId()); requested != "" && requested != prev.GetClientId() {
		return nil, connectx.InvalidArgument("app.client_id",
			"is immutable; register a new Plane App to use another client ID")
	}
	// The client ID is immutable and already validated at registration, so
	// injecting the stored one keeps the caller from having to resend it.
	input.ClientId = prev.GetClientId()
	app, err := normalizePlaneApp(input)
	if err != nil {
		return nil, err
	}
	app.ProviderProfile = prev.GetProviderProfile()
	if _, err := resolveActiveAgent(ctx, s.agentRepo, workspaceID, "app.agent_id", app.GetAgentId()); err != nil {
		return nil, err
	}
	updated, err := s.repo.UpdateApp(ctx, workspaceID, app, input.GetRevision())
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.UpdatePlaneAppResponse{App: updated}), nil
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
	if req.Msg.ClientSecret != nil {
		cred, err := s.encryptPlaneSecret(ctx, "client secret", req.Msg.GetClientSecret())
		if err != nil {
			return nil, err
		}
		change.ClientSecret = &cred
	}
	if req.Msg.WebhookSecret != nil {
		cred, err := s.encryptPlaneSecret(ctx, "webhook secret", req.Msg.GetWebhookSecret())
		if err != nil {
			return nil, err
		}
		change.WebhookSecret = &cred
	}
	updated, err := s.repo.SetAppCredentials(ctx, workspaceID, prev.GetId(), change)
	if err != nil {
		return nil, mapPlaneRepoErr(err)
	}
	// Log only whether a secret is now present; never the value or its
	// ciphertext.
	log.FromContext(ctx).Info("plane app credentials updated",
		"workspace_id", workspaceID, "app_id", prev.GetId(),
		"client_secret_set", updated.GetClientSecretSet(),
		"webhook_secret_set", updated.GetWebhookSecretSet())
	return connect.NewResponse(&agentsv1.PutPlaneAppCredentialsResponse{App: updated}), nil
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
