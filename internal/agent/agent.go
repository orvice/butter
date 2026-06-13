package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/remoteagent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/agent/workflowagents/parallelagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"

	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const defaultMCPTimeout = 5 * time.Second

// MCPHTTPClientFactory creates the HTTP client used by MCP transports. A nil
// client means the MCP SDK should use its default client.
type MCPHTTPClientFactory interface {
	HTTPClientForMCP(ctx context.Context, srv *agentsv1.MCPServer) (*http.Client, error)
}

// ToolsetFactory creates built-in, per-agent toolsets such as agent_files.
type ToolsetFactory func(ctx context.Context, pb *agentsv1.Agent) ([]tool.Toolset, error)

// NewFromProto creates an ADK agent from an agentsv1.Agent proto config.
// providers is the list of model provider mappings used to resolve LLM backends.
// mcpRegistry is the shared MCP server config pool; agents reference entries by ID.
// remoteAgentRegistry is the shared remote agent config pool; agents reference entries by ID.
func NewFromProto(ctx context.Context, pb *agentsv1.Agent, providers []agentsv1.ModelProvider, mcpRegistry []agentsv1.MCPServer, remoteAgentRegistry []agentsv1.RemoteAgent, daemonRegistry *daemon.Registry) (agent.Agent, error) {
	return NewFromProtoWithMCPHTTPClientFactory(ctx, pb, providers, mcpRegistry, remoteAgentRegistry, daemonRegistry, nil)
}

// NewFromProtoWithMCPHTTPClientFactory creates an ADK agent with a custom MCP
// HTTP client factory for static headers, OAuth2 bearer injection, and tests.
func NewFromProtoWithMCPHTTPClientFactory(ctx context.Context, pb *agentsv1.Agent, providers []agentsv1.ModelProvider, mcpRegistry []agentsv1.MCPServer, remoteAgentRegistry []agentsv1.RemoteAgent, daemonRegistry *daemon.Registry, httpFactory MCPHTTPClientFactory) (agent.Agent, error) {
	return NewFromProtoWithToolsetFactory(ctx, pb, providers, mcpRegistry, remoteAgentRegistry, daemonRegistry, httpFactory, nil)
}

// NewFromProtoWithToolsetFactory creates an ADK agent with custom MCP HTTP and
// built-in toolset factories.
func NewFromProtoWithToolsetFactory(ctx context.Context, pb *agentsv1.Agent, providers []agentsv1.ModelProvider, mcpRegistry []agentsv1.MCPServer, remoteAgentRegistry []agentsv1.RemoteAgent, daemonRegistry *daemon.Registry, httpFactory MCPHTTPClientFactory, toolsetFactory ToolsetFactory) (agent.Agent, error) {
	if pb == nil {
		return nil, fmt.Errorf("agent config is nil")
	}

	// Resolve shared MCP servers and merge with inline ones. The merged list
	// is passed alongside the proto so the shared config is never mutated.
	mcpServers, err := resolveMCPServers(pb, mcpRegistry)
	if err != nil {
		return nil, fmt.Errorf("agent %q: %w", pb.GetName(), err)
	}

	// Recursively build sub-agents.
	subAgents := make([]agent.Agent, 0, len(pb.GetSubAgents()))
	for _, sub := range pb.GetSubAgents() {
		sa, err := NewFromProtoWithToolsetFactory(ctx, sub, providers, mcpRegistry, remoteAgentRegistry, daemonRegistry, httpFactory, toolsetFactory)
		if err != nil {
			return nil, fmt.Errorf("building sub-agent %q: %w", sub.GetName(), err)
		}
		subAgents = append(subAgents, sa)
	}

	// Resolve remote agents and add as sub-agents.
	remoteSubAgents, err := resolveRemoteAgents(pb, remoteAgentRegistry, daemonRegistry)
	if err != nil {
		return nil, fmt.Errorf("agent %q: %w", pb.GetName(), err)
	}
	subAgents = append(subAgents, remoteSubAgents...)

	switch pb.GetType() {
	case agentsv1.AgentType_AGENT_TYPE_LLM, agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED:
		return newLLMAgent(ctx, pb, mcpServers, subAgents, providers, httpFactory, toolsetFactory)
	case agentsv1.AgentType_AGENT_TYPE_LOOP:
		return newLoopAgent(pb, subAgents)
	case agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL:
		return newSequentialAgent(pb, subAgents)
	case agentsv1.AgentType_AGENT_TYPE_PARALLEL:
		return newParallelAgent(pb, subAgents)
	default:
		return nil, fmt.Errorf("unsupported agent type: %v", pb.GetType())
	}
}

