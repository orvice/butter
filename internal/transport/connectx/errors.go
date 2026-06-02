// Package connectx contains the adapters that let our existing
// Twirp-shaped service implementations run behind ConnectRPC handlers.
//
// During the Twirp→Connect migration, services in internal/application keep
// returning twirp.Error values. WrapUnary translates those into
// connect.Error via TwirpErrorToConnect so the wire response carries the
// right Connect status codes. Once a service is migrated to native Connect
// signatures, its adapter (and the twirp dependency in that file) can be
// dropped.
package connectx

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/twitchtv/twirp"
)

// TwirpErrorToConnect converts a twirp.Error into the equivalent connect.Error.
// Non-twirp errors are wrapped as CodeInternal so the migration adapter never
// leaks raw Go errors to clients.
func TwirpErrorToConnect(err error) *connect.Error {
	if err == nil {
		return nil
	}
	var twErr twirp.Error
	if !errors.As(err, &twErr) {
		return connect.NewError(connect.CodeInternal, err)
	}
	code := twirpCodeToConnect(twErr.Code())
	cerr := connect.NewError(code, errors.New(twErr.Msg()))
	for k, v := range twErr.MetaMap() {
		// connect.Error metadata is a single value per key; multi-value
		// metadata isn't expressible in twirp so we don't lose anything.
		cerr.Meta().Set(k, v)
	}
	return cerr
}

// twirpCodeToConnect maps every twirp error code to its connect counterpart.
// The mapping mirrors the gRPC status codes both protocols are based on.
//
// twirp's BadRoute and Malformed have no direct gRPC equivalent; they fold
// into Unimplemented and InvalidArgument respectively, matching how grpc-web
// gateways usually surface them.
func twirpCodeToConnect(code twirp.ErrorCode) connect.Code {
	switch code {
	case twirp.Canceled:
		return connect.CodeCanceled
	case twirp.InvalidArgument:
		return connect.CodeInvalidArgument
	case twirp.Malformed:
		return connect.CodeInvalidArgument
	case twirp.DeadlineExceeded:
		return connect.CodeDeadlineExceeded
	case twirp.NotFound:
		return connect.CodeNotFound
	case twirp.BadRoute:
		return connect.CodeUnimplemented
	case twirp.AlreadyExists:
		return connect.CodeAlreadyExists
	case twirp.PermissionDenied:
		return connect.CodePermissionDenied
	case twirp.Unauthenticated:
		return connect.CodeUnauthenticated
	case twirp.ResourceExhausted:
		return connect.CodeResourceExhausted
	case twirp.FailedPrecondition:
		return connect.CodeFailedPrecondition
	case twirp.Aborted:
		return connect.CodeAborted
	case twirp.OutOfRange:
		return connect.CodeOutOfRange
	case twirp.Unimplemented:
		return connect.CodeUnimplemented
	case twirp.Internal:
		return connect.CodeInternal
	case twirp.Unavailable:
		return connect.CodeUnavailable
	case twirp.DataLoss:
		return connect.CodeDataLoss
	case twirp.Unknown:
		return connect.CodeUnknown
	default:
		return connect.CodeUnknown
	}
}

