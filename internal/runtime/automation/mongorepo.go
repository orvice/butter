package automation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	definitionsCollection = "automations"
	runsCollection        = "automation_runs"
	stepRunsCollection    = "automation_step_runs"
)

// runRetention is how long finished run and step-run records are kept, via a
// TTL index on finished_at (30 days, matching the Telegram processing-record
// retention, ADR-0009). Unfinished runs — RUNNING and WAITING_INPUT — omit
// finished_at entirely, so TTL never touches them.
const runRetention = 30 * 24 * time.Hour

type definitionDoc struct {
	ID          string    `bson:"_id"`
	WorkspaceID string    `bson:"workspace_id"`
	Name        string    `bson:"name"`
	Enabled     bool      `bson:"enabled"`
	TriggerType string    `bson:"trigger_type,omitempty"`
	UpdatedAt   time.Time `bson:"updated_at,omitempty"`
	Spec        string    `bson:"spec"`
}

type runDoc struct {
	ID             string    `bson:"_id"`
	WorkspaceID    string    `bson:"workspace_id"`
	AutomationName string    `bson:"automation_name"`
	Status         string    `bson:"status"`
	TriggerType    string    `bson:"trigger_type,omitempty"`
	StartedAt      time.Time `bson:"started_at,omitempty"`
	FinishedAt     time.Time `bson:"finished_at,omitempty"`
	// Session coordinates are promoted out of Spec so a WAITING_INPUT run can be
	// looked up by session on the resume path (a runner turn or session delete).
	SessionAppName string `bson:"session_app_name,omitempty"`
	SessionUserID  string `bson:"session_user_id,omitempty"`
	SessionID      string `bson:"session_id,omitempty"`
	Spec           string `bson:"spec"`
}

type stepRunDoc struct {
	ID             string    `bson:"_id"`
	WorkspaceID    string    `bson:"workspace_id"`
	RunID          string    `bson:"run_id"`
	AutomationName string    `bson:"automation_name"`
	StepName       string    `bson:"step_name"`
	StepType       string    `bson:"step_type,omitempty"`
	Status         string    `bson:"status"`
	Order          int32     `bson:"order"`
	StartedAt      time.Time `bson:"started_at,omitempty"`
	// FinishedAt is promoted out of Spec for the retention TTL index.
	FinishedAt time.Time `bson:"finished_at,omitempty"`
	Spec       string    `bson:"spec"`
}

type MongoDefinitionRepo struct {
	coll *mongo.Collection
}

func NewMongoDefinitionRepo(db *mongo.Database) *MongoDefinitionRepo {
	return &MongoDefinitionRepo{coll: db.Collection(definitionsCollection)}
}

func (r *MongoDefinitionRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "updated_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "trigger_type", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("create automation indexes: %w", err)
	}
	return nil
}

func (r *MongoDefinitionRepo) List(ctx context.Context, workspaceID string) ([]*agentsv1.Automation, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"workspace_id": workspaceID}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list automations: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.Automation
	for cursor.Next(ctx) {
		var d definitionDoc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode automation: %w", err)
		}
		a, err := decodeAutomation(d.Spec)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list automations: %w", err)
	}
	return out, nil
}

func (r *MongoDefinitionRepo) ListAll(ctx context.Context) ([]*agentsv1.Automation, error) {
	cursor, err := r.coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "workspace_id", Value: 1}, {Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list all automations: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.Automation
	for cursor.Next(ctx) {
		var d definitionDoc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode automation: %w", err)
		}
		a, err := decodeAutomation(d.Spec)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list all automations: %w", err)
	}
	return out, nil
}

func (r *MongoDefinitionRepo) Get(ctx context.Context, workspaceID, name string) (*agentsv1.Automation, error) {
	var d definitionDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": automationID(workspaceID, name)}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAutomationNotFound
		}
		return nil, fmt.Errorf("get automation: %w", err)
	}
	return decodeAutomation(d.Spec)
}

