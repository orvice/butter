// Package authapi talks to Butter's AuthService and WorkspaceService.
//
// It converts protobuf responses into the few values the client keeps
// (user, workspaces, expiry) and every ConnectRPC failure into an *Error
// whose Kind the interface can act on: wrong password, disabled user,
// expired session, unreachable server.
package authapi

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// User is the signed-in account.
type User struct {
	ID          string
	Username    string
	DisplayName string
	Role        string
	Disabled    bool
}

// Workspace is a workspace the account can reach.
type Workspace struct {
	ID   string
	Name string
	Slug string
}

// Session is a successful Login.
type Session struct {
	Token      string
	ExpiresAt  time.Time
	User       User
	Workspaces []Workspace
}

// Client is bound to one server's API base URL.
type Client struct {
	auth agentsv1connect.AuthServiceClient
	ws   agentsv1connect.WorkspaceServiceClient
}

// New builds a client for apiBase, for example
// "https://butter.example.com/api". A nil httpClient uses a 30s timeout.
func New(apiBase string, httpClient connect.HTTPClient) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		auth: agentsv1connect.NewAuthServiceClient(httpClient, apiBase),
		ws:   agentsv1connect.NewWorkspaceServiceClient(httpClient, apiBase),
	}
}

// TokenHeader is the header that carries a session token.
const TokenHeader = "Authorization"

// Bearer returns the header value for token.
func Bearer(token string) string { return "Bearer " + token }

// Login exchanges a username and password for a session.
func (c *Client) Login(ctx context.Context, username, password string) (Session, error) {
	res, err := c.auth.Login(ctx, connect.NewRequest(&agentsv1.LoginRequest{
		Username: username,
		Password: password,
	}))
	if err != nil {
		return Session{}, classify("login", err, callUnauthenticated)
	}
	msg := res.Msg
	s := Session{
		Token:      msg.GetToken(),
		User:       user(msg.GetUser()),
		Workspaces: workspaces(msg.GetWorkspaces()),
	}
	if ts := msg.GetExpiresAt(); ts.IsValid() {
		s.ExpiresAt = ts.AsTime()
	}
	return s, nil
}

// Me validates token and returns the account it belongs to.
func (c *Client) Me(ctx context.Context, token string) (User, error) {
	req := connect.NewRequest(&agentsv1.MeRequest{})
	req.Header().Set(TokenHeader, Bearer(token))
	res, err := c.auth.Me(ctx, req)
	if err != nil {
		return User{}, classify("me", err, callExpired)
	}
	return user(res.Msg.GetUser()), nil
}

// Workspaces lists the workspaces token can reach. A global admin with no
// explicit memberships still gets every workspace.
func (c *Client) Workspaces(ctx context.Context, token string) ([]Workspace, error) {
	req := connect.NewRequest(&agentsv1.ListWorkspacesRequest{})
	req.Header().Set(TokenHeader, Bearer(token))
	res, err := c.ws.ListWorkspaces(ctx, req)
	if err != nil {
		return nil, classify("list workspaces", err, callExpired)
	}
	return workspaces(res.Msg.GetWorkspaces()), nil
}

// Logout revokes token on the server. A token that is already gone is not
// an error: the client forgets it either way.
func (c *Client) Logout(ctx context.Context, token string) error {
	req := connect.NewRequest(&agentsv1.LogoutRequest{})
	req.Header().Set(TokenHeader, Bearer(token))
	if _, err := c.auth.Logout(ctx, req); err != nil {
		classified := classify("logout", err, callExpired)
		if KindOf(classified) == KindSessionExpired {
			return nil
		}
		return classified
	}
	return nil
}

func user(u *agentsv1.User) User {
	if u == nil {
		return User{}
	}
	return User{
		ID:          u.GetId(),
		Username:    u.GetUsername(),
		DisplayName: u.GetDisplayName(),
		Role:        u.GetRole(),
		Disabled:    u.GetDisabled(),
	}
}

func workspaces(in []*agentsv1.Workspace) []Workspace {
	out := make([]Workspace, 0, len(in))
	for _, w := range in {
		if w == nil {
			continue
		}
		out = append(out, Workspace{ID: w.GetId(), Name: w.GetName(), Slug: w.GetSlug()})
	}
	return out
}
