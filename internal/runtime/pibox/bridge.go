package pibox

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/a2ui"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	defaultMaxRunSeconds       = 1800
	defaultPollWaitSeconds     = 30
	controlCallTimeout         = 60 * time.Second
	abortTimeout               = 5 * time.Second
	maxConsecutivePollFailures = 3
)

// Bridge is immutable across turns. Pi session bindings live in ADK session
// state; clients and credentials are resolved afresh for every turn.
type Bridge struct {
	factory     ClientFactory
	workspaceID string
	agentID     string
	butterboxID string
	workingDir  string
	provider    string
	model       string
	thinking    string
	maxRun      time.Duration // 0 = unlimited
	cardPolicy  a2ui.CardPolicy

	// Tuning knobs overridden by tests.
	pollWaitSeconds int32
	pollRetryDelay  time.Duration
}

// NewBridge constructs a bridge from a validated Pi agent proto. Factories
// supply inherited Card Policy through AgentBuilder.
func NewBridge(pb *agentsv1.Agent, factory ClientFactory) *Bridge {
	pi := pb.GetConfig().GetPi()
	maxRun := time.Duration(defaultMaxRunSeconds) * time.Second
	if pi.MaxRunSeconds != nil {
		maxRun = time.Duration(pi.GetMaxRunSeconds()) * time.Second
	}
	return &Bridge{
		factory:         factory,
		workspaceID:     pb.GetWorkspaceId(),
		agentID:         pb.GetAgentId(),
		butterboxID:     strings.TrimSpace(pi.GetButterboxId()),
		workingDir:      strings.TrimSpace(pi.GetWorkingDir()),
		provider:        strings.TrimSpace(pi.GetProvider()),
		model:           strings.TrimSpace(pi.GetModel()),
		thinking:        strings.TrimSpace(pi.GetThinkingLevel()),
		maxRun:          maxRun,
		cardPolicy:      a2ui.CardPolicy{}.Narrow(pb.GetConfig().GetResultCards()),
		pollWaitSeconds: defaultPollWaitSeconds,
		pollRetryDelay:  2 * time.Second,
	}
}

// BuildAgent produces the ADK custom agent that delegates each run to the box.
func (b *Bridge) BuildAgent(name, description string) (agent.Agent, error) {
	return agent.New(agent.Config{Name: name, Description: description, Run: b.run})
}

func (b *Bridge) run(ictx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		input, images := extractText(ictx.UserContent()), extractImages(ictx.UserContent())
		if input == "" && len(images) == 0 {
			yield(nil, fmt.Errorf("pibox: empty user input"))
			return
		}
		client, err := b.factory.ClientFor(ictx, b.workspaceID, b.butterboxID)
		if err != nil {
			yield(nil, err)
			return
		}
		runCtx := context.Context(ictx)
		if b.maxRun > 0 {
			var cancel context.CancelFunc
			runCtx, cancel = context.WithTimeout(runCtx, b.maxRun)
			defer cancel()
		}
		bnd, created, err := b.ensureSession(runCtx, ictx, client)
		if err != nil {
			yield(nil, b.classifyInterruption(ictx, runCtx, err))
			return
		}
		if created && !b.yieldBinding(ictx, bnd, yield) {
			return
		}

		submitted, err := b.submit(runCtx, client, bnd.PiSessionID, input, images, ictx)
		if err != nil && connect.CodeOf(err) == connect.CodeNotFound && !created {
			// Recreate once when a previously bound Pi session disappeared.
			bnd, err = b.createSession(runCtx, ictx, client)
			if err != nil {
				yield(nil, b.classifyInterruption(ictx, runCtx, err))
				return
			}
			if !b.yieldBinding(ictx, bnd, yield) {
				return
			}
			submitted, err = b.submit(runCtx, client, bnd.PiSessionID, input, images, ictx)
		}
		if err != nil {
			// Submission is never retried after an uncertain acceptance. A
			// cancelled submission is aborted defensively on the exact session.
			if runCtx.Err() != nil {
				b.abort(client, bnd.PiSessionID)
			}
			if connect.CodeOf(err) == connect.CodeNotFound {
				err = fmt.Errorf("pibox: the box lost a freshly created pi session; the box looks unhealthy — check it before retrying: %w", err)
			}
			yield(nil, b.classifyInterruption(ictx, runCtx, err))
			return
		}

		settled := false
		defer func() {
			// Also covers consumer/persistence failure while a Pi tool awaits
			// its receipt. A disconnected browser does not end a detached run.
			if !settled {
				b.abort(client, bnd.PiSessionID)
			}
		}()
		var result *piv1.TurnResult
		if submitted.host {
			result, err = b.awaitHost(runCtx, client, bnd.PiSessionID, submitted, ictx, yield)
		} else {
			result, err = b.await(runCtx, client, bnd.PiSessionID, submitted.cursor)
		}
		if err != nil {
			yield(nil, b.classifyInterruption(ictx, runCtx, err))
			return
		}
		settled = true
		switch result.GetStopReason() {
		case "aborted":
			yield(nil, fmt.Errorf("pibox: the run was aborted on the box before it produced an answer"))
			return
		case "toolUse", "error":
			// A Box restart may leave only the assistant's tool request in
			// the transcript. That is not a completed answer.
			yield(nil, fmt.Errorf("pibox: the run did not finish on the box (stop reason %s)", result.GetStopReason()))
			return
		}
		evt := session.NewEvent(ictx, ictx.InvocationID())
		evt.Author = ictx.Agent().Name()
		evt.Content = genai.NewContentFromText(result.GetText(), genai.RoleModel)
		yield(evt, nil)
	}
}

