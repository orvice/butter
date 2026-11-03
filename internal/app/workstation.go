package app

import (
	"butterfly.orx.me/core/log"
	"context"
	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/workstation"
	workstationgithub "go.orx.me/apps/butter/internal/workstation/github"
)

func (h *Handlers) wireWorkStation(result *BootstrapResult) {
	if h.workStationSvcServer == nil || result.WorkStationRepo == nil {
		return
	}
	station := &workstation.Service{Repo: result.WorkStationRepo, Keyring: secretbox.NewKeyring(result.CryptoKeyRepo), Provider: workstationgithub.New(nil), CheckHost: application.WorkStationHostCheck(result.GitHostRepo)}
	h.workStationSvcServer.SetService(station, result.WorkspaceRepo)
	h.gitHostSvcServer.SetWorkStationHostGuard(result.WorkStationRepo.HostInUse)
	ctx, cancel := context.WithCancel(context.Background())
	h.stopWorkStation = cancel
	h.workStationDone = make(chan struct{})
	go func() {
		defer close(h.workStationDone)
		station.Run(ctx, func(error) { log.FromContext(ctx).Error("work station reconciliation failed") })
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
