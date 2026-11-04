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

// ValidateConfig retains fail-closed validation for legacy deployment defaults.
// Workspace connections resolve their actual endpoints from approved registries.
func ValidateConfig(c Config) error {
	host := strings.TrimRight(c.Host, "/")
	switch c.Kind {
	case GitHub:
		if (host != "" && host != githubHost) || c.Profile != "" {
			return ws.ErrInvalid
		}
	case GitLab:
		if (host != "" && host != gitlabHost) || c.Profile != "" {
			return ws.ErrInvalid
		}
	case Linear:
		if (host != "" && host != linearHost) || c.Profile != "" {
			return ws.ErrInvalid
		}
	case Plane:
		if c.Profile == "self_hosted" {
			if host == "" || !validSelfHosted(host) {
				return ws.ErrInvalid
			}
			if _, err := ws.ApprovedURL(host); err != nil {
				return err
			}
		} else if (c.Profile != "" && c.Profile != "cloud") || (host != "" && host != planeCloudHost) {
			return ws.ErrInvalid
		}
	default:
		return ws.ErrUnsupported
	}
	return nil
}

type Config struct {
	Kind          Kind
	Host          string
	WorkspaceSlug string
	WebBaseURL    string
	Profile       string
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
	if err := ValidateConfig(c); err != nil {
		return nil, err
	}
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
		return gitlab.New("https://gitlab.com/api/v4", "https://gitlab.com", transport)
	case Linear:
		if host != "" && host != linearHost || c.Profile != "" {
			return nil, ws.ErrInvalid
		}
		return linear.New(transport), nil
	case Plane:
		if c.Profile == "self_hosted" {
			if host == "" || !validSelfHosted(host) {
				return nil, ws.ErrInvalid
			}
			return plane.New(host, host, c.WorkspaceSlug, transport)
		}
		if c.Profile != "" && c.Profile != "cloud" || host != "" && host != planeCloudHost {
			return nil, ws.ErrInvalid
		}
		return plane.New(planeCloudHost, "https://app.plane.so", c.WorkspaceSlug, transport)
	default:
		return nil, ws.ErrUnsupported
	}
}

func validSelfHosted(host string) bool {
	u, err := url.Parse(host)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && u.RawQuery == "" && u.Fragment == ""
}
