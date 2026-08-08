// Package mongo implements githost.Repository backed by MongoDB.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	githostrepo "go.orx.me/apps/butter/internal/repo/githost"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const gitHostsCollection = "git_hosts"

type hostDoc struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
	Spec string `bson:"spec"`
}

// Store implements githost.Repository backed by MongoDB.
type Store struct {
	hosts *mongo.Collection
}

var _ githostrepo.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{hosts: db.Collection(gitHostsCollection)}
}

// EnsureIndexes is a no-op: hosts are keyed by _id and the collection is
// small enough that List needs no secondary index.
func (s *Store) EnsureIndexes(context.Context) error { return nil }

func mapError(id string, err error) error {
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("git host %q: %w", id, githostrepo.ErrAlreadyExists)
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("git host %q: %w", id, githostrepo.ErrNotFound)
	}
	return fmt.Errorf("git host %q: %w", id, err)
}

func decode(doc hostDoc) (*agentsv1.GitHost, error) {
	h := &agentsv1.GitHost{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), h); err != nil {
		return nil, fmt.Errorf("unmarshal git host %q: %w", doc.ID, err)
	}
	return h, nil
}

func encode(h *agentsv1.GitHost) (hostDoc, error) {
	spec, err := protojson.Marshal(h)
	if err != nil {
		return hostDoc{}, fmt.Errorf("marshal git host %q: %w", h.GetId(), err)
	}
	return hostDoc{ID: h.GetId(), Name: h.GetName(), Spec: string(spec)}, nil
}

func (s *Store) List(ctx context.Context) ([]*agentsv1.GitHost, error) {
	cursor, err := s.hosts.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list git hosts: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []hostDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode git hosts: %w", err)
	}
	out := make([]*agentsv1.GitHost, 0, len(docs))
	for _, doc := range docs {
		h, err := decode(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetName() < out[j].GetName() })
	return out, nil
}

func (s *Store) Get(ctx context.Context, id string) (*agentsv1.GitHost, error) {
	var doc hostDoc
	if err := s.hosts.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		return nil, mapError(id, err)
	}
	return decode(doc)
}

func (s *Store) Create(ctx context.Context, host *agentsv1.GitHost) (*agentsv1.GitHost, error) {
	clone := proto.Clone(host).(*agentsv1.GitHost)
	now := timestamppb.New(time.Now().UTC())
	clone.CreatedAt = now
	clone.UpdatedAt = now
	doc, err := encode(clone)
	if err != nil {
		return nil, err
	}
	if _, err := s.hosts.InsertOne(ctx, doc); err != nil {
		return nil, mapError(clone.GetId(), err)
	}
	return clone, nil
}

func (s *Store) Update(ctx context.Context, host *agentsv1.GitHost) (*agentsv1.GitHost, error) {
	prev, err := s.Get(ctx, host.GetId())
	if err != nil {
		return nil, err
	}
	clone := proto.Clone(host).(*agentsv1.GitHost)
	clone.CreatedAt = prev.GetCreatedAt()
	clone.UpdatedAt = timestamppb.New(time.Now().UTC())
	doc, err := encode(clone)
	if err != nil {
		return nil, err
	}
	res, err := s.hosts.ReplaceOne(ctx, bson.M{"_id": clone.GetId()}, doc)
	if err != nil {
		return nil, mapError(clone.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("git host %q: %w", clone.GetId(), githostrepo.ErrNotFound)
	}
	return clone, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.hosts.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return mapError(id, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("git host %q: %w", id, githostrepo.ErrNotFound)
	}
	return nil
}
