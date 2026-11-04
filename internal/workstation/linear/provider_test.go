package linear

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func fixtureClient(t *testing.T, handle func(gqlRequest) (int, http.Header, any)) *Provider {
	t.Helper()
	return New(roundTrip(func(req *http.Request) (*http.Response, error) {
		t.Helper()
		if req.Method != http.MethodPost || req.URL.String() != origin+"/graphql" {
			t.Fatalf("non-profile request: %s %s", req.Method, req.URL)
		}
		if req.Header.Get("Authorization") != "read-key" || req.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("wrong authentication or encoding")
		}
		var payload gqlRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(payload.Query, "query WorkStation") || strings.Contains(strings.ToLower(payload.Query), "mutation") {
			t.Fatalf("non-read query: %s", payload.Query)
		}
		status, headers, data := handle(payload)
		body, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(string(body))), Request: req}, nil
	}))
}
func okay(data any) (int, http.Header, any) { return 200, http.Header{}, map[string]any{"data": data} }
func failCategory(t *testing.T, err error, want string) {
	t.Helper()
	var failure *ws.Failure
	if !errors.As(err, &failure) || failure.Category != want {
		t.Fatalf("error = %v, want %s", err, want)
	}
	if strings.Contains(err.Error(), "read-key") || strings.Contains(err.Error(), "upstream-secret") {
		t.Fatalf("sensitive error: %s", err)
	}
}
func info(next string) map[string]any {
	return map[string]any{"hasNextPage": next != "", "endCursor": nullable(next)}
}
func fixtureTeam() map[string]any {
	return map[string]any{"id": "team-a", "name": "Engineering", "key": "ENG"}
}
func fixtureIssue(id, state string, now time.Time) map[string]any {
	return map[string]any{
		"id": id, "identifier": "ENG-1", "title": "Fixture " + id,
		"description": "**Markdown** remains plain source text", "url": "https://linear.app/test/issue/ENG-1/fixture",
		"createdAt": now.Add(-365 * 24 * time.Hour), "updatedAt": now,
		"team": map[string]any{"id": "team-a"}, "state": map[string]any{"id": "state-" + state, "name": "Visible " + state, "type": state},
		"creator": map[string]any{"id": "creator-a", "name": "Creator"}, "assignee": map[string]any{"id": "user-a", "name": "Ada"},
		"project": map[string]any{"id": "project-a", "name": "Release"},
		"labels":  map[string]any{"nodes": []any{map[string]any{"id": "label-a", "name": "Bug"}}, "pageInfo": info("")},
	}
}
func teamPage(items []any, next string) map[string]any {
	team := fixtureTeam()
	team["issues"] = map[string]any{"nodes": items, "pageInfo": info(next)}
	return map[string]any{"team": team}
}

func TestProbeBindsPrincipalAndOrganization(t *testing.T) {
	p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
		if req.Query != probeQuery {
			t.Fatal("wrong probe")
		}
		return okay(map[string]any{"viewer": map[string]any{"id": "user-a"}, "organization": map[string]any{"id": "org-a"}})
	})
	got, err := p.Probe(context.Background(), "read-key")
	if err != nil {
		t.Fatal(err)
	}
	if got.PrincipalID != "user-a" || got.Source != (ws.SourceIdentity{Provider: "linear", Origin: origin, TenantID: "org-a"}) {
		t.Fatalf("identity = %+v", got)
	}
}

func TestDiscoverAndResolveTeamScopes(t *testing.T) {
	calls := 0
	p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
		calls++
		switch calls {
		case 1:
			if req.Query != teamsQuery || req.Variables["after"] != nil || !strings.Contains(req.Query, "includeArchived: true") {
				t.Fatal("bad first team page")
			}
			return okay(map[string]any{"teams": map[string]any{"nodes": []any{fixtureTeam()}, "pageInfo": info("opaque-cursor")}})
		case 2:
			if req.Variables["after"] != "opaque-cursor" {
				t.Fatal("lost team cursor")
			}
			return okay(map[string]any{"teams": map[string]any{"nodes": []any{}, "pageInfo": info("")}})
		default:
			if req.Query != teamQuery || req.Variables["id"] != "team-a" {
				t.Fatal("scope was not resolved by stable id")
			}
			return okay(map[string]any{"team": fixtureTeam()})
		}
	})
	first, err := p.DiscoverScopes(context.Background(), "read-key", "")
	if err != nil || len(first.Scopes) != 1 || first.Scopes[0].ID != "team-a" || first.Scopes[0].Name != "ENG · Engineering" || first.Next != "opaque-cursor" {
		t.Fatalf("first = %+v, %v", first, err)
	}
	last, err := p.DiscoverScopes(context.Background(), "read-key", first.Next)
	if err != nil || len(last.Scopes) != 0 || last.Next != "" {
		t.Fatalf("last = %+v, %v", last, err)
	}
	scope, err := p.ResolveScope(context.Background(), "read-key", "team-a")
	if err != nil || scope.ID != "team-a" {
		t.Fatalf("scope = %+v, %v", scope, err)
	}
}

