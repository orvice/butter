package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/toolutils"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// testReadonlyContext is what mcptoolset's lazy Tools call needs of the
// invocation; the assertions never leave this test package.
type testReadonlyContext struct{ context.Context }

func (testReadonlyContext) UserContent() *genai.Content          { return nil }
func (testReadonlyContext) InvocationID() string                 { return "inv-1" }
func (testReadonlyContext) AgentName() string                    { return "Writer" }
func (testReadonlyContext) ReadonlyState() session.ReadonlyState { return nil }
func (testReadonlyContext) UserID() string                       { return "u1" }
func (testReadonlyContext) AppName() string                      { return "agent-test" }
func (testReadonlyContext) SessionID() string                    { return "s1" }
func (testReadonlyContext) Branch() string                       { return "" }

var _ agent.ReadonlyContext = testReadonlyContext{}

// fakeMCPTool is one MCP tool as the ADK toolset hands it over: a name, a
// declaration, and a Run that reports which server it belongs to.
type fakeMCPTool struct {
	server string
	name   string
}

func (t *fakeMCPTool) Name() string        { return t.name }
func (t *fakeMCPTool) Description() string { return t.server + " tool " + t.name }
func (t *fakeMCPTool) IsLongRunning() bool { return false }

func (t *fakeMCPTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{Name: t.name, Description: t.Description()}
}

func (t *fakeMCPTool) ProcessRequest(_ agent.Context, req *model.LLMRequest) error {
	return toolutils.PackTool(req, t)
}

func (t *fakeMCPTool) Run(_ agent.Context, args any) (map[string]any, error) {
	return map[string]any{"server": t.server, "tool": t.name, "args": args}, nil
}

type fakeMCPServer struct {
	name  string
	tools []tool.Tool
}

func (s *fakeMCPServer) Name() string { return s.name }

func (s *fakeMCPServer) Tools(agent.ReadonlyContext) ([]tool.Tool, error) {
	return s.tools, nil
}

func fakeServer(name string, toolNames ...string) *fakeMCPServer {
	s := &fakeMCPServer{name: name}
	for _, toolName := range toolNames {
		s.tools = append(s.tools, &fakeMCPTool{server: name, name: toolName})
	}
	return s
}

func entry(s *fakeMCPServer) mcpToolsetEntry {
	return mcpToolsetEntry{server: s.name, toolset: s}
}

func toolNames(tools []tool.Tool) []string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name()
	}
	return names
}

// packAll mirrors ADK's tool preprocess: every tool packs its declaration into
// one model request, which is where duplicate names used to fail the run.
func packAll(t *testing.T, tools []tool.Tool) {
	t.Helper()
	req := &model.LLMRequest{}
	for _, tl := range tools {
		packer, ok := tl.(interface {
			ProcessRequest(agent.Context, *model.LLMRequest) error
		})
		if !ok {
			t.Fatalf("tool %q does not implement ProcessRequest", tl.Name())
		}
		if err := packer.ProcessRequest(nil, req); err != nil {
			t.Fatalf("packing tool %q: %v", tl.Name(), err)
		}
	}
	if len(req.Tools) != len(tools) {
		t.Fatalf("packed %d tools, want %d", len(req.Tools), len(tools))
	}
	// The declarations the model sees are packed alongside the tools.
	var declared []string
	for _, gt := range req.Config.Tools {
		for _, decl := range gt.FunctionDeclarations {
			declared = append(declared, decl.Name)
		}
	}
	if len(declared) != len(tools) {
		t.Fatalf("packed %d declarations, want %d", len(declared), len(tools))
	}
	for i, tl := range tools {
		if declared[i] != tl.Name() {
			t.Fatalf("declaration %d = %q, want %q", i, declared[i], tl.Name())
		}
	}
}

