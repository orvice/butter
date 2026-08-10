// Package mongo implements agentcontent.Repository backed by MongoDB.
package mongo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/agentcontent"
	agentcontentrepo "go.orx.me/apps/butter/internal/repo/agentcontent"
)

const (
	snapshotsCollection = "agent_content_snapshots"
	entriesCollection   = "agent_content_snapshot_entries"
)

type snapshotDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	CommitSHA   string `bson:"commit_sha"`
	EntryCount  int    `bson:"entry_count"`
}

type entryDoc struct {
	ID                string `bson:"_id"`
	WorkspaceID       string `bson:"workspace_id"`
	CommitSHA         string `bson:"commit_sha"`
	AgentID           string `bson:"agent_id"`
	Description       string `bson:"description"`
	Instruction       string `bson:"instruction"`
	GlobalInstruction string `bson:"global_instruction"`
}

// Store persists immutable, revision-addressed Agent Content snapshots. The
// snapshot metadata document is written after its entries, so readers never
// observe a revision until all content has been stored.
type Store struct {
	snapshots *mongo.Collection
	entries   *mongo.Collection
}

var _ agentcontentrepo.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{
		snapshots: db.Collection(snapshotsCollection),
		entries:   db.Collection(entriesCollection),
	}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	indexes := []struct {
		collection *mongo.Collection
		model      mongo.IndexModel
	}{
		{
			collection: s.snapshots,
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "commit_sha", Value: 1}},
				Options: options.Index().SetName("workspace_commit").SetUnique(true),
			},
		},
		{
			collection: s.entries,
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "commit_sha", Value: 1}, {Key: "agent_id", Value: 1}},
				Options: options.Index().SetName("workspace_commit_agent").SetUnique(true),
			},
		},
	}
	for _, index := range indexes {
		if _, err := index.collection.Indexes().CreateOne(ctx, index.model); err != nil {
			return fmt.Errorf("create %s index: %w", index.collection.Name(), err)
		}
	}
	return nil
}

func (s *Store) PutSnapshot(ctx context.Context, workspaceID string, snapshot agentcontent.Snapshot) error {
	if workspaceID == "" {
		return errors.New("workspace ID is required")
	}
	if snapshot.CommitSHA == "" {
		return errors.New("snapshot commit SHA is required")
	}

	for agentID, content := range snapshot.Entries {
		doc := entryDoc{
			ID:                documentID(workspaceID, snapshot.CommitSHA, agentID),
			WorkspaceID:       workspaceID,
			CommitSHA:         snapshot.CommitSHA,
			AgentID:           agentID,
			Description:       content.Description,
			Instruction:       content.Instruction,
			GlobalInstruction: content.GlobalInstruction,
		}
		if _, err := s.entries.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true)); err != nil {
			return fmt.Errorf("put Agent Content entry (workspace %q, commit %q, agent %q): %w", workspaceID, snapshot.CommitSHA, agentID, err)
		}
	}

	doc := snapshotDoc{
		ID:          documentID(workspaceID, snapshot.CommitSHA),
		WorkspaceID: workspaceID,
		CommitSHA:   snapshot.CommitSHA,
		EntryCount:  len(snapshot.Entries),
	}
	if _, err := s.snapshots.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true)); err != nil {
		return fmt.Errorf("put Agent Content snapshot (workspace %q, commit %q): %w", workspaceID, snapshot.CommitSHA, err)
	}
	return nil
}

func (s *Store) GetSnapshot(ctx context.Context, workspaceID, commitSHA string) (agentcontent.Snapshot, error) {
	var metadata snapshotDoc
	err := s.snapshots.FindOne(ctx, bson.M{"_id": documentID(workspaceID, commitSHA)}).Decode(&metadata)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return agentcontent.Snapshot{}, agentcontentrepo.ErrNotFound
	}
	if err != nil {
		return agentcontent.Snapshot{}, fmt.Errorf("get Agent Content snapshot (workspace %q, commit %q): %w", workspaceID, commitSHA, err)
	}

	cursor, err := s.entries.Find(ctx, bson.M{"workspace_id": workspaceID, "commit_sha": commitSHA})
	if err != nil {
		return agentcontent.Snapshot{}, fmt.Errorf("list Agent Content entries (workspace %q, commit %q): %w", workspaceID, commitSHA, err)
	}
	defer cursor.Close(ctx)

	var docs []entryDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return agentcontent.Snapshot{}, fmt.Errorf("decode Agent Content entries (workspace %q, commit %q): %w", workspaceID, commitSHA, err)
	}
	if len(docs) != metadata.EntryCount {
		return agentcontent.Snapshot{}, fmt.Errorf("Agent Content snapshot incomplete (workspace %q, commit %q): have %d entries, want %d: %w",
			workspaceID, commitSHA, len(docs), metadata.EntryCount, agentcontentrepo.ErrNotFound)
	}

	entries := make(map[string]agentcontent.AgentContent, len(docs))
	for _, doc := range docs {
		entries[doc.AgentID] = agentcontent.AgentContent{
			AgentID:           doc.AgentID,
			Description:       doc.Description,
			Instruction:       doc.Instruction,
			GlobalInstruction: doc.GlobalInstruction,
		}
	}
	return agentcontent.Snapshot{CommitSHA: metadata.CommitSHA, Entries: entries}, nil
}

func (s *Store) PruneSnapshots(ctx context.Context, workspaceID, keepCommitSHA string) error {
	filter := bson.M{"workspace_id": workspaceID, "commit_sha": bson.M{"$ne": keepCommitSHA}}
	if _, err := s.snapshots.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("prune Agent Content snapshots (workspace %q): %w", workspaceID, err)
	}
	if _, err := s.entries.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("prune Agent Content entries (workspace %q): %w", workspaceID, err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, workspaceID string) error {
	filter := bson.M{"workspace_id": workspaceID}
	if _, err := s.snapshots.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("delete Agent Content snapshots (workspace %q): %w", workspaceID, err)
	}
	if _, err := s.entries.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("delete Agent Content entries (workspace %q): %w", workspaceID, err)
	}
	return nil
}

func documentID(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
