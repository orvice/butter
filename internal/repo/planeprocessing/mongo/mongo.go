package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Store struct{ col *mongo.Collection }
type doc struct {
	ID                 string                          `bson:"_id"`
	WorkspaceID        string                          `bson:"workspace_id"`
	AppID              string                          `bson:"app_id"`
	Key                string                          `bson:"key"`
	Spec               string                          `bson:"spec"`
	Status             *agentsv1.PlaneProcessingStatus `bson:"status,omitempty"`
	Failure            string                          `bson:"failure_reason,omitempty"`
	ReceivedAt         time.Time                       `bson:"received_at"`
	UpdatedAt          time.Time                       `bson:"updated_at"`
	ProjectionRevision int64                           `bson:"projection_revision,omitempty"`
	ClaimOwner         string                          `bson:"claim_owner,omitempty"`
	ClaimRevision      int64                           `bson:"claim_revision,omitempty"`
	ClaimFence         uint64                          `bson:"claim_fence,omitempty"`
	ClaimExpiresAt     time.Time                       `bson:"claim_expires_at,omitempty"`
}

func New(db *mongo.Database) *Store { return &Store{col: db.Collection("plane_processing")} }
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.col.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "key", Value: 1}}, Options: options.Index().SetUnique(true).SetName("by_business_key")}, {Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "received_at", Value: -1}}, Options: options.Index().SetName("by_workspace_received")}})
	return err
}
func encode(r *agentsv1.PlaneProcessingRecord) (doc, error) {
	b, e := protojson.Marshal(r)
	status := r.GetStatus()
	return doc{ID: r.GetId(), WorkspaceID: r.GetWorkspaceId(), AppID: r.GetAppId(), Key: planeprocessing.BusinessKey(r), Spec: string(b), Status: &status, Failure: r.GetFailureReason(), ReceivedAt: r.GetReceivedAt().AsTime(), UpdatedAt: r.GetUpdatedAt().AsTime()}, e
}
func decode(d doc) (*agentsv1.PlaneProcessingRecord, error) {
	r := &agentsv1.PlaneProcessingRecord{}
	if e := protojson.Unmarshal([]byte(d.Spec), r); e != nil {
		return nil, e
	}
	r.Id = d.ID
	if d.Status != nil {
		r.Status = *d.Status
		r.FailureReason = d.Failure
	}
	if !d.UpdatedAt.IsZero() {
		r.UpdatedAt = timestamppb.New(d.UpdatedAt)
	}
	return r, nil
}
func (s *Store) Create(ctx context.Context, r *agentsv1.PlaneProcessingRecord) (*agentsv1.PlaneProcessingRecord, error) {
	r = planeprocessing.Normalize(r)
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
	// Run lease documents share this collection but are not processing
	// records and may have Mongo ObjectIDs rather than public string IDs.
	q := bson.M{"spec": bson.M{"$exists": true}}
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

func (s *Store) UpdateStatus(ctx context.Context, ws, app, installation, run, activity, action string, status agentsv1.PlaneProcessingStatus, failure string) error {
	r := &agentsv1.PlaneProcessingRecord{WorkspaceId: ws, AppId: app, InstallationId: installation, RunId: run, ActivityId: activity, EventAction: action}
	var d doc
	if err := s.col.FindOne(ctx, projectionQuery(r)).Decode(&d); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return planeprocessing.ErrNotFound
		}
		return err
	}
	current, err := decode(d)
	if err != nil {
		return err
	}
	current.Status, current.FailureReason = status, failure
	return s.Project(ctx, current)
}

func projectionQuery(r *agentsv1.PlaneProcessingRecord) bson.M {
	return bson.M{"workspace_id": r.GetWorkspaceId(), "app_id": r.GetAppId(), "spec": bson.M{"$exists": true}, "key": bson.M{"$in": []string{planeprocessing.BusinessKey(r), planeprocessing.LegacyKey(r)}}}
}

