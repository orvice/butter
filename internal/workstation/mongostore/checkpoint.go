package mongostore

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	ws "go.orx.me/apps/butter/internal/workstation"
)

type checkpointDocument struct {
	ID         string    `bson:"_id"`
	Revision   uint64    `bson:"revision"`
	Generation string    `bson:"generation"`
	Sequence   uint64    `bson:"sequence"`
	Cursor     string    `bson:"cursor"`
	ETag       string    `bson:"etag"`
	UpdatedAt  time.Time `bson:"updated_at"`
}

func checkpointFromDocument(d checkpointDocument) ws.Checkpoint {
	return ws.Checkpoint{Revision: d.Revision, Generation: d.Generation, Sequence: d.Sequence, Cursor: d.Cursor, ETag: d.ETag, UpdatedAt: d.UpdatedAt}
}

func checkpointDocumentFrom(key string, c ws.Checkpoint) checkpointDocument {
	return checkpointDocument{ID: key, Revision: c.Revision, Generation: c.Generation, Sequence: c.Sequence, Cursor: c.Cursor, ETag: c.ETag, UpdatedAt: c.UpdatedAt}
}

func (s *Store) Load(ctx context.Context, key string) (ws.Checkpoint, bool, error) {
	var doc checkpointDocument
	err := s.checkpoints.FindOne(ctx, bson.M{"_id": key}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ws.Checkpoint{}, false, nil
	}
	if err != nil {
		return ws.Checkpoint{}, false, err
	}
	return checkpointFromDocument(doc), true, nil
}

func (s *Store) Delete(ctx context.Context, key string, expected uint64) error {
	result, err := s.checkpoints.DeleteOne(ctx, bson.M{"_id": key, "revision": expected})
	if err != nil {
		return err
	}
	if result.DeletedCount == 1 {
		return nil
	}
	_, exists, err := s.Load(ctx, key)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return ws.ErrCheckpointConflict
}
func (s *Store) CompareAndSwap(ctx context.Context, key string, expected uint64, next ws.Checkpoint) (ws.Checkpoint, error) {
	current, exists, err := s.Load(ctx, key)
	if err != nil {
		return ws.Checkpoint{}, err
	}
	if (!exists && expected != 0) || (exists && current.Revision != expected) {
		return current, ws.ErrCheckpointConflict
	}
	if exists && next.Generation == current.Generation && next.Sequence < current.Sequence {
		return current, ws.ErrCheckpointRegression
	}
	next.Revision = expected + 1
	doc := checkpointDocumentFrom(key, next)
	if !exists {
		_, err = s.checkpoints.InsertOne(ctx, doc)
		if mongo.IsDuplicateKeyError(err) {
			return current, ws.ErrCheckpointConflict
		}
		return next, err
	}
	result, err := s.checkpoints.ReplaceOne(ctx, bson.M{"_id": key, "revision": expected}, doc, options.Replace())
	if err != nil {
		return current, err
	}
	if result.MatchedCount != 1 {
		latest, _, loadErr := s.Load(ctx, key)
		if loadErr != nil {
			return current, loadErr
		}
		return latest, ws.ErrCheckpointConflict
	}
	return next, nil
}
