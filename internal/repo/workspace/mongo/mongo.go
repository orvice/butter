package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/types/known/timestamppb"

	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	workspacesCollection = "workspaces"
	membersCollection    = "workspace_members"
)

type workspaceDoc struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Slug        string    `bson:"slug"`
	Description string    `bson:"description,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

type memberDoc struct {
	WorkspaceID string    `bson:"workspace_id"`
	UserID      string    `bson:"user_id"`
	Role        string    `bson:"role,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
}

type Store struct {
	workspaces *mongo.Collection
	members    *mongo.Collection
}

func New(db *mongo.Database) *Store {
	return &Store{
		workspaces: db.Collection(workspacesCollection),
		members:    db.Collection(membersCollection),
	}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.workspaces.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create workspaces.slug index: %w", err)
	}
	_, err = s.members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create workspace_members compound index: %w", err)
	}
	_, err = s.members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
	if err != nil {
		return fmt.Errorf("create workspace_members.user_id index: %w", err)
	}
	return nil
}

func (s *Store) ListWorkspaces(ctx context.Context) ([]*agentsv1.Workspace, error) {
	cursor, err := s.workspaces.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "slug", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer cursor.Close(ctx)
	var out []*agentsv1.Workspace
	for cursor.Next(ctx) {
		var doc workspaceDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode workspace: %w", err)
		}
		out = append(out, workspaceToProto(&doc))
	}
	return out, cursor.Err()
}

func (s *Store) GetWorkspace(ctx context.Context, id string) (*agentsv1.Workspace, error) {
	var doc workspaceDoc
	err := s.workspaces.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, workspacerepo.ErrNotFound
		}
		return nil, fmt.Errorf("get workspace: %w", err)
	}
	return workspaceToProto(&doc), nil
}

func (s *Store) GetWorkspaceBySlug(ctx context.Context, slug string) (*agentsv1.Workspace, error) {
	var doc workspaceDoc
	err := s.workspaces.FindOne(ctx, bson.M{"slug": slug}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, workspacerepo.ErrNotFound
		}
		return nil, fmt.Errorf("get workspace by slug: %w", err)
	}
	return workspaceToProto(&doc), nil
}

func (s *Store) CreateWorkspace(ctx context.Context, ws *agentsv1.Workspace) (*agentsv1.Workspace, error) {
	doc := protoToWorkspace(ws)
	if _, err := s.workspaces.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, workspacerepo.ErrAlreadyExists
		}
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	return workspaceToProto(&doc), nil
}

func (s *Store) UpdateWorkspace(ctx context.Context, ws *agentsv1.Workspace) (*agentsv1.Workspace, error) {
	doc := protoToWorkspace(ws)
	res, err := s.workspaces.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, workspacerepo.ErrAlreadyExists
		}
		return nil, fmt.Errorf("update workspace: %w", err)
	}
	if res.MatchedCount == 0 {
		return nil, workspacerepo.ErrNotFound
	}
	return workspaceToProto(&doc), nil
}

func (s *Store) DeleteWorkspace(ctx context.Context, id string) error {
	res, err := s.workspaces.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	if res.DeletedCount == 0 {
		return workspacerepo.ErrNotFound
	}
	if _, err := s.members.DeleteMany(ctx, bson.M{"workspace_id": id}); err != nil {
		return fmt.Errorf("delete workspace members: %w", err)
	}
	return nil
}

func (s *Store) CountWorkspaces(ctx context.Context) (int64, error) {
	n, err := s.workspaces.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("count workspaces: %w", err)
	}
	return n, nil
}