func newLLMAgent(ctx context.Context, pb *agentsv1.Agent, mcpServers []*agentsv1.MCPServer, subAgents []agent.Agent, providers []agentsv1.ModelProvider, httpFactory MCPHTTPClientFactory, toolsetFactory ToolsetFactory) (agent.Agent, error) {
	logger := log.FromContext(ctx)
	acfg := pb.GetConfig()

	mcpNames := make([]string, 0, len(mcpServers))
	for _, s := range mcpServers {
		mcpNames = append(mcpNames, s.GetName())
	}

	subAgentNames := make([]string, 0, len(subAgents))
	for _, sa := range subAgents {
		subAgentNames = append(subAgentNames, sa.Name())
	}

	logger.Info("initializing LLM agent",
		"agent", pb.GetName(),
		"model", acfg.GetModel(),
		"mcp_servers", mcpNames,
		"sub_agents", subAgentNames,
		"output_key", acfg.GetOutputKey(),
		"disallow_transfer_to_parent", acfg.GetDisallowTransferToParent(),
		"disallow_transfer_to_peers", acfg.GetDisallowTransferToPeers(),
	)

	m, err := resolveModel(ctx, acfg.GetModel(), providers)
	if err != nil {
		return nil, fmt.Errorf("agent %q: creating model %q: %w", pb.GetName(), acfg.GetModel(), err)
	}

	toolsets, err := buildMCPToolsets(ctx, mcpServers, httpFactory)
	if err != nil {
		return nil, fmt.Errorf("agent %q: building MCP toolsets: %w", pb.GetName(), err)
	}
	if toolsetFactory != nil {
		extraToolsets, err := toolsetFactory(ctx, pb)
		if err != nil {
			return nil, fmt.Errorf("agent %q: building built-in toolsets: %w", pb.GetName(), err)
		}
		toolsets = append(toolsets, extraToolsets...)
	}

	cfg := llmagent.Config{
		Name:                     pb.GetName(),
		Description:              pb.GetDescription(),
		SubAgents:                subAgents,
		Model:                    m,
		Instruction:              acfg.GetInstruction(),
		GlobalInstruction:        acfg.GetGlobalInstruction(),
		DisallowTransferToParent: acfg.GetDisallowTransferToParent(),
		DisallowTransferToPeers:  acfg.GetDisallowTransferToPeers(),
		OutputKey:                acfg.GetOutputKey(),
		Toolsets:                 toolsets,
	}

	switch acfg.GetIncludeContents() {
	case agentsv1.LLMIncludeContents_LLM_INCLUDE_CONTENTS_NONE:
		cfg.IncludeContents = llmagent.IncludeContentsNone
	case agentsv1.LLMIncludeContents_LLM_INCLUDE_CONTENTS_DEFAULT:
		cfg.IncludeContents = llmagent.IncludeContentsDefault
	}

	return llmagent.New(cfg)
}

func newLoopAgent(pb *agentsv1.Agent, subAgents []agent.Agent) (agent.Agent, error) {
	maxIter := uint(pb.GetConfig().GetMaxIterations())

	return loopagent.New(loopagent.Config{
		AgentConfig: agent.Config{
			Name:        pb.GetName(),
			Description: pb.GetDescription(),
			SubAgents:   subAgents,
		},
		MaxIterations: maxIter,
	})
}

