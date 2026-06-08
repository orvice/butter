package mongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	agentsCollection         = "config_agents"
	globalMCPCollection      = "config_global_mcpservers"
	mcpServersCollection     = "config_mcpservers"
	remoteAgentsCollection   = "config_remoteagents"
	daemonRuntimesCollection = "config_daemon_runtimes"
	channelsCollection       = "config_channels"
	modelProvidersCollection = "config_modelproviders"
	notifyGroupsCollection   = "config_notifygroups"
)

// configDoc is the generic MongoDB document for a config entity. The _id is
// a composite of "workspace_id:name" (or "workspace_id:entity_id" depending
// on the entity); workspace_id and name are duplicated as queryable fields.
type configDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	Name        string `bson:"name"`
	Spec        string `bson:"spec"`
}

// Store implements all config repository interfaces backed by MongoDB.
type Store struct {
	agents         *mongo.Collection
	globalMCP      *mongo.Collection
	mcpServers     *mongo.Collection
	remoteAgents   *mongo.Collection
	daemonRuntimes *mongo.Collection
	channels       *mongo.Collection
	modelProviders *mongo.Collection
	notifyGroups   *mongo.Collection
}

func New(db *mongo.Database) *Store {
	return &Store{
		agents:         db.Collection(agentsCollection),
		globalMCP:      db.Collection(globalMCPCollection),
		mcpServers:     db.Collection(mcpServersCollection),
		remoteAgents:   db.Collection(remoteAgentsCollection),
		daemonRuntimes: db.Collection(daemonRuntimesCollection),
		channels:       db.Collection(channelsCollection),
		modelProviders: db.Collection(modelProvidersCollection),
		notifyGroups:   db.Collection(notifyGroupsCollection),
	}
}

// EnsureIndexes creates the (workspace_id, name) compound indexes for fast
// per-workspace listings.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	collections := []*mongo.Collection{s.agents, s.globalMCP, s.mcpServers, s.remoteAgents, s.daemonRuntimes, s.channels, s.modelProviders, s.notifyGroups}
	for _, c := range collections {
		_, err := c.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "name", Value: 1}},
		})
		if err != nil {
			return fmt.Errorf("create %s index: %w", c.Name(), err)
		}
	}
	return nil
}

func mapError(entity, ws, key string, err error) error {
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, configrepo.ErrAlreadyExists)
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, configrepo.ErrNotFound)
	}
	return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, err)
}

func marshal(m proto.Message) (string, error) {
	b, err := protojson.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshal protojson: %w", err)
	}
	return string(b), nil
}

func unmarshal(data string, m proto.Message) error {
	return protojson.Unmarshal([]byte(data), m)
}

func unmarshalMCPServer(data string, m *agentsv1.MCPServer) error {
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal([]byte(data), m)
}

func compositeID(workspaceID, name string) string {
	return workspaceID + ":" + name
}

func listInWorkspace(ctx context.Context, c *mongo.Collection, workspaceID string) ([]configDoc, error) {
	cursor, err := c.Find(ctx, bson.M{"workspace_id": workspaceID})
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", c.Name(), err)
	}
	defer cursor.Close(ctx)
	var docs []configDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode %s: %w", c.Name(), err)
	}
	return docs, nil
}

func listAll(ctx context.Context, c *mongo.Collection) ([]configDoc, error) {
	cursor, err := c.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", c.Name(), err)
	}
	defer cursor.Close(ctx)
	var docs []configDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode %s: %w", c.Name(), err)
	}
	return docs, nil
}

// --- Agents ---

func (s *Store) ListAgents(ctx context.Context, workspaceID string) ([]*agentsv1.Agent, error) {
	docs, err := listInWorkspace(ctx, s.agents, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.Agent, 0, len(docs))
	for _, d := range docs {
		a := &agentsv1.Agent{}
		if err := unmarshal(d.Spec, a); err != nil {
			return nil, fmt.Errorf("unmarshal agent %q: %w", d.ID, err)
		}
		out = append(out, a)
	}
	return out, nil
}

