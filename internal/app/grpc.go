package app

import (
	"fmt"
	"net"
	"strings"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"go.orx.me/apps/butter/internal/authn"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const defaultGRPCPort = 9090

// BuildGRPCServer constructs the gRPC server and registers every service
// implementation, but does not bind a port. Splitting build from listen
// lets the Gin router wrap the same server with grpc-web during route
// registration (which happens before InitFunc/listener startup).
func BuildGRPCServer(cfg *config.AppConfig, registry *daemon.Registry, handlers *Handlers) *grpc.Server {
	opts := []grpc.ServerOption{}
	if handlers != nil && handlers.Resolver() != nil {
		public := authn.PublicMethods()
		opts = append(opts,
			grpc.UnaryInterceptor(authn.UnaryServerInterceptor(handlers.Resolver(), public)),
			grpc.StreamInterceptor(authn.StreamServerInterceptor(handlers.Resolver(), public)),
		)
	}

	srv := grpc.NewServer(opts...)

	// Daemon connector keeps its existing token-based auth contract; the
	// interceptor's PublicMethods list lets it through to the handler.
	daemonHandler := daemon.NewGRPCHandler(registry, func() string {
		if cfg == nil {
			return ""
		}
		return cfg.APIToken
	})
	agentsv1.RegisterDaemonConnectorServiceServer(srv, daemonHandler)

	// User-facing services share the auth interceptor and reuse the same
	// implementations as the Twirp handlers.
	handlers.RegisterGRPCServices(srv)

	return srv
}

// StartGRPCListener binds the native gRPC port for non-web clients
// (daemon connectors, server-to-server gRPC). The browser-facing
// grpc-web handler is mounted on the HTTP listener via NewGRPCWebHandler
// and does not need this port to be reachable.
func StartGRPCListener(cfg *config.AppConfig, srv *grpc.Server) (net.Listener, error) {
	port := cfg.GRPCPort
	if port == 0 {
		port = defaultGRPCPort
	}
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("grpc listen on %s: %w", addr, err)
	}
	slog.Info("gRPC listener bound", "addr", addr)
	return lis, nil
}

// grpcWebDispatcher routes browser grpc-web traffic to the wrapped gRPC
// server. We match any path that looks like a gRPC procedure
// (`/<package>.<Service>/<Method>`) plus its CORS preflight, and short-
// circuit the rest of the Gin chain so the auth middleware doesn't try to
// re-validate the request — the gRPC server runs its own interceptor.
func grpcWebDispatcher(wrapped *grpcweb.WrappedGrpcServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isGRPCWebRequest(c, wrapped) {
			c.Next()
			return
		}
		wrapped.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func isGRPCWebRequest(c *gin.Context, wrapped *grpcweb.WrappedGrpcServer) bool {
	if wrapped.IsGrpcWebRequest(c.Request) || wrapped.IsAcceptableGrpcCorsRequest(c.Request) {
		return true
	}
	// Native fetch from grpc-web JS clients sets Content-Type to
	// application/grpc-web* even when the URL doesn't look like an RPC
	// path (e.g. some proxies rewrite). Fall back to a strict path check
	// for `/<pkg>.<Service>/<Method>` rooted at the agents.v1 namespace
	// so we never swallow Twirp/REST routes by accident.
	path := c.Request.URL.Path
	return strings.HasPrefix(path, "/agents.v1.") && strings.Count(path, "/") == 2
}

// NewGRPCWebHandler wraps the gRPC server so it can be served alongside
// Twirp from the same HTTP listener. CORS is configured to accept the
// custom Authorization and X-Workspace-ID headers the dashboard sends.
func NewGRPCWebHandler(srv *grpc.Server) *grpcweb.WrappedGrpcServer {
	return grpcweb.WrapServer(srv,
		grpcweb.WithOriginFunc(func(string) bool { return true }),
		grpcweb.WithAllowedRequestHeaders([]string{
			"Authorization",
			"X-Workspace-ID",
			"X-Grpc-Web",
			"X-User-Agent",
			"Content-Type",
			"Grpc-Timeout",
		}),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
	)
}
