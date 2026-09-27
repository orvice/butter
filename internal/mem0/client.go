// Package mem0 is a minimal client for the mem0 OSS REST server (ADR-0013).
// It speaks only the self-hosted API: unversioned paths and `X-API-Key`
// authentication. The hosted mem0 platform (`/v3/...`, `Authorization:
// Token`) is deliberately not supported.
package mem0

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
)

// maxErrorBody bounds how much of a failed response body is kept for the
// error message.
const maxErrorBody = 4 << 10

// maxResponseBody bounds a successful response body.
const maxResponseBody = 8 << 20

// Client calls one mem0 OSS server. Timeouts come from the caller's context
// (or the supplied http.Client); the client adds none of its own.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// New returns a client for the server at baseURL. An empty apiKey sends no
// `X-API-Key` header, which only a server with AUTH_DISABLED accepts. A nil
// httpClient uses http.DefaultClient.
func New(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    httpClient,
	}
}

// APIError is a non-2xx answer from the server.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("mem0: HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("mem0: HTTP %d: %s", e.StatusCode, e.Body)
}

// IsAuthError reports whether err is the server refusing the credentials
// (401 or 403).
func IsAuthError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden
}

// SearchRequest is the body of `POST /search`. Filters must carry at least
// one scalar `user_id`, `agent_id`, or `run_id`.
type SearchRequest struct {
	Query     string         `json:"query"`
	Filters   map[string]any `json:"filters,omitempty"`
	TopK      int            `json:"top_k,omitempty"`
	Threshold *float64       `json:"threshold,omitempty"`
}

// Memory is one stored memory as returned by search.
type Memory struct {
	ID        string         `json:"id"`
	Memory    string         `json:"memory"`
	Score     float64        `json:"score,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	UpdatedAt string         `json:"updated_at,omitempty"`
	UserID    string         `json:"user_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
}

type resultsEnvelope struct {
	Results []Memory `json:"results"`
}

// Message is one conversation message submitted for extraction.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AddRequest is the body of `POST /memories`. At least one of UserID or
// AgentID must be set. `run_id` is deliberately not exposed: butter never
// sets it, because it would narrow mem0's add-time dedup (ADR-0013 §3).
type AddRequest struct {
	Messages []Message      `json:"messages"`
	UserID   string         `json:"user_id,omitempty"`
	AgentID  string         `json:"agent_id,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
	// Infer asks the server to extract facts with its LLM (the server
	// default); false stores the messages verbatim.
	Infer *bool `json:"infer,omitempty"`
}

// AddResult is one memory the server stored for an add.
type AddResult struct {
	ID     string `json:"id"`
	Memory string `json:"memory"`
	Event  string `json:"event"`
}

// Add runs `POST /memories`. With inference on, the OSS server answers only
// after its extraction LLM call, so callers should run it off the request
// path with a generous timeout.
func (c *Client) Add(ctx context.Context, req AddRequest) ([]AddResult, error) {
	var out struct {
		Results []AddResult `json:"results"`
	}
	if err := c.post(ctx, "/memories", req, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// Search runs `POST /search` and returns the matching memories.
func (c *Client) Search(ctx context.Context, req SearchRequest) ([]Memory, error) {
	var out resultsEnvelope
	if err := c.post(ctx, "/search", req, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// ListRequest scopes `GET /memories` to exactly one identity. An
// identity-less listing is admin-only on the server and never used.
type ListRequest struct {
	UserID  string
	AgentID string
	// TopK caps the result; the server defaults to 20 and allows 1000. The
	// endpoint has no pagination.
	TopK int
}

// List runs `GET /memories` for one identity.
func (c *Client) List(ctx context.Context, req ListRequest) ([]Memory, error) {
	q := url.Values{}
	switch {
	case req.UserID != "":
		q.Set("user_id", req.UserID)
	case req.AgentID != "":
		q.Set("agent_id", req.AgentID)
	default:
		return nil, errors.New("mem0: list needs a user_id or agent_id")
	}
	if req.TopK > 0 {
		q.Set("top_k", strconv.Itoa(req.TopK))
	}
	var out resultsEnvelope
	if err := c.do(ctx, http.MethodGet, "/memories?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// Get runs `GET /memories/{id}`. It returns nil without error when the
// memory does not exist (the server answers `null`).
func (c *Client) Get(ctx context.Context, id string) (*Memory, error) {
	var out *Memory
	if err := c.do(ctx, http.MethodGet, "/memories/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete runs `DELETE /memories/{id}`. The server does not check who owns
// the memory: callers must verify ownership first.
func (c *Client) Delete(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/memories/"+url.PathEscape(id), nil, nil)
}

func (c *Client) post(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("mem0: encode %s: %w", path, err)
		}
		reader = bytes.NewReader(payload)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("mem0: build %s: %w", path, err)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("mem0: %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(raw))}
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBody)).Decode(out); err != nil {
		return fmt.Errorf("mem0: decode %s: %w", path, err)
	}
	return nil
}
