// Package memory implements telegramprocessing.Repository in memory.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/telegramprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store implements telegramprocessing.Repository in memory.
type Store struct {
	mu      sync.RWMutex
	records map[string]*agentsv1.TelegramProcessingRecord
	// byUpdate indexes (channel, update) so a duplicate delivery finds the
	// existing record rather than creating a second one.
	byUpdate map[string]string
}

var _ telegramprocessing.Repository = (*Store)(nil)

func New() *Store {
	return &Store{
		records:  make(map[string]*agentsv1.TelegramProcessingRecord),
		byUpdate: make(map[string]string),
	}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func updateKey(channelID string, updateID int64) string {
	return fmt.Sprintf("%s:%d", channelID, updateID)
}

func (s *Store) Claim(_ context.Context, record *agentsv1.TelegramProcessingRecord) (*agentsv1.TelegramProcessingRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	key := updateKey(record.GetChannelId(), record.GetUpdateId())

	if id, ok := s.byUpdate[key]; ok {
		existing := s.records[id]
		if existing.GetStatus() == agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_SUCCEEDED {
			// Already completed: the caller acknowledges without re-running.
			return proto.Clone(existing).(*agentsv1.TelegramProcessingRecord), false, nil
		}
		existing.Attempts++
		existing.UpdatedAt = timestamppb.New(now)
		return proto.Clone(existing).(*agentsv1.TelegramProcessingRecord), true, nil
	}

	stored := proto.Clone(record).(*agentsv1.TelegramProcessingRecord)
	if stored.GetId() == "" {
		stored.Id = uuid.NewString()
	}
	if stored.GetInvocationId() == "" {
		stored.InvocationId = uuid.NewString()
	}
	stored.Attempts = 1
	stored.CreatedAt = timestamppb.New(now)
	stored.UpdatedAt = timestamppb.New(now)
	stored.ExpiresAt = timestamppb.New(now.Add(telegramprocessing.RetentionPeriod))
	s.records[stored.GetId()] = stored
	s.byUpdate[key] = stored.GetId()
	return proto.Clone(stored).(*agentsv1.TelegramProcessingRecord), true, nil
}

func (s *Store) Update(_ context.Context, record *agentsv1.TelegramProcessingRecord) (*agentsv1.TelegramProcessingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.records[record.GetId()]
	if !ok {
		return nil, fmt.Errorf("telegram processing record %q: %w", record.GetId(), telegramprocessing.ErrNotFound)
	}
	stored := proto.Clone(record).(*agentsv1.TelegramProcessingRecord)
	stored.CreatedAt = existing.GetCreatedAt()
	stored.ExpiresAt = existing.GetExpiresAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	s.records[stored.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.TelegramProcessingRecord), nil
}

func (s *Store) Get(_ context.Context, workspaceID, id string) (*agentsv1.TelegramProcessingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[id]
	if !ok || record.GetWorkspaceId() != workspaceID {
		return nil, fmt.Errorf("telegram processing record %q: %w", id, telegramprocessing.ErrNotFound)
	}
	return proto.Clone(record).(*agentsv1.TelegramProcessingRecord), nil
}

func (s *Store) List(_ context.Context, filter telegramprocessing.Filter) ([]*agentsv1.TelegramProcessingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.TelegramProcessingRecord
	for _, record := range s.records {
		if record.GetWorkspaceId() != filter.WorkspaceID {
			continue
		}
		if filter.ChannelID != "" && record.GetChannelId() != filter.ChannelID {
			continue
		}
		if filter.DestinationID != "" && record.GetDestinationId() != filter.DestinationID {
			continue
		}
		if filter.Status != agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_UNSPECIFIED &&
			record.GetStatus() != filter.Status {
			continue
		}
		out = append(out, proto.Clone(record).(*agentsv1.TelegramProcessingRecord))
	}
	// Newest first: an operator investigating an incident looks at recent
	// work.
	sort.Slice(out, func(i, j int) bool {
		return out[i].GetCreatedAt().AsTime().After(out[j].GetCreatedAt().AsTime())
	})
	if filter.Limit > 0 && len(out) > filter.Limit {
		out = out[:filter.Limit]
	}
	return out, nil
}
