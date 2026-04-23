package agent

import (
	"context"
	"fmt"
	"net/http"

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

// NewFromProto creates an ADK agent from an agentsv1.Agent proto config.
// providers is the list of model provider mappings used to resolve LLM backends.
// mcpRegistry is the shared MCP server config pool; agents reference entries by ID.
// remoteAgentRegistry is the shared remote agent config pool; agents reference entries by ID.
func NewFromProto(ctx context.Context, pb *agentsv1.Agent, providers []agentsv1.ModelProvider, mcpRegistry []agentsv1.MCPServer, remoteAgentRegistry []agentsv1.RemoteAgent, daemonRegistry *daemon.Registry) (agent.Agent, error) {
	if pb == nil {
		return nil, fmt.Errorf("agent config is nil")
	}

	// Resolve shared MCP servers and merge with inline ones.
	if err := resolveMCPServers(pb, mcpRegistry); err != nil {
		return nil, fmt.Errorf("agent %q: %w", pb.GetName(), err)
	}

	// Recursively build sub-agents.
	subAgents := make([]agent.Agent, 0, len(pb.GetSubAgents()))
	for _, sub := range pb.GetSubAgents() {
		sa, err := NewFromProto(ctx, sub, providers, mcpRegistry, remoteAgentRegistry, daemonRegistry)
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
		return newLLMAgent(ctx, pb, subAgents, providers)
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

func newLLMAgent(ctx context.Context, pb *agentsv1.Agent, subAgents []agent.Agent, providers []agentsv1.ModelProvider) (agent.Agent, error) {
	logger := log.FromContext(ctx)
	acfg := pb.GetConfig()

	mcpServers := acfg.GetMcpServers()
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

	toolsets, err := buildMCPToolsets(acfg.GetMcpServers())
	if err != nil {
		return nil, fmt.Errorf("agent %q: building MCP toolsets: %w", pb.GetName(), err)
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

// resolveMCPServers looks up mcp_server_ids in the registry and merges them
// into the agent's inline mcp_servers. Inline servers with the same name win.
func resolveMCPServers(pb *agentsv1.Agent, registry []agentsv1.MCPServer) error {
	cfg := pb.GetConfig()
	if cfg == nil || len(cfg.GetMcpServerIds()) == 0 {
		return nil
	}

	// Build lookup from registry.
	byID := make(map[string]*agentsv1.MCPServer, len(registry))
	for i := range registry {
		if id := registry[i].GetId(); id != "" {
			byID[id] = &registry[i]
		}
	}

	// Collect inline server names for collision detection.
	inlineNames := make(map[string]struct{}, len(cfg.GetMcpServers()))
	for _, s := range cfg.GetMcpServers() {
		inlineNames[s.GetName()] = struct{}{}
	}

	// Resolve each referenced ID.
	for _, id := range cfg.GetMcpServerIds() {
		srv, ok := byID[id]
		if !ok {
			return fmt.Errorf("unknown mcp_server_id %q", id)
		}
		// Skip if an inline server already has the same name.
		if _, exists := inlineNames[srv.GetName()]; exists {
			continue
		}
		cfg.McpServers = append(cfg.McpServers, srv)
		inlineNames[srv.GetName()] = struct{}{}
	}

	return nil
}

// buildMCPToolsets creates ADK toolsets from the agent's MCP server configs.
// Only HTTP-based transports (streamable HTTP and SSE) are supported.
func buildMCPToolsets(servers []*agentsv1.MCPServer) ([]tool.Toolset, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	toolsets := make([]tool.Toolset, 0, len(servers))
	for _, srv := range servers {
		transport, err := mcpTransport(srv)
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

// mcpTransport builds an MCP transport from the proto config.
func mcpTransport(srv *agentsv1.MCPServer) (mcp.Transport, error) {
	var httpClient *http.Client
	if len(srv.GetHeaders()) > 0 {
		httpClient = &http.Client{
			Transport: &headerTransport{
				base:    http.DefaultTransport,
				headers: srv.GetHeaders(),
			},
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
			if ra.GetDaemonCapability() == "" {
				return nil, fmt.Errorf("remote agent %q: DAEMON protocol requires non-empty daemon_capability", ra.GetName())
			}
			if daemonRegistry == nil {
				return nil, fmt.Errorf("remote agent %q: daemon registry not available", ra.GetName())
			}
			bridge := daemon.NewBridge(daemonRegistry, ra.GetDaemonCapability())
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
