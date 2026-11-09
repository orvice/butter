package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.orx.me/apps/butter/internal/planeapi"
	installationmemory "go.orx.me/apps/butter/internal/repo/planeinstallation/memory"
	installationmongo "go.orx.me/apps/butter/internal/repo/planeinstallation/mongo"
	statememory "go.orx.me/apps/butter/internal/repo/planestate/memory"
	statemongo "go.orx.me/apps/butter/internal/repo/planestate/mongo"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

func installFixture(t *testing.T) (*planeFixture, *agentsv1.PlaneApp, *installationmemory.Store) {
	t.Helper()
	fx := newPlaneFixture(t)
	fx.apps.SetPlaneInstallStateRepo(statememory.New())
	installs := installationmemory.New()
	fx.apps.SetPlaneInstallationRepo(installs)
	fx.apps.SetPlaneDashboardBaseURL(func() string { return "https://dashboard.example" })
	if _, err := fx.settings.Put(t.Context(), &agentsv1.PlaneSettings{PublicBaseUrl: "https://backend.example"}); err != nil {
		t.Fatal(err)
	}
	app := fx.create(t, planeOwnerA(), planeSupportApp("client"))
	if _, err := fx.apps.PutPlaneAppCredentials(planeOwnerA(), connect.NewRequest(&agentsv1.PutPlaneAppCredentialsRequest{
		AppId: app.GetId(), ClientSecret: proto.String(planeClientSecret),
	})); err != nil {
		t.Fatal(err)
	}
	return fx, app, installs
}

func beginPlaneTestInstall(t *testing.T, fx *planeFixture, app *agentsv1.PlaneApp) string {
	t.Helper()
	resp, err := fx.apps.BeginPlaneInstall(planeOwnerA(), connect.NewRequest(&agentsv1.BeginPlaneInstallRequest{
		AppId: app.GetId(), ReturnUrl: "https://dashboard.example/plane-apps/" + app.GetId() + "/edit",
	}))
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(resp.Msg.GetAuthorizeUrl())
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/auth/o/authorize-app/" || u.Query().Get("redirect_uri") != "https://backend.example/api/plane/oauth/callback" ||
		u.Query().Get("state") == "" {
		t.Fatalf("invalid consent URL: %s", u)
	}
	return u.Query().Get("state")
}

func TestPlaneInstallBotFlowAndSingleUseState(t *testing.T) {
	testPlaneInstallBotFlow(t, false)
}

func TestPlaneInstallBotFlowMongo(t *testing.T) {
	testPlaneInstallBotFlow(t, true)
}

