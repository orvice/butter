package gitlab

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string, headers http.Header) *http.Response {
	if headers == nil {
		headers = http.Header{}
	}
	return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}
}

const projectJSON = `{"id":9007199254740993,"path_with_namespace":"group/subgroup/repo","issues_enabled":true,"merge_requests_enabled":true}`
const issueJSON = `{"iid":9007199254740994,"project_id":9007199254740993,"title":"Old open issue","description":"**plain source**","state":"opened","created_at":"2020-01-01T00:00:00Z","updated_at":"2020-02-01T00:00:00Z","author":{"id":9007199254740995,"username":"writer"},"assignees":[{"id":9007199254740996,"username":"owner"}],"labels":["bug"]}`
const mrJSON = `{"iid":2,"project_id":9007199254740993,"title":"Merged work","description":null,"state":"merged","draft":true,"created_at":"2026-09-01T00:00:00Z","updated_at":"2026-11-01T00:00:00Z","closed_at":"2026-10-30T00:00:00Z","merged_at":"2026-10-31T00:00:00Z","author":{"id":1},"assignees":[],"labels":[]}`

func newProvider(t *testing.T, fn transportFunc) *Provider {
	t.Helper()
	p, err := New("https://gitlab.example.com/gitlab/api/v4/", "https://gitlab.example.com/gitlab/", fn)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func scan(t *testing.T, p *Provider) (ws.Scan, error) {
	t.Helper()
	return p.Scan(context.Background(), "test-secret", ws.Scope{ID: "9007199254740993", Name: "old/name"}, time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC))
}
func category(t *testing.T, err error, want string) {
	t.Helper()
	var failure *ws.Failure
	if !errors.As(err, &failure) || failure.Category != want {
		t.Fatalf("want %s, got %v", want, err)
	}
	if strings.Contains(err.Error(), "test-secret") {
		t.Fatal("secret leaked in error")
	}
}

func TestFullProjectScanReadOnlyIdentityAndMappings(t *testing.T) {
	calls := 0
	p := newProvider(t, func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != "GET" || r.URL.Host != "gitlab.example.com" || r.Header.Get("PRIVATE-TOKEN") != "test-secret" || r.URL.Query().Get("private_token") != "" || r.Body != nil {
			t.Fatalf("escaped read profile: %s %s", r.Method, r.URL.Path)
		}
		switch r.URL.Path {
		case "/gitlab/api/v4/projects/9007199254740993":
			return response(200, projectJSON, nil), nil
		case "/gitlab/api/v4/projects/9007199254740993/issues":
			if r.URL.Query().Get("scope") != "all" || r.URL.Query().Get("state") != "all" {
				t.Fatal("omitted full project scope/state")
			}
			if r.URL.Query().Get("page") == "1" {
				next := *r.URL
				q := next.Query()
				q.Set("page", "2")
				next.RawQuery = q.Encode()
				return response(200, "["+issueJSON+"]", http.Header{"X-Next-Page": {"2"}, "Link": {fmt.Sprintf(`<%s>; rel="next"`, next.String())}}), nil
			}
			return response(200, "[]", http.Header{"X-Next-Page": {""}}), nil
		case "/gitlab/api/v4/projects/9007199254740993/merge_requests":
			if r.URL.Query().Get("scope") != "all" || r.URL.Query().Get("state") != "all" {
				t.Fatal("omitted MR scope/state")
			}
			return response(200, "["+mrJSON+"]", nil), nil
		default:
			t.Fatalf("unexpected endpoint %s", r.URL.Path)
			return nil, nil
		}
	})
	result, err := scan(t, p)
	if err != nil || result.Coverage != "COMPLETE" || len(result.Items) != 2 || calls != 5 {
		t.Fatalf("scan: %+v %v, calls %d", result, err, calls)
	}
	issue, mr := result.Items[0], result.Items[1]
	if issue.RemoteId != "9007199254740994" || issue.Author != "9007199254740995" || issue.Assignees[0] != "9007199254740996:owner" || issue.ScopeId != "9007199254740993" || issue.Status != "NOT_STARTED" || issue.Terminal {
		t.Fatalf("issue identity/state: %v", issue)
	}
	if issue.Url != "https://gitlab.example.com/gitlab/group/subgroup/repo/-/issues/9007199254740994" || issue.Description != "**plain source**" {
		t.Fatalf("issue source: %v", issue)
	}
	if mr.Kind != "MERGE_REQUEST" || mr.Status != "MERGED" || !mr.Terminal || !mr.Draft || mr.EndedAt.AsTime().Day() != 31 || mr.Identifier != "!2" {
		t.Fatalf("MR mapping: %v", mr)
	}
}

