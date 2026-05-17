package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	agentsCollection         = "config_agents"
	mcpServersCollection     = "config_mcpservers"
	remoteAgentsCollection   = "config_remoteagents"
	channelsCollection       = "config_channels"
	modelProvidersCollection = "config_modelproviders"
)

// configDoc is the generic MongoDB document for a config entity.
type configDoc struct {
	ID   string `bson:"_id"`
	Spec string `bson:"spec"` // protojson-encoded
}

// Store implements all config repository interfaces backed by MongoDB.
type Store struct {
	agents         *mongo.Collection
	mcpServers     *mongo.Collection
	remoteAgents   *mongo.Collection
	channels       *mongo.Collection
	modelProviders *mongo.Collection
}

func New(db *mongo.Database) *Store {
	return &Store{
		agents:         db.Collection(agentsCollection),
		mcpServers:     db.Collection(mcpServersCollection),
		remoteAgents:   db.Collection(remoteAgentsCollection),
		channels:       db.Collection(channelsCollection),
		modelProviders: db.Collection(modelProvidersCollection),
	}
}

func mapError(entity, key string, err error) error {
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%s %q: %w", entity, key, configrepo.ErrAlreadyExists)
	}
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("%s %q: %w", entity, key, configrepo.ErrNotFound)
	}
	return fmt.Errorf("%s %q: %w", entity, key, err)
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

// --- Agents ---

func (s *Store) ListAgents(ctx context.Context) ([]*agentsv1.Agent, error) {
	cursor, err := s.agents.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []configDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode agents: %w", err)
	}

	result := make([]*agentsv1.Agent, 0, len(docs))
	for _, d := range docs {
		a := &agentsv1.Agent{}
		if err := unmarshal(d.Spec, a); err != nil {
			return nil, fmt.Errorf("unmarshal agent %q: %w", d.ID, err)
		}
		result = append(result, a)
	}
	return result, nil
}

func (s *Store) GetAgent(ctx context.Context, name string) (*agentsv1.Agent, error) {
	var doc configDoc
	err := s.agents.FindOne(ctx, bson.M{"_id": name}).Decode(&doc)
	if err != nil {
		return nil, mapError("agent", name, err)
	}
	a := &agentsv1.Agent{}
	if err := unmarshal(doc.Spec, a); err != nil {
		return nil, fmt.Errorf("unmarshal agent %q: %w", name, err)
	}
	return a, nil
}

func (s *Store) CreateAgent(ctx context.Context, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	spec, err := marshal(agent)
	if err != nil {
		return nil, err
	}
	_, err = s.agents.InsertOne(ctx, configDoc{ID: agent.GetName(), Spec: spec})
	if err != nil {
		return nil, mapError("agent", agent.GetName(), err)
	}
	return proto.Clone(agent).(*agentsv1.Agent), nil
}

func (s *Store) UpdateAgent(ctx context.Context, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	spec, err := marshal(agent)
	if err != nil {
		return nil, err
	}
	result, err := s.agents.ReplaceOne(ctx, bson.M{"_id": agent.GetName()}, configDoc{ID: agent.GetName(), Spec: spec})
	if err != nil {
		return nil, mapError("agent", agent.GetName(), err)
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("agent %q: %w", agent.GetName(), configrepo.ErrNotFound)
	}
	return proto.Clone(agent).(*agentsv1.Agent), nil
}

func (s *Store) DeleteAgent(ctx context.Context, name string) error {
	result, err := s.agents.DeleteOne(ctx, bson.M{"_id": name})
	if err != nil {
		return mapError("agent", name, err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("agent %q: %w", name, configrepo.ErrNotFound)
	}
	return nil
}

// --- MCP Servers ---

func (s *Store) ListMCPServers(ctx context.Context) ([]*agentsv1.MCPServer, error) {
	cursor, err := s.mcpServers.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list mcp servers: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []configDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode mcp servers: %w", err)
	}

	result := make([]*agentsv1.MCPServer, 0, len(docs))
	for _, d := range docs {
		m := &agentsv1.MCPServer{}
		if err := unmarshal(d.Spec, m); err != nil {
			return nil, fmt.Errorf("unmarshal mcp server %q: %w", d.ID, err)
		}
		result = append(result, m)
	}
	return result, nil
}

func (s *Store) GetMCPServer(ctx context.Context, id string) (*agentsv1.MCPServer, error) {
	var doc configDoc
	err := s.mcpServers.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		return nil, mapError("mcp server", id, err)
	}
	m := &agentsv1.MCPServer{}
	if err := unmarshal(doc.Spec, m); err != nil {
		return nil, fmt.Errorf("unmarshal mcp server %q: %w", id, err)
	}
	return m, nil
}

func (s *Store) CreateMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	spec, err := marshal(server)
	if err != nil {
		return nil, err
	}
	_, err = s.mcpServers.InsertOne(ctx, configDoc{ID: server.GetId(), Spec: spec})
	if err != nil {
		return nil, mapError("mcp server", server.GetId(), err)
	}
	return proto.Clone(server).(*agentsv1.MCPServer), nil
}

