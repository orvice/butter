package linearconn

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/linearapi/lineartest"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	linearmemory "go.orx.me/apps/butter/internal/repo/linear/memory"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestAccessTokenIsTheInstallationsDecryptedToken(t *testing.T) {
	ctx := t.Context()
	repo := linearmemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	if _, err := repo.CreateApp(ctx, "ws-a", &agentsv1.LinearApp{Id: "app-1", ClientId: "c", AgentId: "a"}, linearrepo.AppCredentials{}); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	ciphertext, keyID, err := keyring.Encrypt(ctx, []byte("access-1"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := repo.UpsertInstallation(ctx, "ws-a", &agentsv1.LinearInstallation{Id: "inst-1", AppId: "app-1", OrganizationId: "org-1"},
		linearrepo.InstallationTokens{AccessToken: linearrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}
	if _, err := repo.UpsertInstallation(ctx, "ws-a", &agentsv1.LinearInstallation{Id: "inst-2", AppId: "app-1", OrganizationId: "org-2"},
		linearrepo.InstallationTokens{}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	source := NewTokenSource(repo, keyring)
	token, err := source.AccessToken(ctx, "ws-a", "inst-1")
	if err != nil || token != "access-1" {
		t.Fatalf("AccessToken = %q, %v; want the decrypted token", token, err)
	}
	if _, err := source.AccessToken(ctx, "ws-a", "inst-2"); !errors.Is(err, ErrNoToken) {
		t.Fatalf("AccessToken without a token = %v, want ErrNoToken", err)
	}
	if _, err := source.AccessToken(ctx, "ws-b", "inst-1"); !errors.Is(err, linearrepo.ErrNotFound) {
		t.Fatalf("cross-workspace AccessToken = %v, want ErrNotFound", err)
	}
}

// refreshFixture is an installation whose token is about to expire, a fake
// Linear that can refresh it, and a token source over both.
type refreshFixture struct {
	repo    *linearmemory.Store
	keyring *secretbox.Keyring
	linear  *lineartest.Fake
	source  *TokenSource
	now     time.Time
}

func newRefreshFixture(t *testing.T, expiresIn time.Duration) *refreshFixture {
	t.Helper()
	ctx := t.Context()
	fx := &refreshFixture{
		repo:    linearmemory.New(),
		keyring: secretbox.NewKeyring(cryptokeymemory.New()),
		linear:  lineartest.New(t),
		now:     time.Now(),
	}
	secret := fx.encrypt(t, "client-secret")
	if _, err := fx.repo.CreateApp(ctx, "ws-a", &agentsv1.LinearApp{Id: "app-1", ClientId: "client-1", AgentId: "a", DisplayName: "Support"},
		linearrepo.AppCredentials{ClientSecret: secret}); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	if _, err := fx.repo.UpsertInstallation(ctx, "ws-a",
		&agentsv1.LinearInstallation{Id: "inst-1", AppId: "app-1", OrganizationId: "org-1", OrganizationName: "Acme"},
		linearrepo.InstallationTokens{
			AccessToken:  fx.encrypt(t, "access-1"),
			RefreshToken: fx.encrypt(t, "refresh-1"),
			ExpiresAt:    fx.now.Add(expiresIn),
		}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}
	fx.source = NewTokenSource(fx.repo, fx.keyring)
	fx.source.SetLinearClient(fx.linear.Client())
	fx.source.SetRefreshGuard(sessionguard.NewMemory())
	fx.source.SetClock(func() time.Time { return fx.now })
	return fx
}

func (fx *refreshFixture) encrypt(t *testing.T, value string) linearrepo.Credential {
	t.Helper()
	ciphertext, keyID, err := fx.keyring.Encrypt(t.Context(), []byte(value))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return linearrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}
}

func (fx *refreshFixture) refreshes() int {
	n := 0
	for _, form := range fx.linear.TokenRequests() {
		if form.Get("grant_type") == "refresh_token" {
			n++
		}
	}
	return n
}

func (fx *refreshFixture) storedAccess(t *testing.T) string {
	t.Helper()
	tokens, err := fx.repo.GetInstallationTokens(t.Context(), "ws-a", "inst-1")
	if err != nil {
		t.Fatalf("GetInstallationTokens: %v", err)
	}
	plain, err := fx.keyring.Decrypt(t.Context(), tokens.AccessToken.Ciphertext, tokens.AccessToken.KeyID)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	return string(plain)
}

func TestAFreshTokenIsReturnedWithoutRefreshing(t *testing.T) {
	fx := newRefreshFixture(t, time.Hour)
	token, err := fx.source.AccessToken(t.Context(), "ws-a", "inst-1")
	if err != nil || token != "access-1" {
		t.Fatalf("AccessToken = %q, %v", token, err)
	}
	if fx.refreshes() != 0 {
		t.Fatalf("refreshes = %d, want 0", fx.refreshes())
	}
}

func TestANearExpiryTokenIsRefreshedOnceAndTheRotatedTokensAreStored(t *testing.T) {
	fx := newRefreshFixture(t, 2*time.Minute)
	fx.linear.AddRefresh("refresh-1", lineartest.Grant{AccessToken: "access-2", RefreshToken: "refresh-2", ExpiresIn: 86400})

	token, err := fx.source.AccessToken(t.Context(), "ws-a", "inst-1")
	if err != nil || token != "access-2" {
		t.Fatalf("AccessToken = %q, %v; want the refreshed token", token, err)
	}
	form := fx.linear.TokenRequests()[0]
	if form.Get("client_id") != "client-1" || form.Get("client_secret") != "client-secret" || form.Get("refresh_token") != "refresh-1" {
		t.Fatalf("refresh request = %v", form)
	}
	if fx.storedAccess(t) != "access-2" {
		t.Fatal("the refreshed access token was not stored")
	}
	tokens, _ := fx.repo.GetInstallationTokens(t.Context(), "ws-a", "inst-1")
	refresh, _ := fx.keyring.Decrypt(t.Context(), tokens.RefreshToken.Ciphertext, tokens.RefreshToken.KeyID)
	if string(refresh) != "refresh-2" {
		t.Fatalf("stored refresh token = %q, want the rotated one", refresh)
	}
	if again, err := fx.source.AccessToken(t.Context(), "ws-a", "inst-1"); err != nil || again != "access-2" || fx.refreshes() != 1 {
		t.Fatalf("second AccessToken = %q, %v with %d refreshes; want the stored token and no second refresh", again, err, fx.refreshes())
	}
}

func TestConcurrentRefreshesExchangeOnce(t *testing.T) {
	fx := newRefreshFixture(t, time.Minute)
	fx.linear.AddRefresh("refresh-1", lineartest.Grant{AccessToken: "access-2", RefreshToken: "refresh-2", ExpiresIn: 86400})

	var wg sync.WaitGroup
	results := make([]string, 4)
	errs := make([]error, 4)
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = fx.source.AccessToken(t.Context(), "ws-a", "inst-1")
		}()
	}
	wg.Wait()
	for i := range results {
		if errs[i] != nil || results[i] != "access-2" {
			t.Fatalf("caller %d = %q, %v; want the refreshed token", i, results[i], errs[i])
		}
	}
	if fx.refreshes() != 1 {
		t.Fatalf("refreshes = %d, want exactly 1", fx.refreshes())
	}
}

