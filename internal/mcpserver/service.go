package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/agentfile"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	implementationName    = "butter-workspace"
	implementationVersion = "v0.1.0"
)

// Service builds MCP servers exposing read-only workspace tools.
type Service struct {
	workspaceRepo workspacerepo.Repository
	configRepo    configRepository
	agentFileRepo agentfile.Repository
}

type configRepository interface {
	configrepo.AgentRepository
	configrepo.MCPServerRepository
}

func NewService(configRepo configRepository) *Service {
	return &Service{configRepo: configRepo}
}

func (s *Service) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

func (s *Service) SetConfigRepo(repo configRepository) {
	s.configRepo = repo
}

func (s *Service) SetAgentFileRepo(repo agentfile.Repository) {
	s.agentFileRepo = repo
}

func (s *Service) Handler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.ServerForRequest(r)
	}, &mcp.StreamableHTTPOptions{Stateless: true})
}

func (s *Service) ServerForRequest(_ *http.Request) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    implementationName,
		Version: implementationVersion,
	}, nil)
	s.registerTools(server)
	return server
}

func (s *Service) registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "workspace_info",
		Description: "Return the current Butter workspace summary.",
	}, s.workspaceInfo)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_agents",
		Description: "List agents configured in the current workspace.",
	}, s.listAgents)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_agent",
		Description: "Return one agent config from the current workspace.",
	}, s.getAgent)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_mcp_servers",
		Description: "List external MCP server configs in the current workspace with secrets redacted.",
	}, s.listMCPServers)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_file_spaces",
		Description: "List agent file spaces in the current workspace.",
	}, s.listFileSpaces)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_files",
		Description: "List files in an agent file space.",
	}, s.listFiles)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_file",
		Description: "Read a text file from an agent file space.",
	}, s.readFile)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_files",
		Description: "Search files in an agent file space.",
	}, s.searchFiles)
}

func workspaceID(ctx context.Context) (string, error) {
	id, ok := wsctx.FromContext(ctx)
	if !ok {
		return "", errors.New("workspace id missing from request context")
	}
	return id, nil
}

func requireConfigRepo(repo configRepository) (configRepository, error) {
	if repo == nil {
		return nil, errors.New("config repository not available")
	}
	return repo, nil
}

func requireAgentFileRepo(repo agentfile.Repository) (agentfile.Repository, error) {
	if repo == nil {
		return nil, errors.New("agent file repository not available")
	}
	return repo, nil
}

type workspaceInfoInput struct{}

type workspaceInfoOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

func (s *Service) workspaceInfo(ctx context.Context, _ *mcp.CallToolRequest, _ workspaceInfoInput) (*mcp.CallToolResult, workspaceInfoOutput, error) {
	wsID, err := workspaceID(ctx)
	if err != nil {
		return nil, workspaceInfoOutput{}, err
	}
	out := workspaceInfoOutput{ID: wsID}
	if s.workspaceRepo == nil {
		return nil, out, nil
	}
	ws, err := s.workspaceRepo.GetWorkspace(ctx, wsID)
	if err != nil {
		return nil, workspaceInfoOutput{}, fmt.Errorf("get workspace: %w", err)
	}
	out.Name = ws.GetName()
	out.Slug = ws.GetSlug()
	out.Description = ws.GetDescription()
	return nil, out, nil
}

type listAgentsInput struct {
	Limit int `json:"limit,omitempty"`
}

type agentSummary struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Type        string            `json:"type,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type listAgentsOutput struct {
	Agents []agentSummary `json:"agents"`
}

func (s *Service) listAgents(ctx context.Context, _ *mcp.CallToolRequest, in listAgentsInput) (*mcp.CallToolResult, listAgentsOutput, error) {
	repo, err := requireConfigRepo(s.configRepo)
	if err != nil {
		return nil, listAgentsOutput{}, err
	}
	wsID, err := workspaceID(ctx)
	if err != nil {
		return nil, listAgentsOutput{}, err
	}
	agents, err := repo.ListAgents(ctx, wsID)
	if err != nil {
		return nil, listAgentsOutput{}, fmt.Errorf("list agents: %w", err)
	}
	limit := normalizeLimit(in.Limit, 100)
	out := listAgentsOutput{Agents: make([]agentSummary, 0, min(len(agents), limit))}
	for i, agent := range agents {
		if i >= limit {
			break
		}
		out.Agents = append(out.Agents, agentSummary{
			Name:        agent.GetName(),
			Description: agent.GetDescription(),
			Type:        agent.GetType().String(),
			Labels:      agent.GetLabels(),
			Metadata:    redactMap(agent.GetMetadata()),
		})
	}
	return nil, out, nil
}

type getAgentInput struct {
	Name string `json:"name" jsonschema:"the agent name"`
}

type getAgentOutput struct {
	Agent map[string]any `json:"agent"`
}

func (s *Service) getAgent(ctx context.Context, _ *mcp.CallToolRequest, in getAgentInput) (*mcp.CallToolResult, getAgentOutput, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, getAgentOutput{}, errors.New("name is required")
	}
	repo, err := requireConfigRepo(s.configRepo)
	if err != nil {
		return nil, getAgentOutput{}, err
	}
	wsID, err := workspaceID(ctx)
	if err != nil {
		return nil, getAgentOutput{}, err
	}
	agent, err := repo.GetAgent(ctx, wsID, in.Name)
	if err != nil {
		return nil, getAgentOutput{}, fmt.Errorf("get agent: %w", err)
	}
	return nil, getAgentOutput{Agent: protoToMap(redactAgent(agent))}, nil
}

type listMCPServersInput struct{}

type mcpServerSummary struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Transport   string            `json:"transport,omitempty"`
	URL         string            `json:"url,omitempty"`
	ToolFilter  []string          `json:"tool_filter,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	AuthType    string            `json:"auth_type,omitempty"`
	HeaderNames []string          `json:"header_names,omitempty"`
}

