// Package mongo implements linearprocessing.Repository backed by MongoDB.
//
// A unique index on (app_id, delivery_id) is what makes redelivery safe:
// two workers claiming the same Linear redelivery converge on one record
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

	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const recordsCollection = "linear_processing_records"

type recordDoc struct {
	ID          string    `bson:"_id"`
	WorkspaceID string    `bson:"workspace_id"`
	AppID       string    `bson:"app_id"`
	DeliveryID  string    `bson:"delivery_id"`
	Status      int32     `bson:"status"`
	CreatedAt   time.Time `bson:"created_at"`
	ExpiresAt   time.Time `bson:"expires_at"`
	LeaseToken  string    `bson:"lease_token,omitempty"`
	LeaseExpiry time.Time `bson:"lease_expires_at,omitempty"`
	Spec        string    `bson:"spec"`
}

// Store implements linearprocessing.Repository backed by MongoDB.
type Store struct {
	records *mongo.Collection
}

var _ linearprocessing.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{records: db.Collection(recordsCollection)}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	if _, err := s.records.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// The dedupe guarantee: one record per accepted delivery.
			Keys:    bson.D{{Key: "app_id", Value: 1}, {Key: "delivery_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_app_delivery"),
		},
		{
			Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("workspace_recent"),
		},
		{
			// Mongo removes the record — and its persisted reply — once the
			// retention window passes.
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("ttl_expires_at"),
		},
	}); err != nil {
		return fmt.Errorf("create linear processing indexes: %w", err)
	}
	return nil
}

func decode(doc recordDoc) (*agentsv1.LinearProcessingRecord, error) {
	record := &agentsv1.LinearProcessingRecord{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), record); err != nil {
		return nil, fmt.Errorf("unmarshal linear processing record %q: %w", doc.ID, err)
	}
	record.Id = doc.ID
	return record, nil
}

func encode(record *agentsv1.LinearProcessingRecord) (recordDoc, error) {
	spec, err := protojson.Marshal(record)
	if err != nil {
		return recordDoc{}, fmt.Errorf("marshal linear processing record %q: %w", record.GetId(), err)
	}
	return recordDoc{
		ID:          record.GetId(),
		WorkspaceID: record.GetWorkspaceId(),
		AppID:       record.GetAppId(),
		DeliveryID:  record.GetDeliveryId(),
		Status:      int32(record.GetStatus()),
		CreatedAt:   record.GetCreatedAt().AsTime(),
		ExpiresAt:   record.GetExpiresAt().AsTime(),
		Spec:        string(spec),
	}, nil
}

func (s *Store) Claim(ctx context.Context, record *agentsv1.LinearProcessingRecord, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.LinearProcessingRecord, linearprocessing.ClaimAction, error) {
	filter := bson.M{"app_id": record.GetAppId(), "delivery_id": record.GetDeliveryId()}
	for attempt := 0; attempt < 5; attempt++ {
		var existing recordDoc
		err := s.records.FindOne(ctx, filter).Decode(&existing)
		if errors.Is(err, mongo.ErrNoDocuments) {
			stored, created, createErr := s.create(ctx, record, leaseToken, claimedAt, leaseExpiresAt)
			if createErr != nil {
				return nil, linearprocessing.ClaimAcknowledge, createErr
			}
			if created {
				return stored, linearprocessing.ClaimRunAgent, nil
			}
			continue // lost the insert race; re-read the winner
		}
		if err != nil {
			return nil, linearprocessing.ClaimAcknowledge, fmt.Errorf("read linear processing record: %w", err)
		}
		decoded, err := decode(existing)
		if err != nil {
			return nil, linearprocessing.ClaimAcknowledge, err
		}
		if existing.LeaseToken != "" && existing.LeaseExpiry.After(claimedAt) {
			return decoded, linearprocessing.ClaimAcknowledge, linearprocessing.ErrInProgress
		}
		action := linearprocessing.RecoveryAction(decoded)
		linearprocessing.MarkInterruptedUncertain(decoded)
		if action != linearprocessing.ClaimAcknowledge {
			decoded.Attempts++
		}
		decoded.UpdatedAt = timestamppb.New(claimedAt)
		next, err := encode(decoded)
		if err != nil {
			return nil, linearprocessing.ClaimAcknowledge, err
		}
		if action != linearprocessing.ClaimAcknowledge {
			next.LeaseToken = leaseToken
			next.LeaseExpiry = leaseExpiresAt
		}
		claimFilter := bson.M{"_id": existing.ID, "spec": existing.Spec}
		if existing.LeaseToken != "" {
			claimFilter["lease_token"] = existing.LeaseToken
		}
		res, err := s.records.ReplaceOne(ctx, claimFilter, next)
		if err != nil {
			return nil, linearprocessing.ClaimAcknowledge, fmt.Errorf("claim linear processing record: %w", err)
		}
		if res.MatchedCount == 1 {
			return decoded, action, nil
		}
	}
	return nil, linearprocessing.ClaimAcknowledge, linearprocessing.ErrInProgress
}