func newSequentialAgent(pb *agentsv1.Agent, subAgents []agent.Agent) (agent.Agent, error) {
	return sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        pb.GetName(),
			Description: pb.GetDescription(),
			SubAgents:   subAgents,
		},
	})
}

func newParallelAgent(pb *agentsv1.Agent, subAgents []agent.Agent) (agent.Agent, error) {
	return parallelagent.New(parallelagent.Config{
		AgentConfig: agent.Config{
			Name:        pb.GetName(),
			Description: pb.GetDescription(),
			SubAgents:   subAgents,
		},
	})
}

// resolveMCPServers looks up mcp_server_ids in the registry and returns the
// agent's inline mcp_servers merged with the resolved entries. Inline servers
// with the same name win. The agent proto is never mutated.
func resolveMCPServers(pb *agentsv1.Agent, registry []agentsv1.MCPServer) ([]*agentsv1.MCPServer, error) {
	cfg := pb.GetConfig()
	merged := cfg.GetMcpServers()
	if cfg == nil || len(cfg.GetMcpServerIds()) == 0 {
		return merged, nil
	}
	// Copy so appends never write into the proto's backing array.
	merged = append([]*agentsv1.MCPServer(nil), merged...)

	// Build lookup from registry.
	byID := make(map[string]*agentsv1.MCPServer, len(registry))
	for i := range registry {
		if id := registry[i].GetId(); id != "" {
			byID[id] = &registry[i]
		}
	}

	// Collect inline server names for collision detection.
	inlineNames := make(map[string]struct{}, len(merged))
	for _, s := range merged {
		inlineNames[s.GetName()] = struct{}{}
	}

	// Resolve each referenced ID.
	for _, id := range cfg.GetMcpServerIds() {
		srv, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("unknown mcp_server_id %q", id)
		}
		// Skip if an inline server already has the same name.
		if _, exists := inlineNames[srv.GetName()]; exists {
			continue
		}
		merged = append(merged, srv)
		inlineNames[srv.GetName()] = struct{}{}
	}

	return merged, nil
}

// buildMCPToolsets creates ADK toolsets from the agent's MCP server configs.
// Only HTTP-based transports (streamable HTTP and SSE) are supported.
func buildMCPToolsets(ctx context.Context, servers []*agentsv1.MCPServer, httpFactory MCPHTTPClientFactory) ([]tool.Toolset, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	toolsets := make([]tool.Toolset, 0, len(servers))
	for _, srv := range servers {
		transport, err := mcpTransport(ctx, srv, httpFactory)
		if err != nil {
			return nil, fmt.Errorf("mcp server %q: %w", srv.GetName(), err)
		}

		cfg := mcptoolset.Config{
			Transport: transport,
		}
		if len(srv.GetToolFilter()) > 0 {
			cfg.ToolFilter = tool.StringPredicate(srv.GetToolFilter())
		}

		ts, err := mcptoolset.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("mcp server %q: %w", srv.GetName(), err)
		}
		toolsets = append(toolsets, ts)
	}
	return toolsets, nil
}

// MCPTimeout returns the effective timeout for MCP probe operations.
// A zero value preserves the historical default.
func MCPTimeout(srv *agentsv1.MCPServer) (time.Duration, error) {
	if srv == nil || srv.GetTimeoutSeconds() == 0 {
		return defaultMCPTimeout, nil
	}
	if srv.GetTimeoutSeconds() < 0 {
		return 0, fmt.Errorf("timeout_seconds must be greater than or equal to 0")
	}
	return time.Duration(srv.GetTimeoutSeconds()) * time.Second, nil
}

// MCPProbeTool is a single tool exposed by an MCP server, surfaced alongside
// the server's configured allow-list verdict.
type MCPProbeTool struct {
	Name        string
	Description string
	Allowed     bool
}