func (s *Store) ListMembers(ctx context.Context, workspaceID string) ([]*agentsv1.WorkspaceMember, error) {
	cursor, err := s.members.Find(ctx, bson.M{"workspace_id": workspaceID}, options.Find().SetSort(bson.D{{Key: "user_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer cursor.Close(ctx)
	var out []*agentsv1.WorkspaceMember
	for cursor.Next(ctx) {
		var doc memberDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode member: %w", err)
		}
		out = append(out, memberToProto(&doc))
	}
	return out, cursor.Err()
}

func (s *Store) ListMembershipsForUser(ctx context.Context, userID string) ([]*agentsv1.WorkspaceMember, error) {
	cursor, err := s.members.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "workspace_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list memberships for user: %w", err)
	}
	defer cursor.Close(ctx)
	var out []*agentsv1.WorkspaceMember
	for cursor.Next(ctx) {
		var doc memberDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode member: %w", err)
		}
		out = append(out, memberToProto(&doc))
	}
	return out, cursor.Err()
}

func (s *Store) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	n, err := s.members.CountDocuments(ctx, bson.M{"workspace_id": workspaceID, "user_id": userID})
	if err != nil {
		return false, fmt.Errorf("is member: %w", err)
	}
	return n > 0, nil
}

func (s *Store) GetMember(ctx context.Context, workspaceID, userID string) (*agentsv1.WorkspaceMember, error) {
	var doc memberDoc
	err := s.members.FindOne(ctx, bson.M{"workspace_id": workspaceID, "user_id": userID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, workspacerepo.ErrNotFound
		}
		return nil, fmt.Errorf("get member: %w", err)
	}
	return memberToProto(&doc), nil
}

func (s *Store) AddMember(ctx context.Context, member *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error) {
	doc := protoToMember(member)
	if _, err := s.members.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, workspacerepo.ErrAlreadyExists
		}
		return nil, fmt.Errorf("add member: %w", err)
	}
	return memberToProto(&doc), nil
}

func (s *Store) UpdateMember(ctx context.Context, member *agentsv1.WorkspaceMember) (*agentsv1.WorkspaceMember, error) {
	doc := protoToMember(member)
	res := s.members.FindOneAndUpdate(
		ctx,
		bson.M{"workspace_id": doc.WorkspaceID, "user_id": doc.UserID},
		bson.M{"$set": bson.M{"role": doc.Role}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var updated memberDoc
	if err := res.Decode(&updated); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, workspacerepo.ErrNotFound
		}
		return nil, fmt.Errorf("update member: %w", err)
	}
	return memberToProto(&updated), nil
}

func (s *Store) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	res, err := s.members.DeleteOne(ctx, bson.M{"workspace_id": workspaceID, "user_id": userID})
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if res.DeletedCount == 0 {
		return workspacerepo.ErrNotFound
	}
	return nil
}

func protoToWorkspace(ws *agentsv1.Workspace) workspaceDoc {
	created := ws.GetCreatedAt().AsTime()
	updated := ws.GetUpdatedAt().AsTime()
	return workspaceDoc{
		ID:          ws.GetId(),
		Name:        ws.GetName(),
		Slug:        ws.GetSlug(),
		Description: ws.GetDescription(),
		CreatedAt:   created,
		UpdatedAt:   updated,
	}
}

func workspaceToProto(doc *workspaceDoc) *agentsv1.Workspace {
	return &agentsv1.Workspace{
		Id:          doc.ID,
		Name:        doc.Name,
		Slug:        doc.Slug,
		Description: doc.Description,
		CreatedAt:   timestamppb.New(doc.CreatedAt),
		UpdatedAt:   timestamppb.New(doc.UpdatedAt),
	}
}

func protoToMember(m *agentsv1.WorkspaceMember) memberDoc {
	return memberDoc{
		WorkspaceID: m.GetWorkspaceId(),
		UserID:      m.GetUserId(),
		Role:        m.GetRole(),
		CreatedAt:   m.GetCreatedAt().AsTime(),
	}
}

func memberToProto(doc *memberDoc) *agentsv1.WorkspaceMember {
	return &agentsv1.WorkspaceMember{
		WorkspaceId: doc.WorkspaceID,
		UserId:      doc.UserID,
		Role:        doc.Role,
		CreatedAt:   timestamppb.New(doc.CreatedAt),
	}
}