func TestScanStatesRetentionArchiveReopenAndMovedTeam(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	var rows []any
	for _, state := range []string{"triage", "backlog", "unstarted", "started", "completed", "canceled", "future-state"} {
		item := fixtureIssue(state, state, now)
		if state == "completed" {
			item["completedAt"] = now.Add(-90 * 24 * time.Hour)
		}
		if state == "canceled" {
			item["canceledAt"] = now.Add(-time.Hour)
		}
		if state == "started" {
			item["archivedAt"] = now.Add(-time.Hour)
		}
		rows = append(rows, item)
	}
	old := fixtureIssue("expired", "completed", now)
	old["completedAt"] = now.Add(-90*24*time.Hour - time.Nanosecond)
	rows = append(rows, old)
	reopened := fixtureIssue("reopened", "started", now)
	reopened["completedAt"] = now.Add(-365 * 24 * time.Hour)
	rows = append(rows, reopened)
	unresolved := fixtureIssue("unresolved", "duplicate", now)
	rows = append(rows, unresolved)
	moved := fixtureIssue("moved", "started", now)
	moved["team"] = map[string]any{"id": "other-team"}
	rows = append(rows, moved)
	pages := 0
	p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
		if req.Query == teamQuery {
			return okay(map[string]any{"team": fixtureTeam()})
		}
		if req.Query != issuesQuery || req.Variables["id"] != "team-a" || !strings.Contains(req.Query, "includeArchived: true") || !strings.Contains(req.Query, "includeSubTeams: false") || !strings.Contains(req.Query, "orderBy: updatedAt") {
			t.Fatal("unbounded team query")
		}
		filter, _ := json.Marshal(req.Variables["filter"])
		if !strings.Contains(string(filter), `"nin":["completed","canceled","duplicate"]`) || !strings.Contains(string(filter), `"null":true`) || !strings.Contains(string(filter), "2026-08-05T12:00:00Z") {
			t.Fatalf("bad retention filter: %s", filter)
		}
		pages++
		if pages == 1 {
			if req.Variables["after"] != nil {
				t.Fatal("first cursor")
			}
			return okay(teamPage(rows[:5], "next-page"))
		}
		if req.Variables["after"] != "next-page" {
			t.Fatal("lost issue cursor")
		}
		return okay(teamPage(rows[5:], ""))
	})
	scan, err := p.Scan(context.Background(), "read-key", ws.Scope{ID: "team-a"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if scan.Coverage != "PARTIAL" || scan.Unresolved != 1 || len(scan.Items) != 8 {
		t.Fatalf("scan = %+v", scan)
	}
	for _, item := range scan.Items {
		if item.ScopeName != "ENG · Engineering" || item.ProjectId != "project-a" || item.ProjectName != "Release" || len(item.Assignees) != 1 || item.Assignees[0] != "user-a:Ada" || len(item.Labels) != 1 || item.Labels[0] != "Bug" || !item.DescriptionAvailable {
			t.Fatalf("missing metadata: %+v", item)
		}
		switch item.RemoteId {
		case "triage", "backlog", "unstarted":
			if item.Status != "NOT_STARTED" || item.Terminal {
				t.Fatal(item)
			}
		case "started", "reopened":
			if item.Status != "IN_PROGRESS" || item.Terminal || item.EndedAt != nil {
				t.Fatal(item)
			}
		case "completed":
			if item.Status != "DONE" || !item.Terminal || !item.EndedAt.AsTime().Equal(now.Add(-90*24*time.Hour)) {
				t.Fatal(item)
			}
		case "canceled":
			if item.Status != "CANCELED" || !item.Terminal {
				t.Fatal(item)
			}
		case "future-state":
			if item.Status != "OTHER" || item.Terminal {
				t.Fatal(item)
			}
		default:
			t.Fatalf("unexpected retained item %s", item.RemoteId)
		}
	}
}

func TestScanHydratesLabelsWithoutTruncating(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	row := fixtureIssue("issue-a", "started", now)
	row["labels"].(map[string]any)["pageInfo"] = info("label-next")
	p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
		switch req.Query {
		case issuesQuery:
			return okay(teamPage([]any{row}, ""))
		case labelsQuery:
			if req.Variables["id"] != "issue-a" || req.Variables["after"] != "label-next" {
				t.Fatal("bad label cursor")
			}
			return okay(map[string]any{"issue": map[string]any{"id": "issue-a", "team": map[string]any{"id": "team-a"}, "updatedAt": now, "labels": map[string]any{"nodes": []any{map[string]any{"id": "label-b", "name": "Enhancement"}}, "pageInfo": info("")}}})
		case teamQuery:
			return okay(map[string]any{"team": fixtureTeam()})
		default:
			t.Fatal("unexpected query")
			return 0, nil, nil
		}
	})
	scan, err := p.Scan(context.Background(), "read-key", ws.Scope{ID: "team-a"}, now)
	if err != nil || len(scan.Items) != 1 || len(scan.Items[0].Labels) != 2 {
		t.Fatalf("scan = %+v, %v", scan, err)
	}
}

