// Package providerregistry is the explicit, fail-closed provider/profile
// selector. It is separate from workstation to avoid adapter import cycles.
package providerregistry

import (
	"net/http"
	"net/url"
	"strings"

	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/github"
	"go.orx.me/apps/butter/internal/workstation/gitlab"
	"go.orx.me/apps/butter/internal/workstation/linear"
	"go.orx.me/apps/butter/internal/workstation/plane"
)

type Kind string

const (
	GitHub Kind = "github"
	GitLab Kind = "gitlab"
	Linear Kind = "linear"
	Plane  Kind = "plane"
)

type Config struct {
	Kind    Kind
	Host    string
	Profile string
}

const (
	githubHost     = "https://api.github.com"
	gitlabHost     = "https://gitlab.com"
	linearHost     = "https://api.linear.app/graphql"
	planeCloudHost = "https://api.plane.so"
)

// New validates the complete source identity before constructing an adapter.
// An empty host selects the provider's approved cloud host; self-hosted Plane
// requires both an explicit host and the self_hosted profile.
func New(c Config, transport http.RoundTripper) (ws.Provider, error) {
	host := strings.TrimRight(c.Host, "/")
	switch c.Kind {
	case GitHub:
		if host != "" && host != githubHost || c.Profile != "" {
			return nil, ws.ErrInvalid
		}
		return github.New(transport), nil
	case GitLab:
		if host != "" && host != gitlabHost || c.Profile != "" {
			return nil, ws.ErrInvalid
		}
		return gitlab.New(transport), nil
	case Linear:
		if host != "" && host != linearHost || c.Profile != "" {
			return nil, ws.ErrInvalid
		}
		return linear.New(transport), nil
	case Plane:
		if c.Profile == string(plane.SelfHosted) {
			if host == "" || !validSelfHosted(host) {
				return nil, ws.ErrInvalid
			}
			return plane.NewSelfHosted(host, transport)
		}
		if c.Profile != "" && c.Profile != string(plane.Cloud) || host != "" && host != planeCloudHost {
			return nil, ws.ErrInvalid
		}
		return plane.New(transport)
	default:
		return nil, ws.ErrUnsupported
	}
}

func validSelfHosted(host string) bool {
	u, err := url.Parse(host)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && u.RawQuery == "" && u.Fragment == ""
}
