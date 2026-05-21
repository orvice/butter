package mcpoauth

import (
	"context"
	"errors"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var ErrNotFound = errors.New("mcp oauth connection not found")

// Connection stores the server-side OAuth state for one workspace-scoped MCP
// server. Secret fields are encrypted before they reach repository storage.
type Connection struct {
	WorkspaceID             string
	ServerID                string
	UserID                  string
	State                   agentsv1.MCPOAuthConnectionState
	ClientID                string
	EncryptedClientSecret   string
	AuthorizationURL        string
	TokenURL                string
	Resource                string
	Scopes                  []string
	EncryptedToken          string
	ConnectedAt             time.Time
	ExpiresAt               time.Time
	LastCheckedAt           time.Time
	LastError               string
	ReauthorizationRequired bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// Clone returns a detached copy safe for callers to mutate.
func (c *Connection) Clone() *Connection {
	if c == nil {
		return nil
	}
	out := *c
	if c.Scopes != nil {
		out.Scopes = append([]string(nil), c.Scopes...)
	}
	return &out
}

// Repository persists OAuth connection records keyed by workspace and MCP
// server id.
type Repository interface {
	EnsureIndexes(ctx context.Context) error
	Get(ctx context.Context, workspaceID, serverID string) (*Connection, error)
	Save(ctx context.Context, conn *Connection) error
	Delete(ctx context.Context, workspaceID, serverID string) error
	MarkState(ctx context.Context, workspaceID, serverID string, state agentsv1.MCPOAuthConnectionState, detail string, at time.Time) error
}
