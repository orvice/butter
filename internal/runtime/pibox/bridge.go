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

	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// defaultMaxRunSeconds bounds a turn whose config leaves max_run_seconds
	// unset. The bound exists to stop runaway loops from holding a session
	// lease and a box process slot forever; long legitimate runs raise it.
	defaultMaxRunSeconds = 1800

	// defaultPollWaitSeconds is the long-poll window passed to GetTurn
	// (capped at 30 by the box).
	defaultPollWaitSeconds = 30

	// controlCallTimeout bounds the short control RPCs (CreateSession,
	// SubmitMessage). Only GetTurn is long.
	controlCallTimeout = 60 * time.Second

	// abortTimeout bounds the best-effort AbortSession call issued after the
	// turn's own context is already dead.
	abortTimeout = 5 * time.Second

	// maxConsecutivePollFailures is how many GetTurn transport errors in a
	// row the bridge tolerates before declaring the box unreachable. A single
	// dropped poll must not kill a 20-minute run — the cursor makes retrying
	// safe — but a box that stays silent gets an honest error.
	maxConsecutivePollFailures = 3
)

// AgentBuilder adapts a ClientFactory into the internal/agent PI construction
// seam: one Bridge (and one ADK agent) per AGENT_TYPE_PI proto.
func AgentBuilder(factory ClientFactory) internalagent.PiAgentBuilder {
	return func(pb *agentsv1.Agent) (agent.Agent, error) {
		b := NewBridge(pb, factory)
		description := pb.GetDescription()
		if description == "" {
			description = fmt.Sprintf("Pi agent on ButterBox %s", b.butterboxID)
		}
		return b.BuildAgent(pb.GetName(), description)
	}
}

// Bridge holds one PI agent's binding settings. It is stateless across
// turns: the pi session mapping lives in ADK session state, and the
// PiService client is resolved per turn so box edits take effect
// immediately.
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

	// Tuning knobs, overridden by tests.
	pollWaitSeconds int32
	pollRetryDelay  time.Duration
}

// NewBridge constructs a Bridge from a PI agent proto. The caller is
// expected to have validated the config (ValidatePiAgent).
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
		pollWaitSeconds: defaultPollWaitSeconds,
		pollRetryDelay:  2 * time.Second,
	}
}

// BuildAgent produces the ADK agent that delegates each run to the box.
func (b *Bridge) BuildAgent(name, description string) (agent.Agent, error) {
	return agent.New(agent.Config{
		Name:        name,
		Description: description,
		Run:         b.run,
	})
}

// run executes one turn: ensure a pi session, submit the prompt, long-poll
// the cursor until the run settles, and yield the assistant text.
func (b *Bridge) run(ictx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		input := extractText(ictx.UserContent())
		images := extractImages(ictx.UserContent())
		if input == "" && len(images) == 0 {
			yield(nil, fmt.Errorf("pibox: empty user input"))
			return
		}

		client, err := b.factory.ClientFor(ictx, b.workspaceID, b.butterboxID)
		if err != nil {
			yield(nil, err)
			return
		}

		// One deadline bounds the whole turn — session activation, submit,
		// and the poll loop — and ctx cancellation flows through the same
		// path (workflow node timeouts compose here too).
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
		if created {
			if !b.yieldBinding(ictx, bnd, yield) {
				return
			}
		}

		cursor, err := b.submit(runCtx, client, bnd.PiSessionID, input, images)
		if err != nil && connect.CodeOf(err) == connect.CodeNotFound && !created {
			// The box lost the session we were reusing (restart, box-side
			// cleanup). Abandon the stale binding and recreate once.
			bnd, err = b.createSession(runCtx, ictx, client)
			if err != nil {
				yield(nil, b.classifyInterruption(ictx, runCtx, err))
				return
			}
			if !b.yieldBinding(ictx, bnd, yield) {
				return
			}
			cursor, err = b.submit(runCtx, client, bnd.PiSessionID, input, images)
		}
		if err != nil {
			// The submit may have reached the box even though the response
			// did not reach us; when our side gave up, abort defensively.
			if runCtx.Err() != nil {
				b.abort(client, bnd.PiSessionID)
			}
			if connect.CodeOf(err) == connect.CodeNotFound {
				// The session vanished right after we created (or recreated)
				// it — recreating again would loop, so report the box state.
				err = fmt.Errorf("pibox: the box lost a freshly created pi session; the box looks unhealthy — check it before retrying: %w", err)
			}
			yield(nil, b.classifyInterruption(ictx, runCtx, err))
			return
		}

		result, err := b.await(runCtx, client, bnd.PiSessionID, cursor)
		if err != nil {
			yield(nil, b.classifyInterruption(ictx, runCtx, err))
			return
		}
		if result.GetStopReason() == "aborted" {
			yield(nil, fmt.Errorf("pibox: the run was aborted on the box before it produced an answer"))
			return
		}

		evt := session.NewEvent(ictx, ictx.InvocationID())
		evt.Author = ictx.Agent().Name()
		evt.Content = genai.NewContentFromText(result.GetText(), genai.RoleModel)
		yield(evt, nil)
	}
}

