package skill

import (
	"context"
	"fmt"
	"sync"
)

// ContentStore persists SKILL.md bodies outside the metadata repository
// (ADR 0004): metadata lives in Mongo, content behind this interface.
type ContentStore interface {
	Put(ctx context.Context, key, content string) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys []string) error
}

// ContentKey builds the storage key for a skill's SKILL.md body. The
// configured key_prefix is applied by the S3 content store, not here.
func ContentKey(workspaceID, name string) string {
	return workspaceID + "/" + name + "/SKILL.md"
}

// ResourceContentKey builds the storage key for a skill resource file.
// Resource paths live under references/, assets/, or scripts/, so they can
// never collide with the sibling SKILL.md key. The configured key_prefix is
// applied by the S3 content store, not here.
func ResourceContentKey(workspaceID, skillName, resourcePath string) string {
	return workspaceID + "/" + skillName + "/" + resourcePath
}

type memoryContentStore struct {
	mu    sync.RWMutex
	items map[string]string
}

// NewMemoryContentStore returns an in-memory ContentStore used when no S3
// bucket is configured, so local development needs zero infrastructure.
func NewMemoryContentStore() ContentStore {
	return &memoryContentStore{items: make(map[string]string)}
}

func (s *memoryContentStore) Put(_ context.Context, key, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = content
	return nil
}

func (s *memoryContentStore) Get(_ context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	content, ok := s.items[key]
	if !ok {
		return "", fmt.Errorf("skill content %q: %w", key, ErrNotFound)
	}
	return content, nil
}

func (s *memoryContentStore) Delete(_ context.Context, keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range keys {
		delete(s.items, key)
	}
	return nil
}