type listMCPServersOutput struct {
	MCPServers []mcpServerSummary `json:"mcp_servers"`
}

func (s *Service) listMCPServers(ctx context.Context, _ *mcp.CallToolRequest, _ listMCPServersInput) (*mcp.CallToolResult, listMCPServersOutput, error) {
	repo, err := requireConfigRepo(s.configRepo)
	if err != nil {
		return nil, listMCPServersOutput{}, err
	}
	wsID, err := workspaceID(ctx)
	if err != nil {
		return nil, listMCPServersOutput{}, err
	}
	servers, err := repo.ListMCPServers(ctx, wsID)
	if err != nil {
		return nil, listMCPServersOutput{}, fmt.Errorf("list mcp servers: %w", err)
	}
	out := listMCPServersOutput{MCPServers: make([]mcpServerSummary, 0, len(servers))}
	for _, server := range servers {
		out.MCPServers = append(out.MCPServers, mcpServerSummary{
			ID:          server.GetId(),
			Name:        server.GetName(),
			Transport:   server.GetTransport().String(),
			URL:         server.GetUrl(),
			ToolFilter:  server.GetToolFilter(),
			Metadata:    redactMap(server.GetMetadata()),
			AuthType:    server.GetAuth().GetType().String(),
			HeaderNames: mapKeys(server.GetHeaders()),
		})
	}
	return nil, out, nil
}

type listFileSpacesInput struct{}

type fileSpaceSummary struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type listFileSpacesOutput struct {
	Spaces []fileSpaceSummary `json:"spaces"`
}

func (s *Service) listFileSpaces(ctx context.Context, _ *mcp.CallToolRequest, _ listFileSpacesInput) (*mcp.CallToolResult, listFileSpacesOutput, error) {
	repo, err := requireAgentFileRepo(s.agentFileRepo)
	if err != nil {
		return nil, listFileSpacesOutput{}, err
	}
	wsID, err := workspaceID(ctx)
	if err != nil {
		return nil, listFileSpacesOutput{}, err
	}
	spaces, err := repo.ListSpaces(ctx, wsID)
	if err != nil {
		return nil, listFileSpacesOutput{}, fmt.Errorf("list file spaces: %w", err)
	}
	out := listFileSpacesOutput{Spaces: make([]fileSpaceSummary, 0, len(spaces))}
	for _, space := range spaces {
		out.Spaces = append(out.Spaces, fileSpaceSummary{
			ID:          space.GetId(),
			Name:        space.GetName(),
			Description: space.GetDescription(),
			Metadata:    redactMap(space.GetMetadata()),
		})
	}
	return nil, out, nil
}

