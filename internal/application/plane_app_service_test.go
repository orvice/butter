package application

// Service-level tests for PlaneAppService and PlaneAdminService (spec #487,
// ticket #497): workspace isolation, role enforcement, the fixed provider
// profile, write-only encrypted secrets, global client-ID uniqueness,
// optimistic revisions, strong Agent references, and the default-off inbound
// state.
//
// These call the Connect handlers directly, which is the RPC surface the
// dashboard uses; the UI half of the acceptance evidence lives in
// front/src/features/plane/form-state.test.ts.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/planeapi"
	"go.orx.me/apps/butter/internal/repo/auth"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	planememory "go.orx.me/apps/butter/internal/repo/plane/memory"
	planesettingmemory "go.orx.me/apps/butter/internal/repo/planesetting/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	planeClientSecret  = "plane-client-secret-value"
	planeWebhookSecret = "plane-webhook-secret-value"
	planeUserA         = "1d2c3b4a-5f6e-4a7b-8c9d-0e1f2a3b4c5d"
)

type planeFixture struct {
	apps     *PlaneAppServiceServer
	admin    *PlaneAdminServiceServer
	repo     *planememory.Store
	settings *planesettingmemory.Store
	config   *configmemory.Store
	profile  planeapi.Profile
}

func newPlaneFixture(t *testing.T) *planeFixture {
	t.Helper()
	wsRepo := workspacememory.New()
	for _, ws := range []string{"ws-a", "ws-b"} {
		if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: ws, Name: ws, Slug: ws}); err != nil {
			t.Fatalf("seed workspace %s: %v", ws, err)
		}
	}
	for _, m := range []struct{ user, role, ws string }{
		{"owner", "owner", "ws-a"},
		{"admin-a", "admin", "ws-a"},
		{"member", "member", "ws-a"},
		{"owner-b", "owner", "ws-b"},
	} {
		if _, err := wsRepo.AddMember(t.Context(), &agentsv1.WorkspaceMember{
			WorkspaceId: m.ws, UserId: m.user, Role: m.role,
		}); err != nil {
			t.Fatalf("seed member %s: %v", m.user, err)
		}
	}
	configStore := configmemory.New()
	for _, a := range []struct {
		ws, id string
		status agentsv1.AgentLifecycleStatus
	}{
		{"ws-a", "support", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE},
		{"ws-a", "research", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE},
		{"ws-a", "retired", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED},
		{"ws-b", "support", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE},
	} {
		if _, err := configStore.CreateAgent(t.Context(), a.ws, &agentsv1.Agent{
			Name: a.ws + "-" + a.id, AgentId: a.id, LifecycleStatus: a.status,
		}); err != nil {
			t.Fatalf("seed agent %s/%s: %v", a.ws, a.id, err)
		}
	}
	profile, err := planeapi.Build(planeapi.DefaultCloudConfig())
	if err != nil {
		t.Fatalf("build cloud profile: %v", err)
	}
	provider := func() (planeapi.Profile, error) { return profile, nil }
	repo := planememory.New()
	settings := planesettingmemory.New()
	apps := NewPlaneAppServiceServer(repo)
	apps.SetWorkspaceRepo(wsRepo)
	apps.SetAgentRepo(configStore)
	apps.SetKeyring(secretbox.NewKeyring(cryptokeymemory.New()))
	apps.SetProfileProvider(provider)
	admin := NewPlaneAdminServiceServer(settings)
	admin.SetProfileProvider(provider)
	return &planeFixture{
		apps: apps, admin: admin, repo: repo, settings: settings, config: configStore, profile: profile,
	}
}

func planeOwnerA() context.Context  { return ctxAs("owner", "user", "ws-a") }
func planeAdminA() context.Context  { return ctxAs("admin-a", "user", "ws-a") }
func planeMemberA() context.Context { return ctxAs("member", "user", "ws-a") }
func planeOwnerB() context.Context  { return ctxAs("owner-b", "user", "ws-b") }

