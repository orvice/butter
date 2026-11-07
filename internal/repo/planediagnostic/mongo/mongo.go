package mongodiagnostic

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
	"sort"
	"time"
)

type doc struct {
	Key, WorkspaceID, AppID, InstallationID, RunID, ActivityID, WorkflowID string
	Class                                                                  string
	Summary                                                                string
	Recoverable                                                            bool
	CreatedAt                                                              time.Time
}
type Store struct{ c *mongo.Collection }

func New(db *mongo.Database) *Store { return &Store{c: db.Collection("plane_diagnostics")} }
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, e := s.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "app_id", Value: 1}, {Key: "created_at", Value: -1}}}, {Keys: bson.D{{Key: "key", Value: 1}}, Options: nil}})
	return e
}
func toDoc(r *plane.DiagnosticRecord) doc {
	return doc{r.Key, r.WorkspaceID, r.AppID, r.InstallationID, r.RunID, r.ActivityID, r.WorkflowID, string(r.Class), r.Summary, r.Recoverable, r.CreatedAt}
}
func fromDoc(d doc) *plane.DiagnosticRecord {
	return &plane.DiagnosticRecord{Key: d.Key, WorkspaceID: d.WorkspaceID, AppID: d.AppID, InstallationID: d.InstallationID, RunID: d.RunID, ActivityID: d.ActivityID, WorkflowID: d.WorkflowID, Class: plane.FailureClass(d.Class), Summary: d.Summary, Recoverable: d.Recoverable, CreatedAt: d.CreatedAt}
}
func (s *Store) Put(ctx context.Context, r *plane.DiagnosticRecord) error {
	if r == nil {
		return mongo.ErrNilDocument
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	_, e := s.c.ReplaceOne(ctx, bson.M{"key": r.Key, "workspace_id": r.WorkspaceID}, toDoc(r), options.Replace().SetUpsert(true))
	return e
}
func (s *Store) Get(ctx context.Context, ws, k string) (*plane.DiagnosticRecord, error) {
	var d doc
	if e := s.c.FindOne(ctx, bson.M{"key": k, "workspace_id": ws}).Decode(&d); e != nil {
		return nil, e
	}
	return fromDoc(d), nil
}
func (s *Store) List(ctx context.Context, ws, app string) ([]*plane.DiagnosticRecord, error) {
	q := bson.M{"workspace_id": ws}
	if app != "" {
		q["app_id"] = app
	}
	cur, e := s.c.Find(ctx, q)
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var ds []doc
	if e = cur.All(ctx, &ds); e != nil {
		return nil, e
	}
	sort.SliceStable(ds, func(i, j int) bool { return ds[i].CreatedAt.After(ds[j].CreatedAt) })
	out := make([]*plane.DiagnosticRecord, 0, len(ds))
	for _, d := range ds {
		out = append(out, fromDoc(d))
	}
	return out, nil
}
