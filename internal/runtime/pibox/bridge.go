// Package pibox bridges a pi coding-agent session on a ButterBox VM into
// the ADK agent interface (ADR-0011). Each butter session gets one pi session
// per agent, keyed in ADK session state; the bridge drives the async
// SubmitMessage + GetTurn loop so a dropped HTTP connection never kills a
// long-running pi turn.
package pibox

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

const (
	defaultMaxRunSeconds = 1800
	pollWaitSeconds      = 25
	httpTimeout          = 35 * time.Second
)

// stateKey returns the ADK session state key for a pi agent's session info.
func stateKey(agentID string) string { return "pibox:" + agentID }

// piSessionState is the value stored in ADK session state under stateKey.
type piSessionState struct {
	PiSessionID string `json:"pi_session_id"`
	ButterBoxID string `json:"butterbox_id"`
	WorkingDir  string `json:"working_dir"`
}

// PiClient is the PiService surface the bridge uses. Satisfied by the
// generated piv1connect.PiServiceClient; tests substitute a fake.
type PiClient interface {
	CreateSession(context.Context, *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error)
	GetSession(context.Context, *connect.Request[piv1.GetSessionRequest]) (*connect.Response[piv1.GetSessionResponse], error)
	SubmitMessage(context.Context, *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error)
	GetTurn(context.Context, *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error)
	AbortSession(context.Context, *connect.Request[piv1.AbortSessionRequest]) (*connect.Response[piv1.AbortSessionResponse], error)
}

// Bridge holds the settings for one PI agent and creates ADK agent instances
// that delegate to the box.
type Bridge struct {
	client        PiClient
	agentID       string
	butterboxID   string
	workingDir    string
	provider      string
	model         string
	thinkingLevel string
	maxRunSeconds int32
}

// Config groups the parameters for constructing a Bridge.
type Config struct {
	Client        PiClient
	AgentID       string
	ButterBoxID   string
	WorkingDir    string
	Provider      string
	Model         string
	ThinkingLevel string
	MaxRunSeconds int32
}

// NewBridge constructs a Bridge. Call BuildAgent to get an ADK agent.
func NewBridge(cfg Config) *Bridge {
	mrs := cfg.MaxRunSeconds
	if mrs == 0 {
		mrs = defaultMaxRunSeconds
	}
	return &Bridge{
		client:        cfg.Client,
		agentID:       cfg.AgentID,
		butterboxID:   cfg.ButterBoxID,
		workingDir:    cfg.WorkingDir,
		provider:      cfg.Provider,
		model:         cfg.Model,
		thinkingLevel: cfg.ThinkingLevel,
		maxRunSeconds: mrs,
	}
}

// NewPiClient builds a PiService Connect client with a bearer token.
func NewPiClient(baseURL, token string) piv1connect.PiServiceClient {
	httpClient := &http.Client{Timeout: httpTimeout}
	var opts []connect.ClientOption
	if token != "" {
		opts = append(opts, connect.WithInterceptors(bearerInterceptor(token)))
	}
	return piv1connect.NewPiServiceClient(httpClient, baseURL, opts...)
}

func bearerInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// BuildAgent produces an ADK agent whose Run delegates to the pibox bridge.
func (b *Bridge) BuildAgent(name, description string) (agent.Agent, error) {
	return agent.New(agent.Config{
		Name:        name,
		Description: description,
		Run:         b.run,
	})
}

// run is the ADK Run function.
func (b *Bridge) run(ictx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		text, images := extractContent(ictx.UserContent())
		if text == "" && len(images) == 0 {
			yield(nil, fmt.Errorf("pibox: empty user input"))
			return
		}

		piSessID, err := b.ensureSession(ictx)
		if err != nil {
			yield(nil, err)
			return
		}

		// Build a child context with the max-run deadline.
		var runCtx context.Context
		var runCancel context.CancelFunc
		if b.maxRunSeconds > 0 {
			runCtx, runCancel = context.WithTimeout(ictx, time.Duration(b.maxRunSeconds)*time.Second)
		} else {
			runCtx, runCancel = context.WithCancel(ictx)
		}
		defer runCancel()

		// Abort on cancellation (caller cancel or deadline).
		abortDone := make(chan struct{})
		go func() {
			defer close(abortDone)
			<-runCtx.Done()
			_ = b.abort(context.Background(), piSessID)
		}()

		result, err := b.submitAndPoll(runCtx, piSessID, text, images)
		runCancel()
		<-abortDone

		if err != nil {
			if runCtx.Err() != nil && ictx.Err() != nil {
				yield(nil, ictx.Err())
				return
			}
			if runCtx.Err() != nil {
				yield(nil, fmt.Errorf("pibox: turn exceeded max_run_seconds (%d); raise the limit or simplify the task", b.maxRunSeconds))
				return
			}
			yield(nil, err)
			return
		}

		event := session.NewEvent(ictx, ictx.InvocationID())
		event.Author = ictx.Agent().Name()
		event.Content = genai.NewContentFromText(result, genai.RoleModel)
		yield(event, nil)
	}
}

