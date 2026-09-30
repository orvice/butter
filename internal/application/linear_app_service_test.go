package application

// Service-level tests for LinearAppService and LinearAdminService
// (ADR-0015, #361): role enforcement, write-only secrets, derived URLs,
// optimistic revisions, strong Agent references, and the inbound
// preconditions.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/auth"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	linearmemory "go.orx.me/apps/butter/internal/repo/linear/memory"
	linearsettingmemory "go.orx.me/apps/butter/internal/repo/linearsetting/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	linearClientSecret  = "lin-client-secret-value"
	linearWebhookSecret = "lin-webhook-secret-value"
	linearUserA         = "8f7e6d5c-4b3a-4a1b-9c8d-7e6f5a4b3c2d"
)

type linearFixture struct {
	apps     *LinearAppServiceServer
	admin    *LinearAdminServiceServer
	repo     *linearmemory.Store
	settings *linearsettingmemory.Store
	config   *configmemory.Store
	queue    *stubQueueProbe
}

func newLinearFixture(t *testing.T) *linearFixture {
	t.Helper()
	wsRepo := workspacememory.New()
	for _, ws := range []string{"ws-a", "ws-b"} {
		if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: ws, Name: ws, Slug: ws}); err != nil {
			t.Fatalf("seed workspace %s: %v", ws, err)
		}
	}
	for _, m := range []struct{ user, role, ws string }{
		{"owner", "owner", "ws-a"},
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
	settings := linearsettingmemory.New()
	if _, err := settings.Put(t.Context(), &agentsv1.LinearSettings{PublicBaseUrl: "https://butter.test"}); err != nil {
		t.Fatalf("seed linear settings: %v", err)
	}
	repo := linearmemory.New()
	apps := NewLinearAppServiceServer(repo)
	apps.SetWorkspaceRepo(wsRepo)
	apps.SetAgentRepo(configStore)
	apps.SetKeyring(secretbox.NewKeyring(cryptokeymemory.New()))
	apps.SetSettingsRepo(settings)
	queue := &stubQueueProbe{available: true}
	apps.SetQueueProbe(queue)
	return &linearFixture{
		apps: apps, admin: NewLinearAdminServiceServer(settings),
		repo: repo, settings: settings, config: configStore, queue: queue,
	}
}

func ownerA() context.Context  { return ctxAs("owner", "user", "ws-a") }
func memberA() context.Context { return ctxAs("member", "user", "ws-a") }

func globalAdminCtx() context.Context {
	return auth.WithAdmin(auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "root", Role: "admin"}, nil))
}

func (fx *linearFixture) createApp(t *testing.T, ctx context.Context, app *agentsv1.LinearApp, secrets ...string) *agentsv1.LinearApp {
	t.Helper()
	req := &agentsv1.CreateLinearAppRequest{App: app}
	if len(secrets) > 0 {
		req.ClientSecret = proto.String(secrets[0])
	}
	if len(secrets) > 1 {
		req.WebhookSecret = proto.String(secrets[1])
	}
	resp, err := fx.apps.CreateLinearApp(ctx, connect.NewRequest(req))
	if err != nil {
		t.Fatalf("CreateLinearApp: %v", err)
	}
	return resp.Msg.GetApp()
}

func supportApp(clientID string) *agentsv1.LinearApp {
	return &agentsv1.LinearApp{DisplayName: "Support", ClientId: clientID, AgentId: "support"}
}

