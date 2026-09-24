// Package butterboxconn resolves a workspace's ButterBox into the connection
// parameters every box-backed bridge needs: base URL, HTTP client, and the
// bearer credential decrypted through the secretbox keyring. It is shared by
// the PiService bridge (internal/runtime/pibox) and the CursorService bridge
// (internal/runtime/cursorbox); each wraps the result in its own generated
// Connect client.
package butterboxconn

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	"go.orx.me/apps/butter/internal/secretbox"
)

// Conn is one resolved box connection, ready to hand to a generated
// NewXxxServiceClient(conn.HTTPClient, conn.BaseURL, conn.Options...).
type Conn struct {
	BaseURL    string
	HTTPClient connect.HTTPClient
	Options    []connect.ClientOption
}

// Resolver reads the box from the repository and decrypts its access token.
// It resolves per call — no cache — so a base-URL change or token rotation
// takes effect on the next turn (same contract as internal/telegramapi).
type Resolver struct {
	repo    butterboxrepo.Repository
	keyring *secretbox.Keyring
	// HTTPClient overrides the HTTP client, for tests. The default client
	// carries no global timeout: turns are bounded per call by the bridges,
	// not per client.
	HTTPClient connect.HTTPClient
}

func NewResolver(repo butterboxrepo.Repository, keyring *secretbox.Keyring) *Resolver {
	return &Resolver{repo: repo, keyring: keyring}
}

// Resolve returns the connection for one box. Errors are phrased for the
// agent's user; callers prefix them with their own package name.
func (r *Resolver) Resolve(ctx context.Context, workspaceID, butterboxID string) (Conn, error) {
	if r == nil || r.repo == nil {
		return Conn{}, fmt.Errorf("butterbox repository is not configured")
	}
	box, err := r.repo.Get(ctx, workspaceID, butterboxID)
	if err != nil {
		if errors.Is(err, butterboxrepo.ErrNotFound) {
			return Conn{}, fmt.Errorf("butterbox %q no longer exists in this workspace; point the agent at a registered ButterBox", butterboxID)
		}
		return Conn{}, fmt.Errorf("resolve butterbox %q: %w", butterboxID, err)
	}

	token := ""
	if box.GetCredentialSet() {
		if r.keyring == nil {
			return Conn{}, fmt.Errorf("butterbox %q has a stored token but credential encryption is not configured", box.GetName())
		}
		cred, err := r.repo.GetCredential(ctx, workspaceID, butterboxID)
		if err != nil {
			return Conn{}, fmt.Errorf("read butterbox %q credential: %w", box.GetName(), err)
		}
		plaintext, err := r.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
		if err != nil {
			return Conn{}, fmt.Errorf("decrypt butterbox %q token: %w", box.GetName(), err)
		}
		token = string(plaintext)
	}

	httpClient := r.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	conn := Conn{BaseURL: box.GetBaseUrl(), HTTPClient: httpClient}
	if token != "" {
		conn.Options = append(conn.Options, connect.WithInterceptors(BearerInterceptor(token)))
	}
	return conn, nil
}

// BearerInterceptor stamps the box access token on every unary call.
func BearerInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}
