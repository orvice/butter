// Package mongo implements repocache.Repository backed by MongoDB (issue
// #215). Each workspace's cached tree snapshot is stored as a single
// document keyed by workspace ID, holding the commit SHA, tree entries,
// and Markdown blob content. The snapshot is replaced atomically on every
// successful sync.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/repo/repocache"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const collection = "workspace_repo_cache"

type entryDoc struct {
	Path        string `bson:"path"`
	Kind        int32  `bson:"kind"`
	Size        int64  `bson:"size"`
	ContentHash string `bson:"content_hash"`
	Claimed     bool   `bson:"claimed,omitempty"`
}

type blobDoc struct {
	Path    string `bson:"path"`
	Content []byte `bson:"content"`
}

type snapshotDoc struct {
	ID        string     `bson:"_id"`
	CommitSHA string     `bson:"commit_sha"`
	Entries   []entryDoc `bson:"entries"`
	Blobs     []blobDoc  `bson:"blobs"`
}

// Store implements repocache.Repository backed by MongoDB.
type Store struct {
	coll *mongo.Collection
}

var _ repocache.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collection)}
}

func (s *Store) PutSnapshot(ctx context.Context, workspaceID, commitSHA string, entries []*agentsv1.RepoCacheEntry, blobs []repocache.CachedBlob) error {
	edocs := make([]entryDoc, len(entries))
	for i, e := range entries {
		edocs[i] = entryDoc{
			Path:        e.GetPath(),
			Kind:        int32(e.GetKind()),
			Size:        e.GetSize(),
			ContentHash: e.GetContentHash(),
			Claimed:     e.GetClaimed(),
		}
	}
	bdocs := make([]blobDoc, len(blobs))
	for i, b := range blobs {
		bdocs[i] = blobDoc{Path: b.Path, Content: b.Content}
	}
	doc := snapshotDoc{
		ID:        workspaceID,
		CommitSHA: commitSHA,
		Entries:   edocs,
		Blobs:     bdocs,
	}
	_, err := s.coll.UpdateOne(ctx,
		bson.M{"_id": workspaceID},
		bson.M{"$set": doc},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("put repo cache (workspace %q): %w", workspaceID, err)
	}
	return nil
}

func (s *Store) findDoc(ctx context.Context, workspaceID string) (snapshotDoc, error) {
	var doc snapshotDoc
	err := s.coll.FindOne(ctx, bson.M{"_id": workspaceID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return snapshotDoc{}, repocache.ErrNotFound
	}
	if err != nil {
		return snapshotDoc{}, fmt.Errorf("repo cache (workspace %q): %w", workspaceID, err)
	}
	return doc, nil
}

func (s *Store) GetCommitSHA(ctx context.Context, workspaceID string) (string, error) {
	doc, err := s.findDoc(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	return doc.CommitSHA, nil
}

func (s *Store) ListEntries(ctx context.Context, workspaceID, dirPath string) ([]*agentsv1.RepoCacheEntry, error) {
	doc, err := s.findDoc(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	dirPath = strings.TrimRight(dirPath, "/")
	var out []*agentsv1.RepoCacheEntry
	for _, e := range doc.Entries {
		parent := path.Dir(e.Path)
		if parent == "." {
			parent = ""
		}
		if parent == dirPath {
			out = append(out, &agentsv1.RepoCacheEntry{
				Path:        e.Path,
				Kind:        agentsv1.RepoCacheEntryKind(e.Kind),
				Size:        e.Size,
				ContentHash: e.ContentHash,
				Claimed:     e.Claimed,
			})
		}
	}
	return out, nil
}

func (s *Store) GetEntry(ctx context.Context, workspaceID, entryPath string) (*agentsv1.RepoCacheEntry, error) {
	doc, err := s.findDoc(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	for _, e := range doc.Entries {
		if e.Path == entryPath {
			return &agentsv1.RepoCacheEntry{
				Path:        e.Path,
				Kind:        agentsv1.RepoCacheEntryKind(e.Kind),
				Size:        e.Size,
				ContentHash: e.ContentHash,
				Claimed:     e.Claimed,
			}, nil
		}
	}
	return nil, repocache.ErrNotFound
}

func (s *Store) GetBlob(ctx context.Context, workspaceID, filePath string) ([]byte, error) {
	doc, err := s.findDoc(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	for _, b := range doc.Blobs {
		if b.Path == filePath {
			cp := make([]byte, len(b.Content))
			copy(cp, b.Content)
			return cp, nil
		}
	}
	return nil, repocache.ErrNotFound
}

func (s *Store) Delete(ctx context.Context, workspaceID string) error {
	_, err := s.coll.DeleteOne(ctx, bson.M{"_id": workspaceID})
	if err != nil {
		return fmt.Errorf("delete repo cache (workspace %q): %w", workspaceID, err)
	}
	return nil
}
