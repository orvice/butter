package main

import (
	"context"
	"log/slog"

	"butterfly.orx.me/core"
	"butterfly.orx.me/core/app"

	"go.orx.me/apps/butter/internal/bootstrap"
	appconfig "go.orx.me/apps/butter/internal/config"
)

const serviceName = "butter"

func main() {
	cfg := new(appconfig.AppConfig)
	router, handlers := bootstrap.SetupRoutes(cfg)

	channelCtx, channelCancel := context.WithCancel(context.Background())

	svc := core.New(&app.Config{
		Namespace: "ai",
		Service:   serviceName,
		Config:    cfg,
		Router:    router,
		InitFunc: []func() error{
			func() error {
				handlers.SeedConfig(cfg)
				result, err := bootstrap.StartChannels(channelCtx, cfg)
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