func (r *MongoDefinitionRepo) Create(ctx context.Context, automation *agentsv1.Automation) error {
	d, err := definitionDocFromProto(automation)
	if err != nil {
		return err
	}
	_, err = r.coll.InsertOne(ctx, d)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("automation %q already exists in workspace %q: %w", automation.GetName(), automation.GetWorkspaceId(), ErrAutomationAlreadyExists)
		}
		return fmt.Errorf("create automation: %w", err)
	}
	return nil
}

func (r *MongoDefinitionRepo) Update(ctx context.Context, automation *agentsv1.Automation) error {
	d, err := definitionDocFromProto(automation)
	if err != nil {
		return err
	}
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": d.ID}, d)
	if err != nil {
		return fmt.Errorf("update automation: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrAutomationNotFound
	}
	return nil
}

func (r *MongoDefinitionRepo) Delete(ctx context.Context, workspaceID, name string) (*agentsv1.Automation, error) {
	var d definitionDoc
	err := r.coll.FindOneAndDelete(ctx, bson.M{"_id": automationID(workspaceID, name)}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAutomationNotFound
		}
		return nil, fmt.Errorf("delete automation: %w", err)
	}
	return decodeAutomation(d.Spec)
}

type MongoRunRepo struct {
	coll *mongo.Collection
}

func NewMongoRunRepo(db *mongo.Database) *MongoRunRepo {
	return &MongoRunRepo{coll: db.Collection(runsCollection)}
}

func (r *MongoRunRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "automation_name", Value: 1}, {Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "status", Value: 1}}},
		// Stale-run reconciliation lists RUNNING runs across workspaces.
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "started_at", Value: 1}}},
		{Keys: bson.D{{Key: "session_app_name", Value: 1}, {Key: "session_user_id", Value: 1}, {Key: "session_id", Value: 1}, {Key: "status", Value: 1}}},
		// Retention: finished runs expire; unfinished runs omit finished_at.
		{
			Keys:    bson.D{{Key: "finished_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(runRetention.Seconds())),
		},
	})
	if err != nil {
		return fmt.Errorf("create automation run indexes: %w", err)
	}
	return nil
}

func (r *MongoRunRepo) Save(ctx context.Context, run *agentsv1.AutomationRun) error {
	d, err := runDocFromProto(run)
	if err != nil {
		return err
	}
	_, err = r.coll.ReplaceOne(ctx, bson.M{"_id": run.GetId(), "workspace_id": run.GetWorkspaceId()}, d, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("save automation run: %w", err)
	}
	return nil
}

func (r *MongoRunRepo) Get(ctx context.Context, workspaceID, id string) (*agentsv1.AutomationRun, error) {
	var d runDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRunNotFound
		}
		return nil, fmt.Errorf("get automation run: %w", err)
	}
	return decodeRun(d.Spec)
}

