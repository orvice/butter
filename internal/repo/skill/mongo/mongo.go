// Package mongo implements skill.Repository per ADR 0004: parsed frontmatter
// metadata lives in MongoDB so the per-LLM-turn ListFrontmatters lookup is a
// single indexed query, while SKILL.md bodies live behind skill.ContentStore.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const skillsCollection = "skills"

type skillDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	Name        string `bson:"name"`
	Spec        string `bson:"spec"`
	ContentKey  string `bson:"content_key"`
}

// Store implements skill.Repository backed by MongoDB metadata plus a
// pluggable content store for SKILL.md bodies.
type Store struct {
	skills  *mongo.Collection
	content skillrepo.ContentStore
}

func New(db *mongo.Database, content skillrepo.ContentStore) *Store {
	return &Store{
		skills:  db.Collection(skillsCollection),
		content: content,
	}
}

// EnsureIndexes creates the unique (workspace_id, name) index that both
// enforces per-workspace name uniqueness and serves the List hot path.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.skills.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create skills index: %w", err)
	}
	return nil
}

func docID(workspaceID, name string) string {
	return strings.Join([]string{workspaceID, name}, ":")
}

func mapError(ws, name string, err error) error {
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("skill %q (workspace %q): %w", name, ws, skillrepo.ErrAlreadyExists)
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("skill %q (workspace %q): %w", name, ws, skillrepo.ErrNotFound)
	}
	return fmt.Errorf("skill %q (workspace %q): %w", name, ws, err)
}

func decodeSkill(doc skillDoc) (*agentsv1.Skill, error) {
	sk := &agentsv1.Skill{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), sk); err != nil {
		return nil, fmt.Errorf("unmarshal skill %q: %w", doc.ID, err)
	}
	return sk, nil
}

func encodeDoc(workspaceID string, sk *agentsv1.Skill) (skillDoc, error) {
	spec, err := protojson.Marshal(sk)
	if err != nil {
		return skillDoc{}, fmt.Errorf("marshal skill %q: %w", sk.GetName(), err)
	}
	return skillDoc{
		ID:          docID(workspaceID, sk.GetName()),
		WorkspaceID: workspaceID,
		Name:        sk.GetName(),
		Spec:        string(spec),
		ContentKey:  skillrepo.ContentKey(workspaceID, sk.GetName()),
	}, nil
}

func (s *Store) List(ctx context.Context, workspaceID string) ([]*agentsv1.Skill, error) {
	cursor, err := s.skills.Find(ctx, bson.M{"workspace_id": workspaceID})
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []skillDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode skills: %w", err)
	}
	out := make([]*agentsv1.Skill, 0, len(docs))
	for _, doc := range docs {
		sk, err := decodeSkill(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, sk)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (s *Store) findDoc(ctx context.Context, workspaceID, name string) (skillDoc, error) {
	var doc skillDoc
	err := s.skills.FindOne(ctx, bson.M{"_id": docID(workspaceID, name)}).Decode(&doc)
	if err != nil {
		return skillDoc{}, mapError(workspaceID, name, err)
	}
	return doc, nil
}

func (s *Store) Get(ctx context.Context, workspaceID, name string) (*agentsv1.Skill, error) {
	doc, err := s.findDoc(ctx, workspaceID, name)
	if err != nil {
		return nil, err
	}
	return decodeSkill(doc)
}

func (s *Store) GetSkillMD(ctx context.Context, workspaceID, name string) (string, error) {
	doc, err := s.findDoc(ctx, workspaceID, name)
	if err != nil {
		return "", err
	}
	return s.content.Get(ctx, doc.ContentKey)
}

func (s *Store) Create(ctx context.Context, workspaceID string, sk *agentsv1.Skill, skillMD string) (*agentsv1.Skill, error) {
	clone := proto.Clone(sk).(*agentsv1.Skill)
	now := timestamppb.New(time.Now().UTC())
	clone.WorkspaceId = workspaceID
	clone.CreatedAt = now
	clone.UpdatedAt = now
	doc, err := encodeDoc(workspaceID, clone)
	if err != nil {
		return nil, err
	}
	// Insert metadata before writing content: the content key is
	// deterministic per (workspace, name), so a duplicate create must fail
	// here rather than overwrite the existing skill's stored SKILL.md.
	if _, err := s.skills.InsertOne(ctx, doc); err != nil {
		return nil, mapError(workspaceID, clone.GetName(), err)
	}
	if err := s.content.Put(ctx, doc.ContentKey, skillMD); err != nil {
		_, _ = s.skills.DeleteOne(ctx, bson.M{"_id": doc.ID})
		return nil, err
	}
	return clone, nil
}

func (s *Store) Update(ctx context.Context, workspaceID string, sk *agentsv1.Skill, skillMD string) (*agentsv1.Skill, error) {
	prev, err := s.Get(ctx, workspaceID, sk.GetName())
	if err != nil {
		return nil, err
	}
	clone := proto.Clone(sk).(*agentsv1.Skill)
	clone.WorkspaceId = workspaceID
	clone.CreatedAt = prev.GetCreatedAt()
	clone.UpdatedAt = timestamppb.New(time.Now().UTC())
	doc, err := encodeDoc(workspaceID, clone)
	if err != nil {
		return nil, err
	}
	// Content is written before metadata so readers never see updated
	// frontmatter pointing at a stale body. The inverse failure (content
	// written, replace loses to a concurrent delete) only orphans an object
	// under a key the next create overwrites.
	if err := s.content.Put(ctx, doc.ContentKey, skillMD); err != nil {
		return nil, err
	}
	res, err := s.skills.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return nil, mapError(workspaceID, clone.GetName(), err)
	}
	if res.MatchedCount == 0 {
		return nil, mapError(workspaceID, clone.GetName(), mongo.ErrNoDocuments)
	}
	return clone, nil
}

func (s *Store) Delete(ctx context.Context, workspaceID, name string) error {
	doc, err := s.findDoc(ctx, workspaceID, name)
	if err != nil {
		return err
	}
	res, err := s.skills.DeleteOne(ctx, bson.M{"_id": doc.ID})
	if err != nil {
		return mapError(workspaceID, name, err)
	}
	if res.DeletedCount == 0 {
		return mapError(workspaceID, name, mongo.ErrNoDocuments)
	}
	// Metadata is the source of truth: the skill is gone once the document
	// is deleted. A content-store failure here would leave a retry seeing
	// NotFound, so log the orphaned object instead of failing the delete;
	// recreating the same name overwrites the same key.
	if err := s.content.Delete(ctx, []string{doc.ContentKey}); err != nil {
		log.FromContext(ctx).Warn("skill deleted but stored content removal failed; object orphaned",
			"workspace_id", workspaceID,
			"skill", name,
			"content_key", doc.ContentKey,
			"err", err,
		)
	}
	return nil
}

var _ skillrepo.Repository = (*Store)(nil)
