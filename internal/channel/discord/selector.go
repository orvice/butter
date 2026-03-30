package discord

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

const agentKeyPrefix = "butter:agent_sel:"

// AgentSelector stores per-session agent selection in Redis.
type AgentSelector struct {
	rdb *redis.Client
}

// NewAgentSelector creates a new Redis-backed agent selector.
func NewAgentSelector(rdb *redis.Client) *AgentSelector {
	return &AgentSelector{rdb: rdb}
}

func agentSelectorKey(channelName, sessionID string) string {
	return agentKeyPrefix + channelName + ":" + sessionID
}

// Get returns the selected agent name for a session.
func (s *AgentSelector) Get(ctx context.Context, channelName, sessionID string) (string, error) {
	val, err := s.rdb.Get(ctx, agentSelectorKey(channelName, sessionID)).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("redis get agent selection: %w", err)
	}
	return val, nil
}

// Set stores the selected agent name for a session.
func (s *AgentSelector) Set(ctx context.Context, channelName, sessionID, agentName string) error {
	if err := s.rdb.Set(ctx, agentSelectorKey(channelName, sessionID), agentName, 0).Err(); err != nil {
		return fmt.Errorf("redis set agent selection: %w", err)
	}
	return nil
}

const modelKeyPrefix = "butter:model_sel:"

// ModelSelector stores per-session model selection in Redis.
type ModelSelector struct {
	rdb *redis.Client
}

// NewModelSelector creates a new Redis-backed model selector.
func NewModelSelector(rdb *redis.Client) *ModelSelector {
	return &ModelSelector{rdb: rdb}
}

func modelSelectorKey(channelName, sessionID string) string {
	return modelKeyPrefix + channelName + ":" + sessionID
}

// Get returns the selected model alias for a session.
func (s *ModelSelector) Get(ctx context.Context, channelName, sessionID string) (string, error) {
	val, err := s.rdb.Get(ctx, modelSelectorKey(channelName, sessionID)).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("redis get model selection: %w", err)
	}
	return val, nil
}

// Set stores the selected model alias for a session.
func (s *ModelSelector) Set(ctx context.Context, channelName, sessionID, modelAlias string) error {
	if err := s.rdb.Set(ctx, modelSelectorKey(channelName, sessionID), modelAlias, 0).Err(); err != nil {
		return fmt.Errorf("redis set model selection: %w", err)
	}
	return nil
}

// parseAgentCommand parses "/agent <subcommand>" text.
// Returns (subcommand, arg). For "/agent list" → ("list", "").
// For "/agent foo" → ("switch", "foo").
func parseAgentCommand(text string) (subcommand, arg string) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/agent") {
		return "", ""
	}
	parts := strings.Fields(text)
	if len(parts) == 1 {
		return "list", ""
	}
	sub := parts[1]
	if sub == "list" {
		return "list", ""
	}
	return "switch", sub
}
