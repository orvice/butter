package planeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Token struct {
	AccessToken    string
	ExpiresIn      time.Duration
	InstallationID string
}
type Installation struct {
	ID            string   `json:"id"`
	Installed     bool     `json:"installed"`
	WorkspaceID   string   `json:"workspace_id"`
	WorkspaceSlug string   `json:"workspace_slug"`
	BotUserID     string   `json:"bot_user_id"`
	BotName       string   `json:"bot_name"`
	Scopes        []string `json:"scopes"`
}
type Client struct {
	httpClient *http.Client
	OAuthURL   string
	APIURL     string
}

func NewClient(h *http.Client, p Profile) *Client {
	if h == nil {
		h = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{httpClient: h, OAuthURL: p.OAuthURL, APIURL: p.APIBaseURL}
}
func (c *Client) AuthorizeURL(clientID, redirectURI, state string) string {
	q := url.Values{"client_id": {clientID}, "redirect_uri": {redirectURI}, "response_type": {"code"}, "state": {state}}
	return strings.TrimRight(c.OAuthURL, "/") + "/oauth/authorize?" + q.Encode()
}
func (c *Client) ExchangeCode(ctx context.Context, clientID, secret, code string) (Token, error) {
	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.OAuthURL, "/")+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.SetBasicAuth(clientID, secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer res.Body.Close()
	var body struct {
		AccessToken    string `json:"access_token"`
		ExpiresIn      int64  `json:"expires_in"`
		InstallationID string `json:"app_installation_id"`
		Error          string `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return Token{}, err
	}
	if res.StatusCode/100 != 2 || body.AccessToken == "" || body.InstallationID == "" {
		return Token{}, fmt.Errorf("plane authorization exchange failed: %s", body.Error)
	}
	return Token{AccessToken: body.AccessToken, ExpiresIn: time.Duration(body.ExpiresIn) * time.Second, InstallationID: body.InstallationID}, nil
}
func (c *Client) ClientCredentials(ctx context.Context, clientID, secret string) (Token, error) {
	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {clientID}, "client_secret": {secret}}
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.OAuthURL, "/")+"/oauth/token", strings.NewReader(form.Encode()))
	if e != nil {
		return Token{}, e
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, e := c.httpClient.Do(req)
	if e != nil {
		return Token{}, e
	}
	defer res.Body.Close()
	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Error       string `json:"error"`
	}
	if e = json.NewDecoder(res.Body).Decode(&body); e != nil {
		return Token{}, e
	}
	if res.StatusCode/100 != 2 || body.AccessToken == "" {
		return Token{}, fmt.Errorf("plane token exchange failed: %s", body.Error)
	}
	return Token{AccessToken: body.AccessToken, ExpiresIn: time.Duration(body.ExpiresIn) * time.Second}, nil
}
func (c *Client) GetInstallation(ctx context.Context, token, id string) (Installation, error) {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.APIURL, "/")+"/api/instances/"+url.PathEscape(id)+"/app-installation", nil)
	if e != nil {
		return Installation{}, e
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, e := c.httpClient.Do(req)
	if e != nil {
		return Installation{}, e
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return Installation{}, fmt.Errorf("plane installation lookup failed: %s", strconv.Itoa(res.StatusCode))
	}
	var out Installation
	if e = json.NewDecoder(res.Body).Decode(&out); e != nil {
		return Installation{}, e
	}
	if !out.Installed || out.ID == "" || out.WorkspaceID == "" || out.BotUserID == "" {
		return Installation{}, errors.New("plane installation identity is incomplete")
	}
	return out, nil
}
