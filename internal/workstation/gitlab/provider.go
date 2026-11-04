// Package gitlab implements a bounded, read-only GitLab REST v4 profile.
// Endpoint origins come from the administrator-managed Git Host registry.
package gitlab

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/outbound"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	pageSize = 100
	maxPages = 500 // Offset profile deliberately stops at 50,000 records.
	maxBody  = 16 << 20
)

type Provider struct {
	client *http.Client
	api    *url.URL
	web    *url.URL
}

// New accepts only the fixed v4 endpoint below the approved web root, including
// installations hosted under a URL prefix. It never guesses another API root.
func New(apiBase, webBase string, transport http.RoundTripper) (*Provider, error) {
	apiIdentity, err := (ws.SourceIdentity{Provider: "gitlab", Origin: apiBase}).Canonical()
	if err != nil {
		return nil, ws.ErrInvalid
	}
	webIdentity, err := (ws.SourceIdentity{Provider: "gitlab", Origin: webBase}).Canonical()
	if err != nil {
		return nil, ws.ErrInvalid
	}
	api, _ := url.Parse(apiIdentity.Origin)
	web, _ := url.Parse(webIdentity.Origin)
	if api.Scheme != "https" || web.Scheme != "https" || api.Host != web.Host || api.Path != web.Path+"/api/v4" || api.RawPath != "" || web.RawPath != "" || !safePath(web.Path, true) {
		return nil, ws.ErrInvalid
	}
	client, err := outbound.NewClient(apiIdentity.Origin, transport)
	if err != nil {
		return nil, err
	}
	return &Provider{client: client, api: api, web: web}, nil
}

func failure(category string) error { return &ws.Failure{Category: category} }

func (p *Provider) get(ctx context.Context, token, path string, q url.Values, out any) (http.Header, error) {
	u := *p.api
	u.Path += path
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
	if resp.StatusCode == http.StatusTooManyRequests {
		delay := time.Minute
		if sec, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 32); err == nil && sec > 0 {
			delay = time.Duration(sec) * time.Second
		} else if until, err := http.ParseTime(resp.Header.Get("Retry-After")); err == nil {
			delay = max(time.Second, time.Until(until))
		}
		return nil, &ws.Failure{Category: "RATE_LIMITED", RetryAfter: delay}
	}
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, failure("AUTH_INVALID")
	case http.StatusForbidden, http.StatusNotFound:
		return nil, failure("ACCESS_DENIED")
	case http.StatusBadRequest, http.StatusNotAcceptable, http.StatusUnprocessableEntity:
		// An unavailable filter/pagination profile cannot become an empty scan.
		return nil, failure("UNSUPPORTED")
	case http.StatusGone:
		return nil, failure("FEATURE_UNAVAILABLE")
	default:
		return nil, failure("DEGRADED")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, failure("DEGRADED")
	}
	if len(body) > maxBody || json.Unmarshal(body, out) != nil {
		return nil, failure("SCHEMA_ERROR")
	}
	return resp.Header, nil
}

// ProfileEvidence is an instance observation, not a claim of qualification.
// Deployment acceptance records its version/revision and the fixture matrix.
type ProfileEvidence struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
}

