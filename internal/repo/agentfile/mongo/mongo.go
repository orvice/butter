package mongo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/agentfile"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	spacesCollection   = "agent_file_spaces"
	filesCollection    = "agent_files"
	versionsCollection = "agent_file_versions"
)

type specDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	SpaceID     string `bson:"space_id,omitempty"`
	Path        string `bson:"path,omitempty"`
	Name        string `bson:"name,omitempty"`
	Spec        string `bson:"spec"`
}

type versionDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	SpaceID     string `bson:"space_id"`
	FileID      string `bson:"file_id"`
	Path        string `bson:"path"`
	Version     int64  `bson:"version"`
	ContentKey  string `bson:"content_key"`
	ContentType string `bson:"content_type"`
	SizeBytes   int64  `bson:"size_bytes"`
	CreatedAt   int64  `bson:"created_at"`
}

// Store implements agentfile.Repository backed by MongoDB metadata plus a
// pluggable content store for version bodies.
type Store struct {
	spaces  *mongo.Collection
	files   *mongo.Collection
	version *mongo.Collection
	content agentfile.ContentStore
}

func New(db *mongo.Database, content agentfile.ContentStore) *Store {
	return &Store{
		spaces:  db.Collection(spacesCollection),
		files:   db.Collection(filesCollection),
		version: db.Collection(versionsCollection),
		content: content,
	}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	for _, model := range []struct {
		c    *mongo.Collection
		keys bson.D
	}{
		{s.spaces, bson.D{{Key: "workspace_id", Value: 1}, {Key: "name", Value: 1}}},
		{s.files, bson.D{{Key: "workspace_id", Value: 1}, {Key: "space_id", Value: 1}, {Key: "path", Value: 1}}},
		{s.version, bson.D{{Key: "workspace_id", Value: 1}, {Key: "space_id", Value: 1}, {Key: "path", Value: 1}, {Key: "version", Value: 1}}},
	} {
		if _, err := model.c.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: model.keys}); err != nil {
			return fmt.Errorf("create %s index: %w", model.c.Name(), err)
		}
	}
	return nil
}

func mapError(entity, ws, key string, err error) error {
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, agentfile.ErrAlreadyExists)
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, agentfile.ErrNotFound)
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
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal([]byte(data), m)
}

func compositeID(parts ...string) string {
	return strings.Join(parts, ":")
}

func cloneSpace(space *agentsv1.AgentFileSpace) *agentsv1.AgentFileSpace {
	return proto.Clone(space).(*agentsv1.AgentFileSpace)
}

func cloneFile(file *agentsv1.AgentFile) *agentsv1.AgentFile {
	return proto.Clone(file).(*agentsv1.AgentFile)
}

func (s *Store) ListSpaces(ctx context.Context, workspaceID string) ([]*agentsv1.AgentFileSpace, error) {
	cursor, err := s.spaces.Find(ctx, bson.M{"workspace_id": workspaceID})
	if err != nil {
		return nil, fmt.Errorf("list agent file spaces: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []specDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode agent file spaces: %w", err)
	}
	out := make([]*agentsv1.AgentFileSpace, 0, len(docs))
	for _, doc := range docs {
		space := &agentsv1.AgentFileSpace{}
		if err := unmarshal(doc.Spec, space); err != nil {
			return nil, fmt.Errorf("unmarshal agent file space %q: %w", doc.ID, err)
		}
		out = append(out, space)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (s *Store) GetSpace(ctx context.Context, workspaceID, id string) (*agentsv1.AgentFileSpace, error) {
	var doc specDoc
	err := s.spaces.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, id)}).Decode(&doc)
	if err != nil {
		return nil, mapError("agent file space", workspaceID, id, err)
	}
	space := &agentsv1.AgentFileSpace{}
	if err := unmarshal(doc.Spec, space); err != nil {
		return nil, fmt.Errorf("unmarshal agent file space %q: %w", id, err)
	}
	return space, nil
}

