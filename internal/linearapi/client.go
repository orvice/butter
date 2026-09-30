// Package linearapi is Butter's client for Linear's OAuth and GraphQL APIs
// (ADR-0015). It holds no credential state: every call takes the token it
// acts with, so a rotated or refreshed token takes effect on the next call.
package linearapi

import (
	"bytes"
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

// InstallScopes are the scopes an assignable, mentionable agent app needs.
const InstallScopes = "read,write,app:assignable,app:mentionable"

// Endpoints are Linear's API URLs, overridable in tests.
type Endpoints struct {
	AuthorizeURL string
	TokenURL     string
	RevokeURL    string
	GraphQLURL   string
}

// DefaultEndpoints are Linear's production endpoints.
var DefaultEndpoints = Endpoints{
	AuthorizeURL: "https://linear.app/oauth/authorize",
	TokenURL:     "https://api.linear.app/oauth/token",
	RevokeURL:    "https://api.linear.app/oauth/revoke",
	GraphQLURL:   "https://api.linear.app/graphql",
}

var (
	// ErrUnauthorized means Linear rejected the token: it was revoked, or
	// the app was uninstalled.
	ErrUnauthorized = errors.New("linear rejected the token")
	// ErrInvalidGrant means Linear refused an authorization code or refresh
	// token; the installation must be completed again.
	ErrInvalidGrant = errors.New("linear refused the grant")
)

// RateLimitedError reports a Linear rate limit.
type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("linear rate limit, retry after %s", e.RetryAfter)
}

// TransientError wraps a failure worth retrying: a network error or a 5xx.
type TransientError struct{ Err error }

func (e *TransientError) Error() string { return e.Err.Error() }
func (e *TransientError) Unwrap() error { return e.Err }

// IsTransient reports whether retrying the call may succeed.
func IsTransient(err error) bool {
	var transient *TransientError
	var limited *RateLimitedError
	return errors.As(err, &transient) || errors.As(err, &limited)
}

// Token is an OAuth token Linear issued.
type Token struct {
	AccessToken  string
	RefreshToken string
	// ExpiresIn is zero when Linear did not say.
	ExpiresIn time.Duration
	Scopes    []string
}

// ExpiresAt is when the access token expires, or the zero time if unknown.
func (t Token) ExpiresAt(now time.Time) time.Time {
	if t.ExpiresIn <= 0 {
		return time.Time{}
	}
	return now.Add(t.ExpiresIn)
}

// Identity is who a token acts as: the App's own user in one organization.
type Identity struct {
	AppUserID        string
	OrganizationID   string
	OrganizationName string
}

// Client calls Linear.
type Client struct {
	httpClient *http.Client
	endpoints  Endpoints
}

// New builds a client. A nil httpClient uses a 30s-timeout default.
func New(httpClient *http.Client, endpoints Endpoints) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{httpClient: httpClient, endpoints: endpoints}
}

// AuthorizeURL returns Linear's consent URL for installing an app as an
// app actor.
func (c *Client) AuthorizeURL(clientID, redirectURI, state string) string {
	q := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {InstallScopes},
		"state":         {state},
		"actor":         {"app"},
	}
	sep := "?"
	if strings.Contains(c.endpoints.AuthorizeURL, "?") {
		sep = "&"
	}
	return c.endpoints.AuthorizeURL + sep + q.Encode()
}

// ExchangeCode trades an authorization code for a token.
func (c *Client) ExchangeCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (Token, error) {
	return c.tokenGrant(ctx, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	})
}

// RefreshToken trades a refresh token for a new token. Linear rotates
// refresh tokens, so the old one stops working once this succeeds.
func (c *Client) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (Token, error) {
	return c.tokenGrant(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	})
}

type tokenResponse struct {
	AccessToken      string          `json:"access_token"`
	RefreshToken     string          `json:"refresh_token"`
	ExpiresIn        int64           `json:"expires_in"`
	Scope            json.RawMessage `json:"scope"`
	Error            string          `json:"error"`
	ErrorDescription string          `json:"error_description"`
}

