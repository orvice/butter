package cursorbox

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/pkg/proto/butterbox/cursor/v1/cursorv1connect"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	"go.orx.me/apps/butter/internal/secretbox"
)

// ClientFactory resolves a workspace's ButterBox into a CursorService
// client. Implementations resolve per call — no client cache — so a
// base-URL change or token rotation takes effect on the next turn (same
// contract as internal/telegramapi and pibox).
type ClientFactory interface {
	ClientFor(ctx context.Context, workspaceID, butterboxID string) (cursorv1connect.CursorServiceClient, error)
}

// Factory is the production ClientFactory: it reads the box from the
// repository and decrypts its access token through the secretbox keyring,
// then builds a generated Connect client targeting the box's CursorService
// (butterbox.cursor.v1 — the #315 contract, mirrored in this repo's proto
// tree until butter-box publishes it).
type Factory struct {
	repo    butterboxrepo.Repository
	keyring *secretbox.Keyring
	// httpClient overrides the HTTP client, for tests. The default client
	// carries no global timeout: SendMessage is bounded by the bridge's
	// max-run deadline, not per client.
	httpClient connect.HTTPClient
}

func NewFactory(repo butterboxrepo.Repository, keyring *secretbox.Keyring) *Factory {
	return &Factory{repo: repo, keyring: keyring}
}

func (f *Factory) ClientFor(ctx context.Context, workspaceID, butterboxID string) (cursorv1connect.CursorServiceClient, error) {
	if f == nil || f.repo == nil {
		return nil, fmt.Errorf("cursorbox: butterbox repository is not configured")
	}
	box, err := f.repo.Get(ctx, workspaceID, butterboxID)
	if err != nil {
		if errors.Is(err, butterboxrepo.ErrNotFound) {
			return nil, fmt.Errorf("cursorbox: butterbox %q no longer exists in this workspace; point the agent at a registered ButterBox", butterboxID)
		}
		return nil, fmt.Errorf("cursorbox: resolve butterbox %q: %w", butterboxID, err)
	}

	token := ""
	if box.GetCredentialSet() {
		if f.keyring == nil {
			return nil, fmt.Errorf("cursorbox: butterbox %q has a stored token but credential encryption is not configured", box.GetName())
		}
		cred, err := f.repo.GetCredential(ctx, workspaceID, butterboxID)
		if err != nil {
			return nil, fmt.Errorf("cursorbox: read butterbox %q credential: %w", box.GetName(), err)
		}
		plaintext, err := f.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
		if err != nil {
			return nil, fmt.Errorf("cursorbox: decrypt butterbox %q token: %w", box.GetName(), err)
		}
		token = string(plaintext)
	}

	httpClient := f.httpClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	opts := []connect.ClientOption{}
	if token != "" {
		opts = append(opts, connect.WithInterceptors(bearerInterceptor(token)))
	}
	return cursorv1connect.NewCursorServiceClient(httpClient, box.GetBaseUrl(), opts...), nil
}

func bearerInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}
