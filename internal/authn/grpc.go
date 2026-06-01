package authn

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// metadataSource adapts incoming gRPC metadata to HeaderSource. gRPC
// metadata keys are always lowercased by the framework, so we lower-case
// the canonical header name on lookup.
type metadataSource struct{ md metadata.MD }

func (m metadataSource) Get(name string) string {
	if m.md == nil {
		return ""
	}
	values := m.md.Get(strings.ToLower(name))
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// PublicMethods returns the gRPC procedure names that should bypass
// authentication. Keep this in sync with isPublicPath in the HTTP adapter
// — these are the gRPC equivalents of the public Twirp routes.
func PublicMethods() map[string]struct{} {
	return map[string]struct{}{
		"/agents.v1.AuthService/Login":              {},
		"/agents.v1.AuthService/ListOAuthProviders": {},
		"/agents.v1.AuthService/BeginOAuthFlow":     {},
		"/agents.v1.AuthService/CompleteOAuthFlow":  {},
		// DaemonConnectorService is an internal/daemon-only service that
		// runs on a separate, non-web listener and authenticates via its
		// own bearer token in metadata. We treat its methods as public to
		// the user-facing interceptor so it can opt out cleanly when the
		// service is multiplexed onto the same gRPC server.
		"/agents.v1.DaemonConnectorService/Connect": {},
	}
}

// UnaryServerInterceptor authenticates each unary RPC using the resolver.
// Public methods (login, OAuth) bypass the resolver entirely. On
// authentication failure the interceptor returns codes.Unauthenticated.
func UnaryServerInterceptor(r *Resolver, publicMethods map[string]struct{}) grpc.UnaryServerInterceptor {
	if publicMethods == nil {
		publicMethods = PublicMethods()
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		res := r.Resolve(ctx, metadataSource{md})
		if res.Outcome != OutcomeAuthenticated {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		return handler(res.Ctx, req)
	}
}

// StreamServerInterceptor authenticates each streaming RPC using the
// resolver and wraps the server stream so handlers see the resolved
// context.
func StreamServerInterceptor(r *Resolver, publicMethods map[string]struct{}) grpc.StreamServerInterceptor {
	if publicMethods == nil {
		publicMethods = PublicMethods()
	}
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(srv, ss)
		}
		md, _ := metadata.FromIncomingContext(ss.Context())
		res := r.Resolve(ss.Context(), metadataSource{md})
		if res.Outcome != OutcomeAuthenticated {
			return status.Error(codes.Unauthenticated, "unauthorized")
		}
		return handler(srv, &authedServerStream{ServerStream: ss, ctx: res.Ctx})
	}
}

type authedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authedServerStream) Context() context.Context { return s.ctx }
