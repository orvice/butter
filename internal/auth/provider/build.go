package provider

import "go.orx.me/apps/butter/internal/config"

// BuildRegistry constructs a Registry from AuthConfig. Providers whose
// credentials are missing are silently skipped so partially-configured
// deployments still function for the providers that are configured.
//
// Returns nil when no providers are enabled — the AuthService treats a nil
// registry as "OAuth disabled".
func BuildRegistry(cfg config.AuthConfig) *Registry {
	reg := NewRegistry()
	added := 0
	for name, pc := range cfg.OAuthProviders {
		if !pc.Enabled() {
			continue
		}
		p := buildProvider(name, pc)
		if p == nil {
			continue
		}
		reg.Register(p)
		added++
	}
	if added == 0 {
		return nil
	}
	return reg
}

// buildProvider maps a provider key to its implementation. Add new providers
// here (and create a corresponding file in this package).
func buildProvider(name string, cfg config.OAuthProviderConfig) Provider {
	switch name {
	case "github":
		return NewGitHub(GitHubConfig{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			DisplayName:  cfg.DisplayName,
		})
	default:
		return nil
	}
}
