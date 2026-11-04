package gitlab

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

func batchNow() time.Time { return time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC) }

func TestScanBatchResumesFailedPageAfterProviderRestart(t *testing.T) {
	var issuePages []string
	failed := false
	transport := transportFunc(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/projects/9007199254740993") {
			return response(200, projectJSON, nil), nil
		}
		if strings.HasSuffix(r.URL.Path, "/issues") {
			page := r.URL.Query().Get("page")
			issuePages = append(issuePages, page)
			if page == "1" {
				return response(200, "["+issueJSON+"]", http.Header{"X-Next-Page": {"2"}}), nil
			}
			if !failed {
				failed = true
				return response(500, "test-secret private failure", nil), nil
			}
			second := strings.Replace(issueJSON, `"iid":9007199254740994`, `"iid":3`, 1)
			return response(200, "["+second+"]", nil), nil
		}
		return response(200, "["+mrJSON+"]", nil), nil
	})
	scope := ws.Scope{ID: "9007199254740993"}
	first, err := newProvider(t, transport).ScanBatch(context.Background(), "test-secret", scope, batchNow(), "")
	if err != nil || first.Complete || first.Next == "" || len(first.Items) != 1 {
		t.Fatalf("first batch must not publish: %+v %v", first, err)
	}
	metadata, err := newProvider(t, transport).decodeCursor(first.Next, scope, batchNow())
	if err != nil || metadata.Page != 2 || metadata.Phase != 0 || metadata.Scope != scope.ID {
		t.Fatalf("checkpoint metadata: %+v %v", metadata, err)
	}
	if strings.Contains(string(mustDecodeCursor(t, first.Next)), "test-secret") || strings.Contains(string(mustDecodeCursor(t, first.Next)), "Old open issue") {
		t.Fatal("cursor contains credential or item body")
	}
	broken, err := newProvider(t, transport).ScanBatch(context.Background(), "test-secret", scope, batchNow(), first.Next)
	if err == nil || len(broken.Items) != 0 || broken.Next != "" || broken.Complete {
		t.Fatalf("failed page returned publishable batch: %+v %v", broken, err)
	}
	// Recreate the adapter as a new process would. Only the persisted cursor and
	// original scan timestamp are needed; page one must not be requested again.
	restarted := newProvider(t, transport)
	second, err := restarted.ScanBatch(context.Background(), "test-secret", scope, batchNow(), first.Next)
	if err != nil || second.Complete || second.Next == "" || len(second.Items) != 1 || second.Items[0].RemoteId != "3" {
		t.Fatalf("resumed batch: %+v %v", second, err)
	}
	last, err := restarted.ScanBatch(context.Background(), "test-secret", scope, batchNow(), second.Next)
	if err != nil || !last.Complete || last.Next != "" || len(last.Items) != 1 || last.Items[0].Kind != "MERGE_REQUEST" {
		t.Fatalf("final batch: %+v %v", last, err)
	}
	if strings.Join(issuePages, ",") != "1,2,2" {
		t.Fatalf("restart lost checkpoint: %v", issuePages)
	}
}

func mustDecodeCursor(t *testing.T, cursor string) []byte {
	t.Helper()
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBatchCursorCannotCrossScopeInstanceOrScanTimestamp(t *testing.T) {
	called := false
	p := newProvider(t, func(*http.Request) (*http.Response, error) { called = true; return nil, nil })
	cursor := encodeCursor(batchCursor{Version: 1, API: p.api.String(), Scope: "9007199254740993", Path: "group/subgroup/repo", Started: batchNow().UnixNano(), Page: 2})
	for _, tc := range []struct {
		name   string
		scope  ws.Scope
		now    time.Time
		cursor string
	}{
		{"wrong scope", ws.Scope{ID: "2"}, batchNow(), cursor},
		{"wrong time", ws.Scope{ID: "9007199254740993"}, batchNow().Add(time.Second), cursor},
		{"malformed", ws.Scope{ID: "9007199254740993"}, batchNow(), "not-base64"},
		{"wrong instance", ws.Scope{ID: "9007199254740993"}, batchNow(), encodeCursor(batchCursor{Version: 1, API: "https://other.gitlab.example/api/v4", Scope: "9007199254740993", Path: "group/repo", Started: batchNow().UnixNano(), Page: 2})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			batch, err := p.ScanBatch(context.Background(), "test-secret", tc.scope, tc.now, tc.cursor)
			if !errors.Is(err, ws.ErrInvalid) || batch.Complete || called {
				t.Fatalf("invalid checkpoint reached API: %+v %v", batch, err)
			}
		})
	}
}

