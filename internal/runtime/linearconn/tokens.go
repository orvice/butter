// Package linearconn resolves the credentials the Linear runtime acts with
// (ADR-0015). Nothing is cached: every call reads the installation and
// decrypts its token, so a reinstall or refresh takes effect on the next
// call.
package linearconn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/linearapi"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// refreshSkew refreshes an access token this long before it expires.
	refreshSkew = 5 * time.Minute
	// refreshWait bounds how long a caller waits for another Pod's
	// refresh of the same installation.
	refreshWait = 10 * time.Second
	// refreshPoll is how often a waiting caller re-reads the tokens.
	refreshPoll = 50 * time.Millisecond
	// refreshLeasePrefix namespaces the per-installation refresh lease.
	refreshLeasePrefix = "linear-refresh:"
)

var (
	// ErrNeedsReinstall means Linear revoked the installation's token or
	// refused to refresh it. Only installing the App again fixes it.
	ErrNeedsReinstall = errors.New("linear installation needs to be installed again")
	// ErrNoToken means the installation has no stored access token.
	ErrNoToken = errors.New("linear installation has no access token")
)

// TokenSource hands out access tokens for Linear Installations, refreshing
// them shortly before they expire.
//
// Linear rotates refresh tokens, so two Pods refreshing one installation at
// once would leave it with a token only one of them stored — or none. A
// refresh therefore runs under a per-installation lease, and its write is
// fenced on the token revision it read: a caller that loses either race
// re-reads the stored token instead of refreshing again.
type TokenSource struct {
	repo    linearrepo.Repository
	keyring *secretbox.Keyring
	linear  *linearapi.Client
	guard   sessionguard.Guard
	clock   func() time.Time
}

func NewTokenSource(repo linearrepo.Repository, keyring *secretbox.Keyring) *TokenSource {
	return &TokenSource{repo: repo, keyring: keyring}
}

// SetLinearClient sets the client refreshes go through.
func (s *TokenSource) SetLinearClient(client *linearapi.Client) { s.linear = client }

// SetRefreshGuard sets the lease that serializes refreshes of one
// installation across Pods.
func (s *TokenSource) SetRefreshGuard(guard sessionguard.Guard) { s.guard = guard }

// SetClock overrides the clock. Used by tests.
func (s *TokenSource) SetClock(now func() time.Time) { s.clock = now }

func (s *TokenSource) now() time.Time {
	if s.clock != nil {
		return s.clock()
	}
	return time.Now()
}

// AccessToken returns a usable access token for the installation.
func (s *TokenSource) AccessToken(ctx context.Context, workspaceID, installationID string) (string, error) {
	if s == nil || s.repo == nil || s.keyring == nil {
		return "", errors.New("linear token source is not configured")
	}
	inst, tokens, err := s.read(ctx, workspaceID, installationID)
	if err != nil {
		return "", err
	}
	if !s.needsRefresh(tokens) || s.linear == nil {
		return s.decrypt(ctx, tokens.AccessToken)
	}
	return s.refresh(ctx, workspaceID, inst, tokens)
}

// MarkRejected records that Linear rejected the installation's token on
// use: it was revoked or the app was uninstalled.
func (s *TokenSource) MarkRejected(ctx context.Context, workspaceID, installationID string, cause error) {
	reason := "Linear rejected the access token; install the app again"
	if err := s.repo.MarkInstallationNeedsReinstall(ctx, workspaceID, installationID, reason); err != nil {
		log.FromContext(ctx).Warn("could not mark linear installation for reinstall",
			"installation_id", installationID, "err", err)
	}
	log.FromContext(ctx).Warn("linear installation needs reinstall",
		"workspace_id", workspaceID, "installation_id", installationID, "cause", cause)
}

// read loads the installation and its tokens, failing fast on a marked one.
func (s *TokenSource) read(ctx context.Context, workspaceID, installationID string) (*agentsv1.LinearInstallation, linearrepo.InstallationTokens, error) {
	inst, err := s.repo.GetInstallation(ctx, workspaceID, installationID)
	if err != nil {
		return nil, linearrepo.InstallationTokens{}, err
	}
	if inst.GetCredentialState() == agentsv1.LinearInstallationCredentialState_LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL {
		return nil, linearrepo.InstallationTokens{}, s.needsReinstall(ctx, workspaceID, inst)
	}
	tokens, err := s.repo.GetInstallationTokens(ctx, workspaceID, installationID)
	if err != nil {
		return nil, linearrepo.InstallationTokens{}, err
	}
	if !tokens.AccessToken.Set() {
		return nil, linearrepo.InstallationTokens{}, fmt.Errorf("%w: %s", ErrNoToken, installationID)
	}
	return inst, tokens, nil
}

// needsReinstall is the actionable error naming the App and organization.
func (s *TokenSource) needsReinstall(ctx context.Context, workspaceID string, inst *agentsv1.LinearInstallation) error {
	appName := inst.GetAppId()
	if app, err := s.repo.GetApp(ctx, workspaceID, inst.GetAppId()); err == nil && app.GetDisplayName() != "" {
		appName = app.GetDisplayName()
	}
	org := inst.GetOrganizationName()
	if org == "" {
		org = inst.GetOrganizationId()
	}
	return fmt.Errorf("%w: install Linear App %q in organization %q again", ErrNeedsReinstall, appName, org)
}

func (s *TokenSource) needsRefresh(tokens linearrepo.InstallationTokens) bool {
	if tokens.ExpiresAt.IsZero() || !tokens.RefreshToken.Set() {
		return false
	}
	return !s.now().Add(refreshSkew).Before(tokens.ExpiresAt)
}