type listFilesInput struct {
	SpaceID    string `json:"space_id"`
	PathPrefix string `json:"path_prefix,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type fileSummary struct {
	ID          string            `json:"id"`
	SpaceID     string            `json:"space_id"`
	Path        string            `json:"path"`
	ContentType string            `json:"content_type,omitempty"`
	SizeBytes   int64             `json:"size_bytes"`
	Version     int64             `json:"version"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type listFilesOutput struct {
	Files []fileSummary `json:"files"`
}

func (s *Service) listFiles(ctx context.Context, _ *mcp.CallToolRequest, in listFilesInput) (*mcp.CallToolResult, listFilesOutput, error) {
	if strings.TrimSpace(in.SpaceID) == "" {
		return nil, listFilesOutput{}, errors.New("space_id is required")
	}
	repo, err := requireAgentFileRepo(s.agentFileRepo)
	if err != nil {
		return nil, listFilesOutput{}, err
	}
	wsID, err := workspaceID(ctx)
	if err != nil {
		return nil, listFilesOutput{}, err
	}
	files, err := repo.ListFiles(ctx, wsID, in.SpaceID, in.PathPrefix)
	if err != nil {
		return nil, listFilesOutput{}, fmt.Errorf("list files: %w", err)
	}
	limit := normalizeLimit(in.Limit, 200)
	out := listFilesOutput{Files: make([]fileSummary, 0, min(len(files), limit))}
	for i, file := range files {
		if i >= limit {
			break
		}
		out.Files = append(out.Files, summarizeFile(file))
	}
	return nil, out, nil
}

type readFileInput struct {
	SpaceID string `json:"space_id"`
	Path    string `json:"path"`
	Version int64  `json:"version,omitempty"`
}

type readFileOutput struct {
	File    fileSummary `json:"file"`
	Content string      `json:"content"`
}

func (s *Service) readFile(ctx context.Context, _ *mcp.CallToolRequest, in readFileInput) (*mcp.CallToolResult, readFileOutput, error) {
	if strings.TrimSpace(in.SpaceID) == "" {
		return nil, readFileOutput{}, errors.New("space_id is required")
	}
	if strings.TrimSpace(in.Path) == "" {
		return nil, readFileOutput{}, errors.New("path is required")
	}
	repo, err := requireAgentFileRepo(s.agentFileRepo)
	if err != nil {
		return nil, readFileOutput{}, err
	}
	wsID, err := workspaceID(ctx)
	if err != nil {
		return nil, readFileOutput{}, err
	}
	file, content, err := repo.ReadFile(ctx, wsID, in.SpaceID, in.Path, in.Version)
	if err != nil {
		return nil, readFileOutput{}, fmt.Errorf("read file: %w", err)
	}
	return nil, readFileOutput{File: summarizeFile(file), Content: content}, nil
}

type searchFilesInput struct {
	SpaceID string `json:"space_id"`
	Query   string `json:"query"`
	Limit   int    `json:"limit,omitempty"`
}

type fileSearchResult struct {
	File     fileSummary `json:"file"`
	Snippets []string    `json:"snippets,omitempty"`
}

type searchFilesOutput struct {
	Results []fileSearchResult `json:"results"`
}

func (s *Service) searchFiles(ctx context.Context, _ *mcp.CallToolRequest, in searchFilesInput) (*mcp.CallToolResult, searchFilesOutput, error) {
	if strings.TrimSpace(in.SpaceID) == "" {
		return nil, searchFilesOutput{}, errors.New("space_id is required")
	}
	if strings.TrimSpace(in.Query) == "" {
		return nil, searchFilesOutput{}, errors.New("query is required")
	}
	repo, err := requireAgentFileRepo(s.agentFileRepo)
	if err != nil {
		return nil, searchFilesOutput{}, err
	}
	wsID, err := workspaceID(ctx)
	if err != nil {
		return nil, searchFilesOutput{}, err
	}
	limit := normalizeLimit(in.Limit, 100)
	results, err := repo.SearchFiles(ctx, wsID, in.SpaceID, in.Query, limit)
	if err != nil {
		return nil, searchFilesOutput{}, fmt.Errorf("search files: %w", err)
	}
	out := searchFilesOutput{Results: make([]fileSearchResult, 0, len(results))}
	for _, result := range results {
		out.Results = append(out.Results, fileSearchResult{
			File:     summarizeFile(result.GetFile()),
			Snippets: result.GetSnippets(),
		})
	}
	return nil, out, nil
}

func summarizeFile(file *agentsv1.AgentFile) fileSummary {
	return fileSummary{
		ID:          file.GetId(),
		SpaceID:     file.GetSpaceId(),
		Path:        file.GetPath(),
		ContentType: file.GetContentType(),
		SizeBytes:   file.GetSizeBytes(),
		Version:     file.GetVersion(),
		Metadata:    redactMap(file.GetMetadata()),
	}
}

func normalizeLimit(limit, fallback int) int {
	if limit <= 0 || limit > fallback {
		return fallback
	}
	return limit
}

func mapKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func protoToMap(msg proto.Message) map[string]any {
	raw, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(msg)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func redactAgent(agent *agentsv1.Agent) *agentsv1.Agent {
	if agent == nil {
		return nil
	}
	clone := proto.Clone(agent).(*agentsv1.Agent)
	clone.Metadata = redactMap(clone.GetMetadata())
	redactAgentConfig(clone.GetConfig())
	for _, sub := range clone.GetSubAgents() {
		redactAgentInPlace(sub)
	}
	return clone
}

func redactAgentInPlace(agent *agentsv1.Agent) {
	if agent == nil {
		return
	}
	agent.Metadata = redactMap(agent.GetMetadata())
	redactAgentConfig(agent.GetConfig())
	for _, sub := range agent.GetSubAgents() {
		redactAgentInPlace(sub)
	}
}

func redactAgentConfig(cfg *agentsv1.AgentConfig) {
	if cfg == nil {
		return
	}
	for _, server := range cfg.GetMcpServers() {
		redactMCPServer(server)
	}
}

func redactMCPServer(server *agentsv1.MCPServer) {
	if server == nil {
		return
	}
	for k := range server.Headers {
		server.Headers[k] = "***"
	}
	if oauth := server.GetAuth().GetOauth2(); oauth != nil {
		oauth.ClientSecret = ""
	}
	server.Metadata = redactMap(server.GetMetadata())
}

func redactMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if isSecretKey(k) {
			out[k] = "***"
			continue
		}
		out[k] = v
	}
	return out
}

func isSecretKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "secret") ||
		strings.Contains(k, "token") ||
		strings.Contains(k, "password") ||
		strings.Contains(k, "credential") ||
		strings.Contains(k, "authorization") ||
		strings.Contains(k, "api_key") ||
		strings.Contains(k, "apikey")
}
