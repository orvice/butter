// Package memory implements telegramsetting.Repository in process memory.
package memory

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/telegramsetting"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store implements telegramsetting.Repository in memory.
type Store struct {
	mu       sync.RWMutex
	settings *agentsv1.TelegramSettings
}

var _ telegramsetting.Repository = (*Store)(nil)

func New() *Store {
	return &Store{settings: &agentsv1.TelegramSettings{}}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) Get(context.Context) (*agentsv1.TelegramSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return proto.Clone(s.settings).(*agentsv1.TelegramSettings), nil
}

func (s *Store) Put(_ context.Context, settings *agentsv1.TelegramSettings) (*agentsv1.TelegramSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := proto.Clone(settings).(*agentsv1.TelegramSettings)
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	s.settings = stored
	return proto.Clone(stored).(*agentsv1.TelegramSettings), nil
}
