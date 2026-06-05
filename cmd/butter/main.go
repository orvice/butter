package main

import (
	"context"
	"log/slog"

	"butterfly.orx.me/core"
	"butterfly.orx.me/core/app"
	"google.golang.org/grpc"

	butterapp "go.orx.me/apps/butter/internal/app"
	appconfig "go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
)

const serviceName = "butter"

func main() {
	cfg := new(appconfig.AppConfig)
	daemonRegistry := daemon.NewRegistry()
	router, handlers := butterapp.SetupRoutes(cfg, daemonRegistry)

	channelCtx, channelCancel := context.WithCancel(context.Background())

	var grpcServer *grpc.Server

	svc := core.New(&app.Config{
		Namespace: "ai",
		Service:   serviceName,
		Config:    cfg,
		Router:    router,
		InitFunc: []func() error{
			func() error {
				if err := handlers.SeedConfig(channelCtx, cfg); err != nil {
					return err
				}
				result, err := butterapp.StartChannels(channelCtx, cfg, handlers.AgentRepo(), handlers.ChannelRepo(), handlers.NotifyGroupRepo(), daemonRegistry)
				if err != nil {
					return err
				}
				handlers.Wire(result)

				// Start gRPC server for daemon connections.
				srv, lis, err := butterapp.SetupGRPCServer(cfg, daemonRegistry, result.APITokenRepo, handlers.ConfigStore())
				if err != nil {
					return err
				}
				grpcServer = srv
				go func() {
					slog.Info("gRPC server starting", "addr", lis.Addr().String())
					if err := srv.Serve(lis); err != nil {
						slog.Error("gRPC server error", "err", err)
					}
				}()

				return nil
			},
		},
		TeardownFunc: []func() error{
			func() error {
				if grpcServer != nil {
					grpcServer.GracefulStop()
				}
				channelCancel()
				return nil
			},
		},
	})

	slog.Info("starting butterfly service", "service", serviceName)
	svc.Run()
}