func (s *Store) UpdateMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	spec, err := marshal(server)
	if err != nil {
		return nil, err
	}
	result, err := s.mcpServers.ReplaceOne(ctx, bson.M{"_id": server.GetId()}, configDoc{ID: server.GetId(), Spec: spec})
	if err != nil {
		return nil, mapError("mcp server", server.GetId(), err)
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("mcp server %q: %w", server.GetId(), configrepo.ErrNotFound)
	}
	return proto.Clone(server).(*agentsv1.MCPServer), nil
}

func (s *Store) DeleteMCPServer(ctx context.Context, id string) error {
	result, err := s.mcpServers.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return mapError("mcp server", id, err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("mcp server %q: %w", id, configrepo.ErrNotFound)
	}
	return nil
}

// --- Remote Agents ---

func (s *Store) ListRemoteAgents(ctx context.Context) ([]*agentsv1.RemoteAgent, error) {
	cursor, err := s.remoteAgents.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list remote agents: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []configDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode remote agents: %w", err)
	}

	result := make([]*agentsv1.RemoteAgent, 0, len(docs))
	for _, d := range docs {
		r := &agentsv1.RemoteAgent{}
		if err := unmarshal(d.Spec, r); err != nil {
			return nil, fmt.Errorf("unmarshal remote agent %q: %w", d.ID, err)
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *Store) GetRemoteAgent(ctx context.Context, id string) (*agentsv1.RemoteAgent, error) {
	var doc configDoc
	err := s.remoteAgents.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		return nil, mapError("remote agent", id, err)
	}
	r := &agentsv1.RemoteAgent{}
	if err := unmarshal(doc.Spec, r); err != nil {
		return nil, fmt.Errorf("unmarshal remote agent %q: %w", id, err)
	}
	return r, nil
}

func (s *Store) CreateRemoteAgent(ctx context.Context, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	spec, err := marshal(agent)
	if err != nil {
		return nil, err
	}
	_, err = s.remoteAgents.InsertOne(ctx, configDoc{ID: agent.GetId(), Spec: spec})
	if err != nil {
		return nil, mapError("remote agent", agent.GetId(), err)
	}
	return proto.Clone(agent).(*agentsv1.RemoteAgent), nil
}

func (s *Store) UpdateRemoteAgent(ctx context.Context, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	spec, err := marshal(agent)
	if err != nil {
		return nil, err
	}
	result, err := s.remoteAgents.ReplaceOne(ctx, bson.M{"_id": agent.GetId()}, configDoc{ID: agent.GetId(), Spec: spec})
	if err != nil {
		return nil, mapError("remote agent", agent.GetId(), err)
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("remote agent %q: %w", agent.GetId(), configrepo.ErrNotFound)
	}
	return proto.Clone(agent).(*agentsv1.RemoteAgent), nil
}

func (s *Store) DeleteRemoteAgent(ctx context.Context, id string) error {
	result, err := s.remoteAgents.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return mapError("remote agent", id, err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("remote agent %q: %w", id, configrepo.ErrNotFound)
	}
	return nil
}

// --- Channels ---

func (s *Store) ListChannels(ctx context.Context) ([]*agentsv1.AgentChannel, error) {
	cursor, err := s.channels.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []configDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode channels: %w", err)
	}

	result := make([]*agentsv1.AgentChannel, 0, len(docs))
	for _, d := range docs {
		c := &agentsv1.AgentChannel{}
		if err := unmarshal(d.Spec, c); err != nil {
			return nil, fmt.Errorf("unmarshal channel %q: %w", d.ID, err)
		}
		result = append(result, c)
	}
	return result, nil
}

func (s *Store) GetChannel(ctx context.Context, name string) (*agentsv1.AgentChannel, error) {
	var doc configDoc
	err := s.channels.FindOne(ctx, bson.M{"_id": name}).Decode(&doc)
	if err != nil {
		return nil, mapError("channel", name, err)
	}
	c := &agentsv1.AgentChannel{}
	if err := unmarshal(doc.Spec, c); err != nil {
		return nil, fmt.Errorf("unmarshal channel %q: %w", name, err)
	}
	return c, nil
}

func (s *Store) CreateChannel(ctx context.Context, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	spec, err := marshal(channel)
	if err != nil {
		return nil, err
	}
	_, err = s.channels.InsertOne(ctx, configDoc{ID: channel.GetName(), Spec: spec})
	if err != nil {
		return nil, mapError("channel", channel.GetName(), err)
	}
	return proto.Clone(channel).(*agentsv1.AgentChannel), nil
}

func (s *Store) UpdateChannel(ctx context.Context, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	spec, err := marshal(channel)
	if err != nil {
		return nil, err
	}
	result, err := s.channels.ReplaceOne(ctx, bson.M{"_id": channel.GetName()}, configDoc{ID: channel.GetName(), Spec: spec})
	if err != nil {
		return nil, mapError("channel", channel.GetName(), err)
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("channel %q: %w", channel.GetName(), configrepo.ErrNotFound)
	}
	return proto.Clone(channel).(*agentsv1.AgentChannel), nil
}