func (b *Bridge) ensureSession(ctx context.Context, ictx agent.InvocationContext, client piv1connect.PiServiceClient) (binding, bool, error) {
	if bnd, ok := readBinding(ictx.Session().State(), b.agentID); ok && bnd.matches(b.butterboxID, b.workingDir) {
		return bnd, false, nil
	}
	bnd, err := b.createSession(ctx, ictx, client)
	return bnd, true, err
}

func (b *Bridge) createSession(ctx context.Context, ictx agent.InvocationContext, client piv1connect.PiServiceClient) (binding, error) {
	callCtx, cancel := context.WithTimeout(ctx, controlCallTimeout)
	defer cancel()
	resp, err := client.CreateSession(callCtx, connect.NewRequest(&piv1.CreateSessionRequest{
		Name:     fmt.Sprintf("butter:%s:%s", b.agentID, ictx.Session().ID()),
		Provider: b.provider, Model: b.model, ThinkingLevel: b.thinking, Cwd: b.workingDir,
	}))
	if err != nil {
		return binding{}, b.actionable("create pi session", err)
	}
	id := resp.Msg.GetSession().GetId()
	if id == "" {
		return binding{}, fmt.Errorf("pibox: the box answered CreateSession without a session id")
	}
	return binding{PiSessionID: id, ButterboxID: b.butterboxID, WorkingDir: b.workingDir}, nil
}

// The binding is written before the turn's outcome, so it survives failure.
func (b *Bridge) yieldBinding(ictx agent.InvocationContext, bnd binding, yield func(*session.Event, error) bool) bool {
	evt := session.NewEvent(ictx, ictx.InvocationID())
	evt.Author = ictx.Agent().Name()
	evt.Actions.StateDelta[stateKey(b.agentID)] = bnd.stateValue()
	return yield(evt, nil)
}

type submission struct {
	cursor string
	turnID string
	host   bool
}

