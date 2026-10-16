package main

// The mobile file-space screens and the Agent editor's resource associations
// (#456) talk only to the Butter server. These tests drive the production
// AgentFileService, AgentService and the production runner through the
// fixture's HTTP handlers. The model is a scripted double (files.go) that does
// exactly what the run's input says, so what an Agent's saved mounts, MCP
// references and Skills let it do is decided by the real toolsets built from
// the stored record: not by a DTO comparison and not by a UI snapshot. Storage
// is in memory and the model and the MCP server are test doubles; this is not
// a live database, model or device run.

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type filesHarness struct {
	*mcpHarness
	workspace string
	files     agentsv1connect.AgentFileServiceClient
	agents    agentsv1connect.AgentServiceClient
	runs      int
}

func newFilesHarness(t *testing.T, workspace string) *filesHarness {
	t.Helper()
	h := newMCPHarness(t, workspace)
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+h.token)
			req.Header().Set("X-Workspace-ID", workspace)
			return next(ctx, req)
		}
	})
	return &filesHarness{
		mcpHarness: h,
		workspace:  workspace,
		files:      agentsv1connect.NewAgentFileServiceClient(http.DefaultClient, h.url+"/api", connect.WithInterceptors(interceptor)),
		agents:     h.agentClient(workspace),
	}
}

func (h *filesHarness) space(t *testing.T, id, name string) *agentsv1.AgentFileSpace {
	t.Helper()
	resp, err := h.files.CreateAgentFileSpace(context.Background(), connect.NewRequest(&agentsv1.CreateAgentFileSpaceRequest{
		Space: &agentsv1.AgentFileSpace{Id: id, Name: name},
	}))
	if err != nil {
		t.Fatalf("create space %s: %v", id, err)
	}
	return resp.Msg.GetSpace()
}

func (h *filesHarness) write(t *testing.T, space, path, content string) *agentsv1.AgentFile {
	t.Helper()
	resp, err := h.files.WriteAgentFile(context.Background(), connect.NewRequest(&agentsv1.WriteAgentFileRequest{
		SpaceId: space, Path: path, Content: content, ContentType: "text/plain; charset=utf-8",
	}))
	if err != nil {
		t.Fatalf("write %s%s: %v", space, path, err)
	}
	return resp.Msg.GetFile()
}

func (h *filesHarness) read(space, path string, version int64) (*agentsv1.GetAgentFileResponse, error) {
	resp, err := h.files.GetAgentFile(context.Background(), connect.NewRequest(&agentsv1.GetAgentFileRequest{
		SpaceId: space, Path: path, Version: version,
	}))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

func (h *filesHarness) saveAgent(t *testing.T, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	t.Helper()
	resp, err := h.agents.CreateAgent(context.Background(), connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: agent}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetAgent(), nil
}

func (h *filesHarness) updateAgent(t *testing.T, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	t.Helper()
	resp, err := h.agents.UpdateAgent(context.Background(), connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: agent}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetAgent(), nil
}

// run sends one message through the production AG-UI handler (a fresh thread
// each time) and returns the assistant text and the run error, if any.
func (h *filesHarness) run(t *testing.T, agentID, text string) (reply, runError string) {
	t.Helper()
	h.runs++
	thread := fmt.Sprintf("files-%s-%d", agentID, h.runs)
	body, err := json.Marshal(map[string]any{
		"threadId": thread, "runId": "run-" + thread,
		"messages": []map[string]any{{"id": "m1", "role": "user", "content": text}},
		"tools":    []any{}, "context": []any{}, "state": nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, h.url+"/api/agui/"+agentID, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)
	req.Header.Set("X-Workspace-ID", h.workspace)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("run %s: status %d", agentID, resp.StatusCode)
	}
	var out strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		data, ok := strings.CutPrefix(scanner.Text(), "data: ")
		if !ok {
			continue
		}
		var ev map[string]any
		if json.Unmarshal([]byte(data), &ev) != nil {
			continue
		}
		switch ev["type"] {
		case "TEXT_MESSAGE_CONTENT":
			out.WriteString(fmt.Sprint(ev["delta"]))
		case "RUN_ERROR":
			runError = fmt.Sprint(ev["message"])
		}
	}
	return out.String(), runError
}

