package mongodiagnostic

import (
	"cmp"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
)

type doc struct {
	Key            string    `bson:"key"`
	WorkspaceID    string    `bson:"workspace_id"`
	AppID          string    `bson:"app_id"`
	InstallationID string    `bson:"installation_id"`
	RunID          string    `bson:"run_id"`
	ActivityID     string    `bson:"activity_id"`
	WorkflowID     string    `bson:"workflow_id"`
	Class          string    `bson:"class"`
	Summary        string    `bson:"summary"`
	Recoverable    bool      `bson:"recoverable"`
	CreatedAt      time.Time `bson:"created_at"`
	ExpiresAt      time.Time `bson:"expires_at,omitempty"`
	Legacy         legacyDoc `bson:",inline"`
}

// Previous untagged documents used lowercased Go names. Read them without
// rewriting IDs or weakening canonical-field precedence in scoped queries.
type legacyDoc struct {
	WorkspaceID    string    `bson:"workspaceid,omitempty"`
	AppID          string    `bson:"appid,omitempty"`
	InstallationID string    `bson:"installationid,omitempty"`
	RunID          string    `bson:"runid,omitempty"`
	ActivityID     string    `bson:"activityid,omitempty"`
	WorkflowID     string    `bson:"workflowid,omitempty"`
	CreatedAt      time.Time `bson:"createdat,omitempty"`
}

func scopedField(canonical, legacy, value string) bson.M {
	return bson.M{"$or": bson.A{
		bson.M{canonical: value},
		bson.M{canonical: bson.M{"$exists": false}, legacy: value},
	}}
}

type Store struct{ c *mongo.Collection }

func New(db *mongo.Database) *Store { return &Store{c: db.Collection("plane_diagnostics")} }
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, e := s.c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "app_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "key", Value: 1}}},
		// Only new webhook receipts have expires_at. Legacy execution history
		// is not assigned a TTL or deleted as part of this change.
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
	})
	return e
}
func toDoc(r *plane.DiagnosticRecord) doc {
	return doc{Key: r.Key, WorkspaceID: r.WorkspaceID, AppID: r.AppID, InstallationID: r.InstallationID,
		RunID: r.RunID, ActivityID: r.ActivityID, WorkflowID: r.WorkflowID, Class: string(r.Class),
		Summary: r.Summary, Recoverable: r.Recoverable, CreatedAt: r.CreatedAt, ExpiresAt: r.ExpiresAt}
}
func fromDoc(d doc) *plane.DiagnosticRecord {
	created := d.CreatedAt
	if created.IsZero() {
		created = d.Legacy.CreatedAt
	}
	return &plane.DiagnosticRecord{Key: d.Key, WorkspaceID: cmp.Or(d.WorkspaceID, d.Legacy.WorkspaceID),
		AppID: cmp.Or(d.AppID, d.Legacy.AppID), InstallationID: cmp.Or(d.InstallationID, d.Legacy.InstallationID),
		RunID: cmp.Or(d.RunID, d.Legacy.RunID), ActivityID: cmp.Or(d.ActivityID, d.Legacy.ActivityID),
		WorkflowID: cmp.Or(d.WorkflowID, d.Legacy.WorkflowID), Class: plane.FailureClass(d.Class),
		Summary: d.Summary, Recoverable: d.Recoverable, CreatedAt: created, ExpiresAt: d.ExpiresAt}
}
func (s *Store) Put(ctx context.Context, r *plane.DiagnosticRecord) error {
	if r == nil {
		return mongo.ErrNilDocument
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	query := scopedField("workspace_id", "workspaceid", r.WorkspaceID)
	query["key"] = r.Key
	_, e := s.c.ReplaceOne(ctx, query, toDoc(r), options.Replace().SetUpsert(true))
	return e
}
func (s *Store) Get(ctx context.Context, ws, k string) (*plane.DiagnosticRecord, error) {
	var d doc
	query := scopedField("workspace_id", "workspaceid", ws)
	query["key"] = k
	if e := s.c.FindOne(ctx, query).Decode(&d); e != nil {
		return nil, e
	}
	return fromDoc(d), nil
}
func (s *Store) List(ctx context.Context, ws, app string) ([]*plane.DiagnosticRecord, error) {
	q := scopedField("workspace_id", "workspaceid", ws)
	if app != "" {
		q = bson.M{"$and": bson.A{q, scopedField("app_id", "appid", app)}}
	}
	// Sort/limit in Mongo, including legacy timestamps, so an incoming flood
	// cannot make every history read allocate the entire receipt collection.
	cur, e := s.c.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: q}},
		bson.D{{Key: "$set", Value: bson.M{"_diagnostic_created": bson.M{"$ifNull": bson.A{"$created_at", "$createdat"}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_diagnostic_created", Value: -1}, {Key: "key", Value: 1}}}},
		bson.D{{Key: "$limit", Value: plane.DiagnosticHistoryLimit}},
	})
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var ds []doc
	if e = cur.All(ctx, &ds); e != nil {
		return nil, e
	}
	out := make([]*plane.DiagnosticRecord, 0, len(ds))
	for _, d := range ds {
		out = append(out, fromDoc(d))
	}
	return out, nil
}
