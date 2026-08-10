package app

import (
	"context"
	"time"

	"butterfly.orx.me/core/log"
	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
)

const defaultReconcileInterval = 5 * time.Minute

// syncTrigger triggers sync + publication for a workspace.
type syncTrigger interface {
	TriggerSyncAndPublish(ctx context.Context, ws string) error
}

// Reconciler periodically polls all workspace bindings and triggers a
// sync + publish for each one whose remote HEAD may have changed. This
// catches missed webhooks and ensures staleness stays bounded.
type Reconciler struct {
	repo     repobindingrepo.Repository
	trigger  syncTrigger
	interval time.Duration
	cancel   context.CancelFunc
}

func NewReconciler(repo repobindingrepo.Repository, trigger syncTrigger, interval time.Duration) *Reconciler {
	if interval <= 0 {
		interval = defaultReconcileInterval
	}
	return &Reconciler{
		repo:     repo,
		trigger:  trigger,
		interval: interval,
	}
}

// Start runs one reconciliation immediately, then continues periodically in a
// background goroutine.
func (r *Reconciler) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	go r.loop(ctx)
}

// Stop cancels the reconciliation loop.
func (r *Reconciler) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Reconciler) loop(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Info("repository reconciler started", "interval", r.interval)
	r.reconcileAll(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("repository reconciler stopped")
			return
		case <-ticker.C:
			r.reconcileAll(ctx)
		}
	}
}

func (r *Reconciler) reconcileAll(ctx context.Context) {
	logger := log.FromContext(ctx)

	bindings, err := r.repo.ListAcrossWorkspaces(ctx)
	if err != nil {
		logger.Error("reconciler: list bindings failed", "err", err)
		return
	}

	for _, b := range bindings {
		ws := b.GetWorkspaceId()
		if ws == "" {
			continue
		}
		if !b.GetCredentialSet() {
			continue
		}

		if err := r.trigger.TriggerSyncAndPublish(ctx, ws); err != nil {
			logger.Warn("reconciler: sync failed", "workspace_id", ws, "err", err)
		}
	}
}
