// Package mongo implements telegram.Repository backed by MongoDB.
//
// Uniqueness is enforced by indexes rather than read-then-write checks,
// because the invariants this repository owns are exactly the ones that
// concurrent callers break: two Pods creating a Channel for the same Bot, or
// two operators registering the same Forum Topic. A read-then-write check
// would pass in both callers and leave one inbound update matching two
// Destinations.
//
// Credential material lives in dedicated document fields, never inside the
// protojson spec, so decoding a Channel can never surface a Bot Token.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	channelsCollection     = "telegram_channels"
	destinationsCollection = "telegram_destinations"
)

type channelDoc struct {
	ID                 string    `bson:"_id"`
	WorkspaceID        string    `bson:"workspace_id"`
	Key                string    `bson:"key"`
	BotID              string    `bson:"bot_id"`
	Revision           int64     `bson:"revision"`
	Spec               string    `bson:"spec"`
	Credential         string    `bson:"credential,omitempty"`
	CredentialKeyID    string    `bson:"credential_key_id,omitempty"`
	CredentialUpdated  time.Time `bson:"credential_updated_at,omitempty"`
	WebhookSecret      string    `bson:"webhook_secret,omitempty"`
	WebhookSecretKeyID string    `bson:"webhook_secret_key_id,omitempty"`
}

type destinationDoc struct {
	ID              string `bson:"_id"`
	WorkspaceID     string `bson:"workspace_id"`
	Key             string `bson:"key"`
	ChannelID       string `bson:"channel_id"`
	ChatID          string `bson:"chat_id"`
	MessageThreadID string `bson:"message_thread_id"`
	Revision        int64  `bson:"revision"`
	Spec            string `bson:"spec"`
}

// Store implements telegram.Repository backed by MongoDB.
type Store struct {
	channels     *mongo.Collection
	destinations *mongo.Collection
}

var _ telegramrepo.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{
		channels:     db.Collection(channelsCollection),
		destinations: db.Collection(destinationsCollection),
	}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	if _, err := s.channels.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "key", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_workspace_key"),
		},
		{
			// Global, not workspace-scoped: one Telegram Bot must never be
			// consumed by two Channels.
			Keys:    bson.D{{Key: "bot_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_bot_id"),
		},
	}); err != nil {
		return fmt.Errorf("create telegram channel indexes: %w", err)
	}
	if _, err := s.destinations.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "key", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_workspace_key"),
		},
		{
			Keys: bson.D{
				{Key: "channel_id", Value: 1},
				{Key: "chat_id", Value: 1},
				{Key: "message_thread_id", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("uniq_address"),
		},
	}); err != nil {
		return fmt.Errorf("create telegram destination indexes: %w", err)
	}
	return nil
}

// --- Channels --------------------------------------------------------------

func decodeChannel(doc channelDoc) (*agentsv1.TelegramChannel, error) {
	ch := &agentsv1.TelegramChannel{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), ch); err != nil {
		return nil, fmt.Errorf("unmarshal telegram channel %q: %w", doc.ID, err)
	}
	ch.Id = doc.ID
	ch.WorkspaceId = doc.WorkspaceID
	ch.Key = doc.Key
	ch.BotId = doc.BotID
	ch.Revision = doc.Revision
	if doc.Credential == "" {
		ch.CredentialState = agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_MISSING
		ch.CredentialUpdatedAt = nil
	} else if !doc.CredentialUpdated.IsZero() {
		ch.CredentialUpdatedAt = timestamppb.New(doc.CredentialUpdated)
	}
	ch.WebhookSecretSet = doc.WebhookSecret != ""
	return ch, nil
}

// encodeChannelSpec strips the fields carried in dedicated document columns
// so the spec never duplicates — or contradicts — the indexed values.
func encodeChannelSpec(ch *agentsv1.TelegramChannel) (string, error) {
	stored := proto.Clone(ch).(*agentsv1.TelegramChannel)
	stored.CredentialUpdatedAt = nil
	stored.WebhookSecretSet = false
	spec, err := protojson.Marshal(stored)
	if err != nil {
		return "", fmt.Errorf("marshal telegram channel %q: %w", ch.GetId(), err)
	}
	return string(spec), nil
}

