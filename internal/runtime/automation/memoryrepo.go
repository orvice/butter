package automation

import (
	"context"
	"sort"
	"strconv"
	"sync"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

type MemoryDefinitionRepo struct {
	mu   sync.RWMutex
	byID map[string]*agentsv1.Automation
}

func NewMemoryDefinitionRepo() *MemoryDefinitionRepo {
	return &MemoryDefinitionRepo{byID: make(map[string]*agentsv1.Automation)}
}

func (r *MemoryDefinitionRepo) EnsureIndexes(context.Context) error { return nil }

func (r *MemoryDefinitionRepo) List(_ context.Context, workspaceID string) ([]*agentsv1.Automation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.Automation, 0)
	for _, a := range r.byID {
		if a.GetWorkspaceId() == workspaceID {
			out = append(out, proto.Clone(a).(*agentsv1.Automation))
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (r *MemoryDefinitionRepo) ListAll(_ context.Context) ([]*agentsv1.Automation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.Automation, 0, len(r.byID))
	for _, a := range r.byID {
		out = append(out, proto.Clone(a).(*agentsv1.Automation))
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].GetWorkspaceId() == out[j].GetWorkspaceId() {
			return out[i].GetName() < out[j].GetName()
		}
		return out[i].GetWorkspaceId() < out[j].GetWorkspaceId()
	})
	return out, nil
}

func (r *MemoryDefinitionRepo) Get(_ context.Context, workspaceID, name string) (*agentsv1.Automation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.byID[automationID(workspaceID, name)]
	if !ok {
		return nil, ErrAutomationNotFound
	}
	return proto.Clone(a).(*agentsv1.Automation), nil
}

func (r *MemoryDefinitionRepo) Create(_ context.Context, automation *agentsv1.Automation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := automationID(automation.GetWorkspaceId(), automation.GetName())
	if _, exists := r.byID[id]; exists {
		return ErrAutomationAlreadyExists
	}
	r.byID[id] = proto.Clone(automation).(*agentsv1.Automation)
	return nil
}

func (r *MemoryDefinitionRepo) Update(_ context.Context, automation *agentsv1.Automation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := automationID(automation.GetWorkspaceId(), automation.GetName())
	if _, exists := r.byID[id]; !exists {
		return ErrAutomationNotFound
	}
	r.byID[id] = proto.Clone(automation).(*agentsv1.Automation)
	return nil
}

func (r *MemoryDefinitionRepo) Delete(_ context.Context, workspaceID, name string) (*agentsv1.Automation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := automationID(workspaceID, name)
	a, ok := r.byID[id]
	if !ok {
		return nil, ErrAutomationNotFound
	}
	delete(r.byID, id)
	return proto.Clone(a).(*agentsv1.Automation), nil
}

type MemoryRunRepo struct {
	mu   sync.RWMutex
	byID map[string]*agentsv1.AutomationRun
}

func NewMemoryRunRepo() *MemoryRunRepo {
	return &MemoryRunRepo{byID: make(map[string]*agentsv1.AutomationRun)}
}

func (r *MemoryRunRepo) EnsureIndexes(context.Context) error { return nil }

func (r *MemoryRunRepo) Save(_ context.Context, run *agentsv1.AutomationRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[run.GetId()] = proto.Clone(run).(*agentsv1.AutomationRun)
	return nil
}

func (r *MemoryRunRepo) Get(_ context.Context, workspaceID, id string) (*agentsv1.AutomationRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.byID[id]
	if !ok || run.GetWorkspaceId() != workspaceID {
		return nil, ErrRunNotFound
	}
	return proto.Clone(run).(*agentsv1.AutomationRun), nil
}

func (r *MemoryRunRepo) List(_ context.Context, workspaceID, automationName string, pageSize int32, pageToken string) ([]*agentsv1.AutomationRun, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.AutomationRun, 0, len(r.byID))
	for _, run := range r.byID {
		if run.GetWorkspaceId() != workspaceID {
			continue
		}
		if automationName != "" && run.GetAutomationName() != automationName {
			continue
		}
		out = append(out, proto.Clone(run).(*agentsv1.AutomationRun))
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti := out[i].GetStartedAt().AsTime()
		tj := out[j].GetStartedAt().AsTime()
		if ti.Equal(tj) {
			return out[i].GetId() > out[j].GetId()
		}
		return ti.After(tj)
	})
	page, next := paginateRuns(out, pageSize, pageToken)
	return page, next, nil
}

type MemoryStepRunRepo struct {
	mu   sync.RWMutex
	byID map[string]*agentsv1.AutomationStepRun
}

func NewMemoryStepRunRepo() *MemoryStepRunRepo {
	return &MemoryStepRunRepo{byID: make(map[string]*agentsv1.AutomationStepRun)}
}

func (r *MemoryStepRunRepo) EnsureIndexes(context.Context) error { return nil }

func (r *MemoryStepRunRepo) Save(_ context.Context, stepRun *agentsv1.AutomationStepRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[stepRun.GetId()] = proto.Clone(stepRun).(*agentsv1.AutomationStepRun)
	return nil
}

func (r *MemoryStepRunRepo) Get(_ context.Context, workspaceID, id string) (*agentsv1.AutomationStepRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stepRun, ok := r.byID[id]
	if !ok || stepRun.GetWorkspaceId() != workspaceID {
		return nil, ErrStepRunNotFound
	}
	return proto.Clone(stepRun).(*agentsv1.AutomationStepRun), nil
}

func (r *MemoryStepRunRepo) ListByRun(_ context.Context, workspaceID, runID string) ([]*agentsv1.AutomationStepRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.AutomationStepRun, 0)
	for _, stepRun := range r.byID {
		if stepRun.GetWorkspaceId() == workspaceID && stepRun.GetRunId() == runID {
			out = append(out, proto.Clone(stepRun).(*agentsv1.AutomationStepRun))
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].GetOrder() == out[j].GetOrder() {
			return out[i].GetStartedAt().AsTime().Before(out[j].GetStartedAt().AsTime())
		}
		return out[i].GetOrder() < out[j].GetOrder()
	})
	return out, nil
}

func paginateRuns(items []*agentsv1.AutomationRun, pageSize int32, pageToken string) ([]*agentsv1.AutomationRun, string) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := 0
	if pageToken != "" {
		if n, err := strconv.Atoi(pageToken); err == nil && n >= 0 {
			offset = n
		}
	}
	if offset >= len(items) {
		return nil, ""
	}
	end := offset + int(pageSize)
	if end > len(items) {
		end = len(items)
	}
	next := ""
	if end < len(items) {
		next = strconv.Itoa(end)
	}
	return items[offset:end], next
}
