package agentfile

import (
	"context"
	"fmt"
	"sync"
)

type memoryContent struct {
	content     string
	contentType string
}

type memoryContentStore struct {
	mu    sync.RWMutex
	items map[string]memoryContent
}

func NewMemoryContentStore() ContentStore {
	return &memoryContentStore{items: make(map[string]memoryContent)}
}

func (s *memoryContentStore) Put(_ context.Context, key, content, contentType string) error {
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = memoryContent{content: content, contentType: contentType}
	return nil
}

func (s *memoryContentStore) Get(_ context.Context, key string) (string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[key]
	if !ok {
		return "", "", fmt.Errorf("agent file content %q: %w", key, ErrNotFound)
	}
	return item.content, item.contentType, nil
}

func (s *memoryContentStore) Delete(_ context.Context, keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range keys {
		delete(s.items, key)
	}
	return nil
}