func (fx *planeFixture) create(t *testing.T, ctx context.Context, app *agentsv1.PlaneApp, secrets ...string) *agentsv1.PlaneApp {
	t.Helper()
	req := &agentsv1.CreatePlaneAppRequest{App: app}
	if len(secrets) > 0 {
		req.ClientSecret = proto.String(secrets[0])
	}
	if len(secrets) > 1 {
		req.WebhookSecret = proto.String(secrets[1])
	}
	resp, err := fx.apps.CreatePlaneApp(ctx, connect.NewRequest(req))
	if err != nil {
		t.Fatalf("CreatePlaneApp: %v", err)
	}
	return resp.Msg.GetApp()
}

func planeSupportApp(clientID string) *agentsv1.PlaneApp {
	return &agentsv1.PlaneApp{DisplayName: "Support", ClientId: clientID, AgentId: "support"}
}

func wantPlaneCode(t *testing.T, err error, code connect.Code, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected %v, got nil", what, code)
	}
	if got := connect.CodeOf(err); got != code {
		t.Fatalf("%s: code = %v, want %v (%v)", what, got, code, err)
	}
}

func TestPlaneAppWorkspaceIsolationAndRoles(t *testing.T) {
	fx := newPlaneFixture(t)

	// Members may read, but nothing in the list leaks across workspaces.
	if resp, err := fx.apps.ListPlaneApps(planeMemberA(), connect.NewRequest(&agentsv1.ListPlaneAppsRequest{})); err != nil {
		t.Fatalf("member ListPlaneApps: %v", err)
	} else if len(resp.Msg.GetApps()) != 0 {
		t.Fatalf("member saw %d apps, want 0", len(resp.Msg.GetApps()))
	}

	// Reading or writing another workspace's App is NotFound, not denied,
	// so the caller cannot learn it exists.
	app := fx.create(t, planeOwnerA(), planeSupportApp("client-a"))
	_, err := fx.apps.GetPlaneApp(planeOwnerB(), connect.NewRequest(&agentsv1.GetPlaneAppRequest{Id: app.GetId()}))
	wantPlaneCode(t, err, connect.CodeNotFound, "cross-workspace GetPlaneApp")
	if apps, err := fx.apps.ListPlaneApps(planeOwnerB(), connect.NewRequest(&agentsv1.ListPlaneAppsRequest{})); err != nil || len(apps.Msg.GetApps()) != 0 {
		t.Fatalf("ListPlaneApps(ws-b) = %d apps, %v; want 0", len(apps.Msg.GetApps()), err)
	}

	// A plain member cannot register or mutate.
	_, err = fx.apps.CreatePlaneApp(planeMemberA(), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: planeSupportApp("client-member"),
	}))
	wantPlaneCode(t, err, connect.CodePermissionDenied, "member CreatePlaneApp")
	_, err = fx.apps.UpdatePlaneApp(planeMemberA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{Id: app.GetId(), DisplayName: "Hijack", Revision: app.GetRevision()},
	}))
	wantPlaneCode(t, err, connect.CodePermissionDenied, "member UpdatePlaneApp")
	_, err = fx.apps.PutPlaneAppCredentials(planeMemberA(), connect.NewRequest(&agentsv1.PutPlaneAppCredentialsRequest{
		AppId:        app.GetId(),
		ClientSecret: proto.String("member-secret"),
	}))
	wantPlaneCode(t, err, connect.CodePermissionDenied, "member PutPlaneAppCredentials")
	_, err = fx.apps.DeletePlaneApp(planeMemberA(), connect.NewRequest(&agentsv1.DeletePlaneAppRequest{Id: app.GetId()}))
	wantPlaneCode(t, err, connect.CodePermissionDenied, "member DeletePlaneApp")

	// A workspace admin (not just owner) may manage.
	resp, err := fx.apps.CreatePlaneApp(planeAdminA(), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: planeSupportApp("client-admin"),
	}))
	if err != nil {
		t.Fatalf("workspace admin CreatePlaneApp: %v", err)
	}
	if resp.Msg.GetApp().GetWorkspaceId() != "ws-a" {
		t.Fatalf("created workspace = %q, want ws-a", resp.Msg.GetApp().GetWorkspaceId())
	}

	// A global admin bypasses the workspace role check inside the workspace
	// it names; the workspace header is still required.
	globalAdminA := auth.WithAdmin(ctxAs("root", "user", "ws-a"))
	if resp, err := fx.apps.CreatePlaneApp(globalAdminA, connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: &agentsv1.PlaneApp{DisplayName: "Global", ClientId: "client-global", AgentId: "support", ProviderProfile: "cloud"},
	})); err != nil {
		t.Fatalf("global admin CreatePlaneApp: %v", err)
	} else if resp.Msg.GetApp().GetWorkspaceId() != "ws-a" {
		t.Fatalf("global admin Created workspace %q, want ws-a", resp.Msg.GetApp().GetWorkspaceId())
	}
	if _, err := fx.apps.CreatePlaneApp(auth.WithAdmin(context.Background()), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: planeSupportApp("client-no-ws"),
	})); err == nil {
		t.Fatal("CreatePlaneApp succeeded without a workspace")
	}
}

