package http

// Install-flow tests at the public HTTP boundary (ADR-0015, #362): the
// dashboard starts an install, Linear redirects the browser to the shared
// callback, and the Installation is stored with encrypted tokens.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/linearapi/lineartest"
	"go.orx.me/apps/butter/internal/repo/auth"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	linearmemory "go.orx.me/apps/butter/internal/repo/linear/memory"
	linearsettingmemory "go.orx.me/apps/butter/internal/repo/linearsetting/memory"
	linearstatememory "go.orx.me/apps/butter/internal/repo/linearstate/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	installClientSecret = "lin-client-secret"
	installAccessToken  = "lin-access-token-1"
	installRefreshToken = "lin-refresh-token-1"
)

type installFixture struct {
	svc     *application.LinearAppServiceServer
	repo    *linearmemory.Store
	keyring *secretbox.Keyring
	linear  *lineartest.Fake
	router  *gin.Engine
	app     *agentsv1.LinearApp
	now     time.Time
}

func linearCtx(userID string) context.Context {
	ctx := workspace.WithID(context.Background(), "ws-a")
	return auth.WithAuthenticated(ctx, &agentsv1.User{Id: userID, Role: "user"}, nil)
}

func newInstallFixture(t *testing.T) *installFixture {
	t.Helper()
	wsRepo := workspacememory.New()
	if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: "ws-a", Name: "ws-a", Slug: "ws-a"}); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	for _, m := range []struct{ user, role string }{{"owner", "owner"}, {"member", "member"}} {
		if _, err := wsRepo.AddMember(t.Context(), &agentsv1.WorkspaceMember{WorkspaceId: "ws-a", UserId: m.user, Role: m.role}); err != nil {
			t.Fatalf("seed member: %v", err)
		}
	}
	agents := configmemory.New()
	if _, err := agents.CreateAgent(t.Context(), "ws-a", &agentsv1.Agent{Name: "support", AgentId: "support"}); err != nil {
		t.Fatalf("seed agent: %v", err)
	}
	settings := linearsettingmemory.New()
	if _, err := settings.Put(t.Context(), &agentsv1.LinearSettings{PublicBaseUrl: "https://butter.test"}); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	fake := lineartest.New(t)
	repo := linearmemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())

	fx := &installFixture{repo: repo, keyring: keyring, linear: fake, now: time.Now()}
	svc := application.NewLinearAppServiceServer(repo)
	svc.SetWorkspaceRepo(wsRepo)
	svc.SetAgentRepo(agents)
	svc.SetKeyring(keyring)
	svc.SetSettingsRepo(settings)
	svc.SetInstallStateRepo(linearstatememory.New())
	svc.SetLinearClient(fake.Client())
	svc.SetClock(func() time.Time { return fx.now })
	fx.svc = svc

	created, err := svc.CreateLinearApp(linearCtx("owner"), connect.NewRequest(&agentsv1.CreateLinearAppRequest{
		App:          &agentsv1.LinearApp{DisplayName: "Support", ClientId: "client-1", AgentId: "support"},
		ClientSecret: proto.String(installClientSecret),
	}))
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	fx.app = created.Msg.GetApp()

	gin.SetMode(gin.TestMode)
	fx.router = gin.New()
	fx.router.Use(AuthMiddleware(&config.AppConfig{}, nil, nil, nil))
	NewLinearOAuthHandler(svc).Register(fx.router)
	return fx
}

// begin starts an install as the owner and returns Linear's authorize URL.
func (fx *installFixture) begin(t *testing.T, returnURL string) *url.URL {
	t.Helper()
	resp, err := fx.svc.BeginLinearInstall(linearCtx("owner"), connect.NewRequest(&agentsv1.BeginLinearInstallRequest{
		AppId: fx.app.GetId(), ReturnUrl: returnURL,
	}))
	if err != nil {
		t.Fatalf("BeginLinearInstall: %v", err)
	}
	u, err := url.Parse(resp.Msg.GetAuthorizeUrl())
	if err != nil {
		t.Fatalf("parse authorize URL: %v", err)
	}
	return u
}

