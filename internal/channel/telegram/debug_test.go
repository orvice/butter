package telegram

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestIsDebugActive_NilToggle_UsesChannelConfig(t *testing.T) {
	ctx := context.Background()

	cfg := &agentsv1.TelegramChannelConfig{Debug: true}
	if !IsDebugActive(ctx, nil, "ch", "sess", cfg) {
		t.Error("expected debug active when channel config debug=true and toggle is nil")
	}

	cfg = &agentsv1.TelegramChannelConfig{Debug: false}
	if IsDebugActive(ctx, nil, "ch", "sess", cfg) {
		t.Error("expected debug inactive when channel config debug=false and toggle is nil")
	}
}

func TestDebugToggle_Integration(t *testing.T) {
	// This test requires a running Redis instance.
	// Skip if not available.
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("redis not available, skipping integration test")
	}

	toggle := NewDebugToggle(rdb)
	channelName := "test-channel"
	sessionID := "test-session-debug"

	// Cleanup.
	defer rdb.Del(ctx, debugKey(channelName, sessionID))

	// No override set — should return nil.
	got, err := toggle.Get(ctx, channelName, sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil override, got: %v", *got)
	}

	// Toggle from channel default false → should become true.
	newState, err := toggle.Toggle(ctx, channelName, sessionID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newState {
		t.Error("expected toggle to true from default false")
	}

	// Toggle again → should become false.
	newState, err = toggle.Toggle(ctx, channelName, sessionID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState {
		t.Error("expected toggle to false")
	}

	// IsDebugActive should use Redis override (false) not channel config (true).
	cfg := &agentsv1.TelegramChannelConfig{Debug: true}
	if IsDebugActive(ctx, toggle, channelName, sessionID, cfg) {
		t.Error("expected debug inactive due to Redis override=false, despite channel config=true")
	}
}
