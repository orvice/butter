// Package memoryhook wires Memory Recall and Memory Capture into the agent
// runner (ADR-0013 §5, §6), following the hook model of mem0's coding-agent
// plugins without their protocol.
//
//   - Begin runs once per turn, before the ADK run: when the invocation's
//     root agent enables memory, it recalls Workspace and Agent Memory for
//     the user's message and carries the formatted block in the context.
//   - The injection plugin appends that block to the system instruction of
//     every model call in the invocation — including LLM sub-agents of a
//     composite or Workflow root — and never writes it into the session.
//   - Finish runs after a successful turn and captures it into Workspace
//     Memory in the background, off the reply path.
//
// Every mem0 failure degrades: recall injects nothing, capture is dropped,
// and the turn proceeds.
package memoryhook

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"butterfly.orx.me/core/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/mem0memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// DefaultRecallTimeout bounds recall, which sits on the reply path.
	DefaultRecallTimeout = 2 * time.Second
	// DefaultCaptureTimeout bounds one background capture: with inference
	// on, the mem0 OSS server answers only after its extraction LLM call.
	DefaultCaptureTimeout = 60 * time.Second
	// shortQueryRunes is the length below which the previous user message
	// is appended to the recall query, so replies like "ok, go on" still
	// recall context.
	shortQueryRunes = 20
)

var tracer = otel.Tracer("go.orx.me/apps/butter/internal/runtime/memoryhook")

// Turn is one turn's memory state, created by Begin and consumed by the
// injection plugin and Finish.
type Turn struct {
	scope     mem0memory.Scope
	config    *agentsv1.MemoryConfig
	userText  string
	fromEvent int
	block     string
}

// Block is the formatted recall block injected into model calls; empty
// when nothing was recalled.
func (t *Turn) Block() string {
	if t == nil {
		return ""
	}
	return t.block
}

type turnKey struct{}

func withTurn(ctx context.Context, t *Turn) context.Context {
	return context.WithValue(ctx, turnKey{}, t)
}

// TurnFromContext returns the turn Begin attached to the context.
func TurnFromContext(ctx context.Context) *Turn {
	t, _ := ctx.Value(turnKey{}).(*Turn)
	return t
}

// Hooks performs recall and capture against the mem0 memory service.
type Hooks struct {
	svc      *mem0memory.Service
	sessions session.Service

	RecallTimeout  time.Duration
	CaptureTimeout time.Duration

	captures sync.WaitGroup
}

func New(svc *mem0memory.Service, sessions session.Service) *Hooks {
	return &Hooks{
		svc:            svc,
		sessions:       sessions,
		RecallTimeout:  DefaultRecallTimeout,
		CaptureTimeout: DefaultCaptureTimeout,
	}
}

// Begin prepares memory for one turn. root is the invocation's root agent;
// only its MemoryConfig applies. prior is the session as loaded before the
// run (nil for a new session) and userParts the user's input as sent,
// before any workflow-resume rewrap. It returns ctx unchanged and a nil
// Turn when memory does not apply.
func (h *Hooks) Begin(ctx context.Context, root *agentsv1.Agent, info *agentsv1.ContextInfo, prior session.Session, userParts []*genai.Part) (context.Context, *Turn) {
	if h == nil || h.svc == nil || root == nil {
		return ctx, nil
	}
	mc := root.GetConfig().GetMemory()
	if !mc.GetEnabled() || root.GetWorkspaceId() == "" {
		return ctx, nil
	}
	t := &Turn{
		scope: mem0memory.Scope{
			WorkspaceID: root.GetWorkspaceId(),
			AgentID:     root.GetAgentId(),
			SessionID:   info.GetSessionId(),
			Channel:     channelOf(info),
			Principal:   info.GetUserId(),
		},
		config:   mc,
		userText: mem0memory.PartsText(userParts),
	}
	if prior != nil {
		t.fromEvent = prior.Events().Len()
	}
	if !mc.GetDisableAutoRecall() {
		t.block = h.recall(ctx, t, prior)
	}
	// The scope also serves callers that reach memory through ADK's
	// memory.Service interface during the run.
	ctx = mem0memory.WithScope(ctx, t.scope)
	return withTurn(ctx, t), t
}

