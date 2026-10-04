package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	teardown := sync.OnceValue(func() error {
		channelCancel()
		// Stop process-owned async dashboard work and Detached AG-UI runs,
		// and wait for each in-flight run to persist its honest FAILED
		// terminal state and, for AG-UI, release its thread's lease. Both
		// stop at once and share one bound, so a stuck run cannot block
		// process exit; anything still QUEUED/RUNNING afterwards is failed as
		// stale once this process's liveness lapses (without Redis, at next
		// startup).
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		var wg sync.WaitGroup
		wg.Go(func() {
			if err := handlers.ShutdownAsync(shutdownCtx); err != nil {
				slog.Warn("async coordinator shutdown incomplete", "err", err)
			}
		})
		wg.Go(func() {
			if err := handlers.ShutdownAGUI(shutdownCtx); err != nil {
				slog.Warn("detached AG-UI runs shutdown incomplete", "err", err)
			}
		})
		wg.Wait()
		return nil
	})

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
		TeardownFunc: []func() error{teardown},
	})

	// Butterfly's App.Run serves until the process ends and never runs its
	// TeardownFunc, so the teardown runs here on SIGTERM or SIGINT, and then
	// the process exits.
	signals, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	go exitAfterTeardown(signals, stopSignals, teardown, os.Exit)

	slog.Info("starting butterfly service", "service", serviceName, "commit", serverBuildCommit())
	svc.Run()
}

// exitAfterTeardown waits for ctx to end — the first SIGTERM or SIGINT in
// main — then runs teardown and exits. stopSignals hands signals back to
// their default handling first, so a second signal ends the process at once
// even while the teardown runs.
func exitAfterTeardown(ctx context.Context, stopSignals func(), teardown func() error, exit func(int)) {
	<-ctx.Done()
	stopSignals()
	slog.Info("shutting down: ending in-flight runs before exit")
	if err := teardown(); err != nil {
		slog.Warn("teardown failed", "err", err)
	}
	exit(0)
}
