package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	googleAuthorizeURL = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL     = "https://oauth2.googleapis.com/token"
	googleUserURL      = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// GoogleConfig is the subset of OAuth provider config the Google provider
// needs. RedirectURL must exactly match a redirect URI registered in the
// Google Cloud OAuth client. Default scopes are "openid email profile".
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	DisplayName  string
	HTTPClient   *http.Client
	AuthorizeURL string
	TokenURL     string
	UserURL      string
}

// NewGoogle builds a Google provider from the given config. Returns nil when
// ClientID or ClientSecret is empty so callers can skip registering it.
func NewGoogle(cfg GoogleConfig) *GoogleProvider {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	if cfg.AuthorizeURL == "" {
		cfg.AuthorizeURL = googleAuthorizeURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = googleTokenURL
	}
	if cfg.UserURL == "" {
		cfg.UserURL = googleUserURL
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"openid", "email", "profile"}
	}
	if cfg.DisplayName == "" {
		cfg.DisplayName = "Google"
	}
	return &GoogleProvider{cfg: cfg}
}

// GoogleProvider implements Provider for Google OAuth (OIDC).
type GoogleProvider struct {
	cfg GoogleConfig
}

func (p *GoogleProvider) Name() string        { return "google" }
func (p *GoogleProvider) DisplayName() string { return p.cfg.DisplayName }

func (p *GoogleProvider) AuthorizeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", p.cfg.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(p.cfg.Scopes, " "))
	q.Set("state", state)
	q.Set("access_type", "online")
	q.Set("prompt", "select_account")
	return p.cfg.AuthorizeURL + "?" + q.Encode()
}

func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*Claims, error) {
	token, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return p.fetchUser(ctx, token)
}

func (p *GoogleProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	body := url.Values{}
	body.Set("client_id", p.cfg.ClientID)
	body.Set("client_secret", p.cfg.ClientSecret)
	body.Set("code", code)
	body.Set("grant_type", "authorization_code")
	body.Set("redirect_uri", p.cfg.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("google: token request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google: token request status %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("google: decode token response: %w", err)
	}
	if out.Error != "" {
		return "", fmt.Errorf("google: %s: %s", out.Error, out.ErrorDescription)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("google: empty access token in response")
	}
	return out.AccessToken, nil
}

func (p *GoogleProvider) fetchUser(ctx context.Context, token string) (*Claims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.UserURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := p.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google: userinfo request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google: userinfo status %d: %s", resp.StatusCode, string(raw))
	}
	// /oauth2/v3/userinfo returns OIDC standard claims.
	var u struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("google: decode userinfo response: %w", err)
	}
	if u.Sub == "" {
		return nil, fmt.Errorf("google: userinfo response missing sub")
	}
	// Use the local-part of the email as a username hint; falls back to "user"
	// in oauthUsername if empty.
	login := ""
	if u.Email != "" {
		if at := strings.IndexByte(u.Email, '@'); at > 0 {
			login = u.Email[:at]
		}
	}
	return &Claims{
		Provider:   "google",
		ExternalID: u.Sub,
		Login:      login,
		Name:       u.Name,
		Email:      u.Email,
		AvatarURL:  u.Picture,
	}, nil
}
