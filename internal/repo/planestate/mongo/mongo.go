// Package mongo implements planestate.Repository backed by MongoDB, with a
// TTL index that removes abandoned install flows.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/repo/planestate"
)

const statesCollection = "plane_install_states"

type stateDoc struct {
	Hash        string    `bson:"_id"`
	WorkspaceID string    `bson:"workspace_id"`
	AppID       string    `bson:"app_id"`
	ClientID    string    `bson:"client_id"`
	UserID      string    `bson:"user_id"`
	RedirectURI string    `bson:"redirect_uri"`
	ReturnURL   string    `bson:"return_url"`
	CreatedAt   time.Time `bson:"created_at"`
	ExpiresAt   time.Time `bson:"expires_at"`
}

// Store implements planestate.Repository backed by MongoDB.
type Store struct {
	states *mongo.Collection
}

var _ planestate.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store { return &Store{states: db.Collection(statesCollection)} }

func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.states.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0).SetName("ttl_expires_at"),
	})
	if err != nil {
		return fmt.Errorf("plane install state indexes: %w", err)
	}
	return nil
}

func (s *Store) Create(ctx context.Context, entry *planestate.Entry) error {
	_, err := s.states.InsertOne(ctx, stateDoc{
		Hash:        planestate.Hash(entry.State),
		WorkspaceID: entry.WorkspaceID,
		AppID:       entry.AppID,
		ClientID:    entry.ClientID,
		UserID:      entry.UserID,
		RedirectURI: entry.RedirectURI,
		ReturnURL:   entry.ReturnURL,
		CreatedAt:   entry.CreatedAt.UTC(),
		ExpiresAt:   entry.ExpiresAt.UTC(),
	})
	if err != nil {
		return fmt.Errorf("store plane install state: %w", err)
	}
	return nil
}

func (s *Store) Consume(ctx context.Context, state string, now time.Time) (*planestate.Entry, error) {
	var doc stateDoc
	err := s.states.FindOneAndDelete(ctx, bson.M{
		"_id":        planestate.Hash(state),
		"expires_at": bson.M{"$gt": now.UTC()},
	}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, planestate.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("consume plane install state: %w", err)
	}
	return &planestate.Entry{
		State:       state,
		WorkspaceID: doc.WorkspaceID,
		AppID:       doc.AppID,
		ClientID:    doc.ClientID,
		UserID:      doc.UserID,
		RedirectURI: doc.RedirectURI,
		ReturnURL:   doc.ReturnURL,
		CreatedAt:   doc.CreatedAt,
		ExpiresAt:   doc.ExpiresAt,
	}, nil
}
