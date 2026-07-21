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

const (
	skillsCollection         = "skills"
	skillResourcesCollection = "skill_resources"
)

type skillDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	Name        string `bson:"name"`
	Spec        string `bson:"spec"`
	ContentKey  string `bson:"content_key"`
}

// resourceDoc is the Mongo path index for one skill resource (ADR 0004,
// issue #154): ListResources is served entirely from these documents so the
// hot path never issues an S3 List; content lives behind the ContentStore.
type resourceDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	SkillName   string `bson:"skill_name"`
	Path        string `bson:"path"`
	Spec        string `bson:"spec"`
	ContentKey  string `bson:"content_key"`
}

// Store implements skill.Repository backed by MongoDB metadata plus a
// pluggable content store for SKILL.md bodies and resource files.
type Store struct {
	skills    *mongo.Collection
	resources *mongo.Collection
	content   skillrepo.ContentStore
}

func New(db *mongo.Database, content skillrepo.ContentStore) *Store {
	return &Store{
		skills:    db.Collection(skillsCollection),
		resources: db.Collection(skillResourcesCollection),
		content:   content,
	}
}

// EnsureIndexes creates the unique (workspace_id, name) index that both
// enforces per-workspace name uniqueness and serves the List hot path, plus
// the (workspace_id, skill_name) index serving per-skill resource listings.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.skills.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create skills index: %w", err)
	}
	_, err = s.resources.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "skill_name", Value: 1}, {Key: "path", Value: 1}},
	})
	if err != nil {
		return fmt.Errorf("create skill resources index: %w", err)
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
	// Collect resource content keys up front, before deleting anything. A
	// resource lives at an arbitrary path, so unlike the SKILL.md key it is
	// not reconstructable from (workspace, name) — losing the key orphans
	// the object permanently. If the lookup fails we abort with nothing
	// deleted, so a retry still finds the keys in the path index.
	resourceKeys, err := s.resourceContentKeys(ctx, workspaceID, name)
	if err != nil {
		return fmt.Errorf("skill %q (workspace %q): collect resource keys: %w", name, workspaceID, err)
	}

	res, err := s.skills.DeleteOne(ctx, bson.M{"_id": doc.ID})
	if err != nil {
		return mapError(workspaceID, name, err)
	}
	if res.DeletedCount == 0 {
		return mapError(workspaceID, name, mongo.ErrNoDocuments)
	}

	// Metadata is the source of truth: the skill is gone once the document
	// is deleted. Delete content before the path index so the index remains
	// a recovery record if content deletion fails; a content-store failure
	// is logged rather than failing the delete (a retry would see the skill
	// already gone).
	if err := s.content.Delete(ctx, append(resourceKeys, doc.ContentKey)); err != nil {
		log.FromContext(ctx).Warn("skill deleted but stored content removal failed; objects orphaned",
			"workspace_id", workspaceID,
			"skill", name,
			"content_key", doc.ContentKey,
			"resource_count", len(resourceKeys),
			"err", err,
		)
		return nil
	}
	if _, err := s.resources.DeleteMany(ctx, bson.M{"workspace_id": workspaceID, "skill_name": name}); err != nil {
		log.FromContext(ctx).Warn("skill content removed but resource index removal failed; stale index entries remain",
			"workspace_id", workspaceID,
			"skill", name,
			"err", err,
		)
	}
	return nil
}

func resourceDocID(workspaceID, skillName, path string) string {
	return strings.Join([]string{workspaceID, skillName, path}, ":")
}

func mapResourceError(ws, skill, path string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("skill %q resource %q (workspace %q): %w", skill, path, ws, skillrepo.ErrNotFound)
	}
	return fmt.Errorf("skill %q resource %q (workspace %q): %w", skill, path, ws, err)
}

func decodeResource(doc resourceDoc) (*agentsv1.SkillResource, error) {
	res := &agentsv1.SkillResource{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), res); err != nil {
		return nil, fmt.Errorf("unmarshal skill resource %q: %w", doc.ID, err)
	}
	return res, nil
}

func encodeResourceDoc(workspaceID, skillName string, res *agentsv1.SkillResource) (resourceDoc, error) {
	spec, err := protojson.Marshal(res)
	if err != nil {
		return resourceDoc{}, fmt.Errorf("marshal skill resource %q: %w", res.GetPath(), err)
	}
	return resourceDoc{
		ID:          resourceDocID(workspaceID, skillName, res.GetPath()),
		WorkspaceID: workspaceID,
		SkillName:   skillName,
		Path:        res.GetPath(),
		Spec:        string(spec),
		ContentKey:  skillrepo.ResourceContentKey(workspaceID, skillName, res.GetPath()),
	}, nil
}

