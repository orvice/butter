package openaifake

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Backend is an OpenAI-compatible chat completions endpoint for tests. It
// records complete requests and derived user input by actual model ID.
type Backend struct {
	server   *httptest.Server
	scripted map[string]http.HandlerFunc

	mu                sync.Mutex
	inputsByModelID   map[string][]string
	requestsByModelID map[string][]ChatCompletionRequest
}

type ChatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []ChatCompletionMessage `json:"messages"`
	Decoded  map[string]any          `json:"-"`
}

type ChatCompletionMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

func New(t testing.TB) *Backend {
	t.Helper()
	b := &Backend{
		scripted:          make(map[string]http.HandlerFunc),
		inputsByModelID:   make(map[string][]string),
		requestsByModelID: make(map[string][]ChatCompletionRequest),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/chat/completions", b.handleCompletion)
	b.server = httptest.NewServer(mux)
	t.Cleanup(b.server.Close)
	return b
}

func (b *Backend) URL() string {
	return b.server.URL
}

func (b *Backend) handleCompletion(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &req.Decoded); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lastUser := lastUserInput(req.Messages)
	b.mu.Lock()
	b.inputsByModelID[req.Model] = append(b.inputsByModelID[req.Model], lastUser)
	b.requestsByModelID[req.Model] = append(b.requestsByModelID[req.Model], req)
	handler := b.scripted[req.Model]
	b.mu.Unlock()

	if handler != nil {
		handler(w, r)
		return
	}
	WriteCompletion(w, req.Model, fmt.Sprintf("%s(%s)", req.Model, lastUser))
}

func lastUserInput(messages []ChatCompletionMessage) string {
	lastUser := ""
	for _, message := range messages {
		if message.Role != "user" {
			continue
		}
		var text string
		if json.Unmarshal(message.Content, &text) == nil {
			lastUser = text
			continue
		}
		var parts []struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(message.Content, &parts) == nil {
			var joined strings.Builder
			for _, part := range parts {
				joined.WriteString(part.Text)
			}
			lastUser = joined.String()
		}
	}
	return lastUser
}

// Answer scripts a fixed reply for a model, overriding the echo default.
func (b *Backend) Answer(model, reply string) {
	b.Script(model, func(w http.ResponseWriter, _ *http.Request) {
		WriteCompletion(w, model, reply)
	})
}

// Script installs a handler for a model, replacing the echo default.
func (b *Backend) Script(model string, handler http.HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scripted[model] = handler
}

// RequireConcurrent blocks completions until n requests are in flight.
func (b *Backend) RequireConcurrent(model string, n int32) {
	var inFlight atomic.Int32
	proceed := make(chan struct{})
	var once sync.Once
	b.Script(model, func(w http.ResponseWriter, _ *http.Request) {
		if inFlight.Add(1) >= n {
			once.Do(func() { close(proceed) })
		}
		select {
		case <-proceed:
			WriteCompletion(w, model, "done")
		case <-time.After(3 * time.Second):
			http.Error(w, `{"error": {"message": "items were not processed concurrently"}}`, http.StatusBadRequest)
		}
	})
}

// FailFirstCall makes a model fail once with an HTTP 400, then echo normally.
func (b *Backend) FailFirstCall(model string) {
	failed := false
	b.Script(model, func(w http.ResponseWriter, _ *http.Request) {
		b.mu.Lock()
		first := !failed
		failed = true
		input := ""
		if inputs := b.inputsByModelID[model]; len(inputs) > 0 {
			input = inputs[len(inputs)-1]
		}
		b.mu.Unlock()
		if first {
			http.Error(w, `{"error": {"message": "transient failure"}}`, http.StatusBadRequest)
			return
		}
		WriteCompletion(w, model, fmt.Sprintf("%s(%s)", model, input))
	})
}

func (b *Backend) LastInput(model string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	inputs := b.inputsByModelID[model]
	if len(inputs) == 0 {
		return ""
	}
	return inputs[len(inputs)-1]
}

func (b *Backend) LastRequest(model string) (ChatCompletionRequest, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	requests := b.requestsByModelID[model]
	if len(requests) == 0 {
		return ChatCompletionRequest{}, false
	}
	return requests[len(requests)-1], true
}

func (b *Backend) CallCount(model string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.requestsByModelID[model])
}

func WriteCompletion(w http.ResponseWriter, model, reply string) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"id": "cmpl-test",
		"object": "chat.completion",
		"created": 1,
		"model": %q,
		"choices": [{"index": 0, "message": {"role": "assistant", "content": %q}, "finish_reason": "stop"}]
	}`, model, reply)
}
