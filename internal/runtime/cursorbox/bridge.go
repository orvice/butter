package cursorbox

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	defaultMaxRunSeconds = 1800

	// controlCallTimeout bounds CreateSession. SendMessage has its own
	// deadline driven by maxRun since the Cursor agent may run for minutes.
	controlCallTimeout = 60 * time.Second

	abortTimeout = 5 * time.Second
)

// AgentBuilder adapts a ClientFactory into the internal/agent Cursor
// construction seam: one Bridge (and one ADK agent) per AGENT_TYPE_CURSOR proto.
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
	cur := pb.GetConfig().GetCursor()
	maxRun := time.Duration(defaultMaxRunSeconds) * time.Second
	if cur.MaxRunSeconds != nil {
		maxRun = time.Duration(cur.GetMaxRunSeconds()) * time.Second
	}
	return &Bridge{
		factory:     factory,
		workspaceID: pb.GetWorkspaceId(),
		agentID:     pb.GetAgentId(),
		butterboxID: strings.TrimSpace(cur.GetButterboxId()),
		workingDir:  strings.TrimSpace(cur.GetWorkingDir()),
		model:       strings.TrimSpace(cur.GetModel()),
		mode:        strings.TrimSpace(cur.GetMode()),
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

// run executes one turn: ensure a Cursor session, send the prompt, wait for
// the result, and yield the assistant text.
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

		resp, err := b.sendMessage(runCtx, client, bnd.CursorSessionID, input, images)
		if err != nil && isNotFound(err) && !created {
			// The box lost the session (restart, cleanup). Recreate once.
			bnd, err = b.createSession(runCtx, ictx, client)
			if err != nil {
				yield(nil, b.classifyInterruption(ictx, runCtx, err))
				return
			}
			if !b.yieldBinding(ictx, bnd, yield) {
				return
			}
			resp, err = b.sendMessage(runCtx, client, bnd.CursorSessionID, input, images)
		}
		if err != nil {
			if runCtx.Err() != nil {
				b.abort(client, bnd.CursorSessionID)
			}
			if isNotFound(err) {
				err = fmt.Errorf("cursorbox: the box lost a freshly created cursor session; the box looks unhealthy — check it before retrying: %w", err)
			}
			yield(nil, b.classifyInterruption(ictx, runCtx, err))
			return
		}

		evt := session.NewEvent(ictx, ictx.InvocationID())
		evt.Author = ictx.Agent().Name()
		evt.Content = genai.NewContentFromText(resp.Text, genai.RoleModel)
		yield(evt, nil)
	}
}

func (b *Bridge) ensureSession(runCtx context.Context, ictx agent.InvocationContext, client CursorClient) (binding, bool, error) {
	if bnd, ok := readBinding(ictx.Session().State(), b.agentID); ok && bnd.matches(b.butterboxID, b.workingDir) {
		return bnd, false, nil
	}
	bnd, err := b.createSession(runCtx, ictx, client)
	if err != nil {
		return binding{}, false, err
	}
	return bnd, true, nil
}

func (b *Bridge) createSession(runCtx context.Context, ictx agent.InvocationContext, client CursorClient) (binding, error) {
	callCtx, cancel := context.WithTimeout(runCtx, controlCallTimeout)
	defer cancel()
	resp, err := client.CreateSession(callCtx, &CreateSessionRequest{
		Name:       fmt.Sprintf("butter:%s:%s", b.agentID, ictx.Session().ID()),
		WorkingDir: b.workingDir,
		Model:      b.model,
		Mode:       b.mode,
	})
	if err != nil {
		return binding{}, b.actionable("create cursor session", err)
	}
	id := resp.SessionID
	if id == "" {
		return binding{}, fmt.Errorf("cursorbox: the box answered CreateSession without a session id")
	}
	return binding{CursorSessionID: id, ButterboxID: b.butterboxID, WorkingDir: b.workingDir}, nil
}

func (b *Bridge) yieldBinding(ictx agent.InvocationContext, bnd binding, yield func(*session.Event, error) bool) bool {
	evt := session.NewEvent(ictx, ictx.InvocationID())
	evt.Author = ictx.Agent().Name()
	evt.Actions.StateDelta[stateKey(b.agentID)] = bnd.stateValue()
	return yield(evt, nil)
}

func (b *Bridge) sendMessage(runCtx context.Context, client CursorClient, sessionID, input string, images []ImageContent) (*SendMessageResponse, error) {
	resp, err := client.SendMessage(runCtx, &SendMessageRequest{
		SessionID: sessionID,
		Message:   input,
		Images:    images,
	})
	if err != nil {
		if isNotFound(err) {
			return nil, err
		}
		return nil, b.actionable("send message", err)
	}
	return resp, nil
}

func (b *Bridge) abort(client CursorClient, sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), abortTimeout)
	defer cancel()
	_ = client.AbortSession(ctx, &AbortSessionRequest{SessionID: sessionID})
}

func (b *Bridge) classifyInterruption(ictx agent.InvocationContext, runCtx context.Context, err error) error {
	if cerr := ictx.Err(); cerr != nil {
		return cerr
	}
	if runCtx.Err() != nil {
		return fmt.Errorf("cursorbox: the run exceeded max_run_seconds=%d and was aborted on the box; raise max_run_seconds on the agent for long runs", int(b.maxRun/time.Second))
	}
	return err
}

func (b *Bridge) actionable(op string, err error) error {
	switch connect.CodeOf(err) {
	case connect.CodeResourceExhausted:
		return fmt.Errorf("cursorbox: the ButterBox is at its session capacity; retry later or raise CURSOR_MAX_SESSIONS on the box: %w", err)
	case connect.CodeFailedPrecondition:
		return fmt.Errorf("cursorbox: the cursor session is busy with another message; wait for it to finish or cancel it: %w", err)
	case connect.CodeUnauthenticated, connect.CodePermissionDenied:
		return fmt.Errorf("cursorbox: the ButterBox rejected the access token; rotate it via SetButterBoxToken: %w", err)
	default:
		return fmt.Errorf("cursorbox: %s: %w", op, err)
	}
}

func isNotFound(err error) bool {
	return connect.CodeOf(err) == connect.CodeNotFound
}

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

func extractImages(c *genai.Content) []ImageContent {
	if c == nil {
		return nil
	}
	var images []ImageContent
	for _, p := range c.Parts {
		if p.InlineData == nil || len(p.InlineData.Data) == 0 {
			continue
		}
		if !strings.HasPrefix(p.InlineData.MIMEType, "image/") {
			continue
		}
		images = append(images, ImageContent{
			MIMEType: p.InlineData.MIMEType,
			Data:     p.InlineData.Data,
		})
	}
	return images
}
