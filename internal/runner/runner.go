package runner

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"butterfly.orx.me/core/log"
	"github.com/achetronic/adk-utils-go/plugin/contextguard"
	"google.golang.org/adk/agent"
	"google.golang.org/protobuf/proto"
	"google.golang.org/adk/memory"
	adkrunner "google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// AgentStatus holds a snapshot of an agent's configuration for display.
type AgentStatus struct {
	Name        string
	Description string
	MCPServers  []string
	SubAgents   []*AgentStatus
}

// Service manages an agent registry and per-channel ADK runners.
type Service struct {
	agents       map[string]agent.Agent
	agentsProto  map[string]*agentsv1.Agent // original proto configs keyed by name
	providers    []agentsv1.ModelProvider   // model providers for runtime resolution
	mcpRegistry  []agentsv1.MCPServer
	remoteAgents []agentsv1.RemoteAgent
	sessionSvc   session.Service
	memorySvc    memory.Service
	pluginConfig adkrunner.PluginConfig

	mu              sync.Mutex
	runners         map[string]*adkrunner.Runner // keyed by channel name
	overriddenCache map[string]agent.Agent       // keyed by "agentName:modelOverride"
}

// NewService builds the agent registry from proto configs.
func NewService(ctx context.Context, agents []agentsv1.Agent, providers []agentsv1.ModelProvider, mcpRegistry []agentsv1.MCPServer, remoteAgentRegistry []agentsv1.RemoteAgent, sessionSvc session.Service, memorySvc memory.Service, pluginConfig adkrunner.PluginConfig) (*Service, error) {
	logger := log.FromContext(ctx)
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

		a, err := internalagent.NewFromProto(ctx, &agents[i], providers, mcpRegistry, remoteAgentRegistry)
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
		agents:          registry,
		agentsProto:     protoRegistry,
		providers:       providers,
		mcpRegistry:     mcpRegistry,
		remoteAgents:    remoteAgentRegistry,
		sessionSvc:      sessionSvc,
		memorySvc:       memorySvc,
		pluginConfig:    pluginConfig,
		runners:         make(map[string]*adkrunner.Runner),
		overriddenCache: make(map[string]agent.Agent),
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


// AgentNames returns all registered agent names.
func (s *Service) AgentNames() []string {
	names := make([]string, 0, len(s.agents))
	for name := range s.agents {
		names = append(names, name)
	}
	return names
}

// HasAgent checks if an agent with the given name exists.
func (s *Service) HasAgent(name string) bool {
	_, ok := s.agents[name]
	return ok
}

// ModelProviders returns the configured model providers.
func (s *Service) ModelProviders() []agentsv1.ModelProvider {
	return s.providers
}

// GetAgentModel returns the model name configured for the named agent, or empty string.
func (s *Service) GetAgentModel(name string) string {
	pb, ok := s.agentsProto[name]
	if !ok {
		return ""
	}
	return pb.GetConfig().GetModel()
}

// GetAgentStatus returns a status tree for the named agent, or nil if not found.
func (s *Service) GetAgentStatus(name string) *AgentStatus {
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

	pb, ok := s.agentsProto[agentName]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %q", agentName)
	}

	// Resolve the model alias to get the actual model name.
	resolvedName, _ := internalagent.ResolveModelAlias(modelOverride, s.providers)

	// Clone the proto and override the model field.
	clone := proto.Clone(pb).(*agentsv1.Agent)
	clone.Config.Model = resolvedName

	a, err := internalagent.NewFromProto(ctx, clone, s.providers, s.mcpRegistry, s.remoteAgents)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.overriddenCache[cacheKey] = a
	s.mu.Unlock()

	return a, nil
}

// getOrCreateRunner returns a runner for the given channel and agent.
func (s *Service) getOrCreateRunner(ctx context.Context, channelName string, ag agent.Agent) (*adkrunner.Runner, error) {
	logger := log.FromContext(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()

	key := channelName
	if r, ok := s.runners[key]; ok {
		return r, nil
	}

	logger.Info("creating new ADK runner", "channel", channelName)

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
	logger.Info("ADK runner created", "channel", channelName)
	return r, nil
}

// EventCallback is called for each non-final event during agent execution.
// It receives the event and should not block for long.
type EventCallback func(evt *session.Event)

// Run executes an agent with the given context info and input text.
// modelOverride, if non-empty, overrides the agent's configured model (resolved by alias or name).
// If onEvent is non-nil, it is called for each non-final event.
// If onCompaction is non-nil, it is called when context compaction is detected.
func (s *Service) Run(ctx context.Context, agentName, input, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent EventCallback, onCompaction CompactionCallback) (string, error) {
	channelName := ctxInfo.GetChannelName()
	sessionID := ctxInfo.GetSessionId()
	userID := ctxInfo.GetUserId()

	logger := log.FromContext(ctx).With(
		"uuid", ctxInfo.GetUuid(),
		"channel", channelName,
		"agent", agentName,
		"session_id", sessionID,
		"user_id", userID,
		"source", ctxInfo.GetSource().String(),
	)
	ctx = log.WithLogger(ctx, logger)

	ag, ok := s.agents[agentName]
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

	r, err := s.getOrCreateRunner(ctx, channelName, ag)
	if err != nil {
		return "", err
	}

	logger.Debug("invoking ADK runner",
		"input_len", len(input),
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

	msg := genai.NewContentFromText(input, genai.RoleUser)

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
