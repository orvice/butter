package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/workspace"
)

const (
	sessionsCollection = "adk_sessions"
	eventsCollection   = "adk_events"
)

// ErrSessionNotFound is wrapped into errors returned when a session addressed
// by app_name/user_id/session_id does not exist; callers match with errors.Is.
var ErrSessionNotFound = errors.New("session not found")

// sessionDoc is the MongoDB document for a session.
type sessionDoc struct {
	SessionID      string         `bson:"session_id"`
	AppName        string         `bson:"app_name"`
	UserID         string         `bson:"user_id"`
	State          map[string]any `bson:"state"`
	LastUpdateTime time.Time      `bson:"last_update_time"`
	Title          string         `bson:"title,omitempty"`
	WorkspaceID    string         `bson:"workspace_id,omitempty"`
	LastReadAt     *time.Time     `bson:"last_read_at,omitempty"`
}

// eventDoc is the MongoDB document for an event.
type eventDoc struct {
	SessionID    string    `bson:"session_id"`
	AppName      string    `bson:"app_name"`
	EventID      string    `bson:"event_id"`
	InvocationID string    `bson:"invocation_id"`
	Author       string    `bson:"author"`
	Branch       string    `bson:"branch"`
	ContentJSON  []byte    `bson:"content_json,omitempty"`
	Timestamp    time.Time `bson:"timestamp"`
}

// Service implements session.Service backed by MongoDB.
type Service struct {
	db       *mongo.Database
	sessions *mongo.Collection
	events   *mongo.Collection
}

// CountSessions returns the total number of stored ADK sessions, used by the
// dashboard overview. Returns 0 if the underlying collection cannot be read.
func (s *Service) CountSessions(ctx context.Context) (int64, error) {
	return s.sessions.CountDocuments(ctx, bson.M{})
}

// SetSessionTitle updates the first-class title on a session document without
// touching last_update_time, so rename does not reorder conversation history.
func (s *Service) SetSessionTitle(ctx context.Context, appName, userID, sessionID, title string) (SessionTitleResult, error) {
	filter := bson.M{
		"app_name":   appName,
		"user_id":    userID,
		"session_id": sessionID,
	}
	update := bson.M{"$set": bson.M{"title": title}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc sessionDoc
	if err := s.sessions.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return SessionTitleResult{}, fmt.Errorf("%w: %s/%s/%s", ErrSessionNotFound, appName, userID, sessionID)
		}
		return SessionTitleResult{}, fmt.Errorf("updating session title: %w", err)
	}
	return SessionTitleResult{
		SessionID:      doc.SessionID,
		AppName:        doc.AppName,
		UserID:         doc.UserID,
		Title:          doc.Title,
		LastUpdateTime: doc.LastUpdateTime,
	}, nil
}

// SessionTitleResult carries the post-update session snapshot.
type SessionTitleResult struct {
	SessionID      string
	AppName        string
	UserID         string
	Title          string
	LastUpdateTime time.Time
}

// SetSessionTitleIfEmpty is an atomic compare-and-set that writes the title
// only when the first-class title is currently empty AND the legacy
// state["title"] is absent, null, or the empty string. The legacy guard is
// deliberately conservative: any other present value blocks the write, even
// one the read path would ignore (whitespace-only or non-string), so an
// auto-generated title can never shadow a concurrently written legacy title.
// Returns (result, true) when a new title was actually written, or
// (current, false) when an effective title already existed.
// last_update_time is never changed.
func (s *Service) SetSessionTitleIfEmpty(ctx context.Context, appName, userID, sessionID, title string) (SessionTitleResult, bool, error) {
	// {"$in": [nil, ""]} matches documents where the field is missing, null,
	// or the empty string.
	filter := bson.M{
		"app_name":    appName,
		"user_id":     userID,
		"session_id":  sessionID,
		"title":       bson.M{"$in": bson.A{nil, ""}},
		"state.title": bson.M{"$in": bson.A{nil, ""}},
	}
	update := bson.M{"$set": bson.M{"title": title}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc sessionDoc
	if err := s.sessions.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			// Either the session doesn't exist or the title is already set.
			// Check which case by trying to find the session.
			var existing sessionDoc
			lookupFilter := bson.M{
				"app_name":   appName,
				"user_id":    userID,
				"session_id": sessionID,
			}
			if findErr := s.sessions.FindOne(ctx, lookupFilter).Decode(&existing); findErr != nil {
				if findErr == mongo.ErrNoDocuments {
					return SessionTitleResult{}, false, fmt.Errorf("%w: %s/%s/%s", ErrSessionNotFound, appName, userID, sessionID)
				}
				return SessionTitleResult{}, false, fmt.Errorf("looking up session: %w", findErr)
			}
			return SessionTitleResult{
				SessionID:      existing.SessionID,
				AppName:        existing.AppName,
				UserID:         existing.UserID,
				Title:          existing.Title,
				LastUpdateTime: existing.LastUpdateTime,
			}, false, nil
		}
		return SessionTitleResult{}, false, fmt.Errorf("cas update session title: %w", err)
	}
	return SessionTitleResult{
		SessionID:      doc.SessionID,
		AppName:        doc.AppName,
		UserID:         doc.UserID,
		Title:          doc.Title,
		LastUpdateTime: doc.LastUpdateTime,
	}, true, nil
}

