// Package mongo implements repocache.Repository backed by MongoDB.
package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/repo/repocache"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	snapshotsCollection = "workspace_repo_cache"
	entriesCollection   = "workspace_repo_cache_entries"
	blobsCollection     = "workspace_repo_cache_blob_chunks"
	blobChunkBytes      = 1024 * 1024
)

type snapshotDoc struct {
	ID         string `bson:"_id"`
	SnapshotID string `bson:"snapshot_id"`
	BindingKey string `bson:"binding_key"`
	CommitSHA  string `bson:"commit_sha"`
}

type entryDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	SnapshotID  string `bson:"snapshot_id"`
	ParentPath  string `bson:"parent_path"`
	Path        string `bson:"path"`
	Kind        int32  `bson:"kind"`
	Size        int64  `bson:"size"`
	ContentHash string `bson:"content_hash"`
	Claimed     bool   `bson:"claimed"`
}

type blobChunkDoc struct {
	ID          string `bson:"_id"`
	WorkspaceID string `bson:"workspace_id"`
	SnapshotID  string `bson:"snapshot_id"`
	Path        string `bson:"path"`
	ChunkIndex  int    `bson:"chunk_index"`
	Content     []byte `bson:"content"`
}

// Store implements repocache.Repository backed by three collections. A
// workspace metadata document points at an immutable snapshot ID; entries and
// blob chunks are written first, then the pointer is replaced atomically.
type Store struct {
	snapshots *mongo.Collection
	entries   *mongo.Collection
	blobs     *mongo.Collection
}

var _ repocache.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{
		snapshots: db.Collection(snapshotsCollection),
		entries:   db.Collection(entriesCollection),
		blobs:     db.Collection(blobsCollection),
	}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	indexes := []struct {
		collection *mongo.Collection
		model      mongo.IndexModel
	}{
		{
			collection: s.entries,
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "snapshot_id", Value: 1}, {Key: "parent_path", Value: 1}, {Key: "path", Value: 1}},
				Options: options.Index().SetName("workspace_snapshot_parent_path"),
			},
		},
		{
			collection: s.entries,
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "snapshot_id", Value: 1}, {Key: "path", Value: 1}},
				Options: options.Index().SetName("workspace_snapshot_path").SetUnique(true),
			},
		},
		{
			collection: s.blobs,
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "snapshot_id", Value: 1}, {Key: "path", Value: 1}, {Key: "chunk_index", Value: 1}},
				Options: options.Index().SetName("workspace_snapshot_blob_chunks").SetUnique(true),
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

func parentPath(entryPath string) string {
	for i := len(entryPath) - 1; i >= 0; i-- {
		if entryPath[i] == '/' {
			return entryPath[:i]
		}
	}
	return ""
}

func (s *Store) PutSnapshot(ctx context.Context, workspaceID string, metadata repocache.SnapshotMetadata, entries []*agentsv1.RepoCacheEntry, blobs []repocache.CachedBlob) error {
	old, err := s.findSnapshot(ctx, workspaceID)
	if err != nil && !errors.Is(err, repocache.ErrNotFound) {
		return err
	}

	snapshotID := uuid.NewString()
	entryDocs := make([]any, 0, len(entries))
	for _, entry := range entries {
		entryDocs = append(entryDocs, entryDoc{
			ID:          uuid.NewString(),
			WorkspaceID: workspaceID,
			SnapshotID:  snapshotID,
			ParentPath:  parentPath(entry.GetPath()),
			Path:        entry.GetPath(),
			Kind:        int32(entry.GetKind()),
			Size:        entry.GetSize(),
			ContentHash: entry.GetContentHash(),
			Claimed:     entry.GetClaimed(),
		})
	}
	if len(entryDocs) > 0 {
		if _, err := s.entries.InsertMany(ctx, entryDocs); err != nil {
			_ = s.deleteSnapshotContent(ctx, workspaceID, snapshotID)
			return fmt.Errorf("put repo cache entries (workspace %q): %w", workspaceID, err)
		}
	}

	chunkDocs := make([]any, 0, len(blobs))
	for _, blob := range blobs {
		if len(blob.Content) == 0 {
			chunkDocs = append(chunkDocs, newBlobChunk(workspaceID, snapshotID, blob.Path, 0, nil))
			continue
		}
		chunkIndex := 0
		for offset := 0; offset < len(blob.Content); offset += blobChunkBytes {
			end := min(offset+blobChunkBytes, len(blob.Content))
			chunkDocs = append(chunkDocs, newBlobChunk(workspaceID, snapshotID, blob.Path, chunkIndex, blob.Content[offset:end]))
			chunkIndex++
		}
	}
	if len(chunkDocs) > 0 {
		if _, err := s.blobs.InsertMany(ctx, chunkDocs); err != nil {
			_ = s.deleteSnapshotContent(ctx, workspaceID, snapshotID)
			return fmt.Errorf("put repo cache blobs (workspace %q): %w", workspaceID, err)
		}
	}

	doc := snapshotDoc{
		ID:         workspaceID,
		SnapshotID: snapshotID,
		BindingKey: metadata.BindingKey,
		CommitSHA:  metadata.CommitSHA,
	}
	if _, err := s.snapshots.ReplaceOne(ctx, bson.M{"_id": workspaceID}, doc, options.Replace().SetUpsert(true)); err != nil {
		_ = s.deleteSnapshotContent(ctx, workspaceID, snapshotID)
		return fmt.Errorf("put repo cache snapshot (workspace %q): %w", workspaceID, err)
	}

	if old.SnapshotID != "" && old.SnapshotID != snapshotID {
		_ = s.deleteSnapshotContent(ctx, workspaceID, old.SnapshotID)
	}
	return nil
}

