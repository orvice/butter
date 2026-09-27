package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/mem0"
	memoryconfigmem "go.orx.me/apps/butter/internal/repo/memoryconfig/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/runtime/mem0memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// storeMem0 is a stateful fake mem0 OSS server over an in-memory table.
type storeMem0 struct {
	srv *httptest.Server

	mu       sync.Mutex
	memories map[string]mem0.Memory
	searches []map[string]any
	deleted  []string
}

func newStoreMem0(t *testing.T, seed ...mem0.Memory) *storeMem0 {
	f := &storeMem0{memories: map[string]mem0.Memory{}}
	for _, m := range seed {
		f.memories[m.ID] = m
	}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		matches := func(m mem0.Memory, userID, agentID string) bool {
			return (userID != "" && m.UserID == userID) || (agentID != "" && m.AgentID == agentID)
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/memories":
			topK, _ := strconv.Atoi(r.URL.Query().Get("top_k"))
			var out []mem0.Memory
			for _, m := range f.memories {
				if matches(m, r.URL.Query().Get("user_id"), r.URL.Query().Get("agent_id")) {
					out = append(out, m)
				}
			}
			if topK > 0 && len(out) > topK {
				out = out[:topK]
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"results": out})
		case r.Method == http.MethodPost && r.URL.Path == "/search":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			f.searches = append(f.searches, body)
			filters, _ := body["filters"].(map[string]any)
			userID, _ := filters["user_id"].(string)
			agentID, _ := filters["agent_id"].(string)
			var out []mem0.Memory
			for _, m := range f.memories {
				if matches(m, userID, agentID) && strings.Contains(m.Memory, body["query"].(string)) {
					m.Score = 0.7
					out = append(out, m)
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"results": out})
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/memories/"):
			m, ok := f.memories[strings.TrimPrefix(r.URL.Path, "/memories/")]
			if !ok {
				_, _ = w.Write([]byte("null"))
				return
			}
			_ = json.NewEncoder(w).Encode(m)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/memories/"):
			id := strings.TrimPrefix(r.URL.Path, "/memories/")
			if _, ok := f.memories[id]; !ok {
				http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
				return
			}
			delete(f.memories, id)
			f.deleted = append(f.deleted, id)
			_, _ = w.Write([]byte(`{"message":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *storeMem0) state() (searches []map[string]any, deleted []string, remaining int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]any(nil), f.searches...), append([]string(nil), f.deleted...), len(f.memories)
}

func seedMemory(id, text, userID, agentID, created string) mem0.Memory {
	return mem0.Memory{
		ID: id, Memory: text, UserID: userID, AgentID: agentID, CreatedAt: created,
		Metadata: map[string]any{
			"butter_agent_id":   "helper",
			"butter_channel":    "telegram",
			"butter_principal":  "tg:42",
			"butter_session_id": "s1",
		},
	}
}

func newWorkspaceMemoriesFixture(t *testing.T, enabled bool, seed ...mem0.Memory) (*WorkspaceMemoryServiceServer, *storeMem0) {
	t.Helper()
	f := newStoreMem0(t, seed...)
	repo := memoryconfigmem.New()
	if _, err := repo.Put(t.Context(), "ws-a", &agentsv1.WorkspaceMemoryConfig{BaseUrl: f.srv.URL, Enabled: enabled}, nil); err != nil {
		t.Fatal(err)
	}
	memSvc := mem0memory.New(memoryconn.NewResolver(repo, nil))
	memSvc.HTTPClient = f.srv.Client()

	wsRepo := workspacememory.New()
	if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: "ws-a", Name: "ws-a", Slug: "ws-a"}); err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct{ user, role string }{{"owner-user", "owner"}, {"member-user", "member"}} {
		if _, err := wsRepo.AddMember(t.Context(), &agentsv1.WorkspaceMember{WorkspaceId: "ws-a", UserId: m.user, Role: m.role}); err != nil {
			t.Fatal(err)
		}
	}
	svc := NewWorkspaceMemoryServiceServer(memSvc)
	svc.SetWorkspaceRepo(wsRepo)
	return svc, f
}

var memorySeed = []mem0.Memory{
	seedMemory("w-old", "team deploys on friday", "ws:ws-a", "", "2026-09-01T10:00:00+00:00"),
	seedMemory("w-new", "team uses pnpm", "ws:ws-a", "", "2026-09-20T10:00:00.5+00:00"),
	seedMemory("a-1", "answer tersely", "", "ws:ws-a:agent:helper", "2026-09-10T10:00:00+00:00"),
	seedMemory("foreign-w", "other tenant pnpm secret", "ws:ws-b", "", "2026-09-15T10:00:00+00:00"),
	seedMemory("foreign-a", "other tenant agent", "", "ws:ws-b:agent:helper", "2026-09-15T10:00:00+00:00"),
}

func TestListWorkspaceMemoriesScopesAndRedacts(t *testing.T) {
	svc, _ := newWorkspaceMemoriesFixture(t, true, memorySeed...)

	resp, err := svc.ListWorkspaceMemories(memberCtx(), connect.NewRequest(&agentsv1.ListWorkspaceMemoriesRequest{}))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	got := resp.Msg.GetMemories()
	if len(got) != 2 || got[0].GetId() != "w-new" || got[1].GetId() != "w-old" {
		t.Fatalf("workspace memories = %v, want w-new then w-old", got)
	}
	first := got[0]
	if first.GetScope() != agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_WORKSPACE || first.GetAgentId() != "helper" ||
		first.GetChannel() != "telegram" || first.GetSessionId() != "s1" || first.GetCreatedAt() == nil {
		t.Fatalf("provenance = %v", first)
	}
	if first.GetPrincipal() != "" {
		t.Fatalf("principal shown to a member: %q", first.GetPrincipal())
	}
	if resp.Msg.GetTruncated() || resp.Msg.GetLimit() != workspaceMemoryListLimit {
		t.Fatalf("truncated = %v, limit = %d", resp.Msg.GetTruncated(), resp.Msg.GetLimit())
	}

	resp, err = svc.ListWorkspaceMemories(ownerCtx(), connect.NewRequest(&agentsv1.ListWorkspaceMemoriesRequest{}))
	if err != nil || resp.Msg.GetMemories()[0].GetPrincipal() != "tg:42" {
		t.Fatalf("owner principal = %v, %v", resp, err)
	}
	resp, err = svc.ListWorkspaceMemories(ctxAs("root", "admin", "ws-a"), connect.NewRequest(&agentsv1.ListWorkspaceMemoriesRequest{}))
	if err != nil || resp.Msg.GetMemories()[0].GetPrincipal() != "tg:42" {
		t.Fatalf("global admin principal = %v, %v", resp, err)
	}
}

func TestListAgentMemories(t *testing.T) {
	svc, _ := newWorkspaceMemoriesFixture(t, true, memorySeed...)

	_, err := svc.ListWorkspaceMemories(memberCtx(), connect.NewRequest(&agentsv1.ListWorkspaceMemoriesRequest{
		Scope: agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_AGENT,
	}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("agent scope without agent_id: %v", code)
	}
	resp, err := svc.ListWorkspaceMemories(memberCtx(), connect.NewRequest(&agentsv1.ListWorkspaceMemoriesRequest{
		Scope: agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_AGENT, AgentId: "helper",
	}))
	if err != nil {
		t.Fatalf("List agent: %v", err)
	}
	got := resp.Msg.GetMemories()
	if len(got) != 1 || got[0].GetId() != "a-1" || got[0].GetAgentId() != "helper" ||
		got[0].GetScope() != agentsv1.WorkspaceMemoryScope_WORKSPACE_MEMORY_SCOPE_AGENT {
		t.Fatalf("agent memories = %v", got)
	}
}

func TestListWorkspaceMemoriesReportsTruncation(t *testing.T) {
	var many []mem0.Memory
	for i := 0; i < workspaceMemoryListLimit+5; i++ {
		many = append(many, seedMemory(fmt.Sprintf("m%03d", i), "fact", "ws:ws-a", "", fmt.Sprintf("2026-09-01T10:%02d:%02d+00:00", i/60, i%60)))
	}
	svc, _ := newWorkspaceMemoriesFixture(t, true, many...)
	resp, err := svc.ListWorkspaceMemories(memberCtx(), connect.NewRequest(&agentsv1.ListWorkspaceMemoriesRequest{}))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Msg.GetMemories()) != workspaceMemoryListLimit || !resp.Msg.GetTruncated() {
		t.Fatalf("got %d memories, truncated = %v", len(resp.Msg.GetMemories()), resp.Msg.GetTruncated())
	}
}

func TestSearchWorkspaceMemories(t *testing.T) {
	svc, f := newWorkspaceMemoriesFixture(t, true, memorySeed...)

	_, err := svc.SearchWorkspaceMemories(memberCtx(), connect.NewRequest(&agentsv1.SearchWorkspaceMemoriesRequest{Query: " "}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("empty query: %v", code)
	}
	resp, err := svc.SearchWorkspaceMemories(memberCtx(), connect.NewRequest(&agentsv1.SearchWorkspaceMemoriesRequest{Query: "pnpm", TopK: 500}))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	got := resp.Msg.GetMemories()
	if len(got) != 1 || got[0].GetId() != "w-new" || got[0].GetScore() != 0.7 {
		t.Fatalf("search results = %v (the other tenant's pnpm memory must not appear)", got)
	}
	searches, _, _ := f.state()
	if searches[0]["top_k"] != float64(workspaceMemorySearchMaxTopK) || searches[0]["threshold"] != 0.0 {
		t.Fatalf("search body = %v", searches[0])
	}
	if searches[0]["filters"].(map[string]any)["user_id"] != "ws:ws-a" {
		t.Fatalf("search filters = %v", searches[0]["filters"])
	}
}

func TestDeleteWorkspaceMemoryChecksRoleAndOwnership(t *testing.T) {
	svc, f := newWorkspaceMemoriesFixture(t, true, memorySeed...)
	del := func(ctxID string, id string) error {
		ctx := memberCtx()
		if ctxID == "owner" {
			ctx = ownerCtx()
		}
		_, err := svc.DeleteWorkspaceMemory(ctx, connect.NewRequest(&agentsv1.DeleteWorkspaceMemoryRequest{MemoryId: id}))
		return err
	}

	if code := connectCode(t, del("member", "w-old")); code != connect.CodePermissionDenied {
		t.Fatalf("member delete: %v", code)
	}
	for _, id := range []string{"foreign-w", "foreign-a", "missing"} {
		if code := connectCode(t, del("owner", id)); code != connect.CodeNotFound {
			t.Fatalf("owner delete %s: %v, want NotFound", id, code)
		}
	}
	if err := del("owner", "w-old"); err != nil {
		t.Fatalf("delete own workspace memory: %v", err)
	}
	if err := del("owner", "a-1"); err != nil {
		t.Fatalf("delete own agent memory: %v", err)
	}
	_, deleted, remaining := f.state()
	if strings.Join(deleted, ",") != "w-old,a-1" || remaining != 3 {
		t.Fatalf("deleted = %v, remaining = %d; foreign memories must survive", deleted, remaining)
	}
}

func TestWorkspaceMemoriesWithoutConfig(t *testing.T) {
	svc, f := newWorkspaceMemoriesFixture(t, false, memorySeed...)
	_, err := svc.ListWorkspaceMemories(memberCtx(), connect.NewRequest(&agentsv1.ListWorkspaceMemoriesRequest{}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("List without config: %v", code)
	}
	_, err = svc.DeleteWorkspaceMemory(ownerCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceMemoryRequest{MemoryId: "w-old"}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("Delete without config: %v", code)
	}
	if _, deleted, _ := f.state(); len(deleted) != 0 {
		t.Fatal("a disabled workspace reached mem0")
	}
}

func TestWorkspaceMemoriesUnavailableServer(t *testing.T) {
	svc, f := newWorkspaceMemoriesFixture(t, true)
	f.srv.Close()
	_, err := svc.ListWorkspaceMemories(memberCtx(), connect.NewRequest(&agentsv1.ListWorkspaceMemoriesRequest{}))
	if code := connectCode(t, err); code != connect.CodeUnavailable {
		t.Fatalf("List against a down server: %v", code)
	}
}
