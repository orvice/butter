package opencode

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestBridgeBuildAgent(t *testing.T) {
	ra := &agentsv1.RemoteAgent{Url: "http://localhost:4096"}
	b, err := NewBridge(ra)
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	ag, err := b.BuildAgent("opencode-test", "An OpenCode test agent")
	if err != nil {
		t.Fatalf("BuildAgent: %v", err)
	}
	if ag.Name() != "opencode-test" {
		t.Fatalf("name: got %q", ag.Name())
	}
	if ag.Description() != "An OpenCode test agent" {
		t.Fatalf("description: got %q", ag.Description())
	}
}

func TestNewBridgeDefaultUsername(t *testing.T) {
	b, err := NewBridge(&agentsv1.RemoteAgent{
		Url:      "http://x",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	if b.client.BaseURL() != "http://x" {
		t.Fatalf("expected base URL %q, got %q", "http://x", b.client.BaseURL())
	}
}

func TestExtractText(t *testing.T) {
	if got := extractText(nil); got != "" {
		t.Fatalf("nil: got %q", got)
	}
	c := &genai.Content{Parts: []*genai.Part{{Text: "hello"}, {Text: "world"}}}
	if got := extractText(c); got != "hello\nworld" {
		t.Fatalf("got %q", got)
	}
}

func TestParseModelRef(t *testing.T) {
	ref := parseModelRef("anthropic/claude-3-5-sonnet")
	if ref.ProviderID != "anthropic" || ref.ModelID != "claude-3-5-sonnet" {
		t.Fatalf("got provider=%q model=%q", ref.ProviderID, ref.ModelID)
	}
	ref = parseModelRef("gpt-4")
	if ref.ProviderID != "" || ref.ModelID != "gpt-4" {
		t.Fatalf("got provider=%q model=%q", ref.ProviderID, ref.ModelID)
	}
}

// fakeServer captures requests against a stub opencode server and returns
// configurable responses.
type fakeServer struct {
	mu         sync.Mutex
	t          *testing.T
	sessionID  string
	authHeader string
	messageReq map[string]any
	abortHit   atomic.Bool
	srv        *httptest.Server
}

func newFakeServer(t *testing.T) *fakeServer {
	fs := &fakeServer{t: t, sessionID: "sess-123"}
	mux := http.NewServeMux()
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		fs.mu.Lock()
		fs.authHeader = r.Header.Get("Authorization")
		fs.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": fs.sessionID})
	})
	mux.HandleFunc("/session/"+fs.sessionID+"/message", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var msg map[string]any
		_ = json.Unmarshal(body, &msg)
		fs.mu.Lock()
		fs.messageReq = msg
		fs.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"info": map[string]any{
				"id":        "msg-1",
				"sessionID": fs.sessionID,
				"role":      "assistant",
				"time":      map[string]any{"created": 0},
				"tokens":    map[string]any{},
				"path":      map[string]any{},
			},
			"parts": []map[string]any{
				{"id": "p1", "sessionID": fs.sessionID, "messageID": "msg-1", "type": "text", "text": "hello back"},
				{"id": "p2", "sessionID": fs.sessionID, "messageID": "msg-1", "type": "tool"},
				{"id": "p3", "sessionID": fs.sessionID, "messageID": "msg-1", "type": "text", "text": "second line"},
			},
		})
	})
	mux.HandleFunc("/session/"+fs.sessionID+"/abort", func(w http.ResponseWriter, _ *http.Request) {
		fs.abortHit.Store(true)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(true)
	})
	mux.HandleFunc("/global/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"healthy": true, "version": "1.0"})
	})
	fs.srv = httptest.NewServer(mux)
	t.Cleanup(fs.srv.Close)
	return fs
}

func TestBridgeCreateAndSend(t *testing.T) {
	fs := newFakeServer(t)
	b, err := NewBridge(&agentsv1.RemoteAgent{
		Url:           fs.srv.URL,
		OpencodeAgent: "review",
		OpencodeModel: "anthropic/claude-3-5-sonnet",
	})
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}

	sessID, err := b.createSession(t.Context())
	if err != nil {
		t.Fatalf("createSession: %v", err)
	}
	if sessID != fs.sessionID {
		t.Fatalf("session id: got %q", sessID)
	}

	out, err := b.sendMessage(t.Context(), sessID, "hi there")
	if err != nil {
		t.Fatalf("sendMessage: %v", err)
	}
	if out != "hello back\nsecond line" {
		t.Fatalf("output: got %q", out)
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.messageReq["agent"] != "review" {
		t.Fatalf("agent: got %v", fs.messageReq["agent"])
	}
	model, ok := fs.messageReq["model"].(map[string]any)
	if !ok {
		t.Fatalf("model: expected object, got %T %v", fs.messageReq["model"], fs.messageReq["model"])
	}
	if model["providerID"] != "anthropic" || model["modelID"] != "claude-3-5-sonnet" {
		t.Fatalf("model: got %v", model)
	}
	parts, _ := fs.messageReq["parts"].([]any)
	if len(parts) != 1 {
		t.Fatalf("parts len: got %d", len(parts))
	}
}

func TestBridgeBasicAuth(t *testing.T) {
	fs := newFakeServer(t)
	b, err := NewBridge(&agentsv1.RemoteAgent{
		Url:      fs.srv.URL,
		Username: "alice",
		Password: "wonderland",
	})
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	if _, err := b.createSession(t.Context()); err != nil {
		t.Fatalf("createSession: %v", err)
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if !strings.HasPrefix(fs.authHeader, "Basic ") {
		t.Fatalf("expected Basic auth header, got %q", fs.authHeader)
	}
}

func TestBridgeHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	b, err := NewBridge(&agentsv1.RemoteAgent{Url: srv.URL})
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	_, createErr := b.createSession(t.Context())
	if createErr == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(createErr.Error(), "500") {
		t.Fatalf("expected 500 error, got %v", createErr)
	}
}

func TestBridgeAbortOnCancel(t *testing.T) {
	fs := newFakeServer(t)
	b, err := NewBridge(&agentsv1.RemoteAgent{Url: fs.srv.URL})
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	if err := b.abortSession(t.Context(), fs.sessionID); err != nil {
		t.Fatalf("abortSession: %v", err)
	}
	if !fs.abortHit.Load() {
		t.Fatal("expected abort endpoint to be hit")
	}
}
