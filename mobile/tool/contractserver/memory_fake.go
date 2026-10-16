package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"

	"go.orx.me/apps/butter/internal/mem0"
)

// contractMem0Key is the API key every memory contract test uses.
const contractMem0Key = "contract-mem0-key"

// mem0Store is a deterministic in-process mem0 OSS double.
type mem0Store struct {
	srv *httptest.Server

	mu       sync.Mutex
	validKey string
	memories map[string]mem0.Memory
	searches []map[string]any
	deleted  []string
	closed   bool
}

func newMem0Store() *mem0Store {
	f := &mem0Store{
		validKey: contractMem0Key,
		memories: map[string]mem0.Memory{},
	}
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	return f
}

func (f *mem0Store) close() { f.srv.Close() }

func (f *mem0Store) baseURL() string { return f.srv.URL }

func (f *mem0Store) seed(m ...mem0.Memory) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, item := range m {
		f.memories[item.ID] = item
	}
}

func (f *mem0Store) authOK(r *http.Request) bool {
	f.mu.Lock()
	key := f.validKey
	closed := f.closed
	f.mu.Unlock()
	if closed {
		return false
	}
	if key == "" {
		return true
	}
	return r.Header.Get("X-API-Key") == key
}

func (f *mem0Store) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	closed := f.closed
	f.mu.Unlock()
	if closed {
		http.Error(w, `{"detail":"closed"}`, http.StatusServiceUnavailable)
		return
	}
	if !f.authOK(r) {
		http.Error(w, `{"detail":"Invalid API key"}`, http.StatusUnauthorized)
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	matches := func(m mem0.Memory, userID, agentID string) bool {
		return (userID != "" && m.UserID == userID) || (agentID != "" && m.AgentID == agentID)
	}
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/memories":
		topK, _ := strconv.Atoi(r.URL.Query().Get("top_k"))
		var out []mem0.Memory
		for _, m := range f.memories {
			if matches(m, r.URL.Query().Get("user_id"), r.URL.Query().Get("agent_id")) {
				out = append(out, m)
			}
		}
		if topK > 0 && len(out) > topK {
			out = out[:topK]
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": out})
	case r.Method == http.MethodPost && r.URL.Path == "/search":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.searches = append(f.searches, body)
		filters, _ := body["filters"].(map[string]any)
		userID, _ := filters["user_id"].(string)
		agentID, _ := filters["agent_id"].(string)
		query, _ := body["query"].(string)
		var out []mem0.Memory
		for _, m := range f.memories {
			if matches(m, userID, agentID) && strings.Contains(m.Memory, query) {
				m.Score = 0.7
				out = append(out, m)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": out})
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/memories/"):
		id := strings.TrimPrefix(r.URL.Path, "/memories/")
		m, ok := f.memories[id]
		if !ok {
			_, _ = w.Write([]byte("null"))
			return
		}
		_ = json.NewEncoder(w).Encode(m)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/memories/"):
		id := strings.TrimPrefix(r.URL.Path, "/memories/")
		if _, ok := f.memories[id]; !ok {
			http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
			return
		}
		delete(f.memories, id)
		f.deleted = append(f.deleted, id)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	default:
		http.NotFound(w, r)
	}
}

func seedContractMemory(id, text, userID, agentID, created string) mem0.Memory {
	return mem0.Memory{
		ID: id, Memory: text, UserID: userID, AgentID: agentID, CreatedAt: created,
		Metadata: map[string]any{
			"butter_agent_id":   "mem-agent",
			"butter_channel":    "mobile-contract",
			"butter_principal":  "user:contract",
			"butter_session_id": "s-contract",
		},
	}
}