func (s *Store) ListChannels(ctx context.Context, workspaceID string) ([]*agentsv1.TelegramChannel, error) {
	return s.findChannels(ctx, bson.M{"workspace_id": workspaceID})
}

func (s *Store) ListChannelsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.TelegramChannel, error) {
	return s.findChannels(ctx, bson.M{})
}

func (s *Store) findChannels(ctx context.Context, filter bson.M) ([]*agentsv1.TelegramChannel, error) {
	cursor, err := s.channels.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list telegram channels: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []channelDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode telegram channels: %w", err)
	}
	out := make([]*agentsv1.TelegramChannel, 0, len(docs))
	for _, doc := range docs {
		ch, err := decodeChannel(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetKey() < out[j].GetKey() })
	return out, nil
}

func (s *Store) GetChannel(ctx context.Context, workspaceID, id string) (*agentsv1.TelegramChannel, error) {
	doc, err := s.channelDoc(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, id)
	if err != nil {
		return nil, err
	}
	return decodeChannel(doc)
}

func (s *Store) FindChannel(ctx context.Context, id string) (*agentsv1.TelegramChannel, error) {
	doc, err := s.channelDoc(ctx, bson.M{"_id": id}, id)
	if err != nil {
		return nil, err
	}
	return decodeChannel(doc)
}

func (s *Store) channelDoc(ctx context.Context, filter bson.M, id string) (channelDoc, error) {
	var doc channelDoc
	err := s.channels.FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return channelDoc{}, channelNotFound(id)
	}
	if err != nil {
		return channelDoc{}, fmt.Errorf("read telegram channel %q: %w", id, err)
	}
	return doc, nil
}

func (s *Store) CreateChannel(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, cred telegramrepo.Credential) (*agentsv1.TelegramChannel, error) {
	stored := proto.Clone(channel).(*agentsv1.TelegramChannel)
	stored.WorkspaceId = workspaceID
	now := time.Now().UTC()
	stored.CreatedAt = timestamppb.New(now)
	stored.UpdatedAt = timestamppb.New(now)
	stored.Revision = 1

	spec, err := encodeChannelSpec(stored)
	if err != nil {
		return nil, err
	}
	doc := channelDoc{
		ID:          stored.GetId(),
		WorkspaceID: workspaceID,
		Key:         stored.GetKey(),
		BotID:       stored.GetBotId(),
		Revision:    1,
		Spec:        spec,
		Credential:  cred.Ciphertext,
	}
	if cred.Set() {
		doc.CredentialKeyID = cred.KeyID
		doc.CredentialUpdated = now
	}
	if _, err := s.channels.InsertOne(ctx, doc); err != nil {
		return nil, mapChannelWriteErr(stored, err)
	}
	return decodeChannel(doc)
}

func (s *Store) UpdateChannel(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, expectedRevision int64) (*agentsv1.TelegramChannel, error) {
	return s.updateChannel(ctx, workspaceID, channel, nil, expectedRevision)
}

func (s *Store) RotateChannelCredential(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, cred telegramrepo.Credential, expectedRevision int64) (*agentsv1.TelegramChannel, error) {
	return s.updateChannel(ctx, workspaceID, channel, &cred, expectedRevision)
}