func TestPlaneAppRegistersAgainstFixedProviderProfile(t *testing.T) {
	fx := newPlaneFixture(t)

	// The provider profile is server-assigned when the caller omits it.
	app := fx.create(t, planeOwnerA(), planeSupportApp("client-a"))
	if app.GetProviderProfile() != fx.profile.Name {
		t.Fatalf("provider_profile = %q, want %q", app.GetProviderProfile(), fx.profile.Name)
	}

	// Naming the configured profile is accepted.
	if created := fx.create(t, planeOwnerA(), &agentsv1.PlaneApp{
		DisplayName: "Explicit", ClientId: "client-b", AgentId: "support", ProviderProfile: "cloud",
	}); created.GetProviderProfile() != "cloud" {
		t.Fatalf("explicit provider_profile = %q", created.GetProviderProfile())
	}

	// Naming another origin is rejected: a workspace cannot redirect Plane.
	_, err := fx.apps.CreatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: &agentsv1.PlaneApp{DisplayName: "Other", ClientId: "client-c", AgentId: "support", ProviderProfile: "selfhosted"},
	}))
	wantPlaneCode(t, err, connect.CodeInvalidArgument, "create with another provider profile")

	// The provider name is immutable: choosing another provider is a new App.
	_, err = fx.apps.UpdatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{
			Id: app.GetId(), DisplayName: "Moved", AgentId: "support",
			ProviderProfile: "selfhosted", Revision: app.GetRevision(),
		},
	}))
	wantPlaneCode(t, err, connect.CodeInvalidArgument, "update provider profile")

	// A deployment whose profile is invalid fails closed instead of falling
	// back to Cloud origins.
	fx.apps.SetProfileProvider(func() (planeapi.Profile, error) {
		return planeapi.Profile{}, errors.New("plane profile name is required")
	})
	_, err = fx.apps.CreatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: planeSupportApp("client-d"),
	}))
	wantPlaneCode(t, err, connect.CodeFailedPrecondition, "create with a misconfigured profile")
}

func TestPlaneClientIDIsGloballyUniqueAndImmutable(t *testing.T) {
	fx := newPlaneFixture(t)
	fx.create(t, planeOwnerA(), planeSupportApp("client-shared"))

	_, err := fx.apps.CreatePlaneApp(planeOwnerB(), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: planeSupportApp("client-shared"),
	}))
	wantPlaneCode(t, err, connect.CodeAlreadyExists, "duplicate client ID in another workspace")

	app := fx.create(t, planeOwnerA(), planeSupportApp("client-a"))
	_, err = fx.apps.UpdatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{Id: app.GetId(), ClientId: "client-renamed", DisplayName: "Renamed", AgentId: "support", Revision: app.GetRevision()},
	}))
	wantPlaneCode(t, err, connect.CodeInvalidArgument, "update client ID")
}