// scopes normalizes scope, which Linear may send as a delimited string or a
// list.
func (t tokenResponse) scopes() []string {
	var list []string
	if json.Unmarshal(t.Scope, &list) == nil {
		return list
	}
	var joined string
	if json.Unmarshal(t.Scope, &joined) == nil {
		return strings.FieldsFunc(joined, func(r rune) bool { return r == ',' || r == ' ' })
	}
	return nil
}

func (c *Client) tokenGrant(ctx context.Context, form url.Values) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoints.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, body, err := c.do(req)
	if err != nil {
		return Token{}, err
	}
	var out tokenResponse
	decodeErr := json.Unmarshal(body, &out)
	switch {
	case out.Error == "invalid_grant":
		return Token{}, fmt.Errorf("%w: %s", ErrInvalidGrant, describe(out.ErrorDescription, out.Error))
	case res.StatusCode == http.StatusUnauthorized || out.Error == "invalid_client":
		return Token{}, fmt.Errorf("%w: %s", ErrUnauthorized, describe(out.ErrorDescription, out.Error))
	case res.StatusCode != http.StatusOK:
		return Token{}, statusError("token endpoint", res, describe(out.ErrorDescription, out.Error))
	case decodeErr != nil || out.AccessToken == "":
		return Token{}, errors.New("linear token endpoint returned no access token")
	}
	return Token{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    time.Duration(out.ExpiresIn) * time.Second,
		Scopes:       out.scopes(),
	}, nil
}

// Revoke invalidates an access token at Linear.
func (c *Client) Revoke(ctx context.Context, accessToken string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoints.RevokeURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, _, err := c.do(req)
	if err != nil {
		return err
	}
	switch res.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		// Already revoked is as good as revoked.
		return nil
	default:
		return statusError("revoke", res, "")
	}
}

// Identity returns who the token acts as.
func (c *Client) Identity(ctx context.Context, accessToken string) (Identity, error) {
	var out struct {
		Viewer struct {
			ID string `json:"id"`
		} `json:"viewer"`
		Organization struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organization"`
	}
	if err := c.GraphQL(ctx, accessToken, `query ButterIdentity { viewer { id } organization { id name } }`, nil, &out); err != nil {
		return Identity{}, err
	}
	if out.Viewer.ID == "" || out.Organization.ID == "" {
		return Identity{}, errors.New("linear did not report the app user and organization")
	}
	return Identity{
		AppUserID:        out.Viewer.ID,
		OrganizationID:   out.Organization.ID,
		OrganizationName: out.Organization.Name,
	}, nil
}

type graphQLError struct {
	Message    string `json:"message"`
	Extensions struct {
		Code string `json:"code"`
	} `json:"extensions"`
}

// GraphQL runs one operation and decodes its data into out (which may be
// nil).
func (c *Client) GraphQL(ctx context.Context, accessToken, query string, variables map[string]any, out any) error {
	payload, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoints.GraphQLURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, body, err := c.do(req)
	if err != nil {
		return err
	}
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []graphQLError  `json:"errors"`
	}
	decodeErr := json.Unmarshal(body, &envelope)
	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%w: HTTP 401", ErrUnauthorized)
	}
	if len(envelope.Errors) > 0 {
		first := envelope.Errors[0]
		switch strings.ToUpper(first.Extensions.Code) {
		case "AUTHENTICATION_ERROR", "UNAUTHENTICATED", "FORBIDDEN":
			return fmt.Errorf("%w: %s", ErrUnauthorized, first.Message)
		case "RATELIMITED":
			return &RateLimitedError{RetryAfter: retryAfter(res)}
		}
		if res.StatusCode >= 500 {
			return &TransientError{Err: fmt.Errorf("linear graphql: %s", first.Message)}
		}
		return fmt.Errorf("linear graphql: %s", first.Message)
	}
	if res.StatusCode != http.StatusOK {
		return statusError("graphql", res, "")
	}
	if decodeErr != nil || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return errors.New("linear graphql: response has no data")
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return fmt.Errorf("linear graphql: decode data: %w", err)
	}
	return nil
}