// conflictingRepo lets another writer replace the tokens between the token
// source's read and its write, as a Pod without the lease (a reinstall)
// would.
type conflictingRepo struct {
	*linearmemory.Store
	interfere func()
}

func (r *conflictingRepo) ReplaceInstallationTokens(ctx context.Context, workspaceID, id string, expected int64, tokens linearrepo.InstallationTokens) (int64, error) {
	if r.interfere != nil {
		r.interfere()
		r.interfere = nil
	}
	return r.Store.ReplaceInstallationTokens(ctx, workspaceID, id, expected, tokens)
}

func TestALostCompareAndSwapReReadsTheStoredToken(t *testing.T) {
	fx := newRefreshFixture(t, time.Minute)
	fx.linear.AddRefresh("refresh-1", lineartest.Grant{AccessToken: "access-2", RefreshToken: "refresh-2", ExpiresIn: 86400})
	repo := &conflictingRepo{Store: fx.repo}
	repo.interfere = func() {
		if _, err := fx.repo.UpsertInstallation(t.Context(), "ws-a",
			&agentsv1.LinearInstallation{Id: "x", AppId: "app-1", OrganizationId: "org-1", OrganizationName: "Acme"},
			linearrepo.InstallationTokens{AccessToken: fx.encrypt(t, "access-reinstalled"), ExpiresAt: fx.now.Add(24 * time.Hour)}); err != nil {
			t.Errorf("interfering reinstall: %v", err)
		}
	}
	source := NewTokenSource(repo, fx.keyring)
	source.SetLinearClient(fx.linear.Client())
	source.SetRefreshGuard(sessionguard.NewMemory())
	source.SetClock(func() time.Time { return fx.now })

	token, err := source.AccessToken(t.Context(), "ws-a", "inst-1")
	if err != nil || token != "access-reinstalled" {
		t.Fatalf("AccessToken = %q, %v; want the token the other writer stored", token, err)
	}
	if fx.storedAccess(t) != "access-reinstalled" {
		t.Fatal("the losing refresh overwrote the other writer's token")
	}
}

