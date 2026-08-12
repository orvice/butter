// Package telegramapi is the provider seam between Butter and the Telegram
// Bot API (issue #264).
//
// It exists so that credential validation, Webhook reconciliation, and
// message delivery can be exercised without a live Bot, and so that the rest
// of the codebase never constructs Telegram URLs or handles Bot Tokens
// directly. Callers obtain a Client from a Factory using a decrypted token
// and discard it; there is deliberately no client cache, because a rotated
// credential must take effect on the next call.
package telegramapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the public Telegram Bot API endpoint.
const DefaultBaseURL = "https://api.telegram.org"

var (
	// ErrUnauthorized means Telegram rejected the Bot Token itself. Callers
	// map it to a credential problem rather than a transient failure.
	ErrUnauthorized = errors.New("telegram rejected the bot token")
	// ErrNotFound means the addressed resource does not exist.
	ErrNotFound = errors.New("telegram resource not found")
)

// APIError is a structured Telegram error response.
type APIError struct {
	// Code is the HTTP-equivalent error_code Telegram returned.
	Code int
	// Description is Telegram's human-readable message. It never contains
	// the Bot Token, which travels in the URL path and is redacted from any
	// error this package produces.
	Description string
	// RetryAfter is the parameters.retry_after value on a 429, in seconds.
	RetryAfter int
}

func (e *APIError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("telegram error %d: %s (retry after %ds)", e.Code, e.Description, e.RetryAfter)
	}
	return fmt.Sprintf("telegram error %d: %s", e.Code, e.Description)
}

// Is lets callers match ErrUnauthorized / ErrNotFound through errors.Is
// without unwrapping the concrete type.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrUnauthorized:
		return e.Code == http.StatusUnauthorized
	case ErrNotFound:
		return e.Code == http.StatusNotFound
	}
	return false
}

// BotIdentity is the subset of Telegram `getMe` that Butter persists. The ID
// is the immutable Bot identity a Channel is pinned to.
type BotIdentity struct {
	ID                      int64
	Username                string
	FirstName               string
	CanJoinGroups           bool
	CanReadAllGroupMessages bool
	SupportsInlineQueries   bool
}

// Client talks to one Telegram Bot.
type Client interface {
	// GetMe resolves the Bot identity behind the credential. It is the
	// single validation used before a Channel or a rotated credential is
	// committed.
	GetMe(ctx context.Context) (BotIdentity, error)

	// SendMessage delivers one text message. Forum Topic targeting travels
	// in SendMessageParams.MessageThreadID.
	SendMessage(ctx context.Context, params SendMessageParams) (Message, error)

	// EditMessageText replaces the text of a message this Bot sent.
	EditMessageText(ctx context.Context, params EditMessageParams) (Message, error)
}

// Factory builds a Client for a decrypted Bot Token. Services hold a Factory
// rather than a Client so tests can substitute a fake without a live Bot.
type Factory func(token string) Client

// HTTPClient is the live Telegram Bot API implementation.
type HTTPClient struct {
	token   string
	baseURL string
	http    *http.Client
}

var _ Client = (*HTTPClient)(nil)

// Option customizes an HTTPClient.
type Option func(*HTTPClient)

// WithBaseURL points the client at an alternate API host (used by tests and
// by self-hosted Bot API servers).
func WithBaseURL(base string) Option {
	return func(c *HTTPClient) {
		if trimmed := strings.TrimRight(strings.TrimSpace(base), "/"); trimmed != "" {
			c.baseURL = trimmed
		}
	}
}

// WithHTTPClient supplies the underlying transport.
func WithHTTPClient(h *http.Client) Option {
	return func(c *HTTPClient) {
		if h != nil {
			c.http = h
		}
	}
}

// New returns a live client for the given Bot Token.
func New(token string, opts ...Option) *HTTPClient {
	c := &HTTPClient{
		token:   strings.TrimSpace(token),
		baseURL: DefaultBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NewFactory returns a Factory producing live clients.
func NewFactory(opts ...Option) Factory {
	return func(token string) Client { return New(token, opts...) }
}

type getMeResult struct {
	ID                      int64  `json:"id"`
	Username                string `json:"username"`
	FirstName               string `json:"first_name"`
	CanJoinGroups           bool   `json:"can_join_groups"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries"`
}

func (c *HTTPClient) GetMe(ctx context.Context) (BotIdentity, error) {
	var result getMeResult
	if err := c.call(ctx, "getMe", nil, &result); err != nil {
		return BotIdentity{}, err
	}
	return BotIdentity{
		ID:                      result.ID,
		Username:                result.Username,
		FirstName:               result.FirstName,
		CanJoinGroups:           result.CanJoinGroups,
		CanReadAllGroupMessages: result.CanReadAllGroupMessages,
		SupportsInlineQueries:   result.SupportsInlineQueries,
	}, nil
}

type envelope struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
	Parameters  struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

// call performs one Bot API method. Errors never echo the request URL,
// because the URL path contains the Bot Token.
func (c *HTTPClient) call(ctx context.Context, method string, payload any, out any) error {
	if c.token == "" {
		return &APIError{Code: http.StatusUnauthorized, Description: "no bot token configured"}
	}

	var body io.Reader
	contentType := ""
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode %s request: %w", method, err)
		}
		body = bytes.NewReader(encoded)
		contentType = "application/json"
	}

	url := c.baseURL + "/bot" + c.token + "/" + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("build %s request: %w", method, err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		// http.Client wraps the URL — and therefore the token — into its
		// error, so report the method instead of wrapping.
		return fmt.Errorf("telegram %s request failed: %s", method, redactToken(err.Error(), c.token))
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("read %s response: %w", method, err)
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("telegram %s returned a non-JSON response (status %d)", method, resp.StatusCode)
	}
	if !env.OK {
		code := env.ErrorCode
		if code == 0 {
			code = resp.StatusCode
		}
		return &APIError{
			Code:        code,
			Description: redactToken(env.Description, c.token),
			RetryAfter:  env.Parameters.RetryAfter,
		}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(env.Result, out); err != nil {
		return fmt.Errorf("decode %s result: %w", method, err)
	}
	return nil
}

// redactToken removes any occurrence of the Bot Token from text that will be
// logged or returned to a caller.
func redactToken(text, token string) string {
	if token == "" {
		return text
	}
	return strings.ReplaceAll(text, token, "[REDACTED]")
}

// RetryAfter reports the Telegram-requested backoff for err, if any.
func RetryAfter(err error) (time.Duration, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
		return time.Duration(apiErr.RetryAfter) * time.Second, true
	}
	return 0, false
}

// FormatID renders a Telegram int64 identifier in the canonical decimal form
// Butter persists.
func FormatID(id int64) string { return strconv.FormatInt(id, 10) }

// ParseID parses a canonical decimal Telegram identifier.
func ParseID(s string) (int64, error) { return strconv.ParseInt(strings.TrimSpace(s), 10, 64) }
