package mongo

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"

	"go.orx.me/apps/butter/internal/repo/invocation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const collectionName = "invocations"

// Store is a MongoDB-backed implementation of invocation.Repository.
//
// Records keep the canonical protojson payload in `spec` and denormalize the
// indexable identifiers (workspace_id, agent_name, session_id, started_at)
// into top-level fields so List queries can do exact-match BSON filters
// instead of regex scans over the JSON blob.
type Store struct {
	coll *mongo.Collection
}

type doc struct {
	ID          string    `bson:"_id"`
	WorkspaceID string    `bson:"workspace_id,omitempty"`
	AgentName   string    `bson:"agent_name,omitempty"`
	SessionID   string    `bson:"session_id,omitempty"`
	StartedAt   time.Time `bson:"started_at,omitempty"`
	Spec        string    `bson:"spec"`
}

func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collectionName)}
}

// EnsureIndexes creates the indexes used by List and ListRecent queries.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "agent_name", Value: 1}, {Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "session_id", Value: 1}, {Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "started_at", Value: -1}}},
	})
	if err != nil {
		return fmt.Errorf("create invocation indexes: %w", err)
	}
	return nil
}

func (s *Store) Save(ctx context.Context, inv *agentsv1.Invocation) error {
	b, err := protojson.Marshal(inv)
	if err != nil {
		return fmt.Errorf("marshal invocation: %w", err)
	}
	d := doc{
		ID:          inv.GetId(),
		WorkspaceID: inv.GetWorkspaceId(),
		AgentName:   inv.GetAgentName(),
		SessionID:   inv.GetSessionId(),
		Spec:        string(b),
	}
	if ts := inv.GetStartedAt(); ts != nil {
		d.StartedAt = ts.AsTime()
	}
	_, err = s.coll.ReplaceOne(ctx, bson.M{"_id": inv.GetId()}, d, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("upsert invocation: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (*agentsv1.Invocation, error) {
	var d doc
	err := s.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, invocation.ErrNotFound
		}
		return nil, fmt.Errorf("get invocation: %w", err)
	}
	return decode(&d)
}

func (s *Store) List(ctx context.Context, filter invocation.ListFilter, pageSize int32, pageToken string) ([]*agentsv1.Invocation, string, int32, error) {
	q := bson.M{}
	if filter.WorkspaceID != "" {
		q["workspace_id"] = filter.WorkspaceID
	}
	if filter.AgentName != "" {
		q["agent_name"] = filter.AgentName
	}
	if filter.SessionID != "" {
		q["session_id"] = filter.SessionID
	}

	total, err := s.coll.CountDocuments(ctx, q)
	if err != nil {
		return nil, "", 0, fmt.Errorf("count invocations: %w", err)
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	offset := decodeToken(pageToken)

	opts := options.Find().
		SetSort(bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(pageSize))

	cursor, err := s.coll.Find(ctx, q, opts)
	if err != nil {
		return nil, "", 0, fmt.Errorf("list invocations: %w", err)
	}
	defer cursor.Close(ctx)

	out, err := drain(ctx, cursor)
	if err != nil {
		return nil, "", 0, err
	}

	next := ""
	if int64(offset)+int64(len(out)) < total {
		next = encodeToken(offset + len(out))
	}
	return out, next, int32(total), nil
}

func (s *Store) ListRecent(ctx context.Context, limit int32, pageToken string) ([]*agentsv1.Invocation, string, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	offset := decodeToken(pageToken)

	opts := options.Find().
		SetSort(bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := s.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, "", fmt.Errorf("list recent invocations: %w", err)
	}
	defer cursor.Close(ctx)

	out, err := drain(ctx, cursor)
	if err != nil {
		return nil, "", err
	}
	next := ""
	if int32(len(out)) == limit {
		next = encodeToken(offset + len(out))
	}
	return out, next, nil
}

func drain(ctx context.Context, cursor *mongo.Cursor) ([]*agentsv1.Invocation, error) {
	var out []*agentsv1.Invocation
	for cursor.Next(ctx) {
		var d doc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode invocation: %w", err)
		}
		inv, err := decode(&d)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, nil
}

func decode(d *doc) (*agentsv1.Invocation, error) {
	inv := &agentsv1.Invocation{}
	if err := protojson.Unmarshal([]byte(d.Spec), inv); err != nil {
		return nil, fmt.Errorf("unmarshal invocation: %w", err)
	}
	return inv, nil
}

func decodeToken(token string) int {
	if token == "" {
		return 0
	}
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(string(raw))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func encodeToken(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}