// New creates a new MongoDB session service and ensures indexes.
func New(ctx context.Context, db *mongo.Database) (*Service, error) {
	logger := log.FromContext(ctx)
	logger.Info("initializing mongo session service", "database", db.Name())

	s := &Service{
		db:       db,
		sessions: db.Collection(sessionsCollection),
		events:   db.Collection(eventsCollection),
	}
	if err := s.ensureIndexes(ctx); err != nil {
		return nil, fmt.Errorf("ensuring indexes: %w", err)
	}

	logger.Info("mongo session service ready",
		"sessions_collection", sessionsCollection,
		"events_collection", eventsCollection,
	)
	return s, nil
}

func (s *Service) ensureIndexes(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Debug("ensuring mongodb indexes")

	_, err := s.sessions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "app_name", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "session_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = s.events.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "app_name", Value: 1},
			{Key: "session_id", Value: 1},
			{Key: "timestamp", Value: 1},
		},
	})
	if err != nil {
		return err
	}

	_, err = s.sessions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "workspace_id", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "last_update_time", Value: -1},
		},
	})
	if err != nil {
		return err
	}

	logger.Debug("mongodb indexes ensured")
	return nil
}

func (s *Service) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
	logger := log.FromContext(ctx)

	sid := req.SessionID
	if sid == "" {
		sid = uuid.New().String()
	}

	state := req.State
	if state == nil {
		state = make(map[string]any)
	}

	wsID, _ := workspace.FromContext(ctx)

	doc := sessionDoc{
		SessionID:      sid,
		AppName:        req.AppName,
		UserID:         req.UserID,
		State:          state,
		LastUpdateTime: time.Now(),
		WorkspaceID:    wsID,
	}

	logger.Info("creating session",
		"app_name", req.AppName,
		"user_id", req.UserID,
		"session_id", sid,
		"workspace_id", wsID,
	)

	if _, err := s.sessions.InsertOne(ctx, doc); err != nil {
		logger.Error("failed to insert session",
			"app_name", req.AppName,
			"session_id", sid,
			"err", err,
		)
		return nil, fmt.Errorf("inserting session: %w", err)
	}

	sess := &mongoSession{
		id:             sid,
		appName:        req.AppName,
		userID:         req.UserID,
		state:          newState(state),
		events:         newEvents(nil),
		lastUpdateTime: doc.LastUpdateTime,
		workspaceID:    wsID,
	}

	logger.Debug("session created", "app_name", req.AppName, "session_id", sid)
	return &session.CreateResponse{Session: sess}, nil
}

