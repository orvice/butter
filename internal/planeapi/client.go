// Package planeapi contains the small protocol seam used by Plane App
// installation. It deliberately accepts credentials per call and never
// stores or logs them.
package planeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrUnauthorized = errors.New("plane unauthorized")
	ErrInvalidGrant = errors.New("plane invalid grant")
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Now        func() time.Time
}

type Token struct {
	AccessToken string
	TokenType   string
	Scope       string
	ExpiresAt   time.Time
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresIn   int64  `json:"expires_in"`
	Error       string `json:"error"`
}

// MintBotToken exchanges the installation id for a short-lived bot token.
// Plane's bot flow has no refresh token; callers mint again near expiry.
func (c *Client) MintBotToken(ctx context.Context, clientID, clientSecret, installationID, scope string) (Token, error) {
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" || strings.TrimSpace(installationID) == "" {
		return Token{}, fmt.Errorf("plane token request requires client id, secret and installation id")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		return Token{}, fmt.Errorf("plane token endpoint is not configured")
	}
	form := url.Values{"grant_type": {"client_credentials"}, "app_installation_id": {installationID}}
	if scope != "" {
		form.Set("scope", scope)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/auth/o/token/", strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()
	var body tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Token{}, fmt.Errorf("decode plane token response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		if body.Error == "invalid_grant" {
			return Token{}, ErrInvalidGrant
		}
		return Token{}, ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Token{}, fmt.Errorf("plane token endpoint returned %s", resp.Status)
	}
	if body.AccessToken == "" {
		return Token{}, fmt.Errorf("plane token response omitted access_token")
	}
	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	return Token{AccessToken: body.AccessToken, TokenType: body.TokenType, Scope: body.Scope, ExpiresAt: now.Add(time.Duration(body.ExpiresIn) * time.Second)}, nil
}
