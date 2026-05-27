package acp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type fakeClient struct {
	req InvokeRequest
	err error
}

func (c *fakeClient) Invoke(_ context.Context, req InvokeRequest) (InvokeResponse, error) {
	c.req = req
	if c.err != nil {
		return InvokeResponse{}, c.err
	}
	return InvokeResponse{SessionID: "butter-session", Response: "agent response"}, nil
}

func TestServerInitializeAndPrompt(t *testing.T) {
	client := &fakeClient{}
	server, err := NewServer(Config{AgentName: "assistant"}, client)
	if err != nil {
		t.Fatal(err)
	}

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"session/new","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"session/prompt","params":{"sessionId":"s1","content":[{"type":"text","text":"hello"}]}}`,
		"",
	}, "\n")

	var out bytes.Buffer
	if err := server.Serve(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}

	responses := readJSONLines(t, out.String())
	if len(responses) != 4 {
		t.Fatalf("got %d JSON-RPC messages, want 4: %s", len(responses), out.String())
	}
	if got := responses[0]["result"].(map[string]any)["protocolVersion"]; got != float64(protocolVersion) {
		t.Fatalf("protocolVersion = %v", got)
	}
	if got := client.req.Input; got != "hello" {
		t.Fatalf("input = %q, want hello", got)
	}
	if got := client.req.AgentName; got != "assistant" {
		t.Fatalf("agent = %q, want assistant", got)
	}
	if got := responses[2]["method"]; got != "session/update" {
		t.Fatalf("third message method = %v, want session/update", got)
	}
	if responses[3]["error"] != nil {
		t.Fatalf("prompt returned error: %#v", responses[3]["error"])
	}
}

func TestServerPromptReturnsInvokeError(t *testing.T) {
	server, err := NewServer(Config{AgentName: "assistant"}, &fakeClient{err: errors.New("boom")})
	if err != nil {
		t.Fatal(err)
	}

	input := `{"jsonrpc":"2.0","id":1,"method":"session/prompt","params":{"sessionId":"s1","prompt":"hello"}}` + "\n"
	var out bytes.Buffer
	if err := server.Serve(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}

	responses := readJSONLines(t, out.String())
	errBody := responses[0]["error"].(map[string]any)
	if got := errBody["message"]; got != "boom" {
		t.Fatalf("error message = %v, want boom", got)
	}
}

func TestNewServerRequiresAgent(t *testing.T) {
	_, err := NewServer(Config{}, &fakeClient{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func readJSONLines(t *testing.T, content string) []map[string]any {
	t.Helper()
	scanner := bufio.NewScanner(strings.NewReader(content))
	var out []map[string]any
	for scanner.Scan() {
		var msg map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			t.Fatalf("unmarshal %q: %v", scanner.Text(), err)
		}
		out = append(out, msg)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}
