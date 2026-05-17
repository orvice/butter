package runner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/achetronic/adk-utils-go/plugin/contextguard"
	"github.com/google/uuid"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/memory"
	adkrunner "google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// InvocationRecorder receives a per-Run record at start (status=RUNNING) and
// again at completion (status=SUCCEEDED/FAILED). Save must be idempotent /
// upsert by Invocation.Id.
type InvocationRecorder interface {
	Save(ctx context.Context, inv *agentsv1.Invocation) error
}

// AgentStatus holds a snapshot of an agent's configuration for display.
type AgentStatus struct {
	Name        string
	Description string
	MCPServers  []string
	SubAgents   []*AgentStatus
}

// AgentBuilderFunc is a function that builds an agent with a specific model.
// It is used for agents that cannot be rebuilt from proto (e.g., system agent).
type AgentBuilderFunc func(ctx context.Context, model string) (agent.Agent, error)

// Service manages an agent registry and per-channel ADK runners.
type Service struct {
	agents           map[string]agent.Agent
	agentsProto      map[string]*agentsv1.Agent  // original proto configs keyed by name
	agentBuilders    map[string]AgentBuilderFunc // dynamic builder funcs keyed by agent name
	providers        []agentsv1.ModelProvider    // model providers for runtime resolution
	mcpRegistry      []agentsv1.MCPServer
	remoteAgents     []agentsv1.RemoteAgent
	daemonRegistry   *daemon.Registry
	sessionSvc       session.Service
	memorySvc        memory.Service
	basePluginConfig adkrunner.PluginConfig
	pluginConfig     adkrunner.PluginConfig

	mu              sync.Mutex
	runners         map[string]*adkrunner.Runner // keyed by channel name
	overriddenCache map[string]agent.Agent       // keyed by "agentName:modelOverride"

	invRecorder InvocationRecorder

	cancelMu  sync.Mutex
	cancelers map[string]context.CancelFunc
}

// SetInvocationRecorder attaches a recorder that observes every Run call.
// Passing nil disables recording.
func (s *Service) SetInvocationRecorder(rec InvocationRecorder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invRecorder = rec
}

func (s *Service) recorder() InvocationRecorder {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.invRecorder
}

// CancelInvocation signals the in-flight invocation with the given id to stop.
// Returns true if the invocation was found and the cancel signal was delivered.
func (s *Service) CancelInvocation(id string) bool {
	s.cancelMu.Lock()
	cancel, ok := s.cancelers[id]
	s.cancelMu.Unlock()
	if !ok {
		return false
	}
	cancel()
	return true
}

func (s *Service) registerCancel(id string, cancel context.CancelFunc) {
	s.cancelMu.Lock()
	if s.cancelers == nil {
		s.cancelers = make(map[string]context.CancelFunc)
	}
	s.cancelers[id] = cancel
	s.cancelMu.Unlock()
}

func (s *Service) deregisterCancel(id string) {
	s.cancelMu.Lock()
	delete(s.cancelers, id)
	s.cancelMu.Unlock()
}

