// Package mem0memory is the mem0-backed memory service behind Workspace
// Memory and Agent Memory (ADR-0013). It implements ADK's memory.Service
// for callers that reach memory through ADK, and exposes scope-aware
// extension methods (Search, Recall, Add, Capture) for butter's own recall,
// capture, and memory tools, which ADK's interface cannot express.
//
// The service is one global instance. Every call resolves the workspace's
// WorkspaceMemoryConfig afresh; a workspace without an enabled config gets
// memoryconn.ErrNotConfigured from the extension methods and a silent no-op
// from the ADK interface.
package mem0memory

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/memory"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/mem0"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
)

const (
	// DefaultTopK is the number of memories recalled per turn across both
	// scopes when the agent does not override it.
	DefaultTopK = 5
	// DefaultThreshold is the minimum mem0 relevance score when the agent
	// does not override it.
	DefaultThreshold = 0.3
)

// SearchOptions tunes one search. Zero values fall back to the defaults.
type SearchOptions struct {
	TopK      int
	Threshold *float64
}

func (o SearchOptions) topK() int {
	if o.TopK > 0 {
		return o.TopK
	}
	return DefaultTopK
}

func (o SearchOptions) threshold() float64 {
	if o.Threshold != nil {
		return *o.Threshold
	}
	return DefaultThreshold
}

// Recalled is one memory returned by Recall, tagged with its scope.
type Recalled struct {
	mem0.Memory
	Target Target
}

// Service is the mem0-backed memory service.
type Service struct {
	resolver *memoryconn.Resolver
	// HTTPClient carries mem0 requests; nil uses a client without a global
	// timeout, since callers bound each call with their context.
	HTTPClient *http.Client
}

var _ memory.Service = (*Service)(nil)

func New(resolver *memoryconn.Resolver) *Service {
	return &Service{resolver: resolver}
}

func (s *Service) client(ctx context.Context, workspaceID string) (*mem0.Client, error) {
	conn, err := s.resolver.Resolve(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	httpClient := s.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return conn.Client(httpClient), nil
}

// Search reads one scope.
func (s *Service) Search(ctx context.Context, scope Scope, target Target, query string, opts SearchOptions) ([]mem0.Memory, error) {
	filters, err := scope.filters(target)
	if err != nil {
		return nil, err
	}
	client, err := s.client(ctx, scope.WorkspaceID)
	if err != nil {
		return nil, err
	}
	threshold := opts.threshold()
	return client.Search(ctx, mem0.SearchRequest{
		Query:     query,
		Filters:   filters,
		TopK:      opts.topK(),
		Threshold: &threshold,
	})
}

// Recall searches Workspace Memory and, when the scope names an agent,
// Agent Memory; it merges both, drops duplicate memory IDs, and keeps the
// top opts.TopK by score. mem0 cannot OR two identity filters in one
// request, so the scopes are two searches.
func (s *Service) Recall(ctx context.Context, scope Scope, query string, opts SearchOptions) ([]Recalled, error) {
	targets := []Target{TargetWorkspace}
	if scope.AgentID != "" {
		targets = append(targets, TargetAgent)
	}
	var merged []Recalled
	seen := map[string]bool{}
	for _, target := range targets {
		found, err := s.Search(ctx, scope, target, query, opts)
		if err != nil {
			return nil, fmt.Errorf("search %s memory: %w", target, err)
		}
		for _, m := range found {
			if m.ID == "" || seen[m.ID] {
				continue
			}
			seen[m.ID] = true
			merged = append(merged, Recalled{Memory: m, Target: target})
		}
	}
	slices.SortStableFunc(merged, func(a, b Recalled) int { return cmp.Compare(b.Score, a.Score) })
	if limit := opts.topK(); len(merged) > limit {
		merged = merged[:limit]
	}
	return merged, nil
}

// Add submits messages to one scope for extraction (`infer=true`), tagged
// with the scope's provenance metadata.
func (s *Service) Add(ctx context.Context, scope Scope, target Target, messages []mem0.Message) error {
	if len(messages) == 0 {
		return nil
	}
	userID, agentID, err := scope.identity(target)
	if err != nil {
		return err
	}
	client, err := s.client(ctx, scope.WorkspaceID)
	if err != nil {
		return err
	}
	infer := true
	_, err = client.Add(ctx, mem0.AddRequest{
		Messages: messages,
		UserID:   userID,
		AgentID:  agentID,
		Metadata: scope.metadata(),
		Infer:    &infer,
	})
	return err
}

// Capture submits the user and assistant text of the scope's invocation to
// Workspace Memory. Only events carrying scope.InvocationID are sent, so a
// turn is captured once no matter how long the session grows.
func (s *Service) Capture(ctx context.Context, scope Scope, sess session.Session) error {
	if scope.InvocationID == "" {
		return errors.New("memory capture needs the turn's invocation ID")
	}
	return s.Add(ctx, scope, TargetWorkspace, TurnMessages(sess, scope.InvocationID))
}

// SearchMemory implements memory.Service. The scope comes from the context
// (WithScope); the request's AppName/UserID are ignored because butter's
// memory is keyed by workspace and agent, not by ADK app and user. Without
// a scope or an enabled config it returns no memories.
func (s *Service) SearchMemory(ctx context.Context, req *memory.SearchRequest) (*memory.SearchResponse, error) {
	scope, ok := scopeFor(ctx)
	if !ok {
		return &memory.SearchResponse{}, nil
	}
	recalled, err := s.Recall(ctx, scope, req.Query, SearchOptions{})
	if errors.Is(err, memoryconn.ErrNotConfigured) {
		return &memory.SearchResponse{}, nil
	}
	if err != nil {
		return nil, err
	}
	resp := &memory.SearchResponse{Memories: make([]memory.Entry, 0, len(recalled))}
	for _, m := range recalled {
		resp.Memories = append(resp.Memories, toEntry(m))
	}
	return resp, nil
}

// AddSessionToMemory implements memory.Service by capturing the context's
// invocation (see Capture). Without a scope or an enabled config it does
// nothing.
func (s *Service) AddSessionToMemory(ctx context.Context, sess session.Session) error {
	scope, ok := scopeFor(ctx)
	if !ok {
		return nil
	}
	err := s.Capture(ctx, scope, sess)
	if errors.Is(err, memoryconn.ErrNotConfigured) {
		return nil
	}
	return err
}

// scopeFor returns the context's scope, filling the invocation and session
// from an ADK context when the caller is running inside one.
func scopeFor(ctx context.Context) (Scope, bool) {
	scope, ok := ScopeFromContext(ctx)
	if !ok || scope.WorkspaceID == "" {
		return Scope{}, false
	}
	if rc, ok := ctx.(agent.ReadonlyContext); ok {
		if scope.InvocationID == "" {
			scope.InvocationID = rc.InvocationID()
		}
		if scope.SessionID == "" {
			scope.SessionID = rc.SessionID()
		}
	}
	return scope, true
}

func toEntry(m Recalled) memory.Entry {
	entry := memory.Entry{
		ID:      m.ID,
		Content: genai.NewContentFromText(m.Memory.Memory, genai.RoleModel),
		Author:  "memory",
		CustomMetadata: map[string]any{
			"scope": m.Target.String(),
			"score": m.Score,
		},
	}
	if ts, err := time.Parse(time.RFC3339Nano, m.CreatedAt); err == nil {
		entry.Timestamp = ts
	}
	return entry
}
