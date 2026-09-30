// Package mongo implements linear.Repository backed by MongoDB.
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

	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	appsCollection          = "linear_apps"
	installationsCollection = "linear_installations"
)

// appDoc keeps the secrets in dedicated columns, never inside the protojson
// spec, so they cannot leak through a decoded public model. client_id and
// revision are also top-level so the unique index and the optimistic
// revision filter can use them.
type appDoc struct {
	ID                   string    `bson:"_id"`
	WorkspaceID          string    `bson:"workspace_id"`
	ClientID             string    `bson:"client_id"`
	Revision             int64     `bson:"revision"`
	Spec                 string    `bson:"spec"`
	ClientSecret         string    `bson:"client_secret,omitempty"`
	ClientSecretKeyID    string    `bson:"client_secret_key_id,omitempty"`
	WebhookSecret        string    `bson:"webhook_secret,omitempty"`
	WebhookSecretKeyID   string    `bson:"webhook_secret_key_id,omitempty"`
	CredentialsUpdatedAt time.Time `bson:"credentials_updated_at,omitempty"`
}

func (d appDoc) credentials() linearrepo.AppCredentials {
	return linearrepo.AppCredentials{
		ClientSecret:  linearrepo.Credential{Ciphertext: d.ClientSecret, KeyID: d.ClientSecretKeyID},
		WebhookSecret: linearrepo.Credential{Ciphertext: d.WebhookSecret, KeyID: d.WebhookSecretKeyID},
	}
}

// installationDoc keeps the tokens in dedicated columns, like appDoc does
// with the secrets. Identity columns are top-level so the unique index and
// lookups can use them.
type installationDoc struct {
	ID                string    `bson:"_id"`
	WorkspaceID       string    `bson:"workspace_id"`
	AppID             string    `bson:"app_id"`
	OrganizationID    string    `bson:"organization_id"`
	Spec              string    `bson:"spec"`
	InstalledAt       time.Time `bson:"installed_at"`
	AccessToken       string    `bson:"access_token,omitempty"`
	AccessTokenKeyID  string    `bson:"access_token_key_id,omitempty"`
	RefreshToken      string    `bson:"refresh_token,omitempty"`
	RefreshTokenKeyID string    `bson:"refresh_token_key_id,omitempty"`
	ExpiresAt         time.Time `bson:"expires_at,omitempty"`
	TokenRevision     int64     `bson:"token_revision"`
}

func (d installationDoc) tokens() linearrepo.InstallationTokens {
	return linearrepo.InstallationTokens{
		AccessToken:  linearrepo.Credential{Ciphertext: d.AccessToken, KeyID: d.AccessTokenKeyID},
		RefreshToken: linearrepo.Credential{Ciphertext: d.RefreshToken, KeyID: d.RefreshTokenKeyID},
		ExpiresAt:    d.ExpiresAt,
		Revision:     d.TokenRevision,
	}
}

// Store implements linear.Repository backed by MongoDB.
type Store struct {
	apps          *mongo.Collection
	installations *mongo.Collection
}

var _ linearrepo.Repository = (*Store)(nil)

func New(db *mongo.Database) *Store {
	return &Store{
		apps:          db.Collection(appsCollection),
		installations: db.Collection(installationsCollection),
	}
}

// EnsureIndexes creates the global client ID uniqueness index. It is the
// enforcement point: a read-then-write check would let two concurrent
// creates register one Linear app twice.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.apps.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "client_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_client_id"),
		},
		{
			Keys:    bson.D{{Key: "workspace_id", Value: 1}},
			Options: options.Index().SetName("by_workspace"),
		},
	})
	if err != nil {
		return fmt.Errorf("linear app indexes: %w", err)
	}
	_, err = s.installations.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "app_id", Value: 1}, {Key: "organization_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_app_organization"),
		},
		{
			Keys:    bson.D{{Key: "workspace_id", Value: 1}, {Key: "app_id", Value: 1}},
			Options: options.Index().SetName("by_workspace_app"),
		},
	})
	if err != nil {
		return fmt.Errorf("linear installation indexes: %w", err)
	}
	return nil
}

