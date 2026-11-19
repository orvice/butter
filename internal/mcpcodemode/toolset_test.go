package mcpcodemode

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"
)

type fakeTool struct {
	name, desc string
	schema     map[string]any
	run        func(ctx agent.Context, args map[string]any) (map[string]any, error)
	calls      atomic.Int64
}

func (f *fakeTool) Name() string        { return f.name }
func (f *fakeTool) Description() string { return f.desc }
func (f *fakeTool) IsLongRunning() bool { return false }
func (f *fakeTool) ProcessRequest(agent.Context, *model.LLMRequest) error {
	return nil
}
func (f *fakeTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{Name: f.name, Description: f.desc, ParametersJsonSchema: f.schema}
}
func (f *fakeTool) Run(ctx agent.Context, args any) (map[string]any, error) {
	f.calls.Add(1)
	m, _ := args.(map[string]any)
	return f.run(ctx, m)
}

type fakeToolset struct {
	tools []tool.Tool
	err   error
}

func (f *fakeToolset) Name() string { return "fake" }
func (f *fakeToolset) Tools(agent.ReadonlyContext) ([]tool.Tool, error) {
	return f.tools, f.err
}

func testCtx(parent context.Context) agent.Context {
	return deadlineContext{ctx: parent}
}

func output(v any) func(agent.Context, map[string]any) (map[string]any, error) {
	return func(agent.Context, map[string]any) (map[string]any, error) {
		return map[string]any{"output": v}, nil
	}
}

func issueTools() (*fakeTool, *fakeTool, *fakeTool) {
	list := &fakeTool{
		name: "tracker__list_issues", desc: "List issues.\nSecond line.",
		schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"state": map[string]any{"type": "string"}, "limit": map[string]any{"type": "integer"}},
			"required":   []any{"state"},
		},
		run: func(_ agent.Context, args map[string]any) (map[string]any, error) {
			if args["state"] != "open" {
				return nil, fmt.Errorf("unknown state %v", args["state"])
			}
			return map[string]any{"output": []any{
				map[string]any{"id": 1.0, "title": "crash on start", "labels": []any{"bug"}},
				map[string]any{"id": 2.0, "title": "add dark mode", "labels": []any{"feature"}},
				map[string]any{"id": 3.0, "title": "leak in cache", "labels": []any{"bug"}},
			}}, nil
		},
	}
	get := &fakeTool{
		name: "tracker__get_issue", desc: "Get one issue by id.",
		schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"id": map[string]any{"type": "integer"}},
			"required":   []any{"id"},
		},
		run: func(_ agent.Context, args map[string]any) (map[string]any, error) {
			id, _ := args["id"].(float64)
			if id == 3 {
				return nil, errors.New("issue 3 is locked")
			}
			return map[string]any{"output": fmt.Sprintf("issue %d body", int(id))}, nil
		},
	}
	echo := &fakeTool{name: "echo", desc: "Echo.", run: output("pong")}
	return list, get, echo
}

func newTool(t *testing.T, limits Limits, tools ...tool.Tool) *codeTool {
	t.Helper()
	got, err := New(&fakeToolset{tools: tools}, limits).Tools(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d tools, want 1", len(got))
	}
	return got[0].(*codeTool)
}

func run(t *testing.T, ct *codeTool, script string) map[string]any {
	t.Helper()
	out, err := ct.Run(testCtx(context.Background()), map[string]any{"script": script})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return out
}

func TestScriptCallsToolsAndReturnsOnlyItsResult(t *testing.T) {
	list, get, echo := issueTools()
	ct := newTool(t, Limits{}, list, get, echo)

	out := run(t, ct, `
		const open = tools.tracker__list_issues({state: "open"});
		const bugs = open.filter(i => i.labels.includes("bug")).map(i => i.title);
		console.log("found", bugs.length);
		return {bugs, ping: tools.echo()};
	`)
	if out["error"] != nil {
		t.Fatalf("error: %v", out["error"])
	}
	res := out["result"].(map[string]any)
	bugs := res["bugs"].([]any)
	if len(bugs) != 2 || bugs[0] != "crash on start" || bugs[1] != "leak in cache" {
		t.Errorf("bugs = %v", bugs)
	}
	if res["ping"] != "pong" {
		t.Errorf("ping = %v", res["ping"])
	}
	if out["tool_calls"] != 2 {
		t.Errorf("tool_calls = %v, want 2", out["tool_calls"])
	}
	if logs := out["logs"].([]string); len(logs) != 1 || logs[0] != "found 2" {
		t.Errorf("logs = %v", logs)
	}
}