// create inserts a fresh record, reporting false when another worker's
// insert won the unique index.
func (s *Store) create(ctx context.Context, record *agentsv1.LinearProcessingRecord, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.LinearProcessingRecord, bool, error) {
	stored := proto.Clone(record).(*agentsv1.LinearProcessingRecord)
	if stored.GetId() == "" {
		stored.Id = uuid.NewString()
	}
	if stored.GetInvocationId() == "" {
		stored.InvocationId = uuid.NewString()
	}
	stored.Attempts = 1
	stored.CreatedAt = timestamppb.New(claimedAt)
	stored.UpdatedAt = timestamppb.New(claimedAt)
	stored.ExpiresAt = timestamppb.New(claimedAt.Add(linearprocessing.RetentionPeriod))
	doc, err := encode(stored)
	if err != nil {
		return nil, false, err
	}
	doc.LeaseToken = leaseToken
	doc.LeaseExpiry = leaseExpiresAt
	if _, err := s.records.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("create linear processing record: %w", err)
	}
	return stored, true, nil
}

func (s *Store) ClaimForResend(ctx context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.LinearProcessingRecord, error) {
	for attempt := 0; attempt < 5; attempt++ {
		var existing recordDoc
		if err := s.records.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&existing); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, fmt.Errorf("linear processing record %q: %w", id, linearprocessing.ErrNotFound)
			}
			return nil, fmt.Errorf("read linear processing record: %w", err)
		}
		if existing.LeaseToken != "" && existing.LeaseExpiry.After(claimedAt) {
			return nil, linearprocessing.ErrInProgress
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
		}
		res, err := s.records.ReplaceOne(ctx, claimFilter, next)
		if err != nil {
			return nil, fmt.Errorf("claim linear processing record for resend: %w", err)
		}
		if res.MatchedCount == 1 {
			return record, nil
		}
	}
	return nil, linearprocessing.ErrInProgress
}

func (s *Store) set(ctx context.Context, filter bson.M, record *agentsv1.LinearProcessingRecord) (*agentsv1.LinearProcessingRecord, int64, error) {
	stored := proto.Clone(record).(*agentsv1.LinearProcessingRecord)
	stored.UpdatedAt = timestamppb.New(time.Now().UTC())
	doc, err := encode(stored)
	if err != nil {
		return nil, 0, err
	}
	res, err := s.records.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"spec": doc.Spec, "status": doc.Status}})
	if err != nil {
		return nil, 0, fmt.Errorf("update linear processing record %q: %w", stored.GetId(), err)
	}
	return stored, res.MatchedCount, nil
}

func (s *Store) UpdateClaimed(ctx context.Context, record *agentsv1.LinearProcessingRecord, leaseToken string) (*agentsv1.LinearProcessingRecord, error) {
	stored, matched, err := s.set(ctx, bson.M{"_id": record.GetId(), "lease_token": leaseToken}, record)
	if err != nil {
		return nil, err
	}
	if matched == 0 {
		return nil, linearprocessing.ErrLeaseLost
	}
	return stored, nil
}

func (s *Store) Update(ctx context.Context, record *agentsv1.LinearProcessingRecord) (*agentsv1.LinearProcessingRecord, error) {
	stored, matched, err := s.set(ctx, bson.M{"_id": record.GetId()}, record)
	if err != nil {
		return nil, err
	}
	if matched == 0 {
		return nil, fmt.Errorf("linear processing record %q: %w", record.GetId(), linearprocessing.ErrNotFound)
	}
	return stored, nil
}

func (s *Store) RenewClaim(ctx context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error {
	res, err := s.records.UpdateOne(ctx,
		bson.M{"_id": id, "workspace_id": workspaceID, "lease_token": leaseToken},
		bson.M{"$set": bson.M{"lease_expires_at": leaseExpiresAt}})
	if err != nil {
		return fmt.Errorf("renew linear processing claim %q: %w", id, err)
	}
	if res.MatchedCount == 0 {
		return linearprocessing.ErrLeaseLost
	}
	return nil
}

func (s *Store) ReleaseClaim(ctx context.Context, workspaceID, id, leaseToken string) error {
	res, err := s.records.UpdateOne(ctx,
		bson.M{"_id": id, "workspace_id": workspaceID, "lease_token": leaseToken},
		bson.M{"$unset": bson.M{"lease_token": "", "lease_expires_at": ""}})
	if err != nil {
		return fmt.Errorf("release linear processing claim %q: %w", id, err)
	}
	if res.MatchedCount == 0 {
		return linearprocessing.ErrLeaseLost
	}
	return nil
}

func (s *Store) Get(ctx context.Context, workspaceID, id string) (*agentsv1.LinearProcessingRecord, error) {
	var doc recordDoc
	err := s.records.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("linear processing record %q: %w", id, linearprocessing.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("read linear processing record %q: %w", id, err)
	}
	return decode(doc)
}

func (s *Store) List(ctx context.Context, filter linearprocessing.Filter) ([]*agentsv1.LinearProcessingRecord, error) {
	query := bson.M{"workspace_id": filter.WorkspaceID}
	if filter.AppID != "" {
		query["app_id"] = filter.AppID
	}
	if filter.Status != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_UNSPECIFIED {
		query["status"] = int32(filter.Status)
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	cursor, err := s.records.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("list linear processing records: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []recordDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode linear processing records: %w", err)
	}
	out := make([]*agentsv1.LinearProcessingRecord, 0, len(docs))
	for _, doc := range docs {
		record, err := decode(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, nil
}
