package app

import (
	"fmt"
	"net"

	"log/slog"

	"google.golang.org/grpc"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const defaultGRPCPort = 9090

// SetupGRPCServer creates a gRPC server with the DaemonConnectorService service
// registered. It returns the server and a listener. The caller is responsible
// for calling srv.Serve(lis).
func SetupGRPCServer(cfg *config.AppConfig, registry *daemon.Registry, tokenRepo apitoken.Repository, daemonRepo configrepo.DaemonConfigRepository) (*grpc.Server, net.Listener, error) {
	port := cfg.GRPCPort
	if port == 0 {
		port = defaultGRPCPort
	}

	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("grpc listen on %s: %w", addr, err)
	}

	srv := grpc.NewServer()
	handler := daemon.NewGRPCHandler(registry, tokenRepo, daemonRepo)
	agentsv1.RegisterDaemonConnectorServiceServer(srv, handler)

	slog.Info("gRPC server configured", "addr", addr)
	return srv, lis, nil
}
