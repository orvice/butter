package mcpoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const defaultDiscoveryTimeout = 5 * time.Second

type OAuthEndpoints struct {
	AuthorizationURL string
	TokenURL         string
	RegistrationURL  string
	Resource         string
	Scopes           []string
}

type protectedResourceMetadata struct {
	Resource             string   `json:"resource"`
	AuthorizationServers []string `json:"authorization_servers"`
	ScopesSupported      []string `json:"scopes_supported"`
}

type authorizationServerMetadata struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	RegistrationEndpoint  string   `json:"registration_endpoint"`
	ScopesSupported       []string `json:"scopes_supported"`
}

func Discover(ctx context.Context, client *http.Client, srv *agentsv1.MCPServer, allowInsecureHTTP bool) (*OAuthEndpoints, error) {
	if client == nil {
		client = &http.Client{Timeout: defaultDiscoveryTimeout}
	}
	oauth := srv.GetAuth().GetOauth2()
	if oauth == nil {
		return nil, fmt.Errorf("oauth2 config is required")
	}
	resource := strings.TrimSpace(oauth.GetResource())
	if resource == "" {
		resource = strings.TrimSpace(srv.GetUrl())
	}
	if err := validateEndpointURL(resource, allowInsecureHTTP); err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	out := &OAuthEndpoints{
		AuthorizationURL: strings.TrimSpace(oauth.GetAuthorizationUrl()),
		TokenURL:         strings.TrimSpace(oauth.GetTokenUrl()),
		Resource:         resource,
		Scopes:           uniqueStrings(oauth.GetScopes()),
	}

	authServerURL := strings.TrimSpace(oauth.GetAuthorizationServerUrl())
	if needsMetadata(out) {
		resourceMetadataURL := strings.TrimSpace(oauth.GetResourceMetadataUrl())
		if resourceMetadataURL == "" {
			resourceMetadataURL = protectedResourceMetadataURL(srv.GetUrl())
		}
		if resourceMetadataURL != "" {
			if prm, err := fetchJSON[protectedResourceMetadata](ctx, client, resourceMetadataURL, allowInsecureHTTP); err == nil {
				if prm.Resource != "" && out.Resource == "" {
					out.Resource = prm.Resource
				}
				if authServerURL == "" && len(prm.AuthorizationServers) > 0 {
					authServerURL = prm.AuthorizationServers[0]
				}
				out.Scopes = mergeScopes(out.Scopes, prm.ScopesSupported)
			}
		}
	}

	if needsMetadata(out) && authServerURL != "" {
		as, err := fetchAuthorizationServerMetadata(ctx, client, authServerURL, allowInsecureHTTP)
		if err != nil {
			return nil, err
		}
		if out.AuthorizationURL == "" {
			out.AuthorizationURL = as.AuthorizationEndpoint
		}
		if out.TokenURL == "" {
			out.TokenURL = as.TokenEndpoint
		}
		out.RegistrationURL = as.RegistrationEndpoint
		out.Scopes = mergeScopes(out.Scopes, as.ScopesSupported)
	}

	if out.AuthorizationURL == "" {
		return nil, fmt.Errorf("authorization_url or authorization_server_url discovery is required")
	}
	if out.TokenURL == "" {
		return nil, fmt.Errorf("token_url or authorization_server_url discovery is required")
	}
	if err := validateEndpointURL(out.AuthorizationURL, allowInsecureHTTP); err != nil {
		return nil, fmt.Errorf("authorization_url: %w", err)
	}
	if err := validateEndpointURL(out.TokenURL, allowInsecureHTTP); err != nil {
		return nil, fmt.Errorf("token_url: %w", err)
	}
	if out.RegistrationURL != "" {
		if err := validateEndpointURL(out.RegistrationURL, allowInsecureHTTP); err != nil {
			return nil, fmt.Errorf("registration_endpoint: %w", err)
		}
	}
	return out, nil
}

func needsMetadata(out *OAuthEndpoints) bool {
	return out.AuthorizationURL == "" || out.TokenURL == ""
}

func fetchAuthorizationServerMetadata(ctx context.Context, client *http.Client, issuer string, allowInsecureHTTP bool) (*authorizationServerMetadata, error) {
	if strings.Contains(issuer, ".well-known/") {
		return fetchJSON[authorizationServerMetadata](ctx, client, issuer, allowInsecureHTTP)
	}
	u, err := url.Parse(strings.TrimSpace(issuer))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("authorization_server_url must be an absolute URL")
	}
	path := strings.TrimRight(u.Path, "/")
	candidates := []string{
		u.Scheme + "://" + u.Host + "/.well-known/oauth-authorization-server" + path,
		u.Scheme + "://" + u.Host + "/.well-known/openid-configuration" + path,
	}
	var lastErr error
	for _, candidate := range candidates {
		as, err := fetchJSON[authorizationServerMetadata](ctx, client, candidate, allowInsecureHTTP)
		if err == nil {
			return as, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("discover authorization server metadata: %w", lastErr)
}

func fetchJSON[T any](ctx context.Context, client *http.Client, raw string, allowInsecureHTTP bool) (*T, error) {
	if err := validateEndpointURL(raw, allowInsecureHTTP); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s returned HTTP %d", raw, resp.StatusCode)
	}
	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode %s: %w", raw, err)
	}
	return &out, nil
}

func protectedResourceMetadataURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host + "/.well-known/oauth-protected-resource"
}

func validateEndpointURL(raw string, allowInsecureHTTP bool) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("must be an absolute URL")
	}
	if !allowInsecureHTTP && isPrivateLiteralIP(u.Hostname()) && !isLocalhost(u.Hostname()) {
		return fmt.Errorf("must not target private IP literals")
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if allowInsecureHTTP || isLocalhost(u.Hostname()) {
			return nil
		}
		return fmt.Errorf("must use https outside localhost development")
	default:
		return fmt.Errorf("must use http or https")
	}
}

func isLocalhost(host string) bool {
	h := strings.ToLower(strings.Trim(host, "[]"))
	if h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return true
	}
	ip, err := netip.ParseAddr(h)
	if err != nil {
		return false
	}
	return ip.IsLoopback()
}

func isPrivateLiteralIP(host string) bool {
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return false
	}
	return ip.IsPrivate()
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func mergeScopes(primary, discovered []string) []string {
	return uniqueStrings(primary)
}
