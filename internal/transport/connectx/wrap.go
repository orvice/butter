package connectx

import (
	"context"

	"connectrpc.com/connect"
)

// UnaryFunc is the shape of a Twirp-style unary RPC method: it takes the raw
// proto request and returns the raw proto response. Existing service
// implementations in internal/application all satisfy this shape.
type UnaryFunc[Req, Res any] func(ctx context.Context, req *Req) (*Res, error)

// WrapUnary lifts a Twirp-style unary method into a ConnectRPC handler
// method. Errors are translated through TwirpErrorToConnect so wire status
// codes line up.
//
// This adapter is intentionally minimal: it does not propagate request/response
// headers or trailers. Services that need to read Connect headers must be
// migrated to native Connect signatures.
func WrapUnary[Req, Res any](fn UnaryFunc[Req, Res]) func(context.Context, *connect.Request[Req]) (*connect.Response[Res], error) {
	return func(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Res], error) {
		resp, err := fn(ctx, req.Msg)
		if err != nil {
			return nil, TwirpErrorToConnect(err)
		}
		return connect.NewResponse(resp), nil
	}
}