// NewService builds the agent registry from proto configs.
func NewService(ctx context.Context, agents []agentsv1.Agent, providers []agentsv1.ModelProvider, mcpRegistry []agentsv1.MCPServer, remoteAgentRegistry []agentsv1.RemoteAgent, daemonRegistry *daemon.Registry, sessionSvc session.Service, memorySvc memory.Service, pluginConfig adkrunner.PluginConfig) (*Service, error) {
	logger := log.FromContext(ctx)
	basePluginConfig := pluginConfig
	registry := make(map[string]agent.Agent, len(agents))
	protoRegistry := make(map[string]*agentsv1.Agent, len(agents))

	// Validate model alias uniqueness.
	if err := internalagent.ValidateModelAliases(providers); err != nil {
		return nil, fmt.Errorf("model config validation: %w", err)
	}

	logger.Info("building agent registry", "agent_count", len(agents))

	for i := range agents {
		name := agents[i].GetName()
		logger.Info("building agent from proto",
			"agent", name,
			"type", agents[i].GetType().String(),
			"description", agents[i].GetDescription(),
		)

		a, err := internalagent.NewFromProto(ctx, &agents[i], providers, mcpRegistry, remoteAgentRegistry, daemonRegistry)
		if err != nil {
			return nil, fmt.Errorf("building agent %q: %w", name, err)
		}
		registry[name] = a
		protoRegistry[name] = &agents[i]
		logger.Info("agent registered", "agent", name)
	}

	// Build contextguard plugin if any agent has context_guard config.
	guardPC, err := buildContextGuardPlugin(ctx, agents, providers)
	if err != nil {
		return nil, fmt.Errorf("building context guard plugin: %w", err)
	}
	pluginConfig = mergePluginConfigs(pluginConfig, guardPC)

	svc := &Service{
		agents:           registry,
		agentsProto:      protoRegistry,
		agentBuilders:    make(map[string]AgentBuilderFunc),
		providers:        providers,
		mcpRegistry:      mcpRegistry,
		remoteAgents:     remoteAgentRegistry,
		daemonRegistry:   daemonRegistry,
		sessionSvc:       sessionSvc,
		memorySvc:        memorySvc,
		basePluginConfig: basePluginConfig,
		pluginConfig:     pluginConfig,
		runners:          make(map[string]*adkrunner.Runner),
		overriddenCache:  make(map[string]agent.Agent),
	}

	// Add compaction notifier plugin (must be after contextguard).
	if len(guardPC.Plugins) > 0 {
		notifierPC := newCompactionNotifierPlugin()
		svc.pluginConfig = mergePluginConfigs(svc.pluginConfig, notifierPC)
	}

	logger.Info("agent registry ready", "total_agents", len(registry))

	return svc, nil
}

// buildContextGuardPlugin walks agent proto configs and builds a contextguard
// plugin for agents that have context_guard configured.
func buildContextGuardPlugin(ctx context.Context, agents []agentsv1.Agent, providers []agentsv1.ModelProvider) (adkrunner.PluginConfig, error) {
	logger := log.FromContext(ctx)

	// Collect all agents with context_guard config from the proto tree.
	type guardEntry struct {
		name      string
		modelName string
		cfg       *agentsv1.ContextGuardConfig
	}
	var entries []guardEntry
	var walk func(a *agentsv1.Agent)
	walk = func(a *agentsv1.Agent) {
		if cg := a.GetConfig().GetContextGuard(); cg != nil && cg.GetStrategy() != agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_UNSPECIFIED {
			if model := a.GetConfig().GetModel(); model != "" {
				entries = append(entries, guardEntry{
					name:      a.GetName(),
					modelName: model,
					cfg:       cg,
				})
			}
		}
		for _, sub := range a.GetSubAgents() {
			walk(sub)
		}
	}
	for i := range agents {
		walk(&agents[i])
	}

	if len(entries) == 0 {
		return adkrunner.PluginConfig{}, nil
	}

	registry := contextguard.NewCrushRegistry()
	guard := contextguard.New(registry)

	for _, e := range entries {
		m, err := internalagent.ResolveModel(ctx, e.modelName, providers)
		if err != nil {
			return adkrunner.PluginConfig{}, fmt.Errorf("resolving model %q for context guard on agent %q: %w", e.modelName, e.name, err)
		}

		var opts []contextguard.AgentOption
		switch e.cfg.GetStrategy() {
		case agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW:
			maxTurns := int(e.cfg.GetMaxTurns())
			if maxTurns <= 0 {
				maxTurns = 20
			}
			opts = append(opts, contextguard.WithSlidingWindow(maxTurns))
		}
		if e.cfg.GetMaxTokens() > 0 {
			opts = append(opts, contextguard.WithMaxTokens(int(e.cfg.GetMaxTokens())))
		}

		guard.Add(e.name, m, opts...)
		logger.Info("context guard configured",
			"agent", e.name,
			"strategy", e.cfg.GetStrategy().String(),
			"max_turns", e.cfg.GetMaxTurns(),
			"max_tokens", e.cfg.GetMaxTokens(),
		)
	}

	return guard.PluginConfig(), nil
}

