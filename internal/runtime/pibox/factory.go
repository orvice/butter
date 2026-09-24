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
	"fmt"

	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	"go.orx.me/apps/butter/internal/runtime/butterboxconn"
	"go.orx.me/apps/butter/internal/secretbox"
)

// ClientFactory resolves a workspace's ButterBox into a PiService client.
// Implementations resolve per call — no client cache — so a base-URL change
// or token rotation takes effect on the next turn (same contract as
// internal/telegramapi).
type ClientFactory interface {
	ClientFor(ctx context.Context, workspaceID, butterboxID string) (piv1connect.PiServiceClient, error)
}

// Factory is the production ClientFactory: box resolution and token
// decryption go through the shared butterboxconn.Resolver.
type Factory struct {
	resolver *butterboxconn.Resolver
}

func NewFactory(repo butterboxrepo.Repository, keyring *secretbox.Keyring) *Factory {
	return &Factory{resolver: butterboxconn.NewResolver(repo, keyring)}
}

func (f *Factory) ClientFor(ctx context.Context, workspaceID, butterboxID string) (piv1connect.PiServiceClient, error) {
	conn, err := f.resolver.Resolve(ctx, workspaceID, butterboxID)
	if err != nil {
		return nil, fmt.Errorf("pibox: %w", err)
	}
	return piv1connect.NewPiServiceClient(conn.HTTPClient, conn.BaseURL, conn.Options...), nil
}