// callback plays the browser returning from Linear.
func (fx *installFixture) callback(t *testing.T, query url.Values) *url.URL {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/linear/oauth/callback?"+query.Encode(), nil)
	w := httptest.NewRecorder()
	fx.router.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("callback status = %d, want 302; body %s", w.Code, w.Body.String())
	}
	loc, err := url.Parse(w.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect: %v", err)
	}
	return loc
}

func (fx *installFixture) installations(t *testing.T) []*agentsv1.LinearInstallation {
	t.Helper()
	resp, err := fx.svc.ListLinearInstallations(linearCtx("member"), connect.NewRequest(&agentsv1.ListLinearInstallationsRequest{AppId: fx.app.GetId()}))
	if err != nil {
		t.Fatalf("ListLinearInstallations: %v", err)
	}
	return resp.Msg.GetInstallations()
}

func acmeGrant() lineartest.Grant {
	return lineartest.Grant{
		AccessToken: installAccessToken, RefreshToken: installRefreshToken, ExpiresIn: 86400,
		Scopes:   []string{"read", "write", "app:assignable", "app:mentionable"},
		Identity: linearapi.Identity{AppUserID: "app-user-1", OrganizationID: "org-1", OrganizationName: "Acme"},
	}
}

func TestInstallingALinearAppStoresTheInstallationWithEncryptedTokens(t *testing.T) {
	fx := newInstallFixture(t)
	authorize := fx.begin(t, "/linear-apps/"+fx.app.GetId()+"/edit")

	q := authorize.Query()
	if got := authorize.Scheme + "://" + authorize.Host + authorize.Path; got != "https://linear.test/oauth/authorize" {
		t.Fatalf("authorize endpoint = %q", got)
	}
	for key, want := range map[string]string{
		"client_id":     "client-1",
		"redirect_uri":  "https://butter.test/api/linear/oauth/callback",
		"response_type": "code",
		"scope":         "read,write,app:assignable,app:mentionable",
		"actor":         "app",
	} {
		if q.Get(key) != want {
			t.Errorf("authorize %s = %q, want %q", key, q.Get(key), want)
		}
	}
	if q.Get("state") == "" {
		t.Fatal("authorize URL carries no state")
	}

	fx.linear.AddCode("code-1", acmeGrant())
	loc := fx.callback(t, url.Values{"state": {q.Get("state")}, "code": {"code-1"}})
	if loc.Path != "/linear-apps/"+fx.app.GetId()+"/edit" || loc.Query().Get("linear_install") != "success" {
		t.Fatalf("redirect = %s; want the App page with linear_install=success", loc)
	}

	grants := fx.linear.TokenRequests()
	if len(grants) != 1 || grants[0].Get("client_secret") != installClientSecret ||
		grants[0].Get("redirect_uri") != "https://butter.test/api/linear/oauth/callback" {
		t.Fatalf("token requests = %v; want one code exchange with the client secret and redirect URI", grants)
	}

	installs := fx.installations(t)
	if len(installs) != 1 {
		t.Fatalf("installations = %d, want 1", len(installs))
	}
	inst := installs[0]
	if inst.GetOrganizationId() != "org-1" || inst.GetOrganizationName() != "Acme" || inst.GetAppUserId() != "app-user-1" {
		t.Fatalf("installation = %+v", inst)
	}
	if inst.GetCredentialState() != agentsv1.LinearInstallationCredentialState_LINEAR_INSTALLATION_CREDENTIAL_STATE_VALID {
		t.Fatalf("credential_state = %v, want VALID", inst.GetCredentialState())
	}
	if loc.Query().Get("installation_id") != inst.GetId() {
		t.Errorf("redirect installation_id = %q, want %q", loc.Query().Get("installation_id"), inst.GetId())
	}
	if wire := protojson.Format(inst); strings.Contains(wire, installAccessToken) || strings.Contains(wire, installRefreshToken) {
		t.Fatalf("the installation leaked a token: %s", wire)
	}

	tokens, err := fx.repo.GetInstallationTokens(t.Context(), "ws-a", inst.GetId())
	if err != nil {
		t.Fatalf("GetInstallationTokens: %v", err)
	}
	if tokens.AccessToken.Ciphertext == installAccessToken {
		t.Fatal("access token stored in plaintext")
	}
	plain, err := fx.keyring.Decrypt(t.Context(), tokens.AccessToken.Ciphertext, tokens.AccessToken.KeyID)
	if err != nil || string(plain) != installAccessToken {
		t.Fatalf("decrypted access token = %q, %v", plain, err)
	}
	if want := fx.now.Add(24 * time.Hour); tokens.ExpiresAt.Sub(want).Abs() > time.Second {
		t.Fatalf("expires_at = %v, want about %v", tokens.ExpiresAt, want)
	}
}

