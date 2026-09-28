package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/mem0"
	"go.orx.me/apps/butter/internal/repo/auth"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/mem0memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// workspaceMemoryListLimit caps ListWorkspaceMemories: mem0 OSS
	// listing has no pagination, so search reaches older memories.
	workspaceMemoryListLimit         = 200
	workspaceMemorySearchDefaultTopK = 20
	workspaceMemorySearchMaxTopK     = 100
	// workspaceMemoryCallTimeout bounds one mem0 call; none of these RPCs
	// run an extraction.
	workspaceMemoryCallTimeout = 15 * time.Second
)

// WorkspaceMemoryServiceServer implements
// agentsv1connect.WorkspaceMemoryServiceHandler (issue #339): list,
// search, and delete the memories the workspace's mem0 server holds.
type WorkspaceMemoryServiceServer struct {
	memory        *mem0memory.Service
	workspaceRepo workspacerepo.Repository
}

func NewWorkspaceMemoryServiceServer(memory *mem0memory.Service) *WorkspaceMemoryServiceServer {
	return &WorkspaceMemoryServiceServer{memory: memory}
}

// SetMemory wires the mem0 memory service after bootstrap.
func (s *WorkspaceMemoryServiceServer) SetMemory(memory *mem0memory.Service) { s.memory = memory }

// SetWorkspaceRepo wires the membership lookup behind the role checks.
func (s *WorkspaceMemoryServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

func (s *WorkspaceMemoryServiceServer) requireMemory() error {
	if s.memory == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("memory service not configured"))
	}
	return nil
}

func mapWorkspaceMemoriesErr(err error) *connect.Error {
	switch {
	case errors.Is(err, memoryconn.ErrNotConfigured):
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace memory is not configured; connect a mem0 server in Memory settings"))
	case errors.Is(err, mem0memory.ErrMemoryNotFound):
		return connectx.NotFound("memory not found")
	}
	return connect.NewError(connect.CodeUnavailable, fmt.Errorf("mem0 server: %w", err))
}

// memoryScope turns a request's scope and agent into the service scope.
func memoryScope(workspaceID string, scope agentsv1.WorkspaceMemoryScope, agentID string) (mem0memory.Scope, mem0memory.Target, error) {
	switch scope {
	case agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_UNSPECIFIED,
		agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_WORKSPACE:
		return mem0memory.Scope{WorkspaceID: workspaceID}, mem0memory.TargetWorkspace, nil
	case agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_AGENT:
		agentID = strings.TrimSpace(agentID)
		if agentID == "" {
			return mem0memory.Scope{}, 0, connectx.RequiredArgument("agent_id")
		}
		if strings.ContainsAny(agentID, " \t\n:") {
			return mem0memory.Scope{}, 0, connectx.InvalidArgument("agent_id", "must be an Agent ID")
		}
		return mem0memory.Scope{WorkspaceID: workspaceID, AgentID: agentID}, mem0memory.TargetAgent, nil
	default:
		return mem0memory.Scope{}, 0, connectx.InvalidArgument("scope", "unknown memory scope")
	}
}

func (s *WorkspaceMemoryServiceServer) ListWorkspaceMemories(ctx context.Context, req *connect.Request[agentsv1.ListWorkspaceMemoriesRequest]) (*connect.Response[agentsv1.ListWorkspaceMemoriesResponse], error) {
	if err := s.requireMemory(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	scope, target, err := memoryScope(workspaceID, req.Msg.GetScope(), req.Msg.GetAgentId())
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, workspaceMemoryCallTimeout)
	defer cancel()
	// One extra reveals whether the listing was truncated.
	found, err := s.memory.List(callCtx, scope, target, workspaceMemoryListLimit+1)
	if err != nil {
		return nil, mapWorkspaceMemoriesErr(err)
	}
	truncated := len(found) > workspaceMemoryListLimit
	if truncated {
		found = found[:workspaceMemoryListLimit]
	}
	showPrincipal := isWorkspaceManager(ctx, s.workspaceRepo, workspaceID)
	return connect.NewResponse(&agentsv1.ListWorkspaceMemoriesResponse{
		Memories:  toWorkspaceMemories(found, workspaceID, showPrincipal, false),
		Truncated: truncated,
		Limit:     workspaceMemoryListLimit,
	}), nil
}