func TestGraphQLErrorsFailClosedAndRedact(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		code, want string
	}{
		{"partial success", 200, "INTERNAL_SERVER_ERROR", "SCHEMA_ERROR"},
		{"graphql rate", 200, "RATELIMITED", "RATE_LIMITED"},
		{"http rate", 429, "", "RATE_LIMITED"},
		{"graphql auth", 200, "UNAUTHENTICATED", "AUTH_INVALID"},
		{"http auth", 401, "", "AUTH_INVALID"},
		{"graphql forbidden", 200, "FORBIDDEN", "ACCESS_DENIED"},
		{"http forbidden", 403, "", "ACCESS_DENIED"},
		{"validation", 400, "GRAPHQL_VALIDATION_FAILED", "SCHEMA_ERROR"},
		{"server", 503, "", "DEGRADED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := fixtureClient(t, func(gqlRequest) (int, http.Header, any) {
				return tc.status, http.Header{"Retry-After": []string{"120"}}, map[string]any{"data": map[string]any{"viewer": map[string]any{"id": "user-a"}, "organization": map[string]any{"id": "org-a"}}, "errors": []any{map[string]any{"message": "upstream-secret read-key", "extensions": map[string]any{"code": tc.code}}}}
			})
			_, err := p.Probe(context.Background(), "read-key")
			failCategory(t, err, tc.want)
			if tc.want == "RATE_LIMITED" && ws.Classify(err).RetryAfter != 2*time.Minute {
				t.Fatalf("retry delay = %s", ws.Classify(err).RetryAfter)
			}
		})
	}
}

func TestScanNeverReturnsPartialPagesOnFailure(t *testing.T) {
	now := time.Now().UTC()
	for _, mode := range []string{"error", "repeated-cursor", "empty-next", "missing-page-info", "lost-team"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
				calls++
				if calls == 1 {
					return okay(teamPage([]any{fixtureIssue("first", "started", now)}, "page-two"))
				}
				switch mode {
				case "error":
					return 200, nil, map[string]any{"data": teamPage([]any{}, ""), "errors": []any{map[string]any{"message": "upstream-secret read-key"}}}
				case "repeated-cursor":
					return okay(teamPage([]any{fixtureIssue("second", "started", now)}, "page-two"))
				case "empty-next":
					return okay(teamPage([]any{}, "page-three"))
				case "missing-page-info":
					data := teamPage([]any{}, "")
					delete(data["team"].(map[string]any)["issues"].(map[string]any), "pageInfo")
					return okay(data)
				default:
					return okay(map[string]any{"team": nil})
				}
			})
			scan, err := p.Scan(context.Background(), "read-key", ws.Scope{ID: "team-a"}, now)
			if err == nil || len(scan.Items) != 0 || scan.Coverage != "" {
				t.Fatalf("partial published: %+v, %v", scan, err)
			}
		})
	}
}

func TestScanRechecksScopeAccessAfterCompleteTraversal(t *testing.T) {
	now := time.Now().UTC()
	p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
		if req.Query == issuesQuery {
			return okay(teamPage([]any{fixtureIssue("first", "started", now)}, ""))
		}
		return 403, nil, map[string]any{"message": "upstream-secret"}
	})
	scan, err := p.Scan(context.Background(), "read-key", ws.Scope{ID: "team-a"}, now)
	failCategory(t, err, "ACCESS_DENIED")
	if len(scan.Items) != 0 {
		t.Fatal("revoked scope leaked staged scan")
	}
}

func TestSchemaMismatchAndUnsafeSourceLinksAreRejected(t *testing.T) {
	now := time.Now().UTC()
	for _, change := range []func(map[string]any){
		func(row map[string]any) { delete(row, "description") },
		func(row map[string]any) { row["state"] = nil },
		func(row map[string]any) { row["updatedAt"] = "not-time" },
		func(row map[string]any) { row["url"] = "https://linear.app.evil.example/item" },
		func(row map[string]any) { row["url"] = "http://linear.app/item" },
		func(row map[string]any) { row["url"] = "https://person:password@linear.app/item" },
	} {
		row := fixtureIssue("first", "started", now)
		change(row)
		p := fixtureClient(t, func(gqlRequest) (int, http.Header, any) { return okay(teamPage([]any{row}, "")) })
		_, err := p.Scan(context.Background(), "read-key", ws.Scope{ID: "team-a"}, now)
		failCategory(t, err, "SCHEMA_ERROR")
	}
}

func TestEmptyTeamScanIsComplete(t *testing.T) {
	p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
		if req.Query == teamQuery {
			return okay(map[string]any{"team": fixtureTeam()})
		}
		return okay(teamPage([]any{}, ""))
	})
	scan, err := p.Scan(context.Background(), "read-key", ws.Scope{ID: "team-a"}, time.Now())
	if err != nil || scan.Coverage != "COMPLETE" || scan.Unresolved != 0 || len(scan.Items) != 0 {
		t.Fatalf("scan = %+v, %v", scan, err)
	}
}
