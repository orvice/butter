package application

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/auth/provider"
	oauthhandoffmemory "go.orx.me/apps/butter/internal/repo/oauthhandoff/memory"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// RFC 7636 Appendix B: an independent, published verifier/challenge pair.
const (
	rfcVerifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	rfcChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	appReturnURI = "ee.orz.butter:/oauth/callback"
)

// nativeFixture serves the production AuthService Connect handler over HTTP,
// backed by in-memory repositories and a stubbed provider, and exposes a
// controllable clock.
type nativeFixture struct {
	client agentsv1connect.AuthServiceClient
	repo   *fakeAuthRepo
	stub   *stubProvider
	now    time.Time
}

func newNativeFixture(t *testing.T, returnURIs ...string) *nativeFixture {
	t.Helper()
	stub := &stubProvider{
		name:        "github",
		displayName: "GitHub",
		authorize:   "https://provider.test/authorize",
		claims: &provider.Claims{
			Provider: "github", ExternalID: "42", Login: "alice", Name: "Alice",
		},
	}
	srv, repo := newServerWithOAuth(t, stub)
	srv.SetOAuthHandoffRepo(oauthhandoffmemory.New())
	srv.SetNativeReturnURIs(returnURIs)
	f := &nativeFixture{repo: repo, stub: stub, now: time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC)}
	srv.now = func() time.Time { return f.now }

	_, handler := agentsv1connect.NewAuthServiceHandler(srv, connectx.HandlerOptions()...)
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	f.client = agentsv1connect.NewAuthServiceClient(ts.Client(), ts.URL)
	return f
}

func (f *nativeFixture) begin(t *testing.T, native *agentsv1.NativeOAuthClient) (*agentsv1.BeginOAuthFlowResponse, error) {
	t.Helper()
	res, err := f.client.BeginOAuthFlow(context.Background(), connect.NewRequest(&agentsv1.BeginOAuthFlowRequest{
		Provider: "github", Native: native,
	}))
	if err != nil {
		return nil, err
	}
	return res.Msg, nil
}

func (f *nativeFixture) complete(t *testing.T, state string) (*agentsv1.CompleteOAuthFlowResponse, error) {
	t.Helper()
	res, err := f.client.CompleteOAuthFlow(context.Background(), connect.NewRequest(&agentsv1.CompleteOAuthFlowRequest{
		Provider: "github", Code: "provider-code", State: state,
	}))
	if err != nil {
		return nil, err
	}
	return res.Msg, nil
}

func (f *nativeFixture) redeem(t *testing.T, provider, state, handoff, verifier string) (*agentsv1.RedeemOAuthHandoffResponse, error) {
	t.Helper()
	res, err := f.client.RedeemOAuthHandoff(context.Background(), connect.NewRequest(&agentsv1.RedeemOAuthHandoffRequest{
		Provider: provider, State: state, HandoffCode: handoff, CodeVerifier: verifier,
	}))
	if err != nil {
		return nil, err
	}
	return res.Msg, nil
}

func nativeClient() *agentsv1.NativeOAuthClient {
	return &agentsv1.NativeOAuthClient{ReturnUri: appReturnURI, CodeChallenge: rfcChallenge}
}

// beginAndComplete runs the browser leg of a native flow and returns the
// state and the handoff code the App would receive in its return link.
func (f *nativeFixture) beginAndComplete(t *testing.T) (state, handoff string) {
	t.Helper()
	b, err := f.begin(t, nativeClient())
	if err != nil {
		t.Fatalf("BeginOAuthFlow: %v", err)
	}
	done, err := f.complete(t, b.GetState())
	if err != nil {
		t.Fatalf("CompleteOAuthFlow: %v", err)
	}
	link, err := url.Parse(done.GetNativeHandoff().GetReturnUrl())
	if err != nil {
		t.Fatalf("return_url: %v", err)
	}
	return b.GetState(), link.Query().Get("handoff")
}

func requireCode(t *testing.T, err error, want connect.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %v, got success", want)
	}
	if got := connect.CodeOf(err); got != want {
		t.Fatalf("code = %v (%v), want %v", got, err, want)
	}
}