func TestMCPToolsetsRenamesCollidingTools(t *testing.T) {
	alpha := fakeServer("alpha", "search", "read_doc")
	beta := fakeServer("beta", "search")
	group := newMCPToolsets([]mcpToolsetEntry{entry(alpha), entry(beta)})

	tools, err := group.Tools(nil)
	if err != nil {
		t.Fatalf("Tools: %v", err)
	}

	want := []string{"alpha__search", "read_doc", "beta__search"}
	if got := toolNames(tools); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("names = %v, want %v", got, want)
	}

	// Deterministic: the same group lists the same names again.
	again, err := group.Tools(nil)
	if err != nil {
		t.Fatalf("Tools (second call): %v", err)
	}
	if got := toolNames(again); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("second listing names = %v, want %v", got, want)
	}

	// Both servers' search is reachable, and under its own new name.
	packAll(t, tools)

	run := func(toolName string) map[string]any {
		t.Helper()
		var target tool.Tool
		for _, tl := range tools {
			if tl.Name() == toolName {
				target = tl
			}
		}
		if target == nil {
			t.Fatalf("no tool named %q", toolName)
		}
		runnable, ok := target.(interface {
			Run(agent.Context, any) (map[string]any, error)
		})
		if !ok {
			t.Fatalf("tool %q is not runnable", toolName)
		}
		result, err := runnable.Run(nil, "q")
		if err != nil {
			t.Fatalf("running %q: %v", toolName, err)
		}
		return result
	}

	if got := run("alpha__search"); got["server"] != "alpha" || got["tool"] != "search" || got["args"] != "q" {
		t.Fatalf("alpha__search ran %v", got)
	}
	if got := run("beta__search"); got["server"] != "beta" || got["tool"] != "search" {
		t.Fatalf("beta__search ran %v", got)
	}
}

func TestMCPToolsetsKeepsNonCollidingNames(t *testing.T) {
	alpha := fakeServer("alpha", "alpha_only")
	beta := fakeServer("beta", "beta_only")
	group := newMCPToolsets([]mcpToolsetEntry{entry(alpha), entry(beta)})

	tools, err := group.Tools(nil)
	if err != nil {
		t.Fatalf("Tools: %v", err)
	}
	if got, want := toolNames(tools), []string{"alpha_only", "beta_only"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("names = %v, want %v", got, want)
	}
	// Unchanged names keep the toolset's own tool rather than a wrapper.
	if tools[0] != alpha.tools[0] || tools[1] != beta.tools[0] {
		t.Fatal("a tool without a name conflict was wrapped")
	}
	packAll(t, tools)
}

func TestMCPToolsetsDisambiguatesRepeatedNames(t *testing.T) {
	t.Run("same tool twice in one server and once in another", func(t *testing.T) {
		group := newMCPToolsets([]mcpToolsetEntry{
			entry(fakeServer("docs", "echo", "echo")),
			entry(fakeServer("docs", "echo")),
		})
		tools, err := group.Tools(nil)
		if err != nil {
			t.Fatalf("Tools: %v", err)
		}
		want := []string{"docs__echo", "docs__echo__2", "docs__echo__3"}
		if got := toolNames(tools); strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("names = %v, want %v", got, want)
		}
		packAll(t, tools)

		again, err := group.Tools(nil)
		if err != nil {
			t.Fatalf("Tools (second call): %v", err)
		}
		if got := toolNames(again); strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("second listing names = %v, want %v", got, want)
		}
	})

	t.Run("names a provider rejects are sanitized", func(t *testing.T) {
		group := newMCPToolsets([]mcpToolsetEntry{
			entry(fakeServer("My Docs!", "search")),
			entry(fakeServer("9lives", "search")),
			entry(fakeServer("", "search")),
		})
		tools, err := group.Tools(nil)
		if err != nil {
			t.Fatalf("Tools: %v", err)
		}
		want := []string{"My_Docs__search", "mcp_9lives__search", "mcp__search"}
		if got := toolNames(tools); strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("names = %v, want %v", got, want)
		}
		packAll(t, tools)
	})

	t.Run("long names stay within the provider limit", func(t *testing.T) {
		longServer := strings.Repeat("s", 40)
		longTool := strings.Repeat("t", 40)
		group := newMCPToolsets([]mcpToolsetEntry{
			entry(fakeServer(longServer, longTool)),
			entry(fakeServer("other", longTool)),
		})
		tools, err := group.Tools(nil)
		if err != nil {
			t.Fatalf("Tools: %v", err)
		}
		seen := map[string]bool{}
		for _, tl := range tools {
			if n := len([]rune(tl.Name())); n > maxMCPToolNameRunes {
				t.Fatalf("tool name %q has %d runes, limit is %d", tl.Name(), n, maxMCPToolNameRunes)
			}
			if seen[tl.Name()] {
				t.Fatalf("duplicate tool name %q", tl.Name())
			}
			seen[tl.Name()] = true
		}
		packAll(t, tools)
	})
}

