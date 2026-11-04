// Package plane implements one explicitly selected Plane API v2 read profile.
// Instances with another schema fail closed; it never probes v1 as a fallback.
package plane

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

const maxBody = 16 << 20
const maxPages = 1000

var pathAtom = regexp.MustCompile(`^[A-Za-z0-9_-]{1,200}$`)

type Provider struct {
	client              *http.Client
	api, web, workspace string
}

// New only accepts bases already approved by the global Host registry. The
// shared client independently enforces outbound address and redirect policy.
func New(apiBase, webBase, workspaceSlug string, transport http.RoundTripper) (*Provider, error) {
	if !pathAtom.MatchString(workspaceSlug) {
		return nil, ws.ErrInvalid
	}
	source, err := (ws.SourceIdentity{Provider: "plane", Origin: apiBase}).Canonical()
	if err != nil {
		return nil, ws.ErrInvalid
	}
	web, err := (ws.SourceIdentity{Provider: "plane", Origin: webBase}).Canonical()
	if err != nil || !strings.HasPrefix(web.Origin, "https://") {
		return nil, ws.ErrInvalid
	}
	client, err := outbound.NewClient(source.Origin, transport)
	if err != nil {
		return nil, err
	}
	return &Provider{client: client, api: source.Origin, web: web.Origin, workspace: workspaceSlug}, nil
}

func failure(category string) error { return &ws.Failure{Category: category} }
func (p *Provider) base() string    { return "/api/v2/workspaces/" + p.workspace }
func (p *Provider) get(ctx context.Context, token, path string, query url.Values, out any) error {
	u := p.api + path
	if len(query) != 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return failure("CONFIG_ERROR")
	}
	req.Header.Set("X-Api-Key", token)
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return ws.RequestFailure(err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return failure("AUTH_INVALID")
	case http.StatusForbidden, http.StatusNotFound:
		return failure("ACCESS_DENIED")
	case http.StatusPaymentRequired, http.StatusGone:
		return failure("FEATURE_UNAVAILABLE")
	case http.StatusBadRequest, http.StatusNotAcceptable:
		return failure("UNSUPPORTED")
	case http.StatusTooManyRequests:
		delay := time.Minute
		if n, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 32); err == nil && n > 0 {
			delay = time.Duration(n) * time.Second
		} else if until, err := http.ParseTime(resp.Header.Get("Retry-After")); err == nil {
			delay = max(delay, time.Until(until))
		}
		return &ws.Failure{Category: "RATE_LIMITED", RetryAfter: delay}
	default:
		return failure("DEGRADED")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return failure("DEGRADED")
	}
	if len(body) > maxBody || json.Unmarshal(body, out) != nil {
		return failure("SCHEMA_ERROR")
	}
	return nil
}

// opaqueID accepts exact string or integer spellings without float conversion.
type opaqueID string

func (id *opaqueID) UnmarshalJSON(raw []byte) error {
	var s string
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &s); err != nil {
			return err
		}
	} else {
		s = string(raw)
		if s == "" {
			return ws.ErrInvalid
		}
		for _, c := range s {
			if c < '0' || c > '9' {
				return ws.ErrInvalid
			}
		}
	}
	if !pathAtom.MatchString(s) {
		return ws.ErrInvalid
	}
	*id = opaqueID(s)
	return nil
}

func (p *Provider) Probe(ctx context.Context, token string) (ws.Identity, error) {
	var user struct {
		ID   opaqueID `json:"id"`
		Kind string   `json:"principal_kind"`
	}
	if err := p.get(ctx, token, "/api/v2/users/me/", nil, &user); err != nil {
		return ws.Identity{}, err
	}
	if user.ID == "" || user.Kind != "api_key" {
		return ws.Identity{}, failure("UNSUPPORTED")
	}
	// v2 does not expose an immutable workspace ID on its public project/user
	// resources. Keep this explicit rather than inventing an identity endpoint.
	return ws.Identity{PrincipalID: string(user.ID), Source: ws.SourceIdentity{Provider: "plane", Origin: p.api, TenantID: "workspace-slug:" + p.workspace}}, nil
}

type envelope[T any] struct {
	Data       []T     `json:"data"`
	Next       *string `json:"next_cursor"`
	More       *bool   `json:"has_more"`
	Pagination struct {
		Style string `json:"style"`
	} `json:"pagination"`
}

