package mongo

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/invocation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const collectionName = "invocations"

// Store is a MongoDB-backed implementation of invocation.Repository.
//
// Records keep the canonical protojson payload in `spec` and denormalize the
// indexable identifiers (workspace_id, agent_name, session_id, started_at)
// into top-level fields so List queries can do exact-match BSON filters
// instead of regex scans over the JSON blob. The owner stamp lives only in the
// top-level `owner` field, never in `spec`, so it stays out of API responses.
type Store struct {
	coll *mongo.Collection
	// owner is the instance ID stamped on the records this store creates.
	owner string
	// beforeFail runs between selecting a stale record and failing it. Tests
	// use it to race a save against a sweep.
	beforeFail func()
}

type doc struct {
	ID string `bson:"_id"`
	// Fields is named, not embedded: the bson codec skips an embedded field
	// whose type is unexported.
	Fields fields `bson:",inline"`
	// Owner is written once, when Save creates the record (#390).
	Owner string `bson:"owner,omitempty"`
}

// fields are what every Save rewrites.
type fields struct {
	WorkspaceID string    `bson:"workspace_id,omitempty"`
	AgentName   string    `bson:"agent_name,omitempty"`
	AgentID     string    `bson:"agent_id,omitempty"`
	SessionID   string    `bson:"session_id,omitempty"`
	StartedAt   time.Time `bson:"started_at,omitempty"`
	Status      string    `bson:"status,omitempty"`
	RequestID   string    `bson:"request_id,omitempty"`
	Spec        string    `bson:"spec"`
}

var _ invocation.Repository = (*Store)(nil)

// New returns a store whose records carry no owner stamp.
func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collectionName)}
}

// WithOwner returns a store over the same collection that stamps each record
// it creates with owner: the instance ID of the process that runs it.
func (s *Store) WithOwner(owner string) *Store {
	return &Store{coll: s.coll, owner: owner}
}

func activeStatuses() []string {
	return []string{
		agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED.String(),
		agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING.String(),
	}
}

// EnsureIndexes creates the indexes used by the List, ListRecent and
// stale-sweep queries.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "agent_name", Value: 1}, {Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "agent_id", Value: 1}, {Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "session_id", Value: 1}, {Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "started_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "request_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"request_id": bson.M{"$gt": ""}})},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "session_id", Value: 1}, {Key: "status", Value: 1}}},
		// Every Pod's periodic stale sweep reads the active records and their
		// owners.
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "owner", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("create invocation indexes: %w", err)
	}
	return nil
}

// Save upserts the record. The owner stamp is set only when the record is
// created, so a later save from any process (a terminal status, a redaction)
// keeps it.
func (s *Store) Save(ctx context.Context, inv *agentsv1.Invocation) error {
	b, err := protojson.Marshal(inv)
	if err != nil {
		return fmt.Errorf("marshal invocation: %w", err)
	}
	f := fields{
		WorkspaceID: inv.GetWorkspaceId(),
		AgentName:   inv.GetAgentName(),
		AgentID:     inv.GetAgentId(),
		SessionID:   inv.GetSessionId(),
		Status:      inv.GetStatus().String(),
		RequestID:   inv.GetRequestId(),
		Spec:        string(b),
	}
	if ts := inv.GetStartedAt(); ts != nil {
		f.StartedAt = ts.AsTime()
	}
	update := bson.M{"$set": f}
	if s.owner != "" {
		update["$setOnInsert"] = bson.M{"owner": s.owner}
	}
	_, err = s.coll.UpdateOne(ctx, bson.M{"_id": inv.GetId()}, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("upsert invocation: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, workspaceID, id string) (*agentsv1.Invocation, error) {
	return s.get(ctx, bson.M{"_id": id, "workspace_id": workspaceID})
}

func (s *Store) GetAcrossWorkspaces(ctx context.Context, id string) (*agentsv1.Invocation, error) {
	return s.get(ctx, bson.M{"_id": id})
}

func (s *Store) get(ctx context.Context, filter bson.M) (*agentsv1.Invocation, error) {
	var d doc
	err := s.coll.FindOne(ctx, filter).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, invocation.ErrNotFound
		}
		return nil, fmt.Errorf("get invocation: %w", err)
	}
	return decode(&d)
}