func TestOwnerRegistersALinearAppWithWriteOnlySecretsAndDerivedURLs(t *testing.T) {
	fx := newLinearFixture(t)
	app := fx.createApp(t, ownerA(), supportApp("client-1"), linearClientSecret, linearWebhookSecret)

	if app.GetId() == "" || app.GetRevision() != 1 || app.GetWorkspaceId() != "ws-a" {
		t.Fatalf("created = %+v; want an ID, revision 1, workspace ws-a", app)
	}
	if app.GetCredentialState() != agentsv1.LinearAppCredentialState_LINEAR_APP_CREDENTIAL_STATE_COMPLETE {
		t.Fatalf("credential_state = %v, want COMPLETE", app.GetCredentialState())
	}
	if app.GetCallbackUrl() != "https://butter.test/api/linear/oauth/callback" {
		t.Errorf("callback_url = %q", app.GetCallbackUrl())
	}
	if app.GetWebhookUrl() != "https://butter.test/api/linear/webhook/"+app.GetId() {
		t.Errorf("webhook_url = %q", app.GetWebhookUrl())
	}

	// No read path ever returns a secret.
	got, err := fx.apps.GetLinearApp(memberA(), connect.NewRequest(&agentsv1.GetLinearAppRequest{Id: app.GetId()}))
	if err != nil {
		t.Fatalf("member GetLinearApp: %v", err)
	}
	list, err := fx.apps.ListLinearApps(memberA(), connect.NewRequest(&agentsv1.ListLinearAppsRequest{}))
	if err != nil {
		t.Fatalf("member ListLinearApps: %v", err)
	}
	for _, msg := range []proto.Message{app, got.Msg, list.Msg} {
		wire := protojson.Format(msg)
		if strings.Contains(wire, linearClientSecret) || strings.Contains(wire, linearWebhookSecret) {
			t.Fatalf("a response leaked a secret: %s", wire)
		}
	}
	if len(list.Msg.GetApps()) != 1 || list.Msg.GetApps()[0].GetWebhookUrl() == "" {
		t.Fatalf("list = %+v; want the App with its derived URLs", list.Msg.GetApps())
	}
}

func TestDerivedURLsAreEmptyUntilTheBaseURLIsConfigured(t *testing.T) {
	fx := newLinearFixture(t)
	if _, err := fx.settings.Put(t.Context(), &agentsv1.LinearSettings{}); err != nil {
		t.Fatalf("clear settings: %v", err)
	}
	app := fx.createApp(t, ownerA(), supportApp("client-1"))
	if app.GetCallbackUrl() != "" || app.GetWebhookUrl() != "" {
		t.Fatalf("urls = %q / %q; want empty without the setting", app.GetCallbackUrl(), app.GetWebhookUrl())
	}
}

func TestOnlyOwnersAndAdminsManageLinearApps(t *testing.T) {
	fx := newLinearFixture(t)
	_, err := fx.apps.CreateLinearApp(memberA(), connect.NewRequest(&agentsv1.CreateLinearAppRequest{App: supportApp("client-1")}))
	if code := connectCode(t, err); code != connect.CodePermissionDenied {
		t.Fatalf("member create code = %v, want PermissionDenied", code)
	}

	app := fx.createApp(t, ownerA(), supportApp("client-1"))
	update := proto.Clone(app).(*agentsv1.LinearApp)
	update.DisplayName = "Renamed"
	if _, err := fx.apps.UpdateLinearApp(memberA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: update})); connectCode(t, err) != connect.CodePermissionDenied {
		t.Fatalf("member update code = %v, want PermissionDenied", connect.CodeOf(err))
	}
	if _, err := fx.apps.PutLinearAppCredentials(memberA(), connect.NewRequest(&agentsv1.PutLinearAppCredentialsRequest{
		AppId: app.GetId(), ClientSecret: proto.String("x"),
	})); connectCode(t, err) != connect.CodePermissionDenied {
		t.Fatalf("member put credentials code = %v, want PermissionDenied", connect.CodeOf(err))
	}
	if _, err := fx.apps.DeleteLinearApp(memberA(), connect.NewRequest(&agentsv1.DeleteLinearAppRequest{Id: app.GetId()})); connectCode(t, err) != connect.CodePermissionDenied {
		t.Fatalf("member delete code = %v, want PermissionDenied", connect.CodeOf(err))
	}

	// A global admin may manage any workspace.
	adminInA := ctxAs("root", "admin", "ws-a")
	adminInA = auth.WithAdmin(adminInA)
	if _, err := fx.apps.DeleteLinearApp(adminInA, connect.NewRequest(&agentsv1.DeleteLinearAppRequest{Id: app.GetId()})); err != nil {
		t.Fatalf("global admin delete: %v", err)
	}
}

