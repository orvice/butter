// Package sessionshare decides who may hold a session ID.
//
// The session store keeps a conversation's events under (app_name,
// session_id), not per user: every user with a session document under one
// ID reads and appends the same events. A Telegram Destination session
// relies on that — each member of the chat gets their own document under the
// Destination's session ID, and together they share one conversation.
// Anywhere else a second holder would read another principal's conversation,
// so the store refuses to create a session whose ID another user holds unless
// the creating context opts in with Allow.
package sessionshare

import (
	"context"
	"errors"
)

// ErrIDTaken reports that another user already holds the session ID within
// the app.
var ErrIDTaken = errors.New("session id is held by another user")

type allowKey struct{}

// Allow marks ctx so that creating a session may join a session ID other
// users already hold. Only a path that derives the session ID itself may set
// it, never one where the caller chose the ID.
func Allow(ctx context.Context) context.Context {
	return context.WithValue(ctx, allowKey{}, true)
}

// Allowed reports whether ctx was marked by Allow.
func Allowed(ctx context.Context) bool {
	allowed, _ := ctx.Value(allowKey{}).(bool)
	return allowed
}
