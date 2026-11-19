package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool/toolconfirmation"

	"go.orx.me/apps/butter/internal/mcpcodemode"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestMCPCodeModeCallsRealMCPServer(t *testing.T) {
	type addIn struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	srv := mcp.NewServer(&mcp.Implementation{Name: "codemode-test", Version: "0"}, nil)
	mcp.AddTool(srv, &mcp.Tool{Name: "add", Description: "Add two integers."}, func(_ context.Context, _ *mcp.CallToolRequest, in addIn) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "sum"}}}, map[string]any{"sum": in.A + in.B}, nil
	})
	mcp.AddTool(srv, &mcp.Tool{Name: "words", Description: "Return a text blob."}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: strings.Repeat("lorem ipsum ", 1000)}}}, nil, nil
	})
	httpSrv := httptest.NewServer(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil))
	t.Cleanup(httpSrv.Close)

	set, err := buildMCPToolsets(context.Background(), []*agentsv1.MCPServer{{
		Name:      "calc",
		Url:       httpSrv.URL,
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Close before httptest tears the server down: the streamable transport
	// holds a standalone SSE stream open. Closing both layers is safe.
	wrapped := mcpcodemode.New(set, mcpcodemode.Limits{})
	t.Cleanup(func() { _ = wrapped.Close(); _ = set.Close() })

	tools, err := wrapped.Tools(testReadonlyContext{context.Background()})
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name() != mcpcodemode.ToolName {
		t.Fatalf("tools = %v, want only %q", toolNames(tools), mcpcodemode.ToolName)
	}
	if desc := tools[0].Description(); !strings.Contains(desc, "- add(a, b): Add two integers.") {
		t.Errorf("description does not list the MCP tools:\n%s", desc)
	}
	packAll(t, tools)

	run := tools[0].(interface {
		Run(agent.Context, any) (map[string]any, error)
	}).Run
	out, err := run(mcpToolContext{ctx: context.Background()}, map[string]any{"script": `
		const total = tools.add({a: 2, b: 40});
		const blob = tools.words();
		return {total, blobBytes: blob.length};
	`})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(out)
	var got struct {
		Result struct {
			Total     map[string]float64 `json:"total"`
			BlobBytes float64            `json:"blobBytes"`
		} `json:"result"`
		Error     string `json:"error"`
		ToolCalls int    `json:"tool_calls"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Error != "" {
		t.Fatalf("script failed: %s", got.Error)
	}
	if got.Result.Total["sum"] != 42 {
		t.Errorf("total = %v, want sum 42", got.Result.Total)
	}
	if got.Result.BlobBytes != float64(len("lorem ipsum ")*1000) {
		t.Errorf("blobBytes = %v", got.Result.BlobBytes)
	}
	if got.ToolCalls != 2 {
		t.Errorf("tool_calls = %d, want 2", got.ToolCalls)
	}
	// The 12 KB blob stayed in the sandbox: the model's view is small.
	if len(b) > 400 {
		t.Errorf("the model would see %d bytes: %s", len(b), b)
	}
}

// mcpToolContext is what a real MCP tool reads of agent.Context: the Go
// context and "no confirmation pending".
type mcpToolContext struct {
	agent.Context
	ctx context.Context
}

func (c mcpToolContext) Deadline() (time.Time, bool) { return c.ctx.Deadline() }
func (c mcpToolContext) Done() <-chan struct{}       { return c.ctx.Done() }
func (c mcpToolContext) Err() error                  { return c.ctx.Err() }
func (c mcpToolContext) Value(key any) any           { return c.ctx.Value(key) }

func (mcpToolContext) ToolConfirmation() *toolconfirmation.ToolConfirmation { return nil }

// TestNewFromProtoMCPCodeMode builds an LLM agent whose MCP server is behind
// code mode, and releases its resources: the wrapped toolset closes once.
func TestNewFromProtoMCPCodeMode(t *testing.T) {
	url := newTestMCPServer(t, "search")
	for _, codeMode := range []bool{false, true} {
		pb := &agentsv1.Agent{
			AgentId: "coder", Name: "coder",
			Config: &agentsv1.AgentConfig{
				Model:       "m1",
				McpCodeMode: codeMode,
				McpServers: []*agentsv1.MCPServer{{
					Name:      "srv",
					Url:       url,
					Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
				}},
			},
		}
		a, res, err := NewFromProto(context.Background(), pb, workflowProviders(), nil, nil, nil)
		if err != nil {
			t.Fatalf("code mode %v: %v", codeMode, err)
		}
		if a.Name() != "coder" {
			t.Errorf("agent name = %q", a.Name())
		}
		if err := res.Close(); err != nil {
			t.Errorf("code mode %v: closing resources: %v", codeMode, err)
		}
	}
}

func TestValidateMCPCodeMode(t *testing.T) {
	for _, tc := range []struct {
		typ     agentsv1.AgentType
		on      bool
		wantErr bool
	}{
		{agentsv1.AgentType_AGENT_TYPE_LLM, true, false},
		{agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED, true, false},
		{agentsv1.AgentType_AGENT_TYPE_LLM, false, false},
		{agentsv1.AgentType_AGENT_TYPE_LOOP, true, true},
		{agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL, true, true},
		{agentsv1.AgentType_AGENT_TYPE_PARALLEL, true, true},
		{agentsv1.AgentType_AGENT_TYPE_WORKFLOW, true, true},
		{agentsv1.AgentType_AGENT_TYPE_WORKFLOW, false, false},
	} {
		pb := &agentsv1.Agent{Name: "a", Type: tc.typ, Config: &agentsv1.AgentConfig{McpCodeMode: tc.on}}
		err := ValidateMCPCodeMode(pb)
		if (err != nil) != tc.wantErr {
			t.Errorf("type %v on=%v: err = %v, wantErr %v", tc.typ, tc.on, err, tc.wantErr)
		}
	}
}
