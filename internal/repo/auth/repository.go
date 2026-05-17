package auth

import (
	"context"
	"errors"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrSessionNotFound = errors.New("auth session not found")
	ErrUserDisabled    = errors.New("user disabled")
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
	CreateUser(ctx context.Context, user *agentsv1.User, passwordHash string) error
	FindUserByUsername(ctx context.Context, username string) (*agentsv1.User, string, error)
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
