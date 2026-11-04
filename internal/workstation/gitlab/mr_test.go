package gitlab

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// A later MR page is a separate traversal: a successful Issues scan and first
// MR page cannot authorize a truncated snapshot after that page fails.
func TestMergeRequestPaginationAndFailureAreAtomic(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			mrPages := 0
			p := newProvider(t, func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodGet || r.Body != nil {
					t.Fatal("write request")
				}
				if strings.HasSuffix(r.URL.Path, "/projects/9007199254740993") {
					return response(200, projectJSON, nil), nil
				}
				if strings.HasSuffix(r.URL.Path, "/issues") {
					return response(200, "[]", nil), nil
				}
				if !strings.HasSuffix(r.URL.Path, "/merge_requests") || r.URL.Query().Get("scope") != "all" || r.URL.Query().Get("state") != "all" {
					t.Fatal("unapproved MR endpoint")
				}
				mrPages++
				if r.URL.Query().Get("page") == "1" {
					return response(200, "["+mrJSON+"]", http.Header{"X-Next-Page": {"2"}}), nil
				}
				if fail {
					return response(429, `{"message":"private upstream key"}`, http.Header{"Retry-After": {"120"}}), nil
				}
				open := strings.ReplaceAll(mrJSON, `"iid":2`, `"iid":3`)
				open = strings.Replace(open, `"state":"merged"`, `"state":"opened"`, 1)
				open = strings.Replace(open, `"draft":true`, `"draft":false`, 1)
				return response(200, "["+open+"]", http.Header{"X-Next-Page": {""}}), nil
			})
			result, err := scan(t, p)
			if mrPages != 2 {
				t.Fatal("MR continuation skipped", mrPages)
			}
			if fail {
				if err == nil || len(result.Items) != 0 || strings.Contains(err.Error(), "private") {
					t.Fatal("failed MR page published/leaked", result, err)
				}
				return
			}
			if err != nil || result.Coverage != "COMPLETE" || len(result.Items) != 2 {
				t.Fatal(result, err)
			}
			for _, item := range result.Items {
				if item.Kind != "MERGE_REQUEST" {
					t.Fatal(item)
				}
				if item.RemoteId == "2" && (item.Status != "MERGED" || !item.Draft || item.EndedAt == nil || !item.EndedAt.AsTime().Equal(time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC))) {
					t.Fatal("merged/draft/time", item)
				}
				if item.RemoteId == "3" && (item.Status != "OPEN" || item.Terminal || item.EndedAt != nil || item.Draft) {
					t.Fatal("open MR stale terminal fields", item)
				}
			}
		})
	}
}
