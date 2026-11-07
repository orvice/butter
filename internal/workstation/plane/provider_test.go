package plane

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
}

const projectJSON = `{"id":"project-1","name":"Shared project"}`
const itemJSON = `{"id":9007199254740993,"project_id":"project-1","identifier":"PROJ-1","name":"Read-only item","description_html":"<p>Details</p>","is_draft":false,"created_at":"2020-01-01T00:00:00Z","updated_at":"2026-11-01T00:00:00Z","created_by_id":"author-1","completed_at":null,"state_id":"state-1","state":{"id":"state-1","name":"Doing","group":"started"},"assignee_ids":["user-1"],"label_ids":["label-1"],"assignees":[{"id":"user-1","display_name":"Owner"}],"labels":[{"id":"label-1","name":"Bug"}]}`

func envelopeJSON(items, next string) string {
	if next == "" {
		return `{"data":[` + items + `],"next_cursor":null,"has_more":false,"pagination":{"style":"cursor"}}`
	}
	cursor, _ := json.Marshal(next)
	return `{"data":[` + items + `],"next_cursor":` + string(cursor) + `,"has_more":true,"pagination":{"style":"cursor"}}`
}
func newProvider(t *testing.T, handler transportFunc) *Provider {
	t.Helper()
	p, err := New("https://plane.example", "https://web.plane.example", "workspace-1", transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" || r.URL.Host != "plane.example" || r.Header.Get("X-Api-Key") != "test-secret" || !strings.HasPrefix(r.URL.Path, "/api/v2/") {
			t.Fatalf("request escaped fixed read-only profile: %s %s", r.Method, r.URL.Redacted())
		}
		return handler(r)
	}))
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func scan(t *testing.T, p *Provider) (ws.Scan, error) {
	t.Helper()
	return p.Scan(context.Background(), "test-secret", ws.Scope{ID: "project-1"}, time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC))
}
func TestProbeAndScopeDiscoveryUseOnlyV2(t *testing.T) {
	p := newProvider(t, func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/api/v2/users/me/":
			return response(200, `{"id":9007199254740993,"principal_kind":"api_key","scopes":[]}`), nil
		case "/api/v2/workspaces/workspace-1/projects/":
			if r.URL.Query().Get("paginate") != "cursor" || r.URL.Query().Get("include_archived") != "true" {
				t.Fatal("missing archive/deep paging profile")
			}
			if r.URL.Query().Get("cursor") == "" {
				return response(200, envelopeJSON(projectJSON, "https://untrusted.example/?cursor=1")), nil
			}
			if r.URL.Query().Get("cursor") != "https://untrusted.example/?cursor=1" {
				t.Fatal("cursor changed")
			}
			return response(200, envelopeJSON("", "")), nil
		case "/api/v2/workspaces/workspace-1/projects/project-1/":
			return response(200, projectJSON), nil
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
			return nil, nil
		}
	})
	identity, err := p.Probe(context.Background(), "test-secret")
	if err != nil || identity.PrincipalID != "9007199254740993" || identity.Source.TenantID != "workspace-slug:workspace-1" || identity.Source.Origin != "https://plane.example" {
		t.Fatalf("identity=%+v err=%v", identity, err)
	}
	page, err := p.DiscoverScopes(context.Background(), "test-secret", "")
	if err != nil || len(page.Scopes) != 1 || page.Next == "" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	last, err := p.DiscoverScopes(context.Background(), "test-secret", page.Next)
	if err != nil || last.Next != "" || len(last.Scopes) != 0 {
		t.Fatalf("last=%+v err=%v", last, err)
	}
	if _, err := p.ResolveScope(context.Background(), "test-secret", "project-1"); err != nil {
		t.Fatal(err)
	}
}
func TestScanIncludesArchivesExcludesDraftsAndResolvesMetadata(t *testing.T) {
	calls := 0
	p := newProvider(t, func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/project-1/") {
			return response(200, projectJSON), nil
		}
		q := r.URL.Query()
		if q.Get("archived") != "true" || q.Get("is_draft") != "false" || q.Get("paginate") != "cursor" || !strings.Contains(q.Get("fields"), "description_html") || q.Get("expand") != "state,assignees,labels" {
			t.Fatal("incomplete scan request profile")
		}
		calls++
		if q.Get("cursor") == "" {
			draft := strings.Replace(strings.Replace(itemJSON, `9007199254740993`, `"draft-1"`, 1), `"is_draft":false`, `"is_draft":true`, 1)
			return response(200, envelopeJSON(itemJSON+","+draft, "second-page")), nil
		}
		archived := strings.Replace(strings.Replace(itemJSON, `9007199254740993`, `"archived-1"`, 1), `"completed_at":null`, `"archived_at":"2026-10-01T00:00:00Z","completed_at":null`, 1)
		return response(200, envelopeJSON(archived, "")), nil
	})
	result, err := scan(t, p)
	if err != nil || result.Coverage != "COMPLETE" || len(result.Items) != 2 || calls != 2 {
		t.Fatalf("result=%+v err=%v calls=%d", result, err, calls)
	}
	item := result.Items[0]
	if item.RemoteId != "9007199254740993" || item.Status != "IN_PROGRESS" || item.Description != "<p>Details</p>" || !item.DescriptionAvailable || item.Assignees[0] != "user-1:Owner" || item.Labels[0] != "Bug" || !strings.HasPrefix(item.Url, "https://web.plane.example/workspace-1/projects/project-1/") {
		t.Fatalf("normalization=%v", item)
	}
}
func TestTerminalTimeCoverageAndRetention(t *testing.T) {
	for _, tc := range []struct {
		name, group, completed string
		count                  int
		coverage               string
		unresolved             uint32
	}{
		{"completed", "completed", `"2026-10-01T00:00:00Z"`, 1, "COMPLETE", 0},
		{"old_completed_edited", "completed", `"2025-10-01T00:00:00Z"`, 0, "COMPLETE", 0},
		{"unknown_completed", "completed", `null`, 0, "PARTIAL", 1},
		{"cancelled_no_verified_activity", "cancelled", `"2026-10-01T00:00:00Z"`, 0, "PARTIAL", 1},
		{"unknown_state", "future-group", `null`, 1, "COMPLETE", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := strings.Replace(strings.Replace(itemJSON, `"group":"started"`, `"group":"`+tc.group+`"`, 1), `"completed_at":null`, `"completed_at":`+tc.completed, 1)
			p := newProvider(t, func(r *http.Request) (*http.Response, error) {
				if strings.HasSuffix(r.URL.Path, "/project-1/") {
					return response(200, projectJSON), nil
				}
				return response(200, envelopeJSON(item, "")), nil
			})
			result, err := scan(t, p)
			if err != nil || len(result.Items) != tc.count || result.Coverage != tc.coverage || result.Unresolved != tc.unresolved {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			if tc.name == "unknown_state" && (result.Items[0].Terminal || result.Items[0].Status != "OTHER") {
				t.Fatal("unknown state became terminal")
			}
		})
	}
}
func TestFailedTraversalNeverReturnsPartialItems(t *testing.T) {
	for _, scenario := range []string{"later_page_403", "later_page_500", "repeated_cursor", "repeated_rows", "null_data", "v1_envelope", "offset_envelope", "missing_pagination", "metadata_missing", "description_unavailable", "schema_invalid", "scope_revoked", "scope_changed"} {
		t.Run(scenario, func(t *testing.T) {
			projects := 0
			p := newProvider(t, func(r *http.Request) (*http.Response, error) {
				if strings.HasSuffix(r.URL.Path, "/project-1/") {
					projects++
					if projects > 1 && scenario == "scope_revoked" {
						return response(403, "private secret"), nil
					}
					if projects > 1 && scenario == "scope_changed" {
						return response(200, strings.Replace(projectJSON, "Shared project", "Renamed", 1)), nil
					}
					return response(200, projectJSON), nil
				}
				if r.URL.Query().Get("cursor") == "" {
					switch scenario {
					case "null_data":
						return response(200, `{"data":null,"has_more":false,"next_cursor":null,"pagination":{"style":"cursor"}}`), nil
					case "v1_envelope":
						return response(200, `{"results":[],"next_cursor":null}`), nil
					case "offset_envelope":
						return response(200, `{"data":[],"next":null,"pagination":{"style":"offset"}}`), nil
					case "missing_pagination":
						return response(200, `{"data":[]}`), nil
					case "metadata_missing":
						return response(200, envelopeJSON(strings.Replace(itemJSON, `"display_name":"Owner"`, `"email":"should-not-leak@example.com"`, 1), "")), nil
					case "schema_invalid":
						return response(200, `{"data":[`), nil
					case "description_unavailable":
						return response(200, envelopeJSON(strings.Replace(itemJSON, `"description_html":"<p>Details</p>",`, "", 1), "")), nil
					case "scope_revoked", "scope_changed":
						return response(200, envelopeJSON(itemJSON, "")), nil
					}
					return response(200, envelopeJSON(itemJSON, "page-2")), nil
				}
				switch scenario {
				case "later_page_403":
					return response(403, "private token"), nil
				case "later_page_500":
					return response(500, "private token"), nil
				case "repeated_cursor":
					return response(200, envelopeJSON(strings.Replace(itemJSON, `9007199254740993`, `"item-2"`, 1), "page-2")), nil
				default:
					return response(200, envelopeJSON(itemJSON, "page-3")), nil
				}
			})
			result, err := scan(t, p)
			if scenario == "description_unavailable" {
				if err != nil || len(result.Items) != 1 || result.Items[0].DescriptionAvailable {
					t.Fatalf("unavailable description result=%+v err=%v", result, err)
				}
				return
			}
			if err == nil || len(result.Items) != 0 {
				t.Fatalf("failed traversal published items: %+v err=%v", result, err)
			}
			if strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "email") {
				t.Fatal("upstream content leaked")
			}
		})
	}
}
func TestStatusClassificationAndNoDynamicFallback(t *testing.T) {
	for _, tc := range []struct {
		status   int
		category string
	}{{401, "AUTH_INVALID"}, {403, "ACCESS_DENIED"}, {404, "ACCESS_DENIED"}, {402, "FEATURE_UNAVAILABLE"}, {410, "FEATURE_UNAVAILABLE"}, {400, "UNSUPPORTED"}, {429, "RATE_LIMITED"}, {302, "DEGRADED"}, {500, "DEGRADED"}} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			calls := 0
			p := newProvider(t, func(r *http.Request) (*http.Response, error) {
				calls++
				resp := response(tc.status, "credential upstream body")
				resp.Header.Set("Retry-After", "123")
				return resp, nil
			})
			_, err := p.Probe(context.Background(), "test-secret")
			f := ws.Classify(err)
			if f.Category != tc.category || calls != 1 {
				t.Fatalf("err=%v calls=%d", err, calls)
			}
			if tc.status == 429 && f.RetryAfter != 123*time.Second {
				t.Fatalf("retry=%v", f.RetryAfter)
			}
		})
	}
}
func TestConstructorAndScopePathValidation(t *testing.T) {
	for _, slug := range []string{"", "../other", "workspace?key=x", "x/y", "x%2fy"} {
		if _, err := New("https://plane.example", "https://web.example", slug, nil); err == nil {
			t.Fatalf("accepted slug %q", slug)
		}
	}
	for _, base := range []string{"http://plane.example", "https://user:secret@plane.example", "https://plane.example/?secret=x"} {
		if _, err := New(base, "https://web.example", "workspace", nil); err == nil {
			t.Fatalf("accepted base %q", base)
		}
	}
	p := newProvider(t, func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid ID issued HTTP request")
		return nil, nil
	})
	if _, err := p.ResolveScope(context.Background(), "test-secret", "../other"); err == nil {
		t.Fatal("accepted invalid project ID")
	}
}

func TestDeepCursorTraversalDoesNotUseOffset(t *testing.T) {
	const total = 10001
	read := 0
	p := newProvider(t, func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/project-1/") {
			return response(200, projectJSON), nil
		}
		q := r.URL.Query()
		if q.Has("offset") || q.Get("paginate") != "cursor" {
			t.Fatal("deep traversal used offset")
		}
		if read > 0 && q.Get("cursor") != fmt.Sprintf("after-%d", read) {
			t.Fatalf("cursor did not advance from previous boundary: %s", q.Get("cursor"))
		}
		rows := make([]string, 0, 100)
		for len(rows) < 100 && read < total {
			read++
			rows = append(rows, strings.Replace(itemJSON, `9007199254740993`, fmt.Sprintf(`"item-%d"`, read), 1))
		}
		cursor := ""
		if read < total {
			cursor = fmt.Sprintf("after-%d", read)
		}
		return response(200, envelopeJSON(strings.Join(rows, ","), cursor)), nil
	})
	result, err := scan(t, p)
	if err != nil || result.Coverage != "COMPLETE" || len(result.Items) != total || read != total {
		t.Fatalf("deep result count=%d coverage=%s read=%d err=%v", len(result.Items), result.Coverage, read, err)
	}
}
