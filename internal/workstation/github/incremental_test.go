package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestFirstPage304DoesNotSkipRemainingPages(t *testing.T) {
	secondCalls := 0
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/repositories/9007199254740993" {
			return response(200, repoJSON, nil), nil
		}
		if strings.HasSuffix(r.URL.Path, "/pulls") {
			return response(200, "[]", nil), nil
		}
		if r.URL.Query().Get("page") == "1" {
			if r.Header.Get("If-None-Match") == `"issues-page-one"` {
				return response(304, "", nil), nil
			}
			q := r.URL.Query()
			q.Set("page", "2")
			return response(200, "["+issueJSON+"]", http.Header{"Etag": {`"issues-page-one"`}, "Link": {fmt.Sprintf(`<https://api.github.com/repos/org/repo/issues?%s>; rel="next"`, q.Encode())}}), nil
		}
		secondCalls++
		if secondCalls == 2 && r.Header.Get("If-None-Match") != `"issues-page-two-old"` {
			t.Fatal("page two did not revalidate its own ETag")
		}
		body := strings.Replace(issueJSON, `"number":1`, `"number":3`, 1)
		etag := `"issues-page-two-old"`
		if secondCalls == 2 {
			body = strings.Replace(body, `"title":"Issue"`, `"title":"Changed second page"`, 1)
			etag = `"issues-page-two-new"`
		}
		return response(200, "["+body+"]", http.Header{"Etag": {etag}}), nil
	}))
	if _, err := scan(t, p); err != nil {
		t.Fatal(err)
	}
	got, err := scan(t, p)
	if err != nil {
		t.Fatal(err)
	}
	if secondCalls != 2 || len(got.Items) != 2 {
		t.Fatalf("skipped subsequent page: %+v calls=%d", got, secondCalls)
	}
	for _, item := range got.Items {
		if item.RemoteId == "3" && item.Title != "Changed second page" {
			t.Fatal("cached first-page 304 concealed changed next page")
		}
	}
}

func TestETagCacheIsolatedByCredentialCompleteRequestAndProfile(t *testing.T) {
	calls := 0
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if r.Header.Get("If-None-Match") != "" {
				t.Fatal("unexpected initial ETag")
			}
			return response(200, `{"id":1}`, http.Header{"Etag": {`"first-owner"`}}), nil
		case 2:
			if r.Header.Get("If-None-Match") != "" {
				t.Fatal("credential cache crossed identities")
			}
			return response(200, `{"id":2}`, http.Header{"Etag": {`"second-owner"`}}), nil
		default:
			if r.Header.Get("If-None-Match") != `"first-owner"` {
				t.Fatal("first owner lost isolated cache")
			}
			return response(304, "", nil), nil
		}
	}))
	a, err := p.Probe(context.Background(), "first-secret")
	if err != nil || a.PrincipalID != "1" {
		t.Fatal(a, err)
	}
	b, err := p.Probe(context.Background(), "second-secret")
	if err != nil || b.PrincipalID != "2" {
		t.Fatal(b, err)
	}
	a, err = p.Probe(context.Background(), "first-secret")
	if err != nil || a.PrincipalID != "1" {
		t.Fatal(a, err)
	}
	request, _ := http.NewRequest("GET", origin+"/repos/org/repo/issues?page=1&state=all", nil)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	first := requestKey("first-secret", request)
	for _, change := range []func(*http.Request){
		func(r *http.Request) { r.URL.RawQuery = "page=2&state=all" },
		func(r *http.Request) { r.Header.Set("Accept", "application/vnd.github.text+json") },
		func(r *http.Request) { r.Header.Set("X-GitHub-Api-Version", "different-profile") },
	} {
		copy := request.Clone(context.Background())
		u := *request.URL
		copy.URL = &u
		change(copy)
		if requestKey("first-secret", copy) == first {
			t.Fatal("cache ignored request/profile identity")
		}
	}
}

