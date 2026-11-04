package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

func TestBatchScanResumesFailedPageAcrossProvidersAndVerifiesAccess(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	scope := ws.Scope{ID: "9007199254740993"}
	pageOneCalls, pageTwoCalls, repositoryCalls := 0, 0, 0
	transport := transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" || r.URL.Host != "api.github.com" {
			t.Fatal("escaped read profile")
		}
		switch r.URL.Path {
		case "/repositories/9007199254740993":
			repositoryCalls++
			if repositoryCalls == 2 {
				return response(403, "secret body", nil), nil
			}
			return response(200, repoJSON, nil), nil
		case "/repos/org/repo/issues":
			if r.URL.Query().Get("page") == "1" {
				pageOneCalls++
				q := r.URL.Query()
				q.Set("page", "2")
				return response(200, "["+issueJSON+"]", http.Header{"Link": {fmt.Sprintf(`<https://api.github.com/repos/org/repo/issues?%s>; rel="next"`, q.Encode())}}), nil
			}
			pageTwoCalls++
			if pageTwoCalls == 1 {
				return response(500, "secret body", nil), nil
			}
			return response(200, "["+strings.Replace(issueJSON, `"number":1`, `"number":3`, 1)+"]", nil), nil
		case "/repos/org/repo/pulls":
			return response(200, "["+prJSON+"]", nil), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	})
	p := New(transport)
	initial, err := p.ScanBatch(context.Background(), "secret-key", scope, now, "")
	if err != nil || initial.Complete || initial.Next == "" || len(initial.Items) != 0 {
		t.Fatalf("initial=%+v %v", initial, err)
	}
	first, err := p.ScanBatch(context.Background(), "secret-key", scope, now, initial.Next)
	if err != nil || first.Complete || len(first.Items) != 1 {
		t.Fatalf("first=%+v %v", first, err)
	}
	failed, err := p.ScanBatch(context.Background(), "secret-key", scope, now, first.Next)
	if err == nil || failed.Next != "" || failed.Coverage != "" || len(failed.Items) > 0 {
		t.Fatal("failed page returned staging data")
	}
	// A new process/provider resumes solely from its persisted cursor.
	p = New(transport)
	second, err := p.ScanBatch(context.Background(), "secret-key", scope, now, first.Next)
	if err != nil || second.Complete || len(second.Items) != 1 || second.Items[0].RemoteId != "3" {
		t.Fatalf("second=%+v %v", second, err)
	}
	pulls, err := p.ScanBatch(context.Background(), "secret-key", scope, now, second.Next)
	if err != nil || pulls.Complete || len(pulls.Items) != 1 || pulls.Items[0].Status != "MERGED" {
		t.Fatalf("pulls=%+v %v", pulls, err)
	}
	denied, err := p.ScanBatch(context.Background(), "secret-key", scope, now, pulls.Next)
	if err == nil || denied.Complete || ws.Classify(err).Category != "ACCESS_DENIED" {
		t.Fatalf("denied=%+v %v", denied, err)
	}
	final, err := New(transport).ScanBatch(context.Background(), "secret-key", scope, now, pulls.Next)
	if err != nil || !final.Complete || final.Next != "" || final.Coverage != "COMPLETE" {
		t.Fatalf("final=%+v %v", final, err)
	}
	if pageOneCalls != 1 || pageTwoCalls != 2 || repositoryCalls != 3 {
		t.Fatalf("replayed completed pages: first=%d second=%d repo=%d", pageOneCalls, pageTwoCalls, repositoryCalls)
	}
	for _, cursor := range []string{initial.Next, first.Next, second.Next, pulls.Next} {
		if len(cursor) > maxBatchCursor || strings.Contains(cursor, "secret-key") || strings.Contains(cursor, "description") {
			t.Fatal("cursor retains a secret or item body")
		}
	}
}

func TestBatchCursorBindingAndValidationPrecedeRequests(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	scope := ws.Scope{ID: "9007199254740993"}
	valid := batchCursor{Profile: "github-full-v1", Scope: scope.ID, Started: now, Name: "org/repo", Phase: "issues", Page: 1}
	encode := func(c batchCursor) string { b, _ := json.Marshal(c); return string(b) }
	badScope := valid
	badScope.Scope = "2"
	badTime := valid
	badTime.Started = now.Add(time.Second)
	badPhase := valid
	badPhase.Phase = "../user"
	badPage := valid
	badPage.Page = 0
	badName := valid
	badName.Name = "org/repo?url=https://other.example"
	badHash := valid
	badHash.Pages = []byte{1}
	p := New(transportFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid cursor caused a request")
		return nil, nil
	}))
	for _, raw := range []string{encode(badScope), encode(badTime), encode(badPhase), encode(badPage), encode(badName), encode(badHash), encode(valid) + "{}", strings.Repeat("x", maxBatchCursor+1), `{"unknown":true}`} {
		if _, err := p.ScanBatch(context.Background(), "key", scope, now, raw); !errors.Is(err, ws.ErrInvalid) {
			t.Fatalf("cursor accepted: %v", err)
		}
	}
}

