package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
)

type Store struct{ col *mongo.Collection }
type doc struct {
	Key   string                 `bson:"_id"`
	Value *plane.DeliveryAttempt `bson:"value"`
}

func New(db *mongo.Database) *Store { return &Store{col: db.Collection("plane_reply_deliveries")} }
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, e := s.col.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "updated_at", Value: 1}}, Options: options.Index()})
	return e
}
func (s *Store) Get(ctx context.Context, k string) (*plane.DeliveryAttempt, error) {
	var d doc
	e := s.col.FindOne(ctx, bson.M{"_id": k}).Decode(&d)
	if e == mongo.ErrNoDocuments {
		return nil, plane.ErrDeliveryNotFound
	}
	if e != nil {
		return nil, e
	}
	return d.Value, nil
}
func (s *Store) Put(ctx context.Context, r *plane.DeliveryAttempt) error {
	_, e := s.col.ReplaceOne(ctx, bson.M{"_id": r.Key}, doc{Key: r.Key, Value: r}, options.Replace().SetUpsert(true))
	return e
}
func (s *Store) ListAll(ctx context.Context) ([]*plane.DeliveryAttempt, error) {
	cur, e := s.col.Find(ctx, bson.M{})
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var ds []doc
	if e = cur.All(ctx, &ds); e != nil {
		return nil, e
	}
	out := make([]*plane.DeliveryAttempt, 0, len(ds))
	for _, d := range ds {
		if d.Value == nil {
			return nil, errors.New("invalid plane delivery record")
		}
		out = append(out, d.Value)
	}
	return out, nil
}
func (s *Store) ListRecoverable(ctx context.Context) ([]*plane.DeliveryAttempt, error) {
	cur, e := s.col.Find(ctx, bson.M{"value.status": bson.M{"$in": []string{string(plane.DeliveryPending), string(plane.DeliveryRetryable)}}})
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var ds []doc
	if e = cur.All(ctx, &ds); e != nil {
		return nil, e
	}
	out := make([]*plane.DeliveryAttempt, 0, len(ds))
	for _, d := range ds {
		if d.Value == nil {
			return nil, errors.New("invalid plane delivery record")
		}
		out = append(out, d.Value)
	}
	return out, nil
}