func (s *TokenSource) expired(tokens linearrepo.InstallationTokens) bool {
	return !tokens.ExpiresAt.IsZero() && !s.now().Before(tokens.ExpiresAt)
}

func (s *TokenSource) refresh(ctx context.Context, workspaceID string, inst *agentsv1.LinearInstallation, seen linearrepo.InstallationTokens) (string, error) {
	if s.guard == nil {
		return s.refreshHeld(ctx, workspaceID, inst, seen)
	}
	deadline := time.Now().Add(refreshWait)
	for {
		_, release, acquired, err := s.guard.Acquire(ctx, refreshLeasePrefix+inst.GetId())
		if err != nil {
			return "", fmt.Errorf("acquire linear refresh lease: %w", err)
		}
		if acquired {
			defer release()
			return s.refreshHeld(ctx, workspaceID, inst, seen)
		}
		// Another caller is refreshing: wait for its write rather than
		// spending the rotated refresh token a second time.
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(refreshPoll):
		}
		_, current, err := s.read(ctx, workspaceID, inst.GetId())
		if err != nil {
			return "", err
		}
		if current.Revision != seen.Revision && !s.needsRefresh(current) {
			return s.decrypt(ctx, current.AccessToken)
		}
		if time.Now().After(deadline) && !s.expired(current) {
			return s.decrypt(ctx, current.AccessToken)
		}
	}
}

// refreshHeld refreshes while holding the lease.
func (s *TokenSource) refreshHeld(ctx context.Context, workspaceID string, inst *agentsv1.LinearInstallation, seen linearrepo.InstallationTokens) (string, error) {
	logger := log.FromContext(ctx)
	// Someone may have refreshed between our read and the lease.
	_, tokens, err := s.read(ctx, workspaceID, inst.GetId())
	if err != nil {
		return "", err
	}
	if tokens.Revision != seen.Revision && !s.needsRefresh(tokens) {
		return s.decrypt(ctx, tokens.AccessToken)
	}

	app, err := s.repo.GetApp(ctx, workspaceID, inst.GetAppId())
	if err != nil {
		return "", err
	}
	creds, err := s.repo.GetAppCredentials(ctx, workspaceID, inst.GetAppId())
	if err != nil {
		return "", err
	}
	if !creds.ClientSecret.Set() {
		return s.keepOrFail(ctx, tokens, errors.New("the Linear App has no client secret to refresh with"))
	}
	clientSecret, err := s.decrypt(ctx, creds.ClientSecret)
	if err != nil {
		return "", err
	}
	refreshToken, err := s.decrypt(ctx, tokens.RefreshToken)
	if err != nil {
		return "", err
	}

	fresh, err := s.linear.RefreshToken(ctx, app.GetClientId(), clientSecret, refreshToken)
	if err != nil {
		if errors.Is(err, linearapi.ErrInvalidGrant) || errors.Is(err, linearapi.ErrUnauthorized) {
			if markErr := s.repo.MarkInstallationNeedsReinstall(ctx, workspaceID, inst.GetId(),
				"Linear refused to refresh the access token; install the app again"); markErr != nil {
				logger.Warn("could not mark linear installation for reinstall", "installation_id", inst.GetId(), "err", markErr)
			}
			logger.Warn("linear refresh refused", "workspace_id", workspaceID, "installation_id", inst.GetId(), "err", err)
			return "", s.needsReinstall(ctx, workspaceID, inst)
		}
		return s.keepOrFail(ctx, tokens, fmt.Errorf("refresh linear token: %w", err))
	}

	next := linearrepo.InstallationTokens{ExpiresAt: fresh.ExpiresAt(s.now())}
	if next.AccessToken, err = s.encrypt(ctx, fresh.AccessToken); err != nil {
		return "", err
	}
	next.RefreshToken = tokens.RefreshToken
	if fresh.RefreshToken != "" {
		if next.RefreshToken, err = s.encrypt(ctx, fresh.RefreshToken); err != nil {
			return "", err
		}
	}
	if _, err := s.repo.ReplaceInstallationTokens(ctx, workspaceID, inst.GetId(), tokens.Revision, next); err != nil {
		if errors.Is(err, linearrepo.ErrRevisionConflict) {
			// Another writer (a reinstall) stored newer tokens: theirs win.
			_, current, readErr := s.read(ctx, workspaceID, inst.GetId())
			if readErr != nil {
				return "", readErr
			}
			return s.decrypt(ctx, current.AccessToken)
		}
		return "", err
	}
	logger.Info("linear access token refreshed", "workspace_id", workspaceID, "installation_id", inst.GetId())
	return fresh.AccessToken, nil
}

// keepOrFail answers a refresh that could not happen: the current token is
// still fine until it actually expires.
func (s *TokenSource) keepOrFail(ctx context.Context, tokens linearrepo.InstallationTokens, cause error) (string, error) {
	if s.expired(tokens) {
		return "", cause
	}
	log.FromContext(ctx).Warn("linear token refresh deferred; using the current token", "err", cause)
	return s.decrypt(ctx, tokens.AccessToken)
}

func (s *TokenSource) decrypt(ctx context.Context, cred linearrepo.Credential) (string, error) {
	plain, err := s.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
	if err != nil {
		return "", fmt.Errorf("decrypt linear credential: %w", err)
	}
	return string(plain), nil
}

func (s *TokenSource) encrypt(ctx context.Context, value string) (linearrepo.Credential, error) {
	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(value))
	if err != nil {
		return linearrepo.Credential{}, fmt.Errorf("encrypt linear token: %w", err)
	}
	return linearrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}, nil
}
