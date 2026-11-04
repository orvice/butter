package plane

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

func TestBatchResumeAfterFailureWithFreshProvider(t *testing.T) {
	started := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	selected := ws.Scope{ID: "project-1", Name: "outdated display name"}
	failSecond := true
	firstReads, secondReads := 0, 0
	handler := transportFunc(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/project-1/") {
			return response(200, projectJSON), nil
		}
		if r.URL.Query().Get("order_by") != "id" || r.URL.Query().Get("paginate") != "cursor" || r.URL.Query().Get("archived") != "true" || r.URL.Query().Get("expand") != "state,assignees,labels" {
			t.Fatal("batch lost fixed traversal parameters")
		}
		if r.URL.Query().Get("cursor") == "" {
			firstReads++
			return response(200, envelopeJSON(itemJSON, "next-second")), nil
		}
		if r.URL.Query().Get("cursor") != "next-second" {
			t.Fatal("resume changed opaque cursor")
		}
		secondReads++
		if failSecond {
			return response(500, "upstream-private-body"), nil
		}
		second := strings.Replace(itemJSON, `9007199254740993`, `9007199254740994`, 1)
		return response(200, envelopeJSON(second, "")), nil
	})
	p := newProvider(t, handler)
	first, err := p.ScanBatch(context.Background(), "test-secret", selected, started, "")
	if err != nil || first.Complete || first.Next == "" || len(first.Items) != 1 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	if _, err := p.ScanBatch(context.Background(), "test-secret", selected, started, first.Next); err == nil {
		t.Fatal("failed second page returned a batch")
	}
	// A fresh adapter has no shared in-memory traversal state. Only the cursor
	// and immutable scan-start time are needed to continue the same traversal.
	failSecond = false
	restarted := newProvider(t, handler)
	last, err := restarted.ScanBatch(context.Background(), "test-secret", selected, started, first.Next)
	if err != nil || !last.Complete || last.Next != "" || len(last.Items) != 1 || last.Items[0].RemoteId != "9007199254740994" || firstReads != 1 || secondReads != 2 {
		t.Fatalf("last=%+v err=%v reads=%d/%d", last, err, firstReads, secondReads)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(first.Next)
	if err != nil || strings.Contains(string(decoded), "Details") || strings.Contains(string(decoded), "test-secret") || strings.Contains(string(decoded), "Shared project") {
		t.Fatal("cursor contains source body, credential or source name")
	}
}

func TestBatchCursorBindsScopeOriginNameAndStart(t *testing.T) {
	started := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	name := projectJSON
	itemReads := 0
	handler := transportFunc(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/work-items/") {
			itemReads++
			return response(200, envelopeJSON(itemJSON, "next-second")), nil
		}
		if strings.HasSuffix(r.URL.Path, "/other-project/") {
			return response(200, `{"id":"other-project","name":"Shared project"}`), nil
		}
		return response(200, name), nil
	})
	p := newProvider(t, handler)
	first, err := p.ScanBatch(context.Background(), "test-secret", ws.Scope{ID: "project-1"}, started, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		scope  ws.Scope
		start  time.Time
		cursor string
	}{
		{"wrong_scope", ws.Scope{ID: "other-project"}, started, first.Next},
		{"different_scan_start", ws.Scope{ID: "project-1"}, started.Add(time.Millisecond), first.Next},
		{"malformed", ws.Scope{ID: "project-1"}, started, "%%%"},
		{"missing_start", ws.Scope{ID: "project-1"}, time.Time{}, first.Next},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := itemReads
			if _, err := p.ScanBatch(context.Background(), "test-secret", test.scope, test.start, test.cursor); !errors.Is(err, ws.ErrInvalid) {
				t.Fatalf("invalid cursor error=%v", err)
			}
			if itemReads != before {
				t.Fatal("invalid checkpoint issued item request")
			}
		})
	}
	name = strings.Replace(projectJSON, "Shared project", "Renamed project", 1)
	if _, err := p.ScanBatch(context.Background(), "test-secret", ws.Scope{ID: "project-1"}, started, first.Next); !errors.Is(err, ws.ErrInvalid) {
		t.Fatalf("changed project metadata accepted err=%v", err)
	}
	decoded, _ := base64.RawURLEncoding.DecodeString(first.Next)
	var cursor batchCursor
	_ = json.Unmarshal(decoded, &cursor)
	cursor.Instance = digest("https://different.example")
	changed, _ := json.Marshal(cursor)
	name = projectJSON
	if _, err := p.ScanBatch(context.Background(), "test-secret", ws.Scope{ID: "project-1"}, started, base64.RawURLEncoding.EncodeToString(changed)); !errors.Is(err, ws.ErrInvalid) {
		t.Fatalf("different instance cursor accepted err=%v", err)
	}
}

