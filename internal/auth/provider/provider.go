// Package provider defines the OAuth login provider abstraction used by the
// AuthService. Each supported provider (GitHub, Google, …) implements the
// Provider interface and is registered with a Registry. The AuthService
// dispatches BeginOAuthFlow / CompleteOAuthFlow by provider name.
package provider

import (
	"context"
	"errors"
)

// ErrUnknownProvider is returned by Registry.Get when no provider is
// registered under the given name.
var ErrUnknownProvider = errors.New("oauth provider not configured")

// Claims is the normalized set of user attributes a provider returns after a
// successful code exchange. Only ExternalID is strictly required; the rest
// are best-effort.
type Claims struct {
	// Provider is the registry name of the provider that produced these
	// claims (e.g. "github").
	Provider string
	// ExternalID is the provider-issued account ID. Stable across email or
	// username changes.
	ExternalID string
	// Login is the provider-side username (e.g. GitHub login). Optional.
	Login string
	// Name is the provider-side display name. Optional.
	Name string
	// Email is the primary verified email returned by the provider. May be
	// empty if the user has hidden it.
	Email string
	// AvatarURL is a URL to the user's avatar image. Optional.
	AvatarURL string
}

// Provider abstracts a single OAuth login provider.
type Provider interface {
	// Name returns the registry key for this provider (e.g. "github").
	Name() string
	// DisplayName is the human-readable label shown on the login page.
	DisplayName() string
	// AuthorizeURL builds the authorization URL the user is redirected to.
	// state is an opaque CSRF token issued by the AuthService.
	AuthorizeURL(state string) string
	// Exchange swaps an authorization code for normalized Claims. Implementations
	// must validate that code came from the provider and surface transport
	// errors verbatim — the AuthService logs them.
	Exchange(ctx context.Context, code string) (*Claims, error)
}

// Registry holds the configured providers and looks them up by name. The
// zero value is unusable; construct via NewRegistry.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register adds p to the registry under its Name().
func (r *Registry) Register(p Provider) {
	if p == nil || p.Name() == "" {
		return
	}
	r.providers[p.Name()] = p
}

// Get returns the provider registered under name, or ErrUnknownProvider.
func (r *Registry) Get(name string) (Provider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, ErrUnknownProvider
	}
	return p, nil
}

// List returns the registered providers in a deterministic order
// (insertion-order is not preserved; callers that need stability should sort
// the result themselves).
func (r *Registry) List() []Provider {
	out := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	return out
}
