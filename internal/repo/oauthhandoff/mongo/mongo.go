package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/repo/oauthhandoff"
)

const collection = "oauth_handoffs"

type handoffDoc struct {
	CodeHash      string    `bson:"_id"`
	Provider      string    `bson:"provider"`
	State         string    `bson:"state"`
	UserID        string    `bson:"user_id"`
	CodeChallenge string    `bson:"code_challenge"`
	CreatedAt     time.Time `bson:"created_at"`
	ExpiresAt     time.Time `bson:"expires_at"`
}

// Store persists handoffs in MongoDB with a TTL index on expires_at so an
// unredeemed handoff disappears on its own.
type Store struct {
	coll *mongo.Collection
}

var _ oauthhandoff.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collection)}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		return fmt.Errorf("create oauth_handoffs expires_at ttl index: %w", err)
	}
	return nil
}

func (s *Store) Create(ctx context.Context, entry *oauthhandoff.Entry) error {
	if entry == nil || entry.CodeHash == "" {
		return errors.New("oauth handoff entry invalid")
	}
	doc := handoffDoc{
		CodeHash:      entry.CodeHash,
		Provider:      entry.Provider,
		State:         entry.State,
		UserID:        entry.UserID,
		CodeChallenge: entry.CodeChallenge,
		CreatedAt:     entry.CreatedAt,
		ExpiresAt:     entry.ExpiresAt,
	}
	if _, err := s.coll.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert oauth handoff: %w", err)
	}
	return nil
}

func (s *Store) Consume(ctx context.Context, codeHash string, now time.Time) (*oauthhandoff.Entry, error) {
	if codeHash == "" {
		return nil, oauthhandoff.ErrNotFound
	}
	// Delete whether or not it is still valid: an expired handoff is spent.
	res := s.coll.FindOneAndDelete(ctx, bson.M{"_id": codeHash})
	var doc handoffDoc
	if err := res.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, oauthhandoff.ErrNotFound
		}
		return nil, fmt.Errorf("consume oauth handoff: %w", err)
	}
	if !doc.ExpiresAt.After(now) {
		return nil, oauthhandoff.ErrNotFound
	}
	return &oauthhandoff.Entry{
		CodeHash:      doc.CodeHash,
		Provider:      doc.Provider,
		State:         doc.State,
		UserID:        doc.UserID,
		CodeChallenge: doc.CodeChallenge,
		CreatedAt:     doc.CreatedAt,
		ExpiresAt:     doc.ExpiresAt,
	}, nil
}
