// Package mongo is a MongoDB-backed agentop.Repository.
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

	agentoprepo "go.orx.me/apps/butter/internal/repo/agentop"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const operationsCollection = "agent_operations"

// opDoc promotes queryable fields out of the protojson Spec so listing and
// filtering avoid decoding every record.
type opDoc struct {
	ID          string    `bson:"_id"`
	WorkspaceID string    `bson:"workspace_id"`
	AgentID     string    `bson:"agent_id,omitempty"`
	Type        string    `bson:"type,omitempty"`
	Status      string    `bson:"status,omitempty"`
	CreatedAt   time.Time `bson:"created_at,omitempty"`
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
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "agent_id", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("create agent operation indexes: %w", err)
	}
	return nil
}

func (r *Store) Save(ctx context.Context, workspaceID string, op *agentsv1.AgentOperation) error {
	if op.GetWorkspaceId() != workspaceID {
		return fmt.Errorf("workspace mismatch: operation belongs to %q, save requested for %q", op.GetWorkspaceId(), workspaceID)
	}
	d, err := opDocFromProto(op)
	if err != nil {
		return err
	}
	_, err = r.coll.ReplaceOne(ctx, bson.M{"_id": op.GetId(), "workspace_id": workspaceID}, d, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("save agent operation: %w", err)
	}
	return nil
}

func (r *Store) Get(ctx context.Context, workspaceID, id string) (*agentsv1.AgentOperation, error) {
	var d opDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&d)
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
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
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

func (r *Store) ListResumableAcrossWorkspaces(ctx context.Context) ([]*agentsv1.AgentOperation, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"status": bson.M{"$in": bson.A{
		agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING.String(),
		agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED.String(),
	}}}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list resumable agent operations: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.AgentOperation
	for cursor.Next(ctx) {
		var d opDoc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode agent operation: %w", err)
		}
		op, err := decodeOp(d.Spec)
		if err != nil {
			return nil, err
		}
		out = append(out, op)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list resumable agent operations: %w", err)
	}
	return out, nil
}

func opDocFromProto(op *agentsv1.AgentOperation) (*opDoc, error) {
	b, err := protojson.Marshal(op)
	if err != nil {
		return nil, fmt.Errorf("marshal agent operation: %w", err)
	}
	d := &opDoc{
		ID:          op.GetId(),
		WorkspaceID: op.GetWorkspaceId(),
		AgentID:     op.GetAgentId(),
		Type:        op.GetType().String(),
		Status:      op.GetStatus().String(),
		Spec:        string(b),
	}
	if ts := op.GetCreatedAt(); ts != nil {
		d.CreatedAt = ts.AsTime()
	}
	return d, nil
}

func decodeOp(spec string) (*agentsv1.AgentOperation, error) {
	op := &agentsv1.AgentOperation{}
	if err := protojson.Unmarshal([]byte(spec), op); err != nil {
		return nil, fmt.Errorf("unmarshal agent operation: %w", err)
	}
	return op, nil
}
