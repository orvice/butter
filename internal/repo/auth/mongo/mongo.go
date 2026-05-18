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

	"go.orx.me/apps/butter/internal/repo/auth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	usersCollection    = "users"
	sessionsCollection = "auth_sessions"
)

type userDoc struct {
	ID           string    `bson:"_id"`
	Username     string    `bson:"username"`
	DisplayName  string    `bson:"display_name,omitempty"`
	AvatarURL    string    `bson:"avatar_url,omitempty"`
	PasswordHash string    `bson:"password_hash"`
	Role         string    `bson:"role,omitempty"`
	Disabled     bool      `bson:"disabled,omitempty"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

type sessionDoc struct {
	ID         string    `bson:"_id"`
	UserID     string    `bson:"user_id"`
	TokenHash  string    `bson:"token_hash"`
	CreatedAt  time.Time `bson:"created_at"`
	ExpiresAt  time.Time `bson:"expires_at"`
	LastUsedAt time.Time `bson:"last_used_at,omitempty"`
	Revoked    bool      `bson:"revoked,omitempty"`
}

type Store struct {
	users    *mongo.Collection
	sessions *mongo.Collection
}

func New(db *mongo.Database) *Store {
	return &Store{
		users:    db.Collection(usersCollection),
		sessions: db.Collection(sessionsCollection),
	}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create users username index: %w", err)
	}

	_, err = s.sessions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "token_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create auth_sessions token_hash index: %w", err)
	}

	_, err = s.sessions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		return fmt.Errorf("create auth_sessions expires_at ttl index: %w", err)
	}

	return nil
}

func (s *Store) CountUsers(ctx context.Context) (int64, error) {
	count, err := s.users.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

func (s *Store) ListUsers(ctx context.Context) ([]*agentsv1.User, error) {
	cursor, err := s.users.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*agentsv1.User
	for cursor.Next(ctx) {
		var doc userDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode user: %w", err)
		}
		users = append(users, userToProto(&doc))
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func (s *Store) CreateUser(ctx context.Context, user *agentsv1.User, passwordHash string) error {
	doc := userDoc{
		ID:           user.GetId(),
		Username:     user.GetUsername(),
		DisplayName:  user.GetDisplayName(),
		AvatarURL:    user.GetAvatarUrl(),
		PasswordHash: passwordHash,
		Role:         user.GetRole(),
		Disabled:     user.GetDisabled(),
		CreatedAt:    user.GetCreatedAt().AsTime(),
		UpdatedAt:    user.GetUpdatedAt().AsTime(),
	}
	if _, err := s.users.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return auth.ErrUserAlreadyExists
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (s *Store) UpdateUserProfile(ctx context.Context, id string, displayName, avatarURL string, updatedAt time.Time) (*agentsv1.User, error) {
	res := s.users.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"display_name": displayName,
			"avatar_url":   avatarURL,
			"updated_at":   updatedAt,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var doc userDoc
	if err := res.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("update user profile: %w", err)
	}
	return userToProto(&doc), nil
}

func (s *Store) UpdateUserPassword(ctx context.Context, id string, passwordHash string, updatedAt time.Time) (*agentsv1.User, error) {
	res := s.users.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"password_hash": passwordHash, "updated_at": updatedAt}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var doc userDoc
	if err := res.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("update user password: %w", err)
	}
	return userToProto(&doc), nil
}

func (s *Store) SetUserDisabled(ctx context.Context, id string, disabled bool, updatedAt time.Time) (*agentsv1.User, error) {
	res := s.users.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"disabled": disabled, "updated_at": updatedAt}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var doc userDoc
	if err := res.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("set user disabled: %w", err)
	}
	return userToProto(&doc), nil
}

func (s *Store) FindUserByUsername(ctx context.Context, username string) (*agentsv1.User, string, error) {
	var doc userDoc
	err := s.users.FindOne(ctx, bson.M{"username": username}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, "", auth.ErrUserNotFound
		}
		return nil, "", fmt.Errorf("find user by username: %w", err)
	}
	return userToProto(&doc), doc.PasswordHash, nil
}

func (s *Store) GetUser(ctx context.Context, id string) (*agentsv1.User, error) {
	user, _, err := s.FindUserByID(ctx, id)
	return user, err
}

func (s *Store) FindUserByID(ctx context.Context, id string) (*agentsv1.User, string, error) {
	var doc userDoc
	err := s.users.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, "", auth.ErrUserNotFound
		}
		return nil, "", fmt.Errorf("find user by id: %w", err)
	}
	return userToProto(&doc), doc.PasswordHash, nil
}

func (s *Store) CreateSession(ctx context.Context, session *auth.Session) error {
	doc := sessionDoc{
		ID:         session.ID,
		UserID:     session.UserID,
		TokenHash:  session.TokenHash,
		CreatedAt:  session.CreatedAt,
		ExpiresAt:  session.ExpiresAt,
		LastUsedAt: session.LastUsedAt,
		Revoked:    session.Revoked,
	}
	if _, err := s.sessions.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert auth_session: %w", err)
	}
	return nil
}

func (s *Store) LookupSession(ctx context.Context, tokenHash string, now time.Time) (*auth.Session, *agentsv1.User, error) {
	var sess sessionDoc
	err := s.sessions.FindOne(ctx, bson.M{
		"token_hash": tokenHash,
		"revoked":    bson.M{"$ne": true},
		"expires_at": bson.M{"$gt": now},
	}).Decode(&sess)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil, auth.ErrSessionNotFound
		}
		return nil, nil, fmt.Errorf("lookup auth_session: %w", err)
	}

	user, err := s.GetUser(ctx, sess.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user.GetDisabled() {
		return nil, nil, auth.ErrUserDisabled
	}
	return sessionToModel(&sess), user, nil
}

func (s *Store) TouchSession(ctx context.Context, id string, at time.Time) error {
	_, err := s.sessions.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"last_used_at": at}})
	return err
}

func (s *Store) RevokeSession(ctx context.Context, id string) error {
	_, err := s.sessions.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"revoked": true}})
	return err
}

func userToProto(doc *userDoc) *agentsv1.User {
	user := &agentsv1.User{
		Id:          doc.ID,
		Username:    doc.Username,
		DisplayName: doc.DisplayName,
		AvatarUrl:   doc.AvatarURL,
		Role:        doc.Role,
		Disabled:    doc.Disabled,
		CreatedAt:   timestamppb.New(doc.CreatedAt),
		UpdatedAt:   timestamppb.New(doc.UpdatedAt),
	}
	return user
}

func sessionToModel(doc *sessionDoc) *auth.Session {
	return &auth.Session{
		ID:         doc.ID,
		UserID:     doc.UserID,
		TokenHash:  doc.TokenHash,
		CreatedAt:  doc.CreatedAt,
		ExpiresAt:  doc.ExpiresAt,
		LastUsedAt: doc.LastUsedAt,
		Revoked:    doc.Revoked,
	}
}
