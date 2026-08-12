// Package mongo implements cryptokey.Repository backed by MongoDB.
//
// Two document shapes share one collection:
//   - the pointer document, _id "active", naming the current key; and
//   - one document per key, _id = key ID, holding the material.
//
// EnsureActive inserts the key document first (harmless if it ends up
// orphaned) and then races to claim the pointer. A duplicate-key error on the
// pointer means another Pod won, so we read the pointer back and use its key.
// The pointer is never updated in place, which is what makes concurrent
// first-use safe.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go.orx.me/apps/butter/internal/repo/cryptokey"
)

const (
	keysCollection = "crypto_master_keys"
	activeID       = "active"
)

type keyDoc struct {
	ID        string    `bson:"_id"`
	KeyID     string    `bson:"key_id"`
	Material  []byte    `bson:"material"`
	CreatedAt time.Time `bson:"created_at"`
}

// Store implements cryptokey.Repository backed by MongoDB.
type Store struct {
	keys *mongo.Collection
}

var _ cryptokey.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{keys: db.Collection(keysCollection)}
}

// EnsureIndexes is a no-op: documents are addressed by _id only.
func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) EnsureActive(ctx context.Context, generate func() (cryptokey.MasterKey, error)) (cryptokey.MasterKey, error) {
	if active, err := s.readActive(ctx); err == nil {
		return active, nil
	} else if !errors.Is(err, cryptokey.ErrNotFound) {
		return cryptokey.MasterKey{}, err
	}

	candidate, err := generate()
	if err != nil {
		return cryptokey.MasterKey{}, fmt.Errorf("generate master key: %w", err)
	}
	now := time.Now().UTC()

	// Store the material under its own ID first. If we lose the pointer race
	// this document is simply never referenced.
	if _, err := s.keys.InsertOne(ctx, keyDoc{
		ID:        candidate.ID,
		KeyID:     candidate.ID,
		Material:  candidate.Material,
		CreatedAt: now,
	}); err != nil && !mongo.IsDuplicateKeyError(err) {
		return cryptokey.MasterKey{}, fmt.Errorf("store master key: %w", err)
	}

	_, err = s.keys.InsertOne(ctx, keyDoc{
		ID:        activeID,
		KeyID:     candidate.ID,
		Material:  candidate.Material,
		CreatedAt: now,
	})
	switch {
	case err == nil:
		return candidate, nil
	case mongo.IsDuplicateKeyError(err):
		// Another Pod claimed the pointer between our read and our insert.
		return s.readActive(ctx)
	default:
		return cryptokey.MasterKey{}, fmt.Errorf("claim active master key: %w", err)
	}
}

func (s *Store) Get(ctx context.Context, id string) (cryptokey.MasterKey, error) {
	if id == "" {
		return cryptokey.MasterKey{}, fmt.Errorf("master key %q: %w", id, cryptokey.ErrNotFound)
	}
	return s.read(ctx, id)
}

func (s *Store) readActive(ctx context.Context) (cryptokey.MasterKey, error) {
	key, err := s.read(ctx, activeID)
	if err != nil {
		return cryptokey.MasterKey{}, err
	}
	return key, nil
}

func (s *Store) read(ctx context.Context, id string) (cryptokey.MasterKey, error) {
	var doc keyDoc
	err := s.keys.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return cryptokey.MasterKey{}, fmt.Errorf("master key %q: %w", id, cryptokey.ErrNotFound)
	}
	if err != nil {
		return cryptokey.MasterKey{}, fmt.Errorf("read master key %q: %w", id, err)
	}
	keyID := doc.KeyID
	if keyID == "" {
		keyID = doc.ID
	}
	return cryptokey.MasterKey{ID: keyID, Material: doc.Material}, nil
}