func TestNativeOAuth_BrowserReturnsHandoffNotToken_AppRedeemsSession(t *testing.T) {
	f := newNativeFixture(t, appReturnURI)

	b, err := f.begin(t, nativeClient())
	if err != nil {
		t.Fatalf("BeginOAuthFlow: %v", err)
	}
	if !strings.Contains(b.GetAuthorizeUrl(), b.GetState()) {
		t.Errorf("authorize_url must carry the state: %s", b.GetAuthorizeUrl())
	}

	// The browser leg: the fixed web callback completes the flow.
	done, err := f.complete(t, b.GetState())
	if err != nil {
		t.Fatalf("CompleteOAuthFlow: %v", err)
	}
	if f.stub.gotCode != "provider-code" {
		t.Errorf("provider got code %q, want provider-code", f.stub.gotCode)
	}
	if done.GetToken() != "" || done.GetUser() != nil || len(done.GetWorkspaces()) != 0 {
		t.Fatalf("browser leg must not receive a session: %+v", done)
	}
	link, err := url.Parse(done.GetNativeHandoff().GetReturnUrl())
	if err != nil {
		t.Fatalf("return_url: %v", err)
	}
	if link.Scheme != "ee.orz.butter" || link.Path != "/oauth/callback" {
		t.Errorf("return_url = %q, want the allow-listed URI", link)
	}
	if link.Query().Get("state") != b.GetState() {
		t.Errorf("return_url state = %q, want %q", link.Query().Get("state"), b.GetState())
	}
	handoff := link.Query().Get("handoff")
	if handoff == "" {
		t.Fatal("return_url carries no handoff code")
	}
	if len(f.repo.sessions) != 0 {
		t.Fatal("no session may exist before the App redeems the handoff")
	}

	// The App leg: PKCE verifier + handoff yields the session.
	got, err := f.redeem(t, "github", b.GetState(), handoff, rfcVerifier)
	if err != nil {
		t.Fatalf("RedeemOAuthHandoff: %v", err)
	}
	if got.GetToken() == "" {
		t.Error("redeem returned no token")
	}
	if got.GetUser().GetExternalId() != "42" || got.GetUser().GetProvider() != "github" {
		t.Errorf("unexpected user: %+v", got.GetUser())
	}
	if got.GetExpiresAt() == nil || !got.GetExpiresAt().AsTime().Equal(f.now.Add(time.Hour)) {
		t.Errorf("expires_at = %v, want one session TTL after redemption", got.GetExpiresAt())
	}
	if len(f.repo.sessions) != 1 {
		t.Errorf("sessions = %d, want 1", len(f.repo.sessions))
	}
}

func TestNativeOAuth_BeginHonorsOnlyTheServerAllowList(t *testing.T) {
	cases := []struct {
		name   string
		uris   []string
		native *agentsv1.NativeOAuthClient
		want   connect.Code
	}{
		{"disabled by empty allow-list", nil, nativeClient(), connect.CodeFailedPrecondition},
		{"an allow-list entry with a query is never trusted", []string{appReturnURI + "?x=1"},
			&agentsv1.NativeOAuthClient{ReturnUri: appReturnURI + "?x=1", CodeChallenge: rfcChallenge}, connect.CodeFailedPrecondition},
		{"an entry that is not an absolute URI is never trusted", []string{"/oauth/callback", "https:///cb"},
			&agentsv1.NativeOAuthClient{ReturnUri: "/oauth/callback", CodeChallenge: rfcChallenge}, connect.CodeFailedPrecondition},
		{"unlisted scheme", []string{appReturnURI},
			&agentsv1.NativeOAuthClient{ReturnUri: "evil.app:/oauth/callback", CodeChallenge: rfcChallenge}, connect.CodeInvalidArgument},
		{"prefix is not a match", []string{appReturnURI},
			&agentsv1.NativeOAuthClient{ReturnUri: appReturnURI + "/x", CodeChallenge: rfcChallenge}, connect.CodeInvalidArgument},
		{"query is not a match", []string{appReturnURI},
			&agentsv1.NativeOAuthClient{ReturnUri: appReturnURI + "?a=b", CodeChallenge: rfcChallenge}, connect.CodeInvalidArgument},
		{"https web origin is not allowed unless listed", []string{appReturnURI},
			&agentsv1.NativeOAuthClient{ReturnUri: "https://evil.example/cb", CodeChallenge: rfcChallenge}, connect.CodeInvalidArgument},
		{"challenge must be a 43 char base64url S256", []string{appReturnURI},
			&agentsv1.NativeOAuthClient{ReturnUri: appReturnURI, CodeChallenge: "short"}, connect.CodeInvalidArgument},
		{"challenge charset", []string{appReturnURI},
			&agentsv1.NativeOAuthClient{ReturnUri: appReturnURI, CodeChallenge: strings.Repeat("+", 43)}, connect.CodeInvalidArgument},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newNativeFixture(t, tc.uris...)
			_, err := f.begin(t, tc.native)
			requireCode(t, err, tc.want)
		})
	}
}

func TestNativeOAuth_WebFlowIsUnchanged(t *testing.T) {
	// A server with native sign-in enabled still logs a web state in directly.
	f := newNativeFixture(t, appReturnURI)
	b, err := f.begin(t, nil)
	if err != nil {
		t.Fatalf("BeginOAuthFlow: %v", err)
	}
	done, err := f.complete(t, b.GetState())
	if err != nil {
		t.Fatalf("CompleteOAuthFlow: %v", err)
	}
	if done.GetToken() == "" || done.GetNativeHandoff() != nil {
		t.Fatalf("web flow must receive its session and no handoff: %+v", done)
	}
}

