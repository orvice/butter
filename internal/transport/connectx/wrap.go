package connectx

import (
	"context"

	"connectrpc.com/connect"
)

// UnaryFunc is the shape of a unary RPC method in internal/application: it
// takes the raw proto request and returns the raw proto response. Errors are
// already `*connect.Error` values (use connectx.RequiredArgument /
// connectx.NotFound / connect.NewError directly in the service).
type UnaryFunc[Req, Res any] func(ctx context.Context, req *Req) (*Res, error)

// WrapUnary lifts the application-layer method into a ConnectRPC handler
// method. Non-Connect errors are left intact; Connect's handler infrastructure
// will re-wrap any raw `error` with CodeUnknown, so callers don't have to
// translate before returning.
//
// The adapter intentionally does not propagate request headers or trailers.
// Services that need those must be migrated to native Connect signatures
// (`func(ctx, *connect.Request[Req]) (*connect.Response[Res], error)`).
func WrapUnary[Req, Res any](fn UnaryFunc[Req, Res]) func(context.Context, *connect.Request[Req]) (*connect.Response[Res], error) {
	return func(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Res], error) {
		resp, err := fn(ctx, req.Msg)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(resp), nil
	}
}