// mergePluginConfigs combines two PluginConfigs by appending their plugin slices.
func mergePluginConfigs(a, b adkrunner.PluginConfig) adkrunner.PluginConfig {
	merged := a
	merged.Plugins = append(merged.Plugins, b.Plugins...)
	return merged
}

// RegisterAgent adds an agent to the registry. If an agent with the same name
// already exists, it is replaced and a warning is logged.
func (s *Service) RegisterAgent(name string, ag agent.Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[name] = ag
}

// RegisterAgentWithBuilder adds an agent with a builder function that can
// rebuild it with a different model. This is used for agents that are not
// proto-based (e.g., system agent) but still need to support model overrides.
func (s *Service) RegisterAgentWithBuilder(name string, ag agent.Agent, builder AgentBuilderFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[name] = ag
	s.agentBuilders[name] = builder
}

// ReloadProtoAgents rebuilds all proto-configured agents and refreshes runtime registries.
func (s *Service) ReloadProtoAgents(ctx context.Context, agents []agentsv1.Agent, providers []agentsv1.ModelProvider, mcpRegistry []agentsv1.MCPServer, remoteAgentRegistry []agentsv1.RemoteAgent) error {
	logger := log.FromContext(ctx)
	registry := make(map[string]agent.Agent, len(agents))
	protoRegistry := make(map[string]*agentsv1.Agent, len(agents))

	if err := internalagent.ValidateModelAliases(providers); err != nil {
		return fmt.Errorf("model config validation: %w", err)
	}

	for i := range agents {
		name := agents[i].GetName()
		a, err := internalagent.NewFromProto(ctx, &agents[i], providers, mcpRegistry, remoteAgentRegistry, s.daemonRegistry)
		if err != nil {
			return fmt.Errorf("rebuilding agent %q: %w", name, err)
		}
		registry[name] = a
		protoRegistry[name] = &agents[i]
	}

	guardPC, err := buildContextGuardPlugin(ctx, agents, providers)
	if err != nil {
		return fmt.Errorf("building context guard plugin: %w", err)
	}
	pluginConfig := mergePluginConfigs(s.basePluginConfig, guardPC)
	if len(guardPC.Plugins) > 0 {
		pluginConfig = mergePluginConfigs(pluginConfig, newCompactionNotifierPlugin())
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for name, ag := range s.agents {
		if _, isProto := protoRegistry[name]; isProto {
			continue
		}
		if _, hasBuilder := s.agentBuilders[name]; hasBuilder {
			registry[name] = ag
		}
	}

	s.agents = registry
	s.agentsProto = protoRegistry
	s.providers = providers
	s.mcpRegistry = mcpRegistry
	s.remoteAgents = remoteAgentRegistry
	s.pluginConfig = pluginConfig
	s.runners = make(map[string]*adkrunner.Runner)
	s.overriddenCache = make(map[string]agent.Agent)

	logger.Info("runner service reloaded", "proto_agents", len(protoRegistry), "total_agents", len(registry))
	return nil
}

// AgentNames returns all registered agent names.
func (s *Service) AgentNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.agents))
	for name := range s.agents {
		names = append(names, name)
	}
	return names
}

// HasAgent checks if an agent with the given name exists.
func (s *Service) HasAgent(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.agents[name]
	return ok
}

// ModelProviders returns the configured model providers.
func (s *Service) ModelProviders() []agentsv1.ModelProvider {
	return s.providers
}

// GetAgentModel returns the model name configured for the named agent, or empty string.
func (s *Service) GetAgentModel(name string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	pb, ok := s.agentsProto[name]
	if !ok {
		return ""
	}
	return pb.GetConfig().GetModel()
}

// GetAgentStatus returns a status tree for the named agent, or nil if not found.
func (s *Service) GetAgentStatus(name string) *AgentStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	pb, ok := s.agentsProto[name]
	if !ok {
		return nil
	}
	return buildAgentStatus(pb)
}

func buildAgentStatus(pb *agentsv1.Agent) *AgentStatus {
	st := &AgentStatus{
		Name:        pb.GetName(),
		Description: pb.GetDescription(),
	}
	for _, srv := range pb.GetConfig().GetMcpServers() {
		st.MCPServers = append(st.MCPServers, srv.GetName())
	}
	for _, sub := range pb.GetSubAgents() {
		st.SubAgents = append(st.SubAgents, buildAgentStatus(sub))
	}
	return st
}