func TestReinstallingIntoTheSameOrganizationUpdatesTheInstallation(t *testing.T) {
	fx := newInstallFixture(t)
	for i, code := range []string{"code-1", "code-2"} {
		state := fx.begin(t, "").Query().Get("state")
		grant := acmeGrant()
		grant.AccessToken = code + "-access"
		fx.linear.AddCode(code, grant)
		if loc := fx.callback(t, url.Values{"state": {state}, "code": {code}}); loc.Query().Get("linear_install") != "success" {
			t.Fatalf("install %d redirect = %s", i, loc)
		}
	}
	if installs := fx.installations(t); len(installs) != 1 {
		t.Fatalf("installations = %d, want 1 after installing twice", len(installs))
	}
}

func TestInstallCallbackFailuresWriteNothing(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reason string
		setup  func(fx *installFixture, state string) url.Values
	}{
		{
			name:   "Linear reported an error",
			reason: "denied",
			setup: func(_ *installFixture, state string) url.Values {
				return url.Values{"state": {state}, "error": {"access_denied"}}
			},
		},
		{
			name:   "unknown state",
			reason: "state_invalid",
			setup: func(_ *installFixture, _ string) url.Values {
				return url.Values{"state": {"forged"}, "code": {"code-1"}}
			},
		},
		{
			name:   "expired state",
			reason: "state_invalid",
			setup: func(fx *installFixture, state string) url.Values {
				fx.now = fx.now.Add(11 * time.Minute)
				return url.Values{"state": {state}, "code": {"code-1"}}
			},
		},
		{
			name:   "code exchange refused",
			reason: "exchange_failed",
			setup: func(fx *installFixture, state string) url.Values {
				fx.linear.FailNextToken(lineartest.Failure{Status: http.StatusBadRequest, OAuthError: "invalid_grant"})
				return url.Values{"state": {state}, "code": {"code-1"}}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newInstallFixture(t)
			state := fx.begin(t, "").Query().Get("state")
			fx.linear.AddCode("code-1", acmeGrant())
			loc := fx.callback(t, tc.setup(fx, state))
			if loc.Query().Get("linear_install") != "error" || loc.Query().Get("reason") != tc.reason {
				t.Fatalf("redirect = %s; want linear_install=error reason=%s", loc, tc.reason)
			}
			if installs := fx.installations(t); len(installs) != 0 {
				t.Fatalf("installations = %d, want none", len(installs))
			}
		})
	}
}

func TestAnInstallStateIsSingleUse(t *testing.T) {
	fx := newInstallFixture(t)
	state := fx.begin(t, "").Query().Get("state")
	fx.linear.AddCode("code-1", acmeGrant())
	if loc := fx.callback(t, url.Values{"state": {state}, "code": {"code-1"}}); loc.Query().Get("linear_install") != "success" {
		t.Fatalf("first callback = %s", loc)
	}
	fx.linear.AddCode("code-1", acmeGrant())
	if loc := fx.callback(t, url.Values{"state": {state}, "code": {"code-1"}}); loc.Query().Get("reason") != "state_invalid" {
		t.Fatalf("replayed callback = %s; want reason=state_invalid", loc)
	}
}

