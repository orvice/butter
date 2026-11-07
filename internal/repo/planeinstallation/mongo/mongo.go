package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	repo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type doc struct {
	ID, WorkspaceID, AppID, PlaneWorkspaceID, WorkspaceSlug, InstallationID, BotUserID, BotName, Status, CredentialState string
	Scopes                                                                                                               []string
	InstalledAt, UpdatedAt, ExpiresAt                                                                                    time.Time
	Token, TokenKey                                                                                                      string
	Revision                                                                                                             int64
}
type Store struct{ c *mongo.Collection }

func New(db *mongo.Database) *Store { return &Store{c: db.Collection("plane_installations")} }
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, e := s.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "app_id", Value: 1}, {Key: "plane_workspace_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uniq_plane_installation")}, {Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "id", Value: 1}}}})
	return e
}
func from(d doc) *agentsv1.PlaneInstallation {
	return &agentsv1.PlaneInstallation{Id: d.ID, AppId: d.AppID, PlaneWorkspaceId: d.PlaneWorkspaceID, WorkspaceSlug: d.WorkspaceSlug, AppInstallationId: d.InstallationID, AppBotUserId: d.BotUserID, AppBotName: d.BotName, GrantedScopes: d.Scopes, Status: d.Status, CredentialState: d.CredentialState, TokenExpiresAt: timestamppb.New(d.ExpiresAt), InstalledAt: timestamppb.New(d.InstalledAt), UpdatedAt: timestamppb.New(d.UpdatedAt)}
}
func (s *Store) Upsert(ctx context.Context, w, a string, in *agentsv1.PlaneInstallation, t repo.Tokens) (*agentsv1.PlaneInstallation, error) {
	now := time.Now().UTC()
	id := in.Id
	if id == "" {
		id = bson.NewObjectID().Hex()
	}
	var old doc
	s.c.FindOne(ctx, bson.M{"workspace_id": w, "app_id": a, "plane_workspace_id": in.GetPlaneWorkspaceId()}).Decode(&old)
	if old.ID != "" {
		id = old.ID
	}
	d := doc{ID: id, WorkspaceID: w, AppID: a, PlaneWorkspaceID: in.GetPlaneWorkspaceId(), WorkspaceSlug: in.GetWorkspaceSlug(), InstallationID: in.GetAppInstallationId(), BotUserID: in.GetAppBotUserId(), BotName: in.GetAppBotName(), Scopes: append([]string(nil), in.GetGrantedScopes()...), Status: in.GetStatus(), CredentialState: in.GetCredentialState(), InstalledAt: now, UpdatedAt: now, ExpiresAt: t.ExpiresAt, Token: t.AccessToken, TokenKey: t.AccessTokenKeyID}
	if old.ID != "" {
		d.InstalledAt = old.InstalledAt
	}
	_, e := s.c.ReplaceOne(ctx, bson.M{"workspace_id": w, "app_id": a, "plane_workspace_id": d.PlaneWorkspaceID}, d, options.Replace().SetUpsert(true))
	if e != nil {
		return nil, fmt.Errorf("upsert plane installation: %w", e)
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
	if e := s.c.FindOne(ctx, bson.M{"workspace_id": w, "app_id": a, "installation_id": installation}).Decode(&d); e != nil {
		return repo.TokenMaterial{}, e
	}
	return repo.TokenMaterial{WorkspaceID: w, AppID: a, InstallationID: installation, AccessToken: d.Token, ExpiresAt: d.ExpiresAt, Revision: d.Revision}, nil
}
func (s *Store) CASRuntimeToken(ctx context.Context, w, a, installation string, expected int64, t repo.Tokens) (repo.TokenMaterial, error) {
	r, err := s.c.UpdateOne(ctx, bson.M{"workspace_id": w, "app_id": a, "installation_id": installation, "revision": expected}, bson.M{"$set": bson.M{"token": t.AccessToken, "token_key": t.AccessTokenKeyID, "expires_at": t.ExpiresAt}, "$inc": bson.M{"revision": 1}})
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
		return nil, repo.ErrNotFound
	}
	return from(d), nil
}