func TestNoFailedTraversalReturnsPublishableItems(t *testing.T) {
	for _, scenario := range []string{"second_page", "repeat_page", "repeat_cursor", "foreign_link", "changed_query", "null_schema", "wrong_project", "revoked", "project_moved", "feature_disabled", "unknown_draft", "redirect"} {
		t.Run(scenario, func(t *testing.T) {
			projectReads := 0
			p := newProvider(t, func(r *http.Request) (*http.Response, error) {
				if strings.HasSuffix(r.URL.Path, "/projects/9007199254740993") {
					projectReads++
					if projectReads > 1 {
						if scenario == "revoked" {
							return response(404, "test-secret private project", nil), nil
						}
						if scenario == "project_moved" {
							return response(200, strings.Replace(projectJSON, "group/subgroup/repo", "moved/repo", 1), nil), nil
						}
					}
					if scenario == "feature_disabled" {
						return response(200, strings.Replace(projectJSON, `"issues_enabled":true`, `"issues_enabled":false`, 1), nil), nil
					}
					return response(200, projectJSON, nil), nil
				}
				if strings.HasSuffix(r.URL.Path, "/merge_requests") {
					if scenario == "unknown_draft" {
						return response(200, "["+strings.Replace(mrJSON, `"draft":true,`, "", 1)+"]", nil), nil
					}
					return response(200, "[]", nil), nil
				}
				if scenario == "null_schema" {
					return response(200, "null", nil), nil
				}
				if scenario == "wrong_project" {
					return response(200, "["+strings.Replace(issueJSON, `"project_id":9007199254740993`, `"project_id":2`, 1)+"]", nil), nil
				}
				if scenario == "redirect" {
					return response(302, "", http.Header{"Location": {"https://evil.invalid"}}), nil
				}
				if scenario == "revoked" || scenario == "project_moved" || scenario == "unknown_draft" {
					return response(200, "["+issueJSON+"]", nil), nil
				}
				if r.URL.Query().Get("page") == "2" {
					if scenario == "repeat_page" {
						return response(200, "["+issueJSON+"]", http.Header{"X-Next-Page": {"3"}}), nil
					}
					return response(500, "test-secret failed page", nil), nil
				}
				next := *r.URL
				q := next.Query()
				q.Set("page", "2")
				if scenario == "repeat_cursor" {
					q.Set("page", "1")
				}
				if scenario == "changed_query" {
					q.Set("scope", "created_by_me")
				}
				if scenario == "foreign_link" {
					next.Host = "evil.invalid"
				}
				next.RawQuery = q.Encode()
				return response(200, "["+issueJSON+"]", http.Header{"Link": {fmt.Sprintf(`<%s>; rel="next"`, next.String())}}), nil
			})
			result, err := scan(t, p)
			if err == nil || len(result.Items) != 0 || result.Coverage != "" {
				t.Fatalf("failed scan published earlier page: %+v %v", result, err)
			}
			if strings.Contains(err.Error(), "test-secret") {
				t.Fatal("raw error leaked secret")
			}
		})
	}
}

func TestTerminalRetentionAndMissingEndTime(t *testing.T) {
	for _, state := range []string{"closed", "merged"} {
		t.Run(state, func(t *testing.T) {
			p := newProvider(t, func(r *http.Request) (*http.Response, error) {
				if strings.HasSuffix(r.URL.Path, "/projects/9007199254740993") {
					return response(200, projectJSON, nil), nil
				}
				if strings.HasSuffix(r.URL.Path, "/issues") {
					closed := strings.Replace(issueJSON, `"state":"opened"`, `"state":"closed","closed_at":"2020-02-01T00:00:00Z"`, 1)
					return response(200, "["+closed+"]", nil), nil
				}
				mr := strings.Replace(mrJSON, `"state":"merged"`, `"state":"`+state+`"`, 1)
				mr = strings.ReplaceAll(mr, `"closed_at":"2026-10-30T00:00:00Z"`, `"closed_at":null`)
				mr = strings.ReplaceAll(mr, `"merged_at":"2026-10-31T00:00:00Z"`, `"merged_at":null`)
				return response(200, "["+mr+"]", nil), nil
			})
			result, err := scan(t, p)
			if err != nil || result.Coverage != "PARTIAL" || result.Unresolved != 1 || len(result.Items) != 0 {
				t.Fatalf("terminal time was invented or retention ignored: %+v %v", result, err)
			}
		})
	}
}

func TestProbeDiscoveryVersionAndExactNumbers(t *testing.T) {
	p := newProvider(t, func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/gitlab/api/v4/user":
			return response(200, `{"id":9007199254740993}`, nil), nil
		case "/gitlab/api/v4/version":
			return response(200, `{"version":"18.5.0-ee","revision":"abc123"}`, nil), nil
		case "/gitlab/api/v4/projects":
			return response(200, "["+projectJSON+"]", nil), nil
		default:
			return response(200, projectJSON, nil), nil
		}
	})
	id, err := p.Probe(context.Background(), "test-secret")
	if err != nil || id.PrincipalID != "9007199254740993" || id.Source.Provider != "gitlab" || id.Source.Origin != "https://gitlab.example.com/gitlab/api/v4" {
		t.Fatalf("principal/source: %+v %v", id, err)
	}
	profile, err := p.InspectProfile(context.Background(), "test-secret")
	if err != nil || profile.Version != "18.5.0-ee" || profile.Revision != "abc123" {
		t.Fatal(profile, err)
	}
	page, err := p.DiscoverScopes(context.Background(), "test-secret", "")
	if err != nil || len(page.Scopes) != 1 || page.Scopes[0].ID != "9007199254740993" || page.Next != "" {
		t.Fatal(page, err)
	}
	scope, err := p.ResolveScope(context.Background(), "test-secret", page.Scopes[0].ID)
	if err != nil || scope.Name != "group/subgroup/repo" {
		t.Fatal(scope, err)
	}
}

