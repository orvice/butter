// Package memory implements linearsetting.Repository in process memory.
package memory

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/linearsetting"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store implements linearsetting.Repository in memory.
type Store struct {
	mu       sync.RWMutex
	settings *agentsv1.LinearSettings
}

var _ linearsetting.Repository = (*Store)(nil)

func New() *Store {
	return &Store{settings: &agentsv1.LinearSettings{}}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) Get(context.Context) (*agentsv1.LinearSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return proto.Clone(s.settings).(*agentsv1.LinearSettings), nil
}

func (s *Store) Put(_ context.Context, settings *agentsv1.LinearSettings) (*agentsv1.LinearSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := proto.Clone(settings).(*agentsv1.LinearSettings)
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	s.settings = stored
	return proto.Clone(stored).(*agentsv1.LinearSettings), nil
}
