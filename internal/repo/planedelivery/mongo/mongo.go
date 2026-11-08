package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
	"time"
)

type Store struct{ col *mongo.Collection }
type doc struct {
	Key   string                 `bson:"_id"`
	Value *plane.DeliveryAttempt `bson:"value"`
}

func New(db *mongo.Database) *Store { return &Store{col: db.Collection("plane_reply_deliveries")} }
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, e := s.col.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "value.updatedat", Value: 1}}, Options: options.Index()})
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
func (s *Store) Create(ctx context.Context, r *plane.DeliveryAttempt) error {
	_, err := s.col.InsertOne(ctx, doc{Key: r.Key, Value: r})
	if mongo.IsDuplicateKeyError(err) {
		return plane.ErrDeliveryAlreadyExists
	}
	return err
}
func (s *Store) Claim(ctx context.Context, key, token string, now time.Time, lease time.Duration) (*plane.DeliveryAttempt, error) {
	filter := bson.M{
		"_id":          key,
		"value.status": string(plane.DeliveryRetryable),
		"$or": []bson.M{
			{"value.lease_until": bson.M{"$exists": false}},
			{"value.lease_until": nil},
			{"value.lease_until": bson.M{"$lte": now}},
		},
	}
	until := now.Add(lease)
	update := bson.M{"$set": bson.M{
		"value.status":      string(plane.DeliveryPending),
		"value.claim_token": token,
		"value.lease_until": until,
		"value.claimedat":   now,
		"value.updatedat":   now,
	}, "$inc": bson.M{"value.attempts": 1}}
	var d doc
	err := s.col.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&d)
	if err == mongo.ErrNoDocuments {
		return nil, plane.ErrDeliveryInProgress
	}
	if err != nil {
		return nil, err
	}
	if d.Value == nil {
		return nil, errors.New("invalid plane delivery record")
	}
	return d.Value, nil
}
func (s *Store) PutOwned(ctx context.Context, r *plane.DeliveryAttempt, token string) error {
	result, err := s.col.UpdateOne(ctx, bson.M{"_id": r.Key, "value.claim_token": token}, bson.M{
		"$set": bson.M{"value": r},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return plane.ErrDeliveryLeaseLost
	}
	return nil
}
func (s *Store) CompareAndSwap(ctx context.Context, before, after *plane.DeliveryAttempt) error {
	filter := bson.M{"_id": before.Key, "value.status": before.Status, "value.attempts": before.Attempts}
	if before.ClaimToken == "" {
		filter["$or"] = []bson.M{
			{"value.claim_token": ""},
			{"value.claim_token": nil},
		}
	} else {
		filter["value.claim_token"] = before.ClaimToken
	}
	result, err := s.col.ReplaceOne(ctx, filter, doc{Key: before.Key, Value: after})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return plane.ErrDeliveryInProgress
	}
	return nil
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