// ensureSession returns the pi session this turn runs in: the bound session
// when it still matches the agent's box and working directory, otherwise a
// fresh one. A stale binding (repointed agent) is abandoned, never migrated.
func (b *Bridge) ensureSession(runCtx context.Context, ictx agent.InvocationContext, client piv1connect.PiServiceClient) (binding, bool, error) {
	if bnd, ok := readBinding(ictx.Session().State(), b.agentID); ok && bnd.matches(b.butterboxID, b.workingDir) {
		return bnd, false, nil
	}
	bnd, err := b.createSession(runCtx, ictx, client)
	if err != nil {
		return binding{}, false, err
	}
	return bnd, true, nil
}

func (b *Bridge) createSession(runCtx context.Context, ictx agent.InvocationContext, client piv1connect.PiServiceClient) (binding, error) {
	callCtx, cancel := context.WithTimeout(runCtx, controlCallTimeout)
	defer cancel()
	resp, err := client.CreateSession(callCtx, connect.NewRequest(&piv1.CreateSessionRequest{
		Name:          fmt.Sprintf("butter:%s:%s", b.agentID, ictx.Session().ID()),
		Provider:      b.provider,
		Model:         b.model,
		ThinkingLevel: b.thinking,
		Cwd:           b.workingDir,
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

// yieldBinding persists a freshly created session binding through an event's
// StateDelta before the turn's outcome is known, so continuity survives a
// turn that later fails. Returns false when the consumer stopped the run.
func (b *Bridge) yieldBinding(ictx agent.InvocationContext, bnd binding, yield func(*session.Event, error) bool) bool {
	evt := session.NewEvent(ictx, ictx.InvocationID())
	evt.Author = ictx.Agent().Name()
	evt.Actions.StateDelta[stateKey(b.agentID)] = bnd.stateValue()
	return yield(evt, nil)
}

func (b *Bridge) submit(runCtx context.Context, client piv1connect.PiServiceClient, sessionID, input string, images []*piv1.ImageContent) (string, error) {
	callCtx, cancel := context.WithTimeout(runCtx, controlCallTimeout)
	defer cancel()
	resp, err := client.SubmitMessage(callCtx, connect.NewRequest(&piv1.SubmitMessageRequest{
		SessionId: sessionID,
		Message:   input,
		Images:    images,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return "", err // recreate decision belongs to the caller
		}
		return "", b.actionable("submit message", err)
	}
	return resp.Msg.GetTurnCursor(), nil
}

// await long-polls GetTurn until the run settles. The cursor is stable
// across box restarts, so completion is judged honestly: running=false
// without a result means the turn did not finish, never a stale answer.
func (b *Bridge) await(runCtx context.Context, client piv1connect.PiServiceClient, sessionID, cursor string) (*piv1.TurnResult, error) {
	failures := 0
	for {
		if err := runCtx.Err(); err != nil {
			b.abort(client, sessionID)
			return nil, err
		}
		callCtx, cancel := context.WithTimeout(runCtx, time.Duration(b.pollWaitSeconds)*time.Second+15*time.Second)
		resp, err := client.GetTurn(callCtx, connect.NewRequest(&piv1.GetTurnRequest{
			SessionId:   sessionID,
			TurnCursor:  cursor,
			WaitSeconds: b.pollWaitSeconds,
		}))
		cancel()
		if err != nil {
			if runCtx.Err() != nil {
				b.abort(client, sessionID)
				return nil, runCtx.Err()
			}
			if connect.CodeOf(err) == connect.CodeNotFound {
				// The box no longer knows the session mid-run (restart plus a
				// lost session file). The submitted turn can never settle for
				// us, so this is a did-not-finish, not a transport blip.
				return nil, fmt.Errorf("pibox: the box lost the pi session mid-run — the run did not finish and no answer was produced: %w", err)
			}
			failures++
			if failures >= maxConsecutivePollFailures {
				return nil, fmt.Errorf("pibox: lost contact with the box while awaiting the turn (the run may still be executing on the box; it stays visible in pi-web): %w", err)
			}
			select {
			case <-runCtx.Done():
			case <-time.After(b.pollRetryDelay):
			}
			continue
		}
		failures = 0
		if resp.Msg.GetRunning() {
			continue
		}
		if result := resp.Msg.GetResult(); result != nil {
			return result, nil
		}
		return nil, fmt.Errorf("pibox: the run did not finish on the box — no answer was produced (the box may have restarted mid-run, or the run was aborted)")
	}
}

// abort is the single cancellation path (ADR-0011 §4): ctx cancellation and
// the max-run deadline both land here. Best-effort — the turn's own context
// is already dead, so the call runs on a fresh bounded one.
func (b *Bridge) abort(client piv1connect.PiServiceClient, sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), abortTimeout)
	defer cancel()
	_, _ = client.AbortSession(ctx, connect.NewRequest(&piv1.AbortSessionRequest{SessionId: sessionID}))
}

// classifyInterruption turns a turn-ending error into what the caller should
// see: the caller's own cancellation, a raise-the-limit deadline message, or
// the error as-is.
func (b *Bridge) classifyInterruption(ictx agent.InvocationContext, runCtx context.Context, err error) error {
	if cerr := ictx.Err(); cerr != nil {
		// The invocation itself was cancelled (user cancel, workflow node
		// timeout); report that, not a wrapped transport error.
		return cerr
	}
	if runCtx.Err() != nil {
		// Only the bridge's own max-run deadline distinguishes runCtx from
		// the invocation context.
		return fmt.Errorf("pibox: the run exceeded max_run_seconds=%d and was aborted on the box; raise max_run_seconds on the agent for long runs", int(b.maxRun/time.Second))
	}
	return err
}

// actionable maps box error codes onto errors that tell the user what to do.
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

// extractText flattens the user content's text parts (mirrors the opencode
// bridge).
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

// extractImages passes the user's inline images through to pi as raw bytes
// (base64 on the JSON wire).
func extractImages(c *genai.Content) []*piv1.ImageContent {
	if c == nil {
		return nil
	}
	var images []*piv1.ImageContent
	for _, p := range c.Parts {
		if p.InlineData == nil || len(p.InlineData.Data) == 0 {
			continue
		}
		if !strings.HasPrefix(p.InlineData.MIMEType, "image/") {
			continue
		}
		images = append(images, &piv1.ImageContent{
			MimeType: p.InlineData.MIMEType,
			Data:     p.InlineData.Data,
		})
	}
	return images
}
