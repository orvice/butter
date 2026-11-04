// Package linear implements a fixed, read-only Linear GraphQL profile. All
// requests are named queries; API keys never leave the official API origin.
package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/outbound"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const origin = "https://api.linear.app"
const maxPages = 1000
const maxBody = 16 << 20

const probeQuery = `query WorkStationIdentity { viewer { id } organization { id } }`
const teamsQuery = `query WorkStationTeams($after: String) {
  teams(first: 100, after: $after, includeArchived: true) {
    nodes { id name key } pageInfo { hasNextPage endCursor }
  }
}`
const teamQuery = `query WorkStationTeam($id: String!) { team(id: $id) { id name key } }`
const issuesQuery = `query WorkStationIssues($id: String!, $after: String, $filter: IssueFilter!) {
  team(id: $id) {
    id name key
    issues(first: 50, after: $after, includeArchived: true, includeSubTeams: false, orderBy: updatedAt, filter: $filter) {
      nodes {
        id identifier title description url createdAt updatedAt completedAt canceledAt archivedAt
        team { id } state { id name type } creator { id name } assignee { id name }
        project { id name }
        labels(first: 20, includeArchived: true) { nodes { id name } pageInfo { hasNextPage endCursor } }
      }
      pageInfo { hasNextPage endCursor }
    }
  }
}`
const labelsQuery = `query WorkStationIssueLabels($id: String!, $after: String) {
  issue(id: $id) {
    id team { id } updatedAt
    labels(first: 100, after: $after, includeArchived: true) { nodes { id name } pageInfo { hasNextPage endCursor } }
  }
}`

type Provider struct{ client *http.Client }

func New(transport http.RoundTripper) *Provider {
	client, err := outbound.NewClient(origin, transport)
	if err != nil { // The origin above is a compile-time, approved HTTPS origin.
		panic("invalid fixed Linear origin")
	}
	return &Provider{client: client}
}

func failure(category string) error { return &ws.Failure{Category: category} }

func retryAfter(h http.Header) time.Duration {
	delay := time.Minute
	if n, err := strconv.ParseInt(h.Get("Retry-After"), 10, 64); err == nil && n > 0 && n <= 86400*30 {
		delay = time.Duration(n) * time.Second
	} else if until, err := http.ParseTime(h.Get("Retry-After")); err == nil {
		delay = max(delay, time.Until(until))
	}
	for _, key := range []string{"X-RateLimit-Requests-Reset", "X-RateLimit-Complexity-Reset", "X-RateLimit-Endpoint-Requests-Reset"} {
		if n, err := strconv.ParseInt(h.Get(key), 10, 64); err == nil && n > 0 {
			delay = max(delay, time.Until(time.UnixMilli(n)))
		}
	}
	return delay
}

