// Package opencode bridges an `opencode serve` HTTP instance into the ADK
// agent interface so it can be referenced as a RemoteAgent (protocol
// OPENCODE_HTTP). Each invocation creates a one-shot opencode session,
// posts the user input, and yields the joined text response.
//
// API: https://opencode.ai/docs/zh-cn/server/
package opencode

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"strings"
	"time"

	opencodeclient "github.com/orvice/opencode-go"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const defaultHTTPTimeout = 5 * time.Minute

// Bridge holds the connection settings for a single OPENCODE_HTTP remote agent.
type Bridge struct {
	client        *opencodeclient.Client
	opencodeAgent string
	opencodeModel string
}

// NewBridge constructs a Bridge from a RemoteAgent proto. The caller is
// expected to have already validated protocol == OPENCODE_HTTP and a non-empty
// URL via the service-level validator. Extra opencode client options can be
// appended to override defaults (e.g. WithHTTPClient for tests).
func NewBridge(ra *agentsv1.RemoteAgent, opts ...opencodeclient.Option) (*Bridge, error) {
	username := strings.TrimSpace(ra.GetUsername())
	password := ra.GetPassword()
	if username == "" && strings.TrimSpace(password) != "" {
		username = "opencode"
	}

	clientOpts := []opencodeclient.Option{
		opencodeclient.WithBaseURL(strings.TrimRight(strings.TrimSpace(ra.GetUrl()), "/")),
		opencodeclient.WithUsername(username),
		opencodeclient.WithPassword(password),
		opencodeclient.WithHTTPClient(&http.Client{Timeout: defaultHTTPTimeout}),
	}
	clientOpts = append(clientOpts, opts...)

	client, err := opencodeclient.NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("opencode: create client: %w", err)
	}

	return &Bridge{
		client:        client,
		opencodeAgent: strings.TrimSpace(ra.GetOpencodeAgent()),
		opencodeModel: strings.TrimSpace(ra.GetOpencodeModel()),
	}, nil
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
	sess, err := b.client.Session.Create(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("opencode create session: %w", err)
	}
	if sess.ID == "" {
		return "", fmt.Errorf("opencode create session: empty id in response")
	}
	return sess.ID, nil
}

// sendMessage posts a user message to the session and waits for the assistant
// reply, joining all text-typed parts.
func (b *Bridge) sendMessage(ctx context.Context, sessID, input string) (string, error) {
	params := opencodeclient.PromptParams{
		Parts: []opencodeclient.PartInput{opencodeclient.NewTextPart(input)},
	}
	if b.opencodeAgent != "" {
		params.Agent = b.opencodeAgent
	}
	if b.opencodeModel != "" {
		params.Model = parseModelRef(b.opencodeModel)
	}

	resp, err := b.client.Session.Prompt(ctx, sessID, params)
	if err != nil {
		return "", fmt.Errorf("opencode send message: %w", err)
	}

	var out strings.Builder
	for _, p := range resp.Parts {
		if p.Type == opencodeclient.PartTypeText && p.Text != "" {
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
	_, err := b.client.Session.Abort(ctx, sessID)
	return err
}

// parseModelRef converts a model string like "provider/model" into a
// ModelRef. If the string contains no "/", the whole value is used as ModelID.
func parseModelRef(raw string) *opencodeclient.ModelRef {
	if i := strings.IndexByte(raw, '/'); i >= 0 {
		return &opencodeclient.ModelRef{
			ProviderID: raw[:i],
			ModelID:    raw[i+1:],
		}
	}
	return &opencodeclient.ModelRef{ModelID: raw}
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