func TestMCPToolsetsNeverTakesAKetName(t *testing.T) {
	// beta keeps alpha__search itself, so alpha's colliding search cannot
	// become alpha__search; it has to be suffixed instead.
	alpha := fakeServer("alpha", "search")
	beta := fakeServer("beta", "search", "alpha__search")
	group := newMCPToolsets([]mcpToolsetEntry{entry(alpha), entry(beta)})

	tools, err := group.Tools(nil)
	if err != nil {
		t.Fatalf("Tools: %v", err)
	}
	want := []string{"alpha__search__2", "beta__search", "alpha__search"}
	if got := toolNames(tools); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("names = %v, want %v", got, want)
	}
	if tools[2] != beta.tools[1] {
		t.Fatal("the tool whose name was kept is not the original tool")
	}
	packAll(t, tools)
}

func TestMCPToolsetsReportsServerErrors(t *testing.T) {
	failing := &failingMCPServer{name: "broken"}
	group := newMCPToolsets([]mcpToolsetEntry{entry(fakeServer("alpha", "search")), {server: "broken", toolset: failing}})

	if _, err := group.Tools(nil); err == nil || !strings.Contains(err.Error(), `mcp server "broken"`) {
		t.Fatalf("Tools error = %v, want it to name the server", err)
	}
}

type failingMCPServer struct{ name string }

func (s *failingMCPServer) Name() string { return s.name }

func (s *failingMCPServer) Tools(agent.ReadonlyContext) ([]tool.Tool, error) {
	return nil, errors.New("connection refused")
}

// newTestMCPServer serves the named tools from one real MCP server over
// streamable HTTP.
func newTestMCPServer(t *testing.T, toolNames ...string) string {
	t.Helper()
	srv := mcp.NewServer(&mcp.Implementation{Name: "agent-test", Version: "0"}, nil)
	for _, toolName := range toolNames {
		mcp.AddTool(srv, &mcp.Tool{Name: toolName, Description: toolName + " tool"}, func(_ context.Context, _ *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: toolName}}}, nil, nil
		})
	}
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil)
	httpSrv := httptest.NewServer(handler)
	t.Cleanup(httpSrv.Close)
	return httpSrv.URL
}

// TestBuildMCPToolsetsWithRealMCPToolsets runs the conversion against real
// mcptoolsets: the renames have to apply to the ADK tools they actually
// produce, not only to test doubles.
func TestBuildMCPToolsetsWithRealMCPToolsets(t *testing.T) {
	servers := []*agentsv1.MCPServer{
		{
			Name:      "alpha",
			Url:       newTestMCPServer(t, "search", "alpha_only"),
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		},
		{
			Name:      "beta",
			Url:       newTestMCPServer(t, "search"),
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		},
	}

	toolsets, err := buildMCPToolsets(context.Background(), servers, nil)
	if err != nil {
		t.Fatalf("buildMCPToolsets: %v", err)
	}
	if len(toolsets) != 1 {
		t.Fatalf("built %d toolsets, want 1", len(toolsets))
	}
	// Close before httptest tears its servers down: the streamable transport
	// holds a standalone SSE stream open.
	t.Cleanup(func() {
		if c, ok := toolsets[0].(interface{ Close() error }); ok {
			_ = c.Close()
		}
	})

	tools, err := toolsets[0].Tools(testReadonlyContext{context.Background()})
	if err != nil {
		t.Fatalf("Tools: %v", err)
	}

	// The servers' own listing order is theirs; the names are what matters.
	got := toolNames(tools)
	sort.Strings(got)
	want := []string{"alpha__search", "alpha_only", "beta__search"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("names = %v, want %v", got, want)
	}
	packAll(t, tools)
}
