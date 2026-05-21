package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	repo "go.orx.me/apps/butter/internal/repo/mcpoauth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestStoreWorkspaceIsolationAndDisconnect(t *testing.T) {
	store := New()
	ctx := context.Background()
	conn := &repo.Connection{
		WorkspaceID:    "ws-a",
		ServerID:       "srv",
		State:          agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_CONNECTED,
		EncryptedToken: "ciphertext",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	if err := store.Save(ctx, conn); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := store.Get(ctx, "ws-b", "srv"); !errors.Is(err, repo.ErrNotFound) {
		t.Fatalf("expected workspace-isolated not found, got %v", err)
	}
	got, err := store.Get(ctx, "ws-a", "srv")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got.EncryptedToken = "mutated"
	again, err := store.Get(ctx, "ws-a", "srv")
	if err != nil {
		t.Fatalf("get again: %v", err)
	}
	if again.EncryptedToken != "ciphertext" {
		t.Fatalf("store returned mutable connection")
	}
	if err := store.Delete(ctx, "ws-a", "srv"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(ctx, "ws-a", "srv"); !errors.Is(err, repo.ErrNotFound) {
		t.Fatalf("expected deleted connection not found, got %v", err)
	}
}

func TestStoreMarkState(t *testing.T) {
	store := New()
	ctx := context.Background()
	if err := store.Save(ctx, &repo.Connection{WorkspaceID: "ws", ServerID: "srv"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	now := time.Now().UTC()
	err := store.MarkState(ctx, "ws", "srv", agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED, "refresh failed", now)
	if err != nil {
		t.Fatalf("mark state: %v", err)
	}
	conn, err := store.Get(ctx, "ws", "srv")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if conn.State != agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED {
		t.Fatalf("unexpected state %v", conn.State)
	}
	if !conn.ReauthorizationRequired || conn.LastError != "refresh failed" {
		t.Fatalf("reauth fields not updated: %#v", conn)
	}
}
