package cursorbox

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"connectrpc.com/connect"
	cursorv1 "github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1/cursorv1connect"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// defaultMaxRunSeconds bounds a turn whose config leaves max_run_seconds
	// unset, so a runaway Cursor loop cannot hold a session lease and a box
	// bridge process forever.
	defaultMaxRunSeconds = 1800

	// controlCallTimeout bounds CreateSession (the box spawns a bridge
	// process and creates the Cursor agent; its own setup bound is 60s).
	controlCallTimeout = 90 * time.Second

	// abortTimeout bounds the best-effort AbortSession call issued after the
	// turn's own context is already dead.
	abortTimeout = 5 * time.Second

	// apiKeyErrorReason is the stable google.rpc.ErrorInfo reason the box
	// attaches when CURSOR_API_KEY is missing or rejected by Cursor.
	apiKeyErrorReason = "CURSOR_API_KEY_MISSING_OR_INVALID"
)

// AgentBuilder adapts a ClientFactory into the internal/agent CURSOR
// construction seam: one Bridge (and one ADK agent) per AGENT_TYPE_CURSOR
// proto.
func AgentBuilder(factory ClientFactory) internalagent.CursorAgentBuilder {
	return func(pb *agentsv1.Agent) (agent.Agent, error) {
		b := NewBridge(pb, factory)
		description := pb.GetDescription()
		if description == "" {
			description = fmt.Sprintf("Cursor agent on ButterBox %s", b.butterboxID)
		}
		return b.BuildAgent(pb.GetName(), description)
	}
}

// Bridge holds one CURSOR agent's binding settings. It is stateless across
// turns: the Cursor session mapping lives in ADK session state, and the
// CursorService client is resolved per turn so box edits take effect
// immediately.
type Bridge struct {
	factory     ClientFactory
	workspaceID string
	agentID     string
	butterboxID string
	workingDir  string
	model       string
	mode        string
	maxRun      time.Duration // 0 = unlimited
}

// NewBridge constructs a Bridge from a CURSOR agent proto. The caller is
// expected to have validated the config (ValidateCursorAgent).
func NewBridge(pb *agentsv1.Agent, factory ClientFactory) *Bridge {
	c := pb.GetConfig().GetCursor()
	maxRun := time.Duration(defaultMaxRunSeconds) * time.Second
	if c.MaxRunSeconds != nil {
		maxRun = time.Duration(c.GetMaxRunSeconds()) * time.Second
	}
	return &Bridge{
		factory:     factory,
		workspaceID: pb.GetWorkspaceId(),
		agentID:     pb.GetAgentId(),
		butterboxID: strings.TrimSpace(c.GetButterboxId()),
		workingDir:  strings.TrimSpace(c.GetWorkingDir()),
		model:       strings.TrimSpace(c.GetModel()),
		mode:        strings.TrimSpace(c.GetMode()),
		maxRun:      maxRun,
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

// run executes one turn: ensure a Cursor session, send the prompt on one
// held call, and yield the final text.
func (b *Bridge) run(ictx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		input := extractText(ictx.UserContent())
		images := extractImages(ictx.UserContent())
		if input == "" && len(images) == 0 {
			yield(nil, fmt.Errorf("cursorbox: empty user input"))
			return
		}

		client, err := b.factory.ClientFor(ictx, b.workspaceID, b.butterboxID)
		if err != nil {
			yield(nil, err)
			return
		}

		// One deadline bounds the whole turn — session creation and the held
		// SendMessage — and ctx cancellation flows through the same path
		// (workflow node timeouts compose here too).
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

		text, err := b.send(runCtx, client, bnd.CursorSessionID, input, images)
		if err != nil && connect.CodeOf(err) == connect.CodeNotFound && !created {
			// The box lost the session we were reusing (box restart). Abandon
			// the stale binding and recreate once.
			bnd, err = b.createSession(runCtx, ictx, client)
			if err != nil {
				yield(nil, b.classifyInterruption(ictx, runCtx, err))
				return
			}
			if !b.yieldBinding(ictx, bnd, yield) {
				return
			}
			text, err = b.send(runCtx, client, bnd.CursorSessionID, input, images)
		}
		if err != nil {
			if runCtx.Err() != nil {
				// Our side gave up (cancel or max-run): the box stops the run
				// when the held call drops, but abort explicitly so the
				// cancellation never depends on connection teardown.
				b.abort(client, bnd.CursorSessionID)
				yield(nil, b.classifyInterruption(ictx, runCtx, err))
				return
			}
			if connect.CodeOf(err) == connect.CodeNotFound {
				// The session vanished right after we created (or recreated)
				// it — recreating again would loop, so report the box state.
				yield(nil, fmt.Errorf("cursorbox: the box lost a freshly created Cursor session; the box looks unhealthy — check it before retrying: %w", err))
				return
			}
			yield(nil, b.actionable("send message", err))
			return
		}

		evt := session.NewEvent(ictx, ictx.InvocationID())
		evt.Author = ictx.Agent().Name()
		evt.Content = genai.NewContentFromText(text, genai.RoleModel)
		yield(evt, nil)
	}
}

// ensureSession returns the Cursor session this turn runs in: the bound
// session when it still matches the agent's box and working directory,
// otherwise a fresh one.
func (b *Bridge) ensureSession(runCtx context.Context, ictx agent.InvocationContext, client cursorv1connect.CursorServiceClient) (binding, bool, error) {
	if bnd, ok := readBinding(ictx.Session().State(), b.agentID); ok && bnd.matches(b.butterboxID, b.workingDir) {
		return bnd, false, nil
	}
	bnd, err := b.createSession(runCtx, ictx, client)
	if err != nil {
		return binding{}, false, err
	}
	return bnd, true, nil
}

func (b *Bridge) createSession(runCtx context.Context, ictx agent.InvocationContext, client cursorv1connect.CursorServiceClient) (binding, error) {
	callCtx, cancel := context.WithTimeout(runCtx, controlCallTimeout)
	defer cancel()
	resp, err := client.CreateSession(callCtx, connect.NewRequest(&cursorv1.CreateSessionRequest{
		Name:  fmt.Sprintf("butter:%s:%s", b.agentID, ictx.Session().ID()),
		Model: b.model,
		Mode:  b.mode,
		Cwd:   b.workingDir,
	}))
	if err != nil {
		return binding{}, b.actionable("create Cursor session", err)
	}
	id := resp.Msg.GetSessionId()
	if id == "" {
		return binding{}, fmt.Errorf("cursorbox: the box answered CreateSession without a session id")
	}
	return binding{CursorSessionID: id, ButterboxID: b.butterboxID, WorkingDir: b.workingDir}, nil
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

// send holds one SendMessage call for the whole Cursor turn; max_run_seconds
// and cancellation end it. The call context carries runCtx's cancellation
// but not its deadline: a propagated Connect timeout would expire on the box
// first and race our own classification, so butter is always the side that
// gives up and the box sees the dropped call.
func (b *Bridge) send(runCtx context.Context, client cursorv1connect.CursorServiceClient, sessionID, input string, images []*cursorv1.ImageContent) (string, error) {
	callCtx, cancel := context.WithCancel(context.WithoutCancel(runCtx))
	defer cancel()
	stop := context.AfterFunc(runCtx, cancel)
	defer stop()
	resp, err := client.SendMessage(callCtx, connect.NewRequest(&cursorv1.SendMessageRequest{
		SessionId: sessionID,
		Message:   input,
		Images:    images,
	}))
	if err != nil {
		return "", err
	}
	return resp.Msg.GetText(), nil
}

// abort is the single cancellation path: ctx cancellation and the max-run
// deadline both land here. Best-effort — the turn's own context is already
// dead, so the call runs on a fresh bounded one.
func (b *Bridge) abort(client cursorv1connect.CursorServiceClient, sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), abortTimeout)
	defer cancel()
	_, _ = client.AbortSession(ctx, connect.NewRequest(&cursorv1.AbortSessionRequest{SessionId: sessionID}))
}