func TestToolFailureThrowsAndCanBeCaught(t *testing.T) {
	list, get, echo := issueTools()
	ct := newTool(t, Limits{}, list, get, echo)

	out := run(t, ct, `
		try { tools.tracker__get_issue({id: 3}); } catch (e) { return "caught: " + e.message; }
	`)
	if s, _ := out["result"].(string); !strings.Contains(s, "issue 3 is locked") {
		t.Errorf("result = %v, error = %v", out["result"], out["error"])
	}

	out = run(t, ct, `console.log("before"); tools.tracker__get_issue({id: 3});`)
	msg, _ := out["error"].(string)
	if !strings.Contains(msg, "issue 3 is locked") {
		t.Errorf("error = %q", msg)
	}
	if logs, _ := out["logs"].([]string); len(logs) != 1 || logs[0] != "before" {
		t.Errorf("logs = %v, want the log written before the failure", out["logs"])
	}
}

func TestCallToolsRunsBatchAndReportsEachOutcome(t *testing.T) {
	list, get, echo := issueTools()
	ct := newTool(t, Limits{}, list, get, echo)

	out := run(t, ct, `
		return callTools([1, 2, 3].map(id => ({name: "tracker__get_issue", args: {id}})));
	`)
	items := out["result"].([]any)
	if len(items) != 3 {
		t.Fatalf("items = %v (error %v)", items, out["error"])
	}
	for i, wantOK := range []bool{true, true, false} {
		it := items[i].(map[string]any)
		if it["ok"] != wantOK {
			t.Errorf("item %d = %v, want ok=%v", i, it, wantOK)
		}
	}
	if !strings.Contains(items[2].(map[string]any)["error"].(string), "locked") {
		t.Errorf("item 2 = %v", items[2])
	}
	if items[0].(map[string]any)["result"] != "issue 1 body" {
		t.Errorf("item 0 = %v", items[0])
	}
}

func TestDiscovery(t *testing.T) {
	list, get, echo := issueTools()
	ct := newTool(t, Limits{}, list, get, echo)

	out := run(t, ct, `
		return {
			all: ALL_TOOLS,
			found: searchTools("issue by id").map(t => t.name),
			none: searchTools("nothing like this").length,
			desc: describeTool("tracker__get_issue").inputSchema.required,
		};
	`)
	res := out["result"].(map[string]any)
	if got := fmt.Sprint(res["all"]); got != "[tracker__list_issues tracker__get_issue echo]" {
		t.Errorf("all = %v", got)
	}
	if got := fmt.Sprint(res["found"]); got != "[tracker__get_issue]" {
		t.Errorf("found = %v", got)
	}
	if res["none"] != 0.0 {
		t.Errorf("none = %v", res["none"])
	}
	if got := fmt.Sprint(res["desc"]); got != "[id]" {
		t.Errorf("required = %v", got)
	}
}

func TestUnknownToolIsAScriptError(t *testing.T) {
	list, _, _ := issueTools()
	ct := newTool(t, Limits{}, list)
	out := run(t, ct, `return callTool("nope", {});`)
	if msg, _ := out["error"].(string); !strings.Contains(msg, `unknown tool "nope"`) {
		t.Errorf("error = %v", out["error"])
	}
}

func TestSandboxHasNoIO(t *testing.T) {
	_, _, echo := issueTools()
	ct := newTool(t, Limits{}, echo)
	out := run(t, ct, `
		return ["require", "process", "fetch", "XMLHttpRequest", "fs", "setTimeout", "WebSocket", "module"]
			.filter(n => typeof globalThis[n] !== "undefined");
	`)
	if got := fmt.Sprint(out["result"]); got != "[]" {
		t.Errorf("globals present: %v (error %v)", got, out["error"])
	}
}

func TestTimeoutStopsRunawayScript(t *testing.T) {
	_, _, echo := issueTools()
	ct := newTool(t, Limits{Timeout: 100 * time.Millisecond}, echo)

	start := time.Now()
	out := run(t, ct, `while (true) {}`)
	if time.Since(start) > 5*time.Second {
		t.Fatalf("script ran for %v", time.Since(start))
	}
	if msg, _ := out["error"].(string); !strings.Contains(msg, "timed out") {
		t.Errorf("error = %v", out["error"])
	}
}

