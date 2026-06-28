// Package opencode bridges an `opencode serve` HTTP instance into the ADK
// agent interface so it can be referenced as a RemoteAgent (protocol
// OPENCODE_HTTP). Each invocation creates a one-shot opencode session,
// posts the user input, and yields the joined text response.
//
// API: https://opencode.ai/docs/zh-cn/server/
package opencode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	defaultHTTPTimeout = 5 * time.Minute
	defaultUsername    = "opencode"
)

// Bridge holds the connection settings for a single OPENCODE_HTTP remote agent.
type Bridge struct {
	baseURL       string
	username      string
	password      string
	opencodeAgent string
	opencodeModel string
	http          *http.Client
}

// NewBridge constructs a Bridge from a RemoteAgent proto. The caller is
// expected to have already validated protocol == OPENCODE_HTTP and a non-empty
// URL via the service-level validator.
func NewBridge(ra *agentsv1.RemoteAgent) *Bridge {
	username := strings.TrimSpace(ra.GetUsername())
	if username == "" && strings.TrimSpace(ra.GetPassword()) != "" {
		username = defaultUsername
	}
	return &Bridge{
		baseURL:       strings.TrimRight(strings.TrimSpace(ra.GetUrl()), "/"),
		username:      username,
		password:      ra.GetPassword(),
		opencodeAgent: strings.TrimSpace(ra.GetOpencodeAgent()),
		opencodeModel: strings.TrimSpace(ra.GetOpencodeModel()),
		http:          &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// SetHTTPClient overrides the HTTP client (used by tests).
func (b *Bridge) SetHTTPClient(c *http.Client) {
	if c != nil {
		b.http = c
	}
}

// BuildAgent produces an ADK agent that delegates each run to the opencode
// server. ADK's agent.Agent has an unexported method, so we go through
// agent.New (same pattern as the daemon bridge).
func (b *Bridge) BuildAgent(name, description string) (agent.Agent, error) {
	return agent.New(agent.Config{
		Name:        name,
		Description: description,
		Run:         b.run,
	})
}

// run executes one invocation against the opencode server.
func (b *Bridge) run(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		input := extractText(ctx.UserContent())
		if input == "" {
			yield(nil, fmt.Errorf("opencode: empty user input"))
			return
		}

		sessID, err := b.createSession(ctx)
		if err != nil {
			yield(nil, err)
			return
		}

		// Best-effort cancellation: when the invocation is cancelled mid-flight
		// we ask opencode to abort the running session before returning.
		msgCtx, cancelMsg := context.WithCancel(ctx)
		defer cancelMsg()
		abortDone := make(chan struct{})
		go func() {
			defer close(abortDone)
			select {
			case <-ctx.Done():
				cancelMsg()
				_ = b.abortSession(context.Background(), sessID)
			case <-msgCtx.Done():
			}
		}()

		text, err := b.sendMessage(msgCtx, sessID, input)
		cancelMsg()
		<-abortDone
		if err != nil {
			if cerr := ctx.Err(); cerr != nil {
				yield(nil, cerr)
				return
			}
			yield(nil, err)
			return
		}

		event := session.NewEvent(ctx.InvocationID())
		event.Author = ctx.Agent().Name()
		event.Content = genai.NewContentFromText(text, genai.RoleModel)
		yield(event, nil)
	}
}

// createSession calls POST /session and returns the new session id.
func (b *Bridge) createSession(ctx context.Context) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := b.do(ctx, http.MethodPost, "/session", map[string]any{}, &out); err != nil {
		return "", fmt.Errorf("opencode create session: %w", err)
	}
	if out.ID == "" {
		return "", fmt.Errorf("opencode create session: empty id in response")
	}
	return out.ID, nil
}

// sendMessage posts a user message to the session and waits for the assistant
// reply, joining all text-typed parts.
func (b *Bridge) sendMessage(ctx context.Context, sessID, input string) (string, error) {
	body := map[string]any{
		"parts": []map[string]any{
			{"type": "text", "text": input},
		},
	}
	if b.opencodeAgent != "" {
		body["agent"] = b.opencodeAgent
	}
	if b.opencodeModel != "" {
		body["model"] = b.opencodeModel
	}

	var resp struct {
		Parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"parts"`
	}
	path := "/session/" + sessID + "/message"
	if err := b.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return "", fmt.Errorf("opencode send message: %w", err)
	}
	var out strings.Builder
	for _, p := range resp.Parts {
		if p.Type == "text" && p.Text != "" {
			if out.Len() > 0 {
				out.WriteString("\n")
			}
			out.WriteString(p.Text)
		}
	}
	return out.String(), nil
}

// abortSession requests opencode to cancel the running session.
func (b *Bridge) abortSession(ctx context.Context, sessID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return b.do(ctx, http.MethodPost, "/session/"+sessID+"/abort", nil, nil)
}

// do is a small helper that marshals a JSON body, attaches basic auth, sends
// the request, and (if dst != nil) decodes the JSON response.
func (b *Bridge) do(ctx context.Context, method, path string, body any, dst any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, b.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if b.username != "" || b.password != "" {
		req.SetBasicAuth(b.username, b.password)
	}

	resp, err := b.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("opencode %s %s: status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(errBody)))
	}
	if dst == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil && err != io.EOF {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// extractText flattens a genai.Content into a single string by concatenating
// all text parts (mirrors daemon.extractText).
func extractText(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var parts []string
	for _, p := range c.Parts {
		if p.Text != "" {
			parts = append(parts, p.Text)
		}
	}
	return strings.Join(parts, "\n")
}