// classifyInterruption turns a turn-ending error into what the caller should
// see: the caller's own cancellation, a raise-the-limit deadline message, or
// the error as-is.
func (b *Bridge) classifyInterruption(ictx agent.InvocationContext, runCtx context.Context, err error) error {
	if cerr := ictx.Err(); cerr != nil {
		return cerr
	}
	if runCtx.Err() != nil {
		return fmt.Errorf("cursorbox: the run exceeded max_run_seconds=%d and was aborted on the box; raise max_run_seconds on the agent for long runs", int(b.maxRun/time.Second))
	}
	return err
}

// actionable maps box error codes onto errors that tell the user what to do.
func (b *Bridge) actionable(op string, err error) error {
	switch connect.CodeOf(err) {
	case connect.CodeResourceExhausted:
		return fmt.Errorf("cursorbox: the ButterBox is at its Cursor session capacity; retry later or raise CURSOR_MAX_SESSIONS on the box: %w", err)
	case connect.CodeFailedPrecondition:
		return fmt.Errorf("cursorbox: the Cursor session is busy with another message; wait for it to finish or cancel it: %w", err)
	case connect.CodeUnauthenticated, connect.CodePermissionDenied:
		if isAPIKeyError(err) {
			return fmt.Errorf("cursorbox: the Cursor API key is missing or invalid; set CURSOR_API_KEY on the box and restart it: %w", err)
		}
		return fmt.Errorf("cursorbox: the ButterBox rejected the access token; rotate it via SetButterBoxToken: %w", err)
	case connect.CodeInvalidArgument:
		return fmt.Errorf("cursorbox: the box rejected the agent's Cursor settings; check working_dir (it must stay inside the box sandbox) and mode: %w", err)
	case connect.CodeCanceled:
		return fmt.Errorf("cursorbox: the run was cancelled on the box before it produced an answer: %w", err)
	case connect.CodeUnavailable:
		return fmt.Errorf("cursorbox: the ButterBox is unreachable or its Cursor bridge exited — the run did not finish and no answer was produced: %w", err)
	default:
		return fmt.Errorf("cursorbox: %s: %w", op, err)
	}
}

// isAPIKeyError reports whether the box tagged the error with its stable
// Cursor API key ErrorInfo reason.
func isAPIKeyError(err error) bool {
	cerr, ok := err.(*connect.Error)
	if !ok {
		return false
	}
	for _, d := range cerr.Details() {
		v, derr := d.Value()
		if derr != nil {
			continue
		}
		if info, ok := v.(*errdetails.ErrorInfo); ok && info.GetReason() == apiKeyErrorReason {
			return true
		}
	}
	return false
}

// extractText flattens the user content's text parts.
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

// extractImages passes the user's inline images through to Cursor as raw
// bytes.
func extractImages(c *genai.Content) []*cursorv1.ImageContent {
	if c == nil {
		return nil
	}
	var images []*cursorv1.ImageContent
	for _, p := range c.Parts {
		if p.InlineData == nil || len(p.InlineData.Data) == 0 {
			continue
		}
		if !strings.HasPrefix(p.InlineData.MIMEType, "image/") {
			continue
		}
		images = append(images, &cursorv1.ImageContent{
			MimeType: p.InlineData.MIMEType,
			Data:     p.InlineData.Data,
		})
	}
	return images
}
