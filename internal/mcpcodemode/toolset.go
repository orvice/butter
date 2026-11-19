// Package mcpcodemode offers an agent's MCP tools through one `codemode` tool.
//
// Instead of listing every MCP tool in each model request and returning every
// raw result into the conversation, the model writes a short JavaScript
// snippet. The snippet runs in a sandbox (goja: no I/O, no modules, no network
// or filesystem globals) where the MCP tools are plain functions. It can
// discover tools, call several of them, filter and join their results, and
// only what it returns reaches the model. This is the same idea as Pi's
// Code Mode (docs/research/pi-mcp-codemode.md) applied to the MCP tools Butter
// itself manages for LLM agents; Pi agents keep their MCP on the ButterBox.
//
// The sandbox API, all synchronous:
//
//	tools.<name>(args)         call a tool, return its output; throws on failure
//	callTool(name, args)       same, for names that are not valid identifiers
//	callTools([{name, args}])  run calls concurrently; [{ok, result|error}]
//	searchTools(query)         [{name, description}] matching the query
//	describeTool(name)         {name, description, inputSchema}
//	ALL_TOOLS                  every tool name
//	console.log(...)           captured and returned with the result
//
// The script is a function body: `return` the value the model should see.
// Tool calls still go through the wrapped toolset, so renaming, tool filters,
// OAuth and per-call timeouts behave exactly as for direct tool calls.
package mcpcodemode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/toolutils"
	"google.golang.org/genai"
)

// ToolName is the name of the single tool the model sees.
const ToolName = "codemode"

// Limits bound one script. The zero value of a field means its default.
type Limits struct {
	// Timeout bounds the whole script, including its tool calls.
	Timeout time.Duration
	// MaxToolCalls bounds the tool calls one script may make.
	MaxToolCalls int
	// MaxScriptBytes bounds the script's source.
	MaxScriptBytes int
	// MaxOutputBytes bounds the JSON text of the returned value.
	MaxOutputBytes int
	// MaxLogBytes bounds the captured console output.
	MaxLogBytes int
	// MaxParallel bounds the concurrency of one callTools batch.
	MaxParallel int
	// MaxHeapGrowthBytes stops a script when the process's live heap has
	// grown by this much since the script started. See heapWatch.
	MaxHeapGrowthBytes uint64
}

const (
	defaultTimeout        = 60 * time.Second
	defaultMaxToolCalls   = 100
	defaultMaxScriptBytes = 32 << 10
	defaultMaxOutputBytes = 32 << 10
	defaultMaxLogBytes    = 8 << 10
	defaultMaxParallel    = 8
	defaultMaxHeapGrowth  = 512 << 20

	// Bounds of the catalog summary in the tool's description.
	maxListedTools       = 60
	maxListedToolRunes   = 160
	maxDescriptionBytes  = 8 << 10
	maxSearchResultCount = 20
)

func (l Limits) withDefaults() Limits {
	if l.Timeout <= 0 {
		l.Timeout = defaultTimeout
	}
	if l.MaxToolCalls <= 0 {
		l.MaxToolCalls = defaultMaxToolCalls
	}
	if l.MaxScriptBytes <= 0 {
		l.MaxScriptBytes = defaultMaxScriptBytes
	}
	if l.MaxOutputBytes <= 0 {
		l.MaxOutputBytes = defaultMaxOutputBytes
	}
	if l.MaxLogBytes <= 0 {
		l.MaxLogBytes = defaultMaxLogBytes
	}
	if l.MaxParallel <= 0 {
		l.MaxParallel = defaultMaxParallel
	}
	if l.MaxHeapGrowthBytes == 0 {
		l.MaxHeapGrowthBytes = defaultMaxHeapGrowth
	}
	return l
}

// Toolset wraps a toolset of MCP tools and exposes them as one `codemode`
// tool. Tool names must already be unique, which the agent's MCP toolset
// guarantees.
type Toolset struct {
	inner  tool.Toolset
	limits Limits
}

var (
	_ tool.Toolset = (*Toolset)(nil)
	_ io.Closer    = (*Toolset)(nil)
)

// New wraps inner. A nil inner yields a nil Toolset.
func New(inner tool.Toolset, limits Limits) *Toolset {
	if inner == nil {
		return nil
	}
	return &Toolset{inner: inner, limits: limits.withDefaults()}
}

func (t *Toolset) Name() string { return "mcp_codemode" }