func TestPlaneAllowedUsersAreCanonicalized(t *testing.T) {
	fx := newPlaneFixture(t)
	app := fx.create(t, planeOwnerA(), &agentsv1.PlaneApp{
		DisplayName:    "Support",
		ClientId:       "client-a",
		AgentId:        "support",
		AllowedUserIds: []string{strings.ToUpper(planeUserA), planeUserA, " "},
	})
	if got := app.GetAllowedUserIds(); len(got) != 1 || got[0] != planeUserA {
		t.Fatalf("allowed_user_ids = %v; want the canonical ID once", got)
	}
	_, err := fx.apps.CreatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: &agentsv1.PlaneApp{DisplayName: "Bad", ClientId: "client-b", AgentId: "support", AllowedUserIds: []string{"not-a-uuid"}},
	}))
	wantPlaneCode(t, err, connect.CodeInvalidArgument, "invalid allowed user ID")
}

func TestPlaneRevisionIsRequiredAndFenced(t *testing.T) {
	fx := newPlaneFixture(t)
	app := fx.create(t, planeOwnerA(), planeSupportApp("client-a"))

	_, err := fx.apps.UpdatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{Id: app.GetId(), DisplayName: "No revision", AgentId: "support"},
	}))
	wantPlaneCode(t, err, connect.CodeInvalidArgument, "update without revision")

	if _, err := fx.apps.UpdatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{Id: app.GetId(), DisplayName: "Renamed", AgentId: "research", Revision: app.GetRevision()},
	})); err != nil {
		t.Fatalf("UpdatePlaneApp: %v", err)
	}
	updated, err := fx.apps.GetPlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.GetPlaneAppRequest{Id: app.GetId()}))
	if err != nil {
		t.Fatalf("GetPlaneApp: %v", err)
	}
	if got := updated.Msg.GetApp(); got.GetDisplayName() != "Renamed" || got.GetAgentId() != "research" || got.GetRevision() != app.GetRevision()+1 {
		t.Fatalf("updated = %+v; want Renamed/research at revision %d", got, app.GetRevision()+1)
	}

	// The stale revision loses: two people editing the same App cannot
	// silently overwrite each other.
	_, err = fx.apps.UpdatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{Id: app.GetId(), DisplayName: "Stale", AgentId: "support", Revision: app.GetRevision()},
	}))
	wantPlaneCode(t, err, connect.CodeAborted, "stale update")
	again, _ := fx.apps.GetPlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.GetPlaneAppRequest{Id: app.GetId()}))
	if again.Msg.GetApp().GetDisplayName() != "Renamed" {
		t.Fatalf("stale update was applied: %q", again.Msg.GetApp().GetDisplayName())
	}
}