func (r *MongoRunRepo) List(ctx context.Context, workspaceID, automationName string, pageSize int32, pageToken string) ([]*agentsv1.AutomationRun, string, error) {
	q := bson.M{"workspace_id": workspaceID}
	if automationName != "" {
		q["automation_name"] = automationName
	}
	pageSize = clampRunPageSize(pageSize)
	if after, afterID, ok := DecodeRunPageToken(pageToken); ok {
		q["$or"] = []bson.M{
			{"started_at": bson.M{"$lt": after}},
			{"started_at": after, "_id": bson.M{"$lt": afterID}},
		}
	}

	cursor, err := r.coll.Find(ctx, q, options.Find().
		SetSort(bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(int64(pageSize+1)),
	)
	if err != nil {
		return nil, "", fmt.Errorf("list automation runs: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.AutomationRun
	for cursor.Next(ctx) {
		var d runDoc
		if err := cursor.Decode(&d); err != nil {
			return nil, "", fmt.Errorf("decode automation run: %w", err)
		}
		run, err := decodeRun(d.Spec)
		if err != nil {
			return nil, "", err
		}
		out = append(out, run)
	}
	if err := cursor.Err(); err != nil {
		return nil, "", fmt.Errorf("list automation runs: %w", err)
	}

	next := ""
	if len(out) > int(pageSize) {
		out = out[:pageSize]
		last := out[len(out)-1]
		next = EncodeRunPageToken(last.GetStartedAt().AsTime(), last.GetId())
	}
	return out, next, nil
}

func (r *MongoRunRepo) ListStaleRunning(ctx context.Context, before time.Time) ([]*agentsv1.AutomationRun, error) {
	cursor, err := r.coll.Find(ctx, bson.M{
		"status":     agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_RUNNING.String(),
		"started_at": bson.M{"$lt": before},
	})
	if err != nil {
		return nil, fmt.Errorf("list stale automation runs: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.AutomationRun
	for cursor.Next(ctx) {
		var d runDoc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode automation run: %w", err)
		}
		run, err := decodeRun(d.Spec)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list stale automation runs: %w", err)
	}
	return out, nil
}

func (r *MongoRunRepo) ListWaitingBySession(ctx context.Context, appName, userID, sessionID string) ([]*agentsv1.AutomationRun, error) {
	cursor, err := r.coll.Find(ctx, bson.M{
		"session_app_name": appName,
		"session_user_id":  userID,
		"session_id":       sessionID,
		"status":           agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT.String(),
	}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list waiting automation runs: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.AutomationRun
	for cursor.Next(ctx) {
		var d runDoc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode automation run: %w", err)
		}
		run, err := decodeRun(d.Spec)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list waiting automation runs: %w", err)
	}
	return out, nil
}

type MongoStepRunRepo struct {
	coll *mongo.Collection
}

func NewMongoStepRunRepo(db *mongo.Database) *MongoStepRunRepo {
	return &MongoStepRunRepo{coll: db.Collection(stepRunsCollection)}
}

func (r *MongoStepRunRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "run_id", Value: 1}, {Key: "order", Value: 1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "automation_name", Value: 1}, {Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "status", Value: 1}}},
		// Retention: mirrors the run TTL so a run's steps never outlive it by
		// more than the render of their own finish times.
		{
			Keys:    bson.D{{Key: "finished_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(runRetention.Seconds())),
		},
	})
	if err != nil {
		return fmt.Errorf("create automation step run indexes: %w", err)
	}
	return nil
}

func (r *MongoStepRunRepo) Save(ctx context.Context, stepRun *agentsv1.AutomationStepRun) error {
	d, err := stepRunDocFromProto(stepRun)
	if err != nil {
		return err
	}
	_, err = r.coll.ReplaceOne(ctx, bson.M{"_id": stepRun.GetId(), "workspace_id": stepRun.GetWorkspaceId()}, d, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("save automation step run: %w", err)
	}
	return nil
}

func (r *MongoStepRunRepo) Get(ctx context.Context, workspaceID, id string) (*agentsv1.AutomationStepRun, error) {
	var d stepRunDoc
	err := r.coll.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrStepRunNotFound
		}
		return nil, fmt.Errorf("get automation step run: %w", err)
	}
	return decodeStepRun(d.Spec)
}

