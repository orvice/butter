package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserURL      = "https://api.github.com/user"
	githubEmailsURL    = "https://api.github.com/user/emails"
)

// GitHubConfig is the subset of OAuth provider config the GitHub provider
// needs. RedirectURL must match the callback URL registered with the GitHub
// OAuth App. Scopes defaults to ["read:user", "user:email"].
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	DisplayName  string
	// HTTPClient overrides the HTTP client used to talk to GitHub. Tests
	// inject a stubbed transport; production code can leave this nil.
	HTTPClient *http.Client
	// AuthorizeURL, TokenURL, UserURL, EmailsURL allow tests to point at an
	// httptest server. Empty values fall back to the public GitHub endpoints.
	AuthorizeURL string
	TokenURL     string
	UserURL      string
	EmailsURL    string
}

// NewGitHub builds a GitHub provider from the given config. Returns nil when
// ClientID or ClientSecret is empty so callers can skip registering it.
func NewGitHub(cfg GitHubConfig) *GitHubProvider {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	if cfg.AuthorizeURL == "" {
		cfg.AuthorizeURL = githubAuthorizeURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = githubTokenURL
	}
	if cfg.UserURL == "" {
		cfg.UserURL = githubUserURL
	}
	if cfg.EmailsURL == "" {
		cfg.EmailsURL = githubEmailsURL
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"read:user", "user:email"}
	}
	if cfg.DisplayName == "" {
		cfg.DisplayName = "GitHub"
	}
	return &GitHubProvider{cfg: cfg}
}

// GitHubProvider implements Provider for GitHub OAuth.
type GitHubProvider struct {
	cfg GitHubConfig
}

func (p *GitHubProvider) Name() string        { return "github" }
func (p *GitHubProvider) DisplayName() string { return p.cfg.DisplayName }

func (p *GitHubProvider) AuthorizeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", p.cfg.RedirectURL)
	q.Set("scope", strings.Join(p.cfg.Scopes, " "))
	q.Set("state", state)
	q.Set("allow_signup", "true")
	return p.cfg.AuthorizeURL + "?" + q.Encode()
}

func (p *GitHubProvider) Exchange(ctx context.Context, code string) (*Claims, error) {
	token, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}
	user, err := p.fetchUser(ctx, token)
	if err != nil {
		return nil, err
	}
	if user.Email == "" {
		user.Email = p.fetchPrimaryEmail(ctx, token)
	}
	return user, nil
}

func (p *GitHubProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	body := url.Values{}
	body.Set("client_id", p.cfg.ClientID)
	body.Set("client_secret", p.cfg.ClientSecret)
	body.Set("code", code)
	body.Set("redirect_uri", p.cfg.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github: token request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github: token request status %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("github: decode token response: %w", err)
	}
	if out.Error != "" {
		return "", fmt.Errorf("github: %s: %s", out.Error, out.ErrorDescription)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("github: empty access token in response")
	}
	return out.AccessToken, nil
}

func (p *GitHubProvider) fetchUser(ctx context.Context, token string) (*Claims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.UserURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := p.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: user request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: user request status %d: %s", resp.StatusCode, string(raw))
	}
	var u struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("github: decode user response: %w", err)
	}
	if u.ID == 0 {
		return nil, fmt.Errorf("github: user response missing id")
	}
	return &Claims{
		Provider:   "github",
		ExternalID: strconv.FormatInt(u.ID, 10),
		Login:      u.Login,
		Name:       u.Name,
		Email:      u.Email,
		AvatarURL:  u.AvatarURL,
	}, nil
}

// fetchPrimaryEmail is best-effort: if the primary email endpoint fails or the
// user has no verified primary, we return "" and let the AuthService decide
// how to handle a missing email (we just don't store one).
func (p *GitHubProvider) fetchPrimaryEmail(ctx context.Context, token string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.EmailsURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := p.cfg.HTTPClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email
		}
	}
	return ""
}