var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$`)

func (p *Provider) InspectProfile(ctx context.Context, token string) (ProfileEvidence, error) {
	var profile ProfileEvidence
	if _, err := p.get(ctx, token, "/version", nil, &profile); err != nil {
		return ProfileEvidence{}, err
	}
	if !versionPattern.MatchString(profile.Version) || profile.Revision == "" || len(profile.Revision) > 128 {
		return ProfileEvidence{}, failure("SCHEMA_ERROR")
	}
	return profile, nil
}

func (p *Provider) Probe(ctx context.Context, token string) (ws.Identity, error) {
	var me struct {
		ID json.Number `json:"id"`
	}
	if _, err := p.get(ctx, token, "/user", nil, &me); err != nil {
		return ws.Identity{}, err
	}
	if !decimal(me.ID.String()) {
		return ws.Identity{}, failure("SCHEMA_ERROR")
	}
	if _, err := p.InspectProfile(ctx, token); err != nil {
		return ws.Identity{}, err
	}
	return ws.Identity{PrincipalID: me.ID.String(), Source: ws.SourceIdentity{Provider: "gitlab", Origin: p.api.String()}}, nil
}

func decimal(s string) bool {
	if s == "" || s[0] == '0' || len(s) > 30 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func safePath(s string, leadingSlash bool) bool {
	if strings.ContainsAny(s, "?#%\\\x00") || strings.Contains(s, "//") {
		return false
	}
	if leadingSlash {
		s = strings.TrimPrefix(s, "/")
		if s == "" {
			return true
		}
	}
	for _, part := range strings.Split(s, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

type project struct {
	ID                   json.Number `json:"id"`
	Path                 string      `json:"path_with_namespace"`
	IssuesEnabled        *bool       `json:"issues_enabled"`
	MergeRequestsEnabled *bool       `json:"merge_requests_enabled"`
}

func projectScope(r project) (ws.Scope, error) {
	if !decimal(r.ID.String()) || !safePath(r.Path, false) {
		return ws.Scope{}, failure("SCHEMA_ERROR")
	}
	return ws.Scope{ID: r.ID.String(), Name: r.Path}, nil
}

func pageNumber(token string) (int, error) {
	if token == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(token)
	if err != nil || n < 1 || n > maxPages || strconv.Itoa(n) != token {
		return 0, ws.ErrInvalid
	}
	return n, nil
}

var nextLink = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

// nextPage uses GitLab's next-page header or next link, checking agreement when
// both are supplied. An unannotated full page must be followed, never truncated.
func (p *Provider) nextPage(h http.Header, path string, q url.Values, page, count int) (string, error) {
	if count > pageSize {
		return "", failure("SCHEMA_ERROR")
	}
	if current := h.Get("X-Page"); current != "" && current != strconv.Itoa(page) {
		return "", failure("SCHEMA_ERROR")
	}
	total := 0
	if value := h.Get("X-Total-Pages"); value != "" {
		var err error
		total, err = strconv.Atoi(value)
		if err != nil || total < 0 || (total != 0 && total < page) {
			return "", failure("SCHEMA_ERROR")
		}
		if total > maxPages {
			return "", failure("UNSUPPORTED")
		}
	}
	_, headerPresent := h["X-Next-Page"]
	next := h.Get("X-Next-Page")
	if next != "" && next != strconv.Itoa(page+1) {
		return "", failure("SCHEMA_ERROR")
	}
	links := nextLink.FindAllStringSubmatch(h.Get("Link"), -1)
	if len(links) > 1 || (len(links) == 0 && strings.Contains(strings.ToLower(h.Get("Link")), "next")) {
		return "", failure("SCHEMA_ERROR")
	}
	if len(links) == 1 {
		u, err := url.Parse(links[0][1])
		expected := q.Clone()
		expected.Set("page", strconv.Itoa(page+1))
		if err != nil || u.Scheme != p.api.Scheme || u.Host != p.api.Host || u.Path != p.api.Path+path || u.User != nil || u.Fragment != "" || u.Query().Encode() != expected.Encode() || (headerPresent && next == "") {
			return "", failure("SCHEMA_ERROR")
		}
		next = strconv.Itoa(page + 1)
	}
	if total > page {
		if headerPresent && next == "" {
			return "", failure("SCHEMA_ERROR")
		}
		next = strconv.Itoa(page + 1)
	}
	if total > 0 && total == page && next != "" {
		return "", failure("SCHEMA_ERROR")
	}
	if next == "" && !headerPresent && len(links) == 0 && total == 0 && count == pageSize {
		next = strconv.Itoa(page + 1)
	}
	if next != "" && count == 0 {
		return "", failure("SCHEMA_ERROR")
	}
	if next != "" && page >= maxPages {
		return "", failure("UNSUPPORTED")
	}
	return next, nil
}

func (p *Provider) DiscoverScopes(ctx context.Context, token, pageToken string) (ws.Page, error) {
	page, err := pageNumber(pageToken)
	if err != nil {
		return ws.Page{}, err
	}
	q := url.Values{"simple": {"true"}, "order_by": {"id"}, "sort": {"asc"}, "per_page": {strconv.Itoa(pageSize)}, "page": {strconv.Itoa(page)}}
	var remote []project
	h, err := p.get(ctx, token, "/projects", q, &remote)
	if err != nil {
		return ws.Page{}, err
	}
	if remote == nil {
		return ws.Page{}, failure("SCHEMA_ERROR")
	}
	result := ws.Page{}
	seen := map[string]bool{}
	for _, r := range remote {
		scope, err := projectScope(r)
		if err != nil || seen[scope.ID] {
			return ws.Page{}, failure("SCHEMA_ERROR")
		}
		seen[scope.ID] = true
		result.Scopes = append(result.Scopes, scope)
	}
	result.Next, err = p.nextPage(h, "/projects", q, page, len(remote))
	return result, err
}

func (p *Provider) project(ctx context.Context, token, id string) (project, error) {
	if !decimal(id) {
		return project{}, ws.ErrInvalid
	}
	var r project
	if _, err := p.get(ctx, token, "/projects/"+id, nil, &r); err != nil {
		return project{}, err
	}
	if _, err := projectScope(r); err != nil || r.ID.String() != id {
		return project{}, failure("SCHEMA_ERROR")
	}
	return r, nil
}

func (p *Provider) ResolveScope(ctx context.Context, token, id string) (ws.Scope, error) {
	r, err := p.project(ctx, token, id)
	if err != nil {
		return ws.Scope{}, err
	}
	return projectScope(r)
}

type user struct {
	ID       json.Number `json:"id"`
	Username string      `json:"username"`
}

type remoteItem struct {
	IID         json.Number     `json:"iid"`
	ProjectID   json.Number     `json:"project_id"`
	Title       string          `json:"title"`
	Description json.RawMessage `json:"description"`
	State       string          `json:"state"`
	Created     time.Time       `json:"created_at"`
	Updated     time.Time       `json:"updated_at"`
	Closed      *time.Time      `json:"closed_at"`
	Merged      *time.Time      `json:"merged_at"`
	Draft       *bool           `json:"draft"`
	Author      user            `json:"author"`
	Assignees   []user          `json:"assignees"`
	Labels      []string        `json:"labels"`
}

func (p *Provider) normalize(r remoteItem, scope ws.Scope, kind string) (*v1.WorkStationItem, error) {
	if !decimal(r.IID.String()) || r.ProjectID.String() != scope.ID || r.Title == "" || r.State == "" || r.Created.IsZero() || r.Updated.IsZero() || !decimal(r.Author.ID.String()) {
		return nil, failure("SCHEMA_ERROR")
	}
	item := &v1.WorkStationItem{Provider: "gitlab", Kind: kind, RemoteId: r.IID.String(), Identifier: "#" + r.IID.String(), ScopeId: scope.ID, ScopeName: scope.Name, ProjectId: scope.ID, ProjectName: scope.Name, Title: r.Title, DescriptionAvailable: len(r.Description) > 0, RawState: r.State, Author: r.Author.ID.String(), Labels: r.Labels, SourceCreatedAt: timestamppb.New(r.Created), SourceUpdatedAt: timestamppb.New(r.Updated)}
	if len(r.Description) > 0 && string(r.Description) != "null" && json.Unmarshal(r.Description, &item.Description) != nil {
		return nil, failure("SCHEMA_ERROR")
	}
	for _, assignee := range r.Assignees {
		if !decimal(assignee.ID.String()) {
			return nil, failure("SCHEMA_ERROR")
		}
		item.Assignees = append(item.Assignees, assignee.ID.String()+":"+assignee.Username)
	}
	u := *p.web
	u.Path += "/" + scope.Name + "/-/issues/" + r.IID.String()
	item.Status = "OTHER"
	var ended *time.Time
	if kind == "MERGE_REQUEST" {
		u.Path = p.web.Path + "/" + scope.Name + "/-/merge_requests/" + r.IID.String()
		item.Identifier = "!" + r.IID.String()
		if r.Draft == nil {
			return nil, failure("SCHEMA_ERROR")
		}
		item.Draft = *r.Draft
		switch r.State {
		case "opened", "locked":
			item.Status = "OPEN"
		case "merged":
			item.Status, item.Terminal, ended = "MERGED", true, r.Merged
		case "closed":
			item.Status, item.Terminal, ended = "CLOSED", true, r.Closed
		}
	} else {
		switch r.State {
		case "opened":
			item.Status = "NOT_STARTED"
		case "closed":
			item.Status, item.Terminal, ended = "DONE", true, r.Closed
		}
	}
	if ended != nil && !ended.IsZero() && item.Terminal {
		item.EndedAt = timestamppb.New(*ended)
	}
	item.Url = u.String()
	return item, nil
}

// Scan traverses complete project lists, publishing nothing on any failed page.
// Filtering the 90-day terminal window happens after traversal and deduplication;
// stale updated_at cannot cause open issues or reopened work to disappear.
func (p *Provider) Scan(ctx context.Context, token string, scope ws.Scope, now time.Time) (ws.Scan, error) {
	r, err := p.project(ctx, token, scope.ID)
	if err != nil {
		return ws.Scan{}, err
	}
	if r.IssuesEnabled == nil || r.MergeRequestsEnabled == nil {
		return ws.Scan{}, failure("SCHEMA_ERROR")
	}
	if !*r.IssuesEnabled || !*r.MergeRequestsEnabled {
		return ws.Scan{}, failure("FEATURE_UNAVAILABLE")
	}
	scope.Name = r.Path
	items := map[string]*v1.WorkStationItem{}
	for _, endpoint := range []struct{ Path, Kind string }{{"issues", "ISSUE"}, {"merge_requests", "MERGE_REQUEST"}} {
		path := "/projects/" + scope.ID + "/" + endpoint.Path
		seenPages := map[string]bool{}
		for page := 1; ; page++ {
			q := url.Values{"state": {"all"}, "scope": {"all"}, "order_by": {"updated_at"}, "sort": {"desc"}, "per_page": {strconv.Itoa(pageSize)}, "page": {strconv.Itoa(page)}}
			var remote []remoteItem
			h, err := p.get(ctx, token, path, q, &remote)
			if err != nil {
				return ws.Scan{}, err
			}
			if remote == nil {
				return ws.Scan{}, failure("SCHEMA_ERROR")
			}
			var signature strings.Builder
			newItems := 0
			for _, r := range remote {
				item, err := p.normalize(r, scope, endpoint.Kind)
				if err != nil {
					return ws.Scan{}, err
				}
				signature.WriteString(item.RemoteId + ",")
				key := item.Kind + ":" + item.RemoteId
				old, exists := items[key]
				if !exists {
					newItems++
				}
				if !exists || item.SourceUpdatedAt.AsTime().After(old.SourceUpdatedAt.AsTime()) {
					items[key] = item
				}
			}
			if len(remote) > 0 && seenPages[signature.String()] {
				return ws.Scan{}, failure("SCHEMA_ERROR")
			}
			seenPages[signature.String()] = true
			next, err := p.nextPage(h, path, q, page, len(remote))
			if err != nil {
				return ws.Scan{}, err
			}
			if next == "" {
				break
			}
			if newItems == 0 {
				return ws.Scan{}, failure("SCHEMA_ERROR")
			}
		}
	}
	current, err := p.project(ctx, token, scope.ID)
	if err != nil {
		return ws.Scan{}, err
	}
	if current.Path != scope.Name || current.IssuesEnabled == nil || current.MergeRequestsEnabled == nil || !*current.IssuesEnabled || !*current.MergeRequestsEnabled {
		return ws.Scan{}, failure("ACCESS_DENIED")
	}
	result := ws.Scan{Coverage: "COMPLETE"}
	for _, item := range items {
		if item.Terminal && item.EndedAt == nil {
			result.Coverage, result.Unresolved = "PARTIAL", result.Unresolved+1
			continue
		}
		if ws.Retained(item, now) {
			result.Items = append(result.Items, item)
		}
	}
	sort.Slice(result.Items, func(i, j int) bool {
		if result.Items[i].Kind != result.Items[j].Kind {
			return result.Items[i].Kind < result.Items[j].Kind
		}
		return result.Items[i].RemoteId < result.Items[j].RemoteId
	})
	return result, nil
}
