package mongostore

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"regexp"
)

func (s *Store) Facets(ctx context.Context, w string, q ws.Query, dimension, search, after string, limit int) ([]*v1.WorkStationFilterOption, error) {
	p, err := s.queryPipeline(ctx, w, q)
	if err != nil {
		return nil, err
	}
	field := ""
	name := ""
	switch dimension {
	case "provider":
		field = "provider"
	case "scope":
		field = "scope_id"
		name = "scope_name"
	case "project":
		field = "project_id"
		name = "project_name"
	case "status":
		field = "status"
	case "assignee":
		field = "assignees"
		p = append(p, stage("$unwind", "$chosen.assignees"))
	default:
		return nil, ws.ErrInvalid
	}
	if name == "" {
		name = field
	}
	p = append(p, stage("$project", bson.M{"id": "$chosen." + field, "name": "$chosen." + name}))
	match := bson.M{"id": bson.M{"$gt": after}}
	if search != "" {
		rx := bson.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		match["$or"] = bson.A{bson.M{"id": rx}, bson.M{"name": rx}}
	}
	p = append(p, stage("$match", match), stage("$group", bson.M{"_id": "$id", "name": bson.M{"$min": "$name"}}), stage("$sort", bson.D{{Key: "_id", Value: 1}}), stage("$limit", int64(limit)))
	cur, err := s.observations.Aggregate(ctx, p, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []*v1.WorkStationFilterOption{}
	for cur.Next(ctx) {
		var row struct {
			ID   string `bson:"_id"`
			Name string `bson:"name"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		out = append(out, &v1.WorkStationFilterOption{Id: row.ID, Name: row.Name})
	}
	return out, cur.Err()
}