// Close closes the wrapped toolset when it can be closed. The agent's
// resources close the wrapped toolset themselves, and it closes once, so
// calling both is safe.
func (t *Toolset) Close() error {
	if c, ok := t.inner.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Tools lists the wrapped tools and offers them as one tool. With no tool to
// offer, the model gets no tool at all.
func (t *Toolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	listed, err := t.inner.Tools(ctx)
	if err != nil {
		return nil, err
	}
	cat, err := newCatalog(listed)
	if err != nil {
		return nil, err
	}
	if len(cat.entries) == 0 {
		return nil, nil
	}
	return []tool.Tool{&codeTool{catalog: cat, limits: t.limits}}, nil
}

// entry is one callable MCP tool.
type entry struct {
	name        string
	description string
	schema      map[string]any
	params      string // "a, b?" summary of the input schema
	run         func(ctx agent.Context, args any) (map[string]any, error)
}

type catalog struct {
	entries []*entry
	byName  map[string]*entry
}

func newCatalog(listed []tool.Tool) (*catalog, error) {
	c := &catalog{byName: make(map[string]*entry, len(listed))}
	for _, t := range listed {
		declared, ok := t.(interface {
			Declaration() *genai.FunctionDeclaration
		})
		if !ok {
			return nil, fmt.Errorf("tool %q has no declaration", t.Name())
		}
		runnable, ok := t.(interface {
			Run(ctx agent.Context, args any) (map[string]any, error)
		})
		if !ok {
			return nil, fmt.Errorf("tool %q is not runnable", t.Name())
		}
		if _, dup := c.byName[t.Name()]; dup {
			return nil, fmt.Errorf("duplicate tool: %q", t.Name())
		}
		decl := declared.Declaration()
		e := &entry{name: t.Name(), description: t.Description(), run: runnable.Run}
		if decl != nil {
			if e.description == "" {
				e.description = decl.Description
			}
			e.schema = schemaMap(decl)
		}
		e.params = paramSummary(e.schema)
		c.entries = append(c.entries, e)
		c.byName[e.name] = e
	}
	return c, nil
}

// schemaMap returns a declaration's input schema as a plain JSON object.
func schemaMap(decl *genai.FunctionDeclaration) map[string]any {
	var raw any
	switch {
	case decl.ParametersJsonSchema != nil:
		raw = decl.ParametersJsonSchema
	case decl.Parameters != nil:
		raw = decl.Parameters
	default:
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

// paramSummary renders a schema's properties as "a, b?", required ones bare.
func paramSummary(schema map[string]any) string {
	props, _ := schema["properties"].(map[string]any)
	if len(props) == 0 {
		return ""
	}
	required := map[string]bool{}
	if req, ok := schema["required"].([]any); ok {
		for _, r := range req {
			if s, ok := r.(string); ok {
				required[s] = true
			}
		}
	}
	names := make([]string, 0, len(props))
	for n := range props {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool {
		if required[names[i]] != required[names[j]] {
			return required[names[i]]
		}
		return names[i] < names[j]
	})
	parts := make([]string, len(names))
	for i, n := range names {
		if required[n] {
			parts[i] = n
		} else {
			parts[i] = n + "?"
		}
	}
	return strings.Join(parts, ", ")
}

// search returns the tools whose name or description contains every word of
// query, best (name) matches first.
func (c *catalog) search(query string) []*entry {
	words := strings.Fields(strings.ToLower(query))
	type hit struct {
		e     *entry
		score int
	}
	var hits []hit
	for _, e := range c.entries {
		name := strings.ToLower(e.name)
		desc := strings.ToLower(e.description)
		score, all := 0, true
		for _, w := range words {
			switch {
			case strings.Contains(name, w):
				score += 2
			case strings.Contains(desc, w):
				score++
			default:
				all = false
			}
		}
		if all {
			hits = append(hits, hit{e, score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	if len(hits) > maxSearchResultCount {
		hits = hits[:maxSearchResultCount]
	}
	out := make([]*entry, len(hits))
	for i, h := range hits {
		out[i] = h.e
	}
	return out
}

// codeTool is the one tool the model sees.
type codeTool struct {
	catalog *catalog
	limits  Limits
}

var _ tool.Tool = (*codeTool)(nil)

func (t *codeTool) Name() string { return ToolName }

func (t *codeTool) IsLongRunning() bool { return false }

func (t *codeTool) Description() string {
	var b strings.Builder
	b.WriteString("Run a JavaScript snippet that uses this agent's MCP tools and return only what the snippet returns. ")
	b.WriteString("Prefer it to calling tools one at a time: look tools up, call several, loop, filter and combine results in code, then return a small result.\n")
	b.WriteString("The snippet is a function body in a sandbox without network, filesystem or modules; end it with `return <value>`.\n")
	b.WriteString("API (all synchronous): tools.<name>(args) calls a tool and returns its output (it throws on failure); ")
	b.WriteString("callTool(name, args) does the same by name; callTools([{name, args}]) runs calls concurrently and returns [{ok, result|error}]; ")
	b.WriteString("searchTools(query) -> [{name, description}]; describeTool(name) -> {name, description, inputSchema}; ALL_TOOLS lists every tool name; ")
	b.WriteString("console.log(...) output is returned with the result. Tool output is a string or a JSON value; JSON.parse a string result when it holds JSON.\n")
	fmt.Fprintf(&b, "Limits: %s, %d tool calls, %d bytes of returned JSON.\n", t.limits.Timeout, t.limits.MaxToolCalls, t.limits.MaxOutputBytes)

	b.WriteString("Tools:\n")
	listed := 0
	for _, e := range t.catalog.entries {
		if listed >= maxListedTools || b.Len() >= maxDescriptionBytes {
			break
		}
		line := "- " + e.name + "(" + e.params + ")"
		if d := firstLine(e.description); d != "" {
			line += ": " + d
		}
		b.WriteString(truncateRunes(line, maxListedToolRunes))
		b.WriteByte('\n')
		listed++
	}
	if rest := len(t.catalog.entries) - listed; rest > 0 {
		fmt.Fprintf(&b, "... and %d more; find them with searchTools(query) or ALL_TOOLS.\n", rest)
	}
	return b.String()
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return s
}

func truncateRunes(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	r := []rune(s)
	return string(r[:limit-1]) + "…"
}

func (t *codeTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        ToolName,
		Description: t.Description(),
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"script": map[string]any{
					"type":        "string",
					"description": "JavaScript function body. Use `return` to give the value back.",
				},
			},
			"required":             []string{"script"},
			"additionalProperties": false,
		},
	}
}

func (t *codeTool) ProcessRequest(_ agent.Context, req *model.LLMRequest) error {
	return toolutils.PackTool(req, t)
}

// Run executes the script. A failing script, or a call without a usable
// script, is reported in the result (with the logs it produced) rather than as
// a tool error, so the model can fix it.
func (t *codeTool) Run(ctx agent.Context, args any) (map[string]any, error) {
	script, err := scriptArg(args)
	if err == nil && len(script) > t.limits.MaxScriptBytes {
		err = fmt.Errorf("script is too long (%d bytes, max %d)", len(script), t.limits.MaxScriptBytes)
	}
	if err != nil {
		// Same shape as a failing script, so the model fixes its call.
		return map[string]any{"error": err.Error(), "tool_calls": 0}, nil
	}

	res := execute(ctx, t.catalog, t.limits, script)
	out := map[string]any{"tool_calls": res.toolCalls}
	if res.err != nil {
		out["error"] = res.err.Error()
	} else {
		out["result"] = res.value
		if res.truncated {
			out["truncated"] = true
		}
	}
	if len(res.logs) > 0 {
		out["logs"] = res.logs
	}
	return out, nil
}

func scriptArg(args any) (string, error) {
	m, ok := args.(map[string]any)
	if !ok {
		return "", errors.New("arguments must be an object with a `script` string")
	}
	script, ok := m["script"].(string)
	if !ok || strings.TrimSpace(script) == "" {
		return "", errors.New("`script` is required: a JavaScript function body that returns a value")
	}
	return script, nil
}

// deadlineContext gives a tool call the script's deadline while keeping the
// agent context (session, state, invocation) the tool expects.
type deadlineContext struct {
	agent.Context
	ctx context.Context
}

func (d deadlineContext) Deadline() (time.Time, bool) { return d.ctx.Deadline() }
func (d deadlineContext) Done() <-chan struct{}       { return d.ctx.Done() }
func (d deadlineContext) Err() error                  { return d.ctx.Err() }
func (d deadlineContext) Value(key any) any           { return d.ctx.Value(key) }
