// Package mongo implements butterbox.Repository backed by MongoDB.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const butterBoxesCollection = "butter_boxes"

// boxDoc keeps the credential in dedicated columns, never inside the
// protojson spec, so the secret cannot leak through a decoded public model
// and the derived credential fields cannot contradict storage.
type boxDoc struct {
	ID                string    `bson:"_id"`
	WorkspaceID       string    `bson:"workspace_id"`
	Name              string    `bson:"name"`
	Spec              string    `bson:"spec"`
	Credential        string    `bson:"credential,omitempty"`
	CredentialKeyID   string    `bson:"credential_key_id,omitempty"`
	CredentialUpdated time.Time `bson:"credential_updated_at,omitempty"`
}

// Store implements butterbox.Repository backed by MongoDB.
type Store struct {
	boxes *mongo.Collection
}

var _ butterboxrepo.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{boxes: db.Collection(butterBoxesCollection)}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.boxes.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_workspace_name"),
		},
	})
	if err != nil {
		return fmt.Errorf("butterbox indexes: %w", err)
	}
	return nil
}

func mapError(id string, err error) error {
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrAlreadyExists)
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNotFound)
	}
	return fmt.Errorf("butterbox %q: %w", id, err)
}

func decode(doc boxDoc) (*agentsv1.ButterBox, error) {
	box := &agentsv1.ButterBox{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), box); err != nil {
		return nil, fmt.Errorf("unmarshal butterbox %q: %w", doc.ID, err)
	}
	box.WorkspaceId = doc.WorkspaceID
	box.CredentialSet = doc.Credential != ""
	box.CredentialUpdatedAt = nil
	if doc.Credential != "" && !doc.CredentialUpdated.IsZero() {
		box.CredentialUpdatedAt = timestamppb.New(doc.CredentialUpdated)
	}
	return box, nil
}

// encodeSpec strips the derived fields before marshalling so the spec can
// never contradict the credential columns.
func encodeSpec(box *agentsv1.ButterBox) (string, error) {
	stored := proto.Clone(box).(*agentsv1.ButterBox)
	stored.WorkspaceId = ""
	stored.CredentialSet = false
	stored.CredentialUpdatedAt = nil
	spec, err := protojson.Marshal(stored)
	if err != nil {
		return "", fmt.Errorf("marshal butterbox %q: %w", box.GetId(), err)
	}
	return string(spec), nil
}

func (s *Store) List(ctx context.Context, workspaceID string) ([]*agentsv1.ButterBox, error) {
	cursor, err := s.boxes.Find(ctx, bson.M{"workspace_id": workspaceID},
		options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list butterboxes: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []boxDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode butterboxes: %w", err)
	}
	out := make([]*agentsv1.ButterBox, 0, len(docs))
	for _, doc := range docs {
		box, err := decode(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, box)
	}
	return out, nil
}

func (s *Store) Get(ctx context.Context, workspaceID, id string) (*agentsv1.ButterBox, error) {
	var doc boxDoc
	if err := s.boxes.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&doc); err != nil {
		return nil, mapError(id, err)
	}
	return decode(doc)
}

func (s *Store) Create(ctx context.Context, workspaceID string, box *agentsv1.ButterBox, cred butterboxrepo.Credential) (*agentsv1.ButterBox, error) {
	clone := proto.Clone(box).(*agentsv1.ButterBox)
	now := timestamppb.New(time.Now().UTC())
	clone.CreatedAt = now
	clone.UpdatedAt = now
	spec, err := encodeSpec(clone)
	if err != nil {
		return nil, err
	}
	doc := boxDoc{
		ID:          clone.GetId(),
		WorkspaceID: workspaceID,
		Name:        clone.GetName(),
		Spec:        spec,
	}
	if cred.Set() {
		doc.Credential = cred.Ciphertext
		doc.CredentialKeyID = cred.KeyID
		doc.CredentialUpdated = time.Now().UTC()
	}
	if _, err := s.boxes.InsertOne(ctx, doc); err != nil {
		return nil, mapError(clone.GetId(), err)
	}
	return decode(doc)
}

func (s *Store) Update(ctx context.Context, workspaceID string, box *agentsv1.ButterBox) (*agentsv1.ButterBox, error) {
	prev, err := s.Get(ctx, workspaceID, box.GetId())
	if err != nil {
		return nil, err
	}
	clone := proto.Clone(box).(*agentsv1.ButterBox)
	clone.CreatedAt = prev.GetCreatedAt()
	clone.UpdatedAt = timestamppb.New(time.Now().UTC())
	spec, err := encodeSpec(clone)
	if err != nil {
		return nil, err
	}
	res, err := s.boxes.UpdateOne(ctx,
		bson.M{"_id": clone.GetId(), "workspace_id": workspaceID},
		bson.M{"$set": bson.M{"name": clone.GetName(), "spec": spec}})
	if err != nil {
		return nil, mapError(clone.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("butterbox %q: %w", clone.GetId(), butterboxrepo.ErrNotFound)
	}
	return s.Get(ctx, workspaceID, clone.GetId())
}

func (s *Store) Delete(ctx context.Context, workspaceID, id string) error {
	res, err := s.boxes.DeleteOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID})
	if err != nil {
		return mapError(id, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNotFound)
	}
	return nil
}

func (s *Store) SetCredential(ctx context.Context, workspaceID, id string, cred butterboxrepo.Credential) (*agentsv1.ButterBox, error) {
	var update bson.M
	if cred.Set() {
		update = bson.M{"$set": bson.M{
			"credential":            cred.Ciphertext,
			"credential_key_id":     cred.KeyID,
			"credential_updated_at": time.Now().UTC(),
		}}
	} else {
		update = bson.M{"$unset": bson.M{
			"credential":            "",
			"credential_key_id":     "",
			"credential_updated_at": "",
		}}
	}
	res, err := s.boxes.UpdateOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, update)
	if err != nil {
		return nil, mapError(id, err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNotFound)
	}
	return s.Get(ctx, workspaceID, id)
}

func (s *Store) GetCredential(ctx context.Context, workspaceID, id string) (butterboxrepo.Credential, error) {
	var doc boxDoc
	if err := s.boxes.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&doc); err != nil {
		return butterboxrepo.Credential{}, mapError(id, err)
	}
	if doc.Credential == "" {
		return butterboxrepo.Credential{}, fmt.Errorf("butterbox %q: %w", id, butterboxrepo.ErrNoCredential)
	}
	return butterboxrepo.Credential{Ciphertext: doc.Credential, KeyID: doc.CredentialKeyID}, nil
}