func TestScanBatchRepeatPageAndFinalRevocationFailClosed(t *testing.T) {
	for _, scenario := range []string{"repeat", "revoked", "moved"} {
		t.Run(scenario, func(t *testing.T) {
			projectReads := 0
			p := newProvider(t, func(r *http.Request) (*http.Response, error) {
				if strings.HasSuffix(r.URL.Path, "/projects/9007199254740993") {
					projectReads++
					if projectReads == 3 && scenario == "revoked" {
						return response(404, "test-secret", nil), nil
					}
					if projectReads == 3 && scenario == "moved" {
						return response(200, strings.Replace(projectJSON, "group/subgroup/repo", "moved/repo", 1), nil), nil
					}
					return response(200, projectJSON, nil), nil
				}
				if strings.HasSuffix(r.URL.Path, "/issues") {
					if scenario == "repeat" {
						return response(200, "["+issueJSON+"]", http.Header{"X-Next-Page": {"2"}}), nil
					}
					return response(200, "[]", nil), nil
				}
				return response(200, "["+mrJSON+"]", nil), nil
			})
			scope := ws.Scope{ID: "9007199254740993"}
			first, err := p.ScanBatch(context.Background(), "test-secret", scope, batchNow(), "")
			if err != nil || first.Complete {
				t.Fatal(first, err)
			}
			last, err := p.ScanBatch(context.Background(), "test-secret", scope, batchNow(), first.Next)
			if err == nil || last.Complete || len(last.Items) != 0 || last.Next != "" {
				t.Fatalf("failed traversal became complete: %+v %v", last, err)
			}
		})
	}
}

func TestScanBatchReportsUnresolvedWithoutInventingTerminalTime(t *testing.T) {
	p := newProvider(t, func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/projects/9007199254740993") {
			return response(200, projectJSON, nil), nil
		}
		if strings.HasSuffix(r.URL.Path, "/issues") {
			return response(200, "["+strings.Replace(issueJSON, `"state":"opened"`, `"state":"closed"`, 1)+"]", nil), nil
		}
		return response(200, "[]", nil), nil
	})
	batch, err := p.ScanBatch(context.Background(), "test-secret", ws.Scope{ID: "9007199254740993"}, batchNow(), "")
	if err != nil || batch.Coverage != "PARTIAL" || batch.Unresolved != 1 || batch.Complete || len(batch.Items) != 1 || batch.Items[0].EndedAt != nil || batch.Next == "" {
		t.Fatal(batch, err)
	}
}

func TestScanBatchRetainsOverlappingTerminalMarkersAndExpiredRows(t *testing.T) {
	p := newProvider(t, func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/projects/9007199254740993") {
			return response(200, projectJSON, nil), nil
		}
		if r.URL.Query().Get("page") == "1" {
			return response(200, "["+issueJSON+"]", http.Header{"X-Next-Page": {"2"}}), nil
		}
		marker := strings.Replace(issueJSON, `"state":"opened"`, `"state":"closed"`, 1)
		marker = strings.Replace(marker, `"updated_at":"2020-02-01T00:00:00Z"`, `"updated_at":"2026-11-01T00:00:00Z"`, 1)
		expired := strings.Replace(marker, `"iid":9007199254740994`, `"iid":3`, 1)
		expired = strings.Replace(expired, `"state":"closed"`, `"state":"closed","closed_at":"2020-02-01T00:00:00Z"`, 1)
		return response(200, "["+marker+","+expired+"]", nil), nil
	})
	scope := ws.Scope{ID: "9007199254740993"}
	first, err := p.ScanBatch(context.Background(), "test-secret", scope, batchNow(), "")
	if err != nil {
		t.Fatal(err)
	}
	next, err := p.ScanBatch(context.Background(), "test-secret", scope, batchNow(), first.Next)
	if err != nil || len(next.Items) != 2 || next.Unresolved != 1 || next.Coverage != "PARTIAL" {
		t.Fatalf("terminal rows lost before latest-source selection: %+v %v", next, err)
	}
	if next.Items[0].RemoteId != first.Items[0].RemoteId || !next.Items[0].Terminal || next.Items[0].EndedAt != nil {
		t.Fatalf("newest unresolved marker omitted: %v", next.Items[0])
	}
	if !next.Items[1].Terminal || next.Items[1].EndedAt == nil || ws.Retained(next.Items[1], batchNow()) {
		t.Fatalf("expired terminal marker omitted or timestamp changed: %v", next.Items[1])
	}
}
