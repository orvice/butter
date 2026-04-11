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

const (
	collectionName     = "cron_executions"
	jobsCollectionName = "cron_jobs"
)

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

// --- CronJob persistence ---

// cronJobDoc is the MongoDB document for a cron job config.
type cronJobDoc struct {
	Name        string            `bson:"_id"`
	Schedule    string            `bson:"schedule"`
	AgentName   string            `bson:"agent_name"`
	Input       string            `bson:"input"`
	Timezone    string            `bson:"timezone"`
	Enabled     bool              `bson:"enabled"`
	DeliveryType int32            `bson:"delivery_type"`
	WebhookURL  string            `bson:"webhook_url"`
	ChannelName string            `bson:"channel_name"`
	ChatID      string            `bson:"chat_id"`
	Metadata    map[string]string `bson:"metadata,omitempty"`
}

func jobDocFromProto(j *agentsv1.CronJob) *cronJobDoc {
	doc := &cronJobDoc{
		Name:      j.GetName(),
		Schedule:  j.GetSchedule(),
		AgentName: j.GetAgentName(),
		Input:     j.GetInput(),
		Timezone:  j.GetTimezone(),
		Enabled:   j.GetEnabled(),
		Metadata:  j.GetMetadata(),
	}
	if d := j.GetDelivery(); d != nil {
		doc.DeliveryType = int32(d.GetType())
		doc.WebhookURL = d.GetWebhookUrl()
		doc.ChannelName = d.GetChannelName()
		doc.ChatID = d.GetChatId()
	}
	return doc
}

func jobDocToProto(d *cronJobDoc) *agentsv1.CronJob {
	job := &agentsv1.CronJob{
		Name:      d.Name,
		Schedule:  d.Schedule,
		AgentName: d.AgentName,
		Input:     d.Input,
		Timezone:  d.Timezone,
		Enabled:   d.Enabled,
		Metadata:  d.Metadata,
	}
	if d.DeliveryType != 0 || d.WebhookURL != "" || d.ChannelName != "" {
		job.Delivery = &agentsv1.CronDelivery{
			Type:        agentsv1.CronDeliveryType(d.DeliveryType),
			WebhookUrl:  d.WebhookURL,
			ChannelName: d.ChannelName,
			ChatId:      d.ChatID,
		}
	}
	return job
}

// MongoJobRepo implements JobRepo backed by MongoDB.
type MongoJobRepo struct {
	coll *mongo.Collection
}

// NewMongoJobRepo creates a new MongoDB-backed job repo.
func NewMongoJobRepo(db *mongo.Database) *MongoJobRepo {
	return &MongoJobRepo{coll: db.Collection(jobsCollectionName)}
}

func (r *MongoJobRepo) List(ctx context.Context) ([]*agentsv1.CronJob, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list cron jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []cronJobDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode cron jobs: %w", err)
	}

	jobs := make([]*agentsv1.CronJob, len(docs))
	for i := range docs {
		jobs[i] = jobDocToProto(&docs[i])
	}
	return jobs, nil
}

func (r *MongoJobRepo) Get(ctx context.Context, name string) (*agentsv1.CronJob, error) {
	var doc cronJobDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": name}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("cron job %q not found", name)
		}
		return nil, fmt.Errorf("get cron job %q: %w", name, err)
	}
	return jobDocToProto(&doc), nil
}

func (r *MongoJobRepo) Create(ctx context.Context, job *agentsv1.CronJob) error {
	doc := jobDocFromProto(job)
	_, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("cron job %q already exists", job.GetName())
		}
		return fmt.Errorf("create cron job: %w", err)
	}
	return nil
}

func (r *MongoJobRepo) Update(ctx context.Context, job *agentsv1.CronJob) error {
	doc := jobDocFromProto(job)
	result, err := r.coll.ReplaceOne(ctx, bson.M{"_id": job.GetName()}, doc)
	if err != nil {
		return fmt.Errorf("update cron job: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("cron job %q not found", job.GetName())
	}
	return nil
}

func (r *MongoJobRepo) Delete(ctx context.Context, name string) error {
	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": name})
	if err != nil {
		return fmt.Errorf("delete cron job: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("cron job %q not found", name)
	}
	return nil
}
