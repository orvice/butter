package cron

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	collectionName     = "cron_executions"
	jobsCollectionName = "cron_jobs"
)

// executionDoc is the MongoDB document for a cron execution record.
type executionDoc struct {
	ID            string `bson:"_id"`
	WorkspaceID   string `bson:"workspace_id"`
	JobName       string `bson:"job_name"`
	AgentName     string `bson:"agent_name"`
	Status        int32  `bson:"status"`
	Input         string `bson:"input"`
	Output        string `bson:"output"`
	Error         string `bson:"error,omitempty"`
	StartedAt     int64  `bson:"started_at"`  // unix seconds
	FinishedAt    int64  `bson:"finished_at"` // unix seconds
	DurationMs    int64  `bson:"duration_ms,omitempty"`
	AttemptCount  int32  `bson:"attempt_count,omitempty"`
	TriggerType   int32  `bson:"trigger_type,omitempty"`
	SkippedReason string `bson:"skipped_reason,omitempty"`
	Truncated     bool   `bson:"truncated,omitempty"`
	// Session coordinates for answering a WAITING_INPUT execution.
	SessionAppName string `bson:"session_app_name,omitempty"`
	SessionUserID  string `bson:"session_user_id,omitempty"`
	SessionID      string `bson:"session_id,omitempty"`
}

func docFromProto(e *agentsv1.CronExecution) *executionDoc {
	return &executionDoc{
		ID:            e.GetId(),
		WorkspaceID:   e.GetWorkspaceId(),
		JobName:       e.GetJobName(),
		AgentName:     e.GetAgentName(),
		Status:        int32(e.GetStatus()),
		Input:         e.GetInput(),
		Output:        e.GetOutput(),
		Error:         e.GetError(),
		StartedAt:     e.GetStartedAt().GetSeconds(),
		FinishedAt:    e.GetFinishedAt().GetSeconds(),
		DurationMs:    e.GetDurationMs(),
		AttemptCount:  e.GetAttemptCount(),
		TriggerType:    int32(e.GetTriggerType()),
		SkippedReason:  e.GetSkippedReason(),
		Truncated:      e.GetTruncated(),
		SessionAppName: e.GetSessionAppName(),
		SessionUserID:  e.GetSessionUserId(),
		SessionID:      e.GetSessionId(),
	}
}

func docToProto(d *executionDoc) *agentsv1.CronExecution {
	exec := &agentsv1.CronExecution{
		Id:             d.ID,
		WorkspaceId:    d.WorkspaceID,
		JobName:        d.JobName,
		AgentName:      d.AgentName,
		Status:         agentsv1.CronExecutionStatus(d.Status),
		Input:          d.Input,
		Output:         d.Output,
		Error:          d.Error,
		StartedAt:      &timestamppb.Timestamp{Seconds: d.StartedAt},
		DurationMs:     d.DurationMs,
		AttemptCount:   d.AttemptCount,
		TriggerType:    agentsv1.CronExecutionTriggerType(d.TriggerType),
		SkippedReason:  d.SkippedReason,
		Truncated:      d.Truncated,
		SessionAppName: d.SessionAppName,
		SessionUserId:  d.SessionUserID,
		SessionId:      d.SessionID,
	}
	// A WAITING_INPUT execution has no finish time yet — keep it unset
	// instead of surfacing the zero value as 1970-01-01.
	if d.FinishedAt != 0 {
		exec.FinishedAt = &timestamppb.Timestamp{Seconds: d.FinishedAt}
	}
	return exec
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
	// Upsert: an execution is saved again when it leaves WAITING_INPUT for a
	// terminal state after a human answered its session.
	_, err := r.coll.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("save cron execution: %w", err)
	}
	return nil
}

func (r *MongoExecutionRepo) ListWaitingBySession(ctx context.Context, appName, userID, sessionID string) ([]*agentsv1.CronExecution, error) {
	filter := bson.M{
		"status":           int32(agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT),
		"session_app_name": appName,
		"session_user_id":  userID,
		"session_id":       sessionID,
	}
	opts := options.Find().SetSort(bson.D{{Key: "started_at", Value: 1}})
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find waiting cron executions: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []executionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode cron executions: %w", err)
	}
	results := make([]*agentsv1.CronExecution, len(docs))
	for i := range docs {
		results[i] = docToProto(&docs[i])
	}
	return results, nil
}