func TestErrorCategoriesNeverContainBodies(t *testing.T) {
	for _, tc := range []struct {
		status int
		want   string
	}{{401, "AUTH_INVALID"}, {403, "ACCESS_DENIED"}, {404, "ACCESS_DENIED"}, {429, "RATE_LIMITED"}, {400, "UNSUPPORTED"}, {422, "UNSUPPORTED"}, {410, "FEATURE_UNAVAILABLE"}, {500, "DEGRADED"}, {200, "SCHEMA_ERROR"}} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			p := newProvider(t, func(*http.Request) (*http.Response, error) {
				return response(tc.status, "test-secret in upstream body", http.Header{"Retry-After": {"120"}}), nil
			})
			_, err := p.Probe(context.Background(), "test-secret")
			category(t, err, tc.want)
			if tc.status == 429 && ws.Classify(err).RetryAfter != 120*time.Second {
				t.Fatal("Retry-After ignored")
			}
		})
	}
	p := newProvider(t, func(*http.Request) (*http.Response, error) { return nil, errors.New("test-secret transport error") })
	_, err := p.Probe(context.Background(), "test-secret")
	category(t, err, "DEGRADED")
}

func TestPaginationBoundsAndHeaderAgreement(t *testing.T) {
	p := newProvider(t, func(*http.Request) (*http.Response, error) { return nil, nil })
	q := url.Values{"page": {"1"}, "per_page": {"100"}}
	for _, tc := range []struct {
		name        string
		headers     http.Header
		page, count int
		want        string
	}{
		{"repeat", http.Header{"X-Next-Page": {"1"}}, 1, 1, "SCHEMA_ERROR"},
		{"bad_current", http.Header{"X-Page": {"2"}}, 1, 1, "SCHEMA_ERROR"},
		{"known_over_cap", http.Header{"X-Total-Pages": {"501"}}, 1, 1, "UNSUPPORTED"},
		{"next_over_cap", http.Header{"X-Next-Page": {"501"}}, 500, 1, "UNSUPPORTED"},
		{"implicit_over_cap", nil, 500, 100, "UNSUPPORTED"},
		{"empty_next", http.Header{"X-Next-Page": {"2"}}, 1, 0, "SCHEMA_ERROR"},
		{"premature_end", http.Header{"X-Next-Page": {""}, "X-Total-Pages": {"2"}}, 1, 1, "SCHEMA_ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := p.nextPage(tc.headers, "/projects", q, tc.page, tc.count)
			category(t, err, tc.want)
		})
	}
	if next, err := p.nextPage(nil, "/projects", q, 1, 100); err != nil || next != "2" {
		t.Fatal("unannotated full page truncated", next, err)
	}
	if next, err := p.nextPage(http.Header{"X-Next-Page": {""}}, "/projects", q, 1, 100); err != nil || next != "" {
		t.Fatal(next, err)
	}
}

func TestRejectUnapprovedProfileAndUnsafeScope(t *testing.T) {
	for _, tc := range []struct{ api, web string }{
		{"http://gitlab.example.com/api/v4", "http://gitlab.example.com"},
		{"https://gitlab.example.com/api/v3", "https://gitlab.example.com"},
		{"https://gitlab.example.com/api/v4", "https://evil.invalid"},
		{"https://user:pass@gitlab.example.com/api/v4", "https://gitlab.example.com"},
		{"https://gitlab.example.com/api/v4?token=x", "https://gitlab.example.com"},
		{"https://gitlab.example.com/prefix/../api/v4", "https://gitlab.example.com/prefix/.."},
	} {
		if _, err := New(tc.api, tc.web, transportFunc(func(*http.Request) (*http.Response, error) { return nil, nil })); err == nil {
			t.Fatalf("accepted %s", tc.api)
		}
	}
	p := newProvider(t, func(*http.Request) (*http.Response, error) { t.Fatal("invalid id reached HTTP"); return nil, nil })
	for _, id := range []string{"../1", "1/2", "01", "1?token=x", "9007199254740993.0"} {
		if _, err := p.ResolveScope(context.Background(), "test-secret", id); !errors.Is(err, ws.ErrInvalid) {
			t.Fatal(id, err)
		}
	}
}
