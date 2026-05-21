package mcpoauth

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrFlowNotFound = errors.New("mcp oauth flow not found")

type Flow struct {
	ID           string
	State        string
	CodeVerifier string
	WorkspaceID  string
	UserID       string
	ServerID     string
	ReturnURL    string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	Resource     string
	Scopes       []string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

func (f *Flow) Clone() *Flow {
	if f == nil {
		return nil
	}
	out := *f
	out.Scopes = append([]string(nil), f.Scopes...)
	return &out
}

type FlowStore interface {
	Save(ctx context.Context, flow *Flow) error
	Consume(ctx context.Context, flowID, state, workspaceID string, now time.Time) (*Flow, error)
	ConsumeByState(ctx context.Context, state string, now time.Time) (*Flow, error)
}

type MemoryFlowStore struct {
	mu     sync.Mutex
	flows  map[string]*Flow
	states map[string]string
}

func NewMemoryFlowStore() *MemoryFlowStore {
	return &MemoryFlowStore{
		flows:  make(map[string]*Flow),
		states: make(map[string]string),
	}
}

func (s *MemoryFlowStore) Save(_ context.Context, flow *Flow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := flow.Clone()
	s.flows[stored.ID] = stored
	s.states[stored.State] = stored.ID
	return nil
}

func (s *MemoryFlowStore) Consume(_ context.Context, flowID, state, workspaceID string, now time.Time) (*Flow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	flow := s.flows[flowID]
	if flow == nil || flow.State != state || flow.WorkspaceID != workspaceID || now.After(flow.ExpiresAt) {
		return nil, ErrFlowNotFound
	}
	delete(s.flows, flowID)
	delete(s.states, flow.State)
	return flow.Clone(), nil
}

func (s *MemoryFlowStore) ConsumeByState(_ context.Context, state string, now time.Time) (*Flow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	flowID := s.states[state]
	flow := s.flows[flowID]
	if flow == nil || now.After(flow.ExpiresAt) {
		return nil, ErrFlowNotFound
	}
	delete(s.flows, flowID)
	delete(s.states, flow.State)
	return flow.Clone(), nil
}
