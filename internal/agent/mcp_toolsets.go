package agent

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/toolutils"
	"google.golang.org/genai"
)

// maxMCPToolNameRunes bounds a qualified name to what model providers accept
// (64 characters for Gemini and OpenAI function names).
const maxMCPToolNameRunes = 64

// mcpToolsetEntry is one MCP server's toolset under the name the agent's
// config gives that server.
type mcpToolsetEntry struct {
	server  string
	toolset tool.Toolset
}

// mcpToolsets exposes every MCP server of one agent as a single ADK toolset.
//
// ADK packs tools by name into one model request and refuses two tools with
// the same name ("duplicate tool: ..."), so two MCP servers that both expose,
// say, search used to fail the agent's whole run. Names are therefore assigned
// across all servers at once, per Tools call, deterministically:
//
//   - a name exactly one exposed tool carries is left untouched;
//   - a name carried by more than one exposed tool — in different servers or
//     twice in one — becomes <server>__<tool> for every one of them, where
//     <server> is the configured server name sanitized to what providers
//     accept, and __2, __3… is appended where even that repeats;
//   - a kept name is never taken away from its tool: a renamed tool that would
//     land on one is suffixed instead.
//
// The assignment follows the agent's server order and each server's tool
// order, so the same configuration and tool lists always produce the same
// names. A renamed tool wraps the original: Name and Declaration report the
// new name, Run and every other behavior stay the toolset's.
//
// The set closes its servers' toolsets, since it replaces them in the agent
// config; ADK's mcptoolset requires its owner to close it.
type mcpToolsets struct {
	entries []mcpToolsetEntry
}

func newMCPToolsets(entries []mcpToolsetEntry) *mcpToolsets {
	return &mcpToolsets{entries: entries}
}

var (
	_ tool.Toolset = (*mcpToolsets)(nil)
	_ io.Closer    = (*mcpToolsets)(nil)
)

func (*mcpToolsets) Name() string { return "mcp_tools" }

// Close closes every server toolset that can be closed.
func (m *mcpToolsets) Close() error {
	var errs []error
	for _, e := range m.entries {
		if c, ok := e.toolset.(io.Closer); ok {
			if err := c.Close(); err != nil {
				errs = append(errs, fmt.Errorf("mcp server %q: %w", e.server, err))
			}
		}
	}
	return errors.Join(errs...)
}

// Tools lists every server's tools under names that can be packed together.
func (m *mcpToolsets) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	listed := make([][]tool.Tool, len(m.entries))
	counts := make(map[string]int)
	total := 0
	for i, e := range m.entries {
		tools, err := e.toolset.Tools(ctx)
		if err != nil {
			return nil, fmt.Errorf("mcp server %q: %w", e.server, err)
		}
		listed[i] = tools
		total += len(tools)
		for _, t := range tools {
			counts[t.Name()]++
		}
	}

	// Reserve the names that need no qualification first, so qualifying a
	// colliding tool can never take a name another tool keeps as is.
	names := make([][]string, len(listed))
	used := make(map[string]bool, total)
	for i, tools := range listed {
		names[i] = make([]string, len(tools))
		for j, t := range tools {
			if name := t.Name(); name != "" && counts[name] == 1 {
				names[i][j] = name
				used[name] = true
			}
		}
	}

	out := make([]tool.Tool, 0, total)
	for i, e := range m.entries {
		prefix := mcpToolNamePrefix(e.server)
		for j, t := range listed[i] {
			if names[i][j] != "" {
				out = append(out, t)
				continue
			}
			name := uniqueToolName(prefix+"__"+t.Name(), used)
			renamed, err := renameMCPTool(t, name)
			if err != nil {
				return nil, fmt.Errorf("mcp server %q: %w", e.server, err)
			}
			out = append(out, renamed)
		}
	}
	return out, nil
}

// mcpToolNamePrefix turns a configured server name into the provider-safe
// part a qualified tool name starts with.
func mcpToolNamePrefix(server string) string {
	var b strings.Builder
	for _, r := range server {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	prefix := strings.Trim(b.String(), "_")
	if prefix == "" {
		prefix = "mcp"
	}
	// A function name has to start with a letter or underscore.
	if prefix[0] >= '0' && prefix[0] <= '9' {
		prefix = "mcp_" + prefix
	}
	return prefix
}

// uniqueToolName returns candidate, or candidate with __2, __3… appended if it
// is already taken, bounded to maxMCPToolNameRunes, and marks the result used.
func uniqueToolName(candidate string, used map[string]bool) string {
	name := truncateToolName(candidate, maxMCPToolNameRunes)
	if !used[name] {
		used[name] = true
		return name
	}
	for n := 2; ; n++ {
		suffix := fmt.Sprintf("__%d", n)
		stem := truncateToolName(candidate, maxMCPToolNameRunes-len([]rune(suffix)))
		name = stem + suffix
		if !used[name] {
			used[name] = true
			return name
		}
	}
}

func truncateToolName(name string, limit int) string {
	runes := []rune(name)
	if len(runes) <= limit {
		return name
	}
	return string(runes[:limit])
}

// renamedTool exposes an MCP tool under a different name. Everything but the
// name delegates to the original tool, so argument schemas, descriptions,
// long-running and confirmation behavior stay the toolset's; ProcessRequest
// packs the renamed declaration that ADK sends to the model.
type renamedTool struct {
	tool.Tool
	name string
	decl *genai.FunctionDeclaration
	run  func(ctx agent.Context, args any) (map[string]any, error)
}

var _ tool.Tool = (*renamedTool)(nil)

func renameMCPTool(t tool.Tool, name string) (tool.Tool, error) {
	declared, ok := t.(interface {
		Declaration() *genai.FunctionDeclaration
	})
	if !ok {
		return nil, fmt.Errorf("tool %q has no declaration", t.Name())
	}
	decl := declared.Declaration()
	if decl == nil {
		return nil, fmt.Errorf("tool %q has no declaration", t.Name())
	}
	runnable, ok := t.(interface {
		Run(ctx agent.Context, args any) (map[string]any, error)
	})
	if !ok {
		return nil, fmt.Errorf("tool %q is not runnable", t.Name())
	}

	// Copy so the original tool's declaration keeps its own name.
	renamed := *decl
	renamed.Name = name
	return &renamedTool{
		Tool: t,
		name: name,
		decl: &renamed,
		run:  runnable.Run,
	}, nil
}

func (t *renamedTool) Name() string { return t.name }

func (t *renamedTool) Declaration() *genai.FunctionDeclaration { return t.decl }

func (t *renamedTool) ProcessRequest(_ agent.Context, req *model.LLMRequest) error {
	return toolutils.PackTool(req, t)
}

func (t *renamedTool) Run(ctx agent.Context, args any) (map[string]any, error) {
	return t.run(ctx, args)
}
