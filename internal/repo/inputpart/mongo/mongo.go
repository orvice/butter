package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/inputpart"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const collectionName = "invocation_input_parts"

// doc is the BSON layout for one input part record. Each part is stored
// separately so that a 20 MiB combined payload never hits MongoDB's 16 MiB
// per-document limit.
type doc struct {
	ID           string `bson:"_id"`
	InvocationID string `bson:"invocation_id"`
	Index        int    `bson:"index"`
	Data         []byte `bson:"data"` // proto-marshalled InputPart
}

// Store is a MongoDB-backed implementation of inputpart.Repository.
type Store struct {
	coll *mongo.Collection
}

func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collectionName)}
}

// EnsureIndexes creates the indexes needed for efficient lookup and cleanup.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "invocation_id", Value: 1}, {Key: "index", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return fmt.Errorf("create invocation_input_parts indexes: %w", err)
	}
	return nil
}

func (s *Store) SaveAll(ctx context.Context, invocationID string, parts []*agentsv1.InputPart) error {
	// Idempotency: if records already exist for this invocation, skip.
	count, err := s.coll.CountDocuments(ctx, bson.M{"invocation_id": invocationID})
	if err != nil {
		return fmt.Errorf("check existing input parts: %w", err)
	}
	if count > 0 {
		return nil
	}

	docs := make([]interface{}, len(parts))
	for i, p := range parts {
		data, err := proto.Marshal(p)
		if err != nil {
			return fmt.Errorf("marshal input part %d: %w", i, err)
		}
		docs[i] = doc{
			ID:           fmt.Sprintf("%s:%d", invocationID, i),
			InvocationID: invocationID,
			Index:        i,
			Data:         data,
		}
	}

	_, err = s.coll.InsertMany(ctx, docs)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil // idempotent
		}
		return fmt.Errorf("insert input parts: %w", err)
	}
	return nil
}

func (s *Store) Load(ctx context.Context, invocationID string) ([]*agentsv1.InputPart, error) {
	opts := options.Find().SetSort(bson.D{{Key: "index", Value: 1}})
	cursor, err := s.coll.Find(ctx, bson.M{"invocation_id": invocationID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find input parts: %w", err)
	}
	defer cursor.Close(ctx)

	var parts []*agentsv1.InputPart
	for cursor.Next(ctx) {
		var d doc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode input part: %w", err)
		}
		p := &agentsv1.InputPart{}
		if err := proto.Unmarshal(d.Data, p); err != nil {
			return nil, fmt.Errorf("unmarshal input part %d: %w", d.Index, err)
		}
		parts = append(parts, p)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate input parts: %w", err)
	}
	if len(parts) == 0 {
		return nil, inputpart.ErrNotFound
	}
	return parts, nil
}

func (s *Store) Delete(ctx context.Context, invocationID string) error {
	_, err := s.coll.DeleteMany(ctx, bson.M{"invocation_id": invocationID})
	if err != nil {
		return fmt.Errorf("delete input parts: %w", err)
	}
	return nil
}