func (s *Store) ListAgentsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.Agent, error) {
	docs, err := listAll(ctx, s.agents)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.Agent, 0, len(docs))
	for _, d := range docs {
		a := &agentsv1.Agent{}
		if err := unmarshal(d.Spec, a); err != nil {
			return nil, fmt.Errorf("unmarshal agent %q: %w", d.ID, err)
		}
		out = append(out, a)
	}
	return out, nil
}

func (s *Store) GetAgent(ctx context.Context, workspaceID, name string) (*agentsv1.Agent, error) {
	var doc configDoc
	err := s.agents.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, name)}).Decode(&doc)
	if err != nil {
		return nil, mapError("agent", workspaceID, name, err)
	}
	a := &agentsv1.Agent{}
	if err := unmarshal(doc.Spec, a); err != nil {
		return nil, fmt.Errorf("unmarshal agent %q: %w", name, err)
	}
	return a, nil
}

func (s *Store) CreateAgent(ctx context.Context, workspaceID string, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	clone := proto.Clone(agent).(*agentsv1.Agent)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetName()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	if _, err := s.agents.InsertOne(ctx, doc); err != nil {
		return nil, mapError("agent", workspaceID, clone.GetName(), err)
	}
	return clone, nil
}

func (s *Store) UpdateAgent(ctx context.Context, workspaceID string, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	clone := proto.Clone(agent).(*agentsv1.Agent)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetName()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	res, err := s.agents.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("agent", workspaceID, clone.GetName(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("agent", workspaceID, clone.GetName(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteAgent(ctx context.Context, workspaceID, name string) error {
	res, err := s.agents.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, name)})
	if err != nil {
		return mapError("agent", workspaceID, name, err)
	}
	if res.DeletedCount == 0 {
		return mapError("agent", workspaceID, name, mongo.ErrNoDocuments)
	}
	return nil
}

// --- MCP Servers ---

func (s *Store) ListMCPServers(ctx context.Context, workspaceID string) ([]*agentsv1.MCPServer, error) {
	docs, err := listInWorkspace(ctx, s.mcpServers, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.MCPServer, 0, len(docs))
	for _, d := range docs {
		m := &agentsv1.MCPServer{}
		if err := unmarshalMCPServer(d.Spec, m); err != nil {
			return nil, fmt.Errorf("unmarshal mcp server %q: %w", d.ID, err)
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *Store) ListMCPServersAcrossWorkspaces(ctx context.Context) ([]*agentsv1.MCPServer, error) {
	docs, err := listAll(ctx, s.mcpServers)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.MCPServer, 0, len(docs))
	for _, d := range docs {
		m := &agentsv1.MCPServer{}
		if err := unmarshalMCPServer(d.Spec, m); err != nil {
			return nil, fmt.Errorf("unmarshal mcp server %q: %w", d.ID, err)
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *Store) GetMCPServer(ctx context.Context, workspaceID, id string) (*agentsv1.MCPServer, error) {
	var doc configDoc
	err := s.mcpServers.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, id)}).Decode(&doc)
	if err != nil {
		return nil, mapError("mcp server", workspaceID, id, err)
	}
	m := &agentsv1.MCPServer{}
	if err := unmarshalMCPServer(doc.Spec, m); err != nil {
		return nil, fmt.Errorf("unmarshal mcp server %q: %w", id, err)
	}
	return m, nil
}

func (s *Store) CreateMCPServer(ctx context.Context, workspaceID string, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	clone := proto.Clone(server).(*agentsv1.MCPServer)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetId()), WorkspaceID: workspaceID, Name: clone.GetId(), Spec: spec}
	if _, err := s.mcpServers.InsertOne(ctx, doc); err != nil {
		return nil, mapError("mcp server", workspaceID, clone.GetId(), err)
	}
	return clone, nil
}

