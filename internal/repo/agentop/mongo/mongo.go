// Package mongo is a MongoDB-backed agentop.Repository.
package mongo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentoprepo "go.orx.me/apps/butter/internal/repo/agentop"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const operationsCollection = "agent_operations"

// opDoc promotes queryable fields out of the protojson Spec so listing and
// filtering avoid decoding every record.
type opDoc struct {
	ID          string    `bson:"_id"`
	OperationID string    `bson:"operation_id"`
	WorkspaceID string    `bson:"workspace_id"`
	AgentID     string    `bson:"agent_id,omitempty"`
	Type        string    `bson:"type,omitempty"`
	Status      string    `bson:"status,omitempty"`
	CreatedAt   time.Time `bson:"created_at,omitempty"`
	LeaseToken  string    `bson:"lease_token,omitempty"`
	LeaseExpiry time.Time `bson:"lease_expires_at,omitempty"`
	Spec        string    `bson:"spec"`
}

// Store is a MongoDB-backed agentop.Repository.
type Store struct {
	coll *mongo.Collection
}

// New returns a MongoDB-backed operation store.
func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(operationsCollection)}
}

func (r *Store) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "operation_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{
				"operation_id": bson.M{"$type": "string"},
			}),
		},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "created_at", Value: -1}, {Key: "operation_id", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "agent_id", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("create agent operation indexes: %w", err)
	}
	return nil
}

func (r *Store) Create(ctx context.Context, workspaceID string, op *agentsv1.AgentOperation) error {
	if op.GetWorkspaceId() != workspaceID {
		return fmt.Errorf("workspace mismatch: operation belongs to %q, create requested for %q", op.GetWorkspaceId(), workspaceID)
	}
	if _, err := r.Get(ctx, workspaceID, op.GetId()); err == nil {
		return agentoprepo.ErrAlreadyExists
	} else if !errors.Is(err, agentoprepo.ErrNotFound) {
		return err
	}
	d, err := opDocFromProto(op, "", time.Time{})
	if err != nil {
		return err
	}
	_, err = r.coll.InsertOne(ctx, d)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return agentoprepo.ErrAlreadyExists
		}
		return fmt.Errorf("create agent operation: %w", err)
	}
	return nil
}

func (r *Store) Claim(ctx context.Context, workspaceID, id, leaseToken string, claimedAt, leaseExpiresAt time.Time) (*agentsv1.AgentOperation, error) {
	for attempt := 0; attempt < 5; attempt++ {
		var current opDoc
		if err := r.coll.FindOne(ctx, operationLookupFilter(workspaceID, id)).Decode(&current); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, agentoprepo.ErrNotFound
			}
			return nil, fmt.Errorf("get agent operation for claim: %w", err)
		}
		op, err := decodeOp(current.Spec)
		if err != nil {
			return nil, err
		}
		switch op.GetStatus() {
		case agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED:
			return op, agentoprepo.ErrCompleted
		case agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING:
			activeUntil := current.LeaseExpiry
			if activeUntil.IsZero() {
				if updatedAt := op.GetUpdatedAt(); updatedAt != nil {
					activeUntil = updatedAt.AsTime().Add(leaseExpiresAt.Sub(claimedAt))
				}
			}
			if activeUntil.After(claimedAt) {
				return op, agentoprepo.ErrInProgress
			}
		}
		op.AttemptCount++
		op.Status = agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING
		op.Error = ""
		op.UpdatedAt = timestamppb.New(claimedAt)
		next, err := opDocFromProto(op, leaseToken, leaseExpiresAt)
		if err != nil {
			return nil, err
		}
		next.ID = current.ID
		claimFilter := bson.M{"_id": current.ID, "spec": current.Spec}
		if current.LeaseExpiry.IsZero() {
			claimFilter["$or"] = bson.A{
				bson.M{"lease_expires_at": bson.M{"$exists": false}},
				bson.M{"lease_expires_at": time.Time{}},
			}
		} else {
			claimFilter["lease_expires_at"] = current.LeaseExpiry
		}
		res, err := r.coll.ReplaceOne(ctx, claimFilter, next)
		if err != nil {
			return nil, fmt.Errorf("claim agent operation: %w", err)
		}
		if res.MatchedCount == 1 {
			return op, nil
		}
	}
	return nil, agentoprepo.ErrInProgress
}

func (r *Store) RenewLease(ctx context.Context, workspaceID, id, leaseToken string, leaseExpiresAt time.Time) error {
	res, err := r.coll.UpdateOne(ctx, bson.M{
		"workspace_id": workspaceID,
		"operation_id": id,
		"lease_token":  leaseToken,
	}, bson.M{"$set": bson.M{"lease_expires_at": leaseExpiresAt}})
	if err != nil {
		return fmt.Errorf("renew agent operation lease: %w", err)
	}
	if res.MatchedCount == 0 {
		return agentoprepo.ErrLeaseLost
	}
	return nil
}

