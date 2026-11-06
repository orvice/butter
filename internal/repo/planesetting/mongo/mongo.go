// Package mongo implements planesetting.Repository backed by MongoDB. The
// settings share the platform settings collection with Telegram's, under
// their own fixed _id.
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

	"go.orx.me/apps/butter/internal/repo/planesetting"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	settingsCollection = "platform_settings"
	planeSettingsID    = "plane"
)

type settingsDoc struct {
	ID   string `bson:"_id"`
	Spec string `bson:"spec"`
}

// Store implements planesetting.Repository backed by MongoDB.
type Store struct {
	settings *mongo.Collection
}

var _ planesetting.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{settings: db.Collection(settingsCollection)}
}

// EnsureIndexes is a no-op: the document is addressed by _id only.
func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) Get(ctx context.Context) (*agentsv1.PlaneSettings, error) {
	var doc settingsDoc
	err := s.settings.FindOne(ctx, bson.M{"_id": planeSettingsID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return &agentsv1.PlaneSettings{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read plane settings: %w", err)
	}
	settings := &agentsv1.PlaneSettings{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), settings); err != nil {
		return nil, fmt.Errorf("unmarshal plane settings: %w", err)
	}
	return settings, nil
}

func (s *Store) Put(ctx context.Context, settings *agentsv1.PlaneSettings) (*agentsv1.PlaneSettings, error) {
	stored := proto.Clone(settings).(*agentsv1.PlaneSettings)
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	spec, err := protojson.Marshal(stored)
	if err != nil {
		return nil, fmt.Errorf("marshal plane settings: %w", err)
	}
	if _, err := s.settings.UpdateOne(ctx,
		bson.M{"_id": planeSettingsID},
		bson.M{"$set": bson.M{"spec": string(spec)}},
		options.UpdateOne().SetUpsert(true),
	); err != nil {
		return nil, fmt.Errorf("write plane settings: %w", err)
	}
	return stored, nil
}