func (s *Store) List(ctx context.Context, filter invocation.ListFilter, pageSize int32, pageToken string) ([]*agentsv1.Invocation, string, int32, error) {
	q := bson.M{}
	if filter.WorkspaceID != "" {
		q["workspace_id"] = filter.WorkspaceID
	}
	if filter.AgentID != "" {
		q["agent_id"] = filter.AgentID
	}
	if filter.AgentName != "" {
		q["agent_name"] = filter.AgentName
	}
	if filter.SessionID != "" {
		q["session_id"] = filter.SessionID
	}

	total, err := s.coll.CountDocuments(ctx, q)
	if err != nil {
		return nil, "", 0, fmt.Errorf("count invocations: %w", err)
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	offset := decodeToken(pageToken)

	opts := options.Find().
		SetSort(bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(pageSize))

	cursor, err := s.coll.Find(ctx, q, opts)
	if err != nil {
		return nil, "", 0, fmt.Errorf("list invocations: %w", err)
	}
	defer cursor.Close(ctx)

	out, err := drain(ctx, cursor)
	if err != nil {
		return nil, "", 0, err
	}

	next := ""
	if int64(offset)+int64(len(out)) < total {
		next = encodeToken(offset + len(out))
	}
	return out, next, int32(total), nil
}

func (s *Store) ListRecent(ctx context.Context, limit int32, pageToken string) ([]*agentsv1.Invocation, string, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	offset := decodeToken(pageToken)

	opts := options.Find().
		SetSort(bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := s.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, "", fmt.Errorf("list recent invocations: %w", err)
	}
	defer cursor.Close(ctx)

	out, err := drain(ctx, cursor)
	if err != nil {
		return nil, "", err
	}
	next := ""
	if int32(len(out)) == limit {
		next = encodeToken(offset + len(out))
	}
	return out, next, nil
}

// StatusSummaries aggregates each agent's latest invocation and RUNNING count
// in a single query. Documents written before the status field was
// denormalized have no `status` and therefore never count as RUNNING.
func (s *Store) StatusSummaries(ctx context.Context, workspaceID string, agentNames []string) (map[string]invocation.StatusSummary, error) {
	if len(agentNames) == 0 {
		return map[string]invocation.StatusSummary{}, nil
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"workspace_id": workspaceID,
			"agent_name":   bson.M{"$in": agentNames},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}}},
		{{Key: "$group", Value: bson.M{
			"_id":    "$agent_name",
			"latest": bson.M{"$first": "$spec"},
			"running": bson.M{"$sum": bson.M{"$cond": bson.A{
				bson.M{"$eq": bson.A{"$status", agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING.String()}},
				1, 0,
			}}},
		}}},
	}
	cursor, err := s.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate invocation summaries: %w", err)
	}
	defer cursor.Close(ctx)

	out := make(map[string]invocation.StatusSummary, len(agentNames))
	for cursor.Next(ctx) {
		var row struct {
			AgentName string `bson:"_id"`
			Latest    string `bson:"latest"`
			Running   int32  `bson:"running"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, fmt.Errorf("decode invocation summary: %w", err)
		}
		inv := &agentsv1.Invocation{}
		if err := protojson.Unmarshal([]byte(row.Latest), inv); err != nil {
			return nil, fmt.Errorf("unmarshal invocation: %w", err)
		}
		out[row.AgentName] = invocation.StatusSummary{Latest: inv, Running: row.Running}
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("aggregate invocation summaries: %w", err)
	}
	return out, nil
}

func (s *Store) CountByTimeRange(ctx context.Context, start, end time.Time) (int64, int64, error) {
	window := bson.M{"started_at": bson.M{"$gte": start, "$lt": end}}
	total, err := s.coll.CountDocuments(ctx, window)
	if err != nil {
		return 0, 0, fmt.Errorf("count invocations: %w", err)
	}
	failed, err := s.coll.CountDocuments(ctx, bson.M{
		"started_at": bson.M{"$gte": start, "$lt": end},
		"status":     agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED.String(),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("count failed invocations: %w", err)
	}
	return total, failed, nil
}

func (s *Store) FindByRequestID(ctx context.Context, workspaceID, requestID string) (*agentsv1.Invocation, error) {
	var d doc
	err := s.coll.FindOne(ctx, bson.M{
		"workspace_id": workspaceID,
		"request_id":   requestID,
	}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, invocation.ErrNotFound
		}
		return nil, fmt.Errorf("find invocation by request_id: %w", err)
	}
	return decode(&d)
}

func (s *Store) FindActiveBySession(ctx context.Context, workspaceID, sessionID string) (*agentsv1.Invocation, error) {
	var d doc
	err := s.coll.FindOne(ctx, bson.M{
		"workspace_id": workspaceID,
		"session_id":   sessionID,
		"status":       bson.M{"$in": activeStatuses()},
	}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, invocation.ErrNotFound
		}
		return nil, fmt.Errorf("find active invocation by session: %w", err)
	}
	return decode(&d)
}

func (s *Store) ActiveOwners(ctx context.Context) ([]string, error) {
	result := s.coll.Distinct(ctx, "owner", bson.M{"status": bson.M{"$in": activeStatuses()}})
	var values []any
	if err := result.Decode(&values); err != nil {
		return nil, fmt.Errorf("list owners of active invocations: %w", err)
	}
	owners := make([]string, 0, len(values))
	for _, v := range values {
		if owner, ok := v.(string); ok && owner != "" {
			owners = append(owners, owner)
		}
	}
	sort.Strings(owners)
	return owners, nil
}

func (s *Store) MarkStaleRunning(ctx context.Context, sel invocation.StaleSelection) (int64, error) {
	var stale bson.A
	if len(sel.LostOwners) > 0 {
		stale = append(stale, bson.M{"owner": bson.M{"$in": sel.LostOwners}})
	}
	if !sel.LegacyBefore.IsZero() {
		// A null in $in also matches a missing field: records from before
		// owner stamps have none.
		stale = append(stale, bson.M{
			"owner": bson.M{"$in": bson.A{nil, ""}},
			"$or": bson.A{
				bson.M{"started_at": bson.M{"$lt": sel.LegacyBefore}},
				bson.M{"started_at": bson.M{"$exists": false}},
			},
		})
	}
	if len(stale) == 0 {
		return 0, nil
	}
	cursor, err := s.coll.Find(ctx, bson.M{"status": bson.M{"$in": activeStatuses()}, "$or": stale})
	if err != nil {
		return 0, fmt.Errorf("find stale invocations: %w", err)
	}
	var docs []doc
	if err := cursor.All(ctx, &docs); err != nil {
		return 0, fmt.Errorf("decode stale invocations: %w", err)
	}
	if s.beforeFail != nil {
		s.beforeFail()
	}
	var count int64
	for i := range docs {
		d := &docs[i]
		inv, err := decode(d)
		if err != nil {
			return count, err
		}
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED
		inv.Error = ""
		if sel.Reason != nil {
			inv.Error = sel.Reason(d.Owner)
		}
		inv.FinishedAt = timestamppb.Now()
		spec, err := protojson.Marshal(inv)
		if err != nil {
			return count, fmt.Errorf("marshal invocation: %w", err)
		}
		// The authoritative record is the protojson spec, so it is rewritten
		// along with the status field. The filter matches the record only as
		// it was read: if its owner, or a redaction, saved it in between, the
		// newer save stands and the next sweep judges it again.
		res, err := s.coll.UpdateOne(ctx,
			bson.M{"_id": d.ID, "spec": d.Fields.Spec},
			bson.M{"$set": bson.M{"status": inv.GetStatus().String(), "spec": string(spec)}})
		if err != nil {
			return count, fmt.Errorf("mark stale invocation %s: %w", d.ID, err)
		}
		count += res.ModifiedCount
	}
	return count, nil
}

func (s *Store) FindLatestBySession(ctx context.Context, workspaceID, sessionID string) (*agentsv1.Invocation, error) {
	var d doc
	err := s.coll.FindOne(ctx, bson.M{
		"workspace_id": workspaceID,
		"session_id":   sessionID,
	}, options.FindOne().SetSort(bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: -1}})).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, invocation.ErrNotFound
		}
		return nil, fmt.Errorf("find latest invocation by session: %w", err)
	}
	return decode(&d)
}

func (s *Store) ListBySession(ctx context.Context, workspaceID, sessionID string) ([]*agentsv1.Invocation, error) {
	cursor, err := s.coll.Find(ctx, bson.M{
		"workspace_id": workspaceID,
		"session_id":   sessionID,
	}, options.Find().SetSort(bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list invocations by session: %w", err)
	}
	defer cursor.Close(ctx)
	return drain(ctx, cursor)
}

func (s *Store) RedactContent(ctx context.Context, workspaceID, id string) error {
	inv, err := s.Get(ctx, workspaceID, id)
	if err != nil {
		return err
	}
	inv.Input = ""
	inv.Output = ""
	inv.Error = ""
	return s.Save(ctx, inv)
}

func drain(ctx context.Context, cursor *mongo.Cursor) ([]*agentsv1.Invocation, error) {
	var out []*agentsv1.Invocation
	for cursor.Next(ctx) {
		var d doc
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode invocation: %w", err)
		}
		inv, err := decode(&d)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, nil
}

func decode(d *doc) (*agentsv1.Invocation, error) {
	inv := &agentsv1.Invocation{}
	if err := protojson.Unmarshal([]byte(d.Fields.Spec), inv); err != nil {
		return nil, fmt.Errorf("unmarshal invocation: %w", err)
	}
	return inv, nil
}

func decodeToken(token string) int {
	if token == "" {
		return 0
	}
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(string(raw))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func encodeToken(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}