func (r *Store) SaveClaimed(ctx context.Context, workspaceID, leaseToken string, op *agentsv1.AgentOperation) error {
	if op.GetWorkspaceId() != workspaceID {
		return fmt.Errorf("workspace mismatch: operation belongs to %q, save requested for %q", op.GetWorkspaceId(), workspaceID)
	}
	d, err := opDocFromProto(op, leaseToken, time.Time{})
	if err != nil {
		return err
	}
	res, err := r.coll.UpdateOne(ctx, bson.M{
		"workspace_id": workspaceID,
		"operation_id": op.GetId(),
		"lease_token":  leaseToken,
	}, bson.M{"$set": bson.M{
		"operation_id": d.OperationID,
		"workspace_id": d.WorkspaceID,
		"agent_id":     d.AgentID,
		"type":         d.Type,
		"status":       d.Status,
		"created_at":   d.CreatedAt,
		"spec":         d.Spec,
	}})
	if err != nil {
		return fmt.Errorf("save claimed agent operation: %w", err)
	}
	if res.MatchedCount == 0 {
		return agentoprepo.ErrLeaseLost
	}
	return nil
}

func (r *Store) Get(ctx context.Context, workspaceID, id string) (*agentsv1.AgentOperation, error) {
	var d opDoc
	err := r.coll.FindOne(ctx, operationLookupFilter(workspaceID, id)).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, agentoprepo.ErrNotFound
		}
		return nil, fmt.Errorf("get agent operation: %w", err)
	}
	return decodeOp(d.Spec)
}

func (r *Store) List(ctx context.Context, workspaceID string, status agentsv1.AgentOperationStatus, pageSize int32, pageToken string) ([]*agentsv1.AgentOperation, string, error) {
	q := bson.M{"workspace_id": workspaceID}
	if status != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_UNSPECIFIED {
		q["status"] = status.String()
	}
	size := agentoprepo.ClampPageSize(pageSize)
	offset := agentoprepo.DecodePageToken(pageToken)

	cursor, err := r.coll.Find(ctx, q, options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "operation_id", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(size+1)),
	)
	if err != nil {
		return nil, "", fmt.Errorf("list agent operations: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.AgentOperation
	for cursor.Next(ctx) {
		var d opDoc
		if err := cursor.Decode(&d); err != nil {
			return nil, "", fmt.Errorf("decode agent operation: %w", err)
		}
		op, err := decodeOp(d.Spec)
		if err != nil {
			return nil, "", err
		}
		out = append(out, op)
	}
	if err := cursor.Err(); err != nil {
		return nil, "", fmt.Errorf("list agent operations: %w", err)
	}

	next := ""
	if len(out) > int(size) {
		out = out[:size]
		next = agentoprepo.EncodePageToken(offset + len(out))
	}
	return out, next, nil
}

func opDocFromProto(op *agentsv1.AgentOperation, leaseToken string, leaseExpiresAt time.Time) (*opDoc, error) {
	b, err := protojson.Marshal(op)
	if err != nil {
		return nil, fmt.Errorf("marshal agent operation: %w", err)
	}
	d := &opDoc{
		ID:          operationDocumentID(op.GetWorkspaceId(), op.GetId()),
		OperationID: op.GetId(),
		WorkspaceID: op.GetWorkspaceId(),
		AgentID:     op.GetAgentId(),
		Type:        op.GetType().String(),
		Status:      op.GetStatus().String(),
		LeaseToken:  leaseToken,
		LeaseExpiry: leaseExpiresAt,
		Spec:        string(b),
	}
	if ts := op.GetCreatedAt(); ts != nil {
		d.CreatedAt = ts.AsTime()
	}
	return d, nil
}

func operationDocumentID(workspaceID, operationID string) string {
	sum := sha256.Sum256([]byte(workspaceID + "\x00" + operationID))
	return hex.EncodeToString(sum[:])
}

func operationLookupFilter(workspaceID, operationID string) bson.M {
	return bson.M{
		"workspace_id": workspaceID,
		"$or": bson.A{
			bson.M{"_id": operationDocumentID(workspaceID, operationID)},
			bson.M{"_id": operationID},
		},
	}
}

func decodeOp(spec string) (*agentsv1.AgentOperation, error) {
	op := &agentsv1.AgentOperation{}
	if err := protojson.Unmarshal([]byte(spec), op); err != nil {
		return nil, fmt.Errorf("unmarshal agent operation: %w", err)
	}
	return op, nil
}