func TestPlaneSecretsAreWriteOnlyAndEncrypted(t *testing.T) {
	fx := newPlaneFixture(t)
	app := fx.create(t, planeOwnerA(), planeSupportApp("client-a"), planeClientSecret, planeWebhookSecret)

	if !app.GetClientSecretSet() || !app.GetWebhookSecretSet() {
		t.Fatalf("flags = client:%v webhook:%v; want both set", app.GetClientSecretSet(), app.GetWebhookSecretSet())
	}
	if app.GetCredentialState() != agentsv1.PlaneAppCredentialState_PLANE_APP_CREDENTIAL_STATE_COMPLETE {
		t.Fatalf("credential_state = %v, want COMPLETE", app.GetCredentialState())
	}
	// The public model must not carry either secret, in any encoding.
	encoded, err := protojson.Marshal(app)
	if err != nil {
		t.Fatalf("marshal app: %v", err)
	}
	for _, secret := range []string{planeClientSecret, planeWebhookSecret} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("secret %q leaked into the RPC response: %s", secret, encoded)
		}
	}

	// At rest the secrets are ciphertext from the credential seam, never the
	// plaintext, and the key that sealed them is recorded.
	stored, err := fx.repo.GetAppCredentials(t.Context(), "ws-a", app.GetId())
	if err != nil {
		t.Fatalf("GetAppCredentials: %v", err)
	}
	if stored.ClientSecret.Ciphertext == "" || stored.WebhookSecret.Ciphertext == "" {
		t.Fatal("secrets were not stored")
	}
	if stored.ClientSecret.Ciphertext == planeClientSecret || stored.WebhookSecret.Ciphertext == planeWebhookSecret {
		t.Fatal("secret was stored in plaintext")
	}
	if stored.ClientSecret.KeyID == "" || stored.WebhookSecret.KeyID == "" {
		t.Fatal("credential key ID was not recorded")
	}

	// UpdatePlaneApp has no secret fields, so editing the resource keeps
	// them: the flags survive and the ciphertext is unchanged.
	updated, err := fx.apps.UpdatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{Id: app.GetId(), DisplayName: "Support 2", AgentId: "support", Revision: app.GetRevision()},
	}))
	if err != nil {
		t.Fatalf("UpdatePlaneApp: %v", err)
	}
	if !updated.Msg.GetApp().GetClientSecretSet() || !updated.Msg.GetApp().GetWebhookSecretSet() {
		t.Fatal("UpdatePlaneApp cleared the secrets")
	}
	kept, _ := fx.repo.GetAppCredentials(t.Context(), "ws-a", app.GetId())
	if kept != stored {
		t.Fatal("UpdatePlaneApp rewrote the stored credentials")
	}

	// Absent keeps, empty clears: the same semantics as every other
	// credential in Butter.
	webhookOnly, err := fx.apps.PutPlaneAppCredentials(planeOwnerA(), connect.NewRequest(&agentsv1.PutPlaneAppCredentialsRequest{
		AppId:        app.GetId(),
		ClientSecret: proto.String(""),
	}))
	if err != nil {
		t.Fatalf("clear client secret: %v", err)
	}
	if webhookOnly.Msg.GetApp().GetClientSecretSet() || !webhookOnly.Msg.GetApp().GetWebhookSecretSet() {
		t.Fatalf("after clearing = client:%v webhook:%v; want webhook only",
			webhookOnly.Msg.GetApp().GetClientSecretSet(), webhookOnly.Msg.GetApp().GetWebhookSecretSet())
	}
	if webhookOnly.Msg.GetApp().GetCredentialState() != agentsv1.PlaneAppCredentialState_PLANE_APP_CREDENTIAL_STATE_INCOMPLETE {
		t.Fatalf("credential_state = %v with one secret, want INCOMPLETE", webhookOnly.Msg.GetApp().GetCredentialState())
	}
	onlyWebhook, _ := fx.repo.GetAppCredentials(t.Context(), "ws-a", app.GetId())
	if onlyWebhook.ClientSecret.Set() {
		t.Fatal("cleared client secret is still stored")
	}
	if onlyWebhook.WebhookSecret != stored.WebhookSecret {
		t.Fatal("absent client-secret update also rewrote the webhook secret")
	}

	// Rotating one secret leaves the other alone.
	rotated, err := fx.apps.PutPlaneAppCredentials(planeOwnerA(), connect.NewRequest(&agentsv1.PutPlaneAppCredentialsRequest{
		AppId:        app.GetId(),
		ClientSecret: proto.String("rotated-client-secret"),
	}))
	if err != nil {
		t.Fatalf("rotate client secret: %v", err)
	}
	if !rotated.Msg.GetApp().GetClientSecretSet() || !rotated.Msg.GetApp().GetWebhookSecretSet() {
		t.Fatal("rotation lost a secret flag")
	}
	afterRotation, _ := fx.repo.GetAppCredentials(t.Context(), "ws-a", app.GetId())
	if afterRotation.ClientSecret == stored.ClientSecret || afterRotation.WebhookSecret != stored.WebhookSecret {
		t.Fatal("rotation did not replace exactly one secret")
	}
}