// query refuses GraphQL partial success before decoding any data. Errors contain
// only fixed categories, never the upstream message, request, or API key.
func (p *Provider) query(ctx context.Context, token, query string, variables map[string]any, out any) error {
	if !strings.HasPrefix(query, "query WorkStation") || strings.Contains(query, "mutation") {
		return failure("CONFIG_ERROR")
	}
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return failure("SCHEMA_ERROR")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, origin+"/graphql", bytes.NewReader(body))
	if err != nil {
		return failure("CONFIG_ERROR")
	}
	// Personal read API keys use the raw value, unlike OAuth access tokens.
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return failure("DEGRADED")
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return &ws.Failure{Category: "RATE_LIMITED", RetryAfter: retryAfter(resp.Header)}
	case http.StatusUnauthorized:
		return failure("AUTH_INVALID")
	case http.StatusForbidden, http.StatusNotFound:
		return failure("ACCESS_DENIED")
	case http.StatusOK, http.StatusBadRequest:
	default:
		return failure("DEGRADED")
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return failure("DEGRADED")
	}
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Extensions struct {
				Code string `json:"code"`
				Type string `json:"type"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if len(raw) > maxBody || json.Unmarshal(raw, &envelope) != nil {
		return failure("SCHEMA_ERROR")
	}
	if len(envelope.Errors) > 0 {
		category := "SCHEMA_ERROR"
		for _, e := range envelope.Errors {
			for _, code := range []string{e.Extensions.Code, e.Extensions.Type} {
				switch strings.ToUpper(code) {
				case "RATELIMITED", "RATE_LIMITED":
					return &ws.Failure{Category: "RATE_LIMITED", RetryAfter: retryAfter(resp.Header)}
				case "UNAUTHENTICATED", "AUTHENTICATION_ERROR", "AUTHENTICATION_REQUIRED":
					category = "AUTH_INVALID"
				case "FORBIDDEN", "PERMISSION_ERROR", "ENTITY_NOT_FOUND", "NOT_FOUND":
					if category != "AUTH_INVALID" {
						category = "ACCESS_DENIED"
					}
				}
			}
		}
		return failure(category)
	}
	if resp.StatusCode != http.StatusOK || len(envelope.Data) == 0 || string(envelope.Data) == "null" || json.Unmarshal(envelope.Data, out) != nil {
		return failure("SCHEMA_ERROR")
	}
	return nil
}

func validID(id string) bool {
	return id != "" && len(id) <= 256 && !strings.ContainsFunc(id, unicode.IsControl)
}

func (p *Provider) Probe(ctx context.Context, token string) (ws.Identity, error) {
	var data struct {
		Viewer       *entity `json:"viewer"`
		Organization *entity `json:"organization"`
	}
	if err := p.query(ctx, token, probeQuery, nil, &data); err != nil {
		return ws.Identity{}, err
	}
	if data.Viewer == nil || data.Organization == nil || !validID(data.Viewer.ID) || !validID(data.Organization.ID) {
		return ws.Identity{}, failure("SCHEMA_ERROR")
	}
	return ws.Identity{PrincipalID: data.Viewer.ID, Source: ws.SourceIdentity{Provider: "linear", Origin: origin, TenantID: data.Organization.ID}}, nil
}

type entity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}
type pageInfo struct {
	HasNext *bool  `json:"hasNextPage"`
	End     string `json:"endCursor"`
}
type connection[T any] struct {
	Nodes []T      `json:"nodes"`
	Info  pageInfo `json:"pageInfo"`
}

func (info pageInfo) next(current string, count int) (string, error) {
	if info.HasNext == nil {
		return "", failure("SCHEMA_ERROR")
	}
	if !*info.HasNext {
		return "", nil
	}
	if count == 0 || info.End == "" || info.End == current || len(info.End) > 4096 {
		return "", failure("SCHEMA_ERROR")
	}
	return info.End, nil
}
func teamScope(t team) (ws.Scope, error) {
	if !validID(t.ID) || t.Name == "" || t.Key == "" {
		return ws.Scope{}, failure("SCHEMA_ERROR")
	}
	return ws.Scope{ID: t.ID, Name: t.Key + " · " + t.Name}, nil
}
func (p *Provider) DiscoverScopes(ctx context.Context, token, after string) (ws.Page, error) {
	if len(after) > 4096 {
		return ws.Page{}, ws.ErrInvalid
	}
	var data struct {
		Teams *connection[team] `json:"teams"`
	}
	if err := p.query(ctx, token, teamsQuery, map[string]any{"after": nullable(after)}, &data); err != nil {
		return ws.Page{}, err
	}
	if data.Teams == nil || data.Teams.Nodes == nil {
		return ws.Page{}, failure("SCHEMA_ERROR")
	}
	out := ws.Page{}
	seen := map[string]bool{}
	for _, t := range data.Teams.Nodes {
		scope, err := teamScope(t)
		if err != nil {
			return ws.Page{}, err
		}
		if seen[scope.ID] {
			return ws.Page{}, failure("SCHEMA_ERROR")
		}
		seen[scope.ID] = true
		out.Scopes = append(out.Scopes, scope)
	}
	var err error
	out.Next, err = data.Teams.Info.next(after, len(data.Teams.Nodes))
	return out, err
}
func (p *Provider) ResolveScope(ctx context.Context, token, id string) (ws.Scope, error) {
	if !validID(id) {
		return ws.Scope{}, ws.ErrInvalid
	}
	var data struct {
		Team *team `json:"team"`
	}
	if err := p.query(ctx, token, teamQuery, map[string]any{"id": id}, &data); err != nil {
		return ws.Scope{}, err
	}
	if data.Team == nil {
		return ws.Scope{}, failure("ACCESS_DENIED")
	}
	if data.Team.ID != id {
		return ws.Scope{}, failure("SCHEMA_ERROR")
	}
	return teamScope(*data.Team)
}
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type issue struct {
	ID          string          `json:"id"`
	Identifier  string          `json:"identifier"`
	Title       string          `json:"title"`
	Description json.RawMessage `json:"description"`
	URL         string          `json:"url"`
	Created     time.Time       `json:"createdAt"`
	Updated     time.Time       `json:"updatedAt"`
	Completed   *time.Time      `json:"completedAt"`
	Canceled    *time.Time      `json:"canceledAt"`
	Team        *entity         `json:"team"`
	State       *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"state"`
	Creator  *entity             `json:"creator"`
	Assignee *entity             `json:"assignee"`
	Project  *entity             `json:"project"`
	Labels   *connection[entity] `json:"labels"`
}

