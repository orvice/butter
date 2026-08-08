// Package mongo implements repobinding.Repository backed by MongoDB. Each
// workspace holds at most one binding, keyed by workspace ID (_id). The
// encrypted PAT lives in its own document field, never inside the protojson
// spec, so decoding a binding can never surface credential material.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const bindingsCollection = "workspace_repo_bindings"

type bindingDoc struct {
	ID                string    `bson:"_id"` // workspace ID
	Spec              string    `bson:"spec"`
	Credential        string    `bson:"credential,omitempty"`
	CredentialUpdated time.Time `bson:"credential_updated_at,omitempty"`
	WebhookSecret     string    `bson:"webhook_secret,omitempty"`
}

// Store implements repobinding.Repository backed by MongoDB.
type Store struct {
	bindings *mongo.Collection
}

var _ repobindingrepo.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{bindings: db.Collection(bindingsCollection)}
}

// EnsureIndexes is a no-op: bindings are keyed by workspace (_id) and the
// flat overlap listing scans the whole (small) collection.
func (s *Store) EnsureIndexes(context.Context) error { return nil }

func notFound(workspaceID string) error {
	return fmt.Errorf("repo binding (workspace %q): %w", workspaceID, repobindingrepo.ErrNotFound)
}

func decode(doc bindingDoc) (*agentsv1.WorkspaceRepoBinding, error) {
	b := &agentsv1.WorkspaceRepoBinding{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), b); err != nil {
		return nil, fmt.Errorf("unmarshal repo binding %q: %w", doc.ID, err)
	}
	b.CredentialSet = doc.Credential != ""
	if !doc.CredentialUpdated.IsZero() {
		b.CredentialUpdatedAt = timestamppb.New(doc.CredentialUpdated)
	} else {
		b.CredentialUpdatedAt = nil
	}
	b.WebhookSecretSet = doc.WebhookSecret != ""
	return b, nil
}

func (s *Store) findDoc(ctx context.Context, workspaceID string) (bindingDoc, error) {
	var doc bindingDoc
	err := s.bindings.FindOne(ctx, bson.M{"_id": workspaceID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return bindingDoc{}, notFound(workspaceID)
	}
	if err != nil {
		return bindingDoc{}, fmt.Errorf("repo binding (workspace %q): %w", workspaceID, err)
	}
	return doc, nil
}

func (s *Store) Get(ctx context.Context, workspaceID string) (*agentsv1.WorkspaceRepoBinding, error) {
	doc, err := s.findDoc(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return decode(doc)
}

func (s *Store) Put(ctx context.Context, workspaceID string, binding *agentsv1.WorkspaceRepoBinding) (*agentsv1.WorkspaceRepoBinding, error) {
	stored := proto.Clone(binding).(*agentsv1.WorkspaceRepoBinding)
	stored.WorkspaceId = workspaceID
	// Credential fields are repo-owned and derived on read; never persist
	// caller-provided values into the spec.
	stored.CredentialSet = false
	stored.CredentialUpdatedAt = nil
	now := timestamppb.New(time.Now().UTC())
	stored.UpdatedAt = now
	stored.CreatedAt = now
	prev, err := s.findDoc(ctx, workspaceID)
	switch {
	case err == nil:
		prevBinding, decErr := decode(prev)
		if decErr != nil {
			return nil, decErr
		}
		stored.CreatedAt = prevBinding.GetCreatedAt()
	case errors.Is(err, repobindingrepo.ErrNotFound):
		// first write
	default:
		return nil, err
	}
	spec, err := protojson.Marshal(stored)
	if err != nil {
		return nil, fmt.Errorf("marshal repo binding (workspace %q): %w", workspaceID, err)
	}
	_, err = s.bindings.UpdateOne(ctx,
		bson.M{"_id": workspaceID},
		bson.M{"$set": bson.M{"spec": string(spec)}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return nil, fmt.Errorf("put repo binding (workspace %q): %w", workspaceID, err)
	}
	return s.Get(ctx, workspaceID)
}

func (s *Store) Delete(ctx context.Context, workspaceID string) error {
	res, err := s.bindings.DeleteOne(ctx, bson.M{"_id": workspaceID})
	if err != nil {
		return fmt.Errorf("delete repo binding (workspace %q): %w", workspaceID, err)
	}
	if res.DeletedCount == 0 {
		return notFound(workspaceID)
	}
	return nil
}

func (s *Store) SetCredential(ctx context.Context, workspaceID, ciphertext string) error {
	update := bson.M{"$set": bson.M{"credential": ciphertext, "credential_updated_at": time.Now().UTC()}}
	if ciphertext == "" {
		update = bson.M{"$unset": bson.M{"credential": "", "credential_updated_at": ""}}
	}
	res, err := s.bindings.UpdateOne(ctx, bson.M{"_id": workspaceID}, update)
	if err != nil {
		return fmt.Errorf("set repo binding credential (workspace %q): %w", workspaceID, err)
	}
	if res.MatchedCount == 0 {
		return notFound(workspaceID)
	}
	return nil
}

func (s *Store) GetCredential(ctx context.Context, workspaceID string) (string, error) {
	doc, err := s.findDoc(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	if doc.Credential == "" {
		return "", fmt.Errorf("repo binding (workspace %q): %w", workspaceID, repobindingrepo.ErrNoCredential)
	}
	return doc.Credential, nil
}

func (s *Store) SetWebhookSecret(ctx context.Context, workspaceID, ciphertext string) error {
	update := bson.M{"$set": bson.M{"webhook_secret": ciphertext}}
	if ciphertext == "" {
		update = bson.M{"$unset": bson.M{"webhook_secret": ""}}
	}
	res, err := s.bindings.UpdateOne(ctx, bson.M{"_id": workspaceID}, update)
	if err != nil {
		return fmt.Errorf("set webhook secret (workspace %q): %w", workspaceID, err)
	}
	if res.MatchedCount == 0 {
		return notFound(workspaceID)
	}
	return nil
}

func (s *Store) GetWebhookSecret(ctx context.Context, workspaceID string) (string, error) {
	doc, err := s.findDoc(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	if doc.WebhookSecret == "" {
		return "", fmt.Errorf("repo binding (workspace %q): %w", workspaceID, repobindingrepo.ErrNoCredential)
	}
	return doc.WebhookSecret, nil
}

func (s *Store) ListAcrossWorkspaces(ctx context.Context) ([]*agentsv1.WorkspaceRepoBinding, error) {
	cursor, err := s.bindings.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list repo bindings: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []bindingDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode repo bindings: %w", err)
	}
	out := make([]*agentsv1.WorkspaceRepoBinding, 0, len(docs))
	for _, doc := range docs {
		b, err := decode(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetWorkspaceId() < out[j].GetWorkspaceId() })
	return out, nil
}
