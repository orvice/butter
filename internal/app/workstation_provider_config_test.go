package app

import (
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/workstation/providerregistry"
	"testing"
)

func TestConfiguredWorkStationProviderProfiles(t *testing.T) {
	for _, x := range []struct{ kind, host, profile string }{
		{"linear", "https://api.linear.app/graphql", ""}, {"gitlab", "https://gitlab.com", ""}, {"plane", "https://plane.example", "self_hosted"},
	} {
		got := configuredWorkStationProvider(&config.AppConfig{WorkStation: config.WorkStationConfig{ProviderKind: x.kind, Host: x.host, Profile: x.profile}})
		if got.Kind != providerregistry.Kind(x.kind) || got.Host != x.host || got.Profile != x.profile {
			t.Fatalf("got=%+v", got)
		}
	}
	if got := configuredWorkStationProvider(nil); got.Kind != providerregistry.GitHub {
		t.Fatalf("default=%+v", got)
	}
}