func (s *Store) Project(ctx context.Context, r *agentsv1.PlaneProcessingRecord) error {
	for range 16 {
		var existing doc
		err := s.col.FindOne(ctx, projectionQuery(r)).Decode(&existing)
		if errors.Is(err, mongo.ErrNoDocuments) {
			if _, err = s.Create(ctx, r); errors.Is(err, planeprocessing.ErrDuplicate) {
				continue
			}
			return err
		}
		if err != nil {
			return err
		}
		old, err := decode(existing)
		if err != nil {
			return err
		}
		merged := planeprocessing.MergeProjection(old, r)
		d, err := encode(merged)
		if err != nil {
			return err
		}
		d.Failure = merged.GetFailureReason()
		filter := bson.M{"_id": existing.ID, "workspace_id": r.GetWorkspaceId()}
		if existing.ProjectionRevision == 0 {
			filter["$or"] = []bson.M{{"projection_revision": 0}, {"projection_revision": bson.M{"$exists": false}}}
		} else {
			filter["projection_revision"] = existing.ProjectionRevision
		}
		d.ProjectionRevision = existing.ProjectionRevision + 1
		res, err := s.col.ReplaceOne(ctx, filter, d)
		if err != nil {
			return err
		}
		if res.MatchedCount == 1 {
			return nil
		}
	}
	return errors.New("plane processing projection contention")
}

func (s *Store) Claim(ctx context.Context, ws, app, installation, run, owner string, expiry time.Time) (planeprocessing.Claim, error) {
	key := ws + ":" + app + ":" + installation + ":" + run
	now := time.Now()
	f := bson.M{"key": key, "workspace_id": ws, "$or": []bson.M{
		{"claim_expires_at": bson.M{"$lte": now}},
		{"claim_expires_at": bson.M{"$exists": false}},
	}}
	u := bson.M{"$set": bson.M{"claim_owner": owner, "claim_expires_at": expiry}, "$inc": bson.M{"claim_revision": 1, "claim_fence": 1}, "$setOnInsert": bson.M{"workspace_id": ws, "app_id": app, "key": key}}
	// Return the claim from the same atomic operation that acquired it.
	// A separate read could observe a later holder after lease expiry.
	// Claim-only documents may have an ObjectID _id; decode only claim fields.
	var d struct {
		ClaimOwner     string    `bson:"claim_owner"`
		ClaimRevision  int64     `bson:"claim_revision"`
		ClaimFence     uint64    `bson:"claim_fence"`
		ClaimExpiresAt time.Time `bson:"claim_expires_at"`
	}
	e := s.col.FindOneAndUpdate(ctx, f, u, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&d)
	if mongo.IsDuplicateKeyError(e) {
		return planeprocessing.Claim{}, planeprocessing.ErrClaimConflict
	}
	if e != nil {
		return planeprocessing.Claim{}, e
	}
	return planeprocessing.Claim{Owner: d.ClaimOwner, Revision: d.ClaimRevision, Fence: d.ClaimFence, ExpiresAt: d.ClaimExpiresAt}, nil
}
func (s *Store) Renew(ctx context.Context, key string, c planeprocessing.Claim, expiry time.Time) (planeprocessing.Claim, error) {
	r, e := s.col.UpdateOne(ctx, bson.M{"key": key, "claim_owner": c.Owner, "claim_revision": c.Revision, "claim_fence": c.Fence, "claim_expires_at": bson.M{"$gt": time.Now()}}, bson.M{"$set": bson.M{"claim_expires_at": expiry}})
	if e != nil || r.MatchedCount != 1 {
		return planeprocessing.Claim{}, planeprocessing.ErrClaimExpired
	}
	c.ExpiresAt = expiry
	return c, nil
}
func (s *Store) Complete(ctx context.Context, key string, c planeprocessing.Claim) error {
	r, e := s.col.UpdateOne(ctx, bson.M{"key": key, "claim_owner": c.Owner, "claim_revision": c.Revision, "claim_fence": c.Fence, "claim_expires_at": bson.M{"$gt": time.Now()}}, bson.M{"$unset": bson.M{"claim_owner": "", "claim_expires_at": ""}})
	if e != nil {
		return e
	}
	if r.MatchedCount != 1 {
		return planeprocessing.ErrClaimExpired
	}
	return nil
}
func (s *Store) ReclaimExpired(ctx context.Context, ws, app, installation, run, owner string, expiry time.Time) (planeprocessing.Claim, error) {
	return s.Claim(ctx, ws, app, installation, run, owner, expiry)
}
func (s *Store) RecoverExpired(ctx context.Context, now time.Time) error {
	_, e := s.col.UpdateMany(ctx, bson.M{"claim_expires_at": bson.M{"$lte": now}}, bson.M{"$unset": bson.M{"claim_owner": "", "claim_expires_at": ""}})
	return e
}

func (s *Store) FindByEvent(ctx context.Context, r *agentsv1.PlaneProcessingRecord) (*agentsv1.PlaneProcessingRecord, error) {
	var d doc
	err := s.col.FindOne(ctx, projectionQuery(r)).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, planeprocessing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return decode(d)
}