func (h *Hooks) recall(ctx context.Context, t *Turn, prior session.Session) string {
	query := recallQuery(t.userText, prior)
	if query == "" {
		return ""
	}
	logger := log.FromContext(ctx)
	ctx, span := tracer.Start(ctx, "memory.recall")
	defer span.End()
	span.SetAttributes(
		attribute.String("butter.workspace_id", t.scope.WorkspaceID),
		attribute.String("butter.agent_id", t.scope.AgentID),
		attribute.String("butter.session_id", t.scope.SessionID),
	)

	rctx, cancel := context.WithTimeout(ctx, h.RecallTimeout)
	defer cancel()
	started := time.Now()
	recalled, err := h.svc.Recall(rctx, t.scope, query, searchOptions(t.config))
	latency := time.Since(started)
	if errors.Is(err, memoryconn.ErrNotConfigured) {
		logger.Debug("memory recall skipped: workspace memory is not configured")
		span.SetAttributes(attribute.Bool("butter.memory.configured", false))
		return ""
	}
	if err != nil {
		logger.Warn("memory recall degraded: continuing without memories", "err", err, "latency", latency)
		span.RecordError(err)
		span.SetStatus(codes.Error, "recall degraded")
		return ""
	}
	block := FormatBlock(recalled, MaxBlockChars)
	logger.Info("memory recalled", "hits", len(recalled), "latency", latency, "block_chars", utf8.RuneCountInString(block))
	span.SetAttributes(attribute.Int("butter.memory.hits", len(recalled)))
	return block
}

// Finish captures a successful turn into Workspace Memory in the
// background. invocationID is the ADK invocation the turn ran under.
func (h *Hooks) Finish(ctx context.Context, t *Turn, info *agentsv1.ContextInfo, invocationID string, runErr error) {
	if h == nil || h.svc == nil || h.sessions == nil || t == nil || runErr != nil || invocationID == "" {
		return
	}
	if t.config.GetDisableAutoCapture() {
		return
	}
	scope := t.scope
	scope.InvocationID = invocationID
	// Detached from the request: capture must survive the reply returning.
	bg := context.WithoutCancel(ctx)
	h.captures.Add(1)
	go func() {
		defer h.captures.Done()
		h.capture(bg, scope, t, info)
	}()
}

func (h *Hooks) capture(ctx context.Context, scope mem0memory.Scope, t *Turn, info *agentsv1.ContextInfo) {
	logger := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, h.CaptureTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "memory.capture")
	defer span.End()
	span.SetAttributes(
		attribute.String("butter.workspace_id", scope.WorkspaceID),
		attribute.String("butter.agent_id", scope.AgentID),
		attribute.String("butter.session_id", scope.SessionID),
		attribute.String("butter.invocation_id", scope.InvocationID),
	)

	resp, err := h.sessions.Get(ctx, &session.GetRequest{
		AppName:   info.GetChannelName(),
		UserID:    info.GetUserId(),
		SessionID: info.GetSessionId(),
	})
	if err != nil {
		logger.Warn("memory capture dropped: cannot load session", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "session load failed")
		return
	}
	started := time.Now()
	err = h.svc.CaptureTurn(ctx, scope, mem0memory.TurnInput{
		Session:   resp.Session,
		FromEvent: t.fromEvent,
		UserText:  t.userText,
	})
	if errors.Is(err, memoryconn.ErrNotConfigured) {
		logger.Debug("memory capture skipped: workspace memory is not configured")
		return
	}
	if err != nil {
		logger.Warn("memory capture dropped", "err", err, "latency", time.Since(started))
		span.RecordError(err)
		span.SetStatus(codes.Error, "capture dropped")
		return
	}
	logger.Info("memory captured", "invocation_id", scope.InvocationID, "latency", time.Since(started))
}

// Wait blocks until in-flight captures finish, for tests and shutdown.
func (h *Hooks) Wait() {
	if h != nil {
		h.captures.Wait()
	}
}

func searchOptions(mc *agentsv1.MemoryConfig) mem0memory.SearchOptions {
	opts := mem0memory.SearchOptions{}
	if mc.TopK != nil {
		opts.TopK = int(mc.GetTopK())
	}
	if mc.Threshold != nil {
		th := float64(mc.GetThreshold())
		opts.Threshold = &th
	}
	return opts
}

// recallQuery is the current user text, extended with the previous user
// message when the current one is too short to search on.
func recallQuery(userText string, prior session.Session) string {
	query := strings.TrimSpace(userText)
	if query == "" {
		return ""
	}
	if utf8.RuneCountInString(query) >= shortQueryRunes || prior == nil {
		return query
	}
	if prev := lastUserText(prior); prev != "" {
		return prev + "\n" + query
	}
	return query
}

func lastUserText(sess session.Session) string {
	events := sess.Events()
	for i := events.Len() - 1; i >= 0; i-- {
		ev := events.At(i)
		if ev == nil || ev.Author != "user" || ev.Content == nil {
			continue
		}
		var parts []string
		for _, p := range ev.Content.Parts {
			if p != nil && !p.Thought && strings.TrimSpace(p.Text) != "" {
				parts = append(parts, strings.TrimSpace(p.Text))
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n")
		}
	}
	return ""
}

// channelOf names the entry point for provenance: the specific channel type
// when the transport reports one, otherwise the ADK app name the entry
// point runs under (e.g. "web-chat", "cron:<job>").
func channelOf(info *agentsv1.ContextInfo) string {
	if ct := info.GetChannelType(); ct != "" {
		return ct
	}
	return info.GetChannelName()
}
