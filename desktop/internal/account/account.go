// Package account owns the client's signed-in context: which server it
// talks to, which token it holds, which workspace it uses.
//
// Every switch of server or account goes through this type so that a late
// response from the previous context can never be applied to the new one:
// each asynchronous call captures a generation, and results whose
// generation moved are dropped.
package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.orx.me/apps/butter/desktop/internal/authapi"
	"go.orx.me/apps/butter/desktop/internal/credentials"
	"go.orx.me/apps/butter/desktop/internal/endpoint"
)

var (
	// ErrNoServer reports a call before a server address was chosen.
	ErrNoServer = errors.New("account: no server selected")
	// ErrNoSession reports a server with no stored session.
	ErrNoSession = errors.New("account: no stored session")
	// ErrNotSignedIn reports a call that needs a signed-in account.
	ErrNotSignedIn = errors.New("account: not signed in")
	// ErrUnknownWorkspace reports a workspace the account cannot reach.
	ErrUnknownWorkspace = errors.New("account: workspace is not available to this account")
	// ErrStale reports a result from a context the user has left; the
	// caller should ignore it rather than show it.
	ErrStale = errors.New("account: the server or account changed meanwhile")
)

// API is the subset of authapi.Client the manager needs; tests replace it.
type API interface {
	Login(ctx context.Context, username, password string) (authapi.Session, error)
	Me(ctx context.Context, token string) (authapi.User, error)
	Workspaces(ctx context.Context, token string) ([]authapi.Workspace, error)
	Logout(ctx context.Context, token string) error
}

// State is what the interface shows. It never carries the token.
type State struct {
	Server      string
	User        authapi.User
	Workspaces  []authapi.Workspace
	WorkspaceID string
	ExpiresAt   time.Time
}

// SignedIn reports whether state describes an account.
func (s State) SignedIn() bool { return s.User.ID != "" || s.User.Username != "" }

// Manager is the client's signed-in context.
type Manager struct {
	store credentials.Store
	dial  func(apiBase string) API

	mu         sync.Mutex
	server     endpoint.Server
	hasServer  bool
	token      string
	state      State
	generation uint64
}

// New returns a manager that keeps tokens in store and builds an API per
// server with dial.
func New(store credentials.Store, dial func(apiBase string) API) *Manager {
	if dial == nil {
		dial = func(apiBase string) API { return authapi.New(apiBase, nil) }
	}
	return &Manager{store: store, dial: dial}
}

// UseServer switches to raw, clearing the previous server's stored token
// first. Switching to the current server keeps everything.
func (m *Manager) UseServer(raw string) (State, error) {
	srv, err := endpoint.Parse(raw)
	if err != nil {
		return State{}, err
	}
	m.mu.Lock()
	prev, had := m.server, m.hasServer
	m.mu.Unlock()
	if had && prev.String() == srv.String() {
		return m.Snapshot(), nil
	}
	if had {
		if err := m.store.Delete(tokenKey(prev)); err != nil && !errors.Is(err, credentials.ErrNotFound) {
			// Refuse the switch: carrying on would leave a token for the
			// old server behind while the window claims to be elsewhere.
			return State{}, fmt.Errorf("account: clear the previous server's session: %w", err)
		}
	}
	m.mu.Lock()
	m.server, m.hasServer = srv, true
	m.token = ""
	m.state = State{Server: srv.String()}
	m.generation++
	m.mu.Unlock()
	return m.Snapshot(), nil
}

// SignIn logs in on the current server. A password is used for this call
// only and never stored.
func (m *Manager) SignIn(ctx context.Context, username, password string) (State, error) {
	srv, gen, api, ok := m.context()
	if !ok {
		return State{}, ErrNoServer
	}
	sess, err := api.Login(ctx, strings.TrimSpace(username), password)
	if err != nil {
		return State{}, err
	}
	if len(sess.Workspaces) == 0 {
		// A global admin with no explicit memberships gets none from Login;
		// ListWorkspaces is the authority for what they may pick.
		if list, err := api.Workspaces(ctx, sess.Token); err == nil {
			sess.Workspaces = list
		}
	}
	key := tokenKey(srv)
	if err := m.store.Set(key, []byte(sess.Token)); err != nil {
		m.revoke(api, sess.Token)
		return State{}, fmt.Errorf("account: store the session token: %w", err)
	}

	m.mu.Lock()
	if m.generation != gen {
		m.mu.Unlock()
		_ = m.store.Delete(key)
		m.revoke(api, sess.Token)
		return State{}, ErrStale
	}
	prev := m.token
	m.token = sess.Token
	m.state = State{
		Server:     srv.String(),
		User:       sess.User,
		Workspaces: sess.Workspaces,
		ExpiresAt:  sess.ExpiresAt,
	}
	if len(sess.Workspaces) == 1 {
		m.state.WorkspaceID = sess.Workspaces[0].ID
	}
	out := m.state
	m.mu.Unlock()

	if prev != "" && prev != sess.Token {
		m.revoke(api, prev)
	}
	return out, nil
}

