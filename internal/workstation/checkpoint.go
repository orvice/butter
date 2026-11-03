package workstation

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCheckpointConflict   = errors.New("checkpoint revision conflict")
	ErrCheckpointRegression = errors.New("checkpoint sequence regressed")
)

// Checkpoint is the durable state needed to resume one provider endpoint. The
// cursor and ETag are opaque to the shared module; Generation changes when a
// scope, credential, provider profile or scan contract changes.
type Checkpoint struct {
	Revision   uint64
	Generation string
	Sequence   uint64
	Cursor     string
	ETag       string
	UpdatedAt  time.Time
}

// CheckpointStore is the persistence seam for restartable incremental scans.
// Implementations must make CompareAndSwap atomic and durable before returning.
type CheckpointStore interface {
	Load(context.Context, string) (Checkpoint, bool, error)
	CompareAndSwap(context.Context, string, uint64, Checkpoint) (Checkpoint, error)
}

type MemoryCheckpointStore struct {
	mu     sync.RWMutex
	values map[string]Checkpoint
}

func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{values: map[string]Checkpoint{}}
}

func (s *MemoryCheckpointStore) Load(ctx context.Context, key string) (Checkpoint, bool, error) {
	if err := ctx.Err(); err != nil {
		return Checkpoint{}, false, err
	}
	s.mu.RLock()
	value, ok := s.values[key]
	s.mu.RUnlock()
	return value, ok, nil
}

func (s *MemoryCheckpointStore) CompareAndSwap(ctx context.Context, key string, expected uint64, next Checkpoint) (Checkpoint, error) {
	if err := ctx.Err(); err != nil {
		return Checkpoint{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, exists := s.values[key]
	if (!exists && expected != 0) || (exists && current.Revision != expected) {
		return current, ErrCheckpointConflict
	}
	if exists && next.Generation == current.Generation && next.Sequence < current.Sequence {
		return current, ErrCheckpointRegression
	}
	next.Revision = expected + 1
	s.values[key] = next
	return next, nil
}
