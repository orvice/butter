// Package pipeline holds the platform-agnostic channel message handling logic:
// allowlist admission, trigger matching, command handling, per-chat agent and
// model selection, message part assembly, runner invocation, and reply/debug
// delivery decisions. A platform poller shrinks to a transport adapter: it
// normalizes inbound updates into an IncomingMessage, implements the Transport
// interface for outbound I/O, and delegates all routing here. Telegram is the
// only adapter on this module today; Discord still has its own poller and is
// migrated separately.
package pipeline

import (
	"context"
	"time"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// IncomingMessage is a platform-normalized inbound message. The transport
// adapter is responsible for translating a raw platform update into this shape
// (including admission rules and a lazy BuildParts closure) before handing it
// to the Handler.
type IncomingMessage struct {
	// SessionID is the derived session identifier for this message. Session-ID
	// formatting is platform-specific, so the adapter computes it.
	SessionID string
	// UserID is the platform user identifier as a string.
	UserID string
	// ChatID is the platform chat/channel identifier as a string.
	ChatID string
	// MessageID identifies the inbound message, used by reply-mode delivery.
	MessageID string
	// Text is the message text (or photo caption).
	Text string
	// HasMedia reports whether the message carries media (e.g. an image) even
	// when Text is empty; used by the empty-message guard.
	HasMedia bool
	// IsPrivate reports whether the message came from a private (1:1) chat;
	// used by the PRIVATE_CHAT trigger.
	IsPrivate bool
	// ChatType is the normalized chat type recorded on ContextInfo.
	ChatType agentsv1.ChatType
	// ChannelType identifies the platform ("telegram", "discord") on ContextInfo.
	ChannelType string
	// Metadata is copied onto ContextInfo.Metadata (username, native ids, ...).
	Metadata map[string]string
	// Admission holds the normalized allowlist rules evaluated before routing.
	Admission []AdmissionRule
	// BuildParts lazily assembles the multimodal input parts (text + images).
	// It is only invoked on the plain-message path, keeping image download I/O
	// in the adapter and off the command paths.
	BuildParts func(ctx context.Context) ([]*genai.Part, error)
}

// AdmissionRule is one normalized allowlist dimension. The message passes the
// rule when Allowlist is empty (no restriction) or contains Value. When Value
// is empty and SkipWhenEmpty is set, the rule is skipped entirely (e.g. a
// Discord guild allowlist does not apply to direct messages).
type AdmissionRule struct {
	Value         string
	Allowlist     []string
	SkipWhenEmpty bool
}

// AgentChoice is one option in an agent selection list.
type AgentChoice struct {
	Name   string
	Active bool
}

// ModelChoice is one option in a model selection list.
type ModelChoice struct {
	Alias  string
	Active bool
}

// DebugSummary is a snapshot of debug-relevant activity observed during one
// agent turn. ToolCalls counts function-call attempts, while ToolCounts keeps
// the per-tool breakdown used by compact channel renderers.
type DebugSummary struct {
	ToolCalls        int
	ToolCounts       map[string]int
	Transfers        int
	Compactions      int
	LatestEvent      *session.Event
	LatestCompaction string
}

// StatusView is the assembled, platform-neutral data for a /status reply. The
// transport adapter renders it (Telegram markdown, Discord plain text, ...).
type StatusView struct {
	AgentStatus   *runner.AgentStatus
	ActiveAgent   string
	ActiveModel   string // the raw active model override, or "" for agent default
	ResolvedModel string // alias resolved to a concrete model name, if different
	AgentModel    string // the agent's default model (used when ActiveModel is "")
	SessionID     string
	HasSession    bool
	EventCount    int
	LastUpdate    time.Time
	SessionErr    error
	Now           time.Time
}

// Transport is the platform-specific outbound boundary. The Handler decides
// when to send and what data to send; the adapter decides how to render it.
type Transport interface {
	// SendReply delivers a plain text reply (agent responses and command
	// acknowledgements). Reply-mode/threading is the adapter's concern.
	SendReply(ctx context.Context, msg IncomingMessage, text string)
	// SendProcessing sends an initial "processing" placeholder message and
	// returns the platform message ID so it can be edited later. When debug is
	// non-nil, the placeholder includes the initial zero-valued debug summary.
	SendProcessing(ctx context.Context, msg IncomingMessage, agentName string, debug *DebugSummary) string
	// EditDebug refreshes a processing message with the latest aggregated debug
	// snapshot. It must edit messageID rather than append a new message.
	EditDebug(ctx context.Context, msg IncomingMessage, messageID string, agentName string, debug DebugSummary)
	// EditReply edits a previously sent message (identified by messageID) with
	// the final agent response. A non-nil debug snapshot is rendered as a compact
	// summary without the latest-event detail.
	EditReply(ctx context.Context, msg IncomingMessage, messageID string, agentName string, text string, debug *DebugSummary)
	// SendTyping signals that the agent is working, if the platform supports it.
	SendTyping(ctx context.Context, msg IncomingMessage)
	// SendDebugStatus reports the new debug on/off state (with any platform UI).
	SendDebugStatus(ctx context.Context, msg IncomingMessage, active bool)
	// SendAgentList renders the agent selection list with the active one flagged.
	SendAgentList(ctx context.Context, msg IncomingMessage, choices []AgentChoice)
	// SendModelList renders the model selection list with the active one flagged.
	SendModelList(ctx context.Context, msg IncomingMessage, choices []ModelChoice)
	// SendStatus renders the /status view.
	SendStatus(ctx context.Context, msg IncomingMessage, view StatusView)
}

// RunnerPort is the narrow slice of runner.Service the pipeline depends on.
// *runner.Service satisfies it structurally.
type RunnerPort interface {
	RunTurn(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (*runner.TurnResult, error)
	HasAgentInWorkspace(workspaceID, name string) bool
	GetAgentStatus(name string) *runner.AgentStatus
	GetAgentModel(name string) string
	ModelProviders() []agentsv1.ModelProvider
	GetSession(ctx context.Context, channelName, sessionID, userID string) (session.Session, error)
	ClearSession(ctx context.Context, channelName, sessionID, userID string) error
}

// SelectorPort stores a per-session string selection (agent name or model
// alias) in a backing store. The Redis-backed telegram selectors satisfy it.
type SelectorPort interface {
	Get(ctx context.Context, channelName, sessionID string) (string, error)
	Set(ctx context.Context, channelName, sessionID, value string) error
}

// DebugPort manages per-session debug overrides. The telegram DebugToggle
// satisfies it.
type DebugPort interface {
	Get(ctx context.Context, channelName, sessionID string) (*bool, error)
	Toggle(ctx context.Context, channelName, sessionID string, channelDefault bool) (bool, error)
}

// Config carries the platform-neutral settings extracted from an AgentChannel.
type Config struct {
	ChannelName  string
	WorkspaceID  string
	DefaultAgent string
	DefaultModel string
	ChannelType  string
	Triggers     []*agentsv1.AgentTrigger
	SendTyping   bool
	DebugDefault bool
	AgentNames   []string
	ModelNames   []string
}

// Handler routes a normalized IncomingMessage through the channel pipeline. It
// holds no per-message state; collaborators are injected for testability.
type Handler struct {
	cfg       Config
	runner    RunnerPort
	agentSel  SelectorPort
	modelSel  SelectorPort
	debug     DebugPort
	transport Transport
}

// NewHandler wires a Handler. modelSel may be nil (no model selection store),
// matching the existing pollers' tolerance for an absent model selector.
func NewHandler(cfg Config, r RunnerPort, agentSel, modelSel SelectorPort, debug DebugPort, transport Transport) *Handler {
	return &Handler{
		cfg:       cfg,
		runner:    r,
		agentSel:  agentSel,
		modelSel:  modelSel,
		debug:     debug,
		transport: transport,
	}
}
