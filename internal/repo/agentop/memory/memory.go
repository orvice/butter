// Package memory is an in-memory agentop.Repository for tests and unbound
// deployments.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentoprepo "go.orx.me/apps/butter/internal/repo/agentop"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store is an in-memory agentop.Repository.
type Store struct {
	mu   sync.RWMutex
	byID map[string]map[string]*entry
}

type entry struct {
	op             *agentsv1.AgentOperation
	leaseToken     string
	leaseExpiresAt time.Time
}

// New returns an empty in-memory operation store.
func New() *Store {
	return &Store{byID: make(map[string]map[string]*entry)}
}

func (r *Store) EnsureIndexes(context.Context) error { return nil }

func (r *Store) Create(_ context.Context, workspaceID string, op *agentsv1.AgentOperation) error {
	if op.GetWorkspaceId() != workspaceID {
		return fmt.Errorf("workspace mismatch: operation belongs to %q, create requested for %q", op.GetWorkspaceId(), workspaceID)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	bucket := r.byID[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*entry)
		r.byID[workspaceID] = bucket
	}
	if _, exists := bucket[op.GetId()]; exists {
		return agentoprepo.ErrAlreadyExists
	}
	bucket[op.GetId()] = &entry{op: proto.Clone(op).(*agentsv1.AgentOperation)}
	return nil
}

func (r *Store) Claim(_ context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.AgentOperation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.byID[workspaceID][id]
	if !ok {
		return nil, agentoprepo.ErrNotFound
	}
	switch record.op.GetStatus() {
	case agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED:
		return proto.Clone(record.op).(*agentsv1.AgentOperation), agentoprepo.ErrCompleted
	case agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING:
		activeUntil := record.leaseExpiresAt
		if activeUntil.IsZero() {
			if updatedAt := record.op.GetUpdatedAt(); updatedAt != nil {
				activeUntil = updatedAt.AsTime().Add(leaseExpiresAt.Sub(claimedAt))
			}
		}
		if activeUntil.After(claimedAt) {
			return proto.Clone(record.op).(*agentsv1.AgentOperation), agentoprepo.ErrInProgress
		}
	}
	claimed := proto.Clone(record.op).(*agentsv1.AgentOperation)
	claimed.AttemptCount++
	claimed.Status = agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING
	claimed.Error = ""
	claimed.UpdatedAt = timestamppb.New(claimedAt)
	record.op = claimed
	record.leaseToken = leaseToken
	record.leaseExpiresAt = leaseExpiresAt
	return proto.Clone(claimed).(*agentsv1.AgentOperation), nil
}

func (r *Store) RenewLease(_ context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.byID[workspaceID][id]
	if !ok {
		return agentoprepo.ErrNotFound
	}
	if record.leaseToken != leaseToken {
		return agentoprepo.ErrLeaseLost
	}
	record.leaseExpiresAt = leaseExpiresAt
	return nil
}

func (r *Store) SaveClaimed(_ context.Context, workspaceID, leaseToken string, op *agentsv1.AgentOperation) error {
	if op.GetWorkspaceId() != workspaceID {
		return fmt.Errorf("workspace mismatch: operation belongs to %q, save requested for %q", op.GetWorkspaceId(), workspaceID)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.byID[workspaceID][op.GetId()]
	if !ok {
		return agentoprepo.ErrNotFound
	}
	if record.leaseToken != leaseToken {
		return agentoprepo.ErrLeaseLost
	}
	record.op = proto.Clone(op).(*agentsv1.AgentOperation)
	return nil
}

func (r *Store) Get(_ context.Context, workspaceID, id string) (*agentsv1.AgentOperation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, ok := r.byID[workspaceID][id]
	if !ok {
		return nil, agentoprepo.ErrNotFound
	}
	return proto.Clone(record.op).(*agentsv1.AgentOperation), nil
}

func (r *Store) List(_ context.Context, workspaceID string, status agentsv1.AgentOperationStatus, pageSize int32, pageToken string) ([]*agentsv1.AgentOperation, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bucket := r.byID[workspaceID]
	out := make([]*agentsv1.AgentOperation, 0, len(bucket))
	for _, record := range bucket {
		op := record.op
		if status != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_UNSPECIFIED && op.GetStatus() != status {
			continue
		}
		out = append(out, proto.Clone(op).(*agentsv1.AgentOperation))
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti := out[i].GetCreatedAt().AsTime()
		tj := out[j].GetCreatedAt().AsTime()
		if ti.Equal(tj) {
			return out[i].GetId() > out[j].GetId()
		}
		return ti.After(tj)
	})
	page, next := paginate(out, pageSize, pageToken)
	return page, next, nil
}
func paginate(items []*agentsv1.AgentOperation, pageSize int32, pageToken string) ([]*agentsv1.AgentOperation, string) {
	size := int(agentoprepo.ClampPageSize(pageSize))
	offset := agentoprepo.DecodePageToken(pageToken)
	if offset >= len(items) {
		return nil, ""
	}
	end := offset + size
	if end > len(items) {
		end = len(items)
	}
	next := ""
	if end < len(items) {
		next = agentoprepo.EncodePageToken(end)
	}
	return items[offset:end], next
}