func next[T any](page envelope[T], cursor string) (string, error) {
	if page.Data == nil {
		return "", failure("SCHEMA_ERROR")
	}
	if page.Pagination.Style != "cursor" {
		return "", failure("UNSUPPORTED")
	}
	if page.More == nil {
		return "", failure("SCHEMA_ERROR")
	}
	if !*page.More {
		if page.Next != nil && *page.Next != "" {
			return "", failure("SCHEMA_ERROR")
		}
		return "", nil
	}
	if page.Next == nil || *page.Next == "" || *page.Next == cursor || len(*page.Next) > 4096 || len(page.Data) == 0 {
		return "", failure("SCHEMA_ERROR")
	}
	return *page.Next, nil
}
func paging(cursor string) (url.Values, error) {
	if len(cursor) > 4096 {
		return nil, ws.ErrInvalid
	}
	q := url.Values{"paginate": {"cursor"}, "per_page": {"100"}, "order_by": {"id"}}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	return q, nil
}

type project struct {
	ID   opaqueID `json:"id"`
	Name string   `json:"name"`
}

func projectScope(p project) (ws.Scope, error) {
	if p.ID == "" || p.Name == "" {
		return ws.Scope{}, failure("SCHEMA_ERROR")
	}
	return ws.Scope{ID: string(p.ID), Name: p.Name}, nil
}
func (p *Provider) DiscoverScopes(ctx context.Context, token, cursor string) (ws.Page, error) {
	q, err := paging(cursor)
	if err != nil {
		return ws.Page{}, err
	}
	q.Set("include_archived", "true")
	q.Set("fields", "id,name")
	var page envelope[project]
	if err := p.get(ctx, token, p.base()+"/projects/", q, &page); err != nil {
		return ws.Page{}, err
	}
	out := ws.Page{}
	out.Next, err = next(page, cursor)
	if err != nil {
		return ws.Page{}, err
	}
	seen := map[string]bool{}
	for _, project := range page.Data {
		scope, err := projectScope(project)
		if err != nil {
			return ws.Page{}, err
		}
		if seen[scope.ID] {
			return ws.Page{}, failure("SCHEMA_ERROR")
		}
		seen[scope.ID] = true
		out.Scopes = append(out.Scopes, scope)
	}
	return out, nil
}
func (p *Provider) ResolveScope(ctx context.Context, token, id string) (ws.Scope, error) {
	if !pathAtom.MatchString(id) {
		return ws.Scope{}, ws.ErrInvalid
	}
	var project project
	if err := p.get(ctx, token, p.base()+"/projects/"+id+"/", url.Values{"fields": {"id,name"}}, &project); err != nil {
		return ws.Scope{}, err
	}
	if string(project.ID) != id {
		return ws.Scope{}, failure("SCHEMA_ERROR")
	}
	return projectScope(project)
}