func TestBeginLinearInstallPreconditions(t *testing.T) {
	fx := newInstallFixture(t)
	_, err := fx.svc.BeginLinearInstall(linearCtx("member"), connect.NewRequest(&agentsv1.BeginLinearInstallRequest{AppId: fx.app.GetId()}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("member begin = %v, want PermissionDenied", err)
	}

	// Without the client secret the code could never be exchanged.
	if _, err := fx.svc.PutLinearAppCredentials(linearCtx("owner"), connect.NewRequest(&agentsv1.PutLinearAppCredentialsRequest{
		AppId: fx.app.GetId(), ClientSecret: proto.String(""),
	})); err != nil {
		t.Fatalf("clear client secret: %v", err)
	}
	_, err = fx.svc.BeginLinearInstall(linearCtx("owner"), connect.NewRequest(&agentsv1.BeginLinearInstallRequest{AppId: fx.app.GetId()}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("begin without client secret = %v, want FailedPrecondition", err)
	}
}

func TestBeginLinearInstallNeedsTheBaseURL(t *testing.T) {
	fx := newInstallFixture(t)
	settings := linearsettingmemory.New()
	fx.svc.SetSettingsRepo(settings)
	_, err := fx.svc.BeginLinearInstall(linearCtx("owner"), connect.NewRequest(&agentsv1.BeginLinearInstallRequest{AppId: fx.app.GetId()}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("begin without base URL = %v, want FailedPrecondition", err)
	}
}

func TestInstallReturnURLMustStayOnAKnownOrigin(t *testing.T) {
	for returnURL, wantPath := range map[string]string{
		"https://evil.example/steal":            "",
		"//evil.example/steal":                  "",
		"https://butter.test/linear-apps/x/edit": "/linear-apps/x/edit",
		"/linear-apps":                          "/linear-apps",
	} {
		t.Run(returnURL, func(t *testing.T) {
			fx := newInstallFixture(t)
			if wantPath == "" {
				wantPath = "/linear-apps/" + fx.app.GetId() + "/edit"
			}
			state := fx.begin(t, returnURL).Query().Get("state")
			fx.linear.AddCode("code-1", acmeGrant())
			loc := fx.callback(t, url.Values{"state": {state}, "code": {"code-1"}})
			if loc.Host != "" && loc.Host != "butter.test" {
				t.Fatalf("redirected off-origin to %s", loc)
			}
			if loc.Path != wantPath {
				t.Fatalf("redirect path = %q, want %q", loc.Path, wantPath)
			}
		})
	}
}

func TestRemovingAnInstallationRevokesItsTokenBestEffort(t *testing.T) {
	fx := newInstallFixture(t)
	for i, org := range []string{"org-1", "org-2"} {
		state := fx.begin(t, "").Query().Get("state")
		grant := acmeGrant()
		grant.AccessToken = org + "-access"
		grant.Identity.OrganizationID = org
		fx.linear.AddCode(org, grant)
		if loc := fx.callback(t, url.Values{"state": {state}, "code": {org}}); loc.Query().Get("linear_install") != "success" {
			t.Fatalf("install %d = %s", i, loc)
		}
	}
	installs := fx.installations(t)
	del := func(ctx context.Context, id string) error {
		_, err := fx.svc.DeleteLinearInstallation(ctx, connect.NewRequest(&agentsv1.DeleteLinearInstallationRequest{AppId: fx.app.GetId(), Id: id}))
		return err
	}
	if err := del(linearCtx("member"), installs[0].GetId()); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("member delete = %v, want PermissionDenied", err)
	}
	if err := del(linearCtx("owner"), installs[0].GetId()); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if revoked := fx.linear.Revoked(); len(revoked) != 1 || !strings.HasSuffix(revoked[0], "-access") {
		t.Fatalf("revoked = %v; want the removed installation's token", revoked)
	}
	// A failed revoke does not keep the installation.
	fx.linear.FailNextRevoke()
	if err := del(linearCtx("owner"), installs[1].GetId()); err != nil {
		t.Fatalf("delete with a failing revoke: %v", err)
	}
	if left := fx.installations(t); len(left) != 0 {
		t.Fatalf("installations left = %d, want 0", len(left))
	}
}

func TestLinearOAuthCallbackIsPublic(t *testing.T) {
	if !isPublicPath("/api/linear/oauth/callback") {
		t.Fatal("the Linear OAuth callback must bypass authentication: the browser arrives from Linear")
	}
}
