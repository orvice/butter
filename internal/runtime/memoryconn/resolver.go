// Package memoryconn resolves a workspace's WorkspaceMemoryConfig into a
// ready mem0 client: base URL plus the API key decrypted through the
// secretbox keyring (ADR-0013). It is shared by the config service's
// connection probe and the runtime memory service.
package memoryconn

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.orx.me/apps/butter/internal/mem0"
	memoryconfigrepo "go.orx.me/apps/butter/internal/repo/memoryconfig"
	"go.orx.me/apps/butter/internal/secretbox"
)

// ErrNotConfigured means the workspace has no usable memory connection:
// no config at all, or (for Resolve) a disabled one. Callers treat it as
// "memory is off", not as a failure.
var ErrNotConfigured = errors.New("workspace memory is not configured")

// Conn is one workspace's resolved mem0 connection.
type Conn struct {
	BaseURL string
	APIKey  string
	Enabled bool
}

// Client builds a mem0 client for the connection.
func (c Conn) Client(httpClient *http.Client) *mem0.Client {
	return mem0.New(c.BaseURL, c.APIKey, httpClient)
}

// Resolver reads the config from the repository and decrypts its API key.
// It resolves per call — no cache — so a base-URL change or key rotation
// takes effect on the next call (same contract as butterboxconn).
type Resolver struct {
	repo    memoryconfigrepo.Repository
	keyring *secretbox.Keyring
}

func NewResolver(repo memoryconfigrepo.Repository, keyring *secretbox.Keyring) *Resolver {
	return &Resolver{repo: repo, keyring: keyring}
}

// Resolve returns the workspace's enabled connection, or ErrNotConfigured
// when there is no config or it is disabled.
func (r *Resolver) Resolve(ctx context.Context, workspaceID string) (Conn, error) {
	conn, err := r.Load(ctx, workspaceID)
	if err != nil {
		return Conn{}, err
	}
	if !conn.Enabled {
		return Conn{}, ErrNotConfigured
	}
	return conn, nil
}

// Load returns the workspace's connection whether or not it is enabled, or
// ErrNotConfigured when there is no config.
func (r *Resolver) Load(ctx context.Context, workspaceID string) (Conn, error) {
	if r == nil || r.repo == nil {
		return Conn{}, ErrNotConfigured
	}
	cfg, err := r.repo.Get(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, memoryconfigrepo.ErrNotFound) {
			return Conn{}, ErrNotConfigured
		}
		return Conn{}, fmt.Errorf("read workspace memory config: %w", err)
	}

	apiKey := ""
	if cfg.GetCredentialSet() {
		cred, err := r.repo.GetCredential(ctx, workspaceID)
		if err != nil {
			return Conn{}, fmt.Errorf("read workspace memory credential: %w", err)
		}
		plaintext, err := r.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
		if err != nil {
			return Conn{}, fmt.Errorf("decrypt workspace memory API key: %w", err)
		}
		apiKey = string(plaintext)
	}
	return Conn{BaseURL: cfg.GetBaseUrl(), APIKey: apiKey, Enabled: cfg.GetEnabled()}, nil
}