func TestLinearAppMustRouteToAnActiveAgentOfItsWorkspace(t *testing.T) {
	fx := newLinearFixture(t)
	for _, agentID := range []string{"", "missing", "retired"} {
		app := supportApp("client-" + agentID)
		app.AgentId = agentID
		_, err := fx.apps.CreateLinearApp(ownerA(), connect.NewRequest(&agentsv1.CreateLinearAppRequest{App: app}))
		if code := connectCode(t, err); code != connect.CodeInvalidArgument {
			t.Fatalf("agent %q: code = %v, want InvalidArgument", agentID, code)
		}
	}

	app := fx.createApp(t, ownerA(), supportApp("client-1"))
	update := proto.Clone(app).(*agentsv1.LinearApp)
	update.AgentId = "missing"
	_, err := fx.apps.UpdateLinearApp(ownerA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: update}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("update to an unknown agent: code = %v, want InvalidArgument", code)
	}
}

func TestLinearAppFieldValidation(t *testing.T) {
	fx := newLinearFixture(t)
	for name, mutate := range map[string]func(*agentsv1.LinearApp){
		"missing display name": func(a *agentsv1.LinearApp) { a.DisplayName = " " },
		"missing client id":    func(a *agentsv1.LinearApp) { a.ClientId = "" },
		"non-uuid user id":     func(a *agentsv1.LinearApp) { a.AllowedUserIds = []string{"alice"} },
		"negative max run":     func(a *agentsv1.LinearApp) { a.MaxRunSeconds = proto.Int32(-1) },
		"max run over a day":   func(a *agentsv1.LinearApp) { a.MaxRunSeconds = proto.Int32(86401) },
	} {
		t.Run(name, func(t *testing.T) {
			app := supportApp("client-" + strings.ReplaceAll(name, " ", "-"))
			mutate(app)
			_, err := fx.apps.CreateLinearApp(ownerA(), connect.NewRequest(&agentsv1.CreateLinearAppRequest{App: app}))
			if code := connectCode(t, err); code != connect.CodeInvalidArgument {
				t.Fatalf("code = %v, want InvalidArgument", code)
			}
		})
	}
}

func TestAllowlistIsCanonicalizedAndMaxRunKeepsAbsentDistinctFromZero(t *testing.T) {
	fx := newLinearFixture(t)
	app := supportApp("client-1")
	app.AllowedUserIds = []string{strings.ToUpper(linearUserA), " ", linearUserA}
	app.MaxRunSeconds = proto.Int32(0)
	created := fx.createApp(t, ownerA(), app)
	if got := created.GetAllowedUserIds(); len(got) != 1 || got[0] != linearUserA {
		t.Fatalf("allowed_user_ids = %v, want one canonical ID", got)
	}
	if created.MaxRunSeconds == nil || created.GetMaxRunSeconds() != 0 {
		t.Fatalf("max_run_seconds = %v, want an explicit 0 (unlimited)", created.MaxRunSeconds)
	}

	other := fx.createApp(t, ownerA(), supportApp("client-2"))
	if other.MaxRunSeconds != nil {
		t.Fatalf("max_run_seconds = %v, want absent (default)", other.GetMaxRunSeconds())
	}
}

func TestLinearClientIDIsRegisteredOnce(t *testing.T) {
	fx := newLinearFixture(t)
	fx.createApp(t, ownerA(), supportApp("client-1"))
	_, err := fx.apps.CreateLinearApp(ctxAs("owner-b", "user", "ws-b"),
		connect.NewRequest(&agentsv1.CreateLinearAppRequest{App: supportApp("client-1")}))
	if code := connectCode(t, err); code != connect.CodeAlreadyExists {
		t.Fatalf("code = %v, want AlreadyExists", code)
	}
}

func TestUpdateIsGuardedByRevisionAndKeepsTheClientID(t *testing.T) {
	fx := newLinearFixture(t)
	app := fx.createApp(t, ownerA(), supportApp("client-1"))

	update := proto.Clone(app).(*agentsv1.LinearApp)
	update.DisplayName = "Research desk"
	update.AgentId = "research"
	update.ClientId = "client-other"
	resp, err := fx.apps.UpdateLinearApp(ownerA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: update}))
	if err != nil {
		t.Fatalf("UpdateLinearApp: %v", err)
	}
	got := resp.Msg.GetApp()
	if got.GetDisplayName() != "Research desk" || got.GetAgentId() != "research" || got.GetClientId() != "client-1" {
		t.Fatalf("updated = %+v; want new name and agent, original client ID", got)
	}

	// The revision the caller read is now stale.
	_, err = fx.apps.UpdateLinearApp(ownerA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: update}))
	if code := connectCode(t, err); code != connect.CodeAborted {
		t.Fatalf("stale update code = %v, want Aborted", code)
	}
}