func TestPlaneInboundIsDefaultOffAndCannotBeEnabled(t *testing.T) {
	fx := newPlaneFixture(t)

	// Default-off applies even when the caller does not mention the field.
	app := fx.create(t, planeOwnerA(), planeSupportApp("client-a"))
	if app.GetInboundEnabled() {
		t.Fatal("a new Plane App must start with inbound disabled")
	}

	// Default-off is enforced on creation, not merely defaulted.
	_, err := fx.apps.CreatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
		App: &agentsv1.PlaneApp{DisplayName: "Live", ClientId: "client-b", AgentId: "support", InboundEnabled: true},
	}))
	wantPlaneCode(t, err, connect.CodeFailedPrecondition, "create with inbound enabled")

	// And it cannot be enabled later: the execution path is a later ticket.
	_, err = fx.apps.UpdatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{
			Id: app.GetId(), DisplayName: "Support", AgentId: "support",
			InboundEnabled: true, Revision: app.GetRevision(),
		},
	}))
	wantPlaneCode(t, err, connect.CodeFailedPrecondition, "enable inbound")

	stored, err := fx.apps.GetPlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.GetPlaneAppRequest{Id: app.GetId()}))
	if err != nil || stored.Msg.GetApp().GetInboundEnabled() {
		t.Fatalf("stored inbound = %v, %v; want disabled", stored.Msg.GetApp().GetInboundEnabled(), err)
	}
}

func TestPlaneAgentMustBeActiveInTheSameWorkspace(t *testing.T) {
	fx := newPlaneFixture(t)
	for _, tc := range []struct{ name, agent, want string }{
		{"unknown", "missing", "unknown agent"},
		{"deleted", "retired", "not active"},
		{"other workspace", "ws-b-agent", "unknown agent"},
	} {
		agent := tc.agent
		if tc.name == "other workspace" {
			// ws-a requesting an Agent that only exists in ws-b.
			fx.config.CreateAgent(t.Context(), "ws-b", &agentsv1.Agent{Name: "B", AgentId: "b-only"})
			agent = "b-only"
		}
		_, err := fx.apps.CreatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.CreatePlaneAppRequest{
			App: &agentsv1.PlaneApp{DisplayName: "X", ClientId: "client-" + strings.ReplaceAll(tc.name, " ", "-"), AgentId: agent},
		}))
		wantPlaneCode(t, err, connect.CodeInvalidArgument, tc.name)
		if err != nil && !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s: error %q does not explain %q", tc.name, err, tc.want)
		}
	}
}

func TestPlaneReferenceGuardNamesReferencingApps(t *testing.T) {
	fx := newPlaneFixture(t)
	app := fx.create(t, planeOwnerA(), planeSupportApp("client-a"))
	guard := NewPlaneReferenceGuard(fx.repo)

	err := guard.CheckAgentRemovable(t.Context(), "ws-a", "support")
	wantPlaneCode(t, err, connect.CodeFailedPrecondition, "delete a referenced agent")
	if !strings.Contains(err.Error(), app.GetId()) || !strings.Contains(err.Error(), "Support") {
		t.Fatalf("error %q must name the referencing App", err)
	}

	// Another workspace's Apps are not a reference here.
	if err := guard.CheckAgentRemovable(t.Context(), "ws-b", "support"); err != nil {
		t.Fatalf("ws-b guard = %v; ws-a's App must not block it", err)
	}

	// Repointing the App releases the Agent.
	if _, err := fx.apps.UpdatePlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.UpdatePlaneAppRequest{
		App: &agentsv1.PlaneApp{Id: app.GetId(), DisplayName: "Support", AgentId: "research", Revision: app.GetRevision()},
	})); err != nil {
		t.Fatalf("repoint App: %v", err)
	}
	if err := guard.CheckAgentRemovable(t.Context(), "ws-a", "support"); err != nil {
		t.Fatalf("guard after repoint = %v, want nil", err)
	}
}

