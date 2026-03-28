package runner

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"butterfly.orx.me/core/log"
	"google.golang.org/adk/agent"
	adkrunner "google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Service manages an agent registry and per-channel ADK runners.
type Service struct {
	agents       map[string]agent.Agent
	sessionSvc   session.Service
	pluginConfig adkrunner.PluginConfig

	mu      sync.Mutex
	runners map[string]*adkrunner.Runner // keyed by channel name
}

// NewService builds the agent registry from proto configs.
func NewService(ctx context.Context, agents []agentsv1.Agent, providers []agentsv1.ModelProvider, sessionSvc session.Service, pluginConfig adkrunner.PluginConfig) (*Service, error) {
	logger := log.FromContext(ctx)
	registry := make(map[string]agent.Agent, len(agents))

	logger.Info("building agent registry", "agent_count", len(agents))

	for i := range agents {
		name := agents[i].GetName()
		logger.Info("building agent from proto",
			"agent", name,
			"type", agents[i].GetType().String(),
			"description", agents[i].GetDescription(),
		)

		a, err := internalagent.NewFromProto(ctx, &agents[i], providers)
		if err != nil {
			return nil, fmt.Errorf("building agent %q: %w", name, err)
		}
		registry[name] = a
		logger.Info("agent registered", "agent", name)
	}

	logger.Info("agent registry ready", "total_agents", len(registry))

	return &Service{
		agents:       registry,
		sessionSvc:   sessionSvc,
		pluginConfig: pluginConfig,
		runners:      make(map[string]*adkrunner.Runner),
	}, nil
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

// Run executes an agent for a given channel, session, and input text.
// If onEvent is non-nil, it is called for each non-final event.
func (s *Service) Run(ctx context.Context, channelName, agentName, sessionID, userID, input string, onEvent EventCallback) (string, error) {
	logger := log.FromContext(ctx)

	ag, ok := s.agents[agentName]
	if !ok {
		return "", fmt.Errorf("unknown agent: %q", agentName)
	}

	r, err := s.getOrCreateRunner(ctx, channelName, ag)
	if err != nil {
		return "", err
	}

	logger.Debug("invoking ADK runner",
		"channel", channelName,
		"agent", agentName,
		"session_id", sessionID,
		"user_id", userID,
		"input_len", len(input),
	)

	msg := genai.NewContentFromText(input, genai.RoleUser)

	var result strings.Builder
	eventCount := 0
	for evt, err := range r.Run(ctx, userID, sessionID, msg, agent.RunConfig{}) {
		if err != nil {
			logger.Error("ADK runner event error",
				"channel", channelName,
				"agent", agentName,
				"session_id", sessionID,
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
		"channel", channelName,
		"agent", agentName,
		"session_id", sessionID,
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