func (s *Store) UpdateMCPServer(ctx context.Context, workspaceID string, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	clone := proto.Clone(server).(*agentsv1.MCPServer)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetId()), WorkspaceID: workspaceID, Name: clone.GetId(), Spec: spec}
	res, err := s.mcpServers.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("mcp server", workspaceID, clone.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("mcp server", workspaceID, clone.GetId(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteMCPServer(ctx context.Context, workspaceID, id string) error {
	res, err := s.mcpServers.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, id)})
	if err != nil {
		return mapError("mcp server", workspaceID, id, err)
	}
	if res.DeletedCount == 0 {
		return mapError("mcp server", workspaceID, id, mongo.ErrNoDocuments)
	}
	return nil
}

// --- Global MCP Servers ---

func (s *Store) ListGlobalMCPServers(ctx context.Context) ([]*agentsv1.MCPServer, error) {
	docs, err := listAll(ctx, s.globalMCP)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.MCPServer, 0, len(docs))
	for _, d := range docs {
		m := &agentsv1.MCPServer{}
		if err := unmarshalMCPServer(d.Spec, m); err != nil {
			return nil, fmt.Errorf("unmarshal global mcp server %q: %w", d.ID, err)
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *Store) GetGlobalMCPServer(ctx context.Context, id string) (*agentsv1.MCPServer, error) {
	var doc configDoc
	err := s.globalMCP.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		return nil, mapError("global mcp server", "", id, err)
	}
	m := &agentsv1.MCPServer{}
	if err := unmarshalMCPServer(doc.Spec, m); err != nil {
		return nil, fmt.Errorf("unmarshal global mcp server %q: %w", id, err)
	}
	return m, nil
}

func (s *Store) CreateGlobalMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	clone := proto.Clone(server).(*agentsv1.MCPServer)
	clone.WorkspaceId = ""
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: clone.GetId(), Name: clone.GetId(), Spec: spec}
	if _, err := s.globalMCP.InsertOne(ctx, doc); err != nil {
		return nil, mapError("global mcp server", "", clone.GetId(), err)
	}
	return clone, nil
}

func (s *Store) UpdateGlobalMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	clone := proto.Clone(server).(*agentsv1.MCPServer)
	clone.WorkspaceId = ""
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: clone.GetId(), Name: clone.GetId(), Spec: spec}
	res, err := s.globalMCP.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("global mcp server", "", clone.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("global mcp server", "", clone.GetId(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteGlobalMCPServer(ctx context.Context, id string) error {
	res, err := s.globalMCP.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return mapError("global mcp server", "", id, err)
	}
	if res.DeletedCount == 0 {
		return mapError("global mcp server", "", id, mongo.ErrNoDocuments)
	}
	return nil
}

// --- Remote Agents ---

func (s *Store) ListRemoteAgents(ctx context.Context, workspaceID string) ([]*agentsv1.RemoteAgent, error) {
	docs, err := listInWorkspace(ctx, s.remoteAgents, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.RemoteAgent, 0, len(docs))
	for _, d := range docs {
		r := &agentsv1.RemoteAgent{}
		if err := unmarshal(d.Spec, r); err != nil {
			return nil, fmt.Errorf("unmarshal remote agent %q: %w", d.ID, err)
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *Store) ListRemoteAgentsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.RemoteAgent, error) {
	docs, err := listAll(ctx, s.remoteAgents)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.RemoteAgent, 0, len(docs))
	for _, d := range docs {
		r := &agentsv1.RemoteAgent{}
		if err := unmarshal(d.Spec, r); err != nil {
			return nil, fmt.Errorf("unmarshal remote agent %q: %w", d.ID, err)
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *Store) GetRemoteAgent(ctx context.Context, workspaceID, id string) (*agentsv1.RemoteAgent, error) {
	var doc configDoc
	err := s.remoteAgents.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, id)}).Decode(&doc)
	if err != nil {
		return nil, mapError("remote agent", workspaceID, id, err)
	}
	r := &agentsv1.RemoteAgent{}
	if err := unmarshal(doc.Spec, r); err != nil {
		return nil, fmt.Errorf("unmarshal remote agent %q: %w", id, err)
	}
	return r, nil
}

func (s *Store) CreateRemoteAgent(ctx context.Context, workspaceID string, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	clone := proto.Clone(agent).(*agentsv1.RemoteAgent)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetId()), WorkspaceID: workspaceID, Name: clone.GetId(), Spec: spec}
	if _, err := s.remoteAgents.InsertOne(ctx, doc); err != nil {
		return nil, mapError("remote agent", workspaceID, clone.GetId(), err)
	}
	return clone, nil
}

