package app

import (
	"butterfly.orx.me/core/log"
	"context"
	"errors"
	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/providerregistry"
)

func (h *Handlers) wireWorkStation(result *BootstrapResult) {
	if h.workStationSvcServer == nil || result.WorkStationRepo == nil {
		return
	}
	station, err := providerregistry.NewService(providerregistry.Config{Kind: providerregistry.GitHub}, providerregistry.ServiceDeps{Repo: result.WorkStationRepo, Keyring: secretbox.NewKeyring(result.CryptoKeyRepo), CheckHost: application.WorkStationHostCheck(result.GitHostRepo)})
	if err != nil {
		log.FromContext(context.Background()).Error("work station provider wiring failed", "error", err)
		return
	}
	h.workStationSvcServer.SetService(station, result.WorkspaceRepo)
	h.gitHostSvcServer.SetWorkStationHostGuard(result.WorkStationRepo.HostInUse)
	ctx, cancel := context.WithCancel(context.Background())
	h.stopWorkStation = cancel
	h.workStationDone = make(chan struct{})
	go func() {
		defer close(h.workStationDone)
		station.Run(ctx, func(err error) {
			operation := "unknown"
			var failure *workstation.RuntimeError
			if errors.As(err, &failure) {
				operation = failure.Operation
			}
			log.FromContext(ctx).Error("work station reconciliation failed", "operation", operation, "category", workstation.ReportCategory(err), "cause_type", workstation.CauseType(err))
		})
	}()
}

// ShutdownWorkStation cancels and joins the reconciler within the server's
// shutdown deadline. Durable intents and expiring claims survive process exit.
func (h *Handlers) ShutdownWorkStation(ctx context.Context) error {
	if h == nil || h.stopWorkStation == nil {
		return nil
	}
	h.stopWorkStation()
	select {
	case <-h.workStationDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
