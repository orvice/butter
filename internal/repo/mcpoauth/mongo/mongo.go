package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	repo "go.orx.me/apps/butter/internal/repo/mcpoauth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const collectionName = "mcp_oauth_connections"

// Store persists MCP OAuth connection records in MongoDB.
type Store struct {
	coll *mongo.Collection
}

type doc struct {
	ID                      string                           `bson:"_id"`
	WorkspaceID             string                           `bson:"workspace_id"`
	ServerID                string                           `bson:"server_id"`
	UserID                  string                           `bson:"user_id,omitempty"`
	State                   agentsv1.MCPOAuthConnectionState `bson:"state"`
	ClientID                string                           `bson:"client_id,omitempty"`
	EncryptedClientSecret   string                           `bson:"encrypted_client_secret,omitempty"`
	AuthorizationURL        string                           `bson:"authorization_url,omitempty"`
	TokenURL                string                           `bson:"token_url,omitempty"`
	Resource                string                           `bson:"resource,omitempty"`
	Scopes                  []string                         `bson:"scopes,omitempty"`
	EncryptedToken          string                           `bson:"encrypted_token,omitempty"`
	ConnectedAt             time.Time                        `bson:"connected_at,omitempty"`
	ExpiresAt               time.Time                        `bson:"expires_at,omitempty"`
	LastCheckedAt           time.Time                        `bson:"last_checked_at,omitempty"`
	LastError               string                           `bson:"last_error,omitempty"`
	ReauthorizationRequired bool                             `bson:"reauthorization_required,omitempty"`
	CreatedAt               time.Time                        `bson:"created_at,omitempty"`
	UpdatedAt               time.Time                        `bson:"updated_at,omitempty"`
}

func New(db *mongo.Database) *Store {
	return &Store{coll: db.Collection(collectionName)}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "server_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create mcp oauth workspace/server index: %w", err)
	}
	_, err = s.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "workspace_id", Value: 1}},
	})
	if err != nil {
		return fmt.Errorf("create mcp oauth workspace index: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, workspaceID, serverID string) (*repo.Connection, error) {
	var d doc
	err := s.coll.FindOne(ctx, bson.M{"_id": compositeID(workspaceID, serverID)}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("get mcp oauth connection: %w", err)
	}
	return fromDoc(&d), nil
}

func (s *Store) Save(ctx context.Context, conn *repo.Connection) error {
	now := time.Now().UTC()
	clone := conn.Clone()
	if clone.CreatedAt.IsZero() {
		if current, err := s.Get(ctx, clone.WorkspaceID, clone.ServerID); err == nil && !current.CreatedAt.IsZero() {
			clone.CreatedAt = current.CreatedAt
		} else {
			clone.CreatedAt = now
		}
	}
	if clone.UpdatedAt.IsZero() {
		clone.UpdatedAt = now
	}
	d := toDoc(clone)
	_, err := s.coll.ReplaceOne(ctx, bson.M{"_id": d.ID}, d, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("save mcp oauth connection: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, workspaceID, serverID string) error {
	_, err := s.coll.DeleteOne(ctx, bson.M{"_id": compositeID(workspaceID, serverID)})
	if err != nil {
		return fmt.Errorf("delete mcp oauth connection: %w", err)
	}
	return nil
}

func (s *Store) MarkState(ctx context.Context, workspaceID, serverID string, state agentsv1.MCPOAuthConnectionState, detail string, at time.Time) error {
	res, err := s.coll.UpdateOne(ctx, bson.M{"_id": compositeID(workspaceID, serverID)}, bson.M{"$set": bson.M{
		"state":                    state,
		"last_error":               detail,
		"last_checked_at":          at,
		"updated_at":               at,
		"reauthorization_required": state == agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED,
	}})
	if err != nil {
		return fmt.Errorf("mark mcp oauth state: %w", err)
	}
	if res.MatchedCount == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func compositeID(workspaceID, serverID string) string {
	return workspaceID + ":" + serverID
}

func toDoc(conn *repo.Connection) doc {
	return doc{
		ID:                      compositeID(conn.WorkspaceID, conn.ServerID),
		WorkspaceID:             conn.WorkspaceID,
		ServerID:                conn.ServerID,
		UserID:                  conn.UserID,
		State:                   conn.State,
		ClientID:                conn.ClientID,
		EncryptedClientSecret:   conn.EncryptedClientSecret,
		AuthorizationURL:        conn.AuthorizationURL,
		TokenURL:                conn.TokenURL,
		Resource:                conn.Resource,
		Scopes:                  append([]string(nil), conn.Scopes...),
		EncryptedToken:          conn.EncryptedToken,
		ConnectedAt:             conn.ConnectedAt,
		ExpiresAt:               conn.ExpiresAt,
		LastCheckedAt:           conn.LastCheckedAt,
		LastError:               conn.LastError,
		ReauthorizationRequired: conn.ReauthorizationRequired,
		CreatedAt:               conn.CreatedAt,
		UpdatedAt:               conn.UpdatedAt,
	}
}

func fromDoc(d *doc) *repo.Connection {
	return &repo.Connection{
		WorkspaceID:             d.WorkspaceID,
		ServerID:                d.ServerID,
		UserID:                  d.UserID,
		State:                   d.State,
		ClientID:                d.ClientID,
		EncryptedClientSecret:   d.EncryptedClientSecret,
		AuthorizationURL:        d.AuthorizationURL,
		TokenURL:                d.TokenURL,
		Resource:                d.Resource,
		Scopes:                  append([]string(nil), d.Scopes...),
		EncryptedToken:          d.EncryptedToken,
		ConnectedAt:             d.ConnectedAt,
		ExpiresAt:               d.ExpiresAt,
		LastCheckedAt:           d.LastCheckedAt,
		LastError:               d.LastError,
		ReauthorizationRequired: d.ReauthorizationRequired,
		CreatedAt:               d.CreatedAt,
		UpdatedAt:               d.UpdatedAt,
	}
}
