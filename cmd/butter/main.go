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
	"go.orx.me/apps/butter/internal/repo/configstore"
	"go.orx.me/apps/butter/internal/service"
	"go.orx.me/apps/butter/internal/service/configapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const serviceName = "butter"

func main() {
	cfg := new(appconfig.AppConfig)
	healthRepo := repo.NewHealthRepository()
	healthService := service.NewHealthService(healthRepo, cfg)
	healthHandler := httpHandler.NewHealthHandler(healthService)
	a2aHandler := httpHandler.NewA2AHandler(cfg)

	// Config store and Twirp services
	cfgStore := configstore.New()
	agentTwirp := agentsv1.NewAgentServiceServer(configapi.NewAgentServiceServer(cfgStore))
	mcpTwirp := agentsv1.NewMCPServerServiceServer(configapi.NewMCPServerServiceServer(cfgStore))
	remoteTwirp := agentsv1.NewRemoteAgentServiceServer(configapi.NewRemoteAgentServiceServer(cfgStore))

	channelCtx, channelCancel := context.WithCancel(context.Background())

	svc := core.New(&app.Config{
		Namespace: "ai",
		Service:   serviceName,
		Config:    cfg,
		Router: func(r *gin.Engine) {
			healthHandler.Register(r)
			a2aHandler.Register(r)

			// Mount Twirp handlers — Twirp includes /twirp prefix by default
			r.Any(agentTwirp.PathPrefix()+"*path", gin.WrapH(agentTwirp))
			r.Any(mcpTwirp.PathPrefix()+"*path", gin.WrapH(mcpTwirp))
			r.Any(remoteTwirp.PathPrefix()+"*path", gin.WrapH(remoteTwirp))
		},
		InitFunc: []func() error{
			func() error {
				// Seed config store from YAML config
				cfgStore.Seed(cfg.Agents, cfg.MCPServerConfigs, cfg.RemoteAgents)

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
