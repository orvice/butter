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
	Key                                                                           string
	WorkspaceID, AppID, InstallationID, RunID, ActivityID, WorkflowID, DeliveryID string
	Sequence                                                                      int64
	State                                                                         string
	Error                                                                         string
	UpdatedAt                                                                     time.Time
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
	u := bson.M{"$setOnInsert": bson.M{"key": p.Key, "workspace_id": e.WorkspaceID, "app_id": e.AppID, "installation_id": e.InstallationID, "run_id": e.AgentRunID, "activity_id": e.ActivityID, "workflow_id": p.WorkflowID, "delivery_id": p.DeliveryID}, "$set": bson.M{"sequence": p.Sequence, "state": string(p.State), "error": p.Error, "updated_at": time.Now().UTC()}}
	_, err := s.c.UpdateOne(ctx, bson.M{"key": p.Key, "workspace_id": e.WorkspaceID, "sequence": bson.M{"$lt": p.Sequence}}, u, options.UpdateOne().SetUpsert(true))
	return err
}

type Query struct {
	WorkspaceID, AppID, InstallationID, RunID, ActivityID string
	Limit                                                 int
}
type Record struct {
	Key, WorkspaceID, AppID, InstallationID, RunID, ActivityID, WorkflowID, DeliveryID, State, Error string
	Sequence                                                                                         int64
	UpdatedAt                                                                                        time.Time
}

func (s *Store) List(ctx context.Context, q Query) ([]Record, error) {
	filter := bson.M{"workspace_id": q.WorkspaceID}
	if q.AppID != "" {
		filter["app_id"] = q.AppID
	}
	if q.InstallationID != "" {
		filter["installation_id"] = q.InstallationID
	}
	if q.RunID != "" {
		filter["run_id"] = q.RunID
	}
	if q.ActivityID != "" {
		filter["activity_id"] = q.ActivityID
	}
	find := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}, {Key: "sequence", Value: -1}})
	if q.Limit > 0 {
		find.SetLimit(int64(q.Limit))
	}
	cur, err := s.c.Find(ctx, filter, find)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var ds []doc
	if err = cur.All(ctx, &ds); err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(ds))
	for _, d := range ds {
		out = append(out, Record{Key: d.Key, WorkspaceID: d.WorkspaceID, AppID: d.AppID, InstallationID: d.InstallationID, RunID: d.RunID, ActivityID: d.ActivityID, WorkflowID: d.WorkflowID, DeliveryID: d.DeliveryID, Sequence: d.Sequence, State: d.State, Error: d.Error, UpdatedAt: d.UpdatedAt})
	}
	return out, nil
}
