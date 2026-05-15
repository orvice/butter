package mongo

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

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
// Records are encoded as protojson under the `spec` field; pagination is a
// simple offset cursor base64-encoded into page_token.
type Store struct {
	coll *mongo.Collection
}

type doc struct {
	ID   string `bson:"_id"`
	Spec string `bson:"spec"`
}

func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collectionName)}
}

func (s *Store) Save(ctx context.Context, inv *agentsv1.Invocation) error {
	b, err := protojson.Marshal(inv)
	if err != nil {
		return fmt.Errorf("marshal invocation: %w", err)
	}
	_, err = s.coll.ReplaceOne(ctx, bson.M{"_id": inv.GetId()}, doc{ID: inv.GetId(), Spec: string(b)}, options.Replace().SetUpsert(true))
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
	if filter.AgentName != "" {
		q["spec"] = bson.M{"$regex": fmt.Sprintf(`"agentName":\s*"%s"`, regexEscape(filter.AgentName))}
	}
	// Filter by session_id by scanning the encoded spec; for our scale this is
	// acceptable and avoids needing a flattened schema. Combine with agent if both set.
	if filter.SessionID != "" {
		clause := bson.M{"$regex": fmt.Sprintf(`"sessionId":\s*"%s"`, regexEscape(filter.SessionID))}
		if existing, ok := q["spec"]; ok {
			q = bson.M{"$and": []bson.M{{"spec": existing}, {"spec": clause}}}
		} else {
			q["spec"] = clause
		}
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
		SetSort(bson.M{"spec": -1}).
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
		SetSort(bson.M{"spec": -1}).
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

// regexEscape escapes characters that are special in MongoDB $regex patterns.
func regexEscape(s string) string {
	var out []byte
	for _, b := range []byte(s) {
		switch b {
		case '.', '+', '*', '?', '(', ')', '|', '[', ']', '{', '}', '^', '$', '\\':
			out = append(out, '\\', b)
		default:
			out = append(out, b)
		}
	}
	return string(out)
}
