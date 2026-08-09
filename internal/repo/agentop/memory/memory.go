// Package memory is an in-memory agentop.Repository for tests and unbound
// deployments.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"google.golang.org/protobuf/proto"

	agentoprepo "go.orx.me/apps/butter/internal/repo/agentop"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store is an in-memory agentop.Repository.
type Store struct {
	mu   sync.RWMutex
	byID map[string]*agentsv1.AgentOperation
}

// New returns an empty in-memory operation store.
func New() *Store {
	return &Store{byID: make(map[string]*agentsv1.AgentOperation)}
}

func (r *Store) EnsureIndexes(context.Context) error { return nil }

func (r *Store) Save(_ context.Context, workspaceID string, op *agentsv1.AgentOperation) error {
	if op.GetWorkspaceId() != workspaceID {
		return fmt.Errorf("workspace mismatch: operation belongs to %q, save requested for %q", op.GetWorkspaceId(), workspaceID)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.byID[op.GetId()]; ok && existing.GetWorkspaceId() != workspaceID {
		return agentoprepo.ErrNotFound
	}
	r.byID[op.GetId()] = proto.Clone(op).(*agentsv1.AgentOperation)
	return nil
}

func (r *Store) Get(_ context.Context, workspaceID, id string) (*agentsv1.AgentOperation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	op, ok := r.byID[id]
	if !ok || op.GetWorkspaceId() != workspaceID {
		return nil, agentoprepo.ErrNotFound
	}
	return proto.Clone(op).(*agentsv1.AgentOperation), nil
}

func (r *Store) List(_ context.Context, workspaceID string, status agentsv1.AgentOperationStatus, pageSize int32, pageToken string) ([]*agentsv1.AgentOperation, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.AgentOperation, 0, len(r.byID))
	for _, op := range r.byID {
		if op.GetWorkspaceId() != workspaceID {
			continue
		}
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

func (r *Store) ListResumableAcrossWorkspaces(_ context.Context) ([]*agentsv1.AgentOperation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.AgentOperation, 0)
	for _, op := range r.byID {
		switch op.GetStatus() {
		case agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING,
			agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED:
			out = append(out, proto.Clone(op).(*agentsv1.AgentOperation))
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetId() < out[j].GetId() })
	return out, nil
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