func TestNativeOAuth_RedeemRequiresTheAppsVerifier(t *testing.T) {
	f := newNativeFixture(t, appReturnURI)
	state, handoff := f.beginAndComplete(t)

	// An app that merely intercepted the link has no verifier.
	_, err := f.redeem(t, "github", state, handoff, strings.Repeat("a", 43))
	requireCode(t, err, connect.CodePermissionDenied)

	// A wrong guess spends the handoff: the real verifier is too late.
	_, err = f.redeem(t, "github", state, handoff, rfcVerifier)
	requireCode(t, err, connect.CodePermissionDenied)
	if len(f.repo.sessions) != 0 {
		t.Fatalf("no session may be minted, got %d", len(f.repo.sessions))
	}
}

func TestNativeOAuth_HandoffIsSingleUse(t *testing.T) {
	f := newNativeFixture(t, appReturnURI)
	state, handoff := f.beginAndComplete(t)
	if _, err := f.redeem(t, "github", state, handoff, rfcVerifier); err != nil {
		t.Fatalf("first redeem: %v", err)
	}
	_, err := f.redeem(t, "github", state, handoff, rfcVerifier)
	requireCode(t, err, connect.CodePermissionDenied)
	if len(f.repo.sessions) != 1 {
		t.Errorf("replay must not mint a second session, got %d", len(f.repo.sessions))
	}
}

func TestNativeOAuth_RedeemIsBoundToProviderAndState(t *testing.T) {
	t.Run("another state", func(t *testing.T) {
		f := newNativeFixture(t, appReturnURI)
		_, handoff := f.beginAndComplete(t)
		_, err := f.redeem(t, "github", "some-other-state", handoff, rfcVerifier)
		requireCode(t, err, connect.CodePermissionDenied)
	})
	t.Run("another provider", func(t *testing.T) {
		f := newNativeFixture(t, appReturnURI)
		state, handoff := f.beginAndComplete(t)
		_, err := f.redeem(t, "google", state, handoff, rfcVerifier)
		requireCode(t, err, connect.CodePermissionDenied)
	})
	t.Run("the oauth state is not a handoff code", func(t *testing.T) {
		f := newNativeFixture(t, appReturnURI)
		state, _ := f.beginAndComplete(t)
		_, err := f.redeem(t, "github", state, state, rfcVerifier)
		requireCode(t, err, connect.CodePermissionDenied)
	})
}

func TestNativeOAuth_StateCannotBeCompletedTwice(t *testing.T) {
	f := newNativeFixture(t, appReturnURI)
	b, err := f.begin(t, nativeClient())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.complete(t, b.GetState()); err != nil {
		t.Fatal(err)
	}
	_, err = f.complete(t, b.GetState())
	requireCode(t, err, connect.CodePermissionDenied)
}

func TestNativeOAuth_ExpiryOfStateAndHandoff(t *testing.T) {
	t.Run("abandoned browser leg", func(t *testing.T) {
		f := newNativeFixture(t, appReturnURI)
		b, err := f.begin(t, nativeClient())
		if err != nil {
			t.Fatal(err)
		}
		f.now = f.now.Add(11 * time.Minute)
		_, err = f.complete(t, b.GetState())
		requireCode(t, err, connect.CodePermissionDenied)
	})
	t.Run("handoff not redeemed in time", func(t *testing.T) {
		f := newNativeFixture(t, appReturnURI)
		state, handoff := f.beginAndComplete(t)
		f.now = f.now.Add(3 * time.Minute)
		_, err := f.redeem(t, "github", state, handoff, rfcVerifier)
		requireCode(t, err, connect.CodePermissionDenied)
	})
	t.Run("handoff redeemed just in time", func(t *testing.T) {
		f := newNativeFixture(t, appReturnURI)
		state, handoff := f.beginAndComplete(t)
		f.now = f.now.Add(90 * time.Second)
		if _, err := f.redeem(t, "github", state, handoff, rfcVerifier); err != nil {
			t.Fatalf("redeem within the TTL: %v", err)
		}
	})
}

func TestNativeOAuth_UserDisabledBeforeRedeemGetsNoSession(t *testing.T) {
	f := newNativeFixture(t, appReturnURI)
	state, handoff := f.beginAndComplete(t)
	for _, u := range f.repo.users {
		u.Disabled = true
	}
	_, err := f.redeem(t, "github", state, handoff, rfcVerifier)
	requireCode(t, err, connect.CodePermissionDenied)
	if len(f.repo.sessions) != 0 {
		t.Fatalf("a disabled user must not get a session, got %d", len(f.repo.sessions))
	}
}

func TestNativeOAuth_RedeemRejectsMalformedVerifier(t *testing.T) {
	f := newNativeFixture(t, appReturnURI)
	state, handoff := f.beginAndComplete(t)
	for _, v := range []string{"", "short", strings.Repeat("a", 129), strings.Repeat("+", 43)} {
		_, err := f.redeem(t, "github", state, handoff, v)
		requireCode(t, err, connect.CodeInvalidArgument)
	}
	// Malformed input is rejected before the handoff is spent.
	if _, err := f.redeem(t, "github", state, handoff, rfcVerifier); err != nil {
		t.Fatalf("handoff must survive malformed attempts: %v", err)
	}
}