func normalize(r issue, scope ws.Scope) (*v1.WorkStationItem, error) {
	if !validID(r.ID) || r.Identifier == "" || r.Title == "" || r.Created.IsZero() || r.Updated.IsZero() || r.State == nil || r.State.Name == "" || r.State.Type == "" || r.Team == nil || !validID(r.Team.ID) || r.Labels == nil || r.Labels.Nodes == nil || len(r.Description) == 0 {
		return nil, failure("SCHEMA_ERROR")
	}
	u, err := url.Parse(r.URL)
	if err != nil || u.Scheme != "https" || u.Host != "linear.app" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
		return nil, failure("SCHEMA_ERROR")
	}
	item := &v1.WorkStationItem{Provider: "linear", Kind: "ISSUE", RemoteId: r.ID, Identifier: r.Identifier, ScopeId: scope.ID, ScopeName: scope.Name, Title: r.Title, DescriptionAvailable: true, Url: u.String(), RawState: r.State.Name, Status: "OTHER", SourceCreatedAt: timestamppb.New(r.Created), SourceUpdatedAt: timestamppb.New(r.Updated)}
	if item.SourceCreatedAt.CheckValid() != nil || item.SourceUpdatedAt.CheckValid() != nil {
		return nil, failure("SCHEMA_ERROR")
	}
	if string(r.Description) != "null" && json.Unmarshal(r.Description, &item.Description) != nil {
		return nil, failure("SCHEMA_ERROR")
	}
	if r.Creator != nil {
		if !validID(r.Creator.ID) {
			return nil, failure("SCHEMA_ERROR")
		}
		item.Author = r.Creator.ID
	}
	if r.Assignee != nil {
		if !validID(r.Assignee.ID) {
			return nil, failure("SCHEMA_ERROR")
		}
		item.Assignees = []string{r.Assignee.ID + ":" + r.Assignee.Name}
	}
	if r.Project != nil {
		if !validID(r.Project.ID) || r.Project.Name == "" {
			return nil, failure("SCHEMA_ERROR")
		}
		item.ProjectId, item.ProjectName = r.Project.ID, r.Project.Name
	}
	for _, label := range r.Labels.Nodes {
		if !validID(label.ID) || label.Name == "" {
			return nil, failure("SCHEMA_ERROR")
		}
		item.Labels = append(item.Labels, label.Name)
	}
	sort.Strings(item.Labels)
	var ended *time.Time
	switch r.State.Type {
	case "triage", "backlog", "unstarted":
		item.Status = "NOT_STARTED"
	case "started":
		item.Status = "IN_PROGRESS"
	case "completed":
		item.Status, item.Terminal, ended = "DONE", true, r.Completed
	case "canceled", "duplicate":
		item.Status, item.Terminal, ended = "CANCELED", true, r.Canceled
	}
	if item.Terminal && ended != nil && !ended.IsZero() {
		item.EndedAt = timestamppb.New(*ended)
		if item.EndedAt.CheckValid() != nil {
			return nil, failure("SCHEMA_ERROR")
		}
	}
	return item, nil
}

// retentionFilter keeps unknown/new state types and terminal issues without a
// reliable terminal time. The latter must count as PARTIAL, not silently vanish.
func retentionFilter(now time.Time) map[string]any {
	cutoff := now.UTC().Add(-90 * 24 * time.Hour).Format(time.RFC3339Nano)
	return map[string]any{"or": []any{
		map[string]any{"state": map[string]any{"type": map[string]any{"nin": []string{"completed", "canceled", "duplicate"}}}},
		map[string]any{"completedAt": map[string]any{"gte": cutoff}},
		map[string]any{"canceledAt": map[string]any{"gte": cutoff}},
		map[string]any{"state": map[string]any{"type": map[string]any{"eq": "completed"}}, "completedAt": map[string]any{"null": true}},
		map[string]any{"state": map[string]any{"type": map[string]any{"in": []string{"canceled", "duplicate"}}}, "canceledAt": map[string]any{"null": true}},
	}}
}

