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

	channelCtx, channelCancel := context.WithCancel(context.Background())

	svc := core.New(&app.Config{
		Service: serviceName,
		Config:  cfg,
		Router: func(r *gin.Engine) {
			healthHandler.Register(r)
		},
		InitFunc: []func() error{
			func() error {
				return bootstrap.StartChannels(channelCtx, cfg)
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
