package application

import (
	"context"
	"errors"
	"slices"
	"strings"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/repo/auth"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
)

// Session access policy. A session is addressed by (app_name, user_id,
// session_id) and owned by one workspace. Who may act on it:
//
//   - a global admin: every session;
//   - a person: their own sessions (user_id is theirs); with X-Workspace-ID
//     set, only the ones in that workspace;
//   - a workspace owner or admin: every session in that workspace;
//   - any workspace member: the workspace's automation sessions (apps
//     `cron:<job>` and `automation:<name>`), which have no person behind
//     them and belong to resources members manage;
//   - a workspace API token: every session in its workspace (only owners
//     and admins can mint one).
//
// Without X-Workspace-ID a non-admin reaches only their own sessions. A
// denied read or change answers NotFound, so a session's existence is never
// revealed; creating a session for someone else answers PermissionDenied.

// automationAppPrefixes name the apps whose sessions are run by Butter for a
// workspace resource rather than for a person: cron jobs and automations.
var automationAppPrefixes = []string{"cron:", "automation:"}

func isAutomationApp(appName string) bool {
	for _, p := range automationAppPrefixes {
		if strings.HasPrefix(appName, p) {
			return true
		}
	}
	return false
}

// sessionPrincipal is the caller as the session policy sees it.
type sessionPrincipal struct {
	admin    bool
	userID   string // the signed-in person; empty for an API token
	apiToken bool
	wsID     string // the request's workspace; empty without X-Workspace-ID
}

func sessionPrincipalFrom(ctx context.Context) (sessionPrincipal, error) {
	wsID, _ := workspace.FromContext(ctx)
	p := sessionPrincipal{admin: auth.IsAdmin(ctx), wsID: wsID}
	if user, ok := auth.UserFromContext(ctx); ok {
		p.userID = user.GetId()
	}
	if _, ok := auth.APITokenFromContext(ctx); ok && wsID != "" {
		p.apiToken = true
	}
	if !p.admin && p.userID == "" && !p.apiToken {
		return p, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	return p, nil
}

func (p sessionPrincipal) owns(userID string) bool {
	return p.userID != "" && p.userID == userID
}

// SetWorkspaceRepo wires workspace memberships, which decide whether a caller
// is an owner or admin of the request's workspace.
func (s *SessionServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	if repo == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wsRepo = repo
}

func (s *SessionServiceServer) getWSRepo() workspacerepo.Repository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wsRepo
}

// managesWorkspace reports whether the caller is an owner or admin of the
// request's workspace. Without a membership repository it fails closed.
func (s *SessionServiceServer) managesWorkspace(ctx context.Context, p sessionPrincipal) (bool, error) {
	if p.userID == "" || p.wsID == "" {
		return false, nil
	}
	repo := s.getWSRepo()
	if repo == nil {
		return false, nil
	}
	member, err := repo.GetMember(ctx, p.wsID, p.userID)
	if err != nil {
		if errors.Is(err, workspacerepo.ErrNotFound) {
			return false, nil
		}
		return false, connectx.InternalWith(err)
	}
	return slices.Contains([]string{"owner", "admin"}, member.GetRole()), nil
}

// sessionNotFound is the single answer for a missing session and for one the
// caller may not see.
func sessionNotFound() error { return connectx.NotFound("session not found") }

// authorizeExisting decides access to a session that exists in workspace
// sessWS ("" for a legacy session without one).
func (s *SessionServiceServer) authorizeExisting(ctx context.Context, appName, userID, sessWS string) error {
	p, err := sessionPrincipalFrom(ctx)
	if err != nil {
		return err
	}
	if p.admin {
		return nil
	}
	if p.wsID != "" && sessWS != p.wsID {
		return sessionNotFound()
	}
	if p.owns(userID) {
		return nil
	}
	if p.wsID == "" {
		return sessionNotFound()
	}
	if p.apiToken || isAutomationApp(appName) {
		return nil
	}
	manager, err := s.managesWorkspace(ctx, p)
	if err != nil {
		return err
	}
	if !manager {
		return sessionNotFound()
	}
	log.FromContext(ctx).Info("workspace manager accessing another user's session",
		"audit", "workspace_session_access", "workspace_id", p.wsID,
		"user_id", p.userID, "session_app_name", appName, "session_user_id", userID)
	return nil
}

