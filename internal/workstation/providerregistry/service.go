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
	return &ws.Service{Repo: d.Repo, Keyring: d.Keyring, Provider: provider, CheckHost: d.CheckHost, Checkpoint: d.Checkpoint}, nil
}
