package planeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	Scopes         []string
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
	client := *h
	// OAuth credentials and installation tokens must stay on the configured
	// origin, including when a provider answers with a redirect.
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{httpClient: &client, OAuthURL: p.OAuthURL, APIURL: p.APIBaseURL}
}
func (c *Client) AuthorizeURL(clientID, redirectURI, state string) string {
	q := url.Values{"client_id": {clientID}, "redirect_uri": {redirectURI}, "response_type": {"code"}, "state": {state}}
	return strings.TrimRight(c.OAuthURL, "/") + "/auth/o/authorize-app/?" + q.Encode()
}

// BotToken uses the installation-specific client credentials grant for both
// the initial install and renewal. The callback's authorization code is not
// exchanged for a user token.
func (c *Client) BotToken(ctx context.Context, clientID, secret, installationID string) (Token, error) {
	if strings.TrimSpace(installationID) == "" {
		return Token{}, errors.New("plane installation ID is required")
	}
	form := url.Values{"grant_type": {"client_credentials"}, "app_installation_id": {installationID}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.OAuthURL, "/")+"/auth/o/token/", strings.NewReader(form.Encode()))
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
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if res.StatusCode/100 != 2 {
		return Token{}, fmt.Errorf("plane bot token request failed: HTTP %d", res.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); err != nil {
		return Token{}, errors.New("invalid plane bot token response")
	}
	if body.AccessToken == "" || body.ExpiresIn <= 0 || body.ExpiresIn > int64((1<<63-1)/time.Second) {
		return Token{}, errors.New("incomplete plane bot token response")
	}
	return Token{AccessToken: body.AccessToken, ExpiresIn: time.Duration(body.ExpiresIn) * time.Second, InstallationID: installationID, Scopes: strings.Fields(body.Scope)}, nil
}
func (c *Client) GetInstallation(ctx context.Context, token, id string) (Installation, error) {
	query := url.Values{"id": {id}}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.OAuthURL, "/")+"/auth/o/app-installation/?"+query.Encode(), nil)
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
	var rows []struct {
		ID              string `json:"id"`
		Workspace       string `json:"workspace"`
		WorkspaceDetail struct {
			Slug string `json:"slug"`
		} `json:"workspace_detail"`
		AppBot string `json:"app_bot"`
		Status string `json:"status"`
	}
	if e = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&rows); e != nil {
		return Installation{}, errors.New("invalid plane installation response")
	}
	// Never take the first installation from an unfiltered or ambiguous list.
	// The installation-specific token must prove exactly the requested identity.
	if len(rows) != 1 || rows[0].ID != id || rows[0].Status != "installed" ||
		rows[0].Workspace == "" || rows[0].WorkspaceDetail.Slug == "" || rows[0].AppBot == "" {
		return Installation{}, errors.New("plane installation identity is incomplete")
	}
	row := rows[0]
	return Installation{ID: row.ID, Installed: true, WorkspaceID: row.Workspace, WorkspaceSlug: row.WorkspaceDetail.Slug, BotUserID: row.AppBot}, nil
}
