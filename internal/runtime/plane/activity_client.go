package plane

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ActivityType string

const (
	ActivityThought     ActivityType = "thought"
	ActivityAction      ActivityType = "action"
	ActivityResponse    ActivityType = "response"
	ActivityElicitation ActivityType = "elicitation"
	ActivityError       ActivityType = "error"
)

type ActivityRequest struct {
	WorkspaceID, AppID, WorkspaceSlug, RunID, InstallationID string
	Type                                                     ActivityType
	Body                                                     string
	Signal                                                   string
	Metadata                                                 map[string]any
}
type ActivityResult struct {
	ID     string
	Status int
}
type Activity struct {
	ID   string       `json:"id"`
	Type ActivityType `json:"type"`
	Body string       `json:"body"`
}
type ActivityQuery struct {
	WorkspaceID, AppID, WorkspaceSlug, RunID, InstallationID string
}
type ActivityClient interface {
	CreateActivity(context.Context, ActivityRequest) (ActivityResult, error)
	ListActivities(context.Context, ActivityQuery) ([]Activity, error)
}
type ActivityTokenSource func(context.Context, string, string, string) (string, error)
type ForceRefreshTokenSource func(context.Context, string, string, string) (string, error)
type ActivityHTTPClient struct {
	BaseURL      string
	HTTPClient   *http.Client
	Token        ActivityTokenSource
	ForceRefresh ForceRefreshTokenSource
}

var ErrActivityPermanent = errors.New("plane activity request is permanent")

type activityError struct {
	kind   DeliveryStatus
	status int
	err    error
}

func (e *activityError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return "plane activity request failed"
}
func (e *activityError) Unwrap() error                  { return e.err }
func (e *activityError) Retryable() bool                { return e.kind == DeliveryRetryable }
func (e *activityError) DeliveryStatus() DeliveryStatus { return e.kind }
func NewActivityHTTPClient(base string, h *http.Client, t ActivityTokenSource) *ActivityHTTPClient {
	if h == nil {
		h = http.DefaultClient
	}
	return &ActivityHTTPClient{BaseURL: strings.TrimRight(base, "/"), HTTPClient: h, Token: t}
}
func (c *ActivityHTTPClient) CreateActivity(ctx context.Context, a ActivityRequest) (ActivityResult, error) {
	result, err := c.createActivity(ctx, a, false)
	var ae *activityError
	if err != nil && c.ForceRefresh != nil && errors.As(err, &ae) && ae.status == http.StatusUnauthorized {
		if _, refreshErr := c.ForceRefresh(ctx, a.WorkspaceID, a.AppID, a.InstallationID); refreshErr != nil {
			return ActivityResult{}, refreshErr
		}
		return c.createActivity(ctx, a, true)
	}
	return result, err
}
func (c *ActivityHTTPClient) createActivity(ctx context.Context, a ActivityRequest, _ bool) (ActivityResult, error) {
	if c == nil || c.Token == nil {
		return ActivityResult{}, &activityError{kind: DeliveryPermanentFailure, err: ErrReplyUnavailable}
	}
	tok, err := c.Token(ctx, a.WorkspaceID, a.AppID, a.InstallationID)
	if err != nil {
		return ActivityResult{}, &activityError{kind: DeliveryUnknown, err: err}
	}
	content := map[string]any{"type": string(a.Type), "body": a.Body}
	payload := map[string]any{"content": content, "type": string(a.Type)}
	if a.Signal != "" {
		payload["signal"] = a.Signal
	}
	if a.Metadata != nil {
		payload["content_metadata"] = a.Metadata
	}
	b, _ := json.Marshal(payload)
	u := c.BaseURL + "/api/v1/workspaces/" + url.PathEscape(a.WorkspaceSlug) + "/runs/" + url.PathEscape(a.RunID) + "/activities/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return ActivityResult{}, &activityError{kind: DeliveryPermanentFailure, err: err}
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return ActivityResult{}, &activityError{kind: DeliveryUnknown, err: err}
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ActivityResult{}, classifyActivityStatus(res.StatusCode, raw)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return ActivityResult{Status: res.StatusCode}, &activityError{kind: DeliveryUnknown, status: res.StatusCode, err: err}
	}
	return ActivityResult{ID: out.ID, Status: res.StatusCode}, nil
}
func classifyActivityStatus(status int, _ []byte) error {
	kind := DeliveryPermanentFailure
	if status == 408 || status == 409 || status == 429 || (status >= 500 && status <= 599) {
		kind = DeliveryRetryable
	}
	return &activityError{kind: kind, status: status, err: fmt.Errorf("plane activity http status %d", status)}
}
func (c *ActivityHTTPClient) ListActivities(ctx context.Context, q ActivityQuery) ([]Activity, error) {
	if c == nil || c.Token == nil {
		return nil, ErrReplyUnavailable
	}
	tok, err := c.Token(ctx, q.WorkspaceID, q.AppID, q.InstallationID)
	if err != nil {
		return nil, err
	}
	u := c.BaseURL + "/api/v1/workspaces/" + url.PathEscape(q.WorkspaceSlug) + "/runs/" + url.PathEscape(q.RunID) + "/activities/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, classifyActivityStatus(res.StatusCode, raw)
	}
	var out struct {
		Results []Activity `json:"results"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

type ActivityReplyPublisher struct{ Client ActivityClient }

func (p ActivityReplyPublisher) PublishReply(ctx context.Context, e *Event, reply string) error {
	if p.Client == nil {
		return ErrReplyUnavailable
	}
	_, err := p.Client.CreateActivity(ctx, ActivityRequest{WorkspaceID: e.WorkspaceID, AppID: e.AppID, WorkspaceSlug: e.PlaneWorkspaceID, RunID: e.AgentRunID, InstallationID: e.InstallationID, Type: ActivityResponse, Body: reply})
	return err
}
