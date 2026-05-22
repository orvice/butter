package agentfiletool

import (
	"context"
	"testing"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/toolconfirmation"
	"google.golang.org/genai"

	agentfilememory "go.orx.me/apps/butter/internal/repo/agentfile/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestResolveMountPermissions(t *testing.T) {
	ts := &Toolset{mounts: normalizeMounts([]*agentsv1.AgentFileMount{
		{
			SpaceId:    "docs",
			MountPath:  "/docs",
			Permission: agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ,
		},
	})}

	if _, err := ts.resolve("/docs/readme.md", agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ); err != nil {
		t.Fatalf("read resolve failed: %v", err)
	}
	if _, err := ts.resolve("/docs/readme.md", agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE); err == nil {
		t.Fatal("write resolve should fail on read-only mount")
	}
	if _, err := ts.resolve("/other/readme.md", agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ); err == nil {
		t.Fatal("resolve should fail outside mounts")
	}
}

func TestToolsetReadWrite(t *testing.T) {
	ctx := context.Background()
	repo := agentfilememory.New()
	space, err := repo.CreateSpace(ctx, "ws-1", &agentsv1.AgentFileSpace{Name: "Notes"})
	if err != nil {
		t.Fatalf("CreateSpace: %v", err)
	}
	raw, err := NewToolset(repo, []*agentsv1.AgentFileMount{
		{
			SpaceId:    space.GetId(),
			MountPath:  "/notes",
			Permission: agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE,
		},
	}, 1024)
	if err != nil {
		t.Fatalf("NewToolset: %v", err)
	}
	ts := raw.(*Toolset)
	toolCtx := testToolContext{Context: WithRuntimeContext(ctx, RuntimeContext{WorkspaceID: "ws-1", AgentName: "agent"})}
	if _, err := ts.writeFile(toolCtx, writeFileArgs{Path: "/notes/todo.md", Content: "hello"}); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	got, err := ts.readFile(toolCtx, readFileArgs{Path: "/notes/todo.md"})
	if err != nil {
		t.Fatalf("readFile: %v", err)
	}
	if got.Content != "hello" {
		t.Fatalf("content = %q, want hello", got.Content)
	}
}

func TestListFilesAllowsMountRootPrefix(t *testing.T) {
	ctx := context.Background()
	repo := agentfilememory.New()
	space, err := repo.CreateSpace(ctx, "ws-1", &agentsv1.AgentFileSpace{Name: "Docs"})
	if err != nil {
		t.Fatalf("CreateSpace: %v", err)
	}
	raw, err := NewToolset(repo, []*agentsv1.AgentFileMount{
		{
			SpaceId:    space.GetId(),
			MountPath:  "/docs",
			Permission: agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE,
		},
	}, 1024)
	if err != nil {
		t.Fatalf("NewToolset: %v", err)
	}
	ts := raw.(*Toolset)
	toolCtx := testToolContext{Context: WithRuntimeContext(ctx, RuntimeContext{WorkspaceID: "ws-1", AgentName: "agent"})}
	if _, err := ts.writeFile(toolCtx, writeFileArgs{Path: "/docs/readme.md", Content: "hello"}); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	got, err := ts.listFiles(toolCtx, listFilesArgs{PathPrefix: "/docs"})
	if err != nil {
		t.Fatalf("listFiles: %v", err)
	}
	if len(got.Files) != 1 || got.Files[0].Path != "/docs/readme.md" {
		t.Fatalf("files = %#v, want /docs/readme.md", got.Files)
	}
}

type testToolContext struct {
	context.Context
}

func (t testToolContext) FunctionCallID() string { return "call" }
func (t testToolContext) Actions() *session.EventActions {
	return &session.EventActions{}
}
func (t testToolContext) SearchMemory(context.Context, string) (*memory.SearchResponse, error) {
	return nil, nil
}
func (t testToolContext) ToolConfirmation() *toolconfirmation.ToolConfirmation { return nil }
func (t testToolContext) RequestConfirmation(string, any) error                { return nil }
func (t testToolContext) UserContent() *genai.Content                          { return nil }
func (t testToolContext) InvocationID() string                                 { return "invocation" }
func (t testToolContext) AgentName() string                                    { return "agent" }
func (t testToolContext) ReadonlyState() session.ReadonlyState                 { return nil }
func (t testToolContext) UserID() string                                       { return "user" }
func (t testToolContext) AppName() string                                      { return "app" }
func (t testToolContext) SessionID() string                                    { return "session" }
func (t testToolContext) Branch() string                                       { return "" }
func (t testToolContext) Artifacts() agent.Artifacts                           { return nil }
func (t testToolContext) State() session.State                                 { return nil }

var _ tool.Context = testToolContext{}