func mapError(id string, err error) error {
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("linear app %q: %w", id, linearrepo.ErrClientIDExists)
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("linear app %q: %w", id, linearrepo.ErrNotFound)
	}
	return fmt.Errorf("linear app %q: %w", id, err)
}

func decode(doc appDoc) (*agentsv1.LinearApp, error) {
	app := &agentsv1.LinearApp{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), app); err != nil {
		return nil, fmt.Errorf("unmarshal linear app %q: %w", doc.ID, err)
	}
	app.Id = doc.ID
	app.ClientId = doc.ClientID
	app.Revision = doc.Revision
	app.WorkspaceId = doc.WorkspaceID
	linearrepo.StampCredentialState(app, doc.credentials())
	return app, nil
}

func encodeSpec(app *agentsv1.LinearApp) (string, error) {
	stored := proto.Clone(app).(*agentsv1.LinearApp)
	linearrepo.StripDerived(stored)
	spec, err := protojson.Marshal(stored)
	if err != nil {
		return "", fmt.Errorf("marshal linear app %q: %w", app.GetId(), err)
	}
	return string(spec), nil
}

func (s *Store) findOne(ctx context.Context, filter bson.M, id string) (appDoc, error) {
	var doc appDoc
	if err := s.apps.FindOne(ctx, filter).Decode(&doc); err != nil {
		return appDoc{}, mapError(id, err)
	}
	return doc, nil
}

func (s *Store) ListApps(ctx context.Context, workspaceID string) ([]*agentsv1.LinearApp, error) {
	cursor, err := s.apps.Find(ctx, bson.M{"workspace_id": workspaceID})
	if err != nil {
		return nil, fmt.Errorf("list linear apps: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []appDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode linear apps: %w", err)
	}
	out := make([]*agentsv1.LinearApp, 0, len(docs))
	for _, doc := range docs {
		app, err := decode(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, app)
	}
	sortApps(out)
	return out, nil
}

func (s *Store) GetApp(ctx context.Context, workspaceID, id string) (*agentsv1.LinearApp, error) {
	doc, err := s.findOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, id)
	if err != nil {
		return nil, err
	}
	return decode(doc)
}

func (s *Store) FindApp(ctx context.Context, id string) (*agentsv1.LinearApp, error) {
	doc, err := s.findOne(ctx, bson.M{"_id": id}, id)
	if err != nil {
		return nil, err
	}
	return decode(doc)
}

func (s *Store) CreateApp(ctx context.Context, workspaceID string, app *agentsv1.LinearApp, creds linearrepo.AppCredentials) (*agentsv1.LinearApp, error) {
	clone := proto.Clone(app).(*agentsv1.LinearApp)
	now := time.Now().UTC()
	clone.CreatedAt = timestamppb.New(now)
	clone.UpdatedAt = timestamppb.New(now)
	clone.Revision = 1
	spec, err := encodeSpec(clone)
	if err != nil {
		return nil, err
	}
	doc := appDoc{
		ID:                 clone.GetId(),
		WorkspaceID:        workspaceID,
		ClientID:           clone.GetClientId(),
		Revision:           1,
		Spec:               spec,
		ClientSecret:       creds.ClientSecret.Ciphertext,
		ClientSecretKeyID:  creds.ClientSecret.KeyID,
		WebhookSecret:      creds.WebhookSecret.Ciphertext,
		WebhookSecretKeyID: creds.WebhookSecret.KeyID,
	}
	if creds.ClientSecret.Set() || creds.WebhookSecret.Set() {
		doc.CredentialsUpdatedAt = now
	}
	if _, err := s.apps.InsertOne(ctx, doc); err != nil {
		return nil, mapError(clone.GetId(), err)
	}
	return decode(doc)
}

