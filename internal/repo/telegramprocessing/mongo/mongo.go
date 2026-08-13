// Package mongo implements telegramprocessing.Repository backed by MongoDB.
//
// A unique index on (channel_id, update_id) is what makes duplicate delivery
// safe: two workers claiming the same Telegram retry converge on one record
// instead of each running the Agent.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/telegramprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const recordsCollection = "telegram_processing_records"

type recordDoc struct {
	ID            string    `bson:"_id"`
	WorkspaceID   string    `bson:"workspace_id"`
	ChannelID     string    `bson:"channel_id"`
	DestinationID string    `bson:"destination_id,omitempty"`
	UpdateID      int64     `bson:"update_id"`
	Status        int32     `bson:"status"`
	CreatedAt     time.Time `bson:"created_at"`
	ExpiresAt     time.Time `bson:"expires_at"`
	LeaseToken    string    `bson:"lease_token,omitempty"`
	LeaseExpiry   time.Time `bson:"lease_expires_at,omitempty"`
	Spec          string    `bson:"spec"`
}

// Store implements telegramprocessing.Repository backed by MongoDB.
type Store struct {
	records *mongo.Collection
}

var _ telegramprocessing.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{records: db.Collection(recordsCollection)}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	if _, err := s.records.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// The dedupe guarantee: one record per accepted update.
			Keys:    bson.D{{Key: "channel_id", Value: 1}, {Key: "update_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_channel_update"),
		},
		{
			Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("workspace_recent"),
		},
		{
			// Mongo removes the record — including the persisted response and
			// its segments — once the retention window passes.
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("ttl_expires_at"),
		},
	}); err != nil {
		return fmt.Errorf("create telegram processing indexes: %w", err)
	}
	return nil
}

func decode(doc recordDoc) (*agentsv1.TelegramProcessingRecord, error) {
	record := &agentsv1.TelegramProcessingRecord{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), record); err != nil {
		return nil, fmt.Errorf("unmarshal telegram processing record %q: %w", doc.ID, err)
	}
	record.Id = doc.ID
	return record, nil
}

func encode(record *agentsv1.TelegramProcessingRecord) (recordDoc, error) {
	spec, err := protojson.Marshal(record)
	if err != nil {
		return recordDoc{}, fmt.Errorf("marshal telegram processing record %q: %w", record.GetId(), err)
	}
	return recordDoc{
		ID:            record.GetId(),
		WorkspaceID:   record.GetWorkspaceId(),
		ChannelID:     record.GetChannelId(),
		DestinationID: record.GetDestinationId(),
		UpdateID:      record.GetUpdateId(),
		Status:        int32(record.GetStatus()),
		CreatedAt:     record.GetCreatedAt().AsTime(),
		ExpiresAt:     record.GetExpiresAt().AsTime(),
		Spec:          string(spec),
	}, nil
}

func (s *Store) Claim(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.TelegramProcessingRecord, telegramprocessing.ClaimAction, error) {
	filter := bson.M{"channel_id": record.GetChannelId(), "update_id": record.GetUpdateId()}
	for attempt := 0; attempt < 5; attempt++ {
		var existing recordDoc
		err := s.records.FindOne(ctx, filter).Decode(&existing)
		switch {
		case err == nil:
			if existing.LeaseToken != "" && existing.LeaseExpiry.After(claimedAt) {
				decoded, decErr := decode(existing)
				if decErr != nil {
					return nil, telegramprocessing.ClaimAcknowledge, decErr
				}
				return decoded, telegramprocessing.ClaimAcknowledge, telegramprocessing.ErrInProgress
			}
			decoded, decErr := decode(existing)
			if decErr != nil {
				return nil, telegramprocessing.ClaimAcknowledge, decErr
			}
			action := telegramprocessing.RecoveryAction(decoded)
			telegramprocessing.MarkInterruptedUncertain(decoded)
			if action != telegramprocessing.ClaimAcknowledge {
				decoded.Attempts++
			}
			decoded.UpdatedAt = timestamppb.New(claimedAt)
			next, encErr := encode(decoded)
			if encErr != nil {
				return nil, telegramprocessing.ClaimAcknowledge, encErr
			}
			if action != telegramprocessing.ClaimAcknowledge {
				next.LeaseToken = leaseToken
				next.LeaseExpiry = leaseExpiresAt
			}
			claimFilter := bson.M{"_id": existing.ID, "spec": existing.Spec}
			if existing.LeaseToken != "" {
				claimFilter["lease_token"] = existing.LeaseToken
				claimFilter["lease_expires_at"] = existing.LeaseExpiry
			}
			res, replaceErr := s.records.ReplaceOne(ctx, claimFilter, next)
			if replaceErr != nil {
				return nil, telegramprocessing.ClaimAcknowledge, fmt.Errorf("claim telegram processing record: %w", replaceErr)
			}
			if res.MatchedCount == 1 {
				return decoded, action, nil
			}
			continue
		case errors.Is(err, mongo.ErrNoDocuments):
			attempt = 5
		default:
			return nil, telegramprocessing.ClaimAcknowledge, fmt.Errorf("read telegram processing record: %w", err)
		}
	}

	stored := proto.Clone(record).(*agentsv1.TelegramProcessingRecord)
	if stored.GetId() == "" {
		stored.Id = uuid.NewString()
	}
	if stored.GetInvocationId() == "" {
		stored.InvocationId = uuid.NewString()
	}
	stored.Attempts = 1
	stored.CreatedAt = timestamppb.New(claimedAt)
	stored.UpdatedAt = timestamppb.New(claimedAt)
	stored.ExpiresAt = timestamppb.New(claimedAt.Add(telegramprocessing.RetentionPeriod))

	doc, err := encode(stored)
	if err != nil {
		return nil, telegramprocessing.ClaimAcknowledge, err
	}
	doc.LeaseToken = leaseToken
	doc.LeaseExpiry = leaseExpiresAt
	if _, err := s.records.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			// Another worker won the race; re-read and claim that record.
			return s.Claim(ctx, record, leaseToken, claimedAt, leaseExpiresAt)
		}
		return nil, telegramprocessing.ClaimAcknowledge, fmt.Errorf("create telegram processing record: %w", err)
	}
	return stored, telegramprocessing.ClaimRunAgent, nil
}

