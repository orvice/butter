package main

import (
	"context"
	"log/slog"

	"butterfly.orx.me/core"
	"butterfly.orx.me/core/app"

	butterapp "go.orx.me/apps/butter/internal/app"
	appconfig "go.orx.me/apps/butter/internal/config"
)

const serviceName = "butter"

func main() {
	cfg := new(appconfig.AppConfig)
	router, handlers := butterapp.SetupRoutes(cfg)

	channelCtx, channelCancel := context.WithCancel(context.Background())

	svc := core.New(&app.Config{
		Namespace: "ai",
		Service:   serviceName,
		Config:    cfg,
		Router:    router,
		InitFunc: []func() error{
			func() error {
				handlers.SeedConfig(cfg)
				result, err := butterapp.StartChannels(channelCtx, cfg, handlers.AgentRepo())
				if err != nil {
					return err
				}
				handlers.Wire(result)
				return nil
			},
		},
		TeardownFunc: []func() error{
			func() error {
				channelCancel()
				return nil
			},
		},
	})

	slog.Info("starting butterfly service", "service", serviceName)
	svc.Run()
}