func TestBatchRejectsRepeatedCursorRowsAndLateAccessFailure(t *testing.T) {
	started := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	for _, scenario := range []string{"repeated_cursor", "cursor_cycle", "repeated_rows", "unsorted_ids", "final_revocation", "final_rename"} {
		t.Run(scenario, func(t *testing.T) {
			resolved := 0
			p := newProvider(t, func(r *http.Request) (*http.Response, error) {
				if strings.HasSuffix(r.URL.Path, "/project-1/") {
					resolved++
					if resolved >= 3 && scenario == "final_revocation" {
						return response(403, "private"), nil
					}
					if resolved >= 3 && scenario == "final_rename" {
						return response(200, strings.Replace(projectJSON, "Shared project", "Renamed", 1)), nil
					}
					return response(200, projectJSON), nil
				}
				cursor := r.URL.Query().Get("cursor")
				if cursor == "" {
					return response(200, envelopeJSON(itemJSON, "page-2")), nil
				}
				second := strings.Replace(itemJSON, `9007199254740993`, `9007199254740994`, 1)
				switch scenario {
				case "repeated_cursor":
					return response(200, envelopeJSON(second, "page-2")), nil
				case "cursor_cycle":
					if cursor == "page-2" {
						return response(200, envelopeJSON(second, "page-3")), nil
					}
					return response(200, envelopeJSON(strings.Replace(itemJSON, `9007199254740993`, `9007199254740995`, 1), "page-2")), nil
				case "repeated_rows":
					return response(200, envelopeJSON(itemJSON, "")), nil
				case "unsorted_ids":
					return response(200, envelopeJSON(second+","+itemJSON, "")), nil
				default:
					return response(200, envelopeJSON(second, "")), nil
				}
			})
			batch, err := p.ScanBatch(context.Background(), "test-secret", ws.Scope{ID: "project-1"}, started, "")
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "cursor_cycle" {
				batch, err = p.ScanBatch(context.Background(), "test-secret", ws.Scope{ID: "project-1"}, started, batch.Next)
				if err != nil {
					t.Fatal(err)
				}
			}
			last, err := p.ScanBatch(context.Background(), "test-secret", ws.Scope{ID: "project-1"}, started, batch.Next)
			if err == nil || last.Complete || len(last.Items) > 0 {
				t.Fatalf("invalid final batch=%+v err=%v", last, err)
			}
		})
	}
}

func TestBatchFrozenRetentionAndPartialTerminalCoverage(t *testing.T) {
	started := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	boundary := started.Add(-90 * 24 * time.Hour).Format(time.RFC3339)
	completed := strings.Replace(strings.Replace(itemJSON, `"group":"started"`, `"group":"completed"`, 1), `"completed_at":null`, `"completed_at":"`+boundary+`"`, 1)
	canceled := strings.Replace(strings.Replace(itemJSON, `9007199254740993`, `9007199254740994`, 1), `"group":"started"`, `"group":"cancelled"`, 1)
	p := newProvider(t, func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/project-1/") {
			return response(200, projectJSON), nil
		}
		return response(200, envelopeJSON(completed+","+canceled, "")), nil
	})
	result, err := p.ScanBatch(context.Background(), "test-secret", ws.Scope{ID: "project-1"}, started, "")
	if err != nil || !result.Complete || result.Coverage != "PARTIAL" || result.Unresolved != 1 || len(result.Items) != 1 || !result.Items[0].EndedAt.AsTime().Equal(started.Add(-90*24*time.Hour)) {
		t.Fatalf("batch=%+v err=%v", result, err)
	}
}