func (s *Service) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	logger := log.FromContext(ctx)

	logger.Debug("getting session",
		"app_name", req.AppName,
		"user_id", req.UserID,
		"session_id", req.SessionID,
		"num_recent_events", req.NumRecentEvents,
	)

	filter := bson.M{
		"app_name":   req.AppName,
		"user_id":    req.UserID,
		"session_id": req.SessionID,
	}

	var doc sessionDoc
	if err := s.sessions.FindOne(ctx, filter).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Debug("session not found",
				"app_name", req.AppName,
				"session_id", req.SessionID,
			)
			return nil, fmt.Errorf("%w: %s/%s/%s", ErrSessionNotFound, req.AppName, req.UserID, req.SessionID)
		}
		return nil, fmt.Errorf("finding session: %w", err)
	}

	// Query events
	eventFilter := bson.M{
		"app_name":   req.AppName,
		"session_id": req.SessionID,
	}
	if !req.After.IsZero() {
		eventFilter["timestamp"] = bson.M{"$gte": req.After}
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	if req.NumRecentEvents > 0 {
		findOpts = options.Find().
			SetSort(bson.D{{Key: "timestamp", Value: -1}}).
			SetLimit(int64(req.NumRecentEvents))
	}

	cursor, err := s.events.Find(ctx, eventFilter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("finding events: %w", err)
	}
	defer cursor.Close(ctx)

	var eventDocs []eventDoc
	if err := cursor.All(ctx, &eventDocs); err != nil {
		return nil, fmt.Errorf("decoding events: %w", err)
	}

	// If we queried descending for NumRecentEvents, reverse to chronological order.
	if req.NumRecentEvents > 0 {
		for i, j := 0, len(eventDocs)-1; i < j; i, j = i+1, j-1 {
			eventDocs[i], eventDocs[j] = eventDocs[j], eventDocs[i]
		}
	}

	events := make([]*session.Event, 0, len(eventDocs))
	for _, ed := range eventDocs {
		evt := session.NewEvent(ctx, ed.InvocationID)
		evt.ID = ed.EventID
		evt.Timestamp = ed.Timestamp
		evt.Author = ed.Author
		evt.Branch = ed.Branch
		if len(ed.ContentJSON) > 0 {
			var content genai.Content
			if err := json.Unmarshal(ed.ContentJSON, &content); err != nil {
				logger.Warn("failed to unmarshal event content",
					"event_id", ed.EventID,
					"err", err,
				)
			} else {
				evt.LLMResponse = model.LLMResponse{Content: &content}
			}
		}
		events = append(events, evt)
	}

	sess := &mongoSession{
		id:             doc.SessionID,
		appName:        doc.AppName,
		userID:         doc.UserID,
		state:          newState(doc.State),
		events:         newEvents(events),
		lastUpdateTime: doc.LastUpdateTime,
		title:          doc.Title,
		workspaceID:    doc.WorkspaceID,
		lastReadAt:     doc.LastReadAt,
	}

	logger.Debug("session loaded",
		"app_name", req.AppName,
		"session_id", req.SessionID,
		"event_count", len(events),
	)

	return &session.GetResponse{Session: sess}, nil
}

func (s *Service) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	logger := log.FromContext(ctx)

	logger.Debug("listing sessions", "app_name", req.AppName, "user_id", req.UserID)

	// Empty AppName/UserID mean "no filter" — the Session Explorer lists
	// across all channels/users by default. Including them as exact-match
	// empty strings would match nothing.
	filter := bson.M{}
	if req.AppName != "" {
		filter["app_name"] = req.AppName
	}
	if req.UserID != "" {
		filter["user_id"] = req.UserID
	}

	cursor, err := s.sessions.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []sessionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding sessions: %w", err)
	}

	sessions := make([]session.Session, 0, len(docs))
	for _, doc := range docs {
		sessions = append(sessions, &mongoSession{
			id:             doc.SessionID,
			appName:        doc.AppName,
			userID:         doc.UserID,
			state:          newState(doc.State),
			events:         newEvents(nil),
			lastUpdateTime: doc.LastUpdateTime,
			title:          doc.Title,
			workspaceID:    doc.WorkspaceID,
			lastReadAt:     doc.LastReadAt,
		})
	}

	logger.Debug("sessions listed",
		"app_name", req.AppName,
		"user_id", req.UserID,
		"count", len(sessions),
	)

	return &session.ListResponse{Sessions: sessions}, nil
}

