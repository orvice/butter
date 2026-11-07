package app

import (
	"context"
	planeruntime "go.orx.me/apps/butter/internal/runtime/plane"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"net/http"
	"sync/atomic"
)

type planeReceiverHolder struct {
	receiver atomic.Pointer[planeruntime.Receiver]
}

func (h *planeReceiverHolder) FindApp(ctx context.Context, id string) (*agentsv1.PlaneApp, error) {
	r := h.receiver.Load()
	if r == nil {
		return nil, context.Canceled
	}
	return r.FindApp(ctx, id)
}
func (h *planeReceiverHolder) FindScopedApp(ctx context.Context, workspaceID, appID, installationID string) (*agentsv1.PlaneApp, error) {
	r := h.receiver.Load()
	if r == nil {
		return nil, context.Canceled
	}
	return r.FindScopedApp(ctx, workspaceID, appID, installationID)
}
func (h *planeReceiverHolder) Deliver(ctx context.Context, a *agentsv1.PlaneApp, hdr http.Header, b []byte) (planeruntime.Decision, error) {
	r := h.receiver.Load()
	if r == nil {
		return "", context.Canceled
	}
	return r.Deliver(ctx, a, hdr, b)
}
