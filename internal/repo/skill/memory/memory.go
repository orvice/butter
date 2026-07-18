package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Store is an in-memory implementation of skill.Repository.
type Store struct {
	mu       sync.RWMutex
	skills   map[string]map[string]*agentsv1.Skill
	skillMDs map[string]map[string]string
}

func New() *Store {
	return &Store{
		skills:   make(map[string]map[string]*agentsv1.Skill),
		skillMDs: make(map[string]map[string]string),
	}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func cloneSkill(sk *agentsv1.Skill) *agentsv1.Skill {
	return proto.Clone(sk).(*agentsv1.Skill)
}

func notFound(ws, name string) error {
	return fmt.Errorf("skill %q (workspace %q): %w", name, ws, skillrepo.ErrNotFound)
}

func alreadyExists(ws, name string) error {
	return fmt.Errorf("skill %q (workspace %q): %w", name, ws, skillrepo.ErrAlreadyExists)
}

func (s *Store) List(_ context.Context, workspaceID string) ([]*agentsv1.Skill, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.skills[workspaceID]
	out := make([]*agentsv1.Skill, 0, len(bucket))
	for _, sk := range bucket {
		out = append(out, cloneSkill(sk))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (s *Store) Get(_ context.Context, workspaceID, name string) (*agentsv1.Skill, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sk, ok := s.skills[workspaceID][name]
	if !ok {
		return nil, notFound(workspaceID, name)
	}
	return cloneSkill(sk), nil
}

func (s *Store) GetSkillMD(_ context.Context, workspaceID, name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	md, ok := s.skillMDs[workspaceID][name]
	if !ok {
		return "", notFound(workspaceID, name)
	}
	return md, nil
}

func (s *Store) Create(_ context.Context, workspaceID string, sk *agentsv1.Skill, skillMD string) (*agentsv1.Skill, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.skills[workspaceID][sk.GetName()]; ok {
		return nil, alreadyExists(workspaceID, sk.GetName())
	}
	if s.skills[workspaceID] == nil {
		s.skills[workspaceID] = make(map[string]*agentsv1.Skill)
		s.skillMDs[workspaceID] = make(map[string]string)
	}
	stored := cloneSkill(sk)
	now := timestamppb.New(time.Now().UTC())
	stored.WorkspaceId = workspaceID
	stored.CreatedAt = now
	stored.UpdatedAt = now
	s.skills[workspaceID][stored.GetName()] = stored
	s.skillMDs[workspaceID][stored.GetName()] = skillMD
	return cloneSkill(stored), nil
}

func (s *Store) Update(_ context.Context, workspaceID string, sk *agentsv1.Skill, skillMD string) (*agentsv1.Skill, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev, ok := s.skills[workspaceID][sk.GetName()]
	if !ok {
		return nil, notFound(workspaceID, sk.GetName())
	}
	stored := cloneSkill(sk)
	stored.WorkspaceId = workspaceID
	stored.CreatedAt = prev.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	s.skills[workspaceID][stored.GetName()] = stored
	s.skillMDs[workspaceID][stored.GetName()] = skillMD
	return cloneSkill(stored), nil
}

func (s *Store) Delete(_ context.Context, workspaceID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.skills[workspaceID][name]; !ok {
		return notFound(workspaceID, name)
	}
	delete(s.skills[workspaceID], name)
	delete(s.skillMDs[workspaceID], name)
	return nil
}

var _ skillrepo.Repository = (*Store)(nil)