func (r *MongoStepRunRepo) ListByRun(ctx context.Context, workspaceID, runID string) ([]*agentsv1.AutomationStepRun, error) {
	cursor, err := r.coll.Find(ctx,
		bson.M{"workspace_id": workspaceID, "run_id": runID},
		options.Find().SetSort(bson.D{{Key: "order", Value: 1}, {Key: "started_at", Value: 1}, {Key: "_id", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("list automation step runs: %w", err)
	}
	defer cursor.Close(ctx)

	var out []*agentsv1.AutomationStepRun
	for cursor.Next(ctx) {
		var d stepRunDoc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode automation step run: %w", err)
		}
		stepRun, err := decodeStepRun(d.Spec)
		if err != nil {
			return nil, err
		}
		out = append(out, stepRun)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list automation step runs: %w", err)
	}
	return out, nil
}

func automationID(workspaceID, name string) string { return workspaceID + ":" + name }

func definitionDocFromProto(a *agentsv1.Automation) (*definitionDoc, error) {
	b, err := protojson.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("marshal automation: %w", err)
	}
	d := &definitionDoc{
		ID:          automationID(a.GetWorkspaceId(), a.GetName()),
		WorkspaceID: a.GetWorkspaceId(),
		Name:        a.GetName(),
		Enabled:     a.GetEnabled(),
		Spec:        string(b),
	}
	if trigger := a.GetTrigger(); trigger != nil {
		d.TriggerType = trigger.GetType().String()
	}
	if ts := a.GetUpdatedAt(); ts != nil {
		d.UpdatedAt = ts.AsTime()
	}
	return d, nil
}

func runDocFromProto(run *agentsv1.AutomationRun) (*runDoc, error) {
	b, err := protojson.Marshal(run)
	if err != nil {
		return nil, fmt.Errorf("marshal automation run: %w", err)
	}
	d := &runDoc{
		ID:             run.GetId(),
		WorkspaceID:    run.GetWorkspaceId(),
		AutomationName: run.GetAutomationName(),
		Status:         run.GetStatus().String(),
		TriggerType:    run.GetTriggerType().String(),
		SessionAppName: run.GetSessionAppName(),
		SessionUserID:  run.GetSessionUserId(),
		SessionID:      run.GetSessionId(),
		Spec:           string(b),
	}
	if ts := run.GetStartedAt(); ts != nil {
		d.StartedAt = ts.AsTime()
	}
	if ts := run.GetFinishedAt(); ts != nil {
		d.FinishedAt = ts.AsTime()
	}
	return d, nil
}

func stepRunDocFromProto(stepRun *agentsv1.AutomationStepRun) (*stepRunDoc, error) {
	b, err := protojson.Marshal(stepRun)
	if err != nil {
		return nil, fmt.Errorf("marshal automation step run: %w", err)
	}
	d := &stepRunDoc{
		ID:             stepRun.GetId(),
		WorkspaceID:    stepRun.GetWorkspaceId(),
		RunID:          stepRun.GetRunId(),
		AutomationName: stepRun.GetAutomationName(),
		StepName:       stepRun.GetStepName(),
		StepType:       stepRun.GetStepType().String(),
		Status:         stepRun.GetStatus().String(),
		Order:          stepRun.GetOrder(),
		Spec:           string(b),
	}
	if ts := stepRun.GetStartedAt(); ts != nil {
		d.StartedAt = ts.AsTime()
	}
	if ts := stepRun.GetFinishedAt(); ts != nil {
		d.FinishedAt = ts.AsTime()
	}
	return d, nil
}

func decodeAutomation(spec string) (*agentsv1.Automation, error) {
	a := &agentsv1.Automation{}
	if err := protojson.Unmarshal([]byte(spec), a); err != nil {
		return nil, fmt.Errorf("unmarshal automation: %w", err)
	}
	return a, nil
}

func decodeRun(spec string) (*agentsv1.AutomationRun, error) {
	run := &agentsv1.AutomationRun{}
	if err := protojson.Unmarshal([]byte(spec), run); err != nil {
		return nil, fmt.Errorf("unmarshal automation run: %w", err)
	}
	return run, nil
}

func decodeStepRun(spec string) (*agentsv1.AutomationStepRun, error) {
	stepRun := &agentsv1.AutomationStepRun{}
	if err := protojson.Unmarshal([]byte(spec), stepRun); err != nil {
		return nil, fmt.Errorf("unmarshal automation step run: %w", err)
	}
	return stepRun, nil
}

