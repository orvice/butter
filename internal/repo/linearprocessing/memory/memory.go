// Package memory implements linearprocessing.Repository in memory.
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

	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type entry struct {
	record         *agentsv1.LinearProcessingRecord
	leaseToken     string
	leaseExpiresAt time.Time
}

// Store implements linearprocessing.Repository in memory.
type Store struct {
	mu         sync.RWMutex
	records    map[string]*entry
	byDelivery map[string]string
}

var _ linearprocessing.Repository = (*Store)(nil)

func New() *Store {
	return &Store{records: map[string]*entry{}, byDelivery: map[string]string{}}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func deliveryKey(appID, deliveryID string) string { return appID + "\x00" + deliveryID }

func clone(r *agentsv1.LinearProcessingRecord) *agentsv1.LinearProcessingRecord {
	return proto.Clone(r).(*agentsv1.LinearProcessingRecord)
}

func (s *Store) Claim(_ context.Context, record *agentsv1.LinearProcessingRecord, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.LinearProcessingRecord, linearprocessing.ClaimAction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := deliveryKey(record.GetAppId(), record.GetDeliveryId())
	if id, ok := s.byDelivery[key]; ok {
		stored := s.records[id]
		if stored.leaseToken != "" && stored.leaseExpiresAt.After(claimedAt) {
			return clone(stored.record), linearprocessing.ClaimAcknowledge, linearprocessing.ErrInProgress
		}
		action := linearprocessing.RecoveryAction(stored.record)
		linearprocessing.MarkInterruptedUncertain(stored.record)
		if action != linearprocessing.ClaimAcknowledge {
			stored.record.Attempts++
			stored.leaseToken = leaseToken
			stored.leaseExpiresAt = leaseExpiresAt
		} else {
			stored.leaseToken = ""
			stored.leaseExpiresAt = time.Time{}
		}
		stored.record.UpdatedAt = timestamppb.New(claimedAt)
		return clone(stored.record), action, nil
	}
	stored := clone(record)
	if stored.GetId() == "" {
		stored.Id = uuid.NewString()
	}
	if stored.GetInvocationId() == "" {
		stored.InvocationId = uuid.NewString()
	}
	stored.Attempts = 1
	stored.CreatedAt = timestamppb.New(claimedAt)
	stored.UpdatedAt = timestamppb.New(claimedAt)
	stored.ExpiresAt = timestamppb.New(claimedAt.Add(linearprocessing.RetentionPeriod))
	s.records[stored.GetId()] = &entry{record: stored, leaseToken: leaseToken, leaseExpiresAt: leaseExpiresAt}
	s.byDelivery[key] = stored.GetId()
	return clone(stored), linearprocessing.ClaimRunAgent, nil
}

func (s *Store) ClaimForResend(_ context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.LinearProcessingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.records[id]
	if !ok || stored.record.GetWorkspaceId() != workspaceID {
		return nil, fmt.Errorf("linear processing record %q: %w", id, linearprocessing.ErrNotFound)
	}
	if stored.leaseToken != "" && stored.leaseExpiresAt.After(claimedAt) {
		return nil, linearprocessing.ErrInProgress
	}
	stored.record.Attempts++
	stored.record.UpdatedAt = timestamppb.New(claimedAt)
	stored.leaseToken = leaseToken
	stored.leaseExpiresAt = leaseExpiresAt
	return clone(stored.record), nil
}

func (s *Store) replace(existing *entry, record *agentsv1.LinearProcessingRecord) *agentsv1.LinearProcessingRecord {
	stored := clone(record)
	stored.CreatedAt = existing.record.GetCreatedAt()
	stored.ExpiresAt = existing.record.GetExpiresAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	existing.record = stored
	return clone(stored)
}

func (s *Store) UpdateClaimed(_ context.Context, record *agentsv1.LinearProcessingRecord, leaseToken string) (*agentsv1.LinearProcessingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.records[record.GetId()]
	if !ok {
		return nil, fmt.Errorf("linear processing record %q: %w", record.GetId(), linearprocessing.ErrNotFound)
	}
	if existing.leaseToken != leaseToken {
		return nil, linearprocessing.ErrLeaseLost
	}
	return s.replace(existing, record), nil
}

func (s *Store) Update(_ context.Context, record *agentsv1.LinearProcessingRecord) (*agentsv1.LinearProcessingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.records[record.GetId()]
	if !ok {
		return nil, fmt.Errorf("linear processing record %q: %w", record.GetId(), linearprocessing.ErrNotFound)
	}
	return s.replace(existing, record), nil
}

func (s *Store) lease(workspaceID, id, leaseToken string) (*entry, error) {
	stored, ok := s.records[id]
	if !ok || stored.record.GetWorkspaceId() != workspaceID {
		return nil, linearprocessing.ErrNotFound
	}
	if stored.leaseToken != leaseToken {
		return nil, linearprocessing.ErrLeaseLost
	}
	return stored, nil
}

func (s *Store) RenewClaim(_ context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, err := s.lease(workspaceID, id, leaseToken)
	if err != nil {
		return err
	}
	stored.leaseExpiresAt = leaseExpiresAt
	return nil
}

func (s *Store) ReleaseClaim(_ context.Context, workspaceID, id, leaseToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, err := s.lease(workspaceID, id, leaseToken)
	if err != nil {
		return err
	}
	stored.leaseToken = ""
	stored.leaseExpiresAt = time.Time{}
	return nil
}

func (s *Store) Get(_ context.Context, workspaceID, id string) (*agentsv1.LinearProcessingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stored, ok := s.records[id]
	if !ok || stored.record.GetWorkspaceId() != workspaceID {
		return nil, fmt.Errorf("linear processing record %q: %w", id, linearprocessing.ErrNotFound)
	}
	return clone(stored.record), nil
}

func (s *Store) List(_ context.Context, filter linearprocessing.Filter) ([]*agentsv1.LinearProcessingRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.LinearProcessingRecord
	for _, stored := range s.records {
		r := stored.record
		if r.GetWorkspaceId() != filter.WorkspaceID {
			continue
		}
		if filter.AppID != "" && r.GetAppId() != filter.AppID {
			continue
		}
		if filter.Status != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_UNSPECIFIED && r.GetStatus() != filter.Status {
			continue
		}
		out = append(out, clone(r))
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].GetCreatedAt().AsTime().After(out[j].GetCreatedAt().AsTime())
	})
	if filter.Limit > 0 && len(out) > filter.Limit {
		out = out[:filter.Limit]
	}
	return out, nil
}