// resourceContentKeys lists the content keys of every resource of a skill,
// straight from the path index.
func (s *Store) resourceContentKeys(ctx context.Context, workspaceID, skillName string) ([]string, error) {
	cursor, err := s.resources.Find(ctx, bson.M{"workspace_id": workspaceID, "skill_name": skillName})
	if err != nil {
		return nil, fmt.Errorf("list skill resource keys: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []resourceDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode skill resource keys: %w", err)
	}
	keys := make([]string, 0, len(docs))
	for _, doc := range docs {
		keys = append(keys, doc.ContentKey)
	}
	return keys, nil
}

func (s *Store) ListResources(ctx context.Context, workspaceID, skillName string) ([]*agentsv1.SkillResource, error) {
	if _, err := s.findDoc(ctx, workspaceID, skillName); err != nil {
		return nil, err
	}
	cursor, err := s.resources.Find(ctx, bson.M{"workspace_id": workspaceID, "skill_name": skillName})
	if err != nil {
		return nil, fmt.Errorf("list skill resources: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []resourceDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode skill resources: %w", err)
	}
	out := make([]*agentsv1.SkillResource, 0, len(docs))
	for _, doc := range docs {
		res, err := decodeResource(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetPath() < out[j].GetPath() })
	return out, nil
}

func (s *Store) findResourceDoc(ctx context.Context, workspaceID, skillName, path string) (resourceDoc, error) {
	var doc resourceDoc
	err := s.resources.FindOne(ctx, bson.M{"_id": resourceDocID(workspaceID, skillName, path)}).Decode(&doc)
	if err != nil {
		return resourceDoc{}, mapResourceError(workspaceID, skillName, path, err)
	}
	return doc, nil
}

func (s *Store) GetResource(ctx context.Context, workspaceID, skillName, path string) (*agentsv1.SkillResource, []byte, error) {
	if _, err := s.findDoc(ctx, workspaceID, skillName); err != nil {
		return nil, nil, err
	}
	doc, err := s.findResourceDoc(ctx, workspaceID, skillName, path)
	if err != nil {
		return nil, nil, err
	}
	res, err := decodeResource(doc)
	if err != nil {
		return nil, nil, err
	}
	content, err := s.content.Get(ctx, doc.ContentKey)
	if err != nil {
		return nil, nil, err
	}
	return res, []byte(content), nil
}

func (s *Store) PutResource(ctx context.Context, workspaceID, skillName string, resource *agentsv1.SkillResource, content []byte) (*agentsv1.SkillResource, error) {
	if _, err := s.findDoc(ctx, workspaceID, skillName); err != nil {
		return nil, err
	}
	clone := proto.Clone(resource).(*agentsv1.SkillResource)
	now := timestamppb.New(time.Now().UTC())
	clone.SizeBytes = int64(len(content))
	clone.UpdatedAt = now
	clone.CreatedAt = now
	// Preserve created_at across overwrites. Only a genuine miss falls
	// through to the fresh `now`; a transient lookup error must fail the
	// write rather than silently reset the timestamp.
	prev, err := s.findResourceDoc(ctx, workspaceID, skillName, clone.GetPath())
	switch {
	case err == nil:
		prevRes, derr := decodeResource(prev)
		if derr != nil {
			return nil, derr
		}
		clone.CreatedAt = prevRes.GetCreatedAt()
	case errors.Is(err, skillrepo.ErrNotFound):
		// Fresh resource: created_at stays at now.
	default:
		return nil, err
	}
	doc, err := encodeResourceDoc(workspaceID, skillName, clone)
	if err != nil {
		return nil, err
	}
	// Content is written before the index entry so a listed resource is
	// always loadable; the inverse failure only orphans an object the next
	// put to the same path overwrites.
	if err := s.content.Put(ctx, doc.ContentKey, string(content)); err != nil {
		return nil, err
	}
	if _, err := s.resources.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true)); err != nil {
		return nil, mapResourceError(workspaceID, skillName, clone.GetPath(), err)
	}
	return clone, nil
}

func (s *Store) DeleteResource(ctx context.Context, workspaceID, skillName, path string) error {
	if _, err := s.findDoc(ctx, workspaceID, skillName); err != nil {
		return err
	}
	doc, err := s.findResourceDoc(ctx, workspaceID, skillName, path)
	if err != nil {
		return err
	}
	res, err := s.resources.DeleteOne(ctx, bson.M{"_id": doc.ID})
	if err != nil {
		return mapResourceError(workspaceID, skillName, path, err)
	}
	if res.DeletedCount == 0 {
		return mapResourceError(workspaceID, skillName, path, mongo.ErrNoDocuments)
	}
	if err := s.content.Delete(ctx, []string{doc.ContentKey}); err != nil {
		log.FromContext(ctx).Warn("skill resource deleted but stored content removal failed; object orphaned",
			"workspace_id", workspaceID,
			"skill", skillName,
			"path", path,
			"content_key", doc.ContentKey,
			"err", err,
		)
	}
	return nil
}

var _ skillrepo.Repository = (*Store)(nil)