func TestBatchRepeatedPageAndRedirectFailWithoutCompletion(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	scope := ws.Scope{ID: "9007199254740993"}
	for _, mode := range []string{"repeated-page", "redirect"} {
		p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == "/repositories/9007199254740993" {
				return response(200, repoJSON, nil), nil
			}
			if r.URL.Query().Get("page") == "2" && mode == "redirect" {
				return response(302, "", http.Header{"Location": {"https://unapproved.example/"}}), nil
			}
			q := r.URL.Query()
			q.Set("page", fmt.Sprint(2))
			if r.URL.Query().Get("page") == "2" {
				q.Set("page", "3")
			}
			return response(200, "["+issueJSON+"]", http.Header{"Link": {fmt.Sprintf(`<https://api.github.com/repos/org/repo/issues?%s>; rel="next"`, q.Encode())}}), nil
		}))
		initial, err := p.ScanBatch(context.Background(), "key", scope, now, "")
		if err != nil {
			t.Fatal(err)
		}
		first, err := p.ScanBatch(context.Background(), "key", scope, now, initial.Next)
		if err != nil {
			t.Fatal(err)
		}
		last, err := p.ScanBatch(context.Background(), "key", scope, now, first.Next)
		if err == nil || last.Complete || last.Next != "" || len(last.Items) > 0 {
			t.Fatal("invalid page returned a successful batch")
		}
	}
}

func TestBatchPreservesOverlappingUnresolvedMarkersForRuntimeDeduplication(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	scope := ws.Scope{ID: "9007199254740993"}
	unresolved := strings.Replace(issueJSON, `"state":"open"`, `"state":"closed"`, 1)
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/repositories/9007199254740993" {
			return response(200, repoJSON, nil), nil
		}
		if r.URL.Query().Get("page") == "1" {
			q := r.URL.Query()
			q.Set("page", "2")
			return response(200, "["+unresolved+"]", http.Header{"Link": {fmt.Sprintf(`<https://api.github.com/repos/org/repo/issues?%s>; rel="next"`, q.Encode())}}), nil
		}
		return response(200, "["+unresolved+","+strings.Replace(issueJSON, `"number":1`, `"number":3`, 1)+"]", nil), nil
	}))
	initial, err := p.ScanBatch(context.Background(), "key", scope, now, "")
	if err != nil {
		t.Fatal(err)
	}
	first, err := p.ScanBatch(context.Background(), "key", scope, now, initial.Next)
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.ScanBatch(context.Background(), "key", scope, now, first.Next)
	if err != nil {
		t.Fatal(err)
	}
	if first.Unresolved != 1 || first.Coverage != "PARTIAL" || len(first.Items) != 1 || second.Unresolved != 1 || len(second.Items) != 2 {
		t.Fatalf("lost unresolved markers: first=%+v second=%+v", first, second)
	}
}

func TestBatchRetainsTerminalMarkersBeforeRuntimeSourceChoice(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 123456789, time.UTC)
	scope := ws.Scope{ID: "9007199254740993"}
	expired := strings.Replace(issueJSON, `"state":"open"`, `"state":"closed","closed_at":"2020-01-01T00:00:00Z"`, 1)
	unknown := strings.Replace(strings.Replace(issueJSON, `"number":1`, `"number":3`, 1), `"state":"open"`, `"state":"closed"`, 1)
	p := New(transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/repositories/9007199254740993" {
			return response(200, repoJSON, nil), nil
		}
		return response(200, "["+expired+","+unknown+"]", nil), nil
	}))
	initial, err := p.ScanBatch(context.Background(), "key", scope, now, "")
	if err != nil {
		t.Fatal(err)
	}
	// Simulate the fixed start round trip through BSON Date millisecond storage.
	page, err := p.ScanBatch(context.Background(), "key", scope, now.Truncate(time.Millisecond), initial.Next)
	if err != nil || len(page.Items) != 2 || page.Unresolved != 1 {
		t.Fatalf("markers lost before source choice: %+v %v", page, err)
	}
	for _, item := range page.Items {
		if item.RemoteId == "1" && (!item.Terminal || item.EndedAt == nil || ws.Retained(item, now)) {
			t.Fatal("expired marker changed")
		}
		if item.RemoteId == "3" && (!item.Terminal || item.EndedAt != nil) {
			t.Fatal("unresolved marker changed")
		}
	}
}