// MCPProbeResult summarizes a live connectivity probe of an MCP server.
type MCPProbeResult struct {
	// Tools is the full list of tools the server advertised. The Allowed
	// field reflects the configured tool_filter (true when there is no filter).
	Tools []MCPProbeTool
	// ToolCount is the number of tools that pass the configured tool_filter.
	ToolCount int
}

// ProbeMCPServer connects to the configured MCP server transport, runs the
// MCP handshake, and lists the exposed tools.
func ProbeMCPServer(ctx context.Context, srv *agentsv1.MCPServer) (*MCPProbeResult, error) {
	return ProbeMCPServerWithFactory(ctx, srv, nil)
}

// ProbeMCPServerWithFactory connects using the same HTTP client factory used
// by runtime MCP toolsets.
func ProbeMCPServerWithFactory(ctx context.Context, srv *agentsv1.MCPServer, httpFactory MCPHTTPClientFactory) (*MCPProbeResult, error) {
	if srv == nil {
		return nil, fmt.Errorf("nil mcp server")
	}
	transport, err := mcpTransport(ctx, srv, httpFactory)
	if err != nil {
		return nil, err
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "butter-probe", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer session.Close()

	listResult, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}

	filterSet := map[string]struct{}{}
	hasFilter := false
	if filter := srv.GetToolFilter(); len(filter) > 0 {
		hasFilter = true
		for _, name := range filter {
			filterSet[name] = struct{}{}
		}
	}

	tools := make([]MCPProbeTool, 0, len(listResult.Tools))
	allowed := 0
	for _, t := range listResult.Tools {
		allow := true
		if hasFilter {
			_, allow = filterSet[t.Name]
		}
		if allow {
			allowed++
		}
		tools = append(tools, MCPProbeTool{
			Name:        t.Name,
			Description: t.Description,
			Allowed:     allow,
		})
	}
	return &MCPProbeResult{Tools: tools, ToolCount: allowed}, nil
}