func TestPlaneAppOnlyExistsAfterExplicitRegistration(t *testing.T) {
	fx := newPlaneFixture(t)
	// Nothing in this ticket reacts to a redirect, callback or poll: an
	// unknown ID is simply NotFound, never a lazily created App.
	_, err := fx.apps.GetPlaneApp(planeOwnerA(), connect.NewRequest(&agentsv1.GetPlaneAppRequest{Id: "never-created"}))
	wantPlaneCode(t, err, connect.CodeNotFound, "GetPlaneApp without registration")
	if apps, _ := fx.apps.ListPlaneApps(planeOwnerA(), connect.NewRequest(&agentsv1.ListPlaneAppsRequest{})); len(apps.Msg.GetApps()) != 0 {
		t.Fatalf("an App appeared without CreatePlaneApp: %+v", apps.Msg.GetApps())
	}
}

func TestPlaneAdminRequiresGlobalAdminAndFreezesTheProfile(t *testing.T) {
	fx := newPlaneFixture(t)

	for name, ctx := range map[string]context.Context{
		"member": planeMemberA(),
		"owner":  planeOwnerA(),
		"admin":  planeAdminA(),
	} {
		_, err := fx.admin.GetPlaneSettings(ctx, connect.NewRequest(&agentsv1.GetPlaneSettingsRequest{}))
		wantPlaneCode(t, err, connect.CodePermissionDenied, name+" GetPlaneSettings")
		_, err = fx.admin.UpdatePlaneSettings(ctx, connect.NewRequest(&agentsv1.UpdatePlaneSettingsRequest{
			Settings: &agentsv1.PlaneSettings{PublicBaseUrl: "https://evil.test"},
		}))
		wantPlaneCode(t, err, connect.CodePermissionDenied, name+" UpdatePlaneSettings")
	}
	if settings, _ := fx.settings.Get(t.Context()); settings.GetPublicBaseUrl() != "" {
		t.Fatalf("a non-admin write was persisted: %q", settings.GetPublicBaseUrl())
	}

	// A global admin sees the fixed profile and the unset base URL.
	got, err := fx.admin.GetPlaneSettings(globalAdminCtx(), connect.NewRequest(&agentsv1.GetPlaneSettingsRequest{}))
	if err != nil {
		t.Fatalf("admin GetPlaneSettings: %v", err)
	}
	provider := got.Msg.GetProvider()
	if provider.GetName() != "cloud" || provider.GetApiBaseUrl() != "https://api.plane.so" || provider.GetOauthBaseUrl() != "https://api.plane.so" {
		t.Fatalf("provider = %+v; want the fixed Cloud profile", provider)
	}
	// Gate A has not passed: the profile must not claim verification.
	if provider.GetVerified() {
		t.Fatal("the Cloud profile claimed verification before Gate A")
	}

	// Only the public base URL is writable, and each invalid form is
	// rejected rather than stored.
	for _, bad := range []string{"http://butter.example.com", "https://butter.example.com/api", "https://user:pass@butter.example.com", "ftp://butter.example.com"} {
		_, err := fx.admin.UpdatePlaneSettings(globalAdminCtx(), connect.NewRequest(&agentsv1.UpdatePlaneSettingsRequest{
			Settings: &agentsv1.PlaneSettings{PublicBaseUrl: bad},
		}))
		wantPlaneCode(t, err, connect.CodeInvalidArgument, "UpdatePlaneSettings "+bad)
	}
	if settings, _ := fx.settings.Get(t.Context()); settings.GetPublicBaseUrl() != "" {
		t.Fatalf("an invalid base URL was persisted: %q", settings.GetPublicBaseUrl())
	}

	stored, err := fx.admin.UpdatePlaneSettings(globalAdminCtx(), connect.NewRequest(&agentsv1.UpdatePlaneSettingsRequest{
		Settings: &agentsv1.PlaneSettings{PublicBaseUrl: "https://butter.example.com/"},
	}))
	if err != nil {
		t.Fatalf("UpdatePlaneSettings: %v", err)
	}
	if stored.Msg.GetSettings().GetPublicBaseUrl() != "https://butter.example.com" {
		t.Fatalf("stored base URL = %q; want it trimmed", stored.Msg.GetSettings().GetPublicBaseUrl())
	}
	if stored.Msg.GetProvider().GetApiBaseUrl() != "https://api.plane.so" {
		t.Fatal("updating the base URL changed the provider profile")
	}
}
