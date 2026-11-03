package github

import (
	"context"
	"encoding/json"
	"errors"
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
func response(status int, body string, headers http.Header) *http.Response {
	if headers == nil {
		headers = http.Header{}
	}
	return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}
}

const repoJSON = `{"id":9007199254740993,"full_name":"org/repo","has_issues":true}`
const issueJSON = `{"number":1,"title":"Issue","body":"**description**","state":"open","created_at":"2020-01-01T00:00:00Z","updated_at":"2026-11-01T00:00:00Z","user":{"id":1,"login":"author"},"assignees":[{"id":2,"login":"owner"}],"labels":[{"name":"bug"}]}`
const prJSON = `{"number":2,"title":"Merged PR","body":null,"state":"closed","draft":true,"created_at":"2026-10-01T00:00:00Z","updated_at":"2026-11-01T00:00:00Z","closed_at":"2026-10-31T00:00:00Z","merged_at":"2026-10-30T00:00:00Z","user":{"id":1},"assignees":[],"labels":[]}`

func scan(t *testing.T, p *Provider) (ws.Scan, error) {
	t.Helper()
	return p.Scan(context.Background(), "test-secret", ws.Scope{ID: "9007199254740993"}, time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC))
}
func TestReadOnlyMultiPageScanSeparatesIssuesAndPulls(t *testing.T) {
	calls := []string{}
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" || r.URL.Host != "api.github.com" || r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Fatal("request escaped fixed read profile")
		}
		calls = append(calls, r.URL.String())
		switch r.URL.Path {
		case "/repositories/9007199254740993":
			return response(200, repoJSON, nil), nil
		case "/repos/org/repo/issues":
			if r.URL.Query().Get("page") == "1" {
				q := r.URL.Query()
				q.Set("page", "2")
				return response(200, "["+issueJSON+",{\"number\":2,\"pull_request\":{\"url\":\"ignored\"}}]", http.Header{"Link": {fmt.Sprintf(`<https://api.github.com/repos/org/repo/issues?%s>; rel="next"`, q.Encode())}}), nil
			}
			return response(200, "[]", nil), nil
		case "/repos/org/repo/pulls":
			return response(200, "["+prJSON+"]", nil), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	}))
	result, err := scan(t, p)
	if err != nil {
		t.Fatal(err)
	}
	if result.Coverage != "COMPLETE" || len(result.Items) != 2 {
		t.Fatalf("result: %+v", result)
	}
	for _, item := range result.Items {
		if item.Kind == "PULL_REQUEST" && (item.Status != "MERGED" || !item.Draft || item.EndedAt.AsTime().Day() != 30) {
			t.Fatalf("PR mapping: %v", item)
		}
		if item.Kind == "ISSUE" && item.RemoteId != "1" {
			t.Fatal("PR leaked into issues")
		}
	}
	if len(calls) != 5 {
		t.Fatalf("requests: %v", calls)
	}
}
func TestNoFailurePublishesEarlierPages(t *testing.T) {
	for _, scenario := range []string{"page_failure", "foreign_next", "repeated_cursor", "overlap_page", "not_modified", "null_schema", "scope_revoked", "redirect", "incomplete_time"} {
		t.Run(scenario, func(t *testing.T) {
			checks := 0
			p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/repositories/9007199254740993" {
					checks++
					if scenario == "scope_revoked" && checks > 1 {
						return response(404, `{"message":"private test-secret"}`, nil), nil
					}
					return response(200, repoJSON, nil), nil
				}
				if strings.HasSuffix(r.URL.Path, "/pulls") {
					return response(200, "[]", nil), nil
				}
				if scenario == "not_modified" {
					return response(http.StatusNotModified, "", http.Header{"ETag": {`"cached-page"`}}), nil
				}
				if scenario == "null_schema" {
					return response(200, "null", nil), nil
				}
				if scenario == "redirect" {
					return response(302, "", http.Header{"Location": {"https://evil.invalid/steal"}}), nil
				}
				if scenario == "incomplete_time" {
					return response(200, "["+strings.Replace(issueJSON, `"state":"open"`, `"state":"closed"`, 1)+"]", nil), nil
				}
				if scenario == "scope_revoked" {
					return response(200, "["+issueJSON+"]", nil), nil
				}
				if r.URL.Query().Get("page") == "2" {
					if scenario == "overlap_page" {
						return response(200, "["+issueJSON+"]", nil), nil
					}
					return response(500, "test-secret", nil), nil
				}
				next := *r.URL
				q := next.Query()
				q.Set("page", "2")
				next.RawQuery = q.Encode()
				if scenario == "foreign_next" {
					next.Host = "evil.invalid"
				}
				if scenario == "repeated_cursor" {
					q.Set("page", "1")
					next.RawQuery = q.Encode()
				}
				return response(200, "["+issueJSON+"]", http.Header{"Link": {fmt.Sprintf(`<%s>; rel="next"`, next.String())}}), nil
			}))
			result, err := scan(t, p)
			if scenario == "incomplete_time" {
				if err != nil || result.Coverage != "PARTIAL" || result.Unresolved != 1 || len(result.Items) != 0 {
					t.Fatalf("missing terminal time: %+v %v", result, err)
				}
				return
			}
			if err == nil || len(result.Items) != 0 {
				t.Fatalf("failed traversal returned publishable data: %+v %v", result, err)
			}
			if strings.Contains(err.Error(), "test-secret") {
				t.Fatal("raw error leaked token")
			}
		})
	}
}
func TestETagCacheReusesBodyOnlyForMatchingCredential(t *testing.T) {
	cache := NewMemoryETagCache()
	calls := 0
	p := NewWithETagCache(transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/user" {
			return response(500, "unexpected", nil), nil
		}
		if r.Header.Get("Authorization") == "Bearer first" {
			if r.Header.Get("If-None-Match") == `"user-v1"` {
				return response(http.StatusNotModified, "", http.Header{"ETag": {`"user-v1"`}}), nil
			}
			return response(200, `{"id":1}`, http.Header{"ETag": {`"user-v1"`}}), nil
		}
		if r.Header.Get("If-None-Match") != "" {
			t.Fatal("credential reused another credential's validator")
		}
		return response(200, `{"id":2}`, http.Header{"ETag": {`"user-v2"`}}), nil
	}), cache)
	identity, err := p.Probe(context.Background(), "first")
	if err != nil || identity.PrincipalID != "1" {
		t.Fatal(identity, err)
	}
	identity, err = p.Probe(context.Background(), "first")
	if err != nil || identity.PrincipalID != "1" {
		t.Fatal("cached 304 was not decoded", identity, err)
	}
	identity, err = p.Probe(context.Background(), "second")
	if err != nil || identity.PrincipalID != "2" {
		t.Fatal("credential cache isolation", identity, err)
	}
	if calls != 3 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestUncachedNotModifiedFailsClosed(t *testing.T) {
	p := NewWithETagCache(transportFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusNotModified, "", http.Header{"ETag": {`"missing"`}}), nil
	}), NewMemoryETagCache())
	_, err := p.Probe(context.Background(), "token")
	var failure *ws.Failure
	if !errors.As(err, &failure) || failure.Category != "SCHEMA_ERROR" {
		t.Fatalf("want schema error, got %v", err)
	}
}
func TestErrorCategoriesAndRedaction(t *testing.T) {
	for _, tc := range []struct {
		status  int
		headers http.Header
		want    string
	}{{401, nil, "AUTH_INVALID"}, {404, nil, "ACCESS_DENIED"}, {403, nil, "ACCESS_DENIED"}, {403, http.Header{"X-Ratelimit-Remaining": {"0"}, "X-Ratelimit-Reset": {"0"}}, "RATE_LIMITED"}, {403, http.Header{"Retry-After": {"120"}}, "RATE_LIMITED"}, {429, http.Header{"Retry-After": {"120"}}, "RATE_LIMITED"}, {500, nil, "DEGRADED"}, {200, nil, "SCHEMA_ERROR"}} {
		t.Run(fmt.Sprint(tc.status, tc.want), func(t *testing.T) {
			p := New(transportFunc(func(*http.Request) (*http.Response, error) {
				return response(tc.status, "test-secret", tc.headers), nil
			}))
			_, err := p.Probe(context.Background(), "test-secret")
			var f *ws.Failure
			if !errors.As(err, &f) || f.Category != tc.want {
				t.Fatalf("want %s got %v", tc.want, err)
			}
			if strings.Contains(err.Error(), "test-secret") {
				t.Fatal("leaked upstream body")
			}
		})
	}
	p := New(transportFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("test-secret in transport") }))
	_, err := p.Probe(context.Background(), "test-secret")
	if err == nil || strings.Contains(err.Error(), "test-secret") {
		t.Fatal("transport error not redacted")
	}
}
func TestDiscoveryAndScopeIdentityAreDecimalStrings(t *testing.T) {
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/user":
			return response(200, `{"id":9007199254740993}`, nil), nil
		case "/user/repos":
			return response(200, "["+repoJSON+"]", nil), nil
		default:
			return response(200, repoJSON, nil), nil
		}
	}))
	identity, err := p.Probe(context.Background(), "token")
	if err != nil || identity.PrincipalID != "9007199254740993" {
		t.Fatal(identity, err)
	}
	page, err := p.DiscoverScopes(context.Background(), "token", "")
	if err != nil || page.Scopes[0].ID != "9007199254740993" {
		t.Fatal(page, err)
	}
	for _, id := range []string{"../user", "https://evil.invalid", "1?x=2", "01", "0"} {
		if _, err := p.ResolveScope(context.Background(), "token", id); !errors.Is(err, ws.ErrInvalid) {
			t.Fatalf("unsafe scope %q accepted", id)
		}
	}
}
func TestRetainsOldOpenUsesActualEndAndClearsReopenedEnd(t *testing.T) {
	var raw remoteItem
	if err := json.Unmarshal([]byte(issueJSON), &raw); err != nil {
		t.Fatal(err)
	}
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	raw.Closed = &old
	item, err := normalize(raw, ws.Scope{ID: "1", Name: "org/repo"}, "ISSUE")
	if err != nil || item.Terminal || item.EndedAt != nil {
		t.Fatal("reopened retains close time")
	}
	if !ws.Retained(item, time.Now()) {
		t.Fatal("old open lost")
	}
	raw.State = "closed"
	item, err = normalize(raw, ws.Scope{ID: "1", Name: "org/repo"}, "ISSUE")
	if err != nil || ws.Retained(item, time.Now()) {
		t.Fatal("edit time extended end window")
	}
}

func TestUnknownStateRemainsVisibleAndUnmapped(t *testing.T) {
	var raw remoteItem
	if err := json.Unmarshal([]byte(issueJSON), &raw); err != nil {
		t.Fatal(err)
	}
	raw.State = "future_state"
	item, err := normalize(raw, ws.Scope{ID: "1", Name: "org/repo"}, "ISSUE")
	if err != nil || item.Terminal || item.Status != "OTHER" || !ws.Retained(item, time.Now()) {
		t.Fatal("unknown state guessed terminal", item, err)
	}
}
