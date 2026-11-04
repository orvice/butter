package mongostore

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

const budgets = "work_station_budgets"

func (s *Store) ReserveBudget(ctx context.Context, key string, spacing time.Duration) (time.Duration, error) {
	filter := bson.M{"_id": key, "$expr": bson.M{"$lte": bson.A{bson.M{"$ifNull": bson.A{"$next", time.Unix(0, 0)}}, "$$NOW"}}}
	update := mongo.Pipeline{stage("$set", bson.M{"next": bson.M{"$dateAdd": bson.M{"startDate": "$$NOW", "unit": "millisecond", "amount": spacing.Milliseconds()}}})}
	// Mongo forbids $expr in an upsert predicate. Initialize by _id only,
	// then reserve with a server-time conditional write without upsert.
	_, err := s.db.Collection(budgets).UpdateOne(ctx, bson.M{"_id": key}, bson.M{"$setOnInsert": bson.M{"next": time.Unix(0, 0)}}, options.UpdateOne().SetUpsert(true))
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return 0, err
	}
	result, err := s.db.Collection(budgets).UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	if result.MatchedCount == 1 {
		return 0, nil
	}
	var row struct {
		Next time.Time `bson:"next"`
	}
	if err := s.db.Collection(budgets).FindOne(ctx, bson.M{"_id": key}).Decode(&row); err != nil {
		return 0, err
	}
	now, err := s.Now(ctx)
	return max(row.Next.Sub(now), time.Millisecond), err
}
func (s *Store) DeferBudget(ctx context.Context, key string, delay time.Duration) error {
	until := bson.M{"$dateAdd": bson.M{"startDate": "$$NOW", "unit": "millisecond", "amount": max(delay.Milliseconds(), int64(0))}}
	_, err := s.db.Collection(budgets).UpdateOne(ctx, bson.M{"_id": key}, mongo.Pipeline{stage("$set", bson.M{"next": bson.M{"$max": bson.A{"$next", until}}})}, options.UpdateOne().SetUpsert(true))
	return err
}