// ClearSession deletes and recreates a session, effectively clearing its context.
func (s *Service) ClearSession(ctx context.Context, channelName, sessionID, userID string) error {
	// Delete the existing session (ignore not-found errors).
	_ = s.sessionSvc.Delete(ctx, &session.DeleteRequest{
		AppName:   channelName,
		UserID:    userID,
		SessionID: sessionID,
	})

	// Recreate an empty session with the same ID.
	_, err := s.sessionSvc.Create(ctx, &session.CreateRequest{
		AppName:   channelName,
		UserID:    userID,
		SessionID: sessionID,
	})
	return err
}

// GetSession retrieves the session for the given channel, session, and user.
func (s *Service) GetSession(ctx context.Context, channelName, sessionID, userID string) (session.Session, error) {
	resp, err := s.sessionSvc.Get(ctx, &session.GetRequest{
		AppName:   channelName,
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Session, nil
}

// buildOverriddenAgent creates (or returns cached) an agent with its model replaced.
func (s *Service) buildOverriddenAgent(ctx context.Context, agentName, modelOverride string) (agent.Agent, error) {
	cacheKey := agentName + ":" + modelOverride
	s.mu.Lock()
	if cached, ok := s.overriddenCache[cacheKey]; ok {
		s.mu.Unlock()
		return cached, nil
	}
	s.mu.Unlock()

	// Resolve the model alias to get the actual model name.
	resolvedName, _ := internalagent.ResolveModelAlias(modelOverride, s.providers)

	var a agent.Agent
	var err error

	s.mu.Lock()
	pb, hasProto := s.agentsProto[agentName]
	builder, hasBuilder := s.agentBuilders[agentName]
	mcpRegistry := s.mcpRegistry
	remoteAgents := s.remoteAgents
	s.mu.Unlock()

	if hasProto {
		// Proto-based agent: clone proto and override model.
		clone := proto.Clone(pb).(*agentsv1.Agent)
		clone.Config.Model = resolvedName
		a, err = internalagent.NewFromProto(ctx, clone, s.providers, mcpRegistry, remoteAgents, s.daemonRegistry)
	} else if hasBuilder {
		// Builder-based agent: rebuild with the resolved model.
		a, err = builder(ctx, resolvedName)
	} else {
		return nil, fmt.Errorf("agent %q has no proto config or builder for model override", agentName)
	}

	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.overriddenCache[cacheKey] = a
	s.mu.Unlock()

	return a, nil
}

// getOrCreateRunner returns a runner for the given channel, agent, and model override.
func (s *Service) getOrCreateRunner(ctx context.Context, channelName, agentName, modelOverride string, ag agent.Agent) (*adkrunner.Runner, error) {
	logger := log.FromContext(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()

	key := channelName + ":" + agentName + ":" + modelOverride
	if r, ok := s.runners[key]; ok {
		return r, nil
	}

	logger.Info("creating new ADK runner", "channel", channelName, "agent", agentName, "model_override", modelOverride)

	r, err := adkrunner.New(adkrunner.Config{
		AppName:        channelName,
		Agent:          ag,
		SessionService: s.sessionSvc,
		MemoryService:  s.memorySvc,
		PluginConfig:   s.pluginConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("creating runner for channel %q: %w", channelName, err)
	}

	s.runners[key] = r
	logger.Info("ADK runner created", "channel", channelName, "agent", agentName, "model_override", modelOverride)
	return r, nil
}

// EventCallback is called for each non-final event during agent execution.
// It receives the event and should not block for long.
type EventCallback func(evt *session.Event)

// Run executes an agent with the given context info and multimodal input parts.
// parts must contain at least one element (text and/or image parts).
// modelOverride, if non-empty, overrides the agent's configured model (resolved by alias or name).
// If onEvent is non-nil, it is called for each non-final event.
// If onCompaction is non-nil, it is called when context compaction is detected.
func (s *Service) Run(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent EventCallback, onCompaction CompactionCallback) (output string, runErr error) {
	channelName := ctxInfo.GetChannelName()
	sessionID := ctxInfo.GetSessionId()
	userID := ctxInfo.GetUserId()

	inv := s.startInvocation(ctx, agentName, parts, modelOverride, ctxInfo)
	defer s.finishInvocation(ctx, inv, &output, &runErr)
	if inv != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		s.registerCancel(inv.GetId(), cancel)
		defer func() {
			s.deregisterCancel(inv.GetId())
			cancel()
		}()
	}

	logger := log.FromContext(ctx).With(
		"uuid", ctxInfo.GetUuid(),
		"channel", channelName,
		"agent", agentName,
		"session_id", sessionID,
		"user_id", userID,
		"source", ctxInfo.GetSource().String(),
	)
	ctx = log.WithLogger(ctx, logger)

	if len(parts) == 0 {
		return "", fmt.Errorf("empty input: at least one part is required")
	}

	inputSummary := summarizeInputParts(parts)
	logger.Info("starting agent run",
		"parts_count", len(parts),
		"text_parts", inputSummary.textParts,
		"inline_data_parts", inputSummary.inlineDataParts,
		"file_data_parts", inputSummary.fileDataParts,
		"function_response_parts", inputSummary.functionResponseParts,
		"model_override", modelOverride,
	)

	s.mu.Lock()
	ag, ok := s.agents[agentName]
	s.mu.Unlock()
	if !ok {
		return "", fmt.Errorf("unknown agent: %q", agentName)
	}

	// If model override is set, rebuild the agent with the overridden model.
	if modelOverride != "" {
		logger.Info("applying model override", "model_override", modelOverride)
		overriddenAg, err := s.buildOverriddenAgent(ctx, agentName, modelOverride)
		if err != nil {
			return "", fmt.Errorf("building model-overridden agent: %w", err)
		}
		ag = overriddenAg
	}

	r, err := s.getOrCreateRunner(ctx, channelName, agentName, modelOverride, ag)
	if err != nil {
		return "", err
	}

	logger.Debug("invoking ADK runner",
		"parts_count", len(parts),
		"model_override", modelOverride,
	)

	// Ensure session exists; create one if not found.
	if _, err := s.sessionSvc.Get(ctx, &session.GetRequest{
		AppName:   channelName,
		UserID:    userID,
		SessionID: sessionID,
	}); err != nil {
		logger.Info("session not found, creating new session")
		if _, err := s.sessionSvc.Create(ctx, &session.CreateRequest{
			AppName:   channelName,
			UserID:    userID,
			SessionID: sessionID,
		}); err != nil {
			return "", fmt.Errorf("creating session: %w", err)
		}
	}

	// Store compaction callback in context for the notifier plugin.
	if onCompaction != nil {
		ctx = WithCompactionCallback(ctx, onCompaction)
	}

	msg := &genai.Content{Parts: parts, Role: genai.RoleUser}

	var result strings.Builder
	eventCount := 0
	for evt, err := range r.Run(ctx, userID, sessionID, msg, agent.RunConfig{}) {
		if err != nil {
			logger.Error("ADK runner event error",
				"event_count", eventCount,
				"err", err,
			)
			return result.String(), fmt.Errorf("runner error: %w", err)
		}
		eventCount++
		summary := summarizeEvent(evt)
		logger.Info("agent run event",
			"event_count", eventCount,
			"event_id", evt.ID,
			"invocation_id", evt.InvocationID,
			"author", evt.Author,
			"branch", evt.Branch,
			"partial", evt.Partial,
			"final_response", evt.IsFinalResponse(),
			"text_parts", summary.textParts,
			"function_calls", summary.functionCalls,
			"function_responses", summary.functionResponses,
			"code_execution_results", summary.codeExecutionResults,
			"long_running_tools", len(evt.LongRunningToolIDs),
			"transfer_to_agent", evt.Actions.TransferToAgent,
			"escalate", evt.Actions.Escalate,
			"state_delta_keys", summary.stateDeltaKeys,
			"artifact_delta_keys", summary.artifactDeltaKeys,
		)
		if onEvent != nil && !evt.IsFinalResponse() {
			onEvent(evt)
		}
		if evt.IsFinalResponse() && evt.Content != nil {
			for _, part := range evt.Content.Parts {
				if part.Text != "" {
					result.WriteString(part.Text)
				}
			}
		}
	}

	logger.Debug("ADK runner completed",
		"event_count", eventCount,
		"response_len", result.Len(),
	)

	return result.String(), nil
}

func (s *Service) startInvocation(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo) *agentsv1.Invocation {
	rec := s.recorder()
	if rec == nil {
		return nil
	}
	id := ctxInfo.GetUuid()
	if id == "" {
		id = uuid.NewString()
	}
	inv := &agentsv1.Invocation{
		Id:            id,
		AgentName:     agentName,
		AppName:       ctxInfo.GetChannelName(),
		UserId:        ctxInfo.GetUserId(),
		SessionId:     ctxInfo.GetSessionId(),
		Status:        agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		Input:         truncate(joinTextParts(parts), 4096),
		StartedAt:     timestamppb.New(time.Now().UTC()),
		ModelOverride: modelOverride,
		Source:        ctxInfo.GetSource().String(),
	}
	// Best-effort: failures are logged but do not block the run.
	if err := rec.Save(ctx, inv); err != nil {
		log.FromContext(ctx).Warn("failed to record invocation start", "err", err)
	}
	return inv
}

func (s *Service) finishInvocation(ctx context.Context, inv *agentsv1.Invocation, output *string, runErr *error) {
	if inv == nil {
		return
	}
	rec := s.recorder()
	if rec == nil {
		return
	}
	now := time.Now().UTC()
	inv.FinishedAt = timestamppb.New(now)
	inv.LatencyMs = now.Sub(inv.GetStartedAt().AsTime()).Milliseconds()
	if runErr != nil && *runErr != nil {
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
		inv.Error = truncate((*runErr).Error(), 4096)
	} else {
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED
	}
	if output != nil {
		inv.Output = truncate(*output, 4096)
	}
	if err := rec.Save(ctx, inv); err != nil {
		log.FromContext(ctx).Warn("failed to record invocation completion", "err", err)
	}
}

func joinTextParts(parts []*genai.Part) string {
	var b strings.Builder
	for _, p := range parts {
		if p == nil || p.Text == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(p.Text)
	}
	return b.String()
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

type inputPartSummary struct {
	textParts             int
	inlineDataParts       int
	fileDataParts         int
	functionResponseParts int
}

func summarizeInputParts(parts []*genai.Part) inputPartSummary {
	var summary inputPartSummary
	for _, part := range parts {
		if part == nil {
			continue
		}
		switch {
		case part.Text != "":
			summary.textParts++
		case part.InlineData != nil:
			summary.inlineDataParts++
		case part.FileData != nil:
			summary.fileDataParts++
		case part.FunctionResponse != nil:
			summary.functionResponseParts++
		}
	}
	return summary
}

type eventSummary struct {
	textParts            int
	functionCalls        int
	functionResponses    int
	codeExecutionResults int
	stateDeltaKeys       int
	artifactDeltaKeys    int
}

func summarizeEvent(evt *session.Event) eventSummary {
	if evt == nil {
		return eventSummary{}
	}

	summary := eventSummary{
		stateDeltaKeys:    len(evt.Actions.StateDelta),
		artifactDeltaKeys: len(evt.Actions.ArtifactDelta),
	}
	if evt.Content == nil {
		return summary
	}

	for _, part := range evt.Content.Parts {
		switch {
		case part.Text != "":
			summary.textParts++
		case part.FunctionCall != nil:
			summary.functionCalls++
		case part.FunctionResponse != nil:
			summary.functionResponses++
		case part.CodeExecutionResult != nil:
			summary.codeExecutionResults++
		}
	}

	return summary
}

// DeriveSessionID builds a session ID from Telegram update fields based on the session scope.
func DeriveSessionID(scope agentsv1.AgentSessionScope, chatID, userID int64) string {
	switch scope {
	case agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_USER:
		return fmt.Sprintf("user:%d", userID)
	case agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_CHAT:
		return fmt.Sprintf("chat:%d", chatID)
	default:
		return fmt.Sprintf("chat:%d", chatID)
	}
}