func (b *Bridge) submit(ctx context.Context, client piv1connect.PiServiceClient, sessionID, input string, images []*piv1.ImageContent, ictx agent.InvocationContext) (submission, error) {
	host, err := b.hostToolsConfig(ctx, client, sessionID, ictx)
	if err != nil {
		return submission{}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, controlCallTimeout)
	defer cancel()
	resp, err := client.SubmitMessage(callCtx, connect.NewRequest(&piv1.SubmitMessageRequest{
		SessionId: sessionID, Message: input, Images: images, HostTools: host,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return submission{}, err
		}
		return submission{}, b.actionable("submit message", err)
	}
	if host != nil && resp.Msg.GetTurnId() == "" {
		b.abort(client, sessionID)
		return submission{}, fmt.Errorf("pibox: the box accepted a host-tool run without a turn id")
	}
	return submission{cursor: resp.Msg.GetTurnCursor(), turnID: resp.Msg.GetTurnId(), host: host != nil}, nil
}

// await retains the legacy text-only path. The entries cursor remains stable
// across restarts; a missing outcome never becomes a stale previous answer.
func (b *Bridge) await(ctx context.Context, client piv1connect.PiServiceClient, sessionID, cursor string) (*piv1.TurnResult, error) {
	for {
		resp, err := b.pollTurn(ctx, client, sessionID, cursor, b.pollWaitSeconds)
		if err != nil {
			return nil, err
		}
		if !resp.GetRunning() {
			return completedTurn(resp)
		}
	}
}

func (b *Bridge) pollTurn(ctx context.Context, client piv1connect.PiServiceClient, sessionID, cursor string, wait int32) (*piv1.GetTurnResponse, error) {
	var err error
	for attempt := range maxConsecutivePollFailures {
		callCtx, cancel := context.WithTimeout(ctx, time.Duration(wait)*time.Second+15*time.Second)
		var resp *connect.Response[piv1.GetTurnResponse]
		resp, err = client.GetTurn(callCtx, connect.NewRequest(&piv1.GetTurnRequest{SessionId: sessionID, TurnCursor: cursor, WaitSeconds: wait}))
		cancel()
		if err == nil {
			return resp.Msg, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil, fmt.Errorf("pibox: the box lost the pi session mid-run — the run did not finish and no answer was produced: %w", err)
		}
		if attempt+1 < maxConsecutivePollFailures && !b.retry(ctx) {
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("pibox: lost contact with the box while awaiting the turn: %w", err)
}

func completedTurn(resp *piv1.GetTurnResponse) (*piv1.TurnResult, error) {
	if result := resp.GetResult(); result != nil {
		return result, nil
	}
	return nil, fmt.Errorf("pibox: the run did not finish on the box — no answer was produced (the box may have restarted mid-run, or the run was aborted)")
}

func (b *Bridge) retry(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(b.pollRetryDelay):
		return true
	}
}

// abort is the single best-effort cancellation path, on a fresh bounded context.
func (b *Bridge) abort(client piv1connect.PiServiceClient, sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), abortTimeout)
	defer cancel()
	_, _ = client.AbortSession(ctx, connect.NewRequest(&piv1.AbortSessionRequest{SessionId: sessionID}))
}

func (b *Bridge) classifyInterruption(ictx agent.InvocationContext, ctx context.Context, err error) error {
	if ictx.Err() != nil {
		return ictx.Err()
	}
	if ctx.Err() != nil {
		return fmt.Errorf("pibox: the run exceeded max_run_seconds=%d and was aborted on the box; raise max_run_seconds on the agent for long runs", int(b.maxRun/time.Second))
	}
	return err
}

func (b *Bridge) actionable(op string, err error) error {
	switch connect.CodeOf(err) {
	case connect.CodeResourceExhausted:
		return fmt.Errorf("pibox: the ButterBox is at its session capacity; retry later or raise PI_API_MAX_SESSIONS on the box: %w", err)
	case connect.CodeFailedPrecondition:
		return fmt.Errorf("pibox: the pi session is busy with another message; wait for it to finish or cancel it: %w", err)
	case connect.CodeUnauthenticated, connect.CodePermissionDenied:
		return fmt.Errorf("pibox: the ButterBox rejected the access token; rotate it via SetButterBoxToken: %w", err)
	default:
		return fmt.Errorf("pibox: %s: %w", op, err)
	}
}

func extractText(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var parts []string
	for _, p := range c.Parts {
		if p != nil && p.Text != "" {
			parts = append(parts, p.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func extractImages(c *genai.Content) []*piv1.ImageContent {
	if c == nil {
		return nil
	}
	var images []*piv1.ImageContent
	for _, p := range c.Parts {
		if p == nil || p.InlineData == nil || len(p.InlineData.Data) == 0 || !strings.HasPrefix(p.InlineData.MIMEType, "image/") {
			continue
		}
		images = append(images, &piv1.ImageContent{MimeType: p.InlineData.MIMEType, Data: p.InlineData.Data})
	}
	return images
}