// do sends req and reads a bounded body. Network failures are transient.
func (c *Client) do(req *http.Request) (*http.Response, []byte, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return nil, nil, req.Context().Err()
		}
		return nil, nil, &TransientError{Err: fmt.Errorf("linear: %w", err)}
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return nil, nil, &TransientError{Err: fmt.Errorf("linear: read response: %w", err)}
	}
	return res, body, nil
}

func statusError(what string, res *http.Response, detail string) error {
	err := fmt.Errorf("linear %s: HTTP %d", what, res.StatusCode)
	if detail != "" {
		err = fmt.Errorf("linear %s: HTTP %d: %s", what, res.StatusCode, detail)
	}
	switch {
	case res.StatusCode == http.StatusTooManyRequests:
		return &RateLimitedError{RetryAfter: retryAfter(res)}
	case res.StatusCode >= 500:
		return &TransientError{Err: err}
	default:
		return err
	}
}

func retryAfter(res *http.Response) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(res.Header.Get("Retry-After"))); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return 5 * time.Second
}

func describe(description, code string) string {
	if description != "" {
		return description
	}
	return code
}

// Activity is one agent activity posted to a Linear Agent Session.
type Activity struct {
	// Type is thought, action, elicitation, response or error.
	Type string
	// Body is the text of thought, elicitation, response and error.
	Body string
	// Action, Parameter and Result describe an action.
	Action    string
	Parameter string
	Result    string
	// Ephemeral activities are replaced by the next one. Linear accepts it
	// on thought and action only.
	Ephemeral bool
	// Signal and SignalMetadata qualify an elicitation, e.g. "select".
	Signal         string
	SignalMetadata map[string]any
}

// Activity types.
const (
	ActivityThought     = "thought"
	ActivityAction      = "action"
	ActivityElicitation = "elicitation"
	ActivityResponse    = "response"
	ActivityError       = "error"
)

func (a Activity) content() map[string]any {
	if a.Type == ActivityAction {
		content := map[string]any{"type": a.Type, "action": a.Action, "parameter": a.Parameter}
		if a.Result != "" {
			content["result"] = a.Result
		}
		return content
	}
	return map[string]any{"type": a.Type, "body": a.Body}
}

// CreateActivity posts an activity to the agent session.
func (c *Client) CreateActivity(ctx context.Context, accessToken, agentSessionID string, a Activity) error {
	input := map[string]any{"agentSessionId": agentSessionID, "content": a.content()}
	if a.Ephemeral && (a.Type == ActivityThought || a.Type == ActivityAction) {
		input["ephemeral"] = true
	}
	if a.Signal != "" {
		input["signal"] = a.Signal
	}
	if len(a.SignalMetadata) > 0 {
		input["signalMetadata"] = a.SignalMetadata
	}
	var out struct {
		AgentActivityCreate struct {
			Success bool `json:"success"`
		} `json:"agentActivityCreate"`
	}
	if err := c.GraphQL(ctx, accessToken, `mutation ButterAgentActivity($input: AgentActivityCreateInput!) {
  agentActivityCreate(input: $input) { success }
}`, map[string]any{"input": input}, &out); err != nil {
		return err
	}
	if !out.AgentActivityCreate.Success {
		return errors.New("linear agentActivityCreate returned success=false")
	}
	return nil
}

// SetSessionExternalURL links the agent session to a page outside Linear.
func (c *Client) SetSessionExternalURL(ctx context.Context, accessToken, agentSessionID, label, externalURL string) error {
	var out struct {
		AgentSessionUpdate struct {
			Success bool `json:"success"`
		} `json:"agentSessionUpdate"`
	}
	if err := c.GraphQL(ctx, accessToken, `mutation ButterAgentSession($id: String!, $input: AgentSessionUpdateInput!) {
  agentSessionUpdate(id: $id, input: $input) { success }
}`, map[string]any{
		"id":    agentSessionID,
		"input": map[string]any{"externalUrls": []map[string]any{{"label": label, "url": externalURL}}},
	}, &out); err != nil {
		return err
	}
	if !out.AgentSessionUpdate.Success {
		return errors.New("linear agentSessionUpdate returned success=false")
	}
	return nil
}