func TestInboundRequiresBothSecretsAndSecretsCannotBeClearedWhileReceiving(t *testing.T) {
	fx := newLinearFixture(t)

	enabled := supportApp("client-1")
	enabled.InboundEnabled = true
	_, err := fx.apps.CreateLinearApp(ownerA(), connect.NewRequest(&agentsv1.CreateLinearAppRequest{
		App: enabled, ClientSecret: proto.String(linearClientSecret),
	}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("create inbound with one secret: code = %v, want FailedPrecondition", code)
	}

	app := fx.createApp(t, ownerA(), supportApp("client-1"), linearClientSecret)
	update := proto.Clone(app).(*agentsv1.LinearApp)
	update.InboundEnabled = true
	if _, err := fx.apps.UpdateLinearApp(ownerA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: update})); connectCode(t, err) != connect.CodeFailedPrecondition {
		t.Fatalf("enable inbound with one secret: code = %v, want FailedPrecondition", connect.CodeOf(err))
	}

	// Absent keeps the stored client secret; a value sets the webhook one.
	put, err := fx.apps.PutLinearAppCredentials(ownerA(), connect.NewRequest(&agentsv1.PutLinearAppCredentialsRequest{
		AppId: app.GetId(), WebhookSecret: proto.String(linearWebhookSecret),
	}))
	if err != nil {
		t.Fatalf("PutLinearAppCredentials: %v", err)
	}
	if put.Msg.GetApp().GetCredentialState() != agentsv1.LinearAppCredentialState_LINEAR_APP_CREDENTIAL_STATE_COMPLETE {
		t.Fatalf("credential_state = %v, want COMPLETE", put.Msg.GetApp().GetCredentialState())
	}
	resp, err := fx.apps.UpdateLinearApp(ownerA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: update}))
	if err != nil {
		t.Fatalf("enable inbound with both secrets: %v", err)
	}
	if !resp.Msg.GetApp().GetInboundEnabled() {
		t.Fatal("inbound not enabled")
	}

	// Clearing a secret would break every delivery while receiving.
	_, err = fx.apps.PutLinearAppCredentials(ownerA(), connect.NewRequest(&agentsv1.PutLinearAppCredentialsRequest{
		AppId: app.GetId(), WebhookSecret: proto.String(""),
	}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("clear while receiving: code = %v, want FailedPrecondition", code)
	}

	disable := proto.Clone(resp.Msg.GetApp()).(*agentsv1.LinearApp)
	disable.InboundEnabled = false
	if _, err := fx.apps.UpdateLinearApp(ownerA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: disable})); err != nil {
		t.Fatalf("disable inbound: %v", err)
	}
	cleared, err := fx.apps.PutLinearAppCredentials(ownerA(), connect.NewRequest(&agentsv1.PutLinearAppCredentialsRequest{
		AppId: app.GetId(), WebhookSecret: proto.String(""),
	}))
	if err != nil {
		t.Fatalf("clear after disabling: %v", err)
	}
	if cleared.Msg.GetApp().GetWebhookSecretSet() || !cleared.Msg.GetApp().GetClientSecretSet() {
		t.Fatalf("after clearing = %+v; want only the client secret kept", cleared.Msg.GetApp())
	}
}

func TestLinearAppsAreWorkspaceScoped(t *testing.T) {
	fx := newLinearFixture(t)
	app := fx.createApp(t, ownerA(), supportApp("client-1"))
	_, err := fx.apps.GetLinearApp(ctxAs("owner-b", "user", "ws-b"), connect.NewRequest(&agentsv1.GetLinearAppRequest{Id: app.GetId()}))
	if code := connectCode(t, err); code != connect.CodeNotFound {
		t.Fatalf("cross-workspace get code = %v, want NotFound", code)
	}
}

