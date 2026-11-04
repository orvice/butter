package linear

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

func TestBatchResumesFailedPageWithoutReplayingCompletedPage(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	scope := ws.Scope{ID: "team-a"}
	firstCalls, secondCalls, verifyCalls := 0, 0, 0
	makeProvider := func() *Provider {
		return fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
			if req.Query == teamQuery {
				verifyCalls++
				if verifyCalls == 1 {
					return 403, nil, map[string]any{"message": "read-key private"}
				}
				return okay(map[string]any{"team": fixtureTeam()})
			}
			if req.Query != issuesQuery {
				t.Fatal("unexpected batch query")
			}
			if req.Variables["after"] == nil {
				firstCalls++
				return okay(teamPage([]any{fixtureIssue("issue-one", "started", now)}, "next-page"))
			}
			if req.Variables["after"] != "next-page" {
				t.Fatal("lost persisted cursor")
			}
			secondCalls++
			if secondCalls == 1 {
				return 500, nil, map[string]any{"message": "read-key private"}
			}
			return okay(teamPage([]any{fixtureIssue("issue-two", "started", now)}, ""))
		})
	}
	p := makeProvider()
	first, err := p.ScanBatch(context.Background(), "read-key", scope, now, "")
	if err != nil || first.Complete || first.Next == "" || len(first.Items) != 1 {
		t.Fatalf("first=%+v %v", first, err)
	}
	failed, err := p.ScanBatch(context.Background(), "read-key", scope, now, first.Next)
	if err == nil || failed.Next != "" || failed.Coverage != "" || len(failed.Items) > 0 {
		t.Fatal("failed page supplied publishable data")
	}
	p = makeProvider()
	second, err := p.ScanBatch(context.Background(), "read-key", scope, now, first.Next)
	if err != nil || second.Complete || len(second.Items) != 1 || second.Items[0].RemoteId != "issue-two" {
		t.Fatalf("second=%+v %v", second, err)
	}
	denied, err := p.ScanBatch(context.Background(), "read-key", scope, now, second.Next)
	if err == nil || denied.Complete || ws.Classify(err).Category != "ACCESS_DENIED" {
		t.Fatalf("denied=%+v %v", denied, err)
	}
	complete, err := makeProvider().ScanBatch(context.Background(), "read-key", scope, now, second.Next)
	if err != nil || !complete.Complete || complete.Coverage != "COMPLETE" || complete.Next != "" {
		t.Fatalf("complete=%+v %v", complete, err)
	}
	if firstCalls != 1 || secondCalls != 2 || verifyCalls != 2 {
		t.Fatalf("replayed complete page: first=%d second=%d verify=%d", firstCalls, secondCalls, verifyCalls)
	}
	for _, cursor := range []string{first.Next, second.Next} {
		if len(cursor) > maxBatchCursor || strings.Contains(cursor, "read-key") || strings.Contains(cursor, "Markdown") {
			t.Fatal("cursor stores credential or content")
		}
	}
}

func TestBatchCursorRejectsWrongScopeStartAndMalformedState(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	scope := ws.Scope{ID: "team-a"}
	valid := batchCursor{Profile: "linear-full-v1", Scope: scope.ID, Started: now, Name: "ENG · Engineering", Phase: "issues", Page: 2, After: "page-two"}
	encode := func(c batchCursor) string { data, _ := json.Marshal(c); return string(data) }
	wrongScope := valid
	wrongScope.Scope = "team-b"
	wrongStart := valid
	wrongStart.Started = now.Add(time.Second)
	wrongPhase := valid
	wrongPhase.Phase = "mutation"
	wrongPage := valid
	wrongPage.Page = maxPages + 1
	wrongAfter := valid
	wrongAfter.After = ""
	wrongHistory := valid
	wrongHistory.Cursors = []byte{1}
	p := fixtureClient(t, func(gqlRequest) (int, http.Header, any) { t.Fatal("invalid cursor made request"); return 0, nil, nil })
	for _, raw := range []string{encode(wrongScope), encode(wrongStart), encode(wrongPhase), encode(wrongPage), encode(wrongAfter), encode(wrongHistory), encode(valid) + "{}", strings.Repeat("x", maxBatchCursor+1), `{"unknown":true}`} {
		if _, err := p.ScanBatch(context.Background(), "read-key", scope, now, raw); !errors.Is(err, ws.ErrInvalid) {
			t.Fatalf("invalid cursor accepted: %v", err)
		}
	}
}