// ListByWorkspace returns sessions owned by the given workspace and user,
// ordered newest-first. Legacy sessions without workspace_id are excluded.
func (s *Service) ListByWorkspace(ctx context.Context, workspaceID, userID string) ([]session.Session, error) {
	logger := log.FromContext(ctx)
	logger.Debug("listing sessions by workspace", "workspace_id", workspaceID, "user_id", userID)

	filter := bson.M{"workspace_id": workspaceID}
	if userID != "" {
		filter["user_id"] = userID
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "last_update_time", Value: -1}})
	cursor, err := s.sessions.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("listing workspace sessions: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []sessionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding workspace sessions: %w", err)
	}

	result := make([]session.Session, 0, len(docs))
	for _, doc := range docs {
		result = append(result, &mongoSession{
			id:             doc.SessionID,
			appName:        doc.AppName,
			userID:         doc.UserID,
			state:          newState(doc.State),
			events:         newEvents(nil),
			lastUpdateTime: doc.LastUpdateTime,
			title:          doc.Title,
			workspaceID:    doc.WorkspaceID,
			lastReadAt:     doc.LastReadAt,
		})
	}

	logger.Debug("workspace sessions listed", "workspace_id", workspaceID, "user_id", userID, "count", len(result))
	return result, nil
}

// GetWorkspaceID returns the workspace_id for a given session, or empty if
// the session does not exist or is a legacy session without workspace.
func (s *Service) GetWorkspaceID(ctx context.Context, appName, userID, sessionID string) (string, error) {
	filter := bson.M{
		"app_name":   appName,
		"user_id":    userID,
		"session_id": sessionID,
	}
	var doc sessionDoc
	if err := s.sessions.FindOne(ctx, filter, options.FindOne().SetProjection(bson.M{"workspace_id": 1})).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("%w: %s/%s/%s", ErrSessionNotFound, appName, userID, sessionID)
		}
		return "", fmt.Errorf("looking up session workspace: %w", err)
	}
	return doc.WorkspaceID, nil
}

func (s *Service) Delete(ctx context.Context, req *session.DeleteRequest) error {
	logger := log.FromContext(ctx)

	logger.Info("deleting session",
		"app_name", req.AppName,
		"user_id", req.UserID,
		"session_id", req.SessionID,
	)

	filter := bson.M{
		"app_name":   req.AppName,
		"user_id":    req.UserID,
		"session_id": req.SessionID,
	}

	if _, err := s.sessions.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	eventFilter := bson.M{
		"app_name":   req.AppName,
		"session_id": req.SessionID,
	}
	result, err := s.events.DeleteMany(ctx, eventFilter)
	if err != nil {
		return fmt.Errorf("deleting events: %w", err)
	}

	logger.Info("session deleted",
		"app_name", req.AppName,
		"session_id", req.SessionID,
		"events_deleted", result.DeletedCount,
	)

	return nil
}

func (s *Service) AppendEvent(ctx context.Context, sess session.Session, evt *session.Event) error {
	logger := log.FromContext(ctx)

	now := time.Now()
	if evt.Timestamp.IsZero() {
		evt.Timestamp = now
	}
	if evt.ID == "" {
		evt.ID = uuid.New().String()
	}

	logger.Debug("appending event",
		"app_name", sess.AppName(),
		"session_id", sess.ID(),
		"event_id", evt.ID,
		"author", evt.Author,
		"invocation_id", evt.InvocationID,
	)

	ed := eventDoc{
		SessionID:    sess.ID(),
		AppName:      sess.AppName(),
		EventID:      evt.ID,
		InvocationID: evt.InvocationID,
		Author:       evt.Author,
		Branch:       evt.Branch,
		Timestamp:    evt.Timestamp,
	}

	if evt.Content != nil {
		contentBytes, err := json.Marshal(evt.Content)
		if err != nil {
			return fmt.Errorf("marshaling event content: %w", err)
		}
		ed.ContentJSON = contentBytes
	}

	if _, err := s.events.InsertOne(ctx, ed); err != nil {
		logger.Error("failed to insert event",
			"session_id", sess.ID(),
			"event_id", evt.ID,
			"err", err,
		)
		return fmt.Errorf("inserting event: %w", err)
	}

	// Apply state delta from event actions.
	update := bson.M{
		"$set": bson.M{"last_update_time": now},
	}
	if evt.Actions.StateDelta != nil {
		for k, v := range evt.Actions.StateDelta {
			update["$set"].(bson.M)["state."+k] = v
		}
		logger.Debug("applying state delta",
			"session_id", sess.ID(),
			"delta_keys", len(evt.Actions.StateDelta),
		)
	}

	filter := bson.M{
		"app_name":   sess.AppName(),
		"session_id": sess.ID(),
	}
	if _, err := s.sessions.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("updating session: %w", err)
	}

	// Update in-memory session state and events.
	if evt.Actions.StateDelta != nil {
		st := sess.State()
		for k, v := range evt.Actions.StateDelta {
			st.Set(k, v)
		}
	}

	if ms, ok := sess.(*mongoSession); ok {
		ms.events.mu.Lock()
		ms.events.events = append(ms.events.events, evt)
		ms.events.mu.Unlock()
		ms.lastUpdateTime = now
	}

	return nil
}

