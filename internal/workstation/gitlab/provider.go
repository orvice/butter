// Package gitlab implements the fixed GitLab.com REST read profile. It has
// no write methods; its transport is injectable for synthetic acceptance.
package gitlab

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const origin = "https://gitlab.com/api/v4"
const maxBody = 16 << 20

type Provider struct{ client *http.Client }

func New(t http.RoundTripper) *Provider {
	if t == nil {
		t = http.DefaultTransport
	}
	return &Provider{client: &http.Client{Transport: t, Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func failure(c string) error { return &ws.Failure{Category: c} }
func (p *Provider) get(ctx context.Context, token, path string, q url.Values, out any) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+path+"?"+q.Encode(), nil)
	if err != nil {
		return nil, failure("SCHEMA_ERROR")
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, failure("DEGRADED")
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 401:
		return nil, failure("AUTH_INVALID")
	case 403:
		return nil, failure("ACCESS_DENIED")
	case 429:
		return nil, &ws.Failure{Category: "RATE_LIMITED", RetryAfter: time.Minute}
	case 404:
		return nil, failure("ACCESS_DENIED")
	case 200:
	default:
		return nil, failure("DEGRADED")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil || len(data) > maxBody || json.Unmarshal(data, out) != nil {
		return resp.Header, failure("SCHEMA_ERROR")
	}
	return resp.Header, nil
}

type project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func (p *Provider) Probe(ctx context.Context, token string) (ws.Identity, error) {
	var out struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}
	_, err := p.get(ctx, token, "/user", nil, &out)
	if err != nil {
		return ws.Identity{}, err
	}
	if out.ID <= 0 {
		return ws.Identity{}, failure("SCHEMA_ERROR")
	}
	s, e := (ws.SourceIdentity{Provider: "gitlab", Origin: origin, TenantID: strconv.Itoa(out.ID)}).Canonical()
	if e != nil {
		return ws.Identity{}, failure("SCHEMA_ERROR")
	}
	return ws.Identity{PrincipalID: strconv.Itoa(out.ID), Source: s}, nil
}
func (p *Provider) DiscoverScopes(ctx context.Context, token, pageToken string) (ws.Page, error) {
	q := url.Values{"membership": {"true"}, "per_page": {"100"}, "page": {"1"}}
	if pageToken != "" {
		q.Set("page", pageToken)
	}
	var projects []project
	h, err := p.get(ctx, token, "/projects", q, &projects)
	if err != nil {
		return ws.Page{}, err
	}
	page := ws.Page{}
	for _, v := range projects {
		if v.ID <= 0 || v.PathWithNamespace == "" {
			return ws.Page{}, failure("SCHEMA_ERROR")
		}
		page.Scopes = append(page.Scopes, ws.Scope{ID: strconv.Itoa(v.ID), Name: v.PathWithNamespace})
	}
	page.Next = h.Get("X-Next-Page")
	return page, nil
}
func (p *Provider) ResolveScope(ctx context.Context, token, id string) (ws.Scope, error) {
	if id == "" {
		return ws.Scope{}, ws.ErrInvalid
	}
	var v project
	_, err := p.get(ctx, token, "/projects/"+url.PathEscape(id), nil, &v)
	if err != nil {
		return ws.Scope{}, err
	}
	if v.ID <= 0 || v.PathWithNamespace == "" {
		return ws.Scope{}, failure("SCHEMA_ERROR")
	}
	return ws.Scope{ID: strconv.Itoa(v.ID), Name: v.PathWithNamespace}, nil
}

type issue struct {
	ID          int    `json:"id"`
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	WebURL      string `json:"web_url"`
	State       string `json:"state"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (p *Provider) Scan(ctx context.Context, token string, scope ws.Scope, now time.Time) (ws.Scan, error) {
	if scope.ID == "" {
		return ws.Scan{}, ws.ErrInvalid
	}
	q := url.Values{"state": {"all"}, "per_page": {"100"}, "page": {"1"}, "order_by": {"updated_at"}, "sort": {"desc"}}
	var rows []issue
	h, err := p.get(ctx, token, "/projects/"+url.PathEscape(scope.ID)+"/issues", q, &rows)
	if err != nil {
		return ws.Scan{}, err
	}
	r := ws.Scan{Coverage: "COMPLETE"}
	for _, v := range rows {
		created, e1 := time.Parse(time.RFC3339, v.CreatedAt)
		updated, e2 := time.Parse(time.RFC3339, v.UpdatedAt)
		if v.ID <= 0 || v.Title == "" || e1 != nil || e2 != nil {
			return ws.Scan{}, failure("SCHEMA_ERROR")
		}
		r.Items = append(r.Items, &v1.WorkStationItem{Provider: "gitlab", Kind: "ISSUE", RemoteId: strconv.Itoa(v.ID), Identifier: "!" + strconv.Itoa(v.IID), ScopeId: scope.ID, ScopeName: scope.Name, Title: v.Title, Description: v.Description, DescriptionAvailable: v.Description != "", Url: v.WebURL, RawState: v.State, SourceCreatedAt: timestamppb.New(created), SourceUpdatedAt: timestamppb.New(updated), ObservedAt: timestamppb.New(now)})
	}
	if h.Get("X-Next-Page") != "" {
		r.Coverage = "PARTIAL"
		r.Unresolved = 1
	}
	return r, nil
}