type relation struct {
	ID          opaqueID `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
}
type remoteItem struct {
	ID          opaqueID        `json:"id"`
	ProjectID   opaqueID        `json:"project_id"`
	Identifier  string          `json:"identifier"`
	Name        string          `json:"name"`
	Description json.RawMessage `json:"description_html"`
	Draft       *bool           `json:"is_draft"`
	Created     time.Time       `json:"created_at"`
	Updated     time.Time       `json:"updated_at"`
	Completed   *time.Time      `json:"completed_at"`
	Author      opaqueID        `json:"created_by_id"`
	StateID     opaqueID        `json:"state_id"`
	State       struct {
		ID    opaqueID `json:"id"`
		Name  string   `json:"name"`
		Group string   `json:"group"`
	} `json:"state"`
	AssigneeIDs []opaqueID `json:"assignee_ids"`
	LabelIDs    []opaqueID `json:"label_ids"`
	Assignees   []relation `json:"assignees"`
	Labels      []relation `json:"labels"`
}

func names(ids []opaqueID, expanded []relation, user bool) ([]string, error) {
	if ids == nil || expanded == nil || len(ids) != len(expanded) {
		return nil, failure("SCHEMA_ERROR")
	}
	byID := map[opaqueID]string{}
	for _, r := range expanded {
		name := r.Name
		if user {
			name = r.DisplayName
		}
		if r.ID == "" || name == "" || byID[r.ID] != "" {
			return nil, failure("SCHEMA_ERROR")
		}
		byID[r.ID] = name
	}
	var result []string
	seen := map[opaqueID]bool{}
	for _, id := range ids {
		if byID[id] == "" || seen[id] {
			return nil, failure("SCHEMA_ERROR")
		}
		seen[id] = true
		name := byID[id]
		if user {
			name = string(id) + ":" + name
		}
		result = append(result, name)
	}
	return result, nil
}
func (p *Provider) normalize(r remoteItem, scope ws.Scope) (*v1.WorkStationItem, error) {
	if r.ID == "" || string(r.ProjectID) != scope.ID || r.Name == "" || r.Identifier == "" || r.Draft == nil || r.Created.IsZero() || r.Updated.IsZero() || r.StateID == "" || r.State.ID != r.StateID || r.State.Name == "" || r.State.Group == "" {
		return nil, failure("SCHEMA_ERROR")
	}
	item := &v1.WorkStationItem{Provider: "plane", Kind: "ISSUE", RemoteId: string(r.ID), Identifier: r.Identifier, ScopeId: scope.ID, ScopeName: scope.Name, Title: r.Name, RawState: r.State.Name, Status: "OTHER", Draft: *r.Draft, Author: string(r.Author), SourceCreatedAt: timestamppb.New(r.Created), SourceUpdatedAt: timestamppb.New(r.Updated)}
	item.ProjectId, item.ProjectName = scope.ID, scope.Name
	item.Url = p.web + "/" + p.workspace + "/projects/" + scope.ID + "/issues/" + string(r.ID)
	if len(r.Description) > 0 && string(r.Description) != "null" {
		if json.Unmarshal(r.Description, &item.Description) != nil {
			return nil, failure("SCHEMA_ERROR")
		}
		item.DescriptionAvailable = true
	}
	var err error
	if item.Assignees, err = names(r.AssigneeIDs, r.Assignees, true); err != nil {
		return nil, err
	}
	if item.Labels, err = names(r.LabelIDs, r.Labels, false); err != nil {
		return nil, err
	}
	switch r.State.Group {
	case "backlog", "unstarted", "triage":
		item.Status = "NOT_STARTED"
	case "started":
		item.Status = "IN_PROGRESS"
	case "completed":
		item.Status, item.Terminal = "DONE", true
		if r.Completed != nil && !r.Completed.IsZero() {
			item.EndedAt = timestamppb.New(*r.Completed)
		}
	case "cancelled":
		item.Status, item.Terminal = "CANCELED", true
	}
	return item, nil
}

func (p *Provider) Scan(ctx context.Context, token string, selected ws.Scope, now time.Time) (ws.Scan, error) {
	scope, err := p.ResolveScope(ctx, token, selected.ID)
	if err != nil {
		return ws.Scan{}, err
	}
	result := ws.Scan{Coverage: "COMPLETE"}
	items := map[string]*v1.WorkStationItem{}
	seenCursors := map[string]bool{"": true}
	cursor := ""
	for pageNumber := 1; ; pageNumber++ {
		q, _ := paging(cursor)
		q.Set("archived", "true")
		q.Set("is_draft", "false")
		q.Set("fields", "id,project_id,identifier,name,description_html,is_draft,created_at,updated_at,completed_at,created_by_id,state_id,assignee_ids,label_ids")
		q.Set("expand", "state,assignees,labels")
		var page envelope[remoteItem]
		if err := p.get(ctx, token, p.base()+"/projects/"+scope.ID+"/work-items/", q, &page); err != nil {
			return ws.Scan{}, err
		}
		nextCursor, err := next(page, cursor)
		if err != nil {
			return ws.Scan{}, err
		}
		progress := 0
		for _, remote := range page.Data {
			// The server filter is repeated locally so a buggy endpoint cannot
			// accidentally turn drafts into formal work items.
			if remote.Draft != nil && *remote.Draft {
				continue
			}
			item, err := p.normalize(remote, scope)
			if err != nil {
				return ws.Scan{}, err
			}
			if old, ok := items[item.RemoteId]; !ok {
				items[item.RemoteId] = item
				progress++
			} else if item.SourceUpdatedAt.AsTime().After(old.SourceUpdatedAt.AsTime()) {
				items[item.RemoteId] = item
			}
		}
		if nextCursor == "" {
			break
		}
		if seenCursors[nextCursor] || (progress == 0 && len(page.Data) > 0) {
			return ws.Scan{}, failure("SCHEMA_ERROR")
		}
		if pageNumber >= maxPages {
			return ws.Scan{}, failure("UNSUPPORTED")
		}
		seenCursors[nextCursor] = true
		cursor = nextCursor
	}
	current, err := p.ResolveScope(ctx, token, scope.ID)
	if err != nil {
		return ws.Scan{}, err
	}
	if current.Name != scope.Name {
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
	sort.Slice(result.Items, func(i, j int) bool { return result.Items[i].RemoteId < result.Items[j].RemoteId })
	return result, nil
}

var _ ws.Provider = (*Provider)(nil)