func (s *Store) updateChannel(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, cred *telegramrepo.Credential, expectedRevision int64) (*agentsv1.TelegramChannel, error) {
	current, err := s.channelDoc(ctx, bson.M{"_id": channel.GetId(), "workspace_id": workspaceID}, channel.GetId())
	if err != nil {
		return nil, err
	}

	stored := proto.Clone(channel).(*agentsv1.TelegramChannel)
	stored.WorkspaceId = workspaceID
	// Immutable and repo-owned fields always come from the stored document.
	stored.Key = current.Key
	stored.BotId = current.BotID
	existing, err := decodeChannel(current)
	if err != nil {
		return nil, err
	}
	stored.CreatedAt = existing.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	stored.Revision = expectedRevision + 1

	spec, err := encodeChannelSpec(stored)
	if err != nil {
		return nil, err
	}
	set := bson.M{"spec": spec, "revision": stored.GetRevision()}
	if cred != nil {
		set["credential"] = cred.Ciphertext
		set["credential_key_id"] = cred.KeyID
		if cred.Set() {
			set["credential_updated_at"] = time.Now().UTC()
		} else {
			set["credential_updated_at"] = time.Time{}
		}
	}
	res, err := s.channels.UpdateOne(ctx,
		bson.M{"_id": channel.GetId(), "workspace_id": workspaceID, "revision": expectedRevision},
		bson.M{"$set": set},
	)
	if err != nil {
		return nil, fmt.Errorf("update telegram channel %q: %w", channel.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("telegram channel %q (stored revision %d, expected %d): %w",
			channel.GetId(), current.Revision, expectedRevision, telegramrepo.ErrRevisionConflict)
	}

	current.Spec = spec
	current.Revision = stored.GetRevision()
	if cred != nil {
		current.Credential = cred.Ciphertext
		current.CredentialKeyID = cred.KeyID
		if cred.Set() {
			current.CredentialUpdated = set["credential_updated_at"].(time.Time)
		} else {
			current.CredentialUpdated = time.Time{}
		}
	}
	return decodeChannel(current)
}

func (s *Store) DeleteChannel(ctx context.Context, workspaceID, id string) error {
	if _, err := s.channelDoc(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, id); err != nil {
		return err
	}
	var ref destinationDoc
	err := s.destinations.FindOne(ctx, bson.M{"channel_id": id}).Decode(&ref)
	switch {
	case err == nil:
		return fmt.Errorf("telegram channel %q is referenced by destination %q: %w",
			id, ref.ID, telegramrepo.ErrChannelInUse)
	case errors.Is(err, mongo.ErrNoDocuments):
	default:
		return fmt.Errorf("check telegram channel %q references: %w", id, err)
	}
	if _, err := s.channels.DeleteOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}); err != nil {
		return fmt.Errorf("delete telegram channel %q: %w", id, err)
	}
	return nil
}

func (s *Store) SetChannelCredential(ctx context.Context, workspaceID, id string, cred telegramrepo.Credential) error {
	return s.setSecret(ctx, workspaceID, id, "credential", "credential_key_id", "credential_updated_at", cred)
}

func (s *Store) GetChannelCredential(ctx context.Context, workspaceID, id string) (telegramrepo.Credential, error) {
	doc, err := s.channelDoc(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, id)
	if err != nil {
		return telegramrepo.Credential{}, err
	}
	if doc.Credential == "" {
		return telegramrepo.Credential{}, fmt.Errorf("telegram channel %q: %w", id, telegramrepo.ErrNoCredential)
	}
	return telegramrepo.Credential{Ciphertext: doc.Credential, KeyID: doc.CredentialKeyID}, nil
}

func (s *Store) SetWebhookSecret(ctx context.Context, workspaceID, id string, cred telegramrepo.Credential) error {
	return s.setSecret(ctx, workspaceID, id, "webhook_secret", "webhook_secret_key_id", "", cred)
}

func (s *Store) GetWebhookSecret(ctx context.Context, workspaceID, id string) (telegramrepo.Credential, error) {
	doc, err := s.channelDoc(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, id)
	if err != nil {
		return telegramrepo.Credential{}, err
	}
	if doc.WebhookSecret == "" {
		return telegramrepo.Credential{}, fmt.Errorf("telegram channel %q: %w", id, telegramrepo.ErrNoCredential)
	}
	return telegramrepo.Credential{Ciphertext: doc.WebhookSecret, KeyID: doc.WebhookSecretKeyID}, nil
}

func (s *Store) setSecret(ctx context.Context, workspaceID, id, valueField, keyField, updatedField string, cred telegramrepo.Credential) error {
	update := bson.M{}
	if cred.Set() {
		set := bson.M{valueField: cred.Ciphertext, keyField: cred.KeyID}
		if updatedField != "" {
			set[updatedField] = time.Now().UTC()
		}
		update["$set"] = set
	} else {
		unset := bson.M{valueField: "", keyField: ""}
		if updatedField != "" {
			unset[updatedField] = ""
		}
		update["$unset"] = unset
	}
	res, err := s.channels.UpdateOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, update)
	if err != nil {
		return fmt.Errorf("set telegram channel %q secret: %w", id, err)
	}
	if res.MatchedCount == 0 {
		return channelNotFound(id)
	}
	return nil
}

