package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/repo/oauthstate"
)

const collection = "oauth_states"

type stateDoc struct {
	State       string    `bson:"_id"`
	Provider    string    `bson:"provider"`
	RedirectURI string    `bson:"redirect_uri,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	ExpiresAt   time.Time `bson:"expires_at"`
}

// Store persists OAuth state in MongoDB with a TTL index on expires_at so
// stale entries self-destruct.
type Store struct {
	coll *mongo.Collection
}

var _ oauthstate.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collection)}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		return fmt.Errorf("create oauth_states expires_at ttl index: %w", err)
	}
	return nil
}

func (s *Store) Create(ctx context.Context, entry *oauthstate.Entry) error {
	if entry == nil || entry.State == "" {
		return errors.New("oauth state entry invalid")
	}
	doc := stateDoc{
		State:       entry.State,
		Provider:    entry.Provider,
		RedirectURI: entry.RedirectURI,
		CreatedAt:   entry.CreatedAt,
		ExpiresAt:   entry.ExpiresAt,
	}
	if _, err := s.coll.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert oauth state: %w", err)
	}
	return nil
}

func (s *Store) Consume(ctx context.Context, state string, now time.Time) (*oauthstate.Entry, error) {
	if state == "" {
		return nil, oauthstate.ErrNotFound
	}
	res := s.coll.FindOneAndDelete(ctx, bson.M{
		"_id":        state,
		"expires_at": bson.M{"$gt": now},
	})
	var doc stateDoc
	if err := res.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, oauthstate.ErrNotFound
		}
		return nil, fmt.Errorf("consume oauth state: %w", err)
	}
	return &oauthstate.Entry{
		State:       doc.State,
		Provider:    doc.Provider,
		RedirectURI: doc.RedirectURI,
		CreatedAt:   doc.CreatedAt,
		ExpiresAt:   doc.ExpiresAt,
	}, nil
}
