// Package workspace provides workspace scoping primitives shared across
// the HTTP middleware, Twirp services, repositories and runtime.
//
// The current workspace is propagated as a string ID on the request
// context. Services read it via FromContext or MustFromContext; the HTTP
// middleware injects it from the `X-Workspace-ID` header after validating
// the caller's membership.
package workspace

import (
	"context"
	"errors"
)

// HeaderName is the HTTP header used to communicate the selected workspace.
const HeaderName = "X-Workspace-ID"

// DefaultSlug is the slug of the bootstrap workspace created on first startup.
const DefaultSlug = "default"

// ErrMissing is returned when a request requires a workspace but none is
// attached to the context.
var ErrMissing = errors.New("workspace id missing from context")

type contextKey struct{}

// WithID returns a new context carrying the given workspace id.
func WithID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, contextKey{}, id)
}

// FromContext returns the workspace id carried by ctx, if any.
func FromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	id, ok := ctx.Value(contextKey{}).(string)
	return id, ok && id != ""
}

// MustFromContext is a helper for services that require a workspace. It
// returns an error rather than panicking so callers can map it to twirp
// FailedPrecondition / InvalidArgument.
func MustFromContext(ctx context.Context) (string, error) {
	id, ok := FromContext(ctx)
	if !ok {
		return "", ErrMissing
	}
	return id, nil
}
