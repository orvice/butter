package planprogressmongo

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.orx.me/apps/butter/internal/runtime/plane"
	"time"
)

type doc struct {
	Key                                                               string
	WorkspaceID, AppID, InstallationID, RunID, ActivityID, WorkflowID string
	Sequence                                                          int64
	State                                                             string
	Error                                                             string
	UpdatedAt                                                         time.Time
}
type Store struct{ c *mongo.Collection }

func New(db *mongo.Database) *Store { return &Store{c: db.Collection("plane_progress")} }
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, e := s.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "app_id", Value: 1}, {Key: "updated_at", Value: -1}}}, {Keys: bson.D{{Key: "key", Value: 1}}, Options: nil}})
	return e
}
func (s *Store) EmitProgress(ctx context.Context, e *plane.Event, p plane.ProgressEvent) error {
	if e == nil {
		return mongo.ErrNilDocument
	}
	u := bson.M{"$setOnInsert": bson.M{"key": p.Key, "workspace_id": e.WorkspaceID, "app_id": e.AppID, "installation_id": e.InstallationID, "run_id": e.AgentRunID, "activity_id": e.ActivityID}, "$set": bson.M{"sequence": p.Sequence, "state": string(p.State), "error": p.Error, "updated_at": time.Now().UTC()}}
	_, err := s.c.UpdateOne(ctx, bson.M{"key": p.Key, "workspace_id": e.WorkspaceID, "sequence": bson.M{"$lt": p.Sequence}}, u, options.UpdateOne().SetUpsert(true))
	return err
}
