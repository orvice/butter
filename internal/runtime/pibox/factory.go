// Package pibox bridges pi coding-agent sessions hosted on a ButterBox
// (github.com/orvice/butter-box, PiService) into the ADK agent interface so
// an AGENT_TYPE_PI leaf can be invoked like any other agent (ADR-0011).
//
// The bridge drives butter-box's asynchronous turn API: SubmitMessage
// returns an entries cursor and GetTurn long-polls for settlement, so a
// turn's life is never tied to one idle HTTP connection. One pi session
// exists per (butter session × agent), keyed in ADK session state; on a
// repointed agent or a session the box no longer knows, the bridge abandons
// and recreates — sessions are never migrated. butter never deletes data on
// the box.
package pibox

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	"go.orx.me/apps/butter/internal/secretbox"
)

// ClientFactory resolves a workspace's ButterBox into a PiService client.
// Implementations resolve per call — no client cache — so a base-URL change
// or token rotation takes effect on the next turn (same contract as
// internal/telegramapi).
type ClientFactory interface {
	ClientFor(ctx context.Context, workspaceID, butterboxID string) (piv1connect.PiServiceClient, error)
}

// Factory is the production ClientFactory: it reads the box from the
// repository and decrypts its access token through the secretbox keyring.
type Factory struct {
	repo    butterboxrepo.Repository
	keyring *secretbox.Keyring
	// httpClient overrides the HTTP client, for tests. The default client
	// carries no global timeout: turn long-polls are bounded per call by the
	// bridge, not per client.
	httpClient connect.HTTPClient
}

func NewFactory(repo butterboxrepo.Repository, keyring *secretbox.Keyring) *Factory {
	return &Factory{repo: repo, keyring: keyring}
}

func (f *Factory) ClientFor(ctx context.Context, workspaceID, butterboxID string) (piv1connect.PiServiceClient, error) {
	if f == nil || f.repo == nil {
		return nil, fmt.Errorf("pibox: butterbox repository is not configured")
	}
	box, err := f.repo.Get(ctx, workspaceID, butterboxID)
	if err != nil {
		if errors.Is(err, butterboxrepo.ErrNotFound) {
			return nil, fmt.Errorf("pibox: butterbox %q no longer exists in this workspace; point the agent at a registered ButterBox", butterboxID)
		}
		return nil, fmt.Errorf("pibox: resolve butterbox %q: %w", butterboxID, err)
	}

	token := ""
	if box.GetCredentialSet() {
		if f.keyring == nil {
			return nil, fmt.Errorf("pibox: butterbox %q has a stored token but credential encryption is not configured", box.GetName())
		}
		cred, err := f.repo.GetCredential(ctx, workspaceID, butterboxID)
		if err != nil {
			return nil, fmt.Errorf("pibox: read butterbox %q credential: %w", box.GetName(), err)
		}
		plaintext, err := f.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
		if err != nil {
			return nil, fmt.Errorf("pibox: decrypt butterbox %q token: %w", box.GetName(), err)
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
	return piv1connect.NewPiServiceClient(httpClient, box.GetBaseUrl(), opts...), nil
}

func bearerInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}
