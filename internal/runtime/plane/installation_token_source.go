package plane

import (
	"context"
	"errors"
	"go.orx.me/apps/butter/internal/planeapi"
	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	installationrepo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	"go.orx.me/apps/butter/internal/secretbox"
	"sync"
	"time"
)

type TokenLease interface {
	Acquire(context.Context, string) (bool, error)
	Release(context.Context, string) error
}
type InstallationTokenSource struct {
	Apps          planerepo.Repository
	Installations installationrepo.Repository
	Keyring       *secretbox.Keyring
	Client        *planeapi.Client
	Lease         TokenLease
	FencedLease   FencedTokenLease
	Owner         string
	RefreshBefore time.Duration
	mu            sync.Mutex
}

var ErrTokenUnavailable = errors.New("plane installation token unavailable")

func (s *InstallationTokenSource) Token(ctx context.Context, workspaceID, appID, installationID string) (string, error) {
	return s.token(ctx, workspaceID, appID, installationID, false)
}
func (s *InstallationTokenSource) ForceRefresh(ctx context.Context, workspaceID, appID, installationID string) (string, error) {
	return s.token(ctx, workspaceID, appID, installationID, true)
}
func (s *InstallationTokenSource) token(ctx context.Context, workspaceID, appID, installationID string, force bool) (string, error) {
	if s == nil || s.Apps == nil || s.Installations == nil || s.Keyring == nil || s.Client == nil {
		return "", ErrTokenUnavailable
	}
	refreshBefore := s.RefreshBefore
	if refreshBefore <= 0 {
		refreshBefore = 5 * time.Minute
	}
	for attempt := 0; attempt < 2; attempt++ {
		m, err := s.Installations.GetRuntimeToken(ctx, workspaceID, appID, installationID)
		if err != nil {
			return "", ErrTokenUnavailable
		}
		installation, err := s.Installations.GetScoped(ctx, workspaceID, appID, installationID)
		if err != nil || installation.GetStatus() != "installed" || installation.GetCredentialState() != "valid" {
			return "", ErrTokenUnavailable
		}
		if !force && m.AccessToken != "" && time.Until(m.ExpiresAt) > refreshBefore {
			return m.AccessToken, nil
		}
		s.mu.Lock()
		var release func()
		if s.Lease != nil || s.FencedLease != nil {
			key := workspaceID + ":" + appID + ":" + installationID
			var ok bool
			var fence uint64
			var acquireErr error
			if s.FencedLease != nil {
				if s.Owner == "" {
					s.mu.Unlock()
					return "", ErrTokenUnavailable
				}
				fence, acquireErr = s.FencedLease.AcquireToken(ctx, key, s.Owner, m.Revision, time.Now().Add(5*time.Minute))
				ok = acquireErr == nil && fence != 0
			} else {
				ok, acquireErr = s.Lease.Acquire(ctx, key)
			}
			if acquireErr != nil || !ok {
				s.mu.Unlock()
				if acquireErr != nil {
					return "", ErrTokenUnavailable
				}
				continue
			}
			release = func() {
				if s.FencedLease != nil {
					_ = s.FencedLease.ReleaseToken(context.Background(), key, s.Owner, fence)
				} else {
					_ = s.Lease.Release(context.Background(), key)
				}
			}
			defer release()
		}
		latest, e := s.Installations.GetRuntimeToken(ctx, workspaceID, appID, installationID)
		if e != nil {
			s.mu.Unlock()
			return "", ErrTokenUnavailable
		}
		if !force && latest.Revision != m.Revision && latest.AccessToken != "" && time.Until(latest.ExpiresAt) > refreshBefore {
			s.mu.Unlock()
			return latest.AccessToken, nil
		}
		app, e := s.Apps.GetApp(ctx, workspaceID, appID)
		if e != nil {
			s.mu.Unlock()
			return "", ErrTokenUnavailable
		}
		creds, e := s.Apps.GetAppCredentials(ctx, workspaceID, appID)
		if e != nil || !creds.ClientSecret.Set() {
			s.mu.Unlock()
			return "", ErrTokenUnavailable
		}
		secret, e := s.Keyring.Decrypt(ctx, creds.ClientSecret.Ciphertext, creds.ClientSecret.KeyID)
		if e != nil {
			s.mu.Unlock()
			return "", ErrTokenUnavailable
		}
		tok, e := s.Client.BotToken(ctx, app.GetClientId(), string(secret), installation.GetAppInstallationId())
		for i := range secret {
			secret[i] = 0
		}
		if e != nil {
			s.mu.Unlock()
			return "", ErrTokenUnavailable
		}
		saved, e := s.Installations.CASRuntimeToken(ctx, workspaceID, appID, installationID, m.Revision, installationrepo.Tokens{AccessToken: tok.AccessToken, ExpiresAt: time.Now().UTC().Add(tok.ExpiresIn)})
		s.mu.Unlock()
		if e == nil {
			return saved.AccessToken, nil
		}
		if !errors.Is(e, installationrepo.ErrTokenCAS) {
			return "", ErrTokenUnavailable
		}
	}
	return "", ErrTokenUnavailable
}
