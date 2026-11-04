// Package linear implements the read-only Linear GraphQL profile.
// The transport is injected for fixtures and controlled acceptance; callers
// cannot provide an upstream URL through a workspace request.
package linear

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const endpoint = "https://api.linear.app/graphql"
const maxBody = 8 << 20

type Provider struct{ client *http.Client }

func New(transport http.RoundTripper) *Provider {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Provider{client: &http.Client{Transport: transport, Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

type graphRequest struct {
	Query string         `json:"query"`
	Vars  map[string]any `json:"variables,omitempty"`
}
type graphResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func failure(category string) error { return &ws.Failure{Category: category} }

func (p *Provider) query(ctx context.Context, token, query string, vars map[string]any, out any) error {
	body, err := json.Marshal(graphRequest{Query: query, Vars: vars})
	if err != nil {
		return failure("SCHEMA_ERROR")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return failure("SCHEMA_ERROR")
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return failure("DEGRADED")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return failure("AUTH_INVALID")
	}
	if resp.StatusCode == http.StatusForbidden {
		return failure("ACCESS_DENIED")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		delay := time.Minute
		if n, e := strconv.Atoi(resp.Header.Get("Retry-After")); e == nil && n > 0 {
			delay = time.Duration(n) * time.Second
		}
		return &ws.Failure{Category: "RATE_LIMITED", RetryAfter: delay}
	}
	if resp.StatusCode != http.StatusOK {
		return failure("DEGRADED")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil || len(data) > maxBody {
		return failure("SCHEMA_ERROR")
	}
	var envelope graphResponse
	if json.Unmarshal(data, &envelope) != nil || len(envelope.Data) == 0 {
		return failure("SCHEMA_ERROR")
	}
	if len(envelope.Errors) > 0 {
		return failure("SCHEMA_ERROR")
	}
	if json.Unmarshal(envelope.Data, out) != nil {
		return failure("SCHEMA_ERROR")
	}
	return nil
}

func (p *Provider) Probe(ctx context.Context, token string) (ws.Identity, error) {
	var out struct {
		Viewer struct {
			ID string `json:"id"`
		} `json:"viewer"`
	}
	err := p.query(ctx, token, `query Viewer { viewer { id } }`, nil, &out)
	if err != nil {
		return ws.Identity{}, err
	}
	if out.Viewer.ID == "" {
		return ws.Identity{}, failure("SCHEMA_ERROR")
	}
	source, _ := (ws.SourceIdentity{Provider: "linear", Origin: endpoint, TenantID: out.Viewer.ID}).Canonical()
	return ws.Identity{PrincipalID: out.Viewer.ID, Source: source}, nil
}

func (p *Provider) DiscoverScopes(ctx context.Context, token, pageToken string) (ws.Page, error) {
	var out struct {
		Teams struct {
			Nodes    []struct{ ID, Name, Key string }
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			}
		}
	}
	vars := map[string]any{"after": nil}
	if pageToken != "" {
		vars["after"] = pageToken
	}
	err := p.query(ctx, token, `query Teams($after: String) { teams(first: 50, after: $after, includeArchived: false) { nodes { id name key } pageInfo { hasNextPage endCursor } } }`, vars, &out)
	if err != nil {
		return ws.Page{}, err
	}
	page := ws.Page{}
	for _, team := range out.Teams.Nodes {
		if team.ID == "" || team.Name == "" {
			return ws.Page{}, failure("SCHEMA_ERROR")
		}
		page.Scopes = append(page.Scopes, ws.Scope{ID: team.ID, Name: team.Name})
	}
	if out.Teams.PageInfo.HasNextPage {
		if out.Teams.PageInfo.EndCursor == "" {
			return ws.Page{}, failure("SCHEMA_ERROR")
		}
		page.Next = out.Teams.PageInfo.EndCursor
	}
	return page, nil
}

func (p *Provider) ResolveScope(ctx context.Context, token, id string) (ws.Scope, error) {
	var out struct {
		Team struct{ ID, Name string } `json:"team"`
	}
	err := p.query(ctx, token, `query Team($id: String!) { team(id: $id) { id name } }`, map[string]any{"id": id}, &out)
	if err != nil {
		return ws.Scope{}, err
	}
	if out.Team.ID == "" || out.Team.Name == "" {
		return ws.Scope{}, failure("ACCESS_DENIED")
	}
	return ws.Scope{ID: out.Team.ID, Name: out.Team.Name}, nil
}

func (p *Provider) Scan(ctx context.Context, token string, scope ws.Scope, now time.Time) (ws.Scan, error) {
	var out struct {
		Issues struct {
			Nodes []struct {
				ID, Identifier, Title, Description, URL string
				CreatedAt, UpdatedAt                    string
			}
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage"`
			}
		}
	}
	err := p.query(ctx, token, `query Issues($team: ID!) { issues(first: 100, filter: { team: { id: { eq: $team } } }) { nodes { id identifier title description url state { name type } createdAt updatedAt } pageInfo { hasNextPage } } }`, map[string]any{"team": scope.ID}, &out)
	if err != nil {
		return ws.Scan{}, err
	}
	result := ws.Scan{Coverage: "COMPLETE"}
	for _, issue := range out.Issues.Nodes {
		created, err1 := time.Parse(time.RFC3339, issue.CreatedAt)
		updated, err2 := time.Parse(time.RFC3339, issue.UpdatedAt)
		if issue.ID == "" || issue.Title == "" || err1 != nil || err2 != nil {
			return ws.Scan{}, failure("SCHEMA_ERROR")
		}
		result.Items = append(result.Items, &v1.WorkStationItem{Provider: "linear", Kind: "ISSUE", RemoteId: issue.ID, Identifier: issue.Identifier, ScopeId: scope.ID, ScopeName: scope.Name, Title: issue.Title, Description: issue.Description, DescriptionAvailable: issue.Description != "", Url: issue.URL, SourceCreatedAt: timestamppb.New(created), SourceUpdatedAt: timestamppb.New(updated), ObservedAt: timestamppb.New(now)})
	}
	if out.Issues.PageInfo.HasNextPage {
		result.Coverage = "PARTIAL"
		result.Unresolved = 1
	}
	return result, nil
}