func (s *Store) UpdateRemoteAgent(ctx context.Context, workspaceID string, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	clone := proto.Clone(agent).(*agentsv1.RemoteAgent)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetId()), WorkspaceID: workspaceID, Name: clone.GetId(), Spec: spec}
	res, err := s.remoteAgents.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("remote agent", workspaceID, clone.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("remote agent", workspaceID, clone.GetId(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteRemoteAgent(ctx context.Context, workspaceID, id string) error {
	res, err := s.remoteAgents.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, id)})
	if err != nil {
		return mapError("remote agent", workspaceID, id, err)
	}
	if res.DeletedCount == 0 {
		return mapError("remote agent", workspaceID, id, mongo.ErrNoDocuments)
	}
	return nil
}

// --- Daemon Configs ---

func (s *Store) ListDaemonRuntimes(ctx context.Context, workspaceID string) ([]*agentsv1.DaemonRuntime, error) {
	docs, err := listInWorkspace(ctx, s.daemonRuntimes, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.DaemonRuntime, 0, len(docs))
	for _, d := range docs {
		daemon := &agentsv1.DaemonRuntime{}
		if err := unmarshal(d.Spec, daemon); err != nil {
			return nil, fmt.Errorf("unmarshal daemon runtime %q: %w", d.ID, err)
		}
		out = append(out, daemon)
	}
	return out, nil
}

func (s *Store) ListDaemonRuntimesAcrossWorkspaces(ctx context.Context) ([]*agentsv1.DaemonRuntime, error) {
	docs, err := listAll(ctx, s.daemonRuntimes)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.DaemonRuntime, 0, len(docs))
	for _, d := range docs {
		daemon := &agentsv1.DaemonRuntime{}
		if err := unmarshal(d.Spec, daemon); err != nil {
			return nil, fmt.Errorf("unmarshal daemon runtime %q: %w", d.ID, err)
		}
		out = append(out, daemon)
	}
	return out, nil
}

func (s *Store) GetDaemonRuntime(ctx context.Context, workspaceID, id string) (*agentsv1.DaemonRuntime, error) {
	var doc configDoc
	err := s.daemonRuntimes.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, id)}).Decode(&doc)
	if err != nil {
		return nil, mapError("daemon runtime", workspaceID, id, err)
	}
	daemon := &agentsv1.DaemonRuntime{}
	if err := unmarshal(doc.Spec, daemon); err != nil {
		return nil, fmt.Errorf("unmarshal daemon runtime %q: %w", id, err)
	}
	return daemon, nil
}

func (s *Store) CreateDaemonRuntime(ctx context.Context, workspaceID string, daemon *agentsv1.DaemonRuntime) (*agentsv1.DaemonRuntime, error) {
	clone := proto.Clone(daemon).(*agentsv1.DaemonRuntime)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetId()), WorkspaceID: workspaceID, Name: clone.GetId(), Spec: spec}
	if _, err := s.daemonRuntimes.InsertOne(ctx, doc); err != nil {
		return nil, mapError("daemon runtime", workspaceID, clone.GetId(), err)
	}
	return clone, nil
}

func (s *Store) UpdateDaemonRuntime(ctx context.Context, workspaceID string, daemon *agentsv1.DaemonRuntime) (*agentsv1.DaemonRuntime, error) {
	clone := proto.Clone(daemon).(*agentsv1.DaemonRuntime)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetId()), WorkspaceID: workspaceID, Name: clone.GetId(), Spec: spec}
	res, err := s.daemonRuntimes.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("daemon runtime", workspaceID, clone.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("daemon runtime", workspaceID, clone.GetId(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteDaemonRuntime(ctx context.Context, workspaceID, id string) error {
	res, err := s.daemonRuntimes.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, id)})
	if err != nil {
		return mapError("daemon runtime", workspaceID, id, err)
	}
	if res.DeletedCount == 0 {
		return mapError("daemon runtime", workspaceID, id, mongo.ErrNoDocuments)
	}
	return nil
}

