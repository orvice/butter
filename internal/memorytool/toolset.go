// Package memorytool provides the search_memory and add_memory tools an LLM
// agent can use when its MemoryConfig enables them (ADR-0013 §7).
//
// Identity is never the model's to choose: the workspace, Agent ID, and
// provenance come from the turn's memory scope, which the runner derives
// from the invocation's root agent. The model supplies only a query or the
// content to remember, and a scope of "workspace" or "agent".
package memorytool

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"go.orx.me/apps/butter/internal/mem0"
	"go.orx.me/apps/butter/internal/runtime/mem0memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	"go.orx.me/apps/butter/internal/runtime/memoryhook"
)

const (
	// maxTopK caps how many memories one search may return.
	maxTopK = 20
	// maxResultChars caps the text of one search result.
	maxResultChars = 4000
	// maxContentRunes caps one add_memory call: memories are short facts.
	maxContentRunes = 2000

	searchTimeout = 10 * time.Second
	// addTimeout covers mem0's synchronous extraction LLM call.
	addTimeout = 30 * time.Second
)

// Toolset exposes the memory tools of one agent. Tools are offered only in
// turns where that agent is the invocation's root and its workspace has an
// enabled memory config; everywhere else the toolset is empty.
type Toolset struct {
	svc     *mem0memory.Service
	agentID string
	tools   []tool.Tool
}

// NewToolset builds the memory toolset for the agent with agentID.
func NewToolset(svc *mem0memory.Service, agentID string) (tool.Toolset, error) {
	if svc == nil || agentID == "" {
		return nil, nil
	}
	ts := &Toolset{svc: svc, agentID: agentID}
	search, err := functiontool.New(functiontool.Config{
		Name: "search_memory",
		Description: "Search long-term memory from earlier conversations. Use it when the user refers to something " +
			"that may have been discussed before, or when past preferences, decisions, or facts would help. " +
			"Scope 'workspace' (default) searches memory shared across this workspace; 'agent' searches memory specific to you.",
	}, ts.search)
	if err != nil {
		return nil, fmt.Errorf("create search_memory tool: %w", err)
	}
	add, err := functiontool.New(functiontool.Config{
		Name: "add_memory",
		Description: "Save a durable fact, preference, or decision to long-term memory so it can be recalled in future conversations. " +
			"State it on its own, e.g. \"The team deploys with pnpm\". Do not save secrets or transient details. " +
			"Scope 'workspace' (default) is shared with everyone in this workspace; 'agent' is specific to you and may be disabled.",
	}, ts.add)
	if err != nil {
		return nil, fmt.Errorf("create add_memory tool: %w", err)
	}
	ts.tools = []tool.Tool{search, add}
	return ts, nil
}

func (t *Toolset) Name() string { return "memory" }

func (t *Toolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	if !memoryhook.TurnFromContext(ctx).ToolsFor(t.agentID) {
		return nil, nil
	}
	return t.tools, nil
}

type searchArgs struct {
	Query string `json:"query" jsonschema_description:"What to look for, in natural language."`
	Scope string `json:"scope,omitempty" jsonschema_description:"'workspace' (default) or 'agent'."`
	TopK  int    `json:"top_k,omitempty" jsonschema_description:"Maximum number of memories to return, 1-20. Defaults to the agent's configured value."`
}

type memoryItem struct {
	Memory string `json:"memory"`
	Date   string `json:"date,omitempty"`
}

type searchResult struct {
	Scope     string       `json:"scope"`
	Memories  []memoryItem `json:"memories"`
	Truncated bool         `json:"truncated,omitempty"`
}

type addArgs struct {
	Content string `json:"content" jsonschema_description:"The fact, preference, or decision to remember, stated on its own."`
	Scope   string `json:"scope,omitempty" jsonschema_description:"'workspace' (default) or 'agent'."`
}

type addResult struct {
	Scope string `json:"scope"`
	// Stored lists what mem0 extracted and stored; empty when it judged the
	// content a duplicate or not worth keeping.
	Stored []string `json:"stored"`
}

