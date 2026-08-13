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
	records map[string]*entry
	// byUpdate indexes (channel, update) so a duplicate delivery finds the
	// existing record rather than creating a second one.
	byUpdate map[string]string
}

type entry struct {
	record         *agentsv1.TelegramProcessingRecord
	leaseToken     string
	leaseExpiresAt time.Time
}

var _ telegramprocessing.Repository = (*Store)(nil)

func New() *Store {
	return &Store{
		records:  make(map[string]*entry),
		byUpdate: make(map[string]string),
	}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func updateKey(channelID string, updateID int64) string {
	return fmt.Sprintf("%s:%d", channelID, updateID)
}

func (s *Store) Claim(_ context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.TelegramProcessingRecord, telegramprocessing.ClaimAction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := updateKey(record.GetChannelId(), record.GetUpdateId())

	if id, ok := s.byUpdate[key]; ok {
		stored := s.records[id]
		if stored.leaseToken != "" && stored.leaseExpiresAt.After(claimedAt) {
			return proto.Clone(stored.record).(*agentsv1.TelegramProcessingRecord), telegramprocessing.ClaimAcknowledge, telegramprocessing.ErrInProgress
		}
		existing := stored.record
		action := telegramprocessing.RecoveryAction(existing)
		telegramprocessing.MarkInterruptedUncertain(existing)
		if action != telegramprocessing.ClaimAcknowledge {
			existing.Attempts++
			stored.leaseToken = leaseToken
			stored.leaseExpiresAt = leaseExpiresAt
		} else {
			stored.leaseToken = ""
			stored.leaseExpiresAt = time.Time{}
		}
		existing.UpdatedAt = timestamppb.New(claimedAt)
		return proto.Clone(existing).(*agentsv1.TelegramProcessingRecord), action, nil
	}

	stored := proto.Clone(record).(*agentsv1.TelegramProcessingRecord)
	if stored.GetId() == "" {
		stored.Id = uuid.NewString()
	}
	if stored.GetInvocationId() == "" {
		stored.InvocationId = uuid.NewString()
	}
	stored.Attempts = 1
	stored.CreatedAt = timestamppb.New(claimedAt)
	stored.UpdatedAt = timestamppb.New(claimedAt)
	stored.ExpiresAt = timestamppb.New(claimedAt.Add(telegramprocessing.RetentionPeriod))
	s.records[stored.GetId()] = &entry{record: stored, leaseToken: leaseToken, leaseExpiresAt: leaseExpiresAt}
	s.byUpdate[key] = stored.GetId()
	return proto.Clone(stored).(*agentsv1.TelegramProcessingRecord), telegramprocessing.ClaimRunAgent, nil
}

func (s *Store) ClaimDelivery(_ context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.TelegramProcessingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.records[id]
	if !ok || stored.record.GetWorkspaceId() != workspaceID {
		return nil, fmt.Errorf("telegram processing record %q: %w", id, telegramprocessing.ErrNotFound)
	}
	if stored.leaseToken != "" && stored.leaseExpiresAt.After(claimedAt) {
		return proto.Clone(stored.record).(*agentsv1.TelegramProcessingRecord), telegramprocessing.ErrInProgress
	}
	stored.record.Attempts++
	stored.record.UpdatedAt = timestamppb.New(claimedAt)
	stored.leaseToken = leaseToken
	stored.leaseExpiresAt = leaseExpiresAt
	return proto.Clone(stored.record).(*agentsv1.TelegramProcessingRecord), nil
}

func (s *Store) UpdateClaimed(_ context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string) (*agentsv1.TelegramProcessingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.records[record.GetId()]
	if !ok {
		return nil, fmt.Errorf("telegram processing record %q: %w", record.GetId(), telegramprocessing.ErrNotFound)
	}
	if existing.leaseToken != leaseToken {
		return nil, telegramprocessing.ErrLeaseLost
	}
	stored := cloneForUpdate(record, existing.record)
	existing.record = stored
	return proto.Clone(stored).(*agentsv1.TelegramProcessingRecord), nil
}

func (s *Store) RenewClaim(_ context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.records[id]
	if !ok || stored.record.GetWorkspaceId() != workspaceID {
		return telegramprocessing.ErrNotFound
	}
	if stored.leaseToken != leaseToken {
		return telegramprocessing.ErrLeaseLost
	}
	stored.leaseExpiresAt = leaseExpiresAt
	return nil
}

func (s *Store) ReleaseClaim(_ context.Context, workspaceID, id, leaseToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.records[id]
	if !ok || stored.record.GetWorkspaceId() != workspaceID {
		return telegramprocessing.ErrNotFound
	}
	if stored.leaseToken != leaseToken {
		return telegramprocessing.ErrLeaseLost
	}
	stored.leaseToken = ""
	stored.leaseExpiresAt = time.Time{}
	return nil
}

func (s *Store) Update(_ context.Context, record *agentsv1.TelegramProcessingRecord) (*agentsv1.TelegramProcessingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.records[record.GetId()]
	if !ok {
		return nil, fmt.Errorf("telegram processing record %q: %w", record.GetId(), telegramprocessing.ErrNotFound)
	}
	stored := cloneForUpdate(record, existing.record)
	existing.record = stored
	return proto.Clone(stored).(*agentsv1.TelegramProcessingRecord), nil
}

func cloneForUpdate(record, existing *agentsv1.TelegramProcessingRecord) *agentsv1.TelegramProcessingRecord {
	stored := proto.Clone(record).(*agentsv1.TelegramProcessingRecord)
	stored.CreatedAt = existing.GetCreatedAt()
	stored.ExpiresAt = existing.GetExpiresAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	return stored
}

func (s *Store) Get(_ context.Context, workspaceID, id string) (*agentsv1.TelegramProcessingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stored, ok := s.records[id]
	if !ok || stored.record.GetWorkspaceId() != workspaceID {
		return nil, fmt.Errorf("telegram processing record %q: %w", id, telegramprocessing.ErrNotFound)
	}
	return proto.Clone(stored.record).(*agentsv1.TelegramProcessingRecord), nil
}

func (s *Store) List(_ context.Context, filter telegramprocessing.Filter) ([]*agentsv1.TelegramProcessingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.TelegramProcessingRecord
	for _, stored := range s.records {
		record := stored.record
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