// --- Channels ---

func (s *Store) ListChannels(ctx context.Context, workspaceID string) ([]*agentsv1.AgentChannel, error) {
	docs, err := listInWorkspace(ctx, s.channels, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.AgentChannel, 0, len(docs))
	for _, d := range docs {
		c := &agentsv1.AgentChannel{}
		if err := unmarshal(d.Spec, c); err != nil {
			return nil, fmt.Errorf("unmarshal channel %q: %w", d.ID, err)
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *Store) ListChannelsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.AgentChannel, error) {
	docs, err := listAll(ctx, s.channels)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.AgentChannel, 0, len(docs))
	for _, d := range docs {
		c := &agentsv1.AgentChannel{}
		if err := unmarshal(d.Spec, c); err != nil {
			return nil, fmt.Errorf("unmarshal channel %q: %w", d.ID, err)
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *Store) GetChannel(ctx context.Context, workspaceID, name string) (*agentsv1.AgentChannel, error) {
	var doc configDoc
	err := s.channels.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, name)}).Decode(&doc)
	if err != nil {
		return nil, mapError("channel", workspaceID, name, err)
	}
	c := &agentsv1.AgentChannel{}
	if err := unmarshal(doc.Spec, c); err != nil {
		return nil, fmt.Errorf("unmarshal channel %q: %w", name, err)
	}
	return c, nil
}

func (s *Store) CreateChannel(ctx context.Context, workspaceID string, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	clone := proto.Clone(channel).(*agentsv1.AgentChannel)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetName()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	if _, err := s.channels.InsertOne(ctx, doc); err != nil {
		return nil, mapError("channel", workspaceID, clone.GetName(), err)
	}
	return clone, nil
}

