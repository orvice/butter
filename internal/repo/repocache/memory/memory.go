// Package memory provides an in-memory implementation of the repocache
// repository (issue #215). Suitable for tests and single-instance
// development; production deployments should use the Mongo implementation.
package memory

import (
	"context"
	"path"
	"strings"
	"sync"

	"github.com/google/uuid"
	"go.orx.me/apps/butter/internal/repo/repocache"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

type snapshot struct {
	metadata repocache.SnapshotMetadata
	entries  []*agentsv1.RepoCacheEntry
	blobs    map[string][]byte
}

// Store is a thread-safe in-memory repocache.Repository.
type Store struct {
	mu   sync.RWMutex
	data map[string]*snapshot // keyed by workspace ID
}

// New creates a new empty in-memory store.
func New() *Store {
	return &Store{data: make(map[string]*snapshot)}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) PutSnapshot(_ context.Context, workspaceID string, metadata repocache.SnapshotMetadata, entries []*agentsv1.RepoCacheEntry, blobs []repocache.CachedBlob) error {
	metadata.SnapshotID = uuid.NewString()
	cloned := make([]*agentsv1.RepoCacheEntry, len(entries))
	for i, e := range entries {
		cloned[i] = proto.Clone(e).(*agentsv1.RepoCacheEntry)
	}
	blobMap := make(map[string][]byte, len(blobs))
	for _, b := range blobs {
		cp := make([]byte, len(b.Content))
		copy(cp, b.Content)
		blobMap[b.Path] = cp
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[workspaceID] = &snapshot{
		metadata: metadata,
		entries:  cloned,
		blobs:    blobMap,
	}
	return nil
}

func (s *Store) GetMetadata(_ context.Context, workspaceID string) (repocache.SnapshotMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.data[workspaceID]
	if !ok {
		return repocache.SnapshotMetadata{}, repocache.ErrNotFound
	}
	return snap.metadata, nil
}

func (s *Store) ListEntries(_ context.Context, workspaceID, snapshotID, dirPath string) ([]*agentsv1.RepoCacheEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.data[workspaceID]
	if !ok || snap.metadata.SnapshotID != snapshotID {
		return nil, repocache.ErrNotFound
	}
	dirPath = strings.TrimRight(dirPath, "/")
	var out []*agentsv1.RepoCacheEntry
	for _, e := range snap.entries {
		parent := path.Dir(e.GetPath())
		if parent == "." {
			parent = ""
		}
		if parent == dirPath {
			out = append(out, proto.Clone(e).(*agentsv1.RepoCacheEntry))
		}
	}
	return out, nil
}

func (s *Store) GetEntry(_ context.Context, workspaceID, snapshotID, entryPath string) (*agentsv1.RepoCacheEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.data[workspaceID]
	if !ok || snap.metadata.SnapshotID != snapshotID {
		return nil, repocache.ErrNotFound
	}
	for _, e := range snap.entries {
		if e.GetPath() == entryPath {
			return proto.Clone(e).(*agentsv1.RepoCacheEntry), nil
		}
	}
	return nil, repocache.ErrNotFound
}

func (s *Store) GetBlob(_ context.Context, workspaceID, snapshotID, filePath string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.data[workspaceID]
	if !ok || snap.metadata.SnapshotID != snapshotID {
		return nil, repocache.ErrNotFound
	}
	content, ok := snap.blobs[filePath]
	if !ok {
		return nil, repocache.ErrNotFound
	}
	cp := make([]byte, len(content))
	copy(cp, content)
	return cp, nil
}

func (s *Store) Delete(_ context.Context, workspaceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, workspaceID)
	return nil
}