// Restore signs in again from the stored token of the current server. It
// reports ErrNoSession when there is none.
func (m *Manager) Restore(ctx context.Context) (State, error) {
	srv, gen, api, ok := m.context()
	if !ok {
		return State{}, ErrNoServer
	}
	key := tokenKey(srv)
	raw, err := m.store.Get(key)
	switch {
	case errors.Is(err, credentials.ErrNotFound):
		return State{}, ErrNoSession
	case err != nil:
		return State{}, fmt.Errorf("account: read the session token: %w", err)
	}
	token := string(raw)

	user, err := api.Me(ctx, token)
	if err != nil {
		if authapi.Is(err, authapi.KindSessionExpired) {
			_ = m.store.Delete(key)
			m.forget(gen)
		}
		return State{}, err
	}
	workspaces, err := api.Workspaces(ctx, token)
	if err != nil {
		return State{}, err
	}

	m.mu.Lock()
	if m.generation != gen {
		m.mu.Unlock()
		return State{}, ErrStale
	}
	m.token = token
	m.state = State{
		Server:     srv.String(),
		User:       user,
		Workspaces: workspaces,
	}
	m.mu.Unlock()
	return m.Snapshot(), nil
}

// SignOut revokes the session on the server when it can, forgets it
// locally either way, and keeps the server address.
func (m *Manager) SignOut(ctx context.Context) error {
	srv, gen, api, ok := m.context()
	if !ok {
		return ErrNoServer
	}
	key := tokenKey(srv)
	raw, err := m.store.Get(key)
	token := ""
	switch {
	case errors.Is(err, credentials.ErrNotFound):
	case err != nil:
		return fmt.Errorf("account: read the session token: %w", err)
	default:
		token = string(raw)
	}
	if token != "" {
		// Best effort: a server that cannot be reached must not keep the
		// user signed in on this machine.
		_ = api.Logout(ctx, token)
	}
	if err := m.store.Delete(key); err != nil && !errors.Is(err, credentials.ErrNotFound) {
		return fmt.Errorf("account: forget the session token: %w", err)
	}
	m.forget(gen)
	return nil
}

// SelectWorkspace sets the workspace later calls address.
func (m *Manager) SelectWorkspace(id string) (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.state.SignedIn() {
		return State{}, ErrNotSignedIn
	}
	if id != "" {
		found := false
		for _, w := range m.state.Workspaces {
			if w.ID == id {
				found = true
				break
			}
		}
		if !found {
			return State{}, ErrUnknownWorkspace
		}
	}
	m.state.WorkspaceID = id
	return m.state, nil
}

// Snapshot returns the current state.
func (m *Manager) Snapshot() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Server returns the chosen server.
func (m *Manager) Server() (endpoint.Server, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.server, m.hasServer
}

// SignedIn reports whether a session is held.
func (m *Manager) SignedIn() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.token != ""
}

// Token returns the session token for the API layers that need a header.
func (m *Manager) Token() (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.token, m.token != ""
}

// Expired reports whether the session's own expiry has passed. The server
// stays the authority; this only lets the interface warn before a call
// fails.
func (m *Manager) Expired(now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.state.ExpiresAt.IsZero() && !now.Before(m.state.ExpiresAt)
}

func (m *Manager) context() (endpoint.Server, uint64, API, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.hasServer {
		return endpoint.Server{}, 0, nil, false
	}
	return m.server, m.generation, m.dial(m.server.APIBase()), true
}

func (m *Manager) forget(gen uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.generation != gen {
		return
	}
	m.token = ""
	m.state = State{Server: m.state.Server}
}

func (m *Manager) revoke(api API, token string) {
	if api == nil || token == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = api.Logout(ctx, token)
}

// tokenKey scopes a stored token to its server, so one server's session is
// never offered to another.
func tokenKey(s endpoint.Server) string { return "butter-session-" + s.Key() }
