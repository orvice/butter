package main

import (
	"context"
	"log/slog"

	"butterfly.orx.me/core"
	"butterfly.orx.me/core/app"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/bootstrap"
	appconfig "go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/service"
)

const serviceName = "butter"

func main() {
	cfg := new(appconfig.AppConfig)
	healthRepo := repo.NewHealthRepository()
	healthService := service.NewHealthService(healthRepo, cfg)
	healthHandler := httpHandler.NewHealthHandler(healthService)
	a2aHandler := httpHandler.NewA2AHandler(cfg)

	channelCtx, channelCancel := context.WithCancel(context.Background())

	svc := core.New(&app.Config{
		Namespace: "ai",
		Service:   serviceName,
		Config:    cfg,
		Router: func(r *gin.Engine) {
			healthHandler.Register(r)
			a2aHandler.Register(r)
		},
		InitFunc: []func() error{
			func() error {
				runnerSvc, err := bootstrap.StartChannels(channelCtx, cfg)
				if err != nil {
					return err
				}
				if runnerSvc != nil {
					a2aHandler.SetRunnerService(runnerSvc)
				}
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
