package mem0memory

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.orx.me/apps/butter/internal/mem0"
)

// ErrMemoryNotFound means the memory does not exist or does not belong to
// the workspace; the two are indistinguishable to the caller on purpose.
var ErrMemoryNotFound = errors.New("memory not found")

// List returns up to limit memories of one scope, newest first. The mem0
// OSS listing has no pagination, so callers show a capped view and reach
// older memories through Search.
func (s *Service) List(ctx context.Context, scope Scope, target Target, limit int) ([]mem0.Memory, error) {
	userID, agentID, err := scope.identity(target)
	if err != nil {
		return nil, err
	}
	client, err := s.client(ctx, scope.WorkspaceID)
	if err != nil {
		return nil, err
	}
	found, err := client.List(ctx, mem0.ListRequest{UserID: userID, AgentID: agentID, TopK: limit})
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(found, func(a, b mem0.Memory) int { return cmp.Compare(b.CreatedAt, a.CreatedAt) })
	return found, nil
}

// Delete removes one memory of workspaceID. mem0's delete does not check
// ownership, so the memory is read first and must carry this workspace's
// Workspace Memory or Agent Memory identity; anything else is reported as
// ErrMemoryNotFound so IDs from other workspaces cannot be probed.
func (s *Service) Delete(ctx context.Context, workspaceID, memoryID string) error {
	if workspaceID == "" || strings.TrimSpace(memoryID) == "" {
		return ErrMemoryNotFound
	}
	client, err := s.client(ctx, workspaceID)
	if err != nil {
		return err
	}
	m, err := client.Get(ctx, memoryID)
	if err != nil {
		return fmt.Errorf("read memory: %w", err)
	}
	if m == nil {
		return ErrMemoryNotFound
	}
	if _, _, ok := OwnedBy(*m, workspaceID); !ok {
		return ErrMemoryNotFound
	}
	return client.Delete(ctx, memoryID)
}

// OwnedBy reports whether m belongs to workspaceID and, if so, which scope
// holds it and — for Agent Memory — the owning Agent ID.
func OwnedBy(m mem0.Memory, workspaceID string) (Target, string, bool) {
	if workspaceID == "" {
		return 0, "", false
	}
	if m.UserID == WorkspaceUserID(workspaceID) {
		return TargetWorkspace, "", true
	}
	prefix := AgentMemoryID(workspaceID, "")
	if agentID, ok := strings.CutPrefix(m.AgentID, prefix); ok && agentID != "" {
		return TargetAgent, agentID, true
	}
	return 0, "", false
}