// mongoSession implements session.Session and carries the Butter-owned title.
type mongoSession struct {
	id             string
	appName        string
	userID         string
	state          *stateImpl
	events         *eventsImpl
	lastUpdateTime time.Time
	title          string
	workspaceID    string
	lastReadAt     *time.Time
}

func (s *mongoSession) ID() string                { return s.id }
func (s *mongoSession) AppName() string           { return s.appName }
func (s *mongoSession) UserID() string            { return s.userID }
func (s *mongoSession) State() session.State      { return s.state }
func (s *mongoSession) Events() session.Events    { return s.events }
func (s *mongoSession) LastUpdateTime() time.Time { return s.lastUpdateTime }

// Title returns the Butter-owned first-class title. Empty means unset.
func (s *mongoSession) Title() string { return s.title }

// WorkspaceID returns the owning workspace. Empty for legacy sessions.
func (s *mongoSession) WorkspaceID() string { return s.workspaceID }

// LastReadAt returns when the session owner last marked the session read.
// Nil means never read.
func (s *mongoSession) LastReadAt() *time.Time { return s.lastReadAt }

// MarkReadResult carries the post-update session snapshot.
type MarkReadResult struct {
	SessionID      string
	AppName        string
	UserID         string
	Title          string
	LastUpdateTime time.Time
	LastReadAt     time.Time
	WorkspaceID    string
}

// MarkRead atomically records a read timestamp on the session. Returns the
// updated session snapshot. The caller validates ownership; this method only
// ensures the session exists.
func (s *Service) MarkRead(ctx context.Context, appName, userID, sessionID string, readAt time.Time) (MarkReadResult, error) {
	filter := bson.M{
		"app_name":   appName,
		"user_id":    userID,
		"session_id": sessionID,
	}
	update := bson.M{"$set": bson.M{"last_read_at": readAt}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc sessionDoc
	if err := s.sessions.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return MarkReadResult{}, fmt.Errorf("%w: %s/%s/%s", ErrSessionNotFound, appName, userID, sessionID)
		}
		return MarkReadResult{}, fmt.Errorf("marking session read: %w", err)
	}
	ra := readAt
	if doc.LastReadAt != nil {
		ra = *doc.LastReadAt
	}
	return MarkReadResult{
		SessionID:      doc.SessionID,
		AppName:        doc.AppName,
		UserID:         doc.UserID,
		Title:          doc.Title,
		LastUpdateTime: doc.LastUpdateTime,
		LastReadAt:     ra,
		WorkspaceID:    doc.WorkspaceID,
	}, nil
}

// stateImpl implements session.State backed by a map.
type stateImpl struct {
	mu   sync.RWMutex
	data map[string]any
}

func newState(data map[string]any) *stateImpl {
	if data == nil {
		data = make(map[string]any)
	}
	return &stateImpl{data: data}
}

func (s *stateImpl) Get(key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return nil, session.ErrStateKeyNotExist
	}
	return v, nil
}

func (s *stateImpl) Set(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *stateImpl) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		for k, v := range s.data {
			if !yield(k, v) {
				return
			}
		}
	}
}

// eventsImpl implements session.Events.
type eventsImpl struct {
	mu     sync.RWMutex
	events []*session.Event
}

func newEvents(events []*session.Event) *eventsImpl {
	if events == nil {
		events = []*session.Event{}
	}
	return &eventsImpl{events: events}
}

func (e *eventsImpl) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		e.mu.RLock()
		snapshot := make([]*session.Event, len(e.events))
		copy(snapshot, e.events)
		e.mu.RUnlock()
		for _, evt := range snapshot {
			if !yield(evt) {
				return
			}
		}
	}
}

func (e *eventsImpl) Len() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.events)
}

func (e *eventsImpl) At(i int) *session.Event {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.events[i]
}
