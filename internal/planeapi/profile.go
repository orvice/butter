// Package planeapi holds the platform-level Plane provider profile.
//
// Spec #487 delivers Plane support against one verified provider profile
// first, and #497 stops at registering the App resource. The profile — the
// REST, OAuth and web origins plus the edition the deployment talks to — is
// fixed by startup configuration on purpose:
//
//   - a workspace can never point its App at an arbitrary origin, so Plane
//     credentials and callbacks cannot be exfiltrated or redirected;
//   - an App's provider profile is immutable, so an app registered against
//     Cloud cannot later be considered installed against another instance;
//   - the deployment can refuse to start on a half-configured or non-HTTPS
//     profile instead of discovering it on the first OAuth attempt (#498).
//
// It deliberately contains no client and no OAuth flow: those belong to
// #498. What it does provide is the single place where the profile's URLs
// are parsed and checked, which the App service uses to build the public
// callback/webhook URLs.
package planeapi

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// Edition identifies the Plane distribution a profile describes.
type Edition string

const (
	// EditionCloud is Plane Cloud (api.plane.so, app.plane.so).
	EditionCloud Edition = "cloud"
	// EditionSelfHosted is a self-hosted Plane deployment.
	EditionSelfHosted Edition = "selfhosted"
)

var profileNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9])?$`)

// Config is the raw, operator-supplied description of one provider profile.
type Config struct {
	// Name is the profile name an App stores, e.g. "cloud".
	Name string
	// APIBaseURL is the Plane REST origin, e.g. "https://api.plane.so".
	APIBaseURL string
	// OAuthBaseURL is the OAuth origin. Defaults to APIBaseURL.
	OAuthBaseURL string
	// WebBaseURL is the browser origin used only to render links.
	WebBaseURL string
	// Edition is the Plane distribution, "cloud" or "selfhosted".
	Edition string
	// SchemaVerified records whether #496 confirmed this profile's protocol
	// against a real instance. Gate A is not passed yet, so the fixed Cloud
	// profile ships with false and the dashboard must not claim otherwise.
	SchemaVerified bool
}

// DefaultCloudConfig returns the first-version profile: Plane Cloud.
func DefaultCloudConfig() Config {
	return Config{
		Name:           "cloud",
		APIBaseURL:     "https://api.plane.so",
		OAuthBaseURL:   "https://api.plane.so",
		WebBaseURL:     "https://app.plane.so",
		Edition:        string(EditionCloud),
		SchemaVerified: false,
	}
}

// Profile is a validated provider profile.
type Profile struct {
	Name       string
	APIBaseURL string
	OAuthURL   string
	WebURL     string
	Edition    Edition
	Verified   bool
}

// Build validates cfg and returns the profile. It fails closed: an invalid
// profile must stop the deployment instead of silently falling back to a
// default origin, because Apps store the profile name and derive URLs from
// it.
func Build(cfg Config) (Profile, error) {
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		return Profile{}, fmt.Errorf("plane profile name is required")
	}
	if !profileNamePattern.MatchString(name) {
		return Profile{}, fmt.Errorf("plane profile name %q must be lowercase alphanumeric with dashes", name)
	}
	apiBase, err := normaliseOrigin(cfg.APIBaseURL, "api_base_url")
	if err != nil {
		return Profile{}, err
	}
	oauthBase := apiBase
	if strings.TrimSpace(cfg.OAuthBaseURL) != "" {
		oauthBase, err = normaliseOrigin(cfg.OAuthBaseURL, "oauth_base_url")
		if err != nil {
			return Profile{}, err
		}
	}
	webBase := ""
	if strings.TrimSpace(cfg.WebBaseURL) != "" {
		webBase, err = normaliseOrigin(cfg.WebBaseURL, "web_base_url")
		if err != nil {
			return Profile{}, err
		}
	}
	edition := Edition(strings.ToLower(strings.TrimSpace(cfg.Edition)))
	switch edition {
	case "":
		edition = EditionCloud
	case EditionCloud, EditionSelfHosted:
	default:
		return Profile{}, fmt.Errorf("plane edition %q must be %q or %q", cfg.Edition, EditionCloud, EditionSelfHosted)
	}
	return Profile{
		Name:       name,
		APIBaseURL: apiBase,
		OAuthURL:   oauthBase,
		WebURL:     webBase,
		Edition:    edition,
		Verified:   cfg.SchemaVerified,
	}, nil
}

// normaliseOrigin accepts an absolute https origin, plus http on loopback for
// local development, and drops a trailing slash. A path, query, fragment or
// userinfo is rejected: a profile origin is not a URL to arbitrary resources.
func normaliseOrigin(raw, field string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("plane %s is required", field)
	}
	u, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("plane %s %q is not a URL: %w", field, raw, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("plane %s %q must use https", field, raw)
	}
	if u.Host == "" {
		return "", fmt.Errorf("plane %s %q has no host", field, raw)
	}
	if u.User != nil {
		return "", fmt.Errorf("plane %s %q must not carry credentials", field, raw)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("plane %s %q must not carry a query or fragment", field, raw)
	}
	if path := strings.TrimSuffix(u.Path, "/"); path != "" {
		return "", fmt.Errorf("plane %s %q must not carry a path", field, raw)
	}
	if u.Scheme == "http" && !isLoopback(u.Hostname()) {
		return "", fmt.Errorf("plane %s %q must use https unless it targets localhost", field, raw)
	}
	return u.Scheme + "://" + u.Host, nil
}

func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// ValidateOrigin checks that raw is an acceptable public origin for this
// deployment: https (or http on loopback), with a host and no path, query,
// fragment or credentials.
func ValidateOrigin(raw string) error {
	_, err := normaliseOrigin(raw, "public base URL")
	return err
}
