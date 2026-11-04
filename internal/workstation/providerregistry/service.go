package providerregistry

import (
	"context"
	"net/http"

	"go.orx.me/apps/butter/internal/secretbox"
	ws "go.orx.me/apps/butter/internal/workstation"
)

// configured Work Station service. Provider selection remains centralized here.
type ServiceDeps struct {
	Repo       ws.Repository
	Keyring    *secretbox.Keyring
	CheckHost  func(context.Context, string) error
	Transport  http.RoundTripper
	Checkpoint ws.CheckpointStore
}

func NewService(c Config, d ServiceDeps) (*ws.Service, error) {
	if d.Repo == nil || d.Keyring == nil || d.CheckHost == nil {
		return nil, ws.ErrUnsupported
	}
	provider, err := New(c, d.Transport)
	if err != nil {
		return nil, err
	}
	origin, web := "", ""
	switch c.Kind {
	case GitHub:
		origin, web = "https://api.github.com", "https://github.com"
	case GitLab:
		origin, web = "https://gitlab.com/api/v4", "https://gitlab.com"
	case Linear:
		origin, web = "https://api.linear.app", "https://linear.app"
	case Plane:
		origin, web = c.Host, c.Host
		if origin == "" || origin == planeCloudHost {
			origin, web = planeCloudHost, "https://app.plane.so"
		}
	}
	resolver := func(ctx context.Context, r ws.Record) (ws.ResolvedProvider, error) {
		if r.Source.Provider != "" && r.Source.Provider != string(c.Kind) {
			return ws.ResolvedProvider{}, ws.ErrUnsupported
		}
		if err := d.CheckHost(ctx, r.HostID); err != nil {
			return ws.ResolvedProvider{}, err
		}
		return ws.ResolvedProvider{Adapter: provider, Source: ws.SourceIdentity{Provider: string(c.Kind), Origin: origin}, WebBaseURL: web}, nil
	}
	return &ws.Service{DefaultProvider: string(c.Kind), ResolveProvider: resolver, Repo: d.Repo, Keyring: d.Keyring, Provider: provider, CheckHost: d.CheckHost, Checkpoint: d.Checkpoint}, nil
}
