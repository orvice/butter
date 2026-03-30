package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

const (
	memoriesCollection = "adk_memories"
)

// memoryDoc is the MongoDB document for a memory entry.
type memoryDoc struct {
	ID             string         `bson:"_id"`
	AppName        string         `bson:"app_name"`
	UserID         string         `bson:"user_id"`
	SessionID      string         `bson:"session_id"`
	Author         string         `bson:"author"`
	ContentJSON    []byte         `bson:"content_json"`
	ContentText    string         `bson:"content_text"`
	Timestamp      time.Time      `bson:"timestamp"`
	CustomMetadata map[string]any `bson:"custom_metadata,omitempty"`
}

// Service implements memory.Service backed by MongoDB.
type Service struct {
	memories *mongo.Collection
}

// New creates a new MongoDB memory service and ensures indexes.
func New(ctx context.Context, db *mongo.Database) (*Service, error) {
	logger := log.FromContext(ctx)
	logger.Info("initializing mongo memory service", "database", db.Name())

	s := &Service{
		memories: db.Collection(memoriesCollection),
	}
	if err := s.ensureIndexes(ctx); err != nil {
		return nil, fmt.Errorf("ensuring memory indexes: %w", err)
	}

	logger.Info("mongo memory service ready", "collection", memoriesCollection)
	return s, nil
}

func (s *Service) ensureIndexes(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Debug("ensuring memory mongodb indexes")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "app_name", Value: 1},
				{Key: "user_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "app_name", Value: 1},
				{Key: "session_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "content_text", Value: "text"},
			},
		},
	}

	if _, err := s.memories.Indexes().CreateMany(ctx, indexes); err != nil {
		return err
	}

	logger.Debug("memory mongodb indexes ensured")
	return nil
}

// AddSessionToMemory ingests a session's LLM response events into memory.
// Can be called multiple times for the same session — existing entries are replaced.
func (s *Service) AddSessionToMemory(ctx context.Context, sess session.Session) error {
	logger := log.FromContext(ctx)

	appName := sess.AppName()
	userID := sess.UserID()
	sessionID := sess.ID()

	logger.Debug("adding session to memory",
		"app_name", appName,
		"user_id", userID,
		"session_id", sessionID,
	)

	// Delete existing entries for this session to support re-ingestion.
	deleteFilter := bson.M{
		"app_name":   appName,
		"session_id": sessionID,
	}
	if _, err := s.memories.DeleteMany(ctx, deleteFilter); err != nil {
		return fmt.Errorf("clearing existing memory entries: %w", err)
	}

	var docs []any
	for event := range sess.Events().All() {
		if event.LLMResponse.Content == nil {
			continue
		}

		text := extractText(event.LLMResponse.Content)
		if text == "" {
			continue
		}

		contentJSON, err := json.Marshal(event.LLMResponse.Content)
		if err != nil {
			logger.Warn("failed to marshal memory content",
				"event_id", event.ID,
				"err", err,
			)
			continue
		}

		docs = append(docs, memoryDoc{
			ID:             event.ID,
			AppName:        appName,
			UserID:         userID,
			SessionID:      sessionID,
			Author:         event.Author,
			ContentJSON:    contentJSON,
			ContentText:    text,
			Timestamp:      event.Timestamp,
			CustomMetadata: event.CustomMetadata,
		})
	}

	if len(docs) == 0 {
		logger.Debug("no memory entries to add", "session_id", sessionID)
		return nil
	}

	if _, err := s.memories.InsertMany(ctx, docs); err != nil {
		return fmt.Errorf("inserting memory entries: %w", err)
	}

	logger.Debug("session added to memory",
		"session_id", sessionID,
		"entry_count", len(docs),
	)
	return nil
}

// SearchMemory returns memory entries matching the query using MongoDB text search.
func (s *Service) SearchMemory(ctx context.Context, req *memory.SearchRequest) (*memory.SearchResponse, error) {
	logger := log.FromContext(ctx)

	logger.Debug("searching memory",
		"app_name", req.AppName,
		"user_id", req.UserID,
		"query_len", len(req.Query),
	)

	filter := bson.M{
		"app_name": req.AppName,
		"user_id":  req.UserID,
		"$text":    bson.M{"$search": req.Query},
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}}).
		SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}}).
		SetLimit(50)

	cursor, err := s.memories.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("searching memory: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []memoryDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decoding memory results: %w", err)
	}

	entries := make([]memory.Entry, 0, len(docs))
	for _, doc := range docs {
		var content genai.Content
		if err := json.Unmarshal(doc.ContentJSON, &content); err != nil {
			logger.Warn("failed to unmarshal memory content",
				"memory_id", doc.ID,
				"err", err,
			)
			continue
		}

		entries = append(entries, memory.Entry{
			ID:             doc.ID,
			Content:        &content,
			Author:         doc.Author,
			Timestamp:      doc.Timestamp,
			CustomMetadata: doc.CustomMetadata,
		})
	}

	logger.Debug("memory search complete",
		"results", len(entries),
	)

	return &memory.SearchResponse{Memories: entries}, nil
}

// extractText concatenates all text parts from a genai.Content.
func extractText(content *genai.Content) string {
	var b strings.Builder
	for _, part := range content.Parts {
		if part.Text != "" {
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(part.Text)
		}
	}
	return b.String()
}
