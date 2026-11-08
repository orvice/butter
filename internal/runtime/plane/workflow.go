package plane

import (
	"context"
	"errors"
	"fmt"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
	"strings"
	"sync"
)

type WorkflowStep struct {
	ID     string
	Prompt string
}
type WorkflowDefinition struct {
	ID      string
	Enabled bool
	Steps   []WorkflowStep
}
type WorkflowStepState string

const (
	WorkflowPending   WorkflowStepState = "pending"
	WorkflowRunning   WorkflowStepState = "running"
	WorkflowSucceeded WorkflowStepState = "succeeded"
	WorkflowFailed    WorkflowStepState = "failed"
	WorkflowStopped   WorkflowStepState = "stopped"
)

type WorkflowRecord struct {
	Key         string
	WorkspaceID string
	AppID       string
	RunID       string
	ActivityID  string
	WorkflowID  string
	Status      WorkflowStepState
	StepStates  map[string]WorkflowStepState
	Outputs     map[string]string
	Error       string
}
type WorkflowStore interface {
	Put(context.Context, *WorkflowRecord) error
	Get(context.Context, string, string) (*WorkflowRecord, error)
}
type MemoryWorkflowStore struct {
	mu   sync.Mutex
	rows map[string]*WorkflowRecord
}

func NewMemoryWorkflowStore() *MemoryWorkflowStore {
	return &MemoryWorkflowStore{rows: map[string]*WorkflowRecord{}}
}
func (s *MemoryWorkflowStore) Put(_ context.Context, r *WorkflowRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := *r
	c.StepStates = map[string]WorkflowStepState{}
	for k, v := range r.StepStates {
		c.StepStates[k] = v
	}
	c.Outputs = map[string]string{}
	for k, v := range r.Outputs {
		c.Outputs[k] = v
	}
	s.rows[r.Key] = &c
	return nil
}
func (s *MemoryWorkflowStore) Get(_ context.Context, ws, k string) (*WorkflowRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rows[k]
	if !ok || r.WorkspaceID != ws {
		return nil, errors.New("workflow record not found")
	}
	c := *r
	return &c, nil
}

type WorkflowExecutor struct {
	runner Runner
	store  WorkflowStore
	mu     sync.Mutex
	active map[string]context.CancelFunc
}

func NewWorkflowExecutor(r Runner, s WorkflowStore) *WorkflowExecutor {
	return &WorkflowExecutor{runner: r, store: s, active: map[string]context.CancelFunc{}}
}
func (w *WorkflowExecutor) Run(ctx context.Context, app *agentsv1.PlaneApp, e *Event, def WorkflowDefinition) (*WorkflowRecord, error) {
	if w == nil || w.runner == nil {
		return nil, ErrRunnerUnavailable
	}
	if !def.Enabled || def.ID == "" || len(def.Steps) == 0 {
		return nil, errors.New("workflow is not configured")
	}
	for _, step := range def.Steps {
		if step.ID == "" || strings.TrimSpace(step.Prompt) == "" {
			return nil, errors.New("workflow step is invalid")
		}
	}
	key := e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID + ":" + def.ID
	w.mu.Lock()
	if _, ok := w.active[key]; ok {
		w.mu.Unlock()
		return nil, ErrDuplicate
	}
	child, cancel := context.WithCancel(ctx)
	w.active[key] = cancel
	w.mu.Unlock()
	defer func() { cancel(); w.mu.Lock(); delete(w.active, key); w.mu.Unlock() }()
	rec := &WorkflowRecord{Key: key, WorkspaceID: e.WorkspaceID, AppID: e.AppID, RunID: e.AgentRunID, ActivityID: e.ActivityID, WorkflowID: def.ID, Status: WorkflowRunning, StepStates: map[string]WorkflowStepState{}, Outputs: map[string]string{}}
	if w.store != nil {
		_ = w.store.Put(ctx, rec)
	}
	for _, step := range def.Steps {
		rec.StepStates[step.ID] = WorkflowRunning
		if w.store != nil {
			_ = w.store.Put(ctx, rec)
		}
		out, err := w.runner.Run(child, app.GetAgentId(), []*genai.Part{{Text: step.Prompt}}, "", &agentsv1.ContextInfo{UserId: e.BotUserID, SessionId: e.AgentRunID, ChannelName: "plane"}, nil, nil)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				rec.Status = WorkflowStopped
				rec.StepStates[step.ID] = WorkflowStopped
			} else {
				rec.Status = WorkflowFailed
				rec.StepStates[step.ID] = WorkflowFailed
				rec.Error = fmt.Sprintf("step %s: %v", step.ID, err)
			}
			if w.store != nil {
				_ = w.store.Put(ctx, rec)
			}
			return rec, err
		}
		rec.StepStates[step.ID] = WorkflowSucceeded
		rec.Outputs[step.ID] = strings.TrimSpace(out)
		rec.Status = WorkflowSucceeded
		if w.store != nil {
			_ = w.store.Put(ctx, rec)
		}
	}
	return rec, nil
}
func (w *WorkflowExecutor) Stop(e *Event, workflowID string) bool {
	k := e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID + ":" + workflowID
	w.mu.Lock()
	defer w.mu.Unlock()
	if c, ok := w.active[k]; ok {
		c()
		return true
	}
	return false
}
