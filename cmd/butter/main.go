package main

import (
	"context"
	"log/slog"
	"time"

	"butterfly.orx.me/core"
	"butterfly.orx.me/core/app"

	butterapp "go.orx.me/apps/butter/internal/app"
	appconfig "go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
)

const (
	serviceName = "butter"
	h2cAddr     = ":8081"
)

func main() {
	cfg := new(appconfig.AppConfig)
	daemonRegistry := daemon.NewRegistry()
	router, handlers := butterapp.SetupRoutes(cfg, daemonRegistry)

	channelCtx, channelCancel := context.WithCancel(context.Background())

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

				return nil
			},
			func() error {
				_, err := butterapp.StartH2CServer(h2cAddr, router)
				return err
			},
		},
		TeardownFunc: []func() error{
			func() error {
				channelCancel()
				// Stop process-owned async dashboard work and wait for each
				// in-flight run to persist its honest FAILED terminal state.
				// Bounded so a stuck run cannot block process exit; anything
				// still QUEUED/RUNNING afterwards is reconciled at next startup.
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				if err := handlers.ShutdownAsync(shutdownCtx); err != nil {
					slog.Warn("async coordinator shutdown incomplete", "err", err)
				}
				return nil
			},
		},
	})

	slog.Info("starting butterfly service", "service", serviceName, "commit", serverBuildCommit())
	svc.Run()
}