func (s *Store) UpdateApp(ctx context.Context, workspaceID string, app *agentsv1.LinearApp, expectedRevision int64) (*agentsv1.LinearApp, error) {
	prev, err := s.GetApp(ctx, workspaceID, app.GetId())
	if err != nil {
		return nil, err
	}
	clone := proto.Clone(app).(*agentsv1.LinearApp)
	clone.ClientId = prev.GetClientId()
	clone.CreatedAt = prev.GetCreatedAt()
	clone.UpdatedAt = timestamppb.New(time.Now().UTC())
	clone.Revision = expectedRevision + 1
	spec, err := encodeSpec(clone)
	if err != nil {
		return nil, err
	}
	res, err := s.apps.UpdateOne(ctx,
		bson.M{"_id": clone.GetId(), "workspace_id": workspaceID, "revision": expectedRevision},
		bson.M{"$set": bson.M{"spec": spec, "revision": expectedRevision + 1}})
	if err != nil {
		return nil, mapError(clone.GetId(), err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("linear app %q: %w", clone.GetId(), linearrepo.ErrRevisionConflict)
	}
	return s.GetApp(ctx, workspaceID, clone.GetId())
}

func (s *Store) SetAppCredentials(ctx context.Context, workspaceID, id string, change linearrepo.CredentialChange) (*agentsv1.LinearApp, error) {
	set := bson.M{}
	unset := bson.M{}
	apply := func(cred *linearrepo.Credential, field, keyField string) {
		if cred == nil {
			return
		}
		if cred.Set() {
			set[field] = cred.Ciphertext
			set[keyField] = cred.KeyID
		} else {
			unset[field] = ""
			unset[keyField] = ""
		}
	}
	apply(change.ClientSecret, "client_secret", "client_secret_key_id")
	apply(change.WebhookSecret, "webhook_secret", "webhook_secret_key_id")
	if len(set) == 0 && len(unset) == 0 {
		return s.GetApp(ctx, workspaceID, id)
	}
	set["credentials_updated_at"] = time.Now().UTC()
	update := bson.M{"$set": set}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	res, err := s.apps.UpdateOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, update)
	if err != nil {
		return nil, mapError(id, err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("linear app %q: %w", id, linearrepo.ErrNotFound)
	}
	return s.GetApp(ctx, workspaceID, id)
}

func (s *Store) GetAppCredentials(ctx context.Context, workspaceID, id string) (linearrepo.AppCredentials, error) {
	doc, err := s.findOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, id)
	if err != nil {
		return linearrepo.AppCredentials{}, err
	}
	return doc.credentials(), nil
}

func (s *Store) DeleteApp(ctx context.Context, workspaceID, id string) error {
	res, err := s.apps.DeleteOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID})
	if err != nil {
		return mapError(id, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("linear app %q: %w", id, linearrepo.ErrNotFound)
	}
	if _, err := s.installations.DeleteMany(ctx, bson.M{"app_id": id, "workspace_id": workspaceID}); err != nil {
		return fmt.Errorf("delete installations of linear app %q: %w", id, err)
	}
	return nil
}

func decodeInstallation(doc installationDoc) (*agentsv1.LinearInstallation, error) {
	inst := &agentsv1.LinearInstallation{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(doc.Spec), inst); err != nil {
		return nil, fmt.Errorf("unmarshal linear installation %q: %w", doc.ID, err)
	}
	inst.Id = doc.ID
	inst.WorkspaceId = doc.WorkspaceID
	inst.AppId = doc.AppID
	inst.OrganizationId = doc.OrganizationID
	inst.InstalledAt = timestamppb.New(doc.InstalledAt)
	return inst, nil
}

