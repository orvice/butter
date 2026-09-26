// Package mongo implements memoryconfig.Repository backed by MongoDB.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/types/known/timestamppb"

	memoryconfigrepo "go.orx.me/apps/butter/internal/repo/memoryconfig"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const configsCollection = "workspace_memory_configs"

// configDoc is keyed by workspace ID, which is what makes a config a
// per-workspace singleton. The credential lives in dedicated columns so the
// derived credential fields cannot contradict storage.
type configDoc struct {
	WorkspaceID       string    `bson:"_id"`
	BaseURL           string    `bson:"base_url"`
	Enabled           bool      `bson:"enabled"`
	CreatedAt         time.Time `bson:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at"`
	Credential        string    `bson:"credential,omitempty"`
	CredentialKeyID   string    `bson:"credential_key_id,omitempty"`
	CredentialUpdated time.Time `bson:"credential_updated_at,omitempty"`
}

// Store implements memoryconfig.Repository backed by MongoDB.
type Store struct {
	configs *mongo.Collection
}

var _ memoryconfigrepo.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{configs: db.Collection(configsCollection)}
}

// EnsureIndexes is a no-op: the only key is _id.
func (s *Store) EnsureIndexes(context.Context) error { return nil }

func mapError(workspaceID string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("workspace memory config %q: %w", workspaceID, memoryconfigrepo.ErrNotFound)
	}
	return fmt.Errorf("workspace memory config %q: %w", workspaceID, err)
}

func decode(doc configDoc) *agentsv1.WorkspaceMemoryConfig {
	cfg := &agentsv1.WorkspaceMemoryConfig{
		WorkspaceId:   doc.WorkspaceID,
		BaseUrl:       doc.BaseURL,
		Enabled:       doc.Enabled,
		CredentialSet: doc.Credential != "",
		CreatedAt:     timestamppb.New(doc.CreatedAt),
		UpdatedAt:     timestamppb.New(doc.UpdatedAt),
	}
	if doc.Credential != "" && !doc.CredentialUpdated.IsZero() {
		cfg.CredentialUpdatedAt = timestamppb.New(doc.CredentialUpdated)
	}
	return cfg
}

func (s *Store) Get(ctx context.Context, workspaceID string) (*agentsv1.WorkspaceMemoryConfig, error) {
	var doc configDoc
	if err := s.configs.FindOne(ctx, bson.M{"_id": workspaceID}).Decode(&doc); err != nil {
		return nil, mapError(workspaceID, err)
	}
	return decode(doc), nil
}

func (s *Store) Put(ctx context.Context, workspaceID string, cfg *agentsv1.WorkspaceMemoryConfig, cred *memoryconfigrepo.Credential) (*agentsv1.WorkspaceMemoryConfig, error) {
	now := time.Now().UTC()
	set := bson.M{
		"base_url":   cfg.GetBaseUrl(),
		"enabled":    cfg.GetEnabled(),
		"updated_at": now,
	}
	update := bson.M{"$setOnInsert": bson.M{"created_at": now}}
	switch {
	case cred == nil:
	case cred.Set():
		set["credential"] = cred.Ciphertext
		set["credential_key_id"] = cred.KeyID
		set["credential_updated_at"] = now
	default:
		update["$unset"] = bson.M{
			"credential":            "",
			"credential_key_id":     "",
			"credential_updated_at": "",
		}
	}
	update["$set"] = set

	var doc configDoc
	err := s.configs.FindOneAndUpdate(ctx, bson.M{"_id": workspaceID}, update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&doc)
	if err != nil {
		return nil, mapError(workspaceID, err)
	}
	return decode(doc), nil
}

func (s *Store) Delete(ctx context.Context, workspaceID string) error {
	res, err := s.configs.DeleteOne(ctx, bson.M{"_id": workspaceID})
	if err != nil {
		return mapError(workspaceID, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("workspace memory config %q: %w", workspaceID, memoryconfigrepo.ErrNotFound)
	}
	return nil
}

func (s *Store) GetCredential(ctx context.Context, workspaceID string) (memoryconfigrepo.Credential, error) {
	var doc configDoc
	if err := s.configs.FindOne(ctx, bson.M{"_id": workspaceID}).Decode(&doc); err != nil {
		return memoryconfigrepo.Credential{}, mapError(workspaceID, err)
	}
	if doc.Credential == "" {
		return memoryconfigrepo.Credential{}, fmt.Errorf("workspace memory config %q: %w", workspaceID, memoryconfigrepo.ErrNoCredential)
	}
	return memoryconfigrepo.Credential{Ciphertext: doc.Credential, KeyID: doc.CredentialKeyID}, nil
}