// authorizeNew decides whether the caller may start a session for userID in
// the request's workspace.
func (s *SessionServiceServer) authorizeNew(ctx context.Context, userID string) error {
	p, err := sessionPrincipalFrom(ctx)
	if err != nil {
		return err
	}
	if p.admin || p.owns(userID) || p.apiToken {
		return nil
	}
	manager, err := s.managesWorkspace(ctx, p)
	if err != nil {
		return err
	}
	if !manager {
		return connect.NewError(connect.CodePermissionDenied, errors.New("cannot start a session for another user"))
	}
	return nil
}

// lookupSessionWorkspace finds the workspace of an addressed session;
// exists is false when there is no such session.
func (s *SessionServiceServer) lookupSessionWorkspace(ctx context.Context, appName, userID, sessionID string) (wsID string, exists bool, err error) {
	if store := s.getWSStore(); store != nil {
		ws, err := store.GetWorkspaceID(ctx, appName, userID, sessionID)
		if err != nil {
			if strings.Contains(err.Error(), "session not found") {
				return "", false, nil
			}
			return "", false, connectx.InternalWith(err)
		}
		return ws, true, nil
	}
	svc := s.getSessionSvc()
	if svc == nil {
		return "", false, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
	}
	resp, err := svc.Get(ctx, &session.GetRequest{AppName: appName, UserID: userID, SessionID: sessionID, NumRecentEvents: 1})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "session not found") {
			return "", false, nil
		}
		return "", false, connectx.InternalWith(err)
	}
	return sessionWorkspaceID(resp.Session), true, nil
}

func sessionWorkspaceID(sess session.Session) string {
	if ws, ok := sess.(workspacedSession); ok {
		return ws.WorkspaceID()
	}
	return ""
}

// authorizeByID decides access to an addressed session that must exist.
func (s *SessionServiceServer) authorizeByID(ctx context.Context, appName, userID, sessionID string) error {
	if auth.IsAdmin(ctx) {
		return nil
	}
	wsID, exists, err := s.lookupSessionWorkspace(ctx, appName, userID, sessionID)
	if err != nil {
		return err
	}
	if !exists {
		// Authenticate before answering, so an anonymous caller learns
		// nothing either way.
		if _, err := sessionPrincipalFrom(ctx); err != nil {
			return err
		}
		return sessionNotFound()
	}
	return s.authorizeExisting(ctx, appName, userID, wsID)
}

// authorizeReply decides access for a turn on a session that may not exist
// yet: an existing one follows the read rules, a new one the create rules.
func (s *SessionServiceServer) authorizeReply(ctx context.Context, appName, userID, sessionID string) error {
	if auth.IsAdmin(ctx) {
		return nil
	}
	wsID, exists, err := s.lookupSessionWorkspace(ctx, appName, userID, sessionID)
	if err != nil {
		return err
	}
	if exists {
		return s.authorizeExisting(ctx, appName, userID, wsID)
	}
	return s.authorizeNew(ctx, userID)
}

// sessionListScope narrows an unscoped listing to what the caller may see.
type sessionListScope struct {
	userID      string // list this user's sessions; "" lists every user's
	workspaceID string // keep only this workspace's sessions; "" keeps all
}

// listScope applies the policy to an unscoped ListSessions. With
// X-Workspace-ID the listing stays inside that workspace; a member asking
// for every user sees their own sessions, and asking for another person's
// is refused.
func (s *SessionServiceServer) listScope(ctx context.Context, appName, userID string) (sessionListScope, error) {
	p, err := sessionPrincipalFrom(ctx)
	if err != nil {
		return sessionListScope{}, err
	}
	if p.admin {
		return sessionListScope{userID: userID}, nil
	}
	denied := connect.NewError(connect.CodePermissionDenied, errors.New("cannot list another user's sessions"))
	if p.wsID == "" {
		if userID == "" || p.owns(userID) {
			return sessionListScope{userID: p.userID}, nil
		}
		return sessionListScope{}, denied
	}
	scope := sessionListScope{userID: userID, workspaceID: p.wsID}
	if p.owns(userID) || p.apiToken || isAutomationApp(appName) {
		return scope, nil
	}
	manager, err := s.managesWorkspace(ctx, p)
	if err != nil {
		return sessionListScope{}, err
	}
	if manager {
		return scope, nil
	}
	if userID == "" {
		scope.userID = p.userID
		return scope, nil
	}
	return sessionListScope{}, denied
}