func (r *MongoExecutionRepo) List(ctx context.Context, workspaceID, jobName string, pageSize int32, pageToken string) ([]*agentsv1.CronExecution, string, error) {
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := bson.M{}
	if workspaceID != "" {
		filter["workspace_id"] = workspaceID
	}
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

func (r *MongoExecutionRepo) ListByTimeRange(ctx context.Context, workspaceID, jobName string, start, end time.Time) ([]*agentsv1.CronExecution, error) {
	filter := bson.M{"started_at": bson.M{"$gte": start, "$lt": end}}
	if workspaceID != "" {
		filter["workspace_id"] = workspaceID
	}
	if jobName != "" {
		filter["job_name"] = jobName
	}
	opts := options.Find().SetSort(bson.D{{Key: "started_at", Value: 1}})
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find cron executions by range: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []executionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode cron executions: %w", err)
	}
	results := make([]*agentsv1.CronExecution, len(docs))
	for i := range docs {
		results[i] = docToProto(&docs[i])
	}
	return results, nil
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

// cronJobDoc is the MongoDB document for a cron job config. _id is the
// composite "workspace_id:name".
type cronJobDoc struct {
	ID              string            `bson:"_id"`
	WorkspaceID     string            `bson:"workspace_id"`
	Name            string            `bson:"name"`
	Schedule        string            `bson:"schedule"`
	AgentName       string            `bson:"agent_name"`
	Input           string            `bson:"input"`
	Timezone        string            `bson:"timezone"`
	Enabled         bool              `bson:"enabled"`
	DeliveryType    int32             `bson:"delivery_type"`
	WebhookURL      string            `bson:"webhook_url"`
	ChannelName     string            `bson:"channel_name"`
	ChatID          string            `bson:"chat_id"`
	NotifyGroupName string            `bson:"notify_group_name"`
	TimeoutMs       int64             `bson:"timeout_ms,omitempty"`
	RetryMaxAttempts int32            `bson:"retry_max_attempts,omitempty"`
	RetryBackoffMs  int64             `bson:"retry_backoff_ms,omitempty"`
	ConcurrencyPolicy int32           `bson:"concurrency_policy,omitempty"`
	NotifyOn        int32             `bson:"notify_on,omitempty"`
	MaxOutputBytes  int32             `bson:"max_output_bytes,omitempty"`
	Metadata        map[string]string `bson:"metadata,omitempty"`
}

func jobCompositeID(workspaceID, name string) string { return workspaceID + ":" + name }

func jobDocFromProto(j *agentsv1.CronJob) *cronJobDoc {
	doc := &cronJobDoc{
		ID:          jobCompositeID(j.GetWorkspaceId(), j.GetName()),
		WorkspaceID: j.GetWorkspaceId(),
		Name:        j.GetName(),
		Schedule:    j.GetSchedule(),
		AgentName:   j.GetAgentName(),
		Input:       j.GetInput(),
		Timezone:    j.GetTimezone(),
		Enabled:     j.GetEnabled(),
		Metadata:    j.GetMetadata(),
		ConcurrencyPolicy: int32(j.GetConcurrencyPolicy()),
		NotifyOn:    int32(j.GetNotifyOn()),
		MaxOutputBytes: j.GetMaxOutputBytes(),
	}
	if timeout := j.GetTimeout(); timeout != nil {
		doc.TimeoutMs = timeout.AsDuration().Milliseconds()
	}
	if retry := j.GetRetry(); retry != nil {
		doc.RetryMaxAttempts = retry.GetMaxAttempts()
		if retry.GetBackoff() != nil {
			doc.RetryBackoffMs = retry.GetBackoff().AsDuration().Milliseconds()
		}
	}
	if d := j.GetDelivery(); d != nil {
		doc.DeliveryType = int32(d.GetType())
		doc.WebhookURL = d.GetWebhookUrl()
		doc.ChannelName = d.GetChannelName()
		doc.ChatID = d.GetChatId()
		doc.NotifyGroupName = d.GetNotifyGroupName()
	}
	return doc
}

func jobDocToProto(d *cronJobDoc) *agentsv1.CronJob {
	job := &agentsv1.CronJob{
		Name:        d.Name,
		Schedule:    d.Schedule,
		AgentName:   d.AgentName,
		Input:       d.Input,
		Timezone:    d.Timezone,
		Enabled:     d.Enabled,
		Metadata:    d.Metadata,
		WorkspaceId: d.WorkspaceID,
		ConcurrencyPolicy: agentsv1.CronConcurrencyPolicy(d.ConcurrencyPolicy),
		NotifyOn:    agentsv1.CronNotifyOn(d.NotifyOn),
		MaxOutputBytes: d.MaxOutputBytes,
	}
	if d.TimeoutMs > 0 {
		job.Timeout = durationpb.New(time.Duration(d.TimeoutMs) * time.Millisecond)
	}
	if d.RetryMaxAttempts > 0 || d.RetryBackoffMs > 0 {
		job.Retry = &agentsv1.CronRetryPolicy{
			MaxAttempts: d.RetryMaxAttempts,
			Backoff:     durationpb.New(time.Duration(d.RetryBackoffMs) * time.Millisecond),
		}
	}
	if d.DeliveryType != 0 || d.WebhookURL != "" || d.ChannelName != "" || d.NotifyGroupName != "" {
		job.Delivery = &agentsv1.CronDelivery{
			Type:            agentsv1.CronDeliveryType(d.DeliveryType),
			WebhookUrl:      d.WebhookURL,
			ChannelName:     d.ChannelName,
			ChatId:          d.ChatID,
			NotifyGroupName: d.NotifyGroupName,
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

func (r *MongoJobRepo) List(ctx context.Context, workspaceID string) ([]*agentsv1.CronJob, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"workspace_id": workspaceID})
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

func (r *MongoJobRepo) ListAll(ctx context.Context) ([]*agentsv1.CronJob, error) {
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

func (r *MongoJobRepo) Get(ctx context.Context, workspaceID, name string) (*agentsv1.CronJob, error) {
	var doc cronJobDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": jobCompositeID(workspaceID, name)}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("cron job %q (workspace %q) not found", name, workspaceID)
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
			return fmt.Errorf("cron job %q (workspace %q) already exists", job.GetName(), job.GetWorkspaceId())
		}
		return fmt.Errorf("create cron job: %w", err)
	}
	return nil
}

func (r *MongoJobRepo) Update(ctx context.Context, job *agentsv1.CronJob) error {
	doc := jobDocFromProto(job)
	result, err := r.coll.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return fmt.Errorf("update cron job: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("cron job %q (workspace %q) not found", job.GetName(), job.GetWorkspaceId())
	}
	return nil
}

func (r *MongoJobRepo) Delete(ctx context.Context, workspaceID, name string) error {
	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": jobCompositeID(workspaceID, name)})
	if err != nil {
		return fmt.Errorf("delete cron job: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("cron job %q (workspace %q) not found", name, workspaceID)
	}
	return nil
}