// --- Destinations ----------------------------------------------------------

func decodeDestination(doc destinationDoc) (*agentsv1.TelegramDestination, error) {
	d := &agentsv1.TelegramDestination{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), d); err != nil {
		return nil, fmt.Errorf("unmarshal telegram destination %q: %w", doc.ID, err)
	}
	d.Id = doc.ID
	d.WorkspaceId = doc.WorkspaceID
	d.Key = doc.Key
	d.ChannelId = doc.ChannelID
	d.ChatId = doc.ChatID
	d.MessageThreadId = doc.MessageThreadID
	d.Revision = doc.Revision
	return d, nil
}

func encodeDestinationDoc(workspaceID string, d *agentsv1.TelegramDestination) (destinationDoc, error) {
	spec, err := protojson.Marshal(d)
	if err != nil {
		return destinationDoc{}, fmt.Errorf("marshal telegram destination %q: %w", d.GetId(), err)
	}
	return destinationDoc{
		ID:              d.GetId(),
		WorkspaceID:     workspaceID,
		Key:             d.GetKey(),
		ChannelID:       d.GetChannelId(),
		ChatID:          d.GetChatId(),
		MessageThreadID: d.GetMessageThreadId(),
		Revision:        d.GetRevision(),
		Spec:            string(spec),
	}, nil
}

func (s *Store) ListDestinations(ctx context.Context, workspaceID, channelID string) ([]*agentsv1.TelegramDestination, error) {
	filter := bson.M{"workspace_id": workspaceID}
	if channelID != "" {
		filter["channel_id"] = channelID
	}
	return s.findDestinations(ctx, filter)
}

func (s *Store) ListDestinationsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.TelegramDestination, error) {
	return s.findDestinations(ctx, bson.M{})
}

func (s *Store) findDestinations(ctx context.Context, filter bson.M) ([]*agentsv1.TelegramDestination, error) {
	cursor, err := s.destinations.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list telegram destinations: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []destinationDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode telegram destinations: %w", err)
	}
	out := make([]*agentsv1.TelegramDestination, 0, len(docs))
	for _, doc := range docs {
		d, err := decodeDestination(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GetKey() < out[j].GetKey() })
	return out, nil
}

func (s *Store) GetDestination(ctx context.Context, workspaceID, id string) (*agentsv1.TelegramDestination, error) {
	doc, err := s.destinationDoc(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, id)
	if err != nil {
		return nil, err
	}
	return decodeDestination(doc)
}

func (s *Store) destinationDoc(ctx context.Context, filter bson.M, id string) (destinationDoc, error) {
	var doc destinationDoc
	err := s.destinations.FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return destinationDoc{}, destinationNotFound(id)
	}
	if err != nil {
		return destinationDoc{}, fmt.Errorf("read telegram destination %q: %w", id, err)
	}
	return doc, nil
}

func (s *Store) CreateDestination(ctx context.Context, workspaceID string, dest *agentsv1.TelegramDestination) (*agentsv1.TelegramDestination, error) {
	stored := proto.Clone(dest).(*agentsv1.TelegramDestination)
	stored.WorkspaceId = workspaceID
	now := timestamppb.New(time.Now().UTC())
	stored.CreatedAt = now
	stored.UpdatedAt = now
	stored.Revision = 1

	doc, err := encodeDestinationDoc(workspaceID, stored)
	if err != nil {
		return nil, err
	}
	if _, err := s.destinations.InsertOne(ctx, doc); err != nil {
		return nil, mapDestinationWriteErr(stored, err)
	}
	return decodeDestination(doc)
}

