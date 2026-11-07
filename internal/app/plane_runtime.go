package app

import (
	"context"
	"errors"
	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	planeinstallation "go.orx.me/apps/butter/internal/repo/planeinstallation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type scopedPlaneResolver struct {
	apps          planerepo.Repository
	installations planeinstallation.Repository
}

func (r scopedPlaneResolver) FindScopedApp(ctx context.Context, w, a, i string) (*agentsv1.PlaneApp, error) {
	if w == "" || a == "" || i == "" {
		return nil, errors.New("plane app scope is incomplete")
	}
	app, e := r.apps.GetApp(ctx, w, a)
	if e != nil {
		return nil, e
	}
	inst, e := r.installations.GetScoped(ctx, w, a, i)
	if e != nil {
		return nil, e
	}
	if inst.GetAppId() != a || inst.GetAppInstallationId() != i {
		return nil, planeinstallation.ErrNotFound
	}
	return app, nil
}