func (s *Store) ClaimDelivery(ctx context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.TelegramProcessingRecord, error) {
	for attempt := 0; attempt < 5; attempt++ {
		var existing recordDoc
		if err := s.records.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&existing); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, telegramprocessing.ErrNotFound
			}
			return nil, fmt.Errorf("read telegram processing record for delivery claim: %w", err)
		}
		if existing.LeaseToken != "" && existing.LeaseExpiry.After(claimedAt) {
			return nil, telegramprocessing.ErrInProgress
		}
		record, err := decode(existing)
		if err != nil {
			return nil, err
		}
		record.Attempts++
		record.UpdatedAt = timestamppb.New(claimedAt)
		next, err := encode(record)
		if err != nil {
			return nil, err
		}
		next.LeaseToken = leaseToken
		next.LeaseExpiry = leaseExpiresAt
		claimFilter := bson.M{"_id": existing.ID, "spec": existing.Spec}
		if existing.LeaseToken != "" {
			claimFilter["lease_token"] = existing.LeaseToken
			claimFilter["lease_expires_at"] = existing.LeaseExpiry
		}
		res, err := s.records.ReplaceOne(ctx, claimFilter, next)
		if err != nil {
			return nil, fmt.Errorf("claim telegram delivery: %w", err)
		}
		if res.MatchedCount == 1 {
			return record, nil
		}
	}
	return nil, telegramprocessing.ErrInProgress
}

func (s *Store) UpdateClaimed(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string) (*agentsv1.TelegramProcessingRecord, error) {
	stored := proto.Clone(record).(*agentsv1.TelegramProcessingRecord)
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	doc, err := encode(stored)
	if err != nil {
		return nil, err
	}
	res, err := s.records.UpdateOne(ctx, bson.M{"_id": stored.GetId(), "lease_token": leaseToken},
		bson.M{"$set": bson.M{"spec": doc.Spec, "status": doc.Status}})
	if err != nil {
		return nil, fmt.Errorf("update claimed telegram processing record %q: %w", stored.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, telegramprocessing.ErrLeaseLost
	}
	return stored, nil
}

func (s *Store) RenewClaim(ctx context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error {
	res, err := s.records.UpdateOne(ctx, bson.M{
		"_id":          id,
		"workspace_id": workspaceID,
		"lease_token":  leaseToken,
	}, bson.M{"$set": bson.M{"lease_expires_at": leaseExpiresAt}})
	if err != nil {
		return fmt.Errorf("renew telegram processing claim %q: %w", id, err)
	}
	if res.MatchedCount == 0 {
		return telegramprocessing.ErrLeaseLost
	}
	return nil
}

func (s *Store) ReleaseClaim(ctx context.Context, workspaceID, id, leaseToken string) error {
	res, err := s.records.UpdateOne(ctx, bson.M{
		"_id":          id,
		"workspace_id": workspaceID,
		"lease_token":  leaseToken,
	}, bson.M{"$unset": bson.M{"lease_token": "", "lease_expires_at": ""}})
	if err != nil {
		return fmt.Errorf("release telegram processing claim %q: %w", id, err)
	}
	if res.MatchedCount == 0 {
		return telegramprocessing.ErrLeaseLost
	}
	return nil
}

func (s *Store) Update(ctx context.Context, record *agentsv1.TelegramProcessingRecord) (*agentsv1.TelegramProcessingRecord, error) {
	stored := proto.Clone(record).(*agentsv1.TelegramProcessingRecord)
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	doc, err := encode(stored)
	if err != nil {
		return nil, err
	}
	res, err := s.records.UpdateOne(ctx, bson.M{"_id": stored.GetId()},
		bson.M{"$set": bson.M{"spec": doc.Spec, "status": doc.Status}})
	if err != nil {
		return nil, fmt.Errorf("update telegram processing record %q: %w", stored.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("telegram processing record %q: %w", stored.GetId(), telegramprocessing.ErrNotFound)
	}
	return stored, nil
}

func (s *Store) Get(ctx context.Context, workspaceID, id string) (*agentsv1.TelegramProcessingRecord, error) {
	var doc recordDoc
	err := s.records.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("telegram processing record %q: %w", id, telegramprocessing.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("read telegram processing record %q: %w", id, err)
	}
	return decode(doc)
}

func (s *Store) List(ctx context.Context, filter telegramprocessing.Filter) ([]*agentsv1.TelegramProcessingRecord, error) {
	query := bson.M{"workspace_id": filter.WorkspaceID}
	if filter.ChannelID != "" {
		query["channel_id"] = filter.ChannelID
	}
	if filter.DestinationID != "" {
		query["destination_id"] = filter.DestinationID
	}
	if filter.Status != agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_UNSPECIFIED {
		query["status"] = int32(filter.Status)
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	cursor, err := s.records.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("list telegram processing records: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []recordDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode telegram processing records: %w", err)
	}
	out := make([]*agentsv1.TelegramProcessingRecord, 0, len(docs))
	for _, doc := range docs {
		record, err := decode(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, nil
}
