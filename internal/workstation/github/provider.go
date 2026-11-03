// Package github implements the fixed GitHub.com REST read profile. It exposes
// no write methods, follows no redirects and never includes upstream bodies in
// errors. Fixtures exercise it through an injected transport, not another URL.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const origin = "https://api.github.com"
const maxPages = 1000
const maxBody = 16 << 20

type Provider struct{ client *http.Client }

func New(transport http.RoundTripper) *Provider {
	if transport == nil {
		t := http.DefaultTransport.(*http.Transport).Clone()
		// Fixed origins still need DNS checks at connection time. Disable proxy
		// routing so it cannot bypass this check; never disable TLS verification.
		t.Proxy = nil
		t.DialContext = publicDial
		transport = t
	}
	return &Provider{client: &http.Client{Transport: transport, Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func publicDial(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || host != "api.github.com" || port != "443" {
		return nil, fmt.Errorf("unapproved destination")
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("DNS unavailable")
	}
	d := net.Dialer{Timeout: 15 * time.Second}
	for _, ip := range ips {
		if !ip.IP.IsGlobalUnicast() || ip.IP.IsPrivate() || ip.IP.IsLoopback() || ip.IP.IsLinkLocalUnicast() {
			continue
		}
		conn, err := d.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if err == nil {
			return conn, nil
		}
	}
	return nil, fmt.Errorf("no approved destination available")
}
func failure(category string) error { return &ws.Failure{Category: category} }
func (p *Provider) get(ctx context.Context, token, path string, q url.Values, out any) (http.Header, error) {
	u := origin + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, failure("SCHEMA_ERROR")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, failure("DEGRADED")
	}
	defer resp.Body.Close()
	h := resp.Header
	if resp.StatusCode == 429 || (resp.StatusCode == 403 && (h.Get("X-RateLimit-Remaining") == "0" || h.Get("Retry-After") != "")) {
		delay := time.Minute
		if sec, err := strconv.Atoi(h.Get("Retry-After")); err == nil && sec > 0 {
			delay = time.Duration(sec) * time.Second
		} else if until, err := http.ParseTime(h.Get("Retry-After")); err == nil {
			delay = max(delay, time.Until(until))
		}
		if reset, err := strconv.ParseInt(h.Get("X-RateLimit-Reset"), 10, 64); err == nil {
			delay = max(delay, time.Until(time.Unix(reset, 0)))
		}
		return nil, &ws.Failure{Category: "RATE_LIMITED", RetryAfter: delay}
	}
	switch resp.StatusCode {
	case 200:
	case 401:
		return nil, failure("AUTH_INVALID")
	case 403:
		// Secondary rate limits can lack Retry-After. Inspect only a bounded
		// body for classification; never retain it or expose it in an error.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 16384))
		message := strings.ToLower(string(body))
		if strings.Contains(message, "rate limit") || strings.Contains(message, "abuse detection") {
			return nil, &ws.Failure{Category: "RATE_LIMITED", RetryAfter: time.Minute}
		}
		return nil, failure("ACCESS_DENIED")
	case 404:
		return nil, failure("ACCESS_DENIED")
	case 410:
		return nil, failure("FEATURE_UNAVAILABLE")
	default:
		return nil, failure("DEGRADED")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, failure("DEGRADED")
	}
	if len(data) > maxBody || json.Unmarshal(data, out) != nil {
		return nil, failure("SCHEMA_ERROR")
	}
	return h, nil
}
func (p *Provider) Probe(ctx context.Context, token string) (ws.Identity, error) {
	var me struct {
		ID json.Number `json:"id"`
	}
	_, err := p.get(ctx, token, "/user", nil, &me)
	if err != nil {
		return ws.Identity{}, err
	}
	if !decimal(me.ID.String()) {
		return ws.Identity{}, failure("SCHEMA_ERROR")
	}
	return ws.Identity{PrincipalID: me.ID.String()}, nil
}

type repository struct {
	ID        json.Number `json:"id"`
	FullName  string      `json:"full_name"`
	HasIssues *bool       `json:"has_issues"`
}

func decimal(s string) bool {
	if s == "" || s == "0" || len(s) > 30 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return s[0] != '0'
}
func repoScope(r repository) (ws.Scope, error) {
	if !decimal(r.ID.String()) || r.FullName == "" {
		return ws.Scope{}, failure("SCHEMA_ERROR")
	}
	return ws.Scope{ID: r.ID.String(), Name: r.FullName}, nil
}
func pageNumber(token string) (int, error) {
	if token == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(token)
	if err != nil || n < 1 || n > maxPages {
		return 0, ws.ErrInvalid
	}
	return n, nil
}

var nextLink = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func nextPage(h http.Header, path string, q url.Values, page int) (string, error) {
	match := nextLink.FindStringSubmatch(h.Get("Link"))
	if match == nil {
		if strings.Contains(strings.ToLower(h.Get("Link")), "next") {
			return "", failure("SCHEMA_ERROR")
		}
		return "", nil
	}
	if page >= maxPages {
		return "", failure("UNSUPPORTED")
	}
	u, err := url.Parse(match[1])
	if err != nil || u.Scheme != "https" || u.Host != "api.github.com" || u.User != nil || u.Path != path || u.Fragment != "" {
		return "", failure("SCHEMA_ERROR")
	}
	expected := q.Clone()
	expected.Set("page", strconv.Itoa(page+1))
	if u.Query().Encode() != expected.Encode() {
		return "", failure("SCHEMA_ERROR")
	}
	return strconv.Itoa(page + 1), nil
}
func (p *Provider) DiscoverScopes(ctx context.Context, token, pageToken string) (ws.Page, error) {
	n, err := pageNumber(pageToken)
	if err != nil {
		return ws.Page{}, err
	}
	path := "/user/repos"
	q := url.Values{"per_page": {"100"}, "page": {strconv.Itoa(n)}, "sort": {"full_name"}, "direction": {"asc"}}
	var repos []repository
	h, err := p.get(ctx, token, path, q, &repos)
	if err != nil {
		return ws.Page{}, err
	}
	if repos == nil {
		return ws.Page{}, failure("SCHEMA_ERROR")
	}
	out := ws.Page{}
	for _, r := range repos {
		sc, err := repoScope(r)
		if err != nil {
			return ws.Page{}, err
		}
		out.Scopes = append(out.Scopes, sc)
	}
	out.Next, err = nextPage(h, path, q, n)
	return out, err
}
func (p *Provider) repository(ctx context.Context, token, id string) (repository, error) {
	if !decimal(id) {
		return repository{}, ws.ErrInvalid
	}
	var repo repository
	_, err := p.get(ctx, token, "/repositories/"+id, nil, &repo)
	if err != nil {
		return repo, err
	}
	if repo.ID.String() != id || repo.FullName == "" {
		return repo, failure("SCHEMA_ERROR")
	}
	return repo, nil
}
func (p *Provider) ResolveScope(ctx context.Context, token, id string) (ws.Scope, error) {
	r, err := p.repository(ctx, token, id)
	if err != nil {
		return ws.Scope{}, err
	}
	return repoScope(r)
}

type user struct {
	ID    json.Number `json:"id"`
	Login string      `json:"login"`
}
type remoteItem struct {
	Number      json.Number     `json:"number"`
	Title       string          `json:"title"`
	Body        json.RawMessage `json:"body"`
	State       string          `json:"state"`
	StateReason string          `json:"state_reason"`
	Created     time.Time       `json:"created_at"`
	Updated     time.Time       `json:"updated_at"`
	Closed      *time.Time      `json:"closed_at"`
	Merged      *time.Time      `json:"merged_at"`
	Draft       bool            `json:"draft"`
	PullRequest json.RawMessage `json:"pull_request"`
	User        user            `json:"user"`
	Assignees   []user          `json:"assignees"`
	Labels      []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func normalize(r remoteItem, scope ws.Scope, kind string) (*v1.WorkStationItem, error) {
	if !decimal(r.Number.String()) || r.Title == "" || r.Updated.IsZero() || r.Created.IsZero() || r.State == "" {
		return nil, failure("SCHEMA_ERROR")
	}
	item := &v1.WorkStationItem{Provider: "github", Kind: kind, RemoteId: r.Number.String(), Identifier: "#" + r.Number.String(), ScopeId: scope.ID, ScopeName: scope.Name, Title: r.Title, DescriptionAvailable: len(r.Body) > 0, RawState: r.State, Terminal: r.State == "closed", Author: r.User.ID.String(), SourceCreatedAt: timestamppb.New(r.Created), SourceUpdatedAt: timestamppb.New(r.Updated)}
	if len(r.Body) > 0 && string(r.Body) != "null" {
		if json.Unmarshal(r.Body, &item.Description) != nil {
			return nil, failure("SCHEMA_ERROR")
		}
	}
	for _, u := range r.Assignees {
		if !decimal(u.ID.String()) {
			return nil, failure("SCHEMA_ERROR")
		}
		item.Assignees = append(item.Assignees, u.ID.String()+":"+u.Login)
	}
	for _, l := range r.Labels {
		item.Labels = append(item.Labels, l.Name)
	}
	// Build links from approved origin + stable verified repo identity, never
	// trust upstream html_url or description links for the source action.
	item.Url = "https://github.com/" + scope.Name + "/issues/" + r.Number.String()
	item.Status = "NOT_STARTED"
	ended := r.Closed
	if kind == "PULL_REQUEST" {
		item.Url = "https://github.com/" + scope.Name + "/pull/" + r.Number.String()
		item.Status = "OPEN"
		item.Draft = r.Draft
		if item.Terminal {
			item.Status = "CLOSED"
		}
		if r.Merged != nil {
			item.Terminal = true
			item.Status = "MERGED"
			ended = r.Merged
		}
	} else if item.Terminal {
		item.Status = "DONE"
		if r.StateReason == "not_planned" || r.StateReason == "duplicate" {
			item.Status = "CANCELED"
		}
	}
	if r.State != "open" && r.State != "closed" && r.Merged == nil {
		item.Status = "OTHER"
	}
	if item.Terminal && ended != nil && !ended.IsZero() {
		item.EndedAt = timestamppb.New(*ended)
	}
	return item, nil
}

// Scan currently performs a bounded full traversal each cycle. It deliberately
// has no ETag/incremental cache: a 304 or a failed page can never publish an
// empty/partial traversal. Incremental reconciliation is a documented follow-up.
func (p *Provider) Scan(ctx context.Context, token string, scope ws.Scope, now time.Time) (ws.Scan, error) {
	repo, err := p.repository(ctx, token, scope.ID)
	if err != nil {
		return ws.Scan{}, err
	}
	scope.Name = repo.FullName
	if repo.HasIssues == nil {
		return ws.Scan{}, failure("SCHEMA_ERROR")
	}
	if !*repo.HasIssues {
		return ws.Scan{}, failure("FEATURE_UNAVAILABLE")
	}
	// A name is used only after resolving the stable ID. Segments are escaped
	// and forbidden separators are rejected before constructing request paths.
	parts := strings.Split(scope.Name, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ContainsAny(scope.Name, "?#%\\") || parts[0] == ".." || parts[1] == ".." || parts[0] == "." || parts[1] == "." {
		return ws.Scan{}, failure("SCHEMA_ERROR")
	}
	base := "/repos/" + url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1])
	result := ws.Scan{Coverage: "COMPLETE"}
	items := map[string]*v1.WorkStationItem{}
	for _, endpoint := range []struct{ Path, Kind string }{{"/issues", "ISSUE"}, {"/pulls", "PULL_REQUEST"}} {
		path := base + endpoint.Path
		seenPages := map[string]bool{}
		for page := 1; ; page++ {
			q := url.Values{"state": {"all"}, "sort": {"updated"}, "direction": {"desc"}, "per_page": {"100"}, "page": {strconv.Itoa(page)}}
			var remote []remoteItem
			h, err := p.get(ctx, token, path, q, &remote)
			if err != nil {
				return ws.Scan{}, err
			}
			if remote == nil {
				return ws.Scan{}, failure("SCHEMA_ERROR")
			}
			signature := ""
			for _, r := range remote {
				signature += r.Number.String() + ","
			}
			if len(remote) > 0 && seenPages[signature] {
				return ws.Scan{}, failure("SCHEMA_ERROR")
			}
			seenPages[signature] = true
			progress := 0
			for _, r := range remote {
				if endpoint.Kind == "ISSUE" && len(r.PullRequest) > 0 && string(r.PullRequest) != "null" {
					continue
				}
				item, err := normalize(r, scope, endpoint.Kind)
				if err != nil {
					return ws.Scan{}, err
				}
				key := endpoint.Kind + ":" + item.RemoteId
				if old, ok := items[key]; !ok {
					progress++
					items[key] = item
				} else if item.SourceUpdatedAt.AsTime().After(old.SourceUpdatedAt.AsTime()) {
					items[key] = item
				}
			}
			next, err := nextPage(h, path, q, page)
			if err != nil {
				return ws.Scan{}, err
			}
			if next == "" {
				break
			}
			if len(remote) == 0 || (progress == 0 && endpoint.Kind == "PULL_REQUEST") {
				return ws.Scan{}, failure("SCHEMA_ERROR")
			}
		}
	}
	// Recheck access and stable ownership before publishing the traversal.
	current, err := p.repository(ctx, token, scope.ID)
	if err != nil {
		return ws.Scan{}, err
	}
	if current.FullName != scope.Name {
		return ws.Scan{}, failure("SCHEMA_ERROR")
	}
	for _, item := range items {
		if item.Terminal && item.EndedAt == nil {
			result.Coverage = "PARTIAL"
			result.Unresolved++
			continue
		}
		if ws.Retained(item, now) {
			result.Items = append(result.Items, item)
		}
	}
	return result, nil
}