func (s *Store) UpdateDestination(ctx context.Context, workspaceID string, dest *agentsv1.TelegramDestination, expectedRevision int64) (*agentsv1.TelegramDestination, error) {
	current, err := s.destinationDoc(ctx, bson.M{"_id": dest.GetId(), "workspace_id": workspaceID}, dest.GetId())
	if err != nil {
		return nil, err
	}
	existing, err := decodeDestination(current)
	if err != nil {
		return nil, err
	}

	stored := proto.Clone(dest).(*agentsv1.TelegramDestination)
	stored.WorkspaceId = workspaceID
	stored.Key = current.Key
	stored.ChannelId = current.ChannelID
	stored.ChatId = current.ChatID
	stored.MessageThreadId = current.MessageThreadID
	stored.CreatedAt = existing.GetCreatedAt()
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	stored.Revision = expectedRevision + 1

	doc, err := encodeDestinationDoc(workspaceID, stored)
	if err != nil {
		return nil, err
	}
	res, err := s.destinations.UpdateOne(ctx,
		bson.M{"_id": dest.GetId(), "workspace_id": workspaceID, "revision": expectedRevision},
		bson.M{"$set": bson.M{"spec": doc.Spec, "revision": doc.Revision}},
	)
	if err != nil {
		return nil, fmt.Errorf("update telegram destination %q: %w", dest.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("telegram destination %q (stored revision %d, expected %d): %w",
			dest.GetId(), current.Revision, expectedRevision, telegramrepo.ErrRevisionConflict)
	}
	return decodeDestination(doc)
}

func (s *Store) DeleteDestination(ctx context.Context, workspaceID, id string) error {
	res, err := s.destinations.DeleteOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID})
	if err != nil {
		return fmt.Errorf("delete telegram destination %q: %w", id, err)
	}
	if res.DeletedCount == 0 {
		return destinationNotFound(id)
	}
	return nil
}

func (s *Store) FindDestinationByAddress(ctx context.Context, channelID, chatID, threadID string) (*agentsv1.TelegramDestination, error) {
	doc, err := s.destinationDoc(ctx, bson.M{
		"channel_id":        channelID,
		"chat_id":           chatID,
		"message_thread_id": threadID,
	}, fmt.Sprintf("%s/%s/%s", channelID, chatID, threadID))
	if err != nil {
		return nil, err
	}
	return decodeDestination(doc)
}

// --- errors ----------------------------------------------------------------

func channelNotFound(id string) error {
	return fmt.Errorf("telegram channel %q: %w", id, telegramrepo.ErrNotFound)
}

func destinationNotFound(id string) error {
	return fmt.Errorf("telegram destination %q: %w", id, telegramrepo.ErrNotFound)
}

// mapChannelWriteErr turns a duplicate-key error into the specific invariant
// that was violated. Mongo reports the index name, which is the only way to
// tell a key collision from a Bot collision.
func mapChannelWriteErr(ch *agentsv1.TelegramChannel, err error) error {
	if !mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("create telegram channel %q: %w", ch.GetId(), err)
	}
	if indexNamed(err, "uniq_bot_id") {
		return fmt.Errorf("telegram bot %q: %w", ch.GetBotId(), telegramrepo.ErrBotExists)
	}
	return fmt.Errorf("telegram channel %q: %w", ch.GetKey(), telegramrepo.ErrKeyExists)
}

func mapDestinationWriteErr(d *agentsv1.TelegramDestination, err error) error {
	if !mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("create telegram destination %q: %w", d.GetId(), err)
	}
	if indexNamed(err, "uniq_address") {
		return fmt.Errorf("telegram destination for chat %q thread %q: %w",
			d.GetChatId(), d.GetMessageThreadId(), telegramrepo.ErrAddressExists)
	}
	return fmt.Errorf("telegram destination %q: %w", d.GetKey(), telegramrepo.ErrKeyExists)
}

// indexNamed reports whether a write error names the given index. The driver
// surfaces the index name only in the server message, so this is a substring
// match on that message rather than a structured field.
func indexNamed(err error, name string) bool {
	if writeErr, ok := errors.AsType[mongo.WriteException](err); ok {
		for _, we := range writeErr.WriteErrors {
			if strings.Contains(we.Message, name) {
				return true
			}
		}
	}
	if bulkErr, ok := errors.AsType[mongo.BulkWriteException](err); ok {
		for _, we := range bulkErr.WriteErrors {
			if strings.Contains(we.Message, name) {
				return true
			}
		}
	}
	return strings.Contains(err.Error(), name)
}