// turnFor returns the turn's scope for a tool call, refusing calls that
// arrive outside a turn this toolset may serve.
func (t *Toolset) turnFor(ctx agent.Context) (*memoryhook.Turn, mem0memory.Scope, error) {
	turn := memoryhook.TurnFromContext(ctx)
	if !turn.ToolsFor(t.agentID) {
		return nil, mem0memory.Scope{}, errors.New("memory is not available in this conversation")
	}
	scope := turn.Scope()
	scope.InvocationID = ctx.InvocationID()
	return turn, scope, nil
}

func parseTarget(scope string) (mem0memory.Target, error) {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "", "workspace":
		return mem0memory.TargetWorkspace, nil
	case "agent":
		return mem0memory.TargetAgent, nil
	default:
		return 0, fmt.Errorf("scope must be 'workspace' or 'agent', got %q", scope)
	}
}

func (t *Toolset) search(ctx agent.Context, args searchArgs) (searchResult, error) {
	turn, scope, err := t.turnFor(ctx)
	if err != nil {
		return searchResult{}, err
	}
	target, err := parseTarget(args.Scope)
	if err != nil {
		return searchResult{}, err
	}
	query := strings.TrimSpace(args.Query)
	if query == "" {
		return searchResult{}, errors.New("query is required")
	}
	opts := mem0memory.SearchOptions{}
	mc := turn.Config()
	if mc.TopK != nil {
		opts.TopK = int(mc.GetTopK())
	}
	if args.TopK > 0 {
		opts.TopK = min(args.TopK, maxTopK)
	}
	if mc.Threshold != nil {
		th := float64(mc.GetThreshold())
		opts.Threshold = &th
	}

	sctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()
	found, err := t.svc.Search(sctx, scope, target, query, opts)
	if err != nil {
		return searchResult{}, readable("search", err)
	}
	return formatSearch(target, found, maxResultChars), nil
}

func (t *Toolset) add(ctx agent.Context, args addArgs) (addResult, error) {
	turn, scope, err := t.turnFor(ctx)
	if err != nil {
		return addResult{}, err
	}
	target, err := parseTarget(args.Scope)
	if err != nil {
		return addResult{}, err
	}
	if target == mem0memory.TargetAgent && !turn.Config().GetAllowAgentScopeWrite() {
		return addResult{}, errors.New("writing agent memory is not enabled for this agent; save it to workspace memory instead")
	}
	content := strings.TrimSpace(args.Content)
	if content == "" {
		return addResult{}, errors.New("content is required")
	}
	if utf8.RuneCountInString(content) > maxContentRunes {
		return addResult{}, fmt.Errorf("content is too long (%d characters, max %d); save one short fact at a time", utf8.RuneCountInString(content), maxContentRunes)
	}

	actx, cancel := context.WithTimeout(ctx, addTimeout)
	defer cancel()
	stored, err := t.svc.Add(actx, scope, target, []mem0.Message{{Role: "user", Content: content}})
	if err != nil {
		return addResult{}, readable("save", err)
	}
	out := addResult{Scope: target.String(), Stored: []string{}}
	for _, r := range stored {
		if m := strings.TrimSpace(r.Memory); m != "" {
			out.Stored = append(out.Stored, m)
		}
	}
	return out, nil
}

// readable turns a memory failure into an error the model can act on. The
// turn continues either way: ADK hands tool errors back to the model.
func readable(action string, err error) error {
	if errors.Is(err, memoryconn.ErrNotConfigured) {
		return errors.New("workspace memory is not configured")
	}
	return fmt.Errorf("memory %s failed, continue without it: %w", action, err)
}

// formatSearch renders search hits within maxChars of memory text,
// dropping whole entries that do not fit.
func formatSearch(target mem0memory.Target, found []mem0.Memory, maxChars int) searchResult {
	out := searchResult{Scope: target.String(), Memories: []memoryItem{}}
	budget := maxChars
	for _, m := range found {
		text := strings.TrimSpace(m.Memory)
		if text == "" {
			continue
		}
		cost := utf8.RuneCountInString(text)
		if cost > budget {
			out.Truncated = true
			break
		}
		budget -= cost
		out.Memories = append(out.Memories, memoryItem{Memory: text, Date: date(m.CreatedAt)})
	}
	return out
}

func date(createdAt string) string {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999", "2006-01-02"} {
		if ts, err := time.Parse(layout, createdAt); err == nil {
			return ts.Format("2006-01-02")
		}
	}
	return ""
}