func TestETagAuthorizationFailureEvictsPriorRepresentation(t *testing.T) {
	calls := 0
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			return response(200, `{"id":1}`, http.Header{"Etag": {`"old"`}}), nil
		case 2:
			return response(401, "upstream-secret", nil), nil
		default:
			if r.Header.Get("If-None-Match") != "" {
				t.Fatal("unauthorized cached response retained")
			}
			return response(304, "", nil), nil
		}
	}))
	if _, err := p.Probe(context.Background(), "key"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Probe(context.Background(), "key"); ws.Classify(err).Category != "AUTH_INVALID" {
		t.Fatal(err)
	}
	if _, err := p.Probe(context.Background(), "key"); ws.Classify(err).Category != "SCHEMA_ERROR" {
		t.Fatal(err)
	}
}

func TestResponseCacheBounds(t *testing.T) {
	cache := responseCache{}
	for n := 0; n < 200; n++ {
		key := [32]byte{byte(n)}
		cache.put(key, `"valid"`, make([]byte, 512<<10), http.Header{"Link": {"safe"}, "Set-Cookie": {"do-not-cache"}})
	}
	if cache.bytes > cacheBytes || len(cache.entries) > cacheEntries {
		t.Fatal("unbounded cache")
	}
	for _, entry := range cache.entries {
		if entry.header.Get("Set-Cookie") != "" {
			t.Fatal("cached unrelated header")
		}
	}
	cache.put([32]byte{201}, `"valid"`, make([]byte, cacheEntryBytes+1), nil)
	if _, ok := cache.get([32]byte{201}); ok {
		t.Fatal("oversized entry cached")
	}
	cache.put([32]byte{202}, "bad\r\nheader", []byte("{}"), nil)
	if _, ok := cache.get([32]byte{202}); ok {
		t.Fatal("unsafe ETag accepted")
	}
}

func priorItem(t *testing.T, number, kind string, terminal bool, now time.Time) *v1.WorkStationItem {
	t.Helper()
	var raw remoteItem
	text := strings.Replace(issueJSON, `"number":1`, `"number":`+number, 1)
	if kind == "PULL_REQUEST" {
		text = strings.Replace(prJSON, `"number":2`, `"number":`+number, 1)
	}
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		t.Fatal(err)
	}
	if terminal {
		raw.State = "closed"
		ended := now.Add(-time.Hour)
		raw.Closed = &ended
	} else {
		raw.State = "open"
		raw.Merged = nil
		raw.Closed = nil
	}
	item, err := normalize(raw, ws.Scope{ID: "9007199254740993", Name: "org/repo"}, kind)
	if err != nil {
		t.Fatal(err)
	}
	return item
}
func incrementalScope(now time.Time) ws.Scope {
	return ws.Scope{ID: "9007199254740993", Name: "org/repo", Checkpoint: now.Add(-10 * time.Minute), LastFull: now.Add(-time.Hour), Coverage: "COMPLETE"}
}
func TestIncrementalMergesChangesAndReconcilesMissingOpenItems(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	previous := []*v1.WorkStationItem{priorItem(t, "1", "ISSUE", false, now), priorItem(t, "2", "PULL_REQUEST", false, now), priorItem(t, "3", "ISSUE", true, now), priorItem(t, "4", "ISSUE", false, now)}
	mergedPR := strings.Replace(prJSON, `"updated_at":"2026-11-01T00:00:00Z"`, `"updated_at":"2026-11-03T11:59:00Z"`, 1)
	closedIssue := strings.Replace(issueJSON, `"number":1`, `"number":4`, 1)
	closedIssue = strings.Replace(closedIssue, `"state":"open"`, `"state":"closed","closed_at":"2026-11-03T11:58:00Z"`, 1)
	newIssue := strings.Replace(issueJSON, `"number":1`, `"number":5`, 1)
	checks := map[string]int{}
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" {
			t.Fatal("non-read request")
		}
		checks[r.URL.Path]++
		switch r.URL.Path {
		case "/repositories/9007199254740993":
			return response(200, repoJSON, nil), nil
		case "/repos/org/repo/issues":
			if r.URL.Query().Get("state") == "all" {
				if r.URL.Query().Get("since") != "2026-11-03T11:45:00Z" {
					t.Fatal("checkpoint overlap omitted")
				}
				return response(200, "[]", nil), nil
			}
			return response(200, "["+newIssue+"]", nil), nil
		case "/repos/org/repo/pulls":
			if r.URL.Query().Get("state") == "open" {
				return response(200, "[]", nil), nil
			}
			return response(200, "["+mergedPR+"]", nil), nil
		case "/repos/org/repo/issues/1":
			return response(404, "upstream-secret", nil), nil
		case "/repos/org/repo/issues/4":
			return response(200, closedIssue, nil), nil
		default:
			t.Fatalf("unexpected request: %s", r.URL)
			return nil, nil
		}
	}))
	scan, err := p.ScanIncremental(context.Background(), "key", incrementalScope(now), now, previous)
	if err != nil {
		t.Fatal(err)
	}
	if scan.Mode != "INCREMENTAL" || scan.Coverage != "COMPLETE" || len(scan.Items) != 4 || checks["/repos/org/repo/issues/1"] != 1 || checks["/repos/org/repo/issues/4"] != 1 {
		t.Fatalf("scan=%+v requests=%+v", scan, checks)
	}
	for _, item := range scan.Items {
		switch item.RemoteId {
		case "1":
			t.Fatal("deleted item retained")
		case "2":
			if item.Status != "MERGED" {
				t.Fatal("merged PR missed")
			}
		case "3":
			if !item.Terminal {
				t.Fatal("unchanged terminal item lost")
			}
		case "4":
			if item.Status != "DONE" {
				t.Fatal("missing-open item was not rechecked")
			}
		case "5":
			if item.Terminal {
				t.Fatal("new open item missed")
			}
		default:
			t.Fatal("unexpected item")
		}
	}
	if previous[3].Terminal {
		t.Fatal("mutated previous immutable observation")
	}
}

