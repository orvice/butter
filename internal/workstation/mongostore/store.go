// Package mongostore keeps immutable source observations separate from the
// connection's atomic configuration/claim/snapshot-pointer boundary.
package mongostore

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

const connections = "work_station_connections"
const observations = "work_station_observations"

type Store struct {
	db                        *mongo.Database
	connections, observations *mongo.Collection
}

func New(db *mongo.Database) *Store {
	return &Store{db: db, connections: db.Collection(connections), observations: db.Collection(observations)}
}

var _ ws.Repository = (*Store)(nil)

func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.connections.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "paused", Value: 1}, {Key: "scopes.due", Value: 1}}},
		{Keys: bson.D{{Key: "host_id", Value: 1}}},
	})
	if err != nil {
		return err
	}
	_, err = s.observations.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "connection_id", Value: 1}, {Key: "scope_id", Value: 1}, {Key: "snapshot", Value: 1}, {Key: "item_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "item_id", Value: 1}, {Key: "updated", Value: -1}}},
		{Keys: bson.D{{Key: "observed", Value: 1}}},
	})
	return err
}
func (s *Store) Now(ctx context.Context) (time.Time, error) {
	var r struct {
		LocalTime time.Time `bson:"localTime"`
	}
	err := s.db.RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&r)
	return r.LocalTime, err
}
func (s *Store) list(ctx context.Context, filter bson.M) ([]ws.Record, error) {
	cur, err := s.connections.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []ws.Record{}
	err = cur.All(ctx, &out)
	return out, err
}
func (s *Store) List(ctx context.Context, w string) ([]ws.Record, error) {
	return s.list(ctx, bson.M{"workspace_id": w})
}
func (s *Store) Due(ctx context.Context) ([]ws.Record, error) {
	now, err := s.Now(ctx)
	if err != nil {
		return nil, err
	}
	cur, err := s.connections.Find(ctx, bson.M{"paused": false, "credential.ciphertext": bson.M{"$ne": ""}, "scopes": bson.M{"$elemMatch": bson.M{"hidden": false, "due": bson.M{"$lte": now}, "lease_until": bson.M{"$lte": now}}}}, options.Find().SetLimit(100))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []ws.Record
	err = cur.All(ctx, &out)
	return out, err
}
func (s *Store) Get(ctx context.Context, w, id string) (ws.Record, error) {
	var r ws.Record
	err := s.connections.FindOne(ctx, bson.M{"workspace_id": w, "id": id}).Decode(&r)
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = ws.ErrNotFound
	}
	return r, err
}
func (s *Store) Create(ctx context.Context, r ws.Record) error {
	_, err := s.connections.InsertOne(ctx, r)
	if mongo.IsDuplicateKeyError(err) {
		return ws.ErrConflict
	}
	return err
}
func (s *Store) CAS(ctx context.Context, before ws.Record, after *ws.Record, f *ws.Fence) (bool, error) {
	filter := bson.M{"workspace_id": before.WorkspaceID, "id": before.ID, "version": before.Version}
	if f != nil {
		filter["$expr"] = bson.M{"$lt": bson.A{"$$NOW", f.Until}}
		filter["scopes.claim"] = f.Token
	}
	if after == nil {
		r, err := s.connections.DeleteOne(ctx, filter)
		if err != nil {
			return false, err
		}
		return r.DeletedCount == 1, nil
	}
	r, err := s.connections.ReplaceOne(ctx, filter, *after)
	if err != nil {
		return false, err
	}
	return r.MatchedCount == 1, nil
}

type observation struct {
	ID           string     `bson:"_id"`
	WorkspaceID  string     `bson:"workspace_id"`
	ConnectionID string     `bson:"connection_id"`
	ScopeID      string     `bson:"scope_id"`
	Snapshot     string     `bson:"snapshot"`
	ItemID       string     `bson:"item_id"`
	Updated      time.Time  `bson:"updated"`
	Observed     time.Time  `bson:"observed"`
	Ended        *time.Time `bson:"ended"`
	Terminal     bool       `bson:"terminal"`
	Kind         string     `bson:"kind"`
	Status       string     `bson:"status"`
	Assignees    []string   `bson:"assignees"`
	Title        string     `bson:"title"`
	Identifier   string     `bson:"identifier"`
	Payload      []byte     `bson:"payload"`
}

func (s *Store) Stage(ctx context.Context, obs []ws.Observation) error {
	for start := 0; start < len(obs); start += 100 {
		docs := []any{}
		for _, o := range obs[start:min(start+100, len(obs))] {
			b, err := proto.Marshal(o.Item)
			if err != nil {
				return err
			}
			var ended *time.Time
			if o.Item.EndedAt != nil {
				v := o.Item.EndedAt.AsTime()
				ended = &v
			}
			docs = append(docs, observation{ID: uuid.NewString(), WorkspaceID: o.WorkspaceID, ConnectionID: o.ConnectionID, ScopeID: o.ScopeID, Snapshot: o.Snapshot, ItemID: o.Item.Id, Updated: o.Item.SourceUpdatedAt.AsTime(), Observed: o.Item.ObservedAt.AsTime(), Ended: ended, Terminal: o.Item.Terminal, Kind: o.Item.Kind, Status: o.Item.Status, Assignees: o.Item.Assignees, Title: o.Item.Title, Identifier: o.Item.Identifier, Payload: b})
		}
		if _, err := s.observations.InsertMany(ctx, docs); err != nil {
			return err
		}
	}
	return nil
}
func stage(k string, v any) bson.D { return bson.D{{Key: k, Value: v}} }

