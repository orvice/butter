// Package connectx contains the adapters that let our existing
// Twirp-shaped service implementations run behind ConnectRPC handlers.
//
// Service methods in internal/application return *connect.Error values that
// WrapUnary forwards directly. The helpers below cover the common shorthand
// cases (required argument, not found, internal-with-error, …) so callers
// don't need to spell out `connect.NewError(connect.CodeX, errors.New(...))`
// on every line.
package connectx

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
)

// RequiredArgument returns `<name> is required` with CodeInvalidArgument.
func RequiredArgument(name string) *connect.Error {
	cerr := connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s is required", name))
	cerr.Meta().Set("argument", name)
	return cerr
}

// InvalidArgument returns `<name> <validation>` with CodeInvalidArgument.
func InvalidArgument(name, validation string) *connect.Error {
	cerr := connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s %s", name, validation))
	cerr.Meta().Set("argument", name)
	return cerr
}

// NotFound returns the given message with CodeNotFound.
func NotFound(msg string) *connect.Error {
	return connect.NewError(connect.CodeNotFound, errors.New(msg))
}

// Internal returns the given message with CodeInternal.
func Internal(msg string) *connect.Error {
	return connect.NewError(connect.CodeInternal, errors.New(msg))
}

// InternalWith wraps the underlying error with CodeInternal.
func InternalWith(err error) *connect.Error {
	return connect.NewError(connect.CodeInternal, err)
}