func (s *Store) UpsertInstallation(ctx context.Context, workspaceID string, inst *agentsv1.LinearInstallation, tokens linearrepo.InstallationTokens) (*agentsv1.LinearInstallation, error) {
	if _, err := s.GetApp(ctx, workspaceID, inst.GetAppId()); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	spec := proto.Clone(inst).(*agentsv1.LinearInstallation)
	spec.Id = ""
	spec.WorkspaceId = ""
	spec.AppId = ""
	spec.OrganizationId = ""
	spec.InstalledAt = nil
	spec.CredentialState = agentsv1.LinearInstallationCredentialState_LINEAR_INSTALLATION_CREDENTIAL_STATE_VALID
	spec.LastCredentialError = ""
	spec.UpdatedAt = timestamppb.New(now)
	encoded, err := protojson.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal linear installation: %w", err)
	}
	set := bson.M{
		"workspace_id":         workspaceID,
		"spec":                 string(encoded),
		"access_token":         tokens.AccessToken.Ciphertext,
		"access_token_key_id":  tokens.AccessToken.KeyID,
		"refresh_token":        tokens.RefreshToken.Ciphertext,
		"refresh_token_key_id": tokens.RefreshToken.KeyID,
		"expires_at":           tokens.ExpiresAt.UTC(),
	}
	var doc installationDoc
	err = s.installations.FindOneAndUpdate(ctx,
		bson.M{"app_id": inst.GetAppId(), "organization_id": inst.GetOrganizationId()},
		bson.M{
			"$set":         set,
			"$setOnInsert": bson.M{"_id": inst.GetId(), "installed_at": now},
			"$inc":         bson.M{"token_revision": 1},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("upsert linear installation: %w", err)
	}
	return decodeInstallation(doc)
}

func (s *Store) ListInstallations(ctx context.Context, workspaceID, appID string) ([]*agentsv1.LinearInstallation, error) {
	cursor, err := s.installations.Find(ctx, bson.M{"workspace_id": workspaceID, "app_id": appID})
	if err != nil {
		return nil, fmt.Errorf("list linear installations: %w", err)
	}
	defer cursor.Close(ctx)
	var docs []installationDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode linear installations: %w", err)
	}
	out := make([]*agentsv1.LinearInstallation, 0, len(docs))
	for _, doc := range docs {
		inst, err := decodeInstallation(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, inst)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].GetOrganizationName() != out[j].GetOrganizationName() {
			return out[i].GetOrganizationName() < out[j].GetOrganizationName()
		}
		return out[i].GetId() < out[j].GetId()
	})
	return out, nil
}

func (s *Store) findInstallation(ctx context.Context, filter bson.M, what string) (installationDoc, error) {
	var doc installationDoc
	err := s.installations.FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return installationDoc{}, fmt.Errorf("linear installation %s: %w", what, linearrepo.ErrNotFound)
	}
	if err != nil {
		return installationDoc{}, fmt.Errorf("linear installation %s: %w", what, err)
	}
	return doc, nil
}

func (s *Store) GetInstallation(ctx context.Context, workspaceID, id string) (*agentsv1.LinearInstallation, error) {
	doc, err := s.findInstallation(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, fmt.Sprintf("%q", id))
	if err != nil {
		return nil, err
	}
	return decodeInstallation(doc)
}

func (s *Store) FindInstallation(ctx context.Context, workspaceID, appID, organizationID string) (*agentsv1.LinearInstallation, error) {
	doc, err := s.findInstallation(ctx,
		bson.M{"workspace_id": workspaceID, "app_id": appID, "organization_id": organizationID},
		fmt.Sprintf("for organization %q", organizationID))
	if err != nil {
		return nil, err
	}
	return decodeInstallation(doc)
}

func (s *Store) DeleteInstallation(ctx context.Context, workspaceID, id string) error {
	res, err := s.installations.DeleteOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID})
	if err != nil {
		return fmt.Errorf("delete linear installation %q: %w", id, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("linear installation %q: %w", id, linearrepo.ErrNotFound)
	}
	return nil
}

func (s *Store) GetInstallationTokens(ctx context.Context, workspaceID, id string) (linearrepo.InstallationTokens, error) {
	doc, err := s.findInstallation(ctx, bson.M{"_id": id, "workspace_id": workspaceID}, fmt.Sprintf("%q", id))
	if err != nil {
		return linearrepo.InstallationTokens{}, err
	}
	return doc.tokens(), nil
}

func sortApps(apps []*agentsv1.LinearApp) {
	sort.SliceStable(apps, func(i, j int) bool {
		if apps[i].GetDisplayName() != apps[j].GetDisplayName() {
			return apps[i].GetDisplayName() < apps[j].GetDisplayName()
		}
		return apps[i].GetId() < apps[j].GetId()
	})
}