func TestARefusedRefreshMarksTheInstallationForReinstallAndFailsFast(t *testing.T) {
	fx := newRefreshFixture(t, time.Minute)
	fx.linear.FailNextToken(lineartest.Failure{Status: http.StatusBadRequest, OAuthError: "invalid_grant", Message: "refresh token revoked"})

	_, err := fx.source.AccessToken(t.Context(), "ws-a", "inst-1")
	if !errors.Is(err, ErrNeedsReinstall) {
		t.Fatalf("AccessToken = %v, want ErrNeedsReinstall", err)
	}
	if !strings.Contains(err.Error(), "Support") || !strings.Contains(err.Error(), "Acme") {
		t.Errorf("error %q should name the App and the organization", err)
	}
	inst, _ := fx.repo.GetInstallation(t.Context(), "ws-a", "inst-1")
	if inst.GetCredentialState() != agentsv1.LinearInstallationCredentialState_LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL {
		t.Fatalf("credential_state = %v, want NEEDS_REINSTALL", inst.GetCredentialState())
	}
	requests := len(fx.linear.TokenRequests())
	if _, err := fx.source.AccessToken(t.Context(), "ws-a", "inst-1"); !errors.Is(err, ErrNeedsReinstall) {
		t.Fatalf("later AccessToken = %v, want ErrNeedsReinstall", err)
	}
	if len(fx.linear.TokenRequests()) != requests {
		t.Fatal("a marked installation contacted Linear again")
	}
}

func TestATransientRefreshFailureKeepsAStillValidToken(t *testing.T) {
	fx := newRefreshFixture(t, 2*time.Minute)
	fx.linear.FailNextToken(lineartest.Failure{Status: http.StatusBadGateway})
	token, err := fx.source.AccessToken(t.Context(), "ws-a", "inst-1")
	if err != nil || token != "access-1" {
		t.Fatalf("AccessToken = %q, %v; want the current token while it is still valid", token, err)
	}
	inst, _ := fx.repo.GetInstallation(t.Context(), "ws-a", "inst-1")
	if inst.GetCredentialState() != agentsv1.LinearInstallationCredentialState_LINEAR_INSTALLATION_CREDENTIAL_STATE_VALID {
		t.Fatal("a transient failure marked the installation for reinstall")
	}
}