// currentSource joins the live connection and selected snapshot at every read.
// This is also the GC predicate; a stale index can never authorize a body.
func currentSource() bson.D {
	return stage("$lookup", bson.M{"from": connections, "let": bson.M{"w": "$workspace_id", "c": "$connection_id", "s": "$scope_id", "snap": "$snapshot"}, "pipeline": bson.A{
		stage("$match", bson.M{"paused": false, "credential.ciphertext": bson.M{"$ne": ""}, "$expr": bson.M{"$and": bson.A{bson.M{"$eq": bson.A{"$workspace_id", "$$w"}}, bson.M{"$eq": bson.A{"$id", "$$c"}}}}}),
		stage("$unwind", "$scopes"),
		stage("$match", bson.M{"scopes.hidden": false, "scopes.snapshot": bson.M{"$ne": ""}, "$expr": bson.M{"$and": bson.A{bson.M{"$eq": bson.A{"$scopes.id", "$$s"}}, bson.M{"$eq": bson.A{"$scopes.snapshot", "$$snap"}}}}}),
		stage("$project", bson.M{"_id": 1}),
	}, "as": "live"})
}
func (s *Store) Query(ctx context.Context, w string, q ws.Query) ([]*v1.WorkStationItem, error) {
	match := bson.M{"workspace_id": w}
	if q.AllowedConnections != nil {
		match["connection_id"] = bson.M{"$in": q.AllowedConnections}
	}
	if q.ConnectionID != "" {
		match["$and"] = bson.A{bson.M{"connection_id": q.ConnectionID}}
	}
	if q.ScopeID != "" {
		match["scope_id"] = q.ScopeID
	}
	if q.ID != "" {
		match["item_id"] = q.ID
	}
	p := mongo.Pipeline{stage("$match", match), currentSource(), stage("$match", bson.M{"live.0": bson.M{"$exists": true}}),
		stage("$sort", bson.D{{Key: "updated", Value: -1}, {Key: "observed", Value: -1}, {Key: "connection_id", Value: -1}, {Key: "scope_id", Value: -1}}),
		stage("$group", bson.M{"_id": "$item_id", "chosen": bson.M{"$first": "$$ROOT"}, "sources": bson.M{"$push": bson.M{"connection_id": "$connection_id", "scope_id": "$scope_id"}}}),
	}
	filters := bson.A{bson.M{"$or": bson.A{bson.M{"chosen.terminal": false}, bson.M{"chosen.ended": bson.M{"$gte": q.Now.Add(-90 * 24 * time.Hour)}}}}}
	if q.Kind != "" {
		filters = append(filters, bson.M{"chosen.kind": q.Kind})
	}
	if q.Status == "open" {
		filters = append(filters, bson.M{"chosen.terminal": false})
	} else if q.Status != "" && q.Status != "all" {
		filters = append(filters, bson.M{"chosen.status": q.Status})
	}
	if q.Assignee != "" {
		filters = append(filters, bson.M{"chosen.assignees": q.Assignee})
	}
	if q.Keyword != "" {
		rx := bson.Regex{Pattern: regexp.QuoteMeta(q.Keyword), Options: "i"}
		filters = append(filters, bson.M{"$or": bson.A{bson.M{"chosen.title": rx}, bson.M{"chosen.identifier": rx}}})
	}
	if q.AfterID != "" {
		filters = append(filters, bson.M{"$or": bson.A{bson.M{"chosen.updated": bson.M{"$lt": q.AfterTime}}, bson.M{"chosen.updated": q.AfterTime, "_id": bson.M{"$lt": q.AfterID}}}})
	}
	p = append(p, stage("$match", bson.M{"$and": filters}), stage("$sort", bson.D{{Key: "chosen.updated", Value: -1}, {Key: "_id", Value: -1}}), stage("$limit", int64(q.Limit)))
	cur, err := s.observations.Aggregate(ctx, p, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []*v1.WorkStationItem{}
	for cur.Next(ctx) {
		var row struct {
			Chosen  observation `bson:"chosen"`
			Sources []struct {
				ConnectionID string `bson:"connection_id"`
				ScopeID      string `bson:"scope_id"`
			} `bson:"sources"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		item := &v1.WorkStationItem{}
		if err := proto.Unmarshal(row.Chosen.Payload, item); err != nil {
			return nil, err
		}
		item.Sources = nil
		for _, src := range row.Sources {
			item.Sources = append(item.Sources, &v1.WorkStationSource{ConnectionId: src.ConnectionID, ScopeId: src.ScopeID})
		}
		out = append(out, item)
	}
	return out, cur.Err()
}
func (s *Store) HostInUse(ctx context.Context, id string) (bool, error) {
	err := s.connections.FindOne(ctx, bson.M{"host_id": id}, options.FindOne().SetProjection(bson.M{"_id": 1})).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return err == nil, err
}
func (s *Store) Collect(ctx context.Context) error {
	now, err := s.Now(ctx)
	if err != nil {
		return err
	}
	p := mongo.Pipeline{stage("$match", bson.M{"observed": bson.M{"$lt": now.Add(-24 * time.Hour)}}), currentSource(), stage("$match", bson.M{"live.0": bson.M{"$exists": false}}), stage("$limit", int64(1000)), stage("$project", bson.M{"_id": 1})}
	cur, err := s.observations.Aggregate(ctx, p)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	ids := bson.A{}
	for cur.Next(ctx) {
		var r struct {
			ID string `bson:"_id"`
		}
		if err := cur.Decode(&r); err != nil {
			return err
		}
		ids = append(ids, r.ID)
	}
	if err := cur.Err(); err != nil {
		return err
	}
	if len(ids) > 0 {
		_, err = s.observations.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
	}
	return err
}
