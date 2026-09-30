// Package linearconn resolves the credentials the Linear runtime acts with
// (ADR-0015). Nothing is cached: every call reads the installation and
// decrypts its token, so a reinstall or refresh takes effect on the next
// call.
package linearconn

import (
	"context"
	"errors"
	"fmt"

	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	// ErrNeedsReinstall means Linear revoked the installation's token or
	// refused to refresh it. Only installing the App again fixes it.
	ErrNeedsReinstall = errors.New("linear installation needs to be installed again")
	// ErrNoToken means the installation has no stored access token.
	ErrNoToken = errors.New("linear installation has no access token")
)

// TokenSource hands out access tokens for Linear Installations.
type TokenSource struct {
	repo    linearrepo.Repository
	keyring *secretbox.Keyring
}

func NewTokenSource(repo linearrepo.Repository, keyring *secretbox.Keyring) *TokenSource {
	return &TokenSource{repo: repo, keyring: keyring}
}

// AccessToken returns a usable access token for the installation.
func (s *TokenSource) AccessToken(ctx context.Context, workspaceID, installationID string) (string, error) {
	if s == nil || s.repo == nil || s.keyring == nil {
		return "", errors.New("linear token source is not configured")
	}
	inst, err := s.repo.GetInstallation(ctx, workspaceID, installationID)
	if err != nil {
		return "", err
	}
	if inst.GetCredentialState() == agentsv1.LinearInstallationCredentialState_LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL {
		return "", fmt.Errorf("%w: %s (organization %s)", ErrNeedsReinstall, installationID, inst.GetOrganizationName())
	}
	tokens, err := s.repo.GetInstallationTokens(ctx, workspaceID, installationID)
	if err != nil {
		return "", err
	}
	if !tokens.AccessToken.Set() {
		return "", fmt.Errorf("%w: %s", ErrNoToken, installationID)
	}
	plain, err := s.keyring.Decrypt(ctx, tokens.AccessToken.Ciphertext, tokens.AccessToken.KeyID)
	if err != nil {
		return "", fmt.Errorf("decrypt linear access token: %w", err)
	}
	return string(plain), nil
}
