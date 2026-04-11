package cron

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const collectionName = "cron_executions"

// executionDoc is the MongoDB document for a cron execution record.
type executionDoc struct {
	ID         string `bson:"_id"`
	JobName    string `bson:"job_name"`
	AgentName  string `bson:"agent_name"`
	Status     int32  `bson:"status"`
	Input      string `bson:"input"`
	Output     string `bson:"output"`
	StartedAt  int64  `bson:"started_at"`  // unix seconds
	FinishedAt int64  `bson:"finished_at"` // unix seconds
}

func docFromProto(e *agentsv1.CronExecution) *executionDoc {
	return &executionDoc{
		ID:         e.GetId(),
		JobName:    e.GetJobName(),
		AgentName:  e.GetAgentName(),
		Status:     int32(e.GetStatus()),
		Input:      e.GetInput(),
		Output:     e.GetOutput(),
		StartedAt:  e.GetStartedAt().GetSeconds(),
		FinishedAt: e.GetFinishedAt().GetSeconds(),
	}
}

func docToProto(d *executionDoc) *agentsv1.CronExecution {
	return &agentsv1.CronExecution{
		Id:         d.ID,
		JobName:    d.JobName,
		AgentName:  d.AgentName,
		Status:     agentsv1.CronExecutionStatus(d.Status),
		Input:      d.Input,
		Output:     d.Output,
		StartedAt:  &timestamppb.Timestamp{Seconds: d.StartedAt},
		FinishedAt: &timestamppb.Timestamp{Seconds: d.FinishedAt},
	}
}

// MongoExecutionRepo implements ExecutionRepo backed by MongoDB.
type MongoExecutionRepo struct {
	coll *mongo.Collection
}

// NewMongoExecutionRepo creates a new MongoDB-backed execution repo.
func NewMongoExecutionRepo(db *mongo.Database) *MongoExecutionRepo {
	return &MongoExecutionRepo{coll: db.Collection(collectionName)}
}

func (r *MongoExecutionRepo) Save(ctx context.Context, exec *agentsv1.CronExecution) error {
	doc := docFromProto(exec)
	_, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("insert cron execution: %w", err)
	}
	return nil
}

func (r *MongoExecutionRepo) List(ctx context.Context, jobName string, pageSize int32, pageToken string) ([]*agentsv1.CronExecution, string, error) {
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := bson.M{}
	if jobName != "" {
		filter["job_name"] = jobName
	}
	if pageToken != "" {
		filter["_id"] = bson.M{"$lt": pageToken}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "started_at", Value: -1}}).
		SetLimit(int64(pageSize + 1)) // fetch one extra to determine next page token

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", fmt.Errorf("find cron executions: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []executionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, "", fmt.Errorf("decode cron executions: %w", err)
	}

	var nextPageToken string
	if len(docs) > int(pageSize) {
		nextPageToken = docs[pageSize].ID
		docs = docs[:pageSize]
	}

	results := make([]*agentsv1.CronExecution, len(docs))
	for i := range docs {
		results[i] = docToProto(&docs[i])
	}

	return results, nextPageToken, nil
}

func (r *MongoExecutionRepo) GetByID(ctx context.Context, id string) (*agentsv1.CronExecution, error) {
	var doc executionDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("find cron execution %q: %w", id, err)
	}
	return docToProto(&doc), nil
}
