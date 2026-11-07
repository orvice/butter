package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Store struct{ col *mongo.Collection }
type doc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	AppID       string `bson:"app_id"`
	Key         string `bson:"key"`
	Spec        string `bson:"spec"`
}

func New(db *mongo.Database) *Store { return &Store{col: db.Collection("plane_processing")} }
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.col.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "key", Value: 1}}, Options: options.Index().SetUnique(true).SetName("by_business_key")}, {Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "received_at", Value: -1}}, Options: options.Index().SetName("by_workspace_received")}})
	return err
}
func encode(r *agentsv1.PlaneProcessingRecord) (doc, error) {
	b, e := protojson.Marshal(r)
	return doc{ID: r.GetId(), WorkspaceID: r.GetWorkspaceId(), AppID: r.GetAppId(), Key: r.GetInstallationId() + ":" + r.GetRunId() + ":" + r.GetActivityId() + ":" + r.GetEventAction(), Spec: string(b)}, e
}
func decode(d doc) (*agentsv1.PlaneProcessingRecord, error) {
	r := &agentsv1.PlaneProcessingRecord{}
	if e := protojson.Unmarshal([]byte(d.Spec), r); e != nil {
		return nil, e
	}
	r.Id = d.ID
	return r, nil
}
func (s *Store) Create(ctx context.Context, r *agentsv1.PlaneProcessingRecord) (*agentsv1.PlaneProcessingRecord, error) {
	d, e := encode(r)
	if e != nil {
		return nil, e
	}
	_, e = s.col.InsertOne(ctx, d)
	if mongo.IsDuplicateKeyError(e) {
		return nil, planeprocessing.ErrDuplicate
	}
	if e != nil {
		return nil, e
	}
	return r, nil
}
func (s *Store) Get(ctx context.Context, ws, id string) (*agentsv1.PlaneProcessingRecord, error) {
	var d doc
	e := s.col.FindOne(ctx, bson.M{"_id": id, "workspace_id": ws}).Decode(&d)
	if e == mongo.ErrNoDocuments {
		return nil, planeprocessing.ErrNotFound
	}
	if e != nil {
		return nil, e
	}
	return decode(d)
}
func (s *Store) List(ctx context.Context, f planeprocessing.Filter) ([]*agentsv1.PlaneProcessingRecord, error) {
	q := bson.M{}
	if f.WorkspaceID != "" {
		q["workspace_id"] = f.WorkspaceID
	}
	if f.AppID != "" {
		q["app_id"] = f.AppID
	}
	if f.Status != 0 {
		q["status"] = f.Status
	}
	opts := options.Find().SetSort(bson.D{{Key: "received_at", Value: -1}})
	if f.Limit > 0 {
		opts.SetLimit(int64(f.Limit))
	}
	cur, e := s.col.Find(ctx, q, opts)
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var ds []doc
	if e = cur.All(ctx, &ds); e != nil {
		return nil, fmt.Errorf("list plane processing: %w", e)
	}
	out := make([]*agentsv1.PlaneProcessingRecord, 0, len(ds))
	for _, d := range ds {
		r, e := decode(d)
		if e != nil {
			return nil, e
		}
		out = append(out, r)
	}
	return out, nil
}

var _ = proto.Clone