func TestTimeoutCancelsToolInFlight(t *testing.T) {
	slow := &fakeTool{name: "slow", desc: "Slow.", run: func(ctx agent.Context, _ map[string]any) (map[string]any, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	ct := newTool(t, Limits{Timeout: 100 * time.Millisecond}, slow)

	start := time.Now()
	out := run(t, ct, `return tools.slow();`)
	if time.Since(start) > 5*time.Second {
		t.Fatalf("script ran for %v", time.Since(start))
	}
	if out["error"] == nil {
		t.Errorf("want an error, got %v", out)
	}
}

func TestParentCancellationStopsScript(t *testing.T) {
	_, _, echo := issueTools()
	ct := newTool(t, Limits{}, echo)
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	out, err := ct.Run(testCtx(ctx), map[string]any{"script": `while (true) {}`})
	if err != nil {
		t.Fatal(err)
	}
	if msg, _ := out["error"].(string); !strings.Contains(msg, "cancelled") {
		t.Errorf("error = %v", out["error"])
	}
}

func TestToolCallLimit(t *testing.T) {
	_, _, echo := issueTools()
	ct := newTool(t, Limits{MaxToolCalls: 3}, echo)

	out := run(t, ct, `for (let i = 0; i < 10; i++) tools.echo(); return "done";`)
	if msg, _ := out["error"].(string); !strings.Contains(msg, "tool call limit") {
		t.Errorf("error = %v", out["error"])
	}
	if echo.calls.Load() != 3 {
		t.Errorf("echo ran %d times, want 3", echo.calls.Load())
	}

	echo.calls.Store(0)
	out = run(t, ct, `return callTools([1,2,3,4].map(() => ({name: "echo"})));`)
	if msg, _ := out["error"].(string); !strings.Contains(msg, "tool call limit") {
		t.Errorf("batch error = %v", out["error"])
	}
	if echo.calls.Load() != 0 {
		t.Errorf("an over-budget batch ran %d calls, want none", echo.calls.Load())
	}
}

func TestOutputIsBounded(t *testing.T) {
	_, _, echo := issueTools()
	ct := newTool(t, Limits{MaxOutputBytes: 100, MaxLogBytes: 50}, echo)

	out := run(t, ct, `
		for (let i = 0; i < 20; i++) console.log("line number", i);
		return "é".repeat(500);
	`)
	if out["truncated"] != true {
		t.Errorf("not marked truncated: %v", out)
	}
	if s, _ := out["result"].(string); !strings.Contains(s, "output truncated") || len(s) > 200 {
		t.Errorf("result = %q", s)
	}
	logs := out["logs"].([]string)
	total := 0
	for _, l := range logs {
		total += len(l)
	}
	if total > 50+len("… [logs truncated]") {
		t.Errorf("logs hold %d bytes: %v", total, logs)
	}
}

func TestSyntaxErrorAndBadArguments(t *testing.T) {
	_, _, echo := issueTools()
	ct := newTool(t, Limits{}, echo)

	out := run(t, ct, `return (;`)
	if out["error"] == nil {
		t.Errorf("no error for a syntax error: %v", out)
	}

	// A call without a usable script is a result the model can act on, not a
	// tool error.
	for _, args := range []any{nil, map[string]any{}, map[string]any{"script": "  "}, map[string]any{"script": 3}} {
		out, err := ct.Run(testCtx(context.Background()), args)
		if err != nil || out["error"] == nil {
			t.Errorf("Run(%v) = %v, %v; want an error result", args, out, err)
		}
	}

	small := newTool(t, Limits{MaxScriptBytes: 10}, echo)
	out, err := small.Run(testCtx(context.Background()), map[string]any{"script": strings.Repeat("a", 11)})
	if err != nil || out["error"] == nil {
		t.Errorf("an oversize script gave %v, %v; want an error result", out, err)
	}
}

func TestToolIgnoringItsContextCannotOutlastTheDeadline(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	stuck := &fakeTool{name: "stuck", desc: "Ignores ctx.", run: func(agent.Context, map[string]any) (map[string]any, error) {
		<-release
		return map[string]any{"output": "late"}, nil
	}}
	ct := newTool(t, Limits{Timeout: 100 * time.Millisecond}, stuck)

	start := time.Now()
	out := run(t, ct, `return callTools([{name: "stuck"}]).concat([tools.stuck()]);`)
	if time.Since(start) > 5*time.Second {
		t.Fatalf("script ran for %v", time.Since(start))
	}
	if out["error"] == nil && out["result"] == nil {
		t.Errorf("no outcome: %v", out)
	}
}

func TestHeapGrowthStopsScript(t *testing.T) {
	_, _, echo := issueTools()
	ct := newTool(t, Limits{Timeout: 30 * time.Second, MaxHeapGrowthBytes: 64 << 20}, echo)

	start := time.Now()
	out := run(t, ct, `
		const hold = [];
		for (let i = 0; ; i++) hold.push("x".repeat(10000) + i);
	`)
	if time.Since(start) > 20*time.Second {
		t.Fatalf("script ran for %v", time.Since(start))
	}
	if msg, _ := out["error"].(string); !strings.Contains(msg, "heap grew") {
		t.Errorf("error = %v", out["error"])
	}
}

func TestToolArgumentsArePlainJSON(t *testing.T) {
	var got map[string]any
	capture := &fakeTool{name: "capture", desc: "Capture.", run: func(_ agent.Context, args map[string]any) (map[string]any, error) {
		got = args
		return map[string]any{"output": "ok"}, nil
	}}
	ct := newTool(t, Limits{}, capture)
	out := run(t, ct, `return tools.capture({a: 1, b: ["x", {c: null}], d: undefined});`)
	if out["error"] != nil {
		t.Fatalf("error: %v", out["error"])
	}
	if fmt.Sprint(got) != "map[a:1 b:[x map[c:<nil>]]]" {
		t.Errorf("tool got %v", got)
	}

	out = run(t, ct, `return tools.capture("not an object");`)
	if out["error"] == nil {
		t.Errorf("a string argument was accepted: %v", out)
	}
}

func TestStructuredOutputKeepsOtherKeys(t *testing.T) {
	raw := &fakeTool{name: "raw", desc: "Raw.", run: func(agent.Context, map[string]any) (map[string]any, error) {
		return map[string]any{"output": "x", "extra": 1}, nil
	}}
	ct := newTool(t, Limits{}, raw)
	out := run(t, ct, `return tools.raw();`)
	if fmt.Sprint(out["result"]) != "map[extra:1 output:x]" {
		t.Errorf("result = %v", out["result"])
	}
}

func TestToolsetOffersOneToolWithCatalogInDescription(t *testing.T) {
	list, get, echo := issueTools()
	ts := New(&fakeToolset{tools: []tool.Tool{list, get, echo}}, Limits{})
	got, err := ts.Tools(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name() != ToolName {
		t.Fatalf("tools = %v", got)
	}
	desc := got[0].Description()
	for _, want := range []string{
		"- tracker__list_issues(state, limit?): List issues.",
		"- tracker__get_issue(id): Get one issue by id.",
		"- echo(): Echo.",
	} {
		if !strings.Contains(desc, want) {
			t.Errorf("description lacks %q:\n%s", want, desc)
		}
	}
	if strings.Contains(desc, "Second line") {
		t.Error("description carries more than a tool's first line")
	}
	decl := got[0].(interface {
		Declaration() *genai.FunctionDeclaration
	}).Declaration()
	if decl.Name != ToolName || decl.Description != desc {
		t.Errorf("declaration = %+v", decl)
	}
}

func TestLargeCatalogIsSummarized(t *testing.T) {
	var tools []tool.Tool
	for i := range 200 {
		tools = append(tools, &fakeTool{name: fmt.Sprintf("tool_%03d", i), desc: strings.Repeat("d", 500), run: output("x")})
	}
	got, err := New(&fakeToolset{tools: tools}, Limits{}).Tools(nil)
	if err != nil {
		t.Fatal(err)
	}
	desc := got[0].Description()
	if len(desc) > 3*maxDescriptionBytes {
		t.Errorf("description is %d bytes", len(desc))
	}
	listed := strings.Count(desc, "\n- tool_")
	if listed == 0 || listed > maxListedTools {
		t.Errorf("listed %d tools", listed)
	}
	if want := fmt.Sprintf("... and %d more", 200-listed); !strings.Contains(desc, want) {
		t.Errorf("description lacks %q", want)
	}
}

func TestNoToolsOffersNothing(t *testing.T) {
	got, err := New(&fakeToolset{}, Limits{}).Tools(nil)
	if err != nil || len(got) != 0 {
		t.Errorf("Tools() = %v, %v; want none", got, err)
	}
	wantErr := errors.New("listing failed")
	if _, err := New(&fakeToolset{err: wantErr}, Limits{}).Tools(nil); !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}

func TestDuplicateToolNamesAreRefused(t *testing.T) {
	_, _, echo := issueTools()
	if _, err := New(&fakeToolset{tools: []tool.Tool{echo, echo}}, Limits{}).Tools(nil); err == nil {
		t.Error("duplicate names were accepted")
	}
}

type closingToolset struct {
	fakeToolset
	closed int
}

func (c *closingToolset) Close() error { c.closed++; return nil }

func TestCloseClosesWrappedToolset(t *testing.T) {
	inner := &closingToolset{}
	if err := New(inner, Limits{}).Close(); err != nil || inner.closed != 1 {
		t.Errorf("Close() = %v, closed %d times", err, inner.closed)
	}
	if err := New(&fakeToolset{}, Limits{}).Close(); err != nil {
		t.Errorf("Close() of a non-closer = %v", err)
	}
}

func TestToolPanicBecomesScriptError(t *testing.T) {
	boom := &fakeTool{name: "boom", desc: "Boom.", run: func(agent.Context, map[string]any) (map[string]any, error) {
		panic("kaput")
	}}
	ct := newTool(t, Limits{}, boom)
	out := run(t, ct, `return tools.boom();`)
	if msg, _ := out["error"].(string); !strings.Contains(msg, "kaput") {
		t.Errorf("error = %v", out["error"])
	}
	out = run(t, ct, `return callTools([{name: "boom"}]);`)
	items, _ := out["result"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["ok"] != false {
		t.Errorf("batch = %v", out)
	}
}
