package app

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/repo/linearsetting"
	"go.orx.me/apps/butter/internal/runtime/linearconn"
	linearruntime "go.orx.me/apps/butter/internal/runtime/linear"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// linearSessionLeasePrefix serializes turns within one Linear Agent
	// Session across Pods.
	linearSessionLeasePrefix = "butter:linear:lease:session:"
	// linearSessionLeaseTTL bounds how long a crashed worker blocks one
	// session; the lease is renewed while the turn runs.
	linearSessionLeaseTTL = 5 * time.Minute
	// linearRefreshLeasePrefix serializes token refreshes per installation.
	linearRefreshLeasePrefix = "butter:linear:lease:refresh:"
	linearRefreshLeaseTTL    = 30 * time.Second
)

var errLinearReceiveNotReady = errors.New("linear receive is not ready")

// linearReceiverHolder is the webhook handler's receiver, filled in once
// bootstrap has Redis. Until then deliveries answer 503 so Linear retries.
type linearReceiverHolder struct {
	receiver atomic.Pointer[linearruntime.Receiver]
}

func (h *linearReceiverHolder) FindApp(ctx context.Context, appID string) (*agentsv1.LinearApp, error) {
	r := h.receiver.Load()
	if r == nil {
		return nil, errLinearReceiveNotReady
	}
	return r.FindApp(ctx, appID)
}

func (h *linearReceiverHolder) Deliver(ctx context.Context, app *agentsv1.LinearApp, header http.Header, body []byte) (linearruntime.Decision, error) {
	r := h.receiver.Load()
	if r == nil {
		return "", errLinearReceiveNotReady
	}
	return r.Deliver(ctx, app, header, body)
}

// wireLinearRuntime starts the Linear receive path. It exists only with
// Redis: the webhook treats Redis Streams as durable queue infrastructure,
// and without it the route keeps answering 503.
func (h *Handlers) wireLinearRuntime(result *BootstrapResult, keyring *secretbox.Keyring) {
	queue := linearruntime.NewQueue(result.Redis)
	if !queue.Available() {
		return
	}
	if h.linearAppSvcServer != nil {
		h.linearAppSvcServer.SetQueueProbe(queue)
	}
	h.linearReceiver.receiver.Store(linearruntime.NewReceiver(result.LinearRepo, keyring, queue))
	if result.RunnerSvc == nil {
		return
	}

	instanceID := uuid.NewString()
	client := linearapi.New(nil, linearapi.DefaultEndpoints)
	tokens := linearconn.NewTokenSource(result.LinearRepo, keyring)
	tokens.SetLinearClient(client)
	tokens.SetRefreshGuard(sessionguard.NewRedis(result.Redis, instanceID, linearRefreshLeasePrefix, linearRefreshLeaseTTL))

	orchestrator := linearruntime.NewOrchestrator(result.LinearRepo, result.RunnerSvc, tokens, client)
	orchestrator.SetSessionGuard(sessionguard.NewRedis(result.Redis, instanceID, linearSessionLeasePrefix, linearSessionLeaseTTL))
	orchestrator.SetExternalBaseURL(linearExternalBaseURL(h.cfg, result.LinearSettingRepo))

	h.linearWorker = linearruntime.NewWorker(queue, orchestrator, instanceID, linearruntime.DefaultConcurrency)
	if err := h.linearWorker.Start(context.Background()); err != nil {
		log.FromContext(context.Background()).Error("linear worker did not start", "err", err)
		h.linearWorker = nil
	}
}

// linearExternalBaseURL is where links back to the dashboard point: the
// configured dashboard, or else the Linear public base URL, which serves the
// dashboard when both share an origin.
func linearExternalBaseURL(cfg *config.AppConfig, settings linearsetting.Repository) func(context.Context) string {
	return func(ctx context.Context) string {
		if cfg != nil {
			if base := strings.TrimSpace(cfg.MCPOAuth.DashboardBaseURL); base != "" {
				return base
			}
		}
		if settings == nil {
			return ""
		}
		stored, err := settings.Get(ctx)
		if err != nil {
			return ""
		}
		return stored.GetPublicBaseUrl()
	}
}

// StopLinearRuntime halts the Linear worker for a graceful exit.
func (h *Handlers) StopLinearRuntime() {
	if h != nil && h.linearWorker != nil {
		h.linearWorker.Stop()
	}
}