func TestBatchRepeatedCursorAndPartialGraphQLErrorFailClosed(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	scope := ws.Scope{ID: "team-a"}
	for _, mode := range []string{"repeat", "partial-error", "empty-next"} {
		p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
			if req.Variables["after"] == nil {
				return okay(teamPage([]any{fixtureIssue("one", "started", now)}, "same-next"))
			}
			if mode == "partial-error" {
				return 200, nil, map[string]any{"data": teamPage([]any{fixtureIssue("two", "started", now)}, ""), "errors": []any{map[string]any{"message": "private read-key"}}}
			}
			if mode == "empty-next" {
				return okay(teamPage([]any{}, "different-next"))
			}
			return okay(teamPage([]any{fixtureIssue("two", "started", now)}, "same-next"))
		})
		first, err := p.ScanBatch(context.Background(), "read-key", scope, now, "")
		if err != nil {
			t.Fatal(err)
		}
		second, err := p.ScanBatch(context.Background(), "read-key", scope, now, first.Next)
		if err == nil || second.Complete || second.Next != "" || len(second.Items) > 0 {
			t.Fatalf("%s returned successful page: %+v %v", mode, second, err)
		}
	}
}

func TestBatchPreservesUnresolvedMarkersAndExcludesMovedItems(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	scope := ws.Scope{ID: "team-a"}
	unresolved := fixtureIssue("missing-time", "completed", now)
	moved := fixtureIssue("moved", "started", now)
	moved["team"] = map[string]any{"id": "team-b"}
	p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
		if req.Variables["after"] == nil {
			return okay(teamPage([]any{unresolved, moved}, "next-page"))
		}
		return okay(teamPage([]any{unresolved, fixtureIssue("current", "started", now)}, ""))
	})
	first, err := p.ScanBatch(context.Background(), "read-key", scope, now, "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.ScanBatch(context.Background(), "read-key", scope, now, first.Next)
	if err != nil {
		t.Fatal(err)
	}
	if first.Unresolved != 1 || first.Coverage != "PARTIAL" || len(first.Items) != 1 || first.Items[0].RemoteId != "missing-time" || second.Unresolved != 1 || len(second.Items) != 2 {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
}

func TestBatchPreservesTerminalRowsAndMillisecondStartBinding(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 0, 0, 123456789, time.UTC)
	scope := ws.Scope{ID: "team-a"}
	expired := fixtureIssue("expired", "completed", now)
	expired["completedAt"] = now.Add(-365 * 24 * time.Hour)
	unknown := fixtureIssue("unknown", "completed", now)
	p := fixtureClient(t, func(req gqlRequest) (int, http.Header, any) {
		if req.Query == teamQuery {
			return okay(map[string]any{"team": fixtureTeam()})
		}
		return okay(teamPage([]any{expired, unknown}, ""))
	})
	page, err := p.ScanBatch(context.Background(), "read-key", scope, now, "")
	if err != nil || len(page.Items) != 2 || page.Unresolved != 1 {
		t.Fatalf("terminal rows filtered before global choice: %+v %v", page, err)
	}
	for _, item := range page.Items {
		if item.RemoteId == "expired" && (!item.Terminal || item.EndedAt == nil || ws.Retained(item, now)) {
			t.Fatal("expired marker changed")
		}
		if item.RemoteId == "unknown" && (!item.Terminal || item.EndedAt != nil) {
			t.Fatal("unknown marker changed")
		}
	}
	final, err := p.ScanBatch(context.Background(), "read-key", scope, now.Truncate(time.Millisecond), page.Next)
	if err != nil || !final.Complete {
		t.Fatalf("BSON Date round trip broke resume: %+v %v", final, err)
	}
}