// ensureSession looks up or creates the pi session for this agent. On repoint
// (butterbox_id or working_dir changed) or not-found, it abandons and
// recreates.
func (b *Bridge) ensureSession(ictx agent.InvocationContext) (string, error) {
	sk := stateKey(b.agentID)
	var prev piSessionState
	if raw, err := ictx.Session().State().Get(sk); err == nil {
		if data, jsonErr := json.Marshal(raw); jsonErr == nil {
			_ = json.Unmarshal(data, &prev)
		}
	}

	repoint := prev.ButterBoxID != b.butterboxID || prev.WorkingDir != b.workingDir

	if prev.PiSessionID != "" && !repoint {
		_, err := b.client.GetSession(ictx, connect.NewRequest(&piv1.GetSessionRequest{
			SessionId: prev.PiSessionID,
		}))
		if err == nil {
			return prev.PiSessionID, nil
		}
		if connect.CodeOf(err) != connect.CodeNotFound {
			return "", fmt.Errorf("pibox: check session: %w", err)
		}
	}

	resp, err := b.client.CreateSession(ictx, connect.NewRequest(&piv1.CreateSessionRequest{
		Provider:      b.provider,
		Model:         b.model,
		ThinkingLevel: b.thinkingLevel,
		Cwd:           b.workingDir,
	}))
	if err != nil {
		return "", mapPiError(err, "create session")
	}
	piSessID := resp.Msg.GetSession().GetId()
	if piSessID == "" {
		return "", fmt.Errorf("pibox: create session returned empty id")
	}

	newState := piSessionState{
		PiSessionID: piSessID,
		ButterBoxID: b.butterboxID,
		WorkingDir:  b.workingDir,
	}
	stateJSON, _ := json.Marshal(newState)
	var stateMap map[string]any
	_ = json.Unmarshal(stateJSON, &stateMap)
	if err := ictx.Session().State().Set(sk, stateMap); err != nil {
		return "", fmt.Errorf("pibox: persist session state: %w", err)
	}

	return piSessID, nil
}

// submitAndPoll sends the message asynchronously and polls until the turn
// settles.
func (b *Bridge) submitAndPoll(ctx context.Context, sessionID, text string, images []*piv1.ImageContent) (string, error) {
	submitResp, err := b.client.SubmitMessage(ctx, connect.NewRequest(&piv1.SubmitMessageRequest{
		SessionId: sessionID,
		Message:   text,
		Images:    images,
	}))
	if err != nil {
		return "", mapPiError(err, "submit message")
	}
	cursor := submitResp.Msg.GetTurnCursor()

	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		resp, err := b.client.GetTurn(ctx, connect.NewRequest(&piv1.GetTurnRequest{
			SessionId:   sessionID,
			TurnCursor:  cursor,
			WaitSeconds: pollWaitSeconds,
		}))
		if err != nil {
			return "", mapPiError(err, "get turn")
		}

		if resp.Msg.GetResult() != nil {
			return resp.Msg.GetResult().GetText(), nil
		}

		if !resp.Msg.GetRunning() {
			return "", fmt.Errorf("pibox: turn did not finish (pi process may have restarted mid-run)")
		}
	}
}

func (b *Bridge) abort(ctx context.Context, sessionID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := b.client.AbortSession(ctx, connect.NewRequest(&piv1.AbortSessionRequest{
		SessionId: sessionID,
	}))
	return err
}

func mapPiError(err error, op string) error {
	code := connect.CodeOf(err)
	switch code {
	case connect.CodeResourceExhausted:
		return fmt.Errorf("pibox: box is at capacity; retry later or raise PI_API_MAX_SESSIONS on the box: %w", err)
	case connect.CodeUnavailable:
		return fmt.Errorf("pibox: box is unreachable; check the box is running and the token is valid: %w", err)
	case connect.CodeNotFound:
		return fmt.Errorf("pibox: session not found on the box: %w", err)
	default:
		return fmt.Errorf("pibox: %s: %w", op, err)
	}
}

func extractContent(c *genai.Content) (string, []*piv1.ImageContent) {
	if c == nil {
		return "", nil
	}
	var texts []string
	var images []*piv1.ImageContent
	for _, p := range c.Parts {
		if p.Text != "" {
			texts = append(texts, p.Text)
		}
		if p.InlineData != nil && strings.HasPrefix(p.InlineData.MIMEType, "image/") {
			images = append(images, &piv1.ImageContent{
				MimeType: p.InlineData.MIMEType,
				Data:     p.InlineData.Data,
			})
		}
	}
	return strings.Join(texts, "\n"), images
}