// offered is the sorted tool list the model was offered on a run.
func (h *filesHarness) offered(t *testing.T, agentID string) []string {
	t.Helper()
	reply, runErr := h.run(t, agentID, "tools?")
	if runErr != "" {
		t.Fatalf("tools? on %s: run error %q", agentID, runErr)
	}
	inner, ok := strings.CutPrefix(reply, "fixture-tools(")
	if !ok {
		t.Fatalf("tools? on %s: reply %q", agentID, reply)
	}
	inner = strings.TrimSuffix(inner, ")")
	if inner == "" {
		return nil
	}
	return strings.Split(inner, ",")
}

func hasTool(tools []string, wanted string) bool {
	for _, name := range tools {
		if name == wanted || strings.HasSuffix(name, wanted) {
			return true
		}
	}
	return false
}

func llmAgent(id string, mutate func(*agentsv1.AgentConfig)) *agentsv1.Agent {
	cfg := &agentsv1.AgentConfig{Model: chatModel, Instruction: "Be brief."}
	if mutate != nil {
		mutate(cfg)
	}
	return &agentsv1.Agent{AgentId: id, Name: id, Type: agentsv1.AgentType_AGENT_TYPE_LLM, Config: cfg}
}

func TestAgentFileServiceSemanticsTheAppMirrors(t *testing.T) {
	h := newFilesHarness(t, filesWorkspace)
	ctx := context.Background()

	created := h.space(t, "docs", "Docs")
	if created.GetWorkspaceId() != filesWorkspace {
		t.Fatalf("created = %v", created)
	}
	_, err := h.files.CreateAgentFileSpace(ctx, connect.NewRequest(&agentsv1.CreateAgentFileSpaceRequest{Space: &agentsv1.AgentFileSpace{Id: "docs", Name: "Again"}}))
	wantCode(t, err, connect.CodeAlreadyExists)

	// The path is normalised by the server: the App shows the same result.
	f1 := h.write(t, "docs", "  notes//a/../b.md ", "one")
	if f1.GetPath() != "/notes/b.md" || f1.GetVersion() != 1 {
		t.Fatalf("first write = %v", f1)
	}
	for _, bad := range []string{"", "/", "   ", "a\x00b"} {
		_, err := h.files.WriteAgentFile(ctx, connect.NewRequest(&agentsv1.WriteAgentFileRequest{SpaceId: "docs", Path: bad, Content: "x"}))
		wantCode(t, err, connect.CodeInvalidArgument)
	}

	// Every write is a new version with no compare-and-swap, and a write
	// replaces the metadata map: an edit that wants to keep it must send it.
	if _, err := h.files.WriteAgentFile(ctx, connect.NewRequest(&agentsv1.WriteAgentFileRequest{
		SpaceId: "docs", Path: "/notes/b.md", Content: "two", ContentType: "text/markdown",
		Metadata: map[string]string{"owner": "me"},
	})); err != nil {
		t.Fatal(err)
	}
	f3 := h.write(t, "docs", "/notes/b.md", "three")
	if f3.GetVersion() != 3 || len(f3.GetMetadata()) != 0 || f3.GetContentType() != "text/plain; charset=utf-8" {
		t.Fatalf("third write = %v", f3)
	}

	// An earlier version is readable by number; one that does not exist is NotFound.
	v1, err := h.read("docs", "/notes/b.md", 1)
	if err != nil || v1.GetContent() != "one" || v1.GetFile().GetVersion() != 1 {
		t.Fatalf("v1 = %v, %v", v1, err)
	}
	v2, err := h.read("docs", "/notes/b.md", 2)
	// Content and content type are versioned; metadata is not (an older
	// version does not return the map it was written with), so the App never
	// presents an older version's metadata.
	if err != nil || v2.GetContent() != "two" || v2.GetFile().GetContentType() != "text/markdown" || len(v2.GetFile().GetMetadata()) != 0 {
		t.Fatalf("v2 = %v, %v", v2, err)
	}
	_, err = h.read("docs", "/notes/b.md", 9)
	wantCode(t, err, connect.CodeNotFound)
	latest, err := h.read("docs", "/notes/b.md", 0)
	if err != nil || latest.GetContent() != "three" || latest.GetFile().GetVersion() != 3 {
		t.Fatalf("latest = %v, %v", latest, err)
	}

	// Listing by prefix and searching by substring.
	h.write(t, "docs", "/other.md", "needle in a haystack")
	list, err := h.files.ListAgentFiles(ctx, connect.NewRequest(&agentsv1.ListAgentFilesRequest{SpaceId: "docs", PathPrefix: "/notes"}))
	if err != nil || len(list.Msg.GetFiles()) != 1 {
		t.Fatalf("prefix list = %v, %v", list, err)
	}
	hits, err := h.files.SearchAgentFiles(ctx, connect.NewRequest(&agentsv1.SearchAgentFilesRequest{SpaceId: "docs", Query: "needle"}))
	if err != nil || len(hits.Msg.GetResults()) != 1 {
		t.Fatalf("search = %v, %v", hits, err)
	}
	_, err = h.files.SearchAgentFiles(ctx, connect.NewRequest(&agentsv1.SearchAgentFilesRequest{SpaceId: "docs", Query: ""}))
	wantCode(t, err, connect.CodeInvalidArgument)

	// The size limit is a byte count of the UTF-8 content, 256 KiB by default,
	// refused with no side effect.
	limit := 256 * 1024
	h.write(t, "docs", "/exact.txt", strings.Repeat("a", limit))
	_, err = h.files.WriteAgentFile(ctx, connect.NewRequest(&agentsv1.WriteAgentFileRequest{SpaceId: "docs", Path: "/big.txt", Content: strings.Repeat("a", limit+1)}))
	wantCode(t, err, connect.CodeInvalidArgument)
	if !strings.Contains(err.Error(), "262144") {
		t.Fatalf("size refusal should name the limit, got %v", err)
	}
	// Multi-byte text counts bytes, not characters: 87,382 three-byte
	// characters are 262,146 bytes.
	_, err = h.files.WriteAgentFile(ctx, connect.NewRequest(&agentsv1.WriteAgentFileRequest{SpaceId: "docs", Path: "/cjk.txt", Content: strings.Repeat("文", 87382)}))
	wantCode(t, err, connect.CodeInvalidArgument)
	if _, err := h.read("docs", "/big.txt", 0); err == nil {
		t.Fatal("refused write left a file behind")
	}

	// A file in another space is another file; spaces never cross Workspaces.
	peer := newFilesHarness(t, filesPeerWorkspace)
	if _, err := peer.read("docs", "/notes/b.md", 0); err == nil {
		t.Fatal("peer workspace read a space it does not own")
	}
	peerSpaces, err := peer.files.ListAgentFileSpaces(ctx, connect.NewRequest(&agentsv1.ListAgentFileSpacesRequest{}))
	if err != nil || len(peerSpaces.Msg.GetSpaces()) != 0 {
		t.Fatalf("peer spaces = %v, %v", peerSpaces, err)
	}

	// A workspace the account does not belong to is refused outright.
	denied := newFilesHarness(t, mcpDeniedWorkspace)
	_, err = denied.files.ListAgentFileSpaces(ctx, connect.NewRequest(&agentsv1.ListAgentFileSpacesRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)

	// Deleting a file, then the space (which takes its files with it).
	if _, err := h.files.DeleteAgentFile(ctx, connect.NewRequest(&agentsv1.DeleteAgentFileRequest{SpaceId: "docs", Path: "/other.md"})); err != nil {
		t.Fatal(err)
	}
	_, err = h.files.DeleteAgentFile(ctx, connect.NewRequest(&agentsv1.DeleteAgentFileRequest{SpaceId: "docs", Path: "/other.md"}))
	wantCode(t, err, connect.CodeNotFound)
	if _, err := h.files.DeleteAgentFileSpace(ctx, connect.NewRequest(&agentsv1.DeleteAgentFileSpaceRequest{Id: "docs"})); err != nil {
		t.Fatal(err)
	}
	_, err = h.read("docs", "/notes/b.md", 0)
	wantCode(t, err, connect.CodeNotFound)
}

func TestSavedAssociationsDecideWhatARealRunCanDo(t *testing.T) {
	h := newFilesHarness(t, filesWorkspace)
	ctx := context.Background()

	h.space(t, "docs", "Docs")
	h.write(t, "docs", "/guide.md", "mounted content")
	h.create(t, h.at(httpServer("open", "Open", "/fake-mcp/open")))

	mount := func(perm agentsv1.AgentFileMountPermission) []*agentsv1.AgentFileMount {
		return []*agentsv1.AgentFileMount{{SpaceId: "docs", MountPath: "/docs", Permission: perm}}
	}
	read := agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ
	readWrite := agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE
	readWriteDelete := agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE

	// An LLM Agent with a mount, an MCP server reference and a Skill.
	saved, err := h.saveAgent(t, llmAgent("assoc", func(c *agentsv1.AgentConfig) {
		c.FileMounts = mount(read)
		c.McpServerIds = []string{"open"}
		c.Skills = []string{filesSkill}
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.GetConfig().GetFileMounts()) != 1 || saved.GetConfig().GetMcpServerIds()[0] != "open" {
		t.Fatalf("saved = %v", saved)
	}
	tools := h.offered(t, "assoc")
	for _, want := range []string{"agent_files_read_file", "agent_files_write_file", "agent_files_list_files", "echo", "list_skills", "load_skill"} {
		if !hasTool(tools, want) {
			t.Errorf("tool %s not offered; offered %v", want, tools)
		}
	}

	// The mount reaches the stored content, under its virtual path.
	reply, runErr := h.run(t, "assoc", `tool:agent_files_read_file {"path":"/docs/guide.md"}`)
	if runErr != "" || !strings.Contains(reply, "mounted content") {
		t.Fatalf("read via mount = %q (%q)", reply, runErr)
	}

	// READ refuses to write, and nothing changes in the space.
	reply, _ = h.run(t, "assoc", `tool:agent_files_write_file {"path":"/docs/guide.md","content":"changed"}`)
	if !strings.Contains(reply, "does not allow write") {
		t.Fatalf("write on a READ mount = %q", reply)
	}
	if got, _ := h.read("docs", "/guide.md", 0); got.GetContent() != "mounted content" || got.GetFile().GetVersion() != 1 {
		t.Fatalf("file changed under a READ mount: %v", got)
	}

	// The MCP server and Skill associations are real: the MCP tool runs and
	// the Skill is listed.
	reply, _ = h.run(t, "assoc", `tool:echo {"text":"hello"}`)
	if !strings.Contains(reply, "hello") {
		t.Fatalf("mcp echo = %q", reply)
	}
	reply, _ = h.run(t, "assoc", `tool:list_skills {}`)
	if !strings.Contains(reply, filesSkill) {
		t.Fatalf("list_skills = %q", reply)
	}

	// Raising the permission to READ_WRITE is a saved change that lets the
	// Agent write (a new version), but not delete.
	stored, err := h.agents.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "assoc"}))
	if err != nil {
		t.Fatal(err)
	}
	up := stored.Msg.GetAgent()
	up.Config.FileMounts = mount(readWrite)
	if _, err := h.updateAgent(t, up); err != nil {
		t.Fatal(err)
	}
	reply, _ = h.run(t, "assoc", `tool:agent_files_write_file {"path":"/docs/guide.md","content":"changed"}`)
	if strings.Contains(reply, "does not allow") {
		t.Fatalf("write on READ_WRITE = %q", reply)
	}
	if got, _ := h.read("docs", "/guide.md", 0); got.GetContent() != "changed" || got.GetFile().GetVersion() != 2 {
		t.Fatalf("file after agent write: %v", got)
	}
	reply, _ = h.run(t, "assoc", `tool:agent_files_delete_file {"path":"/docs/guide.md"}`)
	if !strings.Contains(reply, "does not allow delete") {
		t.Fatalf("delete on READ_WRITE = %q", reply)
	}
	stored, _ = h.agents.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "assoc"}))
	up = stored.Msg.GetAgent()
	up.Config.FileMounts = mount(readWriteDelete)
	if _, err := h.updateAgent(t, up); err != nil {
		t.Fatal(err)
	}
	reply, _ = h.run(t, "assoc", `tool:agent_files_delete_file {"path":"/docs/guide.md"}`)
	if strings.Contains(reply, "does not allow") {
		t.Fatalf("delete on READ_WRITE_DELETE = %q", reply)
	}
	if _, err := h.read("docs", "/guide.md", 0); err == nil {
		t.Fatal("the Agent's delete did not remove the file")
	}

	// An Agent with none of the associations is offered none of the tools.
	if _, err := h.saveAgent(t, llmAgent("bare", nil)); err != nil {
		t.Fatal(err)
	}
	bare := h.offered(t, "bare")
	for _, name := range bare {
		if strings.HasPrefix(name, "agent_files_") || name == "list_skills" || hasTool([]string{name}, "echo") {
			t.Errorf("bare agent was offered %s", name)
		}
	}

	// The same associations on a composite Agent are ignored by the runtime:
	// the child LLM never sees them. The editor therefore shows them as
	// retained, not editable.
	child := llmAgent("seq-child", nil)
	if _, err := h.saveAgent(t, child); err != nil {
		t.Fatal(err)
	}
	seq := &agentsv1.Agent{
		AgentId: "seq", Name: "seq", Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
		ChildAgentIds: []string{"seq-child"},
		Config: &agentsv1.AgentConfig{
			FileMounts:   mount(readWriteDelete),
			McpServerIds: []string{"open"}, Skills: []string{filesSkill},
		},
	}
	if _, err := h.saveAgent(t, seq); err != nil {
		t.Fatal(err)
	}
	for _, name := range h.offered(t, "seq") {
		if strings.HasPrefix(name, "agent_files_") || name == "list_skills" || hasTool([]string{name}, "echo") {
			t.Errorf("composite agent's child was offered %s", name)
		}
	}
}

