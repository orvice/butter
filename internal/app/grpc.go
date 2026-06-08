package app

import (
	"fmt"
	"net"
	"net/http"

	"log/slog"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const defaultGRPCPort = 9090

// SetupGRPCServer creates an HTTP/2 (h2c) server exposing the
// DaemonConnectorService over ConnectRPC. It returns the server and a listener;
// the caller is responsible for calling srv.Serve(lis).
func SetupGRPCServer(cfg *config.AppConfig, registry *daemon.Registry, tokenRepo apitoken.Repository, daemonRepo configrepo.DaemonRuntimeRepository) (*http.Server, net.Listener, error) {
	port := cfg.GRPCPort
	if port == 0 {
		port = defaultGRPCPort
	}

	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("grpc listen on %s: %w", addr, err)
	}

	handler := daemon.NewGRPCHandler(registry, tokenRepo, daemonRepo)
	mux := http.NewServeMux()
	path, h := agentsv1connect.NewDaemonConnectorServiceHandler(handler)
	mux.Handle(path, h)

	srv := &http.Server{
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	slog.Info("daemon connect server configured", "addr", addr)
	return srv, lis, nil
}
