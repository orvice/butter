// Package cursorbox bridges Cursor SDK Bridge sessions hosted on a ButterBox
// (github.com/orvice/butter-box, CursorService) into the ADK agent interface
// so an AGENT_TYPE_CURSOR leaf can be invoked like any other agent. It mirrors
// internal/runtime/pibox (ADR-0011).
//
// Unlike PiService, CursorService.SendMessage is one held call that returns
// the final text once the Cursor turn ends; the turn's cancellation rides that
// call's context (its deadline is enforced client-side, ADR-0012 §3), and
// AbortSession is issued on the same single cancellation path. One Cursor session exists per (butter
// session × agent), keyed in ADK session state; on a repointed agent or a
// session the box no longer knows, the bridge abandons and recreates —
// sessions are never migrated. The Cursor API key lives in the box's
// environment; butter never sees it.
package cursorbox

import (
	"context"
	"fmt"

	"github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1/cursorv1connect"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	"go.orx.me/apps/butter/internal/runtime/butterboxconn"
	"go.orx.me/apps/butter/internal/secretbox"
)

// ClientFactory resolves a workspace's ButterBox into a CursorService
// client. Implementations resolve per call — no client cache — so a base-URL
// change or token rotation takes effect on the next turn.
type ClientFactory interface {
	ClientFor(ctx context.Context, workspaceID, butterboxID string) (cursorv1connect.CursorServiceClient, error)
}

// Factory is the production ClientFactory: box resolution and token
// decryption go through the shared butterboxconn.Resolver.
type Factory struct {
	resolver *butterboxconn.Resolver
}

func NewFactory(repo butterboxrepo.Repository, keyring *secretbox.Keyring) *Factory {
	return &Factory{resolver: butterboxconn.NewResolver(repo, keyring)}
}

func (f *Factory) ClientFor(ctx context.Context, workspaceID, butterboxID string) (cursorv1connect.CursorServiceClient, error) {
	conn, err := f.resolver.Resolve(ctx, workspaceID, butterboxID)
	if err != nil {
		return nil, fmt.Errorf("cursorbox: %w", err)
	}
	return cursorv1connect.NewCursorServiceClient(conn.HTTPClient, conn.BaseURL, conn.Options...), nil
}