func newBlobChunk(workspaceID, snapshotID, path string, chunkIndex int, content []byte) blobChunkDoc {
	return blobChunkDoc{
		ID:          uuid.NewString(),
		WorkspaceID: workspaceID,
		SnapshotID:  snapshotID,
		Path:        path,
		ChunkIndex:  chunkIndex,
		Content:     content,
	}
}

func (s *Store) findSnapshot(ctx context.Context, workspaceID string) (snapshotDoc, error) {
	var doc snapshotDoc
	err := s.snapshots.FindOne(ctx, bson.M{"_id": workspaceID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return snapshotDoc{}, repocache.ErrNotFound
	}
	if err != nil {
		return snapshotDoc{}, fmt.Errorf("repo cache (workspace %q): %w", workspaceID, err)
	}
	return doc, nil
}

func (s *Store) GetMetadata(ctx context.Context, workspaceID string) (repocache.SnapshotMetadata, error) {
	doc, err := s.findSnapshot(ctx, workspaceID)
	if err != nil {
		return repocache.SnapshotMetadata{}, err
	}
	return repocache.SnapshotMetadata{SnapshotID: doc.SnapshotID, BindingKey: doc.BindingKey, CommitSHA: doc.CommitSHA}, nil
}

func (s *Store) ListEntries(ctx context.Context, workspaceID, snapshotID, dirPath string) ([]*agentsv1.RepoCacheEntry, error) {
	cursor, err := s.entries.Find(ctx, bson.M{
		"workspace_id": workspaceID,
		"snapshot_id":  snapshotID,
		"parent_path":  dirPath,
	}, options.Find().SetSort(bson.D{{Key: "path", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list repo cache entries (workspace %q): %w", workspaceID, err)
	}
	defer cursor.Close(ctx)
	var docs []entryDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode repo cache entries (workspace %q): %w", workspaceID, err)
	}
	out := make([]*agentsv1.RepoCacheEntry, 0, len(docs))
	for _, doc := range docs {
		out = append(out, entryFromDoc(doc))
	}
	return out, nil
}

func (s *Store) GetEntry(ctx context.Context, workspaceID, snapshotID, entryPath string) (*agentsv1.RepoCacheEntry, error) {
	var doc entryDoc
	err := s.entries.FindOne(ctx, bson.M{
		"workspace_id": workspaceID,
		"snapshot_id":  snapshotID,
		"path":         entryPath,
	}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repocache.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get repo cache entry (workspace %q, path %q): %w", workspaceID, entryPath, err)
	}
	return entryFromDoc(doc), nil
}

func entryFromDoc(doc entryDoc) *agentsv1.RepoCacheEntry {
	return &agentsv1.RepoCacheEntry{
		Path:        doc.Path,
		Kind:        agentsv1.RepoCacheEntryKind(doc.Kind),
		Size:        doc.Size,
		ContentHash: doc.ContentHash,
		Claimed:     doc.Claimed,
	}
}

func (s *Store) GetBlob(ctx context.Context, workspaceID, snapshotID, filePath string) ([]byte, error) {
	cursor, err := s.blobs.Find(ctx, bson.M{
		"workspace_id": workspaceID,
		"snapshot_id":  snapshotID,
		"path":         filePath,
	}, options.Find().SetSort(bson.D{{Key: "chunk_index", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("get repo cache blob (workspace %q, path %q): %w", workspaceID, filePath, err)
	}
	defer cursor.Close(ctx)
	var docs []blobChunkDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode repo cache blob (workspace %q, path %q): %w", workspaceID, filePath, err)
	}
	if len(docs) == 0 {
		return nil, repocache.ErrNotFound
	}
	var content []byte
	for index, doc := range docs {
		if doc.ChunkIndex != index {
			return nil, fmt.Errorf("repo cache blob has missing chunk (workspace %q, path %q)", workspaceID, filePath)
		}
		content = append(content, doc.Content...)
	}
	return content, nil
}

func (s *Store) Delete(ctx context.Context, workspaceID string) error {
	if _, err := s.snapshots.DeleteOne(ctx, bson.M{"_id": workspaceID}); err != nil {
		return fmt.Errorf("delete repo cache snapshot (workspace %q): %w", workspaceID, err)
	}
	if _, err := s.entries.DeleteMany(ctx, bson.M{"workspace_id": workspaceID}); err != nil {
		return fmt.Errorf("delete repo cache entries (workspace %q): %w", workspaceID, err)
	}
	if _, err := s.blobs.DeleteMany(ctx, bson.M{"workspace_id": workspaceID}); err != nil {
		return fmt.Errorf("delete repo cache blobs (workspace %q): %w", workspaceID, err)
	}
	return nil
}

func (s *Store) deleteSnapshotContent(ctx context.Context, workspaceID, snapshotID string) error {
	filter := bson.M{"workspace_id": workspaceID, "snapshot_id": snapshotID}
	if _, err := s.entries.DeleteMany(ctx, filter); err != nil {
		return err
	}
	if _, err := s.blobs.DeleteMany(ctx, filter); err != nil {
		return err
	}
	return nil
}
