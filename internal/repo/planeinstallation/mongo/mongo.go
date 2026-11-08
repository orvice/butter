package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	repo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type doc struct {
	ID               string    `bson:"id"`
	WorkspaceID      string    `bson:"workspace_id"`
	AppID            string    `bson:"app_id"`
	PlaneWorkspaceID string    `bson:"plane_workspace_id"`
	WorkspaceSlug    string    `bson:"workspace_slug"`
	InstallationID   string    `bson:"app_installation_id"`
	BotUserID        string    `bson:"bot_user_id"`
	BotName          string    `bson:"bot_name"`
	Status           string    `bson:"status"`
	CredentialState  string    `bson:"credential_state"`
	Scopes           []string  `bson:"scopes"`
	InstalledAt      time.Time `bson:"installed_at"`
	UpdatedAt        time.Time `bson:"updated_at"`
	ExpiresAt        time.Time `bson:"expires_at"`
	Token            string    `bson:"token"`
	TokenKey         string    `bson:"token_key"`
	Revision         int64     `bson:"revision"`
}
type Store struct {
	c       *mongo.Collection
	keyring *secretbox.Keyring
}

func New(db *mongo.Database, keyring *secretbox.Keyring) *Store {
	return &Store{c: db.Collection("plane_installations"), keyring: keyring}
}
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, e := s.c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "app_id", Value: 1}, {Key: "plane_workspace_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uniq_plane_installation")},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "id", Value: 1}}},
		// One fixed provider per deployment; an external installation must not
		// be adopted by another Butter workspace/App.
		{Keys: bson.D{{Key: "app_installation_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uniq_plane_external_installation").SetPartialFilterExpression(bson.M{"app_installation_id": bson.M{"$type": "string"}})},
	})
	return e
}
func from(d doc) *agentsv1.PlaneInstallation {
	return &agentsv1.PlaneInstallation{Id: d.ID, AppId: d.AppID, PlaneWorkspaceId: d.PlaneWorkspaceID, WorkspaceSlug: d.WorkspaceSlug, AppInstallationId: d.InstallationID, AppBotUserId: d.BotUserID, AppBotName: d.BotName, GrantedScopes: d.Scopes, Status: d.Status, CredentialState: d.CredentialState, TokenExpiresAt: timestamppb.New(d.ExpiresAt), InstalledAt: timestamppb.New(d.InstalledAt), UpdatedAt: timestamppb.New(d.UpdatedAt)}
}
func (s *Store) Upsert(ctx context.Context, w, a string, in *agentsv1.PlaneInstallation, t repo.Tokens) (*agentsv1.PlaneInstallation, error) {
	if in.GetAppInstallationId() == "" || in.GetPlaneWorkspaceId() == "" || in.GetAppId() != a {
		return nil, repo.ErrConflict
	}
	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(t.AccessToken))
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	id := in.GetId()
	if id == "" {
		id = bson.NewObjectID().Hex()
	}
	var d doc
	err = s.c.FindOneAndUpdate(ctx,
		bson.M{"workspace_id": w, "app_id": a, "plane_workspace_id": in.GetPlaneWorkspaceId()},
		bson.M{
			"$set": bson.M{
				"app_installation_id": in.GetAppInstallationId(), "workspace_slug": in.GetWorkspaceSlug(),
				"bot_user_id": in.GetAppBotUserId(), "bot_name": in.GetAppBotName(), "scopes": in.GetGrantedScopes(),
				"status": in.GetStatus(), "credential_state": in.GetCredentialState(),
				"updated_at": now, "expires_at": t.ExpiresAt, "token": ciphertext, "token_key": keyID,
			},
			"$setOnInsert": bson.M{"id": id, "installed_at": now},
			"$inc":         bson.M{"revision": 1},
		}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&d)
	if mongo.IsDuplicateKeyError(err) {
		return nil, repo.ErrConflict
	}
	if err != nil {
		return nil, fmt.Errorf("upsert plane installation: %w", err)
	}
	return from(d), nil
}
func (s *Store) List(ctx context.Context, w, a string) ([]*agentsv1.PlaneInstallation, error) {
	cur, e := s.c.Find(ctx, bson.M{"workspace_id": w, "app_id": a})
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var ds []doc
	if e = cur.All(ctx, &ds); e != nil {
		return nil, e
	}
	out := make([]*agentsv1.PlaneInstallation, 0, len(ds))
	for _, d := range ds {
		out = append(out, from(d))
	}
	return out, nil
}
func (s *Store) GetRuntimeToken(ctx context.Context, w, a, installation string) (repo.TokenMaterial, error) {
	var d doc
	if e := s.c.FindOne(ctx, bson.M{"workspace_id": w, "app_id": a, "app_installation_id": installation}).Decode(&d); e != nil {
		return repo.TokenMaterial{}, mapError(e)
	}
	token, err := s.keyring.Decrypt(ctx, d.Token, d.TokenKey)
	if err != nil {
		return repo.TokenMaterial{}, errors.New("plane installation token cannot be decrypted; reinstall the App")
	}
	defer clear(token)
	return repo.TokenMaterial{WorkspaceID: w, AppID: a, InstallationID: installation, AccessToken: string(token), ExpiresAt: d.ExpiresAt, Revision: d.Revision}, nil
}
func (s *Store) CASRuntimeToken(ctx context.Context, w, a, installation string, expected int64, t repo.Tokens) (repo.TokenMaterial, error) {
	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(t.AccessToken))
	if err != nil {
		return repo.TokenMaterial{}, err
	}
	r, err := s.c.UpdateOne(ctx, bson.M{"workspace_id": w, "app_id": a, "app_installation_id": installation, "revision": expected}, bson.M{"$set": bson.M{"token": ciphertext, "token_key": keyID, "expires_at": t.ExpiresAt, "updated_at": time.Now().UTC()}, "$inc": bson.M{"revision": 1}})
	if err != nil {
		return repo.TokenMaterial{}, err
	}
	if r.MatchedCount != 1 {
		return repo.TokenMaterial{}, repo.ErrTokenCAS
	}
	return s.GetRuntimeToken(ctx, w, a, installation)
}
func (s *Store) Get(ctx context.Context, w, id string) (*agentsv1.PlaneInstallation, error) {
	var d doc
	if e := s.c.FindOne(ctx, bson.M{"workspace_id": w, "id": id}).Decode(&d); e != nil {
		return nil, mapError(e)
	}
	return from(d), nil
}
func (s *Store) GetScoped(ctx context.Context, w, a, installation string) (*agentsv1.PlaneInstallation, error) {
	if w == "" || a == "" || installation == "" {
		return nil, repo.ErrNotFound
	}
	var d doc
	if e := s.c.FindOne(ctx, bson.M{"workspace_id": w, "app_id": a, "app_installation_id": installation}).Decode(&d); e != nil {
		return nil, mapError(e)
	}
	return from(d), nil
}

func mapError(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return repo.ErrNotFound
	}
	return err
}