func (s *Store) CreateSpace(ctx context.Context, workspaceID string, space *agentsv1.AgentFileSpace) (*agentsv1.AgentFileSpace, error) {
	clone := cloneSpace(space)
	if clone.Id == "" {
		clone.Id = uuid.NewString()
	}
	now := timestamppb.New(time.Now().UTC())
	clone.WorkspaceId = workspaceID
	clone.CreatedAt = now
	clone.UpdatedAt = now
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := specDoc{ID: compositeID(workspaceID, clone.GetId()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	if _, err := s.spaces.InsertOne(ctx, doc); err != nil {
		return nil, mapError("agent file space", workspaceID, clone.GetId(), err)
	}
	return clone, nil
}

func (s *Store) UpdateSpace(ctx context.Context, workspaceID string, space *agentsv1.AgentFileSpace) (*agentsv1.AgentFileSpace, error) {
	prev, err := s.GetSpace(ctx, workspaceID, space.GetId())
	if err != nil {
		return nil, err
	}
	clone := cloneSpace(space)
	clone.WorkspaceId = workspaceID
	clone.CreatedAt = prev.GetCreatedAt()
	clone.UpdatedAt = timestamppb.New(time.Now().UTC())
	spec, err := marshal(clone)
	if err != nil {
		return nil, err
	}
	doc := specDoc{ID: compositeID(workspaceID, clone.GetId()), WorkspaceID: workspaceID, Name: clone.GetName(), Spec: spec}
	res, err := s.spaces.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError("agent file space", workspaceID, clone.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError("agent file space", workspaceID, clone.GetId(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) DeleteSpace(ctx context.Context, workspaceID, id string) error {
	if _, err := s.GetSpace(ctx, workspaceID, id); err != nil {
		return err
	}
	keys, err := s.contentKeys(ctx, workspaceID, id, "")
	if err != nil {
		return err
	}
	if _, err := s.files.DeleteMany(ctx, bson.M{"workspace_id": workspaceID, "space_id": id}); err != nil {
		return fmt.Errorf("delete agent files for space %q: %w", id, err)
	}
	if _, err := s.version.DeleteMany(ctx, bson.M{"workspace_id": workspaceID, "space_id": id}); err != nil {
		return fmt.Errorf("delete agent file versions for space %q: %w", id, err)
	}
	res, err := s.spaces.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, id)})
	if err != nil {
		return mapError("agent file space", workspaceID, id, err)
	}
	if res.DeletedCount == 0 {
		return mapError("agent file space", workspaceID, id, mongo.ErrNoDocuments)
	}
	return s.content.Delete(ctx, keys)
}

func (s *Store) ListFiles(ctx context.Context, workspaceID, spaceID, pathPrefix string) ([]*agentsv1.AgentFile, error) {
	prefix, err := agentfile.NormalizePrefix(pathPrefix)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetSpace(ctx, workspaceID, spaceID); err != nil {
		return nil, err
	}
	cursor, err := s.files.Find(ctx, bson.M{"workspace_id": workspaceID, "space_id": spaceID})
	if err != nil {
		return nil, fmt.Errorf("list agent files: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []specDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode agent files: %w", err)
	}
	out := make([]*agentsv1.AgentFile, 0, len(docs))
	for _, doc := range docs {
		if prefix != "" && doc.Path != prefix && !strings.HasPrefix(doc.Path, strings.TrimRight(prefix, "/")+"/") {
			continue
		}
		file := &agentsv1.AgentFile{}
		if err := unmarshal(doc.Spec, file); err != nil {
			return nil, fmt.Errorf("unmarshal agent file %q: %w", doc.ID, err)
		}
		out = append(out, file)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetPath() < out[j].GetPath() })
	return out, nil
}

func (s *Store) GetFile(ctx context.Context, workspaceID, spaceID, p string) (*agentsv1.AgentFile, error) {
	clean, err := agentfile.NormalizePath(p)
	if err != nil {
		return nil, err
	}
	var doc specDoc
	err = s.files.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, spaceID, clean)}).Decode(&doc)
	if err != nil {
		return nil, mapError("agent file", workspaceID, spaceID+":"+clean, err)
	}
	file := &agentsv1.AgentFile{}
	if err := unmarshal(doc.Spec, file); err != nil {
		return nil, fmt.Errorf("unmarshal agent file %q: %w", doc.ID, err)
	}
	return file, nil
}

func (s *Store) ReadFile(ctx context.Context, workspaceID, spaceID, p string, version int64) (*agentsv1.AgentFile, string, error) {
	file, err := s.GetFile(ctx, workspaceID, spaceID, p)
	if err != nil {
		return nil, "", err
	}
	if version == 0 {
		version = file.GetVersion()
	}
	var v versionDoc
	err = s.version.FindOne(ctx, bson.M{
		"workspace_id": workspaceID,
		"space_id":     spaceID,
		"path":         file.GetPath(),
		"version":      version,
	}).Decode(&v)
	if err != nil {
		return nil, "", mapError("agent file version", workspaceID, fmt.Sprintf("%s:%s:%d", spaceID, file.GetPath(), version), err)
	}
	content, contentType, err := s.content.Get(ctx, v.ContentKey)
	if err != nil {
		return nil, "", err
	}
	out := cloneFile(file)
	out.Version = v.Version
	out.ContentType = contentType
	out.SizeBytes = int64(len([]byte(content)))
	return out, content, nil
}

