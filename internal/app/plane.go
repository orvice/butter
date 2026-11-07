package app

import (
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/planeapi"
)

// planeProfileProvider builds the fixed Plane provider profile from startup
// configuration. It is a provider, not a value, because SetupRoutes runs
// before the YAML configuration is loaded into *config.AppConfig.
//
// An invalid profile is an error, not a fallback: Apps store the profile name
// and derive origins from it, so silently substituting the Cloud origins for
// a misconfigured self-hosted profile would send tenant credentials to the
// wrong Plane. The services surface it as FailedPrecondition, and Planes stay
// unmanageable until the deployment is fixed.
func planeProfileProvider(cfg *config.AppConfig) func() (planeapi.Profile, error) {
	return func() (planeapi.Profile, error) {
		raw := cfg.Plane.WithDefaults()
		return planeapi.Build(planeapi.Config{
			Name:           raw.ProfileName,
			APIBaseURL:     raw.APIBaseURL,
			OAuthBaseURL:   raw.OAuthBaseURL,
			WebBaseURL:     raw.WebBaseURL,
			Edition:        raw.Edition,
			SchemaVerified: raw.SchemaVerified,
		})
	}
}
