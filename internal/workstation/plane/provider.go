// Package plane implements the read-only Plane REST profile. Host selection is
// explicit so a workspace request cannot redirect credentials to an arbitrary URL.
package plane

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Profile string

const (
	Cloud      Profile = "cloud"
	SelfHosted Profile = "self_hosted"
)
const cloudHost = "https://api.plane.so"
const maxBody = 16 << 20

type Provider struct {
	client  *http.Client
	base    *url.URL
	profile Profile
}

// New accepts only the approved Plane Cloud host. Self-hosted deployments must
// opt into NewSelfHosted explicitly and still use HTTPS without query/fragment.
func New(t http.RoundTripper) (*Provider, error) { return newProvider(cloudHost, Cloud, t) }
func NewSelfHosted(host string, t http.RoundTripper) (*Provider, error) {
	return newProvider(host, SelfHosted, t)
}
func newProvider(host string, profile Profile, t http.RoundTripper) (*Provider, error) {
	u, err := url.Parse(strings.TrimRight(host, "/"))
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, ws.ErrInvalid
	}
	if profile == Cloud && u.String() != cloudHost {
		return nil, ws.ErrInvalid
	}
	if profile != Cloud && profile != SelfHosted {
		return nil, ws.ErrInvalid
	}
	if t == nil {
		t = http.DefaultTransport
	}
	return &Provider{client: &http.Client{Transport: t, Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, base: u, profile: profile}, nil
}
func failure(c string) error { return &ws.Failure{Category: c} }
func (p *Provider) get(ctx context.Context, token, path string, q url.Values, out any) (http.Header, error) {
	u := *p.base
	u.Path = strings.TrimRight(u.Path, "/") + path
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return nil, failure("SCHEMA_ERROR")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	r, e := p.client.Do(req)
	if e != nil {
		return nil, failure("DEGRADED")
	}
	defer r.Body.Close()
	switch r.StatusCode {
	case 401:
		return nil, failure("AUTH_INVALID")
	case 403:
		return nil, failure("ACCESS_DENIED")
	case 404:
		return nil, failure("ACCESS_DENIED")
	case 409:
		return nil, failure("API_VERSION_UNSUPPORTED")
	case 429:
		return nil, &ws.Failure{Category: "RATE_LIMITED", RetryAfter: time.Minute}
	case 200:
	default:
		return nil, failure("DEGRADED")
	}
	b, e := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if e != nil || len(b) > maxBody || json.Unmarshal(b, out) != nil {
		return r.Header, failure("SCHEMA_ERROR")
	}
	return r.Header, nil
}

type envelope[T any] struct {
	Results []T    `json:"results"`
	Next    string `json:"next"`
}
type workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Archived   bool   `json:"archived"`
}
type issue struct {
	ID          string `json:"id"`
	Sequence    int    `json:"sequence_id"`
	Name        string `json:"name"`
	Description string `json:"description_html"`
	URL         string `json:"url"`
	State       string `json:"state"`
	Draft       bool   `json:"is_draft"`
	Archived    bool   `json:"archived_at"`
	Created     string `json:"created_at"`
	Updated     string `json:"updated_at"`
}

func (p *Provider) Probe(ctx context.Context, token string) (ws.Identity, error) {
	var x envelope[workspace]
	_, e := p.get(ctx, token, "/api/workspaces/", nil, &x)
	if e != nil {
		return ws.Identity{}, e
	}
	if len(x.Results) == 0 || x.Results[0].ID == "" {
		return ws.Identity{}, failure("SCHEMA_ERROR")
	}
	s, e := (ws.SourceIdentity{Provider: "plane", Origin: p.base.String(), TenantID: x.Results[0].ID}).Canonical()
	if e != nil {
		return ws.Identity{}, failure("SCHEMA_ERROR")
	}
	return ws.Identity{PrincipalID: x.Results[0].ID, Source: s}, nil
}
func (p *Provider) DiscoverScopes(ctx context.Context, token, page string) (ws.Page, error) {
	if page != "" && !strings.HasPrefix(page, "http") {
		return ws.Page{}, ws.ErrInvalid
	}
	q := url.Values{}
	if page != "" {
		u, e := url.Parse(page)
		if e != nil || u.Host != p.base.Host {
			return ws.Page{}, ws.ErrInvalid
		}
		q = u.Query()
	}
	var x envelope[project]
	h, e := p.get(ctx, token, "/api/workspaces/projects/", q, &x)
	_ = h
	if e != nil {
		return ws.Page{}, e
	}
	out := ws.Page{}
	for _, v := range x.Results {
		if v.ID == "" || v.Name == "" {
			return ws.Page{}, failure("SCHEMA_ERROR")
		}
		if !v.Archived {
			out.Scopes = append(out.Scopes, ws.Scope{ID: v.ID, Name: v.Name})
		}
	}
	out.Next = x.Next
	return out, nil
}
func (p *Provider) ResolveScope(ctx context.Context, token, id string) (ws.Scope, error) {
	var v project
	_, e := p.get(ctx, token, "/api/projects/"+url.PathEscape(id)+"/", nil, &v)
	if e != nil {
		return ws.Scope{}, e
	}
	if v.ID == "" || v.Name == "" {
		return ws.Scope{}, failure("SCHEMA_ERROR")
	}
	return ws.Scope{ID: v.ID, Name: v.Name, Hidden: v.Archived}, nil
}
func (p *Provider) Scan(ctx context.Context, token string, scope ws.Scope, now time.Time) (ws.Scan, error) {
	var x envelope[issue]
	q := url.Values{"limit": {"100"}}
	h, e := p.get(ctx, token, "/api/projects/"+url.PathEscape(scope.ID)+"/issues/", q, &x)
	if e != nil {
		return ws.Scan{}, e
	}
	out := ws.Scan{Coverage: "COMPLETE"}
	for _, v := range x.Results {
		created, e1 := time.Parse(time.RFC3339, v.Created)
		updated, e2 := time.Parse(time.RFC3339, v.Updated)
		if v.ID == "" || v.Name == "" || e1 != nil || e2 != nil {
			return ws.Scan{}, failure("SCHEMA_ERROR")
		}
		if v.Draft || v.Archived {
			continue
		}
		out.Items = append(out.Items, &v1.WorkStationItem{Provider: "plane", Kind: "ISSUE", RemoteId: v.ID, Identifier: scope.ID + "-" + strconv.Itoa(v.Sequence), ScopeId: scope.ID, ScopeName: scope.Name, Title: v.Name, Description: v.Description, DescriptionAvailable: v.Description != "", Url: v.URL, RawState: v.State, SourceCreatedAt: timestamppb.New(created), SourceUpdatedAt: timestamppb.New(updated), ObservedAt: timestamppb.New(now)})
	}
	if h.Get("X-Next-Page") != "" || x.Next != "" {
		out.Coverage = "PARTIAL"
		out.Unresolved = 1
	}
	return out, nil
}
