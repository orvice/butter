package main

import (
	"log/slog"

	"butterfly.orx.me/core"
	"butterfly.orx.me/core/app"
	"github.com/gin-gonic/gin"
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

	svc := core.New(&app.Config{
		Service: serviceName,
		Config:  cfg,
		Router: func(r *gin.Engine) {
			registerRoutes(r, healthHandler)
		},
	})

	slog.Info("starting butterfly service", "service", serviceName)
	svc.Run()
}

func registerRoutes(r *gin.Engine, healthHandler *httpHandler.HealthHandler) {
	healthHandler.Register(r)
}
