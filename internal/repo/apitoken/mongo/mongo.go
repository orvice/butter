package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/apitoken"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const collectionName = "api_tokens"

// tokenDoc is the MongoDB row for an API token.
type tokenDoc struct {
	ID         string    `bson:"_id"`
	Name       string    `bson:"name"`
	Prefix     string    `bson:"prefix"`
	SecretHash string    `bson:"secret_hash"`
	CreatedAt  time.Time `bson:"created_at"`
	LastUsedAt time.Time `bson:"last_used_at,omitempty"`
	Revoked    bool      `bson:"revoked,omitempty"`
}

// Store is a MongoDB-backed implementation of apitoken.Repository.
type Store struct {
	coll *mongo.Collection
}

func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collectionName)}
}

func (s *Store) List(ctx context.Context) ([]*agentsv1.APIToken, error) {
	cursor, err := s.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list api_tokens: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.APIToken
	for cursor.Next(ctx) {
		var doc tokenDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode api_token: %w", err)
		}
		out = append(out, docToProto(&doc))
	}
	return out, nil
}

func (s *Store) Get(ctx context.Context, id string) (*agentsv1.APIToken, error) {
	var doc tokenDoc
	err := s.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apitoken.ErrNotFound
		}
		return nil, fmt.Errorf("get api_token: %w", err)
	}
	return docToProto(&doc), nil
}

func (s *Store) Create(ctx context.Context, token *agentsv1.APIToken, secretHash string) error {
	doc := tokenDoc{
		ID:         token.GetId(),
		Name:       token.GetName(),
		Prefix:     token.GetPrefix(),
		SecretHash: secretHash,
		CreatedAt:  token.GetCreatedAt().AsTime(),
	}
	if _, err := s.coll.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert api_token: %w", err)
	}
	return nil
}

func (s *Store) Revoke(ctx context.Context, id string) (*agentsv1.APIToken, error) {
	res := s.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"revoked": true}})
	var doc tokenDoc
	if err := res.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apitoken.ErrNotFound
		}
		return nil, fmt.Errorf("revoke api_token: %w", err)
	}
	doc.Revoked = true
	return docToProto(&doc), nil
}

func (s *Store) Lookup(ctx context.Context, secretHash string) (*agentsv1.APIToken, error) {
	var doc tokenDoc
	err := s.coll.FindOne(ctx, bson.M{"secret_hash": secretHash, "revoked": bson.M{"$ne": true}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apitoken.ErrNotFound
		}
		return nil, fmt.Errorf("lookup api_token: %w", err)
	}
	return docToProto(&doc), nil
}

func (s *Store) TouchLastUsed(ctx context.Context, id string) error {
	_, err := s.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"last_used_at": time.Now().UTC()}})
	return err
}

func docToProto(doc *tokenDoc) *agentsv1.APIToken {
	t := &agentsv1.APIToken{
		Id:        doc.ID,
		Name:      doc.Name,
		Prefix:    doc.Prefix,
		CreatedAt: timestamppb.New(doc.CreatedAt),
		Revoked:   doc.Revoked,
	}
	if !doc.LastUsedAt.IsZero() {
		t.LastUsedAt = timestamppb.New(doc.LastUsedAt)
	}
	return t
}