func (s *WorkspaceMemoryServiceServer) SearchWorkspaceMemories(ctx context.Context, req *connect.Request[agentsv1.SearchWorkspaceMemoriesRequest]) (*connect.Response[agentsv1.SearchWorkspaceMemoriesResponse], error) {
	if err := s.requireMemory(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	query := strings.TrimSpace(req.Msg.GetQuery())
	if query == "" {
		return nil, connectx.RequiredArgument("query")
	}
	scope, target, err := memoryScope(workspaceID, req.Msg.GetScope(), req.Msg.GetAgentId())
	if err != nil {
		return nil, err
	}
	topK := int(req.Msg.GetTopK())
	if topK <= 0 {
		topK = workspaceMemorySearchDefaultTopK
	}
	topK = min(topK, workspaceMemorySearchMaxTopK)
	// A zero threshold shows every match; people searching to prune want
	// weak matches too, unlike recall.
	threshold := 0.0
	callCtx, cancel := context.WithTimeout(ctx, workspaceMemoryCallTimeout)
	defer cancel()
	found, err := s.memory.Search(callCtx, scope, target, query, mem0memory.SearchOptions{TopK: topK, Threshold: &threshold})
	if err != nil {
		return nil, mapWorkspaceMemoriesErr(err)
	}
	showPrincipal := isWorkspaceManager(ctx, s.workspaceRepo, workspaceID)
	return connect.NewResponse(&agentsv1.SearchWorkspaceMemoriesResponse{
		Memories: toWorkspaceMemories(found, workspaceID, showPrincipal, true),
	}), nil
}

func (s *WorkspaceMemoryServiceServer) DeleteWorkspaceMemory(ctx context.Context, req *connect.Request[agentsv1.DeleteWorkspaceMemoryRequest]) (*connect.Response[agentsv1.DeleteWorkspaceMemoryResponse], error) {
	if err := s.requireMemory(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "workspace_memory"); err != nil {
		return nil, err
	}
	memoryID := strings.TrimSpace(req.Msg.GetMemoryId())
	if memoryID == "" {
		return nil, connectx.RequiredArgument("memory_id")
	}
	callCtx, cancel := context.WithTimeout(ctx, workspaceMemoryCallTimeout)
	defer cancel()
	if err := s.memory.Delete(callCtx, workspaceID, memoryID); err != nil {
		return nil, mapWorkspaceMemoriesErr(err)
	}
	userID := ""
	if user, ok := auth.UserFromContext(ctx); ok {
		userID = user.GetId()
	}
	log.FromContext(ctx).Info("workspace memory deleted",
		"audit", "workspace_memory_delete", "workspace_id", workspaceID, "memory_id", memoryID, "user_id", userID)
	return connect.NewResponse(&agentsv1.DeleteWorkspaceMemoryResponse{}), nil
}

// toWorkspaceMemories converts mem0 memories, skipping any that do not
// belong to the workspace (defense in depth: the identity filter already
// scopes them).
func toWorkspaceMemories(found []mem0.Memory, workspaceID string, showPrincipal, withScore bool) []*agentsv1.WorkspaceMemory {
	out := make([]*agentsv1.WorkspaceMemory, 0, len(found))
	for _, m := range found {
		target, agentID, ok := mem0memory.OwnedBy(m, workspaceID)
		if !ok {
			continue
		}
		wm := &agentsv1.WorkspaceMemory{
			Id:        m.ID,
			Memory:    m.Memory,
			Scope:     agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_WORKSPACE,
			AgentId:   metadataString(m.Metadata, "butter_agent_id"),
			Channel:   metadataString(m.Metadata, "butter_channel"),
			SessionId: metadataString(m.Metadata, "butter_session_id"),
			CreatedAt: memoryTimestamp(m.CreatedAt),
			UpdatedAt: memoryTimestamp(m.UpdatedAt),
		}
		if target == mem0memory.TargetAgent {
			wm.Scope = agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_AGENT
			wm.AgentId = agentID
		}
		if showPrincipal {
			wm.Principal = metadataString(m.Metadata, "butter_principal")
		}
		if withScore {
			wm.Score = m.Score
		}
		out = append(out, wm)
	}
	return out
}

func metadataString(meta map[string]any, key string) string {
	v, _ := meta[key].(string)
	return v
}

// memoryTimestamp parses mem0's ISO-8601 timestamps, with or without a
// zone offset.
func memoryTimestamp(value string) *timestamppb.Timestamp {
	if value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999"} {
		if ts, err := time.Parse(layout, value); err == nil {
			return timestamppb.New(ts)
		}
	}
	return nil
}

// isWorkspaceManager reports, without auditing, whether the caller holds
// the workspace manage role. It only decides what a read reveals; writes
// go through requireWorkspaceManageRole.
func isWorkspaceManager(ctx context.Context, wsRepo workspacerepo.Repository, workspaceID string) bool {
	if auth.IsAdmin(ctx) {
		return true
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok || wsRepo == nil {
		return false
	}
	member, err := wsRepo.GetMember(ctx, workspaceID, user.GetId())
	if err != nil {
		return false
	}
	return slices.Contains([]string{"owner", "admin"}, member.GetRole())
}
