package mem0

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchSendsAPIKeyAndDecodesResults(t *testing.T) {
	var gotKey string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/search" {
			t.Errorf("request = %s %s, want POST /search", r.Method, r.URL.Path)
		}
		gotKey = r.Header.Get("X-API-Key")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"m1","memory":"likes tea","score":0.8,"created_at":"2026-09-01T00:00:00Z","user_id":"ws:w1"}]}`))
	}))
	defer srv.Close()

	threshold := 0.3
	got, err := New(srv.URL+"/", "m0sk_test", srv.Client()).Search(t.Context(), SearchRequest{
		Query:     "tea",
		Filters:   map[string]any{"user_id": "ws:w1"},
		TopK:      5,
		Threshold: &threshold,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if gotKey != "m0sk_test" {
		t.Fatalf("X-API-Key = %q", gotKey)
	}
	if gotBody["query"] != "tea" || gotBody["top_k"] != float64(5) || gotBody["threshold"] != 0.3 {
		t.Fatalf("body = %v", gotBody)
	}
	if filters, _ := gotBody["filters"].(map[string]any); filters["user_id"] != "ws:w1" {
		t.Fatalf("filters = %v", gotBody["filters"])
	}
	if len(got) != 1 || got[0].ID != "m1" || got[0].Memory != "likes tea" || got[0].Score != 0.8 || got[0].UserID != "ws:w1" {
		t.Fatalf("results = %+v", got)
	}
}

func TestAddSendsMessagesAndScope(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/memories" {
			t.Errorf("request = %s %s, want POST /memories", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"m9","memory":"prefers pnpm","event":"ADD"}]}`))
	}))
	defer srv.Close()

	infer := true
	got, err := New(srv.URL, "k", srv.Client()).Add(t.Context(), AddRequest{
		Messages: []Message{{Role: "user", Content: "we use pnpm"}, {Role: "assistant", Content: "noted"}},
		UserID:   "ws:w1",
		Metadata: map[string]any{"session_id": "s1"},
		Infer:    &infer,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if len(got) != 1 || got[0].ID != "m9" || got[0].Event != "ADD" {
		t.Fatalf("results = %+v", got)
	}
	if gotBody["user_id"] != "ws:w1" || gotBody["infer"] != true {
		t.Fatalf("body = %v", gotBody)
	}
	for _, absent := range []string{"agent_id", "run_id"} {
		if _, ok := gotBody[absent]; ok {
			t.Fatalf("body carries %s: %v", absent, gotBody)
		}
	}
	msgs, _ := gotBody["messages"].([]any)
	if len(msgs) != 2 {
		t.Fatalf("messages = %v", gotBody["messages"])
	}
	if meta, _ := gotBody["metadata"].(map[string]any); meta["session_id"] != "s1" {
		t.Fatalf("metadata = %v", gotBody["metadata"])
	}
}

func TestSearchWithoutAPIKeySendsNoHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Header["X-Api-Key"]; ok {
			t.Error("X-API-Key header sent without a key")
		}
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	if _, err := New(srv.URL, "", srv.Client()).Search(t.Context(), SearchRequest{Query: "q", Filters: map[string]any{"user_id": "u"}}); err != nil {
		t.Fatalf("Search: %v", err)
	}
}

func TestSearchReportsAPIErrors(t *testing.T) {
	for _, tc := range []struct {
		status int
		auth   bool
	}{
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, true},
		{http.StatusBadRequest, false},
		{http.StatusBadGateway, false},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"detail":"nope"}`, tc.status)
		}))
		_, err := New(srv.URL, "k", srv.Client()).Search(t.Context(), SearchRequest{Query: "q"})
		srv.Close()

		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status {
			t.Fatalf("status %d: err = %v, want APIError", tc.status, err)
		}
		if IsAuthError(err) != tc.auth {
			t.Fatalf("status %d: IsAuthError = %v, want %v", tc.status, !tc.auth, tc.auth)
		}
	}
}

func TestIsAuthErrorIgnoresTransportErrors(t *testing.T) {
	_, err := New("http://127.0.0.1:1", "k", nil).Search(t.Context(), SearchRequest{Query: "q"})
	if err == nil {
		t.Fatal("Search against a closed port succeeded")
	}
	if IsAuthError(err) {
		t.Fatal("transport error classified as auth error")
	}
}

func TestListGetDelete(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		if r.Header.Get("X-API-Key") != "k" {
			t.Errorf("%s %s without the API key", r.Method, r.URL.Path)
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/memories":
			_, _ = w.Write([]byte(`{"results":[{"id":"m1","memory":"a","user_id":"ws:w1","metadata":{"butter_channel":"web-chat"}}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/memories/m1":
			_, _ = w.Write([]byte(`{"id":"m1","memory":"a","agent_id":"ws:w1:agent:x"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/memories/missing":
			_, _ = w.Write([]byte(`null`))
		case r.Method == http.MethodDelete && r.URL.Path == "/memories/m1":
			_, _ = w.Write([]byte(`{"message":"Memory deleted successfully"}`))
		default:
			http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := New(srv.URL, "k", srv.Client())

	list, err := c.List(t.Context(), ListRequest{UserID: "ws:w1", TopK: 201})
	if err != nil || len(list) != 1 || list[0].Metadata["butter_channel"] != "web-chat" {
		t.Fatalf("List = %+v, %v", list, err)
	}
	if _, err := c.List(t.Context(), ListRequest{AgentID: "ws:w1:agent:x"}); err != nil {
		t.Fatalf("List by agent: %v", err)
	}
	if _, err := c.List(t.Context(), ListRequest{}); err == nil {
		t.Fatal("identity-less List succeeded")
	}

	got, err := c.Get(t.Context(), "m1")
	if err != nil || got == nil || got.AgentID != "ws:w1:agent:x" {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	if got, err := c.Get(t.Context(), "missing"); err != nil || got != nil {
		t.Fatalf("Get missing = %+v, %v; want nil, nil", got, err)
	}
	if err := c.Delete(t.Context(), "m1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	var apiErr *APIError
	if err := c.Delete(t.Context(), "nope"); !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("Delete unknown = %v", err)
	}

	want := []string{
		"GET /memories?top_k=201&user_id=ws%3Aw1",
		"GET /memories?agent_id=ws%3Aw1%3Aagent%3Ax",
		"GET /memories/m1",
		"GET /memories/missing",
		"DELETE /memories/m1",
		"DELETE /memories/nope",
	}
	if len(seen) != len(want) {
		t.Fatalf("requests = %v", seen)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("request %d = %q, want %q", i, seen[i], want[i])
		}
	}
}
