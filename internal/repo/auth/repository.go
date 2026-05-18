package auth

import (
	"context"
	"errors"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrSessionNotFound   = errors.New("auth session not found")
	ErrUserDisabled      = errors.New("user disabled")
)

type Session struct {
	ID         string
	UserID     string
	TokenHash  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastUsedAt time.Time
	Revoked    bool
}

type Repository interface {
	EnsureIndexes(ctx context.Context) error
	CountUsers(ctx context.Context) (int64, error)
	ListUsers(ctx context.Context) ([]*agentsv1.User, error)
	CreateUser(ctx context.Context, user *agentsv1.User, passwordHash string) error
	UpdateUserPassword(ctx context.Context, id string, passwordHash string, updatedAt time.Time) (*agentsv1.User, error)
	// UpdateUserProfile updates the user's display name and (optionally)
	// avatar URL. If avatarURL is nil the stored avatar is left untouched;
	// pass a pointer to an empty string to clear it.
	UpdateUserProfile(ctx context.Context, id string, displayName string, avatarURL *string, updatedAt time.Time) (*agentsv1.User, error)
	SetUserDisabled(ctx context.Context, id string, disabled bool, updatedAt time.Time) (*agentsv1.User, error)
	FindUserByUsername(ctx context.Context, username string) (*agentsv1.User, string, error)
	FindUserByID(ctx context.Context, id string) (*agentsv1.User, string, error)
	GetUser(ctx context.Context, id string) (*agentsv1.User, error)
	CreateSession(ctx context.Context, session *Session) error
	LookupSession(ctx context.Context, tokenHash string, now time.Time) (*Session, *agentsv1.User, error)
	TouchSession(ctx context.Context, id string, at time.Time) error
	RevokeSession(ctx context.Context, id string) error
}

type contextKey string

const (
	userContextKey    contextKey = "auth_user"
	sessionContextKey contextKey = "auth_session"
	adminContextKey   contextKey = "auth_admin"
)

func WithAuthenticated(ctx context.Context, user *agentsv1.User, session *Session) context.Context {
	ctx = context.WithValue(ctx, userContextKey, user)
	ctx = context.WithValue(ctx, sessionContextKey, session)
	return ctx
}

func UserFromContext(ctx context.Context) (*agentsv1.User, bool) {
	user, ok := ctx.Value(userContextKey).(*agentsv1.User)
	return user, ok && user != nil
}

func SessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionContextKey).(*Session)
	return session, ok && session != nil
}

// WithAdmin tags a context as belonging to an admin-equivalent caller: a
// session user with role "admin", a request authenticated via the root API
// token, or the dev/legacy bootstrap path. Services that expose
// cross-workspace data (notably the dashboard) check IsAdmin to scope access.
func WithAdmin(ctx context.Context) context.Context {
	return context.WithValue(ctx, adminContextKey, true)
}

// IsAdmin reports whether the context was tagged by WithAdmin, or whether the
// authenticated user carries role "admin".
func IsAdmin(ctx context.Context) bool {
	if flag, ok := ctx.Value(adminContextKey).(bool); ok && flag {
		return true
	}
	if user, ok := UserFromContext(ctx); ok && user.GetRole() == "admin" {
		return true
	}
	return false
}
