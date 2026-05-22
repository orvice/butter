package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/agentfile"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fileVersion struct {
	version     int64
	content     string
	contentType string
}

// Store is an in-memory implementation of agentfile.Repository.
type Store struct {
	mu       sync.RWMutex
	spaces   map[string]map[string]*agentsv1.AgentFileSpace
	files    map[string]map[string]map[string]*agentsv1.AgentFile
	contents map[string]map[string]map[string][]fileVersion
}

func New() *Store {
	return &Store{
		spaces:   make(map[string]map[string]*agentsv1.AgentFileSpace),
		files:    make(map[string]map[string]map[string]*agentsv1.AgentFile),
		contents: make(map[string]map[string]map[string][]fileVersion),
	}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func cloneSpace(space *agentsv1.AgentFileSpace) *agentsv1.AgentFileSpace {
	return proto.Clone(space).(*agentsv1.AgentFileSpace)
}

func cloneFile(file *agentsv1.AgentFile) *agentsv1.AgentFile {
	return proto.Clone(file).(*agentsv1.AgentFile)
}

func notFound(entity, ws, key string) error {
	return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, agentfile.ErrNotFound)
}

func alreadyExists(entity, ws, key string) error {
	return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, agentfile.ErrAlreadyExists)
}

func (s *Store) ListSpaces(_ context.Context, workspaceID string) ([]*agentsv1.AgentFileSpace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.spaces[workspaceID]
	out := make([]*agentsv1.AgentFileSpace, 0, len(bucket))
	for _, space := range bucket {
		out = append(out, cloneSpace(space))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (s *Store) GetSpace(_ context.Context, workspaceID, id string) (*agentsv1.AgentFileSpace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	space, ok := s.spaces[workspaceID][id]
	if !ok {
		return nil, notFound("agent file space", workspaceID, id)
	}
	return cloneSpace(space), nil
}

func (s *Store) CreateSpace(_ context.Context, workspaceID string, space *agentsv1.AgentFileSpace) (*agentsv1.AgentFileSpace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.spaces[workspaceID] == nil {
		s.spaces[workspaceID] = make(map[string]*agentsv1.AgentFileSpace)
	}
	stored := cloneSpace(space)
	if stored.Id == "" {
		stored.Id = uuid.NewString()
	}
	if _, ok := s.spaces[workspaceID][stored.GetId()]; ok {
		return nil, alreadyExists("agent file space", workspaceID, stored.GetId())
	}
	now := timestamppb.New(time.Now().UTC())
	stored.WorkspaceId = workspaceID
	stored.CreatedAt = now
	stored.UpdatedAt = now
	s.spaces[workspaceID][stored.GetId()] = stored
	return cloneSpace(stored), nil
}

func (s *Store) UpdateSpace(_ context.Context, workspaceID string, space *agentsv1.AgentFileSpace) (*agentsv1.AgentFileSpace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev, ok := s.spaces[workspaceID][space.GetId()]
	if !ok {
		return nil, notFound("agent file space", workspaceID, space.GetId())
	}
	stored := cloneSpace(space)
	stored.WorkspaceId = workspaceID
	stored.CreatedAt = prev.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	s.spaces[workspaceID][stored.GetId()] = stored
	return cloneSpace(stored), nil
}

func (s *Store) DeleteSpace(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.spaces[workspaceID][id]; !ok {
		return notFound("agent file space", workspaceID, id)
	}
	delete(s.spaces[workspaceID], id)
	delete(s.files[workspaceID], id)
	delete(s.contents[workspaceID], id)
	return nil
}

func (s *Store) ListFiles(_ context.Context, workspaceID, spaceID, pathPrefix string) ([]*agentsv1.AgentFile, error) {
	prefix, err := agentfile.NormalizePrefix(pathPrefix)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.spaces[workspaceID][spaceID]; !ok {
		return nil, notFound("agent file space", workspaceID, spaceID)
	}
	bucket := s.files[workspaceID][spaceID]
	out := make([]*agentsv1.AgentFile, 0, len(bucket))
	for p, file := range bucket {
		if prefix != "" && p != prefix && !strings.HasPrefix(p, strings.TrimRight(prefix, "/")+"/") {
			continue
		}
		out = append(out, cloneFile(file))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetPath() < out[j].GetPath() })
	return out, nil
}

func (s *Store) GetFile(_ context.Context, workspaceID, spaceID, p string) (*agentsv1.AgentFile, error) {
	clean, err := agentfile.NormalizePath(p)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	file, ok := s.files[workspaceID][spaceID][clean]
	if !ok {
		return nil, notFound("agent file", workspaceID, spaceID+":"+clean)
	}
	return cloneFile(file), nil
}

func (s *Store) ReadFile(_ context.Context, workspaceID, spaceID, p string, version int64) (*agentsv1.AgentFile, string, error) {
	clean, err := agentfile.NormalizePath(p)
	if err != nil {
		return nil, "", err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	file, ok := s.files[workspaceID][spaceID][clean]
	if !ok {
		return nil, "", notFound("agent file", workspaceID, spaceID+":"+clean)
	}
	versions := s.contents[workspaceID][spaceID][clean]
	if len(versions) == 0 {
		return nil, "", notFound("agent file content", workspaceID, spaceID+":"+clean)
	}
	want := version
	if want == 0 {
		want = file.GetVersion()
	}
	for _, v := range versions {
		if v.version == want {
			out := cloneFile(file)
			out.Version = v.version
			out.ContentType = v.contentType
			out.SizeBytes = int64(len([]byte(v.content)))
			return out, v.content, nil
		}
	}
	return nil, "", notFound("agent file version", workspaceID, fmt.Sprintf("%s:%s:%d", spaceID, clean, version))
}

func (s *Store) WriteFile(_ context.Context, workspaceID, spaceID, p, content, contentType string, metadata map[string]string) (*agentsv1.AgentFile, error) {
	clean, err := agentfile.NormalizePath(p)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.spaces[workspaceID][spaceID]; !ok {
		return nil, notFound("agent file space", workspaceID, spaceID)
	}
	if s.files[workspaceID] == nil {
		s.files[workspaceID] = make(map[string]map[string]*agentsv1.AgentFile)
	}
	if s.files[workspaceID][spaceID] == nil {
		s.files[workspaceID][spaceID] = make(map[string]*agentsv1.AgentFile)
	}
	if s.contents[workspaceID] == nil {
		s.contents[workspaceID] = make(map[string]map[string][]fileVersion)
	}
	if s.contents[workspaceID][spaceID] == nil {
		s.contents[workspaceID][spaceID] = make(map[string][]fileVersion)
	}

	now := timestamppb.New(time.Now().UTC())
	stored := s.files[workspaceID][spaceID][clean]
	version := int64(1)
	if stored == nil {
		stored = &agentsv1.AgentFile{
			Id:          uuid.NewString(),
			WorkspaceId: workspaceID,
			SpaceId:     spaceID,
			Path:        clean,
			CreatedAt:   now,
		}
	} else {
		version = stored.GetVersion() + 1
		stored = cloneFile(stored)
	}
	stored.ContentType = contentType
	stored.SizeBytes = int64(len([]byte(content)))
	stored.Version = version
	stored.Metadata = metadata
	stored.UpdatedAt = now
	s.files[workspaceID][spaceID][clean] = stored
	s.contents[workspaceID][spaceID][clean] = append(s.contents[workspaceID][spaceID][clean], fileVersion{
		version:     version,
		content:     content,
		contentType: contentType,
	})
	return cloneFile(stored), nil
}

func (s *Store) DeleteFile(_ context.Context, workspaceID, spaceID, p string) error {
	clean, err := agentfile.NormalizePath(p)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.files[workspaceID][spaceID][clean]; !ok {
		return notFound("agent file", workspaceID, spaceID+":"+clean)
	}
	delete(s.files[workspaceID][spaceID], clean)
	delete(s.contents[workspaceID][spaceID], clean)
	return nil
}

func (s *Store) SearchFiles(_ context.Context, workspaceID, spaceID, query string, limit int) ([]*agentsv1.AgentFileSearchResult, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.spaces[workspaceID][spaceID]; !ok {
		return nil, notFound("agent file space", workspaceID, spaceID)
	}
	var out []*agentsv1.AgentFileSearchResult
	for p, file := range s.files[workspaceID][spaceID] {
		versions := s.contents[workspaceID][spaceID][p]
		if len(versions) == 0 {
			continue
		}
		content := versions[len(versions)-1].content
		lower := strings.ToLower(content)
		idx := strings.Index(lower, query)
		if idx < 0 {
			continue
		}
		out = append(out, &agentsv1.AgentFileSearchResult{
			File:     cloneFile(file),
			Snippets: []string{snippet(content, idx, len(query))},
		})
		if len(out) >= limit {
			break
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetFile().GetPath() < out[j].GetFile().GetPath() })
	return out, nil
}

func snippet(content string, idx, queryLen int) string {
	start := idx - 80
	if start < 0 {
		start = 0
	}
	end := idx + queryLen + 80
	if end > len(content) {
		end = len(content)
	}
	return content[start:end]
}

var _ agentfile.Repository = (*Store)(nil)
