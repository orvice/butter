package app

import (
	"butterfly.orx.me/core/log"
	"context"
	"errors"
	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/providerregistry"
)

func (h *Handlers) wireWorkStation(result *BootstrapResult) {
	if h.workStationSvcServer == nil || result.WorkStationRepo == nil {
		return
	}
	if err := providerregistry.ValidateConfig(configuredWorkStationProvider(h.cfg)); err != nil {
		log.FromContext(context.Background()).Error("work station defaults rejected", "category", workstation.ReportCategory(err))
		return
	}
	station := &workstation.Service{DefaultProvider: string(configuredWorkStationProvider(h.cfg).Kind), Repo: result.WorkStationRepo, Keyring: secretbox.NewKeyring(result.CryptoKeyRepo), ResolveProvider: application.WorkStationProviders(result.GitHostRepo, result.WorkStationRepo)}
	h.workStationSvcServer.SetService(station, result.WorkspaceRepo)
	if h.workStationHostSvcServer != nil {
		h.workStationHostSvcServer.SetRepo(result.WorkStationRepo)
	}
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

func configuredWorkStationProvider(cfg *config.AppConfig) providerregistry.Config {
	if cfg == nil || cfg.WorkStation.ProviderKind == "" {
		return providerregistry.Config{Kind: providerregistry.GitHub}
	}
	return providerregistry.Config{Kind: providerregistry.Kind(cfg.WorkStation.ProviderKind), Host: cfg.WorkStation.Host, Profile: cfg.WorkStation.Profile}
}