func (p *Provider) completeLabels(ctx context.Context, token string, r *issue) error {
	if r.Labels == nil || r.Labels.Nodes == nil {
		return failure("SCHEMA_ERROR")
	}
	next, err := r.Labels.Info.next("", len(r.Labels.Nodes))
	seen := map[string]bool{}
	for page := 1; next != "" && err == nil; page++ {
		if page >= maxPages {
			return failure("UNSUPPORTED")
		}
		if seen[next] {
			return failure("SCHEMA_ERROR")
		}
		seen[next] = true
		var data struct {
			Issue *issue `json:"issue"`
		}
		if err = p.query(ctx, token, labelsQuery, map[string]any{"id": r.ID, "after": next}, &data); err != nil {
			return err
		}
		part := data.Issue
		// A changed item or Team while hydrating metadata cannot be published as
		// one coherent observation. The next complete scan retries it.
		if part == nil || part.ID != r.ID || part.Team == nil || r.Team == nil || part.Team.ID != r.Team.ID || !part.Updated.Equal(r.Updated) || part.Labels == nil || part.Labels.Nodes == nil {
			return failure("SCHEMA_ERROR")
		}
		for _, label := range part.Labels.Nodes {
			r.Labels.Nodes = append(r.Labels.Nodes, label)
		}
		next, err = part.Labels.Info.next(next, len(part.Labels.Nodes))
	}
	return err
}

// Scan publishes only a complete traversal of one selected Team. All requests
// filter in that Team, include archived issues and never include sub-teams.
func (p *Provider) Scan(ctx context.Context, token string, scope ws.Scope, now time.Time) (ws.Scan, error) {
	if !validID(scope.ID) {
		return ws.Scan{}, ws.ErrInvalid
	}
	result := ws.Scan{Coverage: "COMPLETE"}
	items := map[string]*v1.WorkStationItem{}
	after := ""
	seen := map[string]bool{}
	for page := 0; ; page++ {
		if page >= maxPages {
			return ws.Scan{}, failure("UNSUPPORTED")
		}
		var data struct {
			Team *struct {
				team
				Issues *connection[issue] `json:"issues"`
			} `json:"team"`
		}
		if err := p.query(ctx, token, issuesQuery, map[string]any{"id": scope.ID, "after": nullable(after), "filter": retentionFilter(now)}, &data); err != nil {
			return ws.Scan{}, err
		}
		if data.Team == nil {
			return ws.Scan{}, failure("ACCESS_DENIED")
		}
		if data.Team.ID != scope.ID || data.Team.Issues == nil || data.Team.Issues.Nodes == nil {
			return ws.Scan{}, failure("SCHEMA_ERROR")
		}
		fresh, err := teamScope(data.Team.team)
		if err != nil {
			return ws.Scan{}, err
		}
		scope.Name = fresh.Name
		for _, remote := range data.Team.Issues.Nodes {
			if remote.Team == nil || !validID(remote.Team.ID) {
				return ws.Scan{}, failure("SCHEMA_ERROR")
			}
			// Never retain a moved issue under its former selected Team.
			if remote.Team.ID != scope.ID {
				continue
			}
			if err := p.completeLabels(ctx, token, &remote); err != nil {
				return ws.Scan{}, err
			}
			item, err := normalize(remote, scope)
			if err != nil {
				return ws.Scan{}, err
			}
			if old := items[item.RemoteId]; old == nil || item.SourceUpdatedAt.AsTime().After(old.SourceUpdatedAt.AsTime()) {
				items[item.RemoteId] = item
			}
		}
		next, err := data.Team.Issues.Info.next(after, len(data.Team.Issues.Nodes))
		if err != nil {
			return ws.Scan{}, err
		}
		if next == "" {
			break
		}
		if seen[next] {
			return ws.Scan{}, failure("SCHEMA_ERROR")
		}
		seen[next], after = true, next
	}
	if _, err := p.ResolveScope(ctx, token, scope.ID); err != nil {
		return ws.Scan{}, err
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