// mcpTransport builds an MCP transport from the proto config.
func mcpTransport(ctx context.Context, srv *agentsv1.MCPServer, httpFactory MCPHTTPClientFactory) (mcp.Transport, error) {
	var httpClient *http.Client
	if httpFactory != nil {
		client, err := httpFactory.HTTPClientForMCP(ctx, srv)
		if err != nil {
			return nil, err
		}
		httpClient = client
	} else if len(srv.GetHeaders()) > 0 {
		httpClient = &http.Client{
			Transport: &headerTransport{
				base:    http.DefaultTransport,
				headers: srv.GetHeaders(),
			},
		}
	}

	// An explicitly configured timeout_seconds bounds runtime calls: a server
	// that accepts the connection but never finishes the response would
	// otherwise hang an agent run indefinitely (cron runs have no request
	// deadline). The timeout applies to the whole request/response exchange of
	// message POSTs (which carry tool-call responses, including the streamable
	// HTTP / SSE body the MCP SDK waits on), so a server that streams headers
	// and then stalls the body still gets cancelled. The long-lived standalone
	// SSE listener (a GET) is exempted so server-initiated messages keep
	// flowing for the agent's lifetime.
	if secs := srv.GetTimeoutSeconds(); secs > 0 {
		if httpClient == nil {
			httpClient = &http.Client{}
		}
		base := httpClient.Transport
		if base == nil {
			base = http.DefaultTransport
		}
		httpClient.Transport = &responseHeaderTimeoutTransport{
			base:      base,
			timeout:   time.Duration(secs) * time.Second,
			transport: srv.GetTransport(),
		}
	}

	switch srv.GetTransport() {
	case agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP:
		return &mcp.StreamableClientTransport{
			Endpoint:   srv.GetUrl(),
			HTTPClient: httpClient,
		}, nil
	case agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_SSE:
		return &mcp.SSEClientTransport{
			Endpoint:   srv.GetUrl(),
			HTTPClient: httpClient,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported transport %v, only streamable_http and sse are supported", srv.GetTransport())
	}
}

// responseHeaderTimeoutTransport bounds MCP requests by timeout. Exactly how
// it bounds a request depends on the request and the MCP transport, because
// the MCP SDK relies on long-lived streams that must not be killed:
//
//   - Message POSTs (streamable HTTP tool calls, SSE message sends): the whole
//     exchange — headers plus the response body the SDK reads — must finish
//     within timeout, so a server that streams headers then stalls the body
//     cannot hang the agent.
//   - The streamable HTTP standalone listener (a GET) is exempted once headers
//     arrive: it is a long-lived stream of server-initiated messages that may
//     stay idle for the agent's lifetime.
//   - The legacy SSE stream (a GET) is also long-lived, but the SDK blocks on
//     its body for the initial "endpoint" event before Connect returns. We
//     bound it only until the first bytes arrive (the handshake), then leave
//     the stream unbounded.
type responseHeaderTimeoutTransport struct {
	base      http.RoundTripper
	timeout   time.Duration
	transport agentsv1.MCPServerTransport
}

func (t *responseHeaderTimeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancelCause(req.Context())
	timer := time.AfterFunc(t.timeout, func() {
		cancel(fmt.Errorf("mcp request exceeded %s", t.timeout))
	})
	resp, err := t.base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		timer.Stop()
		cancel(nil)
		if cause := context.Cause(ctx); cause != nil && !errors.Is(cause, context.Canceled) {
			return nil, fmt.Errorf("%w: %w", cause, err)
		}
		return nil, err
	}

	switch {
	case req.Method == http.MethodGet && t.transport == agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP:
		// Standalone listener: may stay idle, so stop bounding at headers.
		timer.Stop()
		resp.Body = &cancelOnCloseBody{ReadCloser: resp.Body, cancel: func() { cancel(nil) }}
	case req.Method == http.MethodGet:
		// Legacy SSE stream: bound the handshake until the first complete SSE
		// event (the endpoint event the SDK waits on) has been delivered, then
		// leave the long-lived stream unbounded.
		resp.Body = &stopTimerOnSSEEventBody{
			ReadCloser: resp.Body,
			stop:       func() { timer.Stop() },
			cancel:     func() { cancel(nil) },
		}
	default:
		// Message exchange: keep the timer running so an unfinished body read
		// is cancelled. A normally consumed body stops the timer on close.
		resp.Body = &cancelOnCloseBody{ReadCloser: resp.Body, cancel: func() {
			timer.Stop()
			cancel(nil)
		}}
	}
	return resp, nil
}

type cancelOnCloseBody struct {
	io.ReadCloser
	cancel func()
}

func (b *cancelOnCloseBody) Close() error {
	err := b.ReadCloser.Close()
	b.cancel()
	return err
}

// stopTimerOnSSEEventBody stops the handshake timer once a complete SSE event
// (delimited by a blank line) has been read from a long-lived SSE stream — the
// MCP SDK only returns from Connect after the full endpoint event, so stopping
// on the first partial bytes could let a stalled mid-event server hang. After
// the first complete event the stream is left unbounded. The context is
// released on Close.
type stopTimerOnSSEEventBody struct {
	io.ReadCloser
	once   sync.Once
	stop   func()
	cancel func()
	tail   []byte // trailing bytes carried between reads to catch a split delimiter
	done   bool   // a complete event boundary has been seen
}

func (b *stopTimerOnSSEEventBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 && !b.done {
		b.scan(p[:n])
	}
	return n, err
}