func TestIncrementalPRStopsAtOverlapButStillTraversesAllOpen(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	old := strings.Replace(prJSON, `"number":2`, `"number":8`, 1)
	open := strings.Replace(issueJSON, `"number":1`, `"number":9`, 1)
	allOpenCalls := 0
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/repositories/9007199254740993" {
			return response(200, repoJSON, nil), nil
		}
		if strings.HasSuffix(r.URL.Path, "/issues") {
			return response(200, "[]", nil), nil
		}
		if r.URL.Query().Get("state") == "all" {
			if r.URL.Query().Get("page") != "1" {
				t.Fatal("read beyond complete sorted overlap")
			}
			q := r.URL.Query()
			q.Set("page", "2")
			return response(200, "["+old+"]", http.Header{"Link": {fmt.Sprintf(`<https://api.github.com/repos/org/repo/pulls?%s>; rel="next"`, q.Encode())}}), nil
		}
		allOpenCalls++
		if r.URL.Query().Get("page") == "1" {
			q := r.URL.Query()
			q.Set("page", "2")
			return response(200, "["+open+"]", http.Header{"Link": {fmt.Sprintf(`<https://api.github.com/repos/org/repo/pulls?%s>; rel="next"`, q.Encode())}}), nil
		}
		return response(200, "[]", nil), nil
	}))
	scan, err := p.ScanIncremental(context.Background(), "key", incrementalScope(now), now, nil)
	if err != nil || allOpenCalls != 2 || len(scan.Items) != 1 || scan.Items[0].RemoteId != "9" {
		t.Fatalf("scan=%+v err=%v open-calls=%d", scan, err, allOpenCalls)
	}
}

