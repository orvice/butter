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
	b := NewBridge(ra)
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
	b := NewBridge(&agentsv1.RemoteAgent{
		Url:      "http://x",
		Password: "secret",
	})
	if b.username != defaultUsername {
		t.Fatalf("expected default username %q, got %q", defaultUsername, b.username)
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
			"parts": []map[string]any{
				{"type": "text", "text": "hello back"},
				{"type": "tool", "text": "ignored"},
				{"type": "text", "text": "second line"},
			},
		})
	})
	mux.HandleFunc("/session/"+fs.sessionID+"/abort", func(w http.ResponseWriter, _ *http.Request) {
		fs.abortHit.Store(true)
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/global/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	fs.srv = httptest.NewServer(mux)
	t.Cleanup(fs.srv.Close)
	return fs
}

func TestBridgeCreateAndSend(t *testing.T) {
	fs := newFakeServer(t)
	b := NewBridge(&agentsv1.RemoteAgent{
		Url:           fs.srv.URL,
		OpencodeAgent: "review",
		OpencodeModel: "anthropic/claude-3-5-sonnet",
	})

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
	if fs.messageReq["model"] != "anthropic/claude-3-5-sonnet" {
		t.Fatalf("model: got %v", fs.messageReq["model"])
	}
	parts, _ := fs.messageReq["parts"].([]any)
	if len(parts) != 1 {
		t.Fatalf("parts len: got %d", len(parts))
	}
}

func TestBridgeBasicAuth(t *testing.T) {
	fs := newFakeServer(t)
	b := NewBridge(&agentsv1.RemoteAgent{
		Url:      fs.srv.URL,
		Username: "alice",
		Password: "wonderland",
	})
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
	b := NewBridge(&agentsv1.RemoteAgent{Url: srv.URL})
	_, err := b.createSession(t.Context())
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Fatalf("expected 500 error, got %v", err)
	}
}

func TestBridgeAbortOnCancel(t *testing.T) {
	// Verifies abortSession hits the right endpoint.
	fs := newFakeServer(t)
	b := NewBridge(&agentsv1.RemoteAgent{Url: fs.srv.URL})
	if err := b.abortSession(t.Context(), fs.sessionID); err != nil {
		t.Fatalf("abortSession: %v", err)
	}
	if !fs.abortHit.Load() {
		t.Fatal("expected abort endpoint to be hit")
	}
}