func (s *Store) DeleteChannel(ctx context.Context, name string) error {
	result, err := s.channels.DeleteOne(ctx, bson.M{"_id": name})
	if err != nil {
		return mapError("channel", name, err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("channel %q: %w", name, configrepo.ErrNotFound)
	}
	return nil
}

// --- Model Providers ---

func (s *Store) ListModelProviders(ctx context.Context) ([]*agentsv1.ModelProvider, error) {
	cursor, err := s.modelProviders.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list model providers: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []configDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode model providers: %w", err)
	}

	result := make([]*agentsv1.ModelProvider, 0, len(docs))
	for _, d := range docs {
		p := &agentsv1.ModelProvider{}
		if err := unmarshal(d.Spec, p); err != nil {
			return nil, fmt.Errorf("unmarshal model provider %q: %w", d.ID, err)
		}
		result = append(result, p)
	}
	return result, nil
}

func (s *Store) GetModelProvider(ctx context.Context, name string) (*agentsv1.ModelProvider, error) {
	var doc configDoc
	err := s.modelProviders.FindOne(ctx, bson.M{"_id": name}).Decode(&doc)
	if err != nil {
		return nil, mapError("model provider", name, err)
	}
	p := &agentsv1.ModelProvider{}
	if err := unmarshal(doc.Spec, p); err != nil {
		return nil, fmt.Errorf("unmarshal model provider %q: %w", name, err)
	}
	return p, nil
}

func (s *Store) CreateModelProvider(ctx context.Context, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error) {
	spec, err := marshal(provider)
	if err != nil {
		return nil, err
	}
	_, err = s.modelProviders.InsertOne(ctx, configDoc{ID: provider.GetName(), Spec: spec})
	if err != nil {
		return nil, mapError("model provider", provider.GetName(), err)
	}
	return proto.Clone(provider).(*agentsv1.ModelProvider), nil
}

func (s *Store) UpdateModelProvider(ctx context.Context, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error) {
	spec, err := marshal(provider)
	if err != nil {
		return nil, err
	}
	result, err := s.modelProviders.ReplaceOne(ctx, bson.M{"_id": provider.GetName()}, configDoc{ID: provider.GetName(), Spec: spec})
	if err != nil {
		return nil, mapError("model provider", provider.GetName(), err)
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("model provider %q: %w", provider.GetName(), configrepo.ErrNotFound)
	}
	return proto.Clone(provider).(*agentsv1.ModelProvider), nil
}

func (s *Store) DeleteModelProvider(ctx context.Context, name string) error {
	result, err := s.modelProviders.DeleteOne(ctx, bson.M{"_id": name})
	if err != nil {
		return mapError("model provider", name, err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("model provider %q: %w", name, configrepo.ErrNotFound)
	}
	return nil
}

// Seed upserts config entries into MongoDB. Existing entries with matching keys are overwritten.
func (s *Store) Seed(ctx context.Context, agents []agentsv1.Agent, mcpServers []agentsv1.MCPServer, remoteAgents []agentsv1.RemoteAgent, channels []agentsv1.AgentChannel, modelProviders []agentsv1.ModelProvider) error {
	for i := range agents {
		spec, err := marshal(&agents[i])
		if err != nil {
			return err
		}
		doc := configDoc{ID: agents[i].GetName(), Spec: spec}
		_, err = s.agents.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, replaceUpsert())
		if err != nil {
			return fmt.Errorf("seed agent %q: %w", doc.ID, err)
		}
	}
	for i := range mcpServers {
		spec, err := marshal(&mcpServers[i])
		if err != nil {
			return err
		}
		doc := configDoc{ID: mcpServers[i].GetId(), Spec: spec}
		_, err = s.mcpServers.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, replaceUpsert())
		if err != nil {
			return fmt.Errorf("seed mcp server %q: %w", doc.ID, err)
		}
	}
	for i := range remoteAgents {
		spec, err := marshal(&remoteAgents[i])
		if err != nil {
			return err
		}
		doc := configDoc{ID: remoteAgents[i].GetId(), Spec: spec}
		_, err = s.remoteAgents.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, replaceUpsert())
		if err != nil {
			return fmt.Errorf("seed remote agent %q: %w", doc.ID, err)
		}
	}
	for i := range channels {
		spec, err := marshal(&channels[i])
		if err != nil {
			return err
		}
		doc := configDoc{ID: channels[i].GetName(), Spec: spec}
		_, err = s.channels.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, replaceUpsert())
		if err != nil {
			return fmt.Errorf("seed channel %q: %w", doc.ID, err)
		}
	}
	for i := range modelProviders {
		spec, err := marshal(&modelProviders[i])
		if err != nil {
			return err
		}
		doc := configDoc{ID: modelProviders[i].GetName(), Spec: spec}
		_, err = s.modelProviders.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, replaceUpsert())
		if err != nil {
			return fmt.Errorf("seed model provider %q: %w", doc.ID, err)
		}
	}
	return nil
}

func replaceUpsert() *options.ReplaceOptionsBuilder {
	return options.Replace().SetUpsert(true)
}