func TestAgentRoutedByALinearAppCannotBeRemoved(t *testing.T) {
	fx := newLinearFixture(t)
	app := fx.createApp(t, ownerA(), supportApp("client-1"))
	guard := NewLinearReferenceGuard(fx.repo)

	err := guard.CheckAgentRemovable(t.Context(), "ws-a", "support")
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, want FailedPrecondition", code)
	}
	if !strings.Contains(err.Error(), app.GetId()) {
		t.Errorf("error should name the App, got %q", err.Error())
	}
	if err := guard.CheckAgentRemovable(t.Context(), "ws-a", "research"); err != nil {
		t.Fatalf("an unreferenced agent was blocked: %v", err)
	}
	if err := guard.CheckAgentRemovable(t.Context(), "ws-b", "support"); err != nil {
		t.Fatalf("a cross-workspace reference blocked the delete: %v", err)
	}
}

func TestDeletingAnAgentALinearAppRoutesToIsRefused(t *testing.T) {
	fx := newLinearFixture(t)
	app := fx.createApp(t, ownerA(), supportApp("client-1"))
	agents := NewAgentServiceServer(fx.config)
	agents.SetLinearGuard(NewLinearReferenceGuard(fx.repo))

	_, err := agents.DeleteAgent(ownerA(), connect.NewRequest(&agentsv1.DeleteAgentRequest{AgentId: "support"}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("delete code = %v, want FailedPrecondition", code)
	}
	if !strings.Contains(err.Error(), app.GetId()) {
		t.Errorf("error should name the App, got %q", err.Error())
	}
}

func TestOnlyGlobalAdminsManageLinearSettings(t *testing.T) {
	fx := newLinearFixture(t)
	if _, err := fx.admin.GetLinearSettings(ownerA(), connect.NewRequest(&agentsv1.GetLinearSettingsRequest{})); connectCode(t, err) != connect.CodePermissionDenied {
		t.Fatalf("owner get code = %v, want PermissionDenied", connect.CodeOf(err))
	}
	update := func(url string) error {
		_, err := fx.admin.UpdateLinearSettings(globalAdminCtx(), connect.NewRequest(&agentsv1.UpdateLinearSettingsRequest{
			Settings: &agentsv1.LinearSettings{PublicBaseUrl: url},
		}))
		return err
	}
	for _, bad := range []string{"butter.example.com", "http://butter.example.com", "https://butter.example.com/api"} {
		if code := connectCode(t, update(bad)); code != connect.CodeInvalidArgument {
			t.Fatalf("base URL %q: code = %v, want InvalidArgument", bad, code)
		}
	}
	if err := update("https://butter.example.com/"); err != nil {
		t.Fatalf("valid base URL: %v", err)
	}
	// Loopback HTTP is allowed for local development.
	if err := update("http://localhost:8080"); err != nil {
		t.Fatalf("loopback base URL: %v", err)
	}
	got, err := fx.admin.GetLinearSettings(globalAdminCtx(), connect.NewRequest(&agentsv1.GetLinearSettingsRequest{}))
	if err != nil || got.Msg.GetSettings().GetPublicBaseUrl() != "http://localhost:8080" {
		t.Fatalf("settings = %+v, %v", got.Msg.GetSettings(), err)
	}
}

func TestInboundRequiresADurableQueue(t *testing.T) {
	fx := newLinearFixture(t)
	app := fx.createApp(t, ownerA(), supportApp("client-1"), linearClientSecret, linearWebhookSecret)
	update := proto.Clone(app).(*agentsv1.LinearApp)
	update.InboundEnabled = true

	for name, probe := range map[string]QueueProbe{
		"no queue":        nil,
		"queue not ready": &stubQueueProbe{available: true, readyErr: errors.New("maxmemory-policy must be noeviction")},
	} {
		fx.apps.SetQueueProbe(probe)
		_, err := fx.apps.UpdateLinearApp(ownerA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: update}))
		if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
			t.Fatalf("%s: code = %v, want FailedPrecondition", name, code)
		}
	}
	fx.apps.SetQueueProbe(fx.queue)
	if _, err := fx.apps.UpdateLinearApp(ownerA(), connect.NewRequest(&agentsv1.UpdateLinearAppRequest{App: update})); err != nil {
		t.Fatalf("enable with a durable queue: %v", err)
	}
}