func TestIncrementalNeverPublishesOnLaterPageOrPermissionFailure(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	for _, status := range []int{403, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/repositories/9007199254740993" {
					return response(200, repoJSON, nil), nil
				}
				if strings.HasSuffix(r.URL.Path, "/issues/1") {
					return response(status, "upstream-secret", nil), nil
				}
				return response(200, "[]", nil), nil
			}))
			scan, err := p.ScanIncremental(context.Background(), "key", incrementalScope(now), now, []*v1.WorkStationItem{priorItem(t, "1", "ISSUE", false, now)})
			if err == nil || len(scan.Items) > 0 {
				t.Fatalf("partial incremental published: %+v, %v", scan, err)
			}
			if status == 403 && ws.Classify(err).Category != "ACCESS_DENIED" {
				t.Fatal(err)
			}
		})
	}
}

func TestIncrementalMissingAndTransferredItemsAreDropped(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	for _, status := range []int{301, 302, 307, 308, 404, 410} {
		p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == "/repositories/9007199254740993" {
				return response(200, repoJSON, nil), nil
			}
			if strings.HasSuffix(r.URL.Path, "/issues/1") {
				return response(status, "upstream-secret", http.Header{"Location": {"https://unapproved.example/item"}}), nil
			}
			return response(200, "[]", nil), nil
		}))
		scan, err := p.ScanIncremental(context.Background(), "key", incrementalScope(now), now, []*v1.WorkStationItem{priorItem(t, "1", "ISSUE", false, now)})
		if err != nil || len(scan.Items) > 0 || scan.Coverage != "COMPLETE" {
			t.Fatalf("status %d: %+v %v", status, scan, err)
		}
	}
}

func TestIncrementalIncompleteCoverageFallsBackToFull(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/repositories/9007199254740993" {
			return response(200, repoJSON, nil), nil
		}
		if r.URL.Query().Get("since") != "" || r.URL.Query().Get("state") != "all" {
			t.Fatal("partial coverage incorrectly used as incremental base")
		}
		return response(200, "[]", nil), nil
	}))
	scope := incrementalScope(now)
	scope.Coverage = "PARTIAL"
	scope.Unresolved = 1
	scan, err := p.ScanIncremental(context.Background(), "key", scope, now, nil)
	if err != nil || scan.Coverage != "COMPLETE" || scan.Mode == "INCREMENTAL" {
		t.Fatalf("fallback = %+v %v", scan, err)
	}
}

func TestItemPagesRejectOutOfOrderUpdatedBoundaryAndFailure(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	for _, mode := range []string{"out-of-order", "later-page-failure"} {
		p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
			if mode == "out-of-order" {
				newer := strings.Replace(prJSON, `"updated_at":"2026-11-01T00:00:00Z"`, `"updated_at":"2026-11-03T11:59:00Z"`, 1)
				return response(200, "["+prJSON+","+newer+"]", nil), nil
			}
			if r.URL.Query().Get("page") == "2" {
				return response(500, "upstream-secret", nil), nil
			}
			q := r.URL.Query()
			q.Set("page", "2")
			return response(200, "["+prJSON+"]", http.Header{"Link": {fmt.Sprintf(`<https://api.github.com/repos/org/repo/pulls?%s>; rel="next"`, q.Encode())}}), nil
		}))
		cutoff := now.Add(-time.Hour)
		if mode == "later-page-failure" {
			cutoff = time.Time{}
		}
		rows, err := p.itemPages(context.Background(), "key", "/repos/org/repo/pulls", url.Values{"state": {"all"}, "sort": {"updated"}, "direction": {"desc"}, "per_page": {"100"}}, cutoff)
		if err == nil || len(rows) > 0 {
			t.Fatal("invalid traversal returned data", rows, err)
		}
	}
}

func TestItemMissingSentinelDoesNotHideRepositoryFailures(t *testing.T) {
	p := New(transportFunc(func(*http.Request) (*http.Response, error) { return response(404, "private", nil), nil }))
	var raw remoteItem
	_, itemErr := p.getResource(context.Background(), "key", "/repos/org/repo/issues/1", nil, &raw, true)
	if !errors.Is(itemErr, errItemMissing) {
		t.Fatal(itemErr)
	}
	_, repoErr := p.repository(context.Background(), "key", "1")
	if ws.Classify(repoErr).Category != "ACCESS_DENIED" {
		t.Fatal(repoErr)
	}
}