func (s *Store) UpdateChannel(ctx context.Context, workspaceID string, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	clone := proto.Clone(channel).(*agentsv1.AgentChannel)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetName()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	res, err := s.channels.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("channel", workspaceID, clone.GetName(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("channel", workspaceID, clone.GetName(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteChannel(ctx context.Context, workspaceID, name string) error {
	res, err := s.channels.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, name)})
	if err != nil {
		return mapError("channel", workspaceID, name, err)
	}
	if res.DeletedCount == 0 {
		return mapError("channel", workspaceID, name, mongo.ErrNoDocuments)
	}
	return nil
}

// --- Model Providers ---

func (s *Store) ListModelProviders(ctx context.Context, workspaceID string) ([]*agentsv1.ModelProvider, error) {
	docs, err := listInWorkspace(ctx, s.modelProviders, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.ModelProvider, 0, len(docs))
	for _, d := range docs {
		p := &agentsv1.ModelProvider{}
		if err := unmarshal(d.Spec, p); err != nil {
			return nil, fmt.Errorf("unmarshal model provider %q: %w", d.ID, err)
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *Store) ListModelProvidersAcrossWorkspaces(ctx context.Context) ([]*agentsv1.ModelProvider, error) {
	docs, err := listAll(ctx, s.modelProviders)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.ModelProvider, 0, len(docs))
	for _, d := range docs {
		p := &agentsv1.ModelProvider{}
		if err := unmarshal(d.Spec, p); err != nil {
			return nil, fmt.Errorf("unmarshal model provider %q: %w", d.ID, err)
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *Store) GetModelProvider(ctx context.Context, workspaceID, name string) (*agentsv1.ModelProvider, error) {
	var doc configDoc
	err := s.modelProviders.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, name)}).Decode(&doc)
	if err != nil {
		return nil, mapError("model provider", workspaceID, name, err)
	}
	p := &agentsv1.ModelProvider{}
	if err := unmarshal(doc.Spec, p); err != nil {
		return nil, fmt.Errorf("unmarshal model provider %q: %w", name, err)
	}
	return p, nil
}

func (s *Store) CreateModelProvider(ctx context.Context, workspaceID string, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error) {
	clone := proto.Clone(provider).(*agentsv1.ModelProvider)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetName()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	if _, err := s.modelProviders.InsertOne(ctx, doc); err != nil {
		return nil, mapError("model provider", workspaceID, clone.GetName(), err)
	}
	return clone, nil
}

func (s *Store) UpdateModelProvider(ctx context.Context, workspaceID string, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error) {
	clone := proto.Clone(provider).(*agentsv1.ModelProvider)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetName()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	res, err := s.modelProviders.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("model provider", workspaceID, clone.GetName(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("model provider", workspaceID, clone.GetName(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteModelProvider(ctx context.Context, workspaceID, name string) error {
	res, err := s.modelProviders.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, name)})
	if err != nil {
		return mapError("model provider", workspaceID, name, err)
	}
	if res.DeletedCount == 0 {
		return mapError("model provider", workspaceID, name, mongo.ErrNoDocuments)
	}
	return nil
}

// --- Notify Groups ---

func (s *Store) ListNotifyGroups(ctx context.Context, workspaceID string) ([]*agentsv1.NotifyGroup, error) {
	docs, err := listInWorkspace(ctx, s.notifyGroups, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.NotifyGroup, 0, len(docs))
	for _, d := range docs {
		g := &agentsv1.NotifyGroup{}
		if err := unmarshal(d.Spec, g); err != nil {
			return nil, fmt.Errorf("unmarshal notify group %q: %w", d.ID, err)
		}
		out = append(out, g)
	}
	return out, nil
}

func (s *Store) ListNotifyGroupsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.NotifyGroup, error) {
	docs, err := listAll(ctx, s.notifyGroups)
	if err != nil {
		return nil, err
	}
	out := make([]*agentsv1.NotifyGroup, 0, len(docs))
	for _, d := range docs {
		g := &agentsv1.NotifyGroup{}
		if err := unmarshal(d.Spec, g); err != nil {
			return nil, fmt.Errorf("unmarshal notify group %q: %w", d.ID, err)
		}
		out = append(out, g)
	}
	return out, nil
}

func (s *Store) GetNotifyGroup(ctx context.Context, workspaceID, name string) (*agentsv1.NotifyGroup, error) {
	var doc configDoc
	err := s.notifyGroups.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, name)}).Decode(&doc)
	if err != nil {
		return nil, mapError("notify group", workspaceID, name, err)
	}
	g := &agentsv1.NotifyGroup{}
	if err := unmarshal(doc.Spec, g); err != nil {
		return nil, fmt.Errorf("unmarshal notify group %q: %w", name, err)
	}
	return g, nil
}

func (s *Store) CreateNotifyGroup(ctx context.Context, workspaceID string, group *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error) {
	clone := proto.Clone(group).(*agentsv1.NotifyGroup)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetName()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	if _, err := s.notifyGroups.InsertOne(ctx, doc); err != nil {
		return nil, mapError("notify group", workspaceID, clone.GetName(), err)
	}
	return clone, nil
}

func (s *Store) UpdateNotifyGroup(ctx context.Context, workspaceID string, group *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error) {
	clone := proto.Clone(group).(*agentsv1.NotifyGroup)
	clone.WorkspaceId = workspaceID
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := configDoc{ID: compositeID(workspaceID, clone.GetName()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	res, err := s.notifyGroups.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("notify group", workspaceID, clone.GetName(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("notify group", workspaceID, clone.GetName(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteNotifyGroup(ctx context.Context, workspaceID, name string) error {
	res, err := s.notifyGroups.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, name)})
	if err != nil {
		return mapError("notify group", workspaceID, name, err)
	}
	if res.DeletedCount == 0 {
		return mapError("notify group", workspaceID, name, mongo.ErrNoDocuments)
	}
	return nil
}