func (s *Store) WriteFile(ctx context.Context, workspaceID, spaceID, p, content, contentType string, metadata map[string]string) (*agentsv1.AgentFile, error) {
	clean, err := agentfile.NormalizePath(p)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetSpace(ctx, workspaceID, spaceID); err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}
	now := timestamppb.New(time.Now().UTC())
	file, err := s.GetFile(ctx, workspaceID, spaceID, clean)
	version := int64(1)
	if err != nil {
		if !errors.Is(err, agentfile.ErrNotFound) {
			return nil, err
		}
		file = &agentsv1.AgentFile{
			Id:          uuid.NewString(),
			WorkspaceId: workspaceID,
			SpaceId:     spaceID,
			Path:        clean,
			CreatedAt:   now,
		}
	} else {
		version = file.GetVersion() + 1
	}
	contentKey := fmt.Sprintf("agent-files/%s/%s/%s/%d", workspaceID, spaceID, file.GetId(), version)
	if err := s.content.Put(ctx, contentKey, content, contentType); err != nil {
		return nil, err
	}
	file.ContentType = contentType
	file.SizeBytes = int64(len([]byte(content)))
	file.Version = version
	file.Metadata = metadata
	file.UpdatedAt = now
	spec, err := marshal(file)
	if err != nil {
		return nil, err
	}
	doc := specDoc{
		ID:          compositeID(workspaceID, spaceID, clean),
		WorkspaceID: workspaceID,
		SpaceID:     spaceID,
		Path:        clean,
		Spec:        spec,
	}
	if version == 1 {
		if _, err := s.files.InsertOne(ctx, doc); err != nil {
			return nil, mapError("agent file", workspaceID, spaceID+":"+clean, err)
		}
	} else {
		res, err := s.files.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
		if err != nil {
			return nil, mapError("agent file", workspaceID, spaceID+":"+clean, err)
		}
		if res.MatchedCount == 0 {
			return nil, mapError("agent file", workspaceID, spaceID+":"+clean, mongo.ErrNoDocuments)
		}
	}
	v := versionDoc{
		ID:          compositeID(workspaceID, spaceID, clean, fmt.Sprintf("%d", version)),
		WorkspaceID: workspaceID,
		SpaceID:     spaceID,
		FileID:      file.GetId(),
		Path:        clean,
		Version:     version,
		ContentKey:  contentKey,
		ContentType: contentType,
		SizeBytes:   int64(len([]byte(content))),
		CreatedAt:   now.AsTime().Unix(),
	}
	if _, err := s.version.InsertOne(ctx, v); err != nil {
		return nil, mapError("agent file version", workspaceID, v.ID, err)
	}
	return cloneFile(file), nil
}

func (s *Store) DeleteFile(ctx context.Context, workspaceID, spaceID, p string) error {
	clean, err := agentfile.NormalizePath(p)
	if err != nil {
		return err
	}
	if _, err := s.GetFile(ctx, workspaceID, spaceID, clean); err != nil {
		return err
	}
	keys, err := s.contentKeys(ctx, workspaceID, spaceID, clean)
	if err != nil {
		return err
	}
	if _, err := s.version.DeleteMany(ctx, bson.M{"workspace_id": workspaceID, "space_id": spaceID, "path": clean}); err != nil {
		return fmt.Errorf("delete agent file versions: %w", err)
	}
	res, err := s.files.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, spaceID, clean)})
	if err != nil {
		return mapError("agent file", workspaceID, spaceID+":"+clean, err)
	}
	if res.DeletedCount == 0 {
		return mapError("agent file", workspaceID, spaceID+":"+clean, mongo.ErrNoDocuments)
	}
	return s.content.Delete(ctx, keys)
}

func (s *Store) SearchFiles(ctx context.Context, workspaceID, spaceID, query string, limit int) ([]*agentsv1.AgentFileSearchResult, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	files, err := s.ListFiles(ctx, workspaceID, spaceID, "")
	if err != nil {
		return nil, err
	}
	var out []*agentsv1.AgentFileSearchResult
	for _, file := range files {
		_, content, err := s.ReadFile(ctx, workspaceID, spaceID, file.GetPath(), 0)
		if err != nil {
			return nil, err
		}
		idx := strings.Index(strings.ToLower(content), query)
		if idx < 0 {
			continue
		}
		out = append(out, &agentsv1.AgentFileSearchResult{
			File:     file,
			Snippets: []string{snippet(content, idx, len(query))},
		})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *Store) contentKeys(ctx context.Context, workspaceID, spaceID, p string) ([]string, error) {
	filter := bson.M{"workspace_id": workspaceID, "space_id": spaceID}
	if p != "" {
		filter["path"] = p
	}
	cursor, err := s.version.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list agent file version keys: %w", err)
	}
	defer cursor.Close(ctx)
	var versions []versionDoc
	if err := cursor.All(ctx, &versions); err != nil {
		return nil, fmt.Errorf("decode agent file version keys: %w", err)
	}
	keys := make([]string, 0, len(versions))
	for _, v := range versions {
		keys = append(keys, v.ContentKey)
	}
	return keys, nil
}

func snippet(content string, idx, queryLen int) string {
	start := idx - 80
	if start < 0 {
		start = 0
	}
	end := idx + queryLen + 80
	if end > len(content) {
		end = len(content)
	}
	return content[start:end]
}

var _ agentfile.Repository = (*Store)(nil)