func TestDanglingReferencesAreRefusedBySaveAndAMissingSpaceFailsHonestly(t *testing.T) {
	h := newFilesHarness(t, filesWorkspace)
	h.create(t, h.at(httpServer("open", "Open", "/fake-mcp/open")))

	// Dangling MCP or Remote Agent ids are refused at save, and nothing is stored.
	_, err := h.saveAgent(t, llmAgent("dangling-mcp", func(c *agentsv1.AgentConfig) { c.McpServerIds = []string{"nope"} }))
	if err == nil {
		t.Fatal("a dangling MCP reference was saved")
	}
	t.Logf("dangling mcp: %v", err)
	_, err = h.saveAgent(t, llmAgent("dangling-remote", func(c *agentsv1.AgentConfig) { c.RemoteAgentIds = []string{"nope"} }))
	if err == nil {
		t.Fatal("a dangling Remote Agent reference was saved")
	}
	t.Logf("dangling remote: %v", err)
	if _, err := h.agents.GetAgent(context.Background(), connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "dangling-mcp"})); err == nil {
		t.Fatal("refused agent exists")
	}

	// A mount of a space that does not exist is accepted (the server does not
	// check) and fails at run time: the tool answers with an error, never data.
	if _, err := h.saveAgent(t, llmAgent("ghost", func(c *agentsv1.AgentConfig) {
		c.FileMounts = []*agentsv1.AgentFileMount{{SpaceId: "ghost-space", MountPath: "/g", Permission: agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ}}
	})); err != nil {
		t.Fatal(err)
	}
	reply, _ := h.run(t, "ghost", `tool:agent_files_read_file {"path":"/g/x.md"}`)
	if !strings.Contains(strings.ToLower(reply), "not found") {
		t.Fatalf("missing space read = %q", reply)
	}

	// A Remote Agent reference that exists is accepted on a composite Agent
	// and refused on a Pi leaf.
	if _, err := h.remote.CreateRemoteAgent(context.Background(), connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "ra", Name: "RA", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A, Url: h.url + "/fake-a2a/ok"},
	})); err != nil {
		t.Fatal(err)
	}
	loop := &agentsv1.Agent{
		AgentId: "loop-ra", Name: "loop-ra", Type: agentsv1.AgentType_AGENT_TYPE_LOOP,
		Config: &agentsv1.AgentConfig{RemoteAgentIds: []string{"ra"}},
	}
	if _, err := h.saveAgent(t, loop); err != nil {
		t.Logf("loop with remote only: %v", err)
	}
	pi := piLeaf("pi-with-files", "no-box", nil)
	pi.Config.FileMounts = []*agentsv1.AgentFileMount{{SpaceId: "docs", MountPath: "/d"}}
	_, err = h.saveAgent(t, pi)
	wantCode(t, err, connect.CodeInvalidArgument)
}
