package mongostore

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	ws "go.orx.me/apps/butter/internal/workstation"
)

const hosts = "work_station_hosts"

func (s *Store) ListHosts(ctx context.Context) ([]ws.Host, error) {
	cur, err := s.db.Collection(hosts).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []ws.Host{ws.CloudHost()}
	var stored []ws.Host
	err = cur.All(ctx, &stored)
	return append(out, stored...), err
}
func (s *Store) GetHost(ctx context.Context, id string) (ws.Host, error) {
	if id == ws.PlaneCloudHost {
		return ws.CloudHost(), nil
	}
	var h ws.Host
	err := s.db.Collection(hosts).FindOne(ctx, bson.M{"id": id}).Decode(&h)
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = ws.ErrNotFound
	}
	return h, err
}
func (s *Store) SaveHost(ctx context.Context, h ws.Host, revision int64) error {
	if h.ID == ws.PlaneCloudHost {
		return ws.ErrInvalid
	}
	if revision == 0 {
		_, err := s.db.Collection(hosts).InsertOne(ctx, h)
		if mongo.IsDuplicateKeyError(err) {
			return ws.ErrConflict
		}
		return err
	}
	r, err := s.db.Collection(hosts).ReplaceOne(ctx, bson.M{"id": h.ID, "revision": revision}, h)
	if err != nil {
		return err
	}
	if r.MatchedCount == 0 {
		return ws.ErrConflict
	}
	return nil
}
func (s *Store) DeleteHost(ctx context.Context, id string, revision int64) error {
	if id == ws.PlaneCloudHost {
		return ws.ErrInvalid
	}
	used, err := s.HostInUse(ctx, id)
	if err != nil {
		return err
	}
	if used {
		return ws.ErrConflict
	}
	r, err := s.db.Collection(hosts).DeleteOne(ctx, bson.M{"id": id, "revision": revision})
	if err != nil {
		return err
	}
	if r.DeletedCount == 0 {
		_, err = s.GetHost(ctx, id)
		if errors.Is(err, ws.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		return ws.ErrConflict
	}
	return nil
}
