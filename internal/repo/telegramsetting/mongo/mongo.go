// Package mongo implements telegramsetting.Repository backed by MongoDB.
// The settings live in one document with a fixed _id so every Pod reads the
// same row without any coordination.
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

	"go.orx.me/apps/butter/internal/repo/telegramsetting"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	settingsCollection = "platform_settings"
	telegramSettingsID = "telegram"
)

type settingsDoc struct {
	ID   string `bson:"_id"`
	Spec string `bson:"spec"`
}

// Store implements telegramsetting.Repository backed by MongoDB.
type Store struct {
	settings *mongo.Collection
}

var _ telegramsetting.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{settings: db.Collection(settingsCollection)}
}

// EnsureIndexes is a no-op: the document is addressed by _id only.
func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) Get(ctx context.Context) (*agentsv1.TelegramSettings, error) {
	var doc settingsDoc
	err := s.settings.FindOne(ctx, bson.M{"_id": telegramSettingsID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// Never configured is a normal state, not an error.
		return &agentsv1.TelegramSettings{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read telegram settings: %w", err)
	}
	settings := &agentsv1.TelegramSettings{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), settings); err != nil {
		return nil, fmt.Errorf("unmarshal telegram settings: %w", err)
	}
	return settings, nil
}

func (s *Store) Put(ctx context.Context, settings *agentsv1.TelegramSettings) (*agentsv1.TelegramSettings, error) {
	stored := proto.Clone(settings).(*agentsv1.TelegramSettings)
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	spec, err := protojson.Marshal(stored)
	if err != nil {
		return nil, fmt.Errorf("marshal telegram settings: %w", err)
	}
	if _, err := s.settings.UpdateOne(ctx,
		bson.M{"_id": telegramSettingsID},
		bson.M{"$set": bson.M{"spec": string(spec)}},
		options.UpdateOne().SetUpsert(true),
	); err != nil {
		return nil, fmt.Errorf("write telegram settings: %w", err)
	}
	return stored, nil
}
