package discord

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const debugKeyPrefix = "butter:debug:"

// DebugToggle manages per-session debug mode state in Redis.
type DebugToggle struct {
	rdb *redis.Client
}

// NewDebugToggle creates a new Redis-backed debug toggle.
func NewDebugToggle(rdb *redis.Client) *DebugToggle {
	return &DebugToggle{rdb: rdb}
}

func debugKey(channelName, sessionID string) string {
	return debugKeyPrefix + channelName + ":" + sessionID
}

// Get returns the per-session debug override.
func (d *DebugToggle) Get(ctx context.Context, channelName, sessionID string) (*bool, error) {
	val, err := d.rdb.Get(ctx, debugKey(channelName, sessionID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get debug toggle: %w", err)
	}
	v := val == "1"
	return &v, nil
}

// Set stores the per-session debug state.
func (d *DebugToggle) Set(ctx context.Context, channelName, sessionID string, enabled bool) error {
	val := "0"
	if enabled {
		val = "1"
	}
	if err := d.rdb.Set(ctx, debugKey(channelName, sessionID), val, 0).Err(); err != nil {
		return fmt.Errorf("redis set debug toggle: %w", err)
	}
	return nil
}

// Toggle flips the debug state for a session and returns the new state.
func (d *DebugToggle) Toggle(ctx context.Context, channelName, sessionID string, channelDefault bool) (bool, error) {
	current, err := d.Get(ctx, channelName, sessionID)
	if err != nil {
		return false, err
	}
	var newState bool
	if current != nil {
		newState = !*current
	} else {
		newState = !channelDefault
	}
	if err := d.Set(ctx, channelName, sessionID, newState); err != nil {
		return false, err
	}
	return newState, nil
}

// IsDebugActive checks if debug mode is active for a session.
func IsDebugActive(ctx context.Context, toggle *DebugToggle, channelName, sessionID string, discordCfg *agentsv1.DiscordChannelConfig) bool {
	if toggle != nil {
		override, err := toggle.Get(ctx, channelName, sessionID)
		if err == nil && override != nil {
			return *override
		}
	}
	return discordCfg.GetDebug()
}