// scan looks for an SSE event boundary (a blank line) across read chunks,
// keeping a few trailing bytes so a delimiter split between reads is detected.
func (b *stopTimerOnSSEEventBody) scan(data []byte) {
	buf := make([]byte, 0, len(b.tail)+len(data))
	buf = append(buf, b.tail...)
	buf = append(buf, data...)
	if bytes.Contains(buf, []byte("\n\n")) || bytes.Contains(buf, []byte("\r\n\r\n")) {
		b.done = true
		b.tail = nil
		b.once.Do(b.stop)
		return
	}
	const keep = 3 // longest delimiter prefix worth carrying ("\r\n\r")
	if len(buf) > keep {
		buf = buf[len(buf)-keep:]
	}
	b.tail = buf
}

func (b *stopTimerOnSSEEventBody) Close() error {
	err := b.ReadCloser.Close()
	b.once.Do(b.stop)
	b.cancel()
	return err
}

// headerTransport is an http.RoundTripper that injects custom headers.
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
}

// resolveRemoteAgents looks up remote_agent_ids in the registry and creates
// ADK agents for each. Supports A2A and DAEMON protocols.
func resolveRemoteAgents(pb *agentsv1.Agent, registry []agentsv1.RemoteAgent, daemonRegistry *daemon.Registry) ([]agent.Agent, error) {
	cfg := pb.GetConfig()
	if cfg == nil || len(cfg.GetRemoteAgentIds()) == 0 {
		return nil, nil
	}

	// Build lookup from registry.
	byID := make(map[string]*agentsv1.RemoteAgent, len(registry))
	for i := range registry {
		if id := registry[i].GetId(); id != "" {
			byID[id] = &registry[i]
		}
	}

	var agents []agent.Agent
	for _, id := range cfg.GetRemoteAgentIds() {
		ra, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("unknown remote_agent_id %q", id)
		}

		switch ra.GetProtocol() {
		case agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A:
			if ra.GetUrl() == "" {
				return nil, fmt.Errorf("remote agent %q: A2A protocol requires non-empty url", ra.GetName())
			}
			a, err := remoteagent.NewA2A(remoteagent.A2AConfig{
				Name:            ra.GetName(),
				Description:     fmt.Sprintf("Remote A2A agent: %s", ra.GetName()),
				AgentCardSource: ra.GetUrl(),
			})
			if err != nil {
				return nil, fmt.Errorf("creating remote agent %q: %w", ra.GetName(), err)
			}
			agents = append(agents, a)

		case agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON:
			runtimeID := strings.TrimSpace(ra.GetDaemonRuntimeId())
			if runtimeID == "" {
				return nil, fmt.Errorf("remote agent %q: DAEMON protocol requires non-empty daemon_runtime_id", ra.GetName())
			}
			acpRuntime := strings.TrimSpace(ra.GetAcpRuntime())
			if !isSupportedAcpRuntime(acpRuntime) {
				return nil, fmt.Errorf("remote agent %q: DAEMON protocol requires acp_runtime to be one of opencode, codex", ra.GetName())
			}
			if daemonRegistry == nil {
				return nil, fmt.Errorf("remote agent %q: daemon registry not available", ra.GetName())
			}
			workspaceID := ra.GetWorkspaceId()
			if workspaceID == "" {
				workspaceID = pb.GetWorkspaceId()
			}
			if workspaceID == "" {
				return nil, fmt.Errorf("remote agent %q: DAEMON protocol requires workspace_id", ra.GetName())
			}
			bridge := daemon.NewBridge(daemonRegistry, workspaceID, runtimeID, acpRuntime)
			a, err := bridge.BuildAgent(ra.GetName(), fmt.Sprintf("Daemon agent: %s", ra.GetName()))
			if err != nil {
				return nil, fmt.Errorf("creating daemon agent %q: %w", ra.GetName(), err)
			}
			agents = append(agents, a)

		default:
			return nil, fmt.Errorf("remote agent %q: unsupported protocol %v", ra.GetName(), ra.GetProtocol())
		}
	}

	return agents, nil
}

func isSupportedAcpRuntime(runtime string) bool {
	switch runtime {
	case "opencode", "codex":
		return true
	default:
		return false
	}
}