func testPlaneInstallBotFlow(t *testing.T, persistent bool) {
	t.Helper()
	fx, app, _ := installFixture(t)
	if persistent {
		uri := os.Getenv("BUTTER_TEST_MONGO_URI")
		if uri == "" {
			t.Skip("BUTTER_TEST_MONGO_URI is required")
		}
		client, err := mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			t.Fatal(err)
		}
		db := client.Database("butter_plane_flow_" + uuid.NewString())
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = db.Drop(ctx)
			_ = client.Disconnect(ctx)
		})
		states := statemongo.New(db)
		storedInstalls := installationmongo.New(db, fx.apps.keyring)
		if err := states.EnsureIndexes(t.Context()); err != nil {
			t.Fatal(err)
		}
		if err := storedInstalls.EnsureIndexes(t.Context()); err != nil {
			t.Fatal(err)
		}
		fx.apps.SetPlaneInstallStateRepo(states)
		fx.apps.SetPlaneInstallationRepo(storedInstalls)
	}
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/o/token/":
			client, secret, _ := r.BasicAuth()
			_ = r.ParseForm()
			if client != "client" || secret != planeClientSecret || r.Form.Get("grant_type") != "client_credentials" ||
				r.Form.Get("app_installation_id") != "external-install" || r.Form.Get("code") != "" {
				t.Error("incorrect Bot grant")
			}
			_, _ = w.Write([]byte(`{"access_token":"private-bot-token","expires_in":86400,"scope":"agents.runs:read agents.run_activities:read agents.run_activities:write"}`))
		case "/auth/o/app-installation/":
			if r.URL.Query().Get("id") != "external-install" || r.Header.Get("Authorization") != "Bearer private-bot-token" {
				t.Error("incorrect installation lookup")
			}
			_, _ = w.Write([]byte(`[{"id":"external-install","workspace":"plane-ws","workspace_detail":{"slug":"team"},"app_bot":"plane-bot","status":"installed"}]`))
		default:
			t.Errorf("unexpected Plane request %s", r.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	fx.apps.SetPlaneHTTPClient(planeapi.NewClient(srv.Client(), planeapi.Profile{OAuthURL: srv.URL}))
	state := beginPlaneTestInstall(t, fx, app)
	target := fx.apps.CompletePlaneInstall(t.Context(), state, "external-install", "")
	if !strings.HasPrefix(target, "https://dashboard.example/plane-apps/"+app.GetId()+"/edit?") ||
		!strings.Contains(target, "plane_install=success") || strings.Contains(target, "private-bot-token") {
		t.Fatalf("unexpected redirect: %s", target)
	}
	rows, err := fx.apps.installRepo.List(t.Context(), "ws-a", app.GetId())
	if err != nil || len(rows) != 1 || rows[0].GetPlaneWorkspaceId() != "plane-ws" ||
		rows[0].GetAppBotUserId() != "plane-bot" || strings.Join(rows[0].GetGrantedScopes(), " ") != "agents.runs:read agents.run_activities:read agents.run_activities:write" {
		t.Fatalf("installations=%v err=%v", rows, err)
	}
	if token, err := fx.apps.installRepo.GetRuntimeToken(t.Context(), "ws-a", app.GetId(), "external-install"); err != nil || token.AccessToken != "private-bot-token" {
		t.Fatalf("runtime token unavailable: %v", err)
	}
	repeated := fx.apps.CompletePlaneInstall(t.Context(), state, "external-install", "")
	if !strings.Contains(repeated, "reason=state_invalid") || calls.Load() != 2 {
		t.Fatalf("state replay: target=%s calls=%d", repeated, calls.Load())
	}
}

func TestPlaneInstallRejectsInvalidCallbacksBeforeNetwork(t *testing.T) {
	for _, tc := range []struct{ name, id, oauthErr, reason string }{
		{"missing installation", "", "", "installation_missing"},
		{"denied", "external", "access_denied", "denied"},
		{"expired", "external", "", "state_invalid"},
		{"permission revoked", "external", "", "permission_changed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx, app, installs := installFixture(t)
			state := beginPlaneTestInstall(t, fx, app)
			if tc.name == "expired" {
				fx.apps.SetPlaneClock(func() time.Time { return time.Now().Add(time.Hour) })
			}
			if tc.name == "permission revoked" {
				fx.apps.SetWorkspaceRepo(nil)
			}
			result := fx.apps.CompletePlaneInstall(t.Context(), state, tc.id, tc.oauthErr)
			if !strings.Contains(result, "reason="+tc.reason) {
				t.Fatalf("redirect=%s want reason %s", result, tc.reason)
			}
			rows, _ := installs.List(t.Context(), "ws-a", app.GetId())
			if len(rows) != 0 {
				t.Fatal("invalid callback created an installation")
			}
		})
	}
}

func TestPlaneBeginInstallRequiresPublicEndpointAndManageRole(t *testing.T) {
	fx, app, _ := installFixture(t)
	_, err := fx.apps.BeginPlaneInstall(planeMemberA(), connect.NewRequest(&agentsv1.BeginPlaneInstallRequest{AppId: app.GetId()}))
	wantPlaneCode(t, err, connect.CodePermissionDenied, "member installation")
	_, _ = fx.settings.Put(t.Context(), &agentsv1.PlaneSettings{})
	_, err = fx.apps.BeginPlaneInstall(planeOwnerA(), connect.NewRequest(&agentsv1.BeginPlaneInstallRequest{AppId: app.GetId()}))
	wantPlaneCode(t, err, connect.CodeFailedPrecondition, "missing public endpoint")
}

func TestPlaneInstallReturnIsBoundToDashboard(t *testing.T) {
	for _, raw := range []string{"https://evil.example/steal", "//evil.example/steal", "/\\evil.example", "/%2f%2fevil.example", "javascript:alert(1)"} {
		if got := sanitizePlaneReturn(raw, "https://dashboard.example", "app"); got != "https://dashboard.example/plane-apps/app/edit" {
			t.Fatalf("unsafe return %q => %q", raw, got)
		}
	}
}
